package modbus

import "sort"

// TODO: add tests

const (
	// limits to how many registers you can read / write with Modbus at once
	maxModbusFunc16Quantity = 123
	maxModbusFunc3Quantity  = 2047
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
			// FIXME: both optimizations functions should check whether or not
			// the condition is satisfied before, and not after merging
			if op.quantity >= maxModbusFunc3Quantity {
				break
			}
			if preopt[j].register == op.register+op.quantity {
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
			if op.quantity >= maxModbusFunc16Quantity {
				break
			}
			if preopt[j].register == op.register+op.quantity {
				op.quantity += preopt[j].quantity
				op.value = append(op.value, preopt[j].value...)
				i++
			}
		}
		opt = append(opt, op)
	}

	return opt
}
