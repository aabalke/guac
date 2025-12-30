package arm9

import (
	"math/bits"
	"unsafe"
)

const (
	THUMB_AND = iota
	THUMB_EOR
	THUMB_LSL
	THUMB_LSR
	THUMB_ASR
	THUMB_ADC
	THUMB_SBC
	THUMB_ROR
	THUMB_TST
	THUMB_NEG // arm eq is RSBS
	THUMB_CMP
	THUMB_CMN
	THUMB_ORR
	THUMB_MUL
	THUMB_BIC
	THUMB_MVN
)

func (cpu *Cpu) ThumbAlu(opcode uint16) {

	var (
		r    = &cpu.Reg.R
		cpsr = &cpu.Reg.CPSR

		inst = (opcode >> 6) & 0xF
		rsv  = r[(opcode>>3)&0x7]
		rd   = opcode & 0x7
		rdv  = r[rd]

		res uint64
	)

	switch inst {
	case THUMB_MUL:
		res = uint64(rdv) * uint64(rsv)
		r[rd] = uint32(res)
		// ARM < 4, carry flag destroyed, ARM >= 5, carry flag unchanged
		//cpsr.C = false

	case THUMB_TST:
		res = uint64(rdv) & uint64(rsv)

	case THUMB_CMN:
		res = uint64(rdv) + uint64(rsv)
		cpsr.V = ((^(rdv ^ rsv))&(rdv^uint32(res)))>>31 != 0
		cpsr.C = res >= 0x1_0000_0000

	case THUMB_CMP:
		res = uint64(rdv) - uint64(rsv)
		cpsr.C = res < 0x1_0000_0000
		cpsr.V = ((rdv^rsv)&(rdv^uint32(res)))>>31 != 0

	case THUMB_AND:
		res = uint64(rdv) & uint64(rsv)
		r[rd] = uint32(res)

	case THUMB_EOR:
		res = uint64(rdv) ^ uint64(rsv)
		r[rd] = uint32(res)

	case THUMB_ORR:
		res = uint64(rdv) | uint64(rsv)
		r[rd] = uint32(res)

	case THUMB_BIC:
		res = uint64(rdv) &^ uint64(rsv)
		r[rd] = uint32(res)

	case THUMB_MVN:
		res = ^uint64(rsv)
		r[rd] = uint32(res)

	case THUMB_NEG:
		res = 0 - uint64(rsv)
		r[rd] = uint32(res)

		cpsr.C = res < 0x1_0000_0000
		cpsr.V = ((rdv^rsv)&(rdv^uint32(res)))>>31 != 0

	case THUMB_SBC:
		res = uint64(rdv) - uint64(rsv)
		if !cpsr.C {
			res--
		}

		cpsr.V = ((rdv^rsv)&(rdv^uint32(res)))>>31 != 0
		cpsr.C = res < 0x1_0000_0000

		r[rd] = uint32(res)

	case THUMB_ADC:
		res = uint64(rdv) + uint64(rsv)
		if cpsr.C {
			res++
		}

		cpsr.V = ((^(rdv ^ rsv))&(rdv^uint32(res)))>>31 != 0
		cpsr.C = res >= 0x1_0000_0000

		r[rd] = uint32(res)

	case THUMB_LSL:
		rsv &= 0xFF

		if rsv > 32 {
			res = 0
		} else {
			res = uint64(rdv) << rsv
			if rsv != 0 {
				cpsr.C = rdv&(1<<(32-rsv)) != 0
			}
		}
		r[rd] = uint32(res)

	case THUMB_LSR:
		rsv &= 0xFF

		res = uint64(rdv) >> rsv
		r[rd] = uint32(res)

		if rsv != 0 {
			cpsr.C = rdv&(1<<(rsv-1)) != 0
		}

	case THUMB_ASR:
		rsv &= 0xFF

		if rsv > 32 {
			rsv = 32
		}

		if rsv != 0 {
			cpsr.C = rdv&(1<<(rsv-1)) != 0
		}

		res = uint64(int32(rdv) >> rsv)
		r[rd] = uint32(res)

	case THUMB_ROR:
		rsv &= 0xFF

		if rsv != 0 {
			cpsr.C = (rdv>>((rsv-1)%32))&1 != 0
		}

		res = uint64(bits.RotateLeft32(rdv, -int(rsv)))
		r[rd] = uint32(res)
	}

	cpsr.N = (res>>31)&1 != 0
	cpsr.Z = uint32(res) == 0

	r[PC] += 2
}

