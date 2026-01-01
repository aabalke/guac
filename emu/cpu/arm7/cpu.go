package arm7

import (
	"unsafe"

	"github.com/aabalke/guac/emu/cpu"
)

type Cpu struct {
	mem    cpu.MemoryInterface
	Irq    *cpu.Irq
	Reg    Reg
	Halted bool

	PcPtr       unsafe.Pointer
	PcOff       int
	isBranching bool
	BranchPc    uint32
	LoopCnt     uint32
	LoopLen     uint32
}

const (
	SP = 13
	LR = 14
	PC = 15

	FLAG_N = 31
	FLAG_Z = 30
	FLAG_C = 29
	FLAG_V = 28
	FLAG_Q = 27
	FLAG_I = 7
	FLAG_F = 6
	FLAG_T = 5

	MODE_USR = 0x10
	MODE_FIQ = 0x11
	MODE_IRQ = 0x12
	MODE_SWI = 0x13
	MODE_ABT = 0x17
	MODE_UND = 0x1B
	MODE_SYS = 0x1F

	BIOS_STARTUP  = 0
	BIOS_SWI      = 1
	BIOS_IRQ      = 2
	BIOS_IRQ_POST = 3
)

func (cpu *Cpu) CheckCond(cond uint32) bool {

	cpsr := cpu.Reg.CPSR

	switch cond {
	case 0xE: // AL (always)
		return true
	case 0x0: // EQ
		return cpsr.Z
	case 0x1: // NE
		return !cpsr.Z
	case 0x2: // CS/HS
		return cpsr.C
	case 0x3: // CC/LO
		return !cpsr.C
	case 0x4: // MI
		return cpsr.N
	case 0x5: // PL
		return !cpsr.N
	case 0x6: // VS
		return cpsr.V
	case 0x7: // VC
		return !cpsr.V
	case 0x8: // HI
		return cpsr.C && !cpsr.Z
	case 0x9: // LS
		return !cpsr.C || cpsr.Z
	case 0xA: // GE
		return cpsr.N == cpsr.V
	case 0xB: // LT
		return cpsr.N != cpsr.V
	case 0xC: // GT
		return !cpsr.Z && (cpsr.N == cpsr.V)
	case 0xD: // LE
		return cpsr.Z || (cpsr.N != cpsr.V)
	default: // NV
		return false
	}
}

var BANK_ID = map[uint32]uint32{
	MODE_USR: 0,
	MODE_SYS: 0,
	MODE_FIQ: 1,
	MODE_IRQ: 2,
	MODE_SWI: 3,
	MODE_ABT: 4,
	MODE_UND: 5,
}

var BIOS_ADDR = map[uint32]uint32{
	BIOS_STARTUP:  0xE129F000,
	BIOS_SWI:      0xE3A02004,
	BIOS_IRQ:      0xE25EF004,
	BIOS_IRQ_POST: 0xE55EC002,
}

func NewCpu(mem cpu.MemoryInterface, irq *cpu.Irq) *Cpu {

	c := &Cpu{
		mem: mem,
		Irq: irq,
	}

	// skip bios
	c.Irq.IME = true

	return c
}

func (c *Cpu) Execute() (int, bool) {
	if c.Reg.CPSR.T {
		return c.DecodeTHUMB()
	}

	return c.DecodeARM()
}

type Reg struct {
	R    [16]uint32
	SP   [6]uint32
	LR   [6]uint32
	FIQ  [5]uint32 // r8 - r12
	USR  [5]uint32 // r8 - r12 // tmp to restore after FIQ
	CPSR Cond
	SPSR [6]Cond
}

type Cond struct {
	N, Z, C, V, Q, I, F, T bool
	Mode                   uint32
}

func (c *Cond) Get() uint32 {

	v := c.Mode

	if c.N {
		v |= 1 << FLAG_N
	}
	if c.Z {
		v |= 1 << FLAG_Z
	}
	if c.C {
		v |= 1 << FLAG_C
	}
	if c.V {
		v |= 1 << FLAG_V
	}
	if c.Q {
		v |= 1 << FLAG_Q
	}
	if c.I {
		v |= 1 << FLAG_I
	}
	if c.F {
		v |= 1 << FLAG_F
	}
	if c.T {
		v |= 1 << FLAG_T
	}

	return v
}

