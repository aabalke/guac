package debug

import "fmt"

var pcs [50]struct {
	r    [16]uint32
	op   uint32
	cpsr uint32
}

var head int

func UpdatePcs(r [16]uint32, op, cpsr uint32) {
	pcs[head] = struct {
		r    [16]uint32
		op   uint32
		cpsr uint32
	}{r, op, cpsr}

	head = (head + 1) % len(pcs)
}

func PrintPcs() {
	for i := range len(pcs) {
		idx := (head + i) % len(pcs)
		v := pcs[idx]
		fmt.Printf("R %08X OP %08X CPSR %08X\n", v.r, v.op, v.cpsr)
	}
}
