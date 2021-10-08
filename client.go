// Package modbus implements a proof-of-concept thread-safe Modbus
// client that operates on batches of requests.
//
// The client is designed to work on high-latency connections where an
// average delay between a Modbus request and the response to it gets up
// to a few seconds by using Modbus functions 3 and 16 to combine
// requests which can be merged into one. See doc of
// BatchRead/BatchWrite for details.
//
// The client can also optimize batch writes by difference against an
// older set of values. See doc of BatchWrite for details.
//
// Algorithm
//
// 1. For write operations: if differential optimization is enabled,
// exclude all operations whose register and value match the older
// written data.
//
// 2. Sort operations by register number in ascending order.
//
// 3. If out of two consecutive operations A and B the following
// condition holds true:
//
//  A.register + A.quantity = B.register
//
// and the total quantity after merge does not exceed 2047 for reads and
// 123 for writes, merge operations.
package modbus

import (
	"bytes"
	"errors"
	"fmt"
	"sync"

	"github.com/goburrow/modbus"
	"github.com/tdemin/opmodbus/internal/containers"
	"github.com/tdemin/opmodbus/types"
)

// Client is an optimizing Modbus client that operates on chains of
// requests. It can only execute functions 3 and 16.
//
// Client is only thread-safe if Client and ClientHandler are untouched.
type Client struct {
	modbus.Client
	modbus.ClientHandler

	mtx sync.Mutex
}

// NewClient builds a Modbus client from ClientHandler.
func NewClient(handler modbus.ClientHandler) *Client {
	return &Client{modbus.NewClient(handler), handler, sync.Mutex{}}
}

// Read represents Modbus function 3 call for a single value.
type Read interface {
	Register() uint16
	Type() types.Type
}

// Write represents Modbus function 16 call for a single value.
type Write interface {
	Register() uint16
	Value() types.Value
}

// ErrTooManyRegisters is returned when a number of registers exceeds
// 123 for writes, and 2047 for reads.
var ErrTooManyRegisters = errors.New("too many registers in an operation")

const maxUint16 int = 65536 // covers the maximum number of Modbus registers in place

// Registers holds a mapping of a Modbus registers set to their values.
type Registers map[uint16]types.Value

// BatchRead optimizes a batch of read operations, performs them with
// function 3 and returns a map of Modbus registers with their
// corresponding values.
//
// See package documentation for the optimization algoritm.
func (c *Client) BatchRead(ops []Read) (Registers, error) {
	preopt := make([]readOp, 0, len(ops))
	for _, op := range ops {
		rop, err := convertReadOp(op)
		if err != nil {
			return nil, err
		}
		preopt = append(preopt, rop)
	}

	optimized := optimizeRead(preopt)
	results, err := c.batchRead(optimized)
	if err != nil {
		return nil, err
	}

	// align results in a flat map, get and convert results by offset
	// which is equal to Modbus register number
	mem := containers.NewSlice(maxUint16)
	resultMap := make(Registers)
	for index, result := range results {
		mem.Set(int(index)*2, result)
	}
	for _, op := range ops {
		result, err := op.Type().Converter()(mem.Get(int(op.Register())*2, int(op.Type().Size())*2))
		if err != nil {
			return nil, fmt.Errorf("%v: %w", op, err)
		}
		resultMap[op.Register()] = result
	}

	return resultMap, nil
}

// BatchWrite optimizes a batch of write operations, performs them with
// function 16 and returns on the first error encountered.
//
// If oldData is not nil, BatchWrite will perform differential
// optimization. See package documentation for the optimization
// algorithm.
//
// Only use differential optimization if it is well-known that the slave
// registers values never change between BatchWrite invocations.
func (c *Client) BatchWrite(ops []Write, oldData Registers) error {
	diffOpt := make([]writeOp, 0, len(ops))

	if oldData != nil {
		for _, op := range ops {
			value, ok := oldData[op.Register()]
			if !ok || !bytes.Equal(op.Value().Bytes(), value.Bytes()) {
				wop, err := convertWriteOp(op)
				if err != nil {
					return err
				}
				diffOpt = append(diffOpt, wop)
			}
		}
	} else {
		// skip differential optimization
		for _, op := range ops {
			wop, err := convertWriteOp(op)
			if err != nil {
				return err
			}
			diffOpt = append(diffOpt, wop)
		}
	}

	optimized := optimizeWrite(diffOpt)
	return c.batchWrite(optimized)
}

// Read reads a single value from one or more Modbus registers with
// function 3 and converts it to Value. The number of Modbus registers
// is automatically picked based on provided type.
func (c *Client) Read(register uint16, t types.Type) (types.Value, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	op, err := newReadOp(register, t.Size())
	if err != nil {
		return nil, err
	}
	res, err := c.read(op)
	if err != nil {
		return nil, err
	}

	return t.Converter()(res)
}

// Write writes a single value to one or more Modbus registers with
// function 16. The number of Modbus registers is automatically picked
// based on value size.
func (c *Client) Write(register uint16, value types.Value) error {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	op, err := newWriteOp(register, value.Bytes())
	if err != nil {
		return err
	}

	return c.write(op)
}

func (c *Client) batchRead(ops []readOp) (map[uint16][]byte, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	results := make(map[uint16][]byte)
	for i, v := range ops {
		b, err := c.read(v)
		if err != nil {
			return nil, fmt.Errorf("read request %d at %d: %w", i+1, v.register, err)
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
			return fmt.Errorf("write request %d at %d: %w", i+1, v.register, err)
		}
	}

	return nil
}

func (c *Client) read(r readOp) ([]byte, error) {
	return c.ReadHoldingRegisters(r.register, r.quantity)
}

func (c *Client) write(w writeOp) error {
	_, err := c.WriteMultipleRegisters(w.register, w.quantity, w.value)
	return err
}
