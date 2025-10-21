package uhh

import "fmt"

var pcs [50][3]uint32

func PrintPcs() {
	for _, v := range pcs {
		fmt.Printf("PC %08X OP %08X CPSR %08X\n", v[0], v[1], v[2])
	}
}

func UpdatePcs(pc, opcode, cpsr uint32) {

    for i := 1; i < len(pcs); i++ {
        pcs[i-1] = pcs[i]
    }

    pcs[len(pcs) - 1] = [3]uint32{pc, opcode, cpsr}
}