const (
	HI_ADD = 0
	HI_CMP = 1
	HI_MOV = 2
	HI_BX  = 3
)

func (cpu *Cpu) HiRegBX(opcode uint16) {

	var (
		r    = &cpu.Reg.R
		cpsr = &cpu.Reg.CPSR

		inst = (opcode >> 8) & 0b11
		mSBd = (opcode>>7)&1 != 0
		mSBs = (opcode>>6)&1 != 0
		rs   = (opcode >> 3) & 0xF
		rd   = (opcode) & 0x7
	)

	if inst != 3 && mSBd {
		rd |= 0b1000
	}

	if mSBs {
		rs |= 0b1000
	}

	switch inst {
	case HI_ADD:

		rsv := r[rs]
		rdv := r[rd]

		if rs == PC {
			rsv += 4
		}

		res := uint64(rsv) + uint64(rdv)

		if rd == PC {
			r[rd] = (uint32(res &^ 1)) + 4
		} else {
			r[rd] = uint32(res)
			r[PC] += 2
		}
		return

	case HI_CMP:

		rsv := r[rs]
		rdv := r[rd]

		if rs == PC {
			rsv += 4
		}

		res := uint64(rdv) - uint64(rsv)

		if rd != PC {
			r[PC] += 2
		}

		cpsr.N = (uint32(res)>>31)&1 != 0
		cpsr.Z = uint32(res) == 0
		cpsr.C = res < 0x1_0000_0000
		cpsr.V = ((rdv^rsv)&(rdv^uint32(res)))>>31 != 0
		return

	case HI_MOV:

		if nop := rs == 8 && rd == 8; nop {
			r[PC] += 2
			return
		}

		rsv := r[rs]
		if rs == PC {
			rsv += 4
		}

		if rd == PC {
			r[rd] = rsv &^ 0b1
			return
		}

		r[rd] = rsv
		r[PC] += 2
		return

	case HI_BX:

		if blx := mSBd; blx {
			if rs == LR {
				panic("Setup Thumb Blx LR, LR")
			}

			r[LR] = r[PC] + 3
		}

		if rs == PC {
			cpsr.T = false
			r[PC] = (r[PC] + 4) &^ 2
			return
		}

		if thumb := r[rs]&1 != 0; thumb {
			r[PC] = r[rs] &^ 0b1
			return
		}

		cpsr.T = false
		r[PC] = r[rs] &^ 0b11

		return
	}
}

const (
	THUMB_ADD = iota
	THUMB_SUB
	THUMB_ADDImm
	THUMB_SUBImm
)

func (cpu *Cpu) ThumbAddSub(opcode uint16) {

	var (
		r    = &cpu.Reg.R
		cpsr = &cpu.Reg.CPSR

		inst = (opcode >> 9) & 0b11
		rsv  = uint32(r[(opcode>>3)&0x7])
		rd   = (opcode) & 0x7

		res uint64
	)

	var op2 uint32
	if reg := inst < 2; reg {
		op2 = uint32(r[(opcode>>6)&0x7])
	} else {
		op2 = uint32((opcode >> 6) & 0x7)
	}

	switch inst {
	case THUMB_ADD, THUMB_ADDImm:
		res = uint64(rsv) + uint64(op2)

		cpsr.V = ((^(rsv ^ op2))&(rsv^uint32(res)))>>31 != 0
		cpsr.C = res >= 0x1_0000_0000
	case THUMB_SUB, THUMB_SUBImm:
		res = uint64(rsv) - uint64(op2)

		cpsr.V = ((rsv^op2)&(rsv^uint32(res)))>>31 != 0
		cpsr.C = res < 0x1_0000_0000
	}

	r[rd] = uint32(res)

	cpsr.N = (uint32(res)>>31)&1 != 0
	cpsr.Z = uint32(res) == 0

	r[PC] += 2
}

const (
	THUMB_IMM_MOV = iota
	THUMB_IMM_CMP
	THUMB_IMM_ADD
	THUMB_IMM_SUB
)

