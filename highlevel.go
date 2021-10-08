package modbus

import (
	"bytes"
	"errors"
	"fmt"
	"sync"
	"time"

	"git.tdem.in/tdemin/vkr_awp/internal/memory"
)

type Client struct {
	*client
	mtx sync.RWMutex

	DoDifferentialOptimization bool
}

func NewClient(devicePath string, timeout time.Duration) (*Client, error) {
	client, err := newClient(devicePath, timeout)
	if err != nil {
		return nil, err
	}
	return &Client{
		client,
		sync.RWMutex{},
		false,
	}, nil
}

func (c *Client) ID() string {
	return c.client.ID()
}

const maxUint16 int = 65536 // covers the maximum number of Modbus registers in place

func (c *Client) BatchRead(ops []ReadOp) (map[uint16][]byte, error) {
	preopt := make([]readOp, 0, len(ops))
	for _, op := range ops {
		preopt = append(preopt, convertReadOp(op))
	}

	optimized := optimizeRead(preopt)
	results, err := c.batchRead(optimized)
	if err != nil {
		return nil, err
	}

	// align results in a flat map, get and convert results by offset
	// which is equal to Modbus register number
	mem := memory.NewMap(maxUint16)
	resultMap := make(map[uint16][]byte)
	for index, result := range results {
		mem.Set(int(index)*2, result)
	}
	for _, op := range ops {
		resultMap[op.Register()] = mem.Get(int(op.Register()*2), int(op.Quantity()*2))
	}

	return resultMap, nil
}

func (c *Client) BatchWrite(ops []WriteOp, oldData map[uint16][]byte) error {
	diffOpt := make([]writeOp, 0, len(ops))

	if oldData != nil && c.DoDifferentialOptimization {
		for _, op := range ops {
			value, ok := oldData[op.Register()]
			if !ok || !bytes.Equal(op.Bits(), value) {
				diffOpt = append(diffOpt, convertWriteOp(op))
			}
		}
	} else {
		// skip differential optimization
		for _, op := range ops {
			diffOpt = append(diffOpt, convertWriteOp(op))
		}
	}

	optimized := optimizeWrite(diffOpt)

	return c.batchWrite(optimized)
}

func (c *Client) Read(r uint16, q uint16) (interface{}, error) {
	c.mtx.RLock()
	defer c.mtx.RUnlock()

	bits, err := c.read(newReadOp(r, q))
	if err != nil {
		return nil, fmt.Errorf("Read %v: %w", r, err)
	}
	return bits, nil
}

func (c *Client) Write(r uint16, v []byte) error {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if err := c.write(newWriteOp(r, v)); err != nil {
		return fmt.Errorf("Write %v: %w", r, err)
	}

	return nil
}

func (c *Client) batchRead(ops []readOp) (map[uint16][]byte, error) {
	c.mtx.RLock()
	defer c.mtx.RUnlock()

	results := make(map[uint16][]byte)
	for i, v := range ops {
		b, err := c.read(v)
		if err != nil {
			return nil, fmt.Errorf("request %d at %d: %w", i+1, v.register, err)
		}
		results[v.register] = b
	}

	return results, nil
}

func (c *Client) batchWrite(ops []writeOp) error {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	for i, v := range ops {
		if err := c.write(v); err != nil {
			return fmt.Errorf("request %d at %d: %w", i+1, v.register, err)
		}
	}

	return nil
}

func (c *Client) read(r readOp) ([]byte, error) {
	// retry request on first error
	result, err := c.client.ReadHoldingRegisters(r.register, r.quantity)
	if err != nil {
		return c.client.ReadHoldingRegisters(r.register, r.quantity)
	}
	return result, err
}

func (c *Client) write(r writeOp) error {
	// retry request on first error
	_, err := c.client.WriteMultipleRegisters(r.register, r.quantity, r.value)
	if err != nil {
		_, err := c.client.WriteMultipleRegisters(r.register, r.quantity, r.value)
		return err
	}
	return err
}

type ReadOp interface {
	Register() uint16
	Quantity() uint16
}

type WriteOp interface {
	Register() uint16
	Bits() []byte
}

var (
	ErrOddByteNumber    = errors.New("odd bytes number not allowed")
	ErrTooManyRegisters = errors.New("too many registers in an operation")
)

func convertReadOp(r ReadOp) readOp {
	ro := readOp{
		register: uint16(r.Register()),
		quantity: r.Quantity(),
	}
	if err := ro.validate(); err != nil {
		panic(err)
	}
	return ro
}

func newReadOp(r, q uint16) readOp {
	ro := readOp{r, q}
	if err := ro.validate(); err != nil {
		panic(err)
	}
	return ro
}

func newWriteOp(r uint16, v []byte) writeOp {
	wo := writeOp{r, uint16(len(v) / 2), v}
	if err := wo.validate(); err != nil {
		panic(err)
	}
	return wo
}

func convertWriteOp(w WriteOp) writeOp {
	wo := writeOp{
		register: uint16(w.Register()),
		quantity: uint16(len(w.Bits()) / 2),
		value:    w.Bits(),
	}
	if err := wo.validate(); err != nil {
		panic(err)
	}
	return wo
}

type readOp struct {
	register uint16
	quantity uint16
}

func (r readOp) validate() error {
	if r.quantity > maxModbusFunc3Quantity {
		return fmt.Errorf("%d: %w", maxModbusFunc3Quantity, ErrTooManyRegisters)
	}
	return nil
}

type writeOp struct {
	register uint16
	quantity uint16
	value    []byte
}

func (w writeOp) validate() error {
	// single Modbus register is 2 bytes, no more than 123 registers
	// allowed per write operation
	if len(w.value)%2 != 0 {
		return ErrOddByteNumber
	}
	if w.quantity > maxModbusFunc16Quantity {
		return fmt.Errorf("%d: %w", maxModbusFunc16Quantity, ErrTooManyRegisters)
	}
	return nil
}
