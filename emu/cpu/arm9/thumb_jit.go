package arm9

import amd64 "github.com/aabalke/gojit"

func (j *Jit) emitLSSP(op uint32) {

	rd := (op >> 8) & 0x7

	j.Movl(j.REG(SP), amd64.Eax)
    j.Add(amd64.Imm((op&0xFF)<<2), amd64.Eax)

	if ldr := (op>>11)&1 != 0; ldr {

        j.Movl(amd64.Eax, j.SCRATCH(0))
        j.And(amd64.Imm(^0b11), amd64.Eax)
        j.CallFunc(Read32)

        j.Movl(j.SCRATCH(0), amd64.Ecx)
        j.And(amd64.Imm(0b11), amd64.Ecx)
        j.Shl(amd64.Imm(3), amd64.Ecx)

        j.RorCl(amd64.Eax)

        j.Movl(amd64.Eax, j.REG(rd))



		//v := cpu.mem.Read32(addr&^0b11, true)
		//is := (addr & 0b11) << 3
		//r[rd] = bits.RotateLeft32(v, -int(is))
	} else {

        j.Movl(j.REG(rd), amd64.Ebx)
        j.CallFunc(Write32)
		//cpu.mem.Write32(addr, r[rd], true)
	}
}