func (cpu *Cpu) thumbImm(opcode uint16) {

	var (
		r    = &cpu.Reg.R
		cpsr = &cpu.Reg.CPSR
		inst = (opcode >> 11) & 0b11
		rd   = (opcode >> 8) & 0x7
		nn   = uint32(opcode & 0xFF)
		rdv  = r[rd]

		res uint64
	)

	switch inst {
	case THUMB_IMM_MOV:
		res = uint64(nn)
		r[rd] = uint32(res)
	case THUMB_IMM_CMP:
		res = uint64(rdv) - uint64(nn)
		cpsr.V = ((rdv^nn)&(rdv^uint32(res)))>>31 != 0
		cpsr.C = res < 0x1_0000_0000
	case THUMB_IMM_ADD:
		res = uint64(rdv) + uint64(nn)
		r[rd] = uint32(res)
		cpsr.V = ((^(rdv ^ nn))&(rdv^uint32(res)))>>31 != 0
		cpsr.C = res >= 0x1_0000_0000
	case THUMB_IMM_SUB:
		res = uint64(rdv) - uint64(nn)
		r[rd] = uint32(res)
		cpsr.V = ((rdv^nn)&(rdv^uint32(res)))>>31 != 0
		cpsr.C = res < 0x1_0000_0000
	}

	cpsr.N = (res>>31)&1 != 0
	cpsr.Z = uint32(res) == 0

	r[PC] += 2
}

func (cpu *Cpu) thumbLSHalf(opcode uint16) {

	var (
		r = &cpu.Reg.R

		offset = uint32((opcode >> 6) & 0x1F << 1)
		addr   = r[(opcode>>3)&0x7] + offset
		rd     = (opcode) & 0x7
	)

	if ldr := (opcode>>11)&1 != 0; ldr {
		v := uint32(cpu.mem.Read16(addr&^1, true))
		is := (addr & 1) << 3
		r[rd] = bits.RotateLeft32(v, -int(is))
	} else {
		cpu.mem.Write16(addr&^1, uint16(r[rd]), true)
	}

	r[PC] += 2
}

const (
	THUMB_STRH = iota
	THUMB_LDSB
	THUMB_LDRH
	THUMB_LDSH
)

const (
	THUMB_STR_REG = iota
	THUMB_STRB_REG
	THUMB_LDR_REG
	THUMB_LDRB_REG
)

func (cpu *Cpu) thumbSdt(opcode uint16) {
	var (
		r    = &cpu.Reg.R
		inst = (opcode >> 10) & 0b11
		rd   = opcode & 0x7
		addr = r[(opcode>>3)&0x7] + r[(opcode>>6)&0x7]
	)

	if signed := (opcode>>9)&1 != 0; signed {

		switch inst {
		case THUMB_STRH:

			cpu.mem.Write16(addr&^1, uint16(r[rd]), true)

		case THUMB_LDSB:

			// sign-expand byte value
			r[rd] = uint32(int32(int8(cpu.mem.Read8(addr, true))))

		case THUMB_LDRH:

			v := cpu.mem.Read16(addr&^1, true)
			is := (addr & 1) << 3
			r[rd] = bits.RotateLeft32(v, -int(is))

		case THUMB_LDSH:

			// sign-expand half value
			r[rd] = uint32(int32(int16(cpu.mem.Read16(addr&^1, true))))
		}

		r[PC] += 2

		return
	}

	switch inst {
	case THUMB_STR_REG:
		cpu.mem.Write32(addr&^0b11, r[rd], true)
	case THUMB_STRB_REG:
		cpu.mem.Write8(addr, uint8(r[rd]), true)
	case THUMB_LDR_REG:
		v := cpu.mem.Read32(addr&^0b11, true)
		is := (addr & 0b11) << 3
		r[rd] = bits.RotateLeft32(v, -int(is))
	case THUMB_LDRB_REG:
		r[rd] = cpu.mem.Read8(addr, true)
	}

	r[PC] += 2

}

func (cpu *Cpu) thumbLPC(opcode uint16) {

	var (
		r    = &cpu.Reg.R
		rd   = (opcode >> 8) & 0x7
		nn   = uint32(opcode&0xFF) << 2
		addr = (r[PC] + 4 + nn) &^ 0b11
	)

	r[rd] = cpu.mem.Read32(addr, true)
	r[PC] += 2
}

const (
	THUMB_STR_IMM = iota
	THUMB_LDR_IMM
	THUMB_STRB_IMM
	THUMB_LDRB_IMM
)

