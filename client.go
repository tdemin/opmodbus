// modbus implements a thread-safe Modbus client that operates on chains
// of requests.
package modbus

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/goburrow/modbus"
)

var (
	// rtu://[/dev/ttyUSB0]:[16]
	rtuDevicePathRegex = regexp.MustCompile(`^rtu:\/\/(.+):(\d+)$`)
	// tcp://[127.0.0.1:502]
	tcpDevicePathRegex = regexp.MustCompile(`^tcp:\/\/(.+)$`)
)

type handler interface {
	// these two are not present on Modbus ASCII client handler

	Connect() error
	Close() error
	modbus.ClientHandler
}

// client implements a thread-safe Modbus client, wrapping goburrow/modbus.
type client struct {
	client  modbus.Client
	handler handler // only stored for closing/opening connections
	id      string

	mtx sync.Mutex
}

var ErrInvalidURI = errors.New("not a device URI")

func newClient(devicePath string, timeout time.Duration) (*client, error) {
	var (
		h  handler
		id string
	)

	if match := tcpDevicePathRegex.FindStringSubmatch(devicePath); len(match) == 2 {
		h = modbus.NewTCPClientHandler(match[1])
		id = match[0]
	} else if match := rtuDevicePathRegex.FindStringSubmatch(devicePath); len(match) == 3 {
		slaveID, err := strconv.ParseUint(match[2], 10, 16)
		if err != nil {
			return nil, fmt.Errorf("%w: slave ID not specified: %s", ErrInvalidURI, devicePath)
		}
		rtuHandler := modbus.NewRTUClientHandler(match[1])
		rtuHandler.BaudRate = 115200
		rtuHandler.DataBits = 8
		rtuHandler.Parity = "N"
		rtuHandler.StopBits = 1
		rtuHandler.SlaveId = byte(slaveID)
		rtuHandler.Timeout = timeout
		h = rtuHandler
		id = match[2]
	} else {
		return nil, fmt.Errorf("%w: %s", ErrInvalidURI, devicePath)
	}

	return &client{
		client:  modbus.NewClient(h),
		handler: h,
		id:      id,
		mtx:     sync.Mutex{},
	}, nil
}

func (c *client) ID() string {
	return c.id
}

func (c *client) Connect() error {
	return c.handler.Connect()
}

func (c *client) Close() error {
	return c.handler.Close()
}

func (c *client) lock() {
	c.mtx.Lock()
}

func (c *client) unlock() {
	c.mtx.Unlock()
}

// ReadCoils reads from 1 to 2000 contiguous status of coils in a remote
// device and returns coil status.
func (c *client) ReadCoils(address uint16, quantity uint16) (results []byte, err error) {
	c.lock()
	defer c.unlock()
	return c.client.ReadCoils(address, quantity)
}

// ReadDiscreteInputs reads from 1 to 2000 contiguous status of discrete
// inputs in a remote device and returns input status.
func (c *client) ReadDiscreteInputs(address uint16, quantity uint16) (results []byte, err error) {
	c.lock()
	defer c.unlock()
	return c.client.ReadDiscreteInputs(address, quantity)
}

// WriteSingleCoil write a single output to either ON or OFF in a remote
// device and returns output value.
func (c *client) WriteSingleCoil(address uint16, value uint16) (results []byte, err error) {
	c.lock()
	defer c.unlock()
	return c.client.WriteSingleCoil(address, value)
}

// WriteMultipleCoils forces each coil in a sequence of coils to either
// ON or OFF in a remote device and returns quantity of outputs.
func (c *client) WriteMultipleCoils(address uint16, quantity uint16, value []byte) (results []byte, err error) {
	c.lock()
	defer c.unlock()
	return c.client.WriteMultipleCoils(address, quantity, value)
}

// ReadInputRegisters reads from 1 to 125 contiguous input registers in
// a remote device and returns input registers.
func (c *client) ReadInputRegisters(address uint16, quantity uint16) (results []byte, err error) {
	c.lock()
	defer c.unlock()
	return c.client.ReadInputRegisters(address, quantity)
}

// ReadHoldingRegisters reads the contents of a contiguous block of
// holding registers in a remote device and returns register value.
func (c *client) ReadHoldingRegisters(address uint16, quantity uint16) (results []byte, err error) {
	c.lock()
	defer c.unlock()
	return c.client.ReadHoldingRegisters(address, quantity)
}

// WriteSingleRegister writes a single holding register in a remote
// device and returns register value.
func (c *client) WriteSingleRegister(address uint16, value uint16) (results []byte, err error) {
	c.lock()
	defer c.unlock()
	return c.client.WriteSingleRegister(address, value)
}

// WriteMultipleRegisters writes a block of contiguous registers (1 to
// 123 registers) in a remote device and returns quantity of registers.
func (c *client) WriteMultipleRegisters(address uint16, quantity uint16, value []byte) (results []byte, err error) {
	c.lock()
	defer c.unlock()
	return c.client.WriteMultipleRegisters(address, quantity, value)
}

// ReadWriteMultipleRegisters performs a combination of one read
// operation and one write operation. It returns read registers value.
func (c *client) ReadWriteMultipleRegisters(readAddress uint16, readQuantity uint16, writeAddress uint16, writeQuantity uint16, value []byte) (results []byte, err error) {
	c.lock()
	defer c.unlock()
	return c.client.ReadWriteMultipleRegisters(readAddress, readQuantity, writeAddress, writeQuantity, value)
}

// MaskWriteRegister modify the contents of a specified holding register
// using a combination of an AND mask, an OR mask, and the register's
// current contents. The function returns AND-mask and OR-mask.
func (c *client) MaskWriteRegister(address uint16, andMask uint16, orMask uint16) (results []byte, err error) {
	c.lock()
	defer c.unlock()
	return c.client.MaskWriteRegister(address, andMask, orMask)
}

// ReadFIFOQueue reads the contents of a First-In-First-Out (FIFO) queue
// of register in a remote device and returns FIFO value register.
func (c *client) ReadFIFOQueue(address uint16) (results []byte, err error) {
	c.lock()
	defer c.unlock()
	return c.client.ReadFIFOQueue(address)
}