func (c *Cond) Set(v uint32) {
	c.N = (v>>FLAG_N)&1 == 1
	c.Z = (v>>FLAG_Z)&1 == 1
	c.C = (v>>FLAG_C)&1 == 1
	c.V = (v>>FLAG_V)&1 == 1
	c.Q = (v>>FLAG_Q)&1 == 1
	c.I = (v>>FLAG_I)&1 == 1
	c.F = (v>>FLAG_F)&1 == 1
	c.T = (v>>FLAG_T)&1 == 1
	c.Mode = v & 0x1F
}

func (cpu *Cpu) toggleThumb() {

	reg := &cpu.Reg

	reg.CPSR.T = reg.R[PC]&1 > 0

	if reg.CPSR.T {
		reg.R[PC] &^= 1
		return
	}

	reg.R[PC] &^= 3
}

func (cpu *Cpu) CheckIrq() {

	if interrupts := cpu.Irq.IE&cpu.Irq.IF != 0; !interrupts {
		return
	}

	cpu.Halted = false

	if !cpu.Reg.CPSR.I && cpu.Irq.IME {
		cpu.exception(VEC_IRQ, MODE_IRQ)
		cpu.isBranching = true
	}
}

//type Cond uint32
//
//func (c *Cond) GetFlag(flag uint32) bool {
//	return (uint32(*c)>>flag)&0b1 == 0b1
//}
//
//func (c *Cond) SetThumb(value bool, cpu *Cpu) {
//	cpu.Reg.IsThumb = value
//	c.SetFlag(FLAG_T, value)
//}
//
//func (c *Cond) SetFlag(flag uint32, value bool) {
//
//	if value {
//		*c |= (0b1 << flag)
//		return
//	}
//
//	*c &^= (0b1 << flag)
//}
//
//func (c *Cond) SetField(loBit uint32, value uint32) {
//	mask := 0b1111_1111 << loBit
//	*c &^= Cond(mask)
//	value <<= loBit
//	*c |= Cond(value)
//}
//
//func (c *Cond) SetMode(mode uint32) {
//	*c = Cond((uint32(*c) &^ 0b11111) | mode)
//}
//
//func (r *Reg) getMode() uint32 {
//	return uint32(r.CPSR) & 0b11111
//}
//
//func (r *Reg) setMode(prev, curr uint32) {
//
//	if prev == curr {
//		return
//	}
//
//	r.CPSR.SetMode(curr)
//
//	//r._setMode(prev, curr)
//	//}
//	//
//	//func (r *Reg) _setMode(prev, curr uint32) {
//
//	if BANK_ID[prev] == BANK_ID[curr] {
//		return
//	}
//
//	r.switchRegisterBanks(prev, curr)
//}
//
//func (r *Reg) switchRegisterBanks(prev, curr uint32) {
//
//	//if BANK_ID[prev] == BANK_ID[curr] {
//	//    return
//	//}
//
//	if prev != MODE_FIQ {
//		for i := range 5 {
//			r.USR[i] = r.R[8+i]
//		}
//	}
//
//	r.SP[BANK_ID[prev]] = r.R[SP]
//	r.LR[BANK_ID[prev]] = r.R[LR]
//
//	if prev == MODE_FIQ {
//		for i := range 5 {
//			r.FIQ[i] = r.R[8+i]
//		}
//	}
//
//	if curr != MODE_FIQ {
//		for i := range 5 {
//			r.R[8+i] = r.USR[i]
//		}
//	}
//
//	r.R[SP] = r.SP[BANK_ID[curr]]
//	r.R[LR] = r.LR[BANK_ID[curr]]
//
//	if curr == MODE_FIQ {
//		for i := range 5 {
//			r.R[8+i] = r.FIQ[i]
//		}
//	}
//}

