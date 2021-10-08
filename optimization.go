package modbus

import (
	"fmt"
	"sort"
)

const (
	// limits to how many registers you can read / write with Modbus at once
	maxFunc16Quantity = 123
	maxFunc3Quantity  = 2047
)

func optimizeRead(r []readOp) []readOp {
	preopt := make([]readOp, len(r))
	copy(preopt, r)
	sort.Slice(preopt, func(i, j int) bool {
		return preopt[i].register < preopt[j].register
	})

	opt := make([]readOp, 0, len(preopt))
	for i := 0; i < len(preopt); i++ {
		op := preopt[i]
		for j := i + 1; j < len(preopt); j++ {
			if preopt[j].register == op.register+op.quantity &&
				op.quantity+preopt[j].quantity <= maxFunc3Quantity {
				op.quantity += preopt[j].quantity
				i++
			}
		}
		opt = append(opt, op)
	}

	return opt
}

func optimizeWrite(w []writeOp) []writeOp {
	preopt := make([]writeOp, len(w))
	copy(preopt, w)
	sort.Slice(preopt, func(i, j int) bool {
		return preopt[i].register < preopt[j].register
	})

	opt := make([]writeOp, 0, len(preopt))
	for i := 0; i < len(preopt); i++ {
		op := preopt[i]
		for j := i + 1; j < len(preopt); j++ {
			if preopt[j].register == op.register+op.quantity &&
				op.quantity+preopt[j].quantity <= maxFunc16Quantity {
				op.quantity += preopt[j].quantity
				op.value = append(op.value, preopt[j].value...)
				i++
			}
		}
		opt = append(opt, op)
	}

	return opt
}

func convertReadOp(r Read) (readOp, error) {
	ro := readOp{
		register: r.Register(),
		quantity: r.Type().Size(),
	}
	return ro, ro.validate()
}

func convertWriteOp(w Write) (writeOp, error) {
	wo := writeOp{
		register: w.Register(),
		quantity: uint16(len(w.Value().Bytes()) / 2),
		value:    w.Value().Bytes(),
	}
	return wo, wo.validate()
}

type readOp struct {
	register uint16
	quantity uint16
}

func (r readOp) validate() error {
	if r.quantity > maxFunc3Quantity {
		return fmt.Errorf("%w: %d: %v", ErrTooManyRegisters, maxFunc3Quantity, r)
	}
	return nil
}

type writeOp struct {
	register uint16
	quantity uint16
	value    []byte
}

func (w writeOp) validate() error {
	if w.quantity > maxFunc16Quantity {
		// no more than 123 registers are allowed per write operation
		return fmt.Errorf("%w: %d: %v", ErrTooManyRegisters, maxFunc16Quantity, w)
	}
	return nil
}

func newReadOp(r, q uint16) (readOp, error) {
	ro := readOp{r, q}
	return ro, ro.validate()
}

func newWriteOp(r uint16, v []byte) (writeOp, error) {
	wo := writeOp{r, uint16(len(v) / 2), v}
	return wo, wo.validate()
}
