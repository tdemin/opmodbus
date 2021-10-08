# Optimizing Modbus client

[![Go Reference](https://pkg.go.dev/badge/github.com/tdemin/opmodbus.svg)](https://pkg.go.dev/github.com/tdemin/opmodbus)
[![Go Report Card](https://goreportcard.com/badge/github.com/tdemin/opmodbus)](https://goreportcard.com/report/github.com/tdemin/opmodbus)
![License](https://img.shields.io/github/license/tdemin/opmodbus)

This is a proof-of-concept Modbus master library that operates in
batches of requests, based on
[goburrow/modbus](https://github.com/goburrow/modbus).

The client itself is a refactored extract from a bachelor's thesis on
high-delay connection communication optimization methods, where Modbus
RTU would go over a connection with very high latency.

## Optimization methods

The client uses the properties of Modbus functions 3/16 (Read Holding
Registers / Write Multiple Holding Registers), specifically their
ability to operate on multiple registers at once. If registers +
quantity are closely connected (e.g. requests with register 82, quantity
2, and register 84 will be connected), they can be merged into a single
request.

For write requests, the master can also use the previous values to only
send the requests containing differentiating values to the slave.

## Example

```go
package main

import (
    "fmt"

    goburrow "github.com/goburrow/modbus"
    "github.com/tdemin/opmodbus/types"
    modbus "github.com/tdemin/opmodbus"
)

func main() {
    handler := goburrow.NewTCPClientHandler("localhost:502")
    defer handler.Close()
    client := modbus.NewClient(handler)
    // client will batch these three reads into a single request
    // with function 3, register 82 and quantity 6
    results, err := client.BatchRead([]modbus.Read{
        ReadFloat{82},
        ReadFloat{84},
        ReadFloat{86},
    })
    if err != nil {
        panic(err)
    }
    for register, value := range results {
        fmt.Printf("%v: %v", register, value)
    }
    // client will batch these two writes into a single request
    // with function 16, register 82, quantity 4, and merged data from
    // both requests
    if client.BatchWrite([]modbus.Write{
        WriteFloat{82, 0.4},
        WriteFloat{84, 0.6},
    }, nil); err != nil {
        panic(err)
    }
    // with differential optimization against old values
    if client.BatchWrite([]modbus.Write{
        WriteFloat{82, 0.3},
        WriteFloat{84, 0.7},
    }, results); err != nil {
        panic(err)
    }
}

// implements read/write operation interfaces
type ReadFloat struct {
    register uint16
}
func (r ReadFloat) Register() uint16 {
    return r.register
}
func (ReadFloat) Type() types.Type {
    // opmodbus has pre-built types for float32 and uint16
    return types.Float32CDABType
}

type WriteFloat struct {
    register uint16
    value float32
}
func (w WriteFloat) Register() uint16 {
    return w.register
}
func (w WriteFloat) Value() types.Value {
    return types.Float32CDAB(w.value)
}
```

## Caveats

The client is currently only capable of using functions 3/16 for
read/write operations.

The optimization premise is based on the assumption that the Modbus
slave is capable of having its registers grouped one after other in PLC
configuration. This has proven to be not true for some PLCs like OVEN
PLC-63. You should verify that your PLC works in such a configuration
first.

Differential optimization should only be used under two conditions:

1. You know the registers you set will never change between `BatchWrite`
   invocations (otherwise invoking `BatchWrite` involves a risk of
   skipping required updates).
2. The registers to be set aren't going in the straight order right
   after another (otherwise skipping differential optimization is more
   likely to save time).

## Copying

See [COPYING](COPYING).