func (cpu *Cpu) thumbLSImm(opcode uint16) {

	var (
		r = &cpu.Reg.R

		inst = (opcode >> 11) & 0b11
		rd   = opcode & 0x7
		rb   = (opcode >> 3) & 0x7
		nn   = uint32(opcode>>6) & 0x1F
	)

	switch inst {
	case THUMB_STR_IMM:
		addr := r[rb] + (nn << 2)
		cpu.mem.Write32(addr&^0b11, r[rd], true)
	case THUMB_LDR_IMM:
		addr := r[rb] + (nn << 2)
		v := cpu.mem.Read32(addr&^0b11, true)
		is := (addr & 0b11) << 3
		r[rd] = bits.RotateLeft32(v, -int(is))

	case THUMB_STRB_IMM:
		addr := r[rb] + nn
		cpu.mem.Write8(addr, uint8(r[rd]), true)
	case THUMB_LDRB_IMM:
		addr := r[rb] + nn
		r[rd] = uint32(cpu.mem.Read8(addr, true))
	}

	r[PC] += 2
}

func (cpu *Cpu) thumbPushPop(opcode uint16) {

	var (
		r     = &cpu.Reg.R
		pclr  = (opcode>>8)&1 != 0
		rlist = opcode & 0xFF
		pop   = (opcode>>11)&1 != 0
	)

	reg := 0
	if !pop {
		reg = 7
	}

	p, ok := cpu.mem.ReadPtr(r[SP], true)
	if !ok {
		p = nil
	}

	if !pop && pclr {
		r[SP] -= 4
		if p != nil {
			p = unsafe.Add(p, -4)
			*(*uint32)(p) = r[14]
		} else {
			cpu.mem.Write32(r[SP], r[14], true)
		}
	}

	for range 8 {
		if disabled := (rlist>>reg)&1 == 0; disabled {
			if pop {
				reg++
			} else {
				reg--
			}
			continue
		}

		if pop {
			if p != nil {
				r[reg] = *(*uint32)(p)
				p = unsafe.Add(p, 4)
			} else {
				r[reg] = cpu.mem.Read32(r[SP], true)
			}
			r[SP] += 4
		} else {
			r[SP] -= 4
			if p != nil {
				p = unsafe.Add(p, -4)
				*(*uint32)(p) = r[reg]
			} else {
				cpu.mem.Write32(r[SP], r[reg], true)
			}
		}

		if pop {
			reg++
		} else {
			reg--
		}
	}

	if pop && pclr {

		if p != nil {
			r[PC] = *(*uint32)(p)
		} else {
			r[PC] = cpu.mem.Read32(r[SP], true)
		}

		cpu.toggleThumb()

		r[SP] += 4
		return
	}

	r[PC] += 2
}

func (cpu *Cpu) thumbRelative(opcode uint16) {

	var (
		r  = &cpu.Reg.R
		rd = (opcode >> 8) & 0x7
		nn = uint32(opcode&0xFF) << 2
	)

	if isSP := (opcode>>11)&1 != 0; isSP {
		r[rd] = r[SP] + nn
		r[PC] += 2
		return
	}

	r[rd] = ((r[PC] + 4) &^ 0b11) + nn
	r[PC] += 2
}

func (cpu *Cpu) thumbJumpCalls(opcode uint16) {

	r := &cpu.Reg.R

	if !cpu.CheckCond(uint32(opcode>>8) & 0xF) {
		r[PC] += 2
		return
	}

	nn := int(int8(opcode&0xFF)) << 1
	r[PC] = uint32(int(r[PC]) + 4 + nn)
}

func (cpu *Cpu) thumbB(opcode uint16) {

	r := &cpu.Reg.R

	const shift = 32 - 11 // int32 - offset size

	nn := (int32(uint32(opcode)<<shift) >> shift) << 1
	r[PC] = uint32(int32(r[PC]) + 4 + nn)
}