//func (cpu *Cpu) toggleThumb() {
//
//	reg := &cpu.Reg
//
//	newFlag := reg.R[PC]&1 > 0
//
//	reg.CPSR.SetThumb(newFlag, cpu)
//
//	if newFlag {
//		reg.R[PC] &^= 1
//		return
//	}
//
//	reg.R[PC] &^= 3
//}
//
//func (cpu *Cpu) CheckIrq() {
//
//	if interrupts := cpu.Irq.IE&cpu.Irq.IF != 0; !interrupts {
//        return
//    }
//
//    cpu.Halted = false
//
//	if !cpu.Reg.CPSR.GetFlag(FLAG_I) && cpu.Irq.IME {
//		cpu.exception(VEC_IRQ, MODE_IRQ)
//        cpu.PcPtr = nil
//	}
//}

//func (cpu *Cpu) GetOpArm() (uint32, int) {
//
//    if cpu.PcPtr == nil {
//        r := &cpu.Reg.R
//        if p, ok := cpu.mem.ReadPtr(r[PC], false); ok {
//            cpu.PcPtr = p
//        } else {
//            return cpu.mem.Read32(r[PC], false), 0
//        }
//    }
//
//    op := *(*uint32)(cpu.PcPtr)
//    cpu.PcPtr = unsafe.Add(cpu.PcPtr, 4)
//
//    if isBranching := ((op >> 27) & 1 == 1) || (op >> 12) & 0xF == 0xF; isBranching {
//        cpu.PcPtr = nil
//    }
//
//    return op, 0
//}
//
//func (cpu *Cpu) GetOpThumb() uint16 {
//
//    if cpu.PcPtr == nil {
//
//        r := &cpu.Reg.R
//        if p, ok := cpu.mem.ReadPtr(r[PC], false); ok {
//            cpu.PcPtr = p
//        } else {
//            return uint16(cpu.mem.Read16(r[PC], false))
//        }
//    }
//
//    op := *(*uint16)(cpu.PcPtr)
//    cpu.PcPtr = unsafe.Add(cpu.PcPtr, 2)
//
//    if isBranching := (op >> 14) != 0; isBranching {
//        cpu.PcPtr = nil
//    }
//
//    return op
//}

func (cpu *Cpu) GetOpArm() (uint32, int) {

	r := &cpu.Reg.R

	if cpu.isBranching {
		cpu.isBranching = false
		cpu.PcOff = 0

		if r[PC] != cpu.BranchPc {
			cpu.PcPtr = nil

			// imm loop ended

		} else {
			cpu.LoopCnt++
		}

		// this is here for debugging above, could probably move earlier
		cpu.LoopLen = 0
	}

	if cpu.PcPtr == nil {
		cpu.LoopCnt = 0
		cpu.BranchPc = r[PC]
		if p, ok := cpu.mem.ReadPtr(r[PC], false); ok {
			cpu.PcPtr = p
		} else {
			return cpu.mem.Read32(r[PC], false), 0
		}
	}

	op := *(*uint32)(unsafe.Add(cpu.PcPtr, cpu.PcOff))
	cpu.PcOff += 4
	cpu.LoopLen++
	cpu.isBranching = ((op>>27)&1 == 1) || (op>>12)&0xF == 0xF

	return op, 0
}

func (cpu *Cpu) GetOpThumb() uint16 {

	r := &cpu.Reg.R

	if cpu.isBranching {
		cpu.isBranching = false
		cpu.PcOff = 0
		if r[PC] != cpu.BranchPc {
			cpu.PcPtr = nil
		} else {
			cpu.LoopCnt++
		}
	}

	if cpu.PcPtr == nil {
		if p, ok := cpu.mem.ReadPtr(r[PC], false); ok {
			cpu.LoopCnt = 0
			cpu.LoopLen = 0
			cpu.BranchPc = r[PC]
			cpu.PcPtr = p
		} else {
			return uint16(cpu.mem.Read16(r[PC], false))
		}
	}

	op := *(*uint16)(unsafe.Add(cpu.PcPtr, cpu.PcOff))
	cpu.PcOff += 2
	cpu.LoopLen++
	cpu.isBranching = (op >> 14) != 0

	return op
}