func (cpu *Cpu) thumbShifted(opcode uint16) {

	var (
		cpsr = &cpu.Reg.CPSR
		r    = &cpu.Reg.R

		shType = (opcode >> 11) & 0b11
		is     = (opcode >> 6) & 0x1F
		rsv    = r[(opcode>>3)&0x7]
		rd     = opcode & 0x7

		res uint32
	)

	switch shType {
	case LSL:

		switch {
		case is == 0:
			res = rsv
		case is > 32:
			res = 0
			cpsr.C = false
		default:
			res = rsv << is
			cpsr.C = rsv&(1<<(32-is)) != 0
		}

	case LSR:

		if is == 0 {
			is = 32
		}

		cpsr.C = rsv&(1<<(is-1)) != 0
		res = rsv >> is

	case ASR:

		if (is == 0) || is > 32 {
			is = 32
		}

		cpsr.C = rsv&(1<<(is-1)) != 0
		res = uint32(int32(rsv) >> is)
	}

	cpsr.N = (res>>31)&1 != 0
	cpsr.Z = uint32(res) == 0

	r[rd] = res
	r[PC] += 2
}

func (cpu *Cpu) thumbStack(opcode uint16) {

	r := &cpu.Reg.R
	nn := int(opcode*0x3F) << 2

	if sub := int((opcode>>7)&1) != 0; sub {
		nn = -nn
	}

	r[SP] = uint32(int(r[SP]) + nn)
	r[PC] += 2
}

func (cpu *Cpu) thumbLongBranch(op uint16) {

	const shift = 32 - 22
	var (
		r   = &cpu.Reg.R
		op2 = cpu.mem.Read16(r[PC]+2, true)
		hi  = uint32(op & 0x7FF)
		lo  = uint32(op2 & 0x7FF)
		nn  = int32(((hi<<12)|(lo<<1))<<shift) >> shift
	)

	r[LR] = ((r[PC] + 4) &^ 1) + 1
	r[PC] = uint32(int32(r[PC]) + 4 + nn)

	if exc := (op2>>12)&1 == 0; exc {
		cpu.toggleThumb()
	}
}

func (cpu *Cpu) thumbShortLongBranch(opcode uint16) {
	// Using only the 2nd half of BL as "BL LR+imm" is possible
	// (for example, Mario Golf Advance Tour for GBA uses opcode F800h as "BL LR+0").
	// BL LR + nn
	// bottom half never signed?

	r := &cpu.Reg.R
	nn := uint32(opcode&0x7FF) << 1
	tmpLR := r[LR]
	r[LR] = ((r[PC] + 2) &^ 0b1) + 1
	r[PC] = (tmpLR + nn) &^ 0b1
}

func (cpu *Cpu) thumbLSSP(opcode uint16) {

	r := &cpu.Reg.R
	rd := (opcode >> 8) & 0x7
	addr := r[SP] + (uint32(opcode&0xFF) << 2)

	if ldr := (opcode>>11)&1 != 0; ldr {
		v := cpu.mem.Read32(addr&^0b11, true)
		is := (addr & 0b11) << 3
		r[rd] = bits.RotateLeft32(v, -int(is))
	} else {
		cpu.mem.Write32(addr, r[rd], true)
	}

	r[PC] += 2
}

func (cpu *Cpu) thumbBlock(opcode uint16) {

	var (
		r     = &cpu.Reg.R
		ldmia = (opcode>>11)&1 != 0
		rb    = (opcode >> 8) & 0x7
		rlist = opcode & 0xFF
	)

	if rlist == 0 {
		if ldmia {
			r[rb] += 0x40
			r[PC] += 8
		} else {
			cpu.mem.Write32(r[rb], r[PC]+6, true)
			r[rb] += 0x40
			r[PC] += 2
		}

		return
	}

	var (
		oldBase = r[rb]
		rbv     = r[rb]
		addr    = r[rb] &^ 0b11
	)

	p, ok := cpu.mem.ReadPtr(addr, true)
	if !ok {
		p = nil
	}

	for reg := range 8 {
		if disabled := (rlist>>reg)&1 == 0; disabled {
			continue
		}

		if p != nil {
			if ldmia {
				r[reg] = *(*uint32)(p)
			} else {
				*(*uint32)(p) = r[reg]
			}

			p = unsafe.Add(p, 4)
		} else {
			if ldmia {
				r[reg] = cpu.mem.Read32(addr, true)
			} else {
				cpu.mem.Write32(addr, r[reg], true)
			}
		}

		rbv += 4
		addr += 4
	}

	if wb := (rlist>>rb)&1 == 0; wb {
		r[rb] = rbv
		r[PC] += 2
		return
	}

	if !ldmia {
		cpu.mem.Write32(r[rb]&^0b11, oldBase, true)
	}

	r[PC] += 2
}
