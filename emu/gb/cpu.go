package gameboy

import (
	"fmt"
	"unsafe"
)

// 4 cycles per m clock (if inst/opcode take 4 cycles, inc m by 1, 8 cycles, by 2

const right, throughCarry, acc = true, true, true

var branchingOps [256 / 8]uint8 // 32 bytes = 256 bits

func init() {
	for _, op := range []uint8{
		0xC2, 0xD2, 0xC3, 0xCA, 0xDA, 0xE9,
		0x18, 0x20, 0x28, 0x30, 0x38,
		0xC7, 0xD7, 0xE7, 0xF7, 0xCF, 0xDF, 0xEF, 0xFF,
		0xC4, 0xD4, 0xCC, 0xDC, 0xCD,
		0xC0, 0xD0, 0xC8, 0xD8, 0xC9, 0xD9, 0xCB,
	} {
		branchingOps[op/8] |= 1 << (op & 7)
	}
}

func isBranching(op uint8) bool {
	return branchingOps[op>>3]&(1<<(op&7)) != 0
}

type Cpu struct {
	IME              bool
	IE, IF           uint8
	PendingInterrupt bool
	Halted           bool

	a uint8
	f Flags

	c uint8
	b uint8

	e uint8
	d uint8

	l uint8
	h uint8

	//AF
	BC *uint16
	DE *uint16
	HL *uint16

	PC uint16
	SP uint16

	// optimizations
	isBranching bool
	PcPtr       unsafe.Pointer
	PcOff       int
	BranchPc    uint16
}

type Flags struct {
	Z, S, H, C bool
}

func (f *Flags) Get() uint8 {

	var v uint8

	if f.Z {
		v |= 1 << 7
	}

	if f.S {
		v |= 1 << 6
	}

	if f.H {
		v |= 1 << 5
	}

	if f.C {
		v |= 1 << 4
	}

	return v
}

func (f *Flags) Set(v uint8) {
	f.Z = (v>>7)&1 != 0
	f.S = (v>>6)&1 != 0
	f.H = (v>>5)&1 != 0
	f.C = (v>>4)&1 != 0
}

func NewCpu() *Cpu {
	c := &Cpu{
		a: 0x01,
		b: 0x00,
		c: 0x13,
		d: 0x00,
		e: 0xD8,
		h: 0x01,
		l: 0x4D,
		f: Flags{
			Z: true,
			S: false,
			H: true,
			C: true,
		},

		IME:              false,
		PendingInterrupt: false,
		PC:               0x0100,
		SP:               0xFFFE,
	}

	c.BC = (*uint16)(unsafe.Pointer(&c.c))
	c.DE = (*uint16)(unsafe.Pointer(&c.e))
	c.HL = (*uint16)(unsafe.Pointer(&c.l))

	return c
}

func (gb *GameBoy) GetOp() uint8 {

	cpu := gb.Cpu

	if cpu.isBranching {
		cpu.isBranching = false
		cpu.PcOff = 0

		if cpu.PC != cpu.BranchPc {
			cpu.PcPtr = nil
		}
	}

	if sequential := cpu.PcPtr == nil; sequential {
		cpu.BranchPc = cpu.PC
		if p := gb.ReadPtr(cpu.PC); p != nil {
			cpu.PcPtr = p
		} else {
			//cpu.isBranching = true
			return gb.Read(cpu.PC)
		}
	}

	op := *(*uint8)(unsafe.Add(cpu.PcPtr, cpu.PcOff))
	cpu.PcOff++

	cpu.isBranching = isBranching(op)

	return op
}

func (gb *GameBoy) getImm8() uint8 {

	if gb.Cpu.PcPtr != nil {
		return *(*uint8)(unsafe.Add(gb.Cpu.PcPtr, gb.Cpu.PcOff))
	}

	return gb.Read(gb.Cpu.PC + 1)
}

func (gb *GameBoy) getImm16() uint16 {

	if gb.Cpu.PcPtr != nil {
		return *(*uint16)(unsafe.Add(gb.Cpu.PcPtr, gb.Cpu.PcOff))
	}

	return uint16(gb.Read(gb.Cpu.PC+2))<<8 | uint16(gb.Read(gb.Cpu.PC+1))
}

func (gb *GameBoy) Execute() {

	cycles := 0
	pc := gb.Cpu.PC + 1
	reg := gb.Cpu

	gb.Tick(4)
	op := gb.GetOp()
	//op := gb.Read(gb.Cpu.PC)

	//L.WriteLog(cnt, op)
	//cnt++

	switch op {
	case 0x00: // nop
	case 0x10: // stop / toggle speed

		if gb.Color && gb.PrepareSpeedToggle {
			gb.Tick(8200)
			gb.toggleDoubleSpeed()
		} else {
			gb.Cpu.Halted = true
			gb.Timer.Div = 0
		}

		pc++
		reg.PcOff++

	// Load State Move
	case 0x40:
		reg.b = reg.b
	case 0x41:
		reg.b = reg.c
	case 0x42:
		reg.b = reg.d
	case 0x43:
		reg.b = reg.e
	case 0x44:
		reg.b = reg.h
	case 0x45:
		reg.b = reg.l
	case 0x46:
		gb.Tick(4)
		reg.b = gb.Read(*reg.HL)
	case 0x47:
		reg.b = reg.a
	case 0x48:
		reg.c = reg.b
	case 0x49:
		reg.c = reg.c
	case 0x4A:
		reg.c = reg.d
	case 0x4B:
		reg.c = reg.e
	case 0x4C:
		reg.c = reg.h
	case 0x4D:
		reg.c = reg.l
	case 0x4E:
		gb.Tick(4)
		reg.c = gb.Read(*reg.HL)
	case 0x4F:
		reg.c = reg.a

	case 0x50:
		reg.d = reg.b
	case 0x51:
		reg.d = reg.c
	case 0x52:
		reg.d = reg.d
	case 0x53:
		reg.d = reg.e
	case 0x54:
		reg.d = reg.h
	case 0x55:
		reg.d = reg.l
	case 0x56:
		gb.Tick(4)
		reg.d = gb.Read(*reg.HL)
	case 0x57:
		reg.d = reg.a

	case 0x58:
		reg.e = reg.b
	case 0x59:
		reg.e = reg.c
	case 0x5A:
		reg.e = reg.d
	case 0x5B:
		reg.e = reg.e
	case 0x5C:
		reg.e = reg.h
	case 0x5D:
		reg.e = reg.l
	case 0x5E:
		gb.Tick(4)
		reg.e = gb.Read(*reg.HL)
	case 0x5F:
		reg.e = reg.a

	case 0x60:
		reg.h = reg.b
	case 0x61:
		reg.h = reg.c
	case 0x62:
		reg.h = reg.d
	case 0x63:
		reg.h = reg.e
	case 0x64:
		reg.h = reg.h
	case 0x65:
		reg.h = reg.l
	case 0x66:
		gb.Tick(4)
		reg.h = gb.Read(*reg.HL)
	case 0x67:
		reg.h = reg.a

	case 0x68:
		reg.l = reg.b
	case 0x69:
		reg.l = reg.c
	case 0x6A:
		reg.l = reg.d
	case 0x6B:
		reg.l = reg.e
	case 0x6C:
		reg.l = reg.h
	case 0x6D:
		reg.l = reg.l
	case 0x6E:
		gb.Tick(4)
		reg.l = gb.Read(*reg.HL)
	case 0x6F:
		reg.l = reg.a
	case 0x01:
		gb.Tick(8)
		*reg.BC = gb.getImm16()
		pc = pc + 2
		reg.PcOff += 2
	case 0x11:
		gb.Tick(8)
		*reg.DE = gb.getImm16()
		pc = pc + 2
		reg.PcOff += 2
	case 0x21:
		gb.Tick(8)
		*reg.HL = gb.getImm16()
		pc = pc + 2
		reg.PcOff += 2
	case 0x31:
		gb.Tick(8)
		reg.SP = gb.getImm16()
		pc = pc + 2
		reg.PcOff += 2

	case 0x70:

		gb.Tick(4)
		gb.Write(*reg.HL, reg.b)
	case 0x71:
		gb.Tick(4)
		gb.Write(*reg.HL, reg.c)
	case 0x72:
		gb.Tick(4)
		gb.Write(*reg.HL, reg.d)
	case 0x73:
		gb.Tick(4)
		gb.Write(*reg.HL, reg.e)
	case 0x74:
		gb.Tick(4)
		gb.Write(*reg.HL, reg.h)
	case 0x75:
		gb.Tick(4)
		gb.Write(*reg.HL, reg.l)
	case 0x76:
		gb.Cpu.Halted = true

	case 0x77:
		gb.Tick(4)
		gb.Write(*reg.HL, reg.a)

	case 0x0E:
		gb.Tick(4)
		reg.c = gb.getImm8()
		pc++
		reg.PcOff++
	case 0x1E:
		gb.Tick(4)
		reg.e = gb.getImm8()
		pc++
		reg.PcOff++
	case 0x2E:
		gb.Tick(4)
		reg.l = gb.getImm8()
		pc++
		reg.PcOff++
	case 0x3E:
		gb.Tick(4)
		reg.a = gb.getImm8()
		pc++
		reg.PcOff++

	case 0x78:
		reg.a = reg.b
	case 0x79:
		reg.a = reg.c
	case 0x7A:
		reg.a = reg.d
	case 0x7B:
		reg.a = reg.e
	case 0x7C:
		reg.a = reg.h
	case 0x7D:
		reg.a = reg.l
	case 0x7E:
		gb.Tick(4)
		reg.a = gb.Read(*reg.HL)
	case 0x7F:
		reg.a = reg.a

	case 0x80:
		reg.a = gb.execAdd(reg.a, reg.b)
	case 0x81:
		reg.a = gb.execAdd(reg.a, reg.c)
	case 0x82:
		reg.a = gb.execAdd(reg.a, reg.d)
	case 0x83:
		reg.a = gb.execAdd(reg.a, reg.e)
	case 0x84:
		reg.a = gb.execAdd(reg.a, reg.h)
	case 0x85:
		reg.a = gb.execAdd(reg.a, reg.l)
	case 0x86:
		gb.Tick(4)
		reg.a = gb.execAdd(reg.a, gb.Read(*reg.HL))
	case 0x87:
		reg.a = gb.execAdd(reg.a, reg.a)

	case 0x09:
		gb.Tick(4)
		*reg.HL = gb.execAddHl(*reg.HL, *reg.BC)
	case 0x19:
		gb.Tick(4)
		*reg.HL = gb.execAddHl(*reg.HL, *reg.DE)
	case 0x29:
		gb.Tick(4)
		*reg.HL = gb.execAddHl(*reg.HL, *reg.HL)
	case 0x39:
		gb.Tick(4)
		*reg.HL = gb.execAddHl(*reg.HL, gb.Cpu.SP)

	case 0x88:
		reg.a = gb.execAdc(reg.a, reg.b)
	case 0x89:
		reg.a = gb.execAdc(reg.a, reg.c)
	case 0x8A:
		reg.a = gb.execAdc(reg.a, reg.d)
	case 0x8B:
		reg.a = gb.execAdc(reg.a, reg.e)
	case 0x8C:
		reg.a = gb.execAdc(reg.a, reg.h)
	case 0x8D:
		reg.a = gb.execAdc(reg.a, reg.l)
	case 0x8E:
		gb.Tick(4)
		reg.a = gb.execAdc(reg.a, gb.Read(*reg.HL))
	case 0x8F:
		reg.a = gb.execAdc(reg.a, reg.a)

	case 0x90:
		reg.a = gb.execSub(reg.a, reg.b)
	case 0x91:
		reg.a = gb.execSub(reg.a, reg.c)
	case 0x92:
		reg.a = gb.execSub(reg.a, reg.d)
	case 0x93:
		reg.a = gb.execSub(reg.a, reg.e)
	case 0x94:
		reg.a = gb.execSub(reg.a, reg.h)
	case 0x95:
		reg.a = gb.execSub(reg.a, reg.l)
	case 0x96:
		gb.Tick(4)
		reg.a = gb.execSub(reg.a, gb.Read(*reg.HL))
	case 0x97:
		reg.a = gb.execSub(reg.a, reg.a)

	case 0x98:
		reg.a = gb.execSbc(reg.a, reg.b)
	case 0x99:
		reg.a = gb.execSbc(reg.a, reg.c)
	case 0x9A:
		reg.a = gb.execSbc(reg.a, reg.d)
	case 0x9B:
		reg.a = gb.execSbc(reg.a, reg.e)
	case 0x9C:
		reg.a = gb.execSbc(reg.a, reg.h)
	case 0x9D:
		reg.a = gb.execSbc(reg.a, reg.l)
	case 0x9E:
		gb.Tick(4)
		reg.a = gb.execSbc(reg.a, gb.Read(*reg.HL))
	case 0x9F:
		reg.a = gb.execSbc(reg.a, reg.a)

	case 0xA0:
		reg.a = gb.execAnd(reg.a, reg.b)
	case 0xA1:
		reg.a = gb.execAnd(reg.a, reg.c)
	case 0xA2:
		reg.a = gb.execAnd(reg.a, reg.d)
	case 0xA3:
		reg.a = gb.execAnd(reg.a, reg.e)
	case 0xA4:
		reg.a = gb.execAnd(reg.a, reg.h)
	case 0xA5:
		reg.a = gb.execAnd(reg.a, reg.l)
	case 0xA6:
		gb.Tick(4)
		reg.a = gb.execAnd(reg.a, gb.Read(*reg.HL))
	case 0xA7:
		reg.a = gb.execAnd(reg.a, reg.a)

	case 0xA8:
		reg.a = gb.execXor(reg.a, reg.b)
	case 0xA9:
		reg.a = gb.execXor(reg.a, reg.c)
	case 0xAA:
		reg.a = gb.execXor(reg.a, reg.d)
	case 0xAB:
		reg.a = gb.execXor(reg.a, reg.e)
	case 0xAC:
		reg.a = gb.execXor(reg.a, reg.h)
	case 0xAD:
		reg.a = gb.execXor(reg.a, reg.l)
	case 0xAE:
		gb.Tick(4)
		reg.a = gb.execXor(reg.a, gb.Read(*reg.HL))
	case 0xAF:
		reg.a = gb.execXor(reg.a, reg.a)

	case 0xB0:
		reg.a = gb.execOr(reg.a, reg.b)
	case 0xB1:
		reg.a = gb.execOr(reg.a, reg.c)
	case 0xB2:
		reg.a = gb.execOr(reg.a, reg.d)
	case 0xB3:
		reg.a = gb.execOr(reg.a, reg.e)
	case 0xB4:
		reg.a = gb.execOr(reg.a, reg.h)
	case 0xB5:
		reg.a = gb.execOr(reg.a, reg.l)
	case 0xB6:
		gb.Tick(4)
		reg.a = gb.execOr(reg.a, gb.Read(*reg.HL))
	case 0xB7:
		reg.a = gb.execOr(reg.a, reg.a)

	case 0xB8:
		gb.execCp(reg.a, reg.b)
	case 0xB9:
		gb.execCp(reg.a, reg.c)
	case 0xBA:
		gb.execCp(reg.a, reg.d)
	case 0xBB:
		gb.execCp(reg.a, reg.e)
	case 0xBC:
		gb.execCp(reg.a, reg.h)
	case 0xBD:
		gb.execCp(reg.a, reg.l)
	case 0xBE:
		gb.Tick(4)
		gb.execCp(reg.a, gb.Read(*reg.HL))
	case 0xBF:
		gb.execCp(reg.a, reg.a)

	case 0x04:
		reg.b = gb.execInc(reg.b)
	case 0x14:
		reg.d = gb.execInc(reg.d)
	case 0x24:
		reg.h = gb.execInc(reg.h)
	case 0x34:
		gb.Tick(4)
		v := gb.execInc(gb.Read(*reg.HL))
		gb.Tick(4)
		gb.Write(*reg.HL, v)
	case 0x0C:
		reg.c = gb.execInc(reg.c)
	case 0x1C:
		reg.e = gb.execInc(reg.e)
	case 0x2C:
		reg.l = gb.execInc(reg.l)
	case 0x3C:
		reg.a = gb.execInc(reg.a)
	case 0x03:
		gb.Tick(4)
		*reg.BC++
	case 0x13:
		gb.Tick(4)
		*reg.DE++
	case 0x23:
		gb.Tick(4)
		*reg.HL++
	case 0x33:
		gb.Tick(4)
		gb.Cpu.SP += 1

	case 0x05:
		reg.b = gb.execDec(reg.b)
	case 0x15:
		reg.d = gb.execDec(reg.d)
	case 0x25:
		reg.h = gb.execDec(reg.h)
	case 0x35:
		gb.Tick(4)
		v := gb.execDec(gb.Read(*reg.HL))
		gb.Tick(4)
		gb.Write(*reg.HL, v)
	case 0x0D:
		reg.c = gb.execDec(reg.c)
	case 0x1D:
		reg.e = gb.execDec(reg.e)
	case 0x2D:
		reg.l = gb.execDec(reg.l)
	case 0x3D:
		reg.a = gb.execDec(reg.a)
	case 0x0B:
		gb.Tick(4)
		*reg.BC--
	case 0x1B:
		gb.Tick(4)
		*reg.DE--
	case 0x2B:
		gb.Tick(4)
		*reg.HL--
	case 0x3B:
		gb.Tick(4)
		gb.Cpu.SP -= 1

	case 0xC6:
		gb.Tick(4)
		reg.a = gb.execAdd(reg.a, gb.getImm8())
		pc++
		reg.PcOff++
	case 0xD6:
		gb.Tick(4)
		reg.a = gb.execSub(reg.a, gb.getImm8())
		pc++
		reg.PcOff++
	case 0xE6:
		gb.Tick(4)
		reg.a = gb.execAnd(reg.a, gb.getImm8())
		pc++
		reg.PcOff++
	case 0xF6:
		gb.Tick(4)
		reg.a = gb.execOr(reg.a, gb.getImm8())
		pc++
		reg.PcOff++
	case 0xCE:
		gb.Tick(4)
		reg.a = gb.execAdc(reg.a, gb.getImm8())
		pc++
		reg.PcOff++
	case 0xDE:
		gb.Tick(4)
		reg.a = gb.execSbc(reg.a, gb.getImm8())
		pc++
		reg.PcOff++
	case 0xEE:
		gb.Tick(4)
		reg.a = gb.execXor(reg.a, gb.getImm8())
		pc++
		reg.PcOff++
	case 0xFE:
		gb.Tick(4)
		gb.execCp(reg.a, gb.getImm8())
		pc++
		reg.PcOff++
	case 0x06:
		gb.Tick(4)
		reg.b = gb.getImm8()
		pc++
		reg.PcOff++
	case 0x16:
		gb.Tick(4)
		reg.d = gb.getImm8()
		pc++
		reg.PcOff++
	case 0x26:
		gb.Tick(4)
		reg.h = gb.getImm8()
		pc++
		reg.PcOff++
	case 0x36:
		gb.Tick(8)
		gb.Write(*reg.HL, gb.getImm8())
		pc++
		reg.PcOff++

	case 0x0A:
		gb.Tick(4)
		reg.a = gb.Read(*reg.BC)
	case 0x1A:
		gb.Tick(4)
		reg.a = gb.Read(*reg.DE)
	case 0x2A:
		gb.Tick(4)
		reg.a = gb.Read(*reg.HL)
		*reg.HL++
	case 0x3A:
		gb.Tick(4)
		reg.a = gb.Read(*reg.HL)
		*reg.HL--

	case 0x02:
		gb.Tick(4)
		gb.Write(*reg.BC, reg.a)
	case 0x12:
		gb.Tick(4)
		gb.Write(*reg.DE, reg.a)
	case 0x22:
		gb.Tick(4)
		gb.Write(*reg.HL, reg.a)
		*reg.HL++

	case 0x32:
		gb.Tick(4)
		gb.Write(*reg.HL, reg.a)
		*reg.HL--

	//other misc arth
	case 0x27:
		gb.execDAA()
	case 0x37:
		gb.execSCF()
	case 0x2F:
		gb.execCPL()
	case 0x3F:
		gb.execCCF()

	// Register A Rotations
	case 0x07:
		reg.a = gb.execRot(reg.a, acc, !right, !throughCarry)
	case 0x17:
		reg.a = gb.execRot(reg.a, acc, !right, throughCarry)
	case 0x0F:
		reg.a = gb.execRot(reg.a, acc, right, !throughCarry)
	case 0x1F:
		reg.a = gb.execRot(reg.a, acc, right, throughCarry)

	// CB
	case 0xCB:
		// 8 ticks 1 op 1 cb
		gb.Tick(4)
		gb.execCB(gb.getImm8())
		pc++

	// jump abs
	case 0xC2:
		cycles, pc = gb.execJP(gb.getImm16(), !reg.f.Z, 4, 3)
		gb.Tick((cycles - 1) * 4)
	case 0xD2:
		cycles, pc = gb.execJP(gb.getImm16(), !reg.f.C, 4, 3)
		gb.Tick((cycles - 1) * 4)
	case 0xC3:
		cycles, pc = gb.execJP(gb.getImm16(), true, 4, 4)
		gb.Tick((cycles - 1) * 4)
	case 0xCA:
		cycles, pc = gb.execJP(gb.getImm16(), reg.f.Z, 4, 3)
		gb.Tick((cycles - 1) * 4)
	case 0xDA:
		cycles, pc = gb.execJP(gb.getImm16(), reg.f.C, 4, 3)
		gb.Tick((cycles - 1) * 4)
	case 0xE9:
		cycles, pc = gb.execJP(*reg.HL, true, 1, 1)
		gb.Tick((cycles - 1) * 4)

	// jump relative
	case 0x20:
		cycles, pc = gb.execJR(gb.getImm8(), !reg.f.Z, 3, 2)
		gb.Tick((cycles - 1) * 4)
	case 0x30:
		cycles, pc = gb.execJR(gb.getImm8(), !reg.f.C, 3, 2)
		gb.Tick((cycles - 1) * 4)
	case 0x18:
		cycles, pc = gb.execJR(gb.getImm8(), true, 3, 3)
		gb.Tick((cycles - 1) * 4)
	case 0x28:
		cycles, pc = gb.execJR(gb.getImm8(), reg.f.Z, 3, 2)
		gb.Tick((cycles - 1) * 4)
	case 0x38:
		cycles, pc = gb.execJR(gb.getImm8(), reg.f.C, 3, 2)
		gb.Tick((cycles - 1) * 4)

	// Interrupts
	case 0xF3:
		gb.Cpu.IME = false
	case 0xFB:
		gb.Cpu.PendingInterrupt = true

	// misc ld
	case 0x08:
		gb.Tick(4 * 4)
		gb.Write(gb.getImm16()+0, uint8(gb.Cpu.SP))
		gb.Write(gb.getImm16()+1, uint8(gb.Cpu.SP>>8))
		pc = pc + 2
		reg.PcOff += 2
	case 0xFA:
		gb.Tick(4 * 3)
		reg.a = gb.Read(gb.getImm16())
		pc = pc + 2
		reg.PcOff += 2
	case 0xEA:
		gb.Tick(4 * 3)
		gb.Write(gb.getImm16(), reg.a)
		pc = pc + 2
		reg.PcOff += 2
	case 0xF9:
		gb.Tick(4)
		reg.SP = *reg.HL

	// push
	case 0xC5:
		gb.Tick(4 * 3)
		gb.StackPush(*reg.BC)
	case 0xD5:
		gb.Tick(4 * 3)
		gb.StackPush(*reg.DE)
	case 0xE5:
		gb.Tick(4 * 3)
		gb.StackPush(*reg.HL)
	case 0xF5:
		gb.Tick(4 * 3)
		gb.StackPush(uint16(reg.a)<<8 | uint16(reg.f.Get()))

	// pop
	case 0xC1:
		gb.Tick(8)
		*reg.BC = gb.StackPop()
	case 0xD1:
		gb.Tick(8)
		*reg.DE = gb.StackPop()
	case 0xE1:
		gb.Tick(8)
		*reg.HL = gb.StackPop()
	case 0xF1:
		gb.Tick(8)
		v := gb.StackPop() & 0xFFF0
		reg.a = uint8(v >> 8)
		reg.f.Set(uint8(v))

		// rst
	case 0xC7:
		gb.Tick(3 * 4)
		gb.StackPush(gb.Cpu.PC + 1)
		pc = 0x00
	case 0xD7:
		gb.Tick(3 * 4)
		gb.StackPush(gb.Cpu.PC + 1)
		pc = 0x10
	case 0xE7:
		gb.Tick(3 * 4)
		gb.StackPush(gb.Cpu.PC + 1)
		pc = 0x20
	case 0xF7:
		gb.Tick(3 * 4)
		gb.StackPush(gb.Cpu.PC + 1)
		pc = 0x30
	case 0xCF:
		gb.Tick(3 * 4)
		gb.StackPush(gb.Cpu.PC + 1)
		pc = 0x08
	case 0xDF:
		gb.Tick(3 * 4)
		gb.StackPush(gb.Cpu.PC + 1)
		pc = 0x18
	case 0xEF:
		gb.Tick(3 * 4)
		gb.StackPush(gb.Cpu.PC + 1)
		pc = 0x28
	case 0xFF:
		gb.Tick(3 * 4)
		gb.StackPush(gb.Cpu.PC + 1)
		pc = 0x38

	// call
	case 0xC4:
		cycles, pc = gb.execCall(gb.getImm16(), !reg.f.Z, 6, 3)
		gb.Tick((cycles - 1) * 4)
	case 0xD4:
		cycles, pc = gb.execCall(gb.getImm16(), !reg.f.C, 6, 3)
		gb.Tick((cycles - 1) * 4)
	case 0xCC:
		cycles, pc = gb.execCall(gb.getImm16(), reg.f.Z, 6, 3)
		gb.Tick((cycles - 1) * 4)
	case 0xDC:
		cycles, pc = gb.execCall(gb.getImm16(), reg.f.C, 6, 3)
		gb.Tick((cycles - 1) * 4)
	case 0xCD:
		cycles, pc = gb.execCall(gb.getImm16(), true, 6, 6)
		gb.Tick((cycles - 1) * 4)

	// ret
	case 0xC0:
		cycles, pc = gb.execRet(!reg.f.Z, 5, 2)
		gb.Tick((cycles - 1) * 4)
	case 0xD0:
		cycles, pc = gb.execRet(!reg.f.C, 5, 2)
		gb.Tick((cycles - 1) * 4)
	case 0xC8:
		cycles, pc = gb.execRet(reg.f.Z, 5, 2)
		gb.Tick((cycles - 1) * 4)
	case 0xD8:
		cycles, pc = gb.execRet(reg.f.C, 5, 2)
		gb.Tick((cycles - 1) * 4)
	case 0xC9:
		cycles, pc = gb.execRet(true, 4, 4)
		gb.Tick((cycles - 1) * 4)
	case 0xD9:
		gb.Tick(3 * 4)
		pc = gb.StackPop()
		gb.Cpu.IME = true
		//gb.Cpu.PendingInterrupt = true

	case 0xE0:
		gb.Tick(2 * 4)
		gb.Write(0xFF00+uint16(gb.getImm8()), reg.a)
		pc++
		reg.PcOff++
	case 0xF0:
		gb.Tick(8)
		reg.a = gb.Read(0xFF00 | uint16(gb.getImm8()))
		pc++
		reg.PcOff++

	case 0xE8:
		gb.Tick(3 * 4)
		gb.Cpu.SP = gb.execAddSp(gb.Cpu.SP, uint16(gb.getImm8()))
		pc++
		reg.PcOff++
	case 0xF8:

		gb.Tick(8)
		a := int32(gb.Cpu.SP)
		b := int32(int8(gb.getImm8()))
		newValue := a + b
		temp := a ^ b ^ newValue
		*gb.Cpu.HL = uint16(newValue)

		gb.Cpu.f.Z = false
		gb.Cpu.f.S = false
		gb.Cpu.f.H = (temp & 0x10) != 0
		gb.Cpu.f.C = (temp & 0x100) != 0

		pc++
		reg.PcOff++
	case 0xE2:
		gb.Tick(4)
		gb.Write(0xFF00+uint16(reg.c), reg.a)
	case 0xF2:
		gb.Tick(4)
		reg.a = gb.Read(0xFF00 | uint16(reg.c))

	// empty opcode
	case 0xD3, 0xE3, 0xE4, 0xF4, 0xDB, 0xEB, 0xEC, 0xFC, 0xDD, 0xED, 0xFD:
		panic(fmt.Sprintf("EMPTY OPCODE INSTRUCTION HIT %X", op))
	}

	gb.Cpu.PC = pc
}

func (gb *GameBoy) execCB(op uint8) {

	reg := gb.Cpu
	//cycles = 2

	switch op {
	case 0x00:
		reg.b = gb.execRot(reg.b, !acc, !right, !throughCarry)
	case 0x01:
		reg.c = gb.execRot(reg.c, !acc, !right, !throughCarry)
	case 0x02:
		reg.d = gb.execRot(reg.d, !acc, !right, !throughCarry)
	case 0x03:
		reg.e = gb.execRot(reg.e, !acc, !right, !throughCarry)
	case 0x04:
		reg.h = gb.execRot(reg.h, !acc, !right, !throughCarry)
	case 0x05:
		reg.l = gb.execRot(reg.l, !acc, !right, !throughCarry)
	case 0x06:
		gb.Tick(4)
		v := gb.execRot(gb.Read(*reg.HL), !acc, !right, !throughCarry)
		gb.Tick(4)
		gb.Write(*reg.HL, v)
	case 0x07:
		reg.a = gb.execRot(reg.a, !acc, !right, !throughCarry)

	case 0x10:
		reg.b = gb.execRot(reg.b, !acc, !right, throughCarry)
	case 0x11:
		reg.c = gb.execRot(reg.c, !acc, !right, throughCarry)
	case 0x12:
		reg.d = gb.execRot(reg.d, !acc, !right, throughCarry)
	case 0x13:
		reg.e = gb.execRot(reg.e, !acc, !right, throughCarry)
	case 0x14:
		reg.h = gb.execRot(reg.h, !acc, !right, throughCarry)
	case 0x15:
		reg.l = gb.execRot(reg.l, !acc, !right, throughCarry)
	case 0x16:
		gb.Tick(4)
		v := gb.execRot(gb.Read(*reg.HL), !acc, !right, throughCarry)
		gb.Tick(4)
		gb.Write(*reg.HL, v)

	case 0x17:
		reg.a = gb.execRot(reg.a, !acc, !right, throughCarry)

	case 0x08:
		reg.b = gb.execRot(reg.b, !acc, right, !throughCarry)
	case 0x09:
		reg.c = gb.execRot(reg.c, !acc, right, !throughCarry)
	case 0x0A:
		reg.d = gb.execRot(reg.d, !acc, right, !throughCarry)
	case 0x0B:
		reg.e = gb.execRot(reg.e, !acc, right, !throughCarry)
	case 0x0C:
		reg.h = gb.execRot(reg.h, !acc, right, !throughCarry)
	case 0x0D:
		reg.l = gb.execRot(reg.l, !acc, right, !throughCarry)
	case 0x0E:
		gb.Tick(4)
		v := gb.execRot(gb.Read(*reg.HL), !acc, right, !throughCarry)
		gb.Tick(4)
		gb.Write(*reg.HL, v)

	case 0x0F:
		reg.a = gb.execRot(reg.a, !acc, right, !throughCarry)

	case 0x18:
		reg.b = gb.execRot(reg.b, !acc, right, throughCarry)
	case 0x19:
		reg.c = gb.execRot(reg.c, !acc, right, throughCarry)
	case 0x1A:
		reg.d = gb.execRot(reg.d, !acc, right, throughCarry)
	case 0x1B:
		reg.e = gb.execRot(reg.e, !acc, right, throughCarry)
	case 0x1C:
		reg.h = gb.execRot(reg.h, !acc, right, throughCarry)
	case 0x1D:
		reg.l = gb.execRot(reg.l, !acc, right, throughCarry)
	case 0x1E:
		gb.Tick(4)
		v := gb.execRot(gb.Read(*reg.HL), !acc, right, throughCarry)
		gb.Tick(4)
		gb.Write(*reg.HL, v)

	case 0x1F:
		reg.a = gb.execRot(reg.a, !acc, right, throughCarry)

	case 0x20:
		reg.b = gb.execSLA(reg.b)
	case 0x21:
		reg.c = gb.execSLA(reg.c)
	case 0x22:
		reg.d = gb.execSLA(reg.d)
	case 0x23:
		reg.e = gb.execSLA(reg.e)
	case 0x24:
		reg.h = gb.execSLA(reg.h)
	case 0x25:
		reg.l = gb.execSLA(reg.l)
	case 0x26:
		gb.Tick(4)
		v := gb.execSLA(gb.Read(*reg.HL))
		gb.Tick(4)
		gb.Write(*reg.HL, v)

	case 0x27:
		reg.a = gb.execSLA(reg.a)

	case 0x28:
		reg.b = gb.execSRA(reg.b)
	case 0x29:
		reg.c = gb.execSRA(reg.c)
	case 0x2A:
		reg.d = gb.execSRA(reg.d)
	case 0x2B:
		reg.e = gb.execSRA(reg.e)
	case 0x2C:
		reg.h = gb.execSRA(reg.h)
	case 0x2D:
		reg.l = gb.execSRA(reg.l)
	case 0x2E:
		gb.Tick(4)
		v := gb.execSRA(gb.Read(*reg.HL))
		gb.Tick(4)
		gb.Write(*reg.HL, v)

	case 0x2F:
		reg.a = gb.execSRA(reg.a)

	case 0x30:
		reg.b = gb.execSWAP(reg.b)
	case 0x31:
		reg.c = gb.execSWAP(reg.c)
	case 0x32:
		reg.d = gb.execSWAP(reg.d)
	case 0x33:
		reg.e = gb.execSWAP(reg.e)
	case 0x34:
		reg.h = gb.execSWAP(reg.h)
	case 0x35:
		reg.l = gb.execSWAP(reg.l)
	case 0x36:
		gb.Tick(4)
		v := gb.execSWAP(gb.Read(*reg.HL))
		gb.Tick(4)
		gb.Write(*reg.HL, v)

	case 0x37:
		reg.a = gb.execSWAP(reg.a)

	case 0x38:
		reg.b = gb.execSRL(reg.b)
	case 0x39:
		reg.c = gb.execSRL(reg.c)
	case 0x3A:
		reg.d = gb.execSRL(reg.d)
	case 0x3B:
		reg.e = gb.execSRL(reg.e)
	case 0x3C:
		reg.h = gb.execSRL(reg.h)
	case 0x3D:
		reg.l = gb.execSRL(reg.l)
	case 0x3E:
		gb.Tick(4)
		v := gb.execSRL(gb.Read(*reg.HL))
		gb.Tick(4)
		gb.Write(*reg.HL, v)

	case 0x3F:
		reg.a = gb.execSRL(reg.a)

	case 0x40:
		gb.execBIT(reg.b, 0)
	case 0x41:
		gb.execBIT(reg.c, 0)
	case 0x42:
		gb.execBIT(reg.d, 0)
	case 0x43:
		gb.execBIT(reg.e, 0)
	case 0x44:
		gb.execBIT(reg.h, 0)
	case 0x45:
		gb.execBIT(reg.l, 0)
	case 0x46:
		gb.Tick(4)
		gb.execBIT(gb.Read(*reg.HL), 0)
	case 0x47:
		gb.execBIT(reg.a, 0)

	case 0x48:
		gb.execBIT(reg.b, 1)
	case 0x49:
		gb.execBIT(reg.c, 1)
	case 0x4A:
		gb.execBIT(reg.d, 1)
	case 0x4B:
		gb.execBIT(reg.e, 1)
	case 0x4C:
		gb.execBIT(reg.h, 1)
	case 0x4D:
		gb.execBIT(reg.l, 1)
	case 0x4E:
		gb.Tick(4)
		gb.execBIT(gb.Read(*reg.HL), 1)
	case 0x4F:
		gb.execBIT(reg.a, 1)

	case 0x50:
		gb.execBIT(reg.b, 2)
	case 0x51:
		gb.execBIT(reg.c, 2)
	case 0x52:
		gb.execBIT(reg.d, 2)
	case 0x53:
		gb.execBIT(reg.e, 2)
	case 0x54:
		gb.execBIT(reg.h, 2)
	case 0x55:
		gb.execBIT(reg.l, 2)
	case 0x56:
		gb.Tick(4 * 1)
		gb.execBIT(gb.Read(*reg.HL), 2)
	case 0x57:
		gb.execBIT(reg.a, 2)

	case 0x58:
		gb.execBIT(reg.b, 3)
	case 0x59:
		gb.execBIT(reg.c, 3)
	case 0x5A:
		gb.execBIT(reg.d, 3)
	case 0x5B:
		gb.execBIT(reg.e, 3)
	case 0x5C:
		gb.execBIT(reg.h, 3)
	case 0x5D:
		gb.execBIT(reg.l, 3)
	case 0x5E:
		gb.Tick(4 * 1)
		gb.execBIT(gb.Read(*reg.HL), 3)
	case 0x5F:
		gb.execBIT(reg.a, 3)

	case 0x60:
		gb.execBIT(reg.b, 4)
	case 0x61:
		gb.execBIT(reg.c, 4)
	case 0x62:
		gb.execBIT(reg.d, 4)
	case 0x63:
		gb.execBIT(reg.e, 4)
	case 0x64:
		gb.execBIT(reg.h, 4)
	case 0x65:
		gb.execBIT(reg.l, 4)
	case 0x66:
		gb.Tick(4 * 1)
		gb.execBIT(gb.Read(*reg.HL), 4)
	case 0x67:
		gb.execBIT(reg.a, 4)

	case 0x68:
		gb.execBIT(reg.b, 5)
	case 0x69:
		gb.execBIT(reg.c, 5)
	case 0x6A:
		gb.execBIT(reg.d, 5)
	case 0x6B:
		gb.execBIT(reg.e, 5)
	case 0x6C:
		gb.execBIT(reg.h, 5)
	case 0x6D:
		gb.execBIT(reg.l, 5)
	case 0x6E:
		gb.Tick(4 * 1)
		gb.execBIT(gb.Read(*reg.HL), 5)
	case 0x6F:
		gb.execBIT(reg.a, 5)

	case 0x70:
		gb.execBIT(reg.b, 6)
	case 0x71:
		gb.execBIT(reg.c, 6)
	case 0x72:
		gb.execBIT(reg.d, 6)
	case 0x73:
		gb.execBIT(reg.e, 6)
	case 0x74:
		gb.execBIT(reg.h, 6)
	case 0x75:
		gb.execBIT(reg.l, 6)
	case 0x76:
		gb.Tick(4 * 1)
		gb.execBIT(gb.Read(*reg.HL), 6)
	case 0x77:
		gb.execBIT(reg.a, 6)

	case 0x78:
		gb.execBIT(reg.b, 7)
	case 0x79:
		gb.execBIT(reg.c, 7)
	case 0x7A:
		gb.execBIT(reg.d, 7)
	case 0x7B:
		gb.execBIT(reg.e, 7)
	case 0x7C:
		gb.execBIT(reg.h, 7)
	case 0x7D:
		gb.execBIT(reg.l, 7)
	case 0x7E:
		gb.Tick(4 * 1)
		gb.execBIT(gb.Read(*reg.HL), 7)
	case 0x7F:
		gb.execBIT(reg.a, 7)

	case 0x80:
		reg.b = gb.execRES(reg.b, 0)
	case 0x81:
		reg.c = gb.execRES(reg.c, 0)
	case 0x82:
		reg.d = gb.execRES(reg.d, 0)
	case 0x83:
		reg.e = gb.execRES(reg.e, 0)
	case 0x84:
		reg.h = gb.execRES(reg.h, 0)
	case 0x85:
		reg.l = gb.execRES(reg.l, 0)
	case 0x86:
		gb.Tick(4)
		v := gb.execRES(gb.Read(*reg.HL), 0)
		gb.Tick(4)
		gb.Write(*reg.HL, v)

	case 0x87:
		reg.a = gb.execRES(reg.a, 0)

	case 0x88:
		reg.b = gb.execRES(reg.b, 1)
	case 0x89:
		reg.c = gb.execRES(reg.c, 1)
	case 0x8A:
		reg.d = gb.execRES(reg.d, 1)
	case 0x8B:
		reg.e = gb.execRES(reg.e, 1)
	case 0x8C:
		reg.h = gb.execRES(reg.h, 1)
	case 0x8D:
		reg.l = gb.execRES(reg.l, 1)
	case 0x8E:
		gb.Tick(4)
		v := gb.execRES(gb.Read(*reg.HL), 1)
		gb.Tick(4)
		gb.Write(*reg.HL, v)
	case 0x8F:
		reg.a = gb.execRES(reg.a, 1)

	case 0x90:
		reg.b = gb.execRES(reg.b, 2)
	case 0x91:
		reg.c = gb.execRES(reg.c, 2)
	case 0x92:
		reg.d = gb.execRES(reg.d, 2)
	case 0x93:
		reg.e = gb.execRES(reg.e, 2)
	case 0x94:
		reg.h = gb.execRES(reg.h, 2)
	case 0x95:
		reg.l = gb.execRES(reg.l, 2)
	case 0x96:
		gb.Tick(4)
		v := gb.execRES(gb.Read(*reg.HL), 2)
		gb.Tick(4)
		gb.Write(*reg.HL, v)
	case 0x97:
		reg.a = gb.execRES(reg.a, 2)

	case 0x98:
		reg.b = gb.execRES(reg.b, 3)
	case 0x99:
		reg.c = gb.execRES(reg.c, 3)
	case 0x9A:
		reg.d = gb.execRES(reg.d, 3)
	case 0x9B:
		reg.e = gb.execRES(reg.e, 3)
	case 0x9C:
		reg.h = gb.execRES(reg.h, 3)
	case 0x9D:
		reg.l = gb.execRES(reg.l, 3)
	case 0x9E:
		gb.Tick(4)
		v := gb.execRES(gb.Read(*reg.HL), 3)
		gb.Tick(4)
		gb.Write(*reg.HL, v)
	case 0x9F:
		reg.a = gb.execRES(reg.a, 3)

	case 0xA0:
		reg.b = gb.execRES(reg.b, 4)
	case 0xA1:
		reg.c = gb.execRES(reg.c, 4)
	case 0xA2:
		reg.d = gb.execRES(reg.d, 4)
	case 0xA3:
		reg.e = gb.execRES(reg.e, 4)
	case 0xA4:
		reg.h = gb.execRES(reg.h, 4)
	case 0xA5:
		reg.l = gb.execRES(reg.l, 4)
	case 0xA6:
		gb.Tick(4)
		v := gb.execRES(gb.Read(*reg.HL), 4)
		gb.Tick(4)
		gb.Write(*reg.HL, v)
	case 0xA7:
		reg.a = gb.execRES(reg.a, 4)

	case 0xA8:
		reg.b = gb.execRES(reg.b, 5)
	case 0xA9:
		reg.c = gb.execRES(reg.c, 5)
	case 0xAA:
		reg.d = gb.execRES(reg.d, 5)
	case 0xAB:
		reg.e = gb.execRES(reg.e, 5)
	case 0xAC:
		reg.h = gb.execRES(reg.h, 5)
	case 0xAD:
		reg.l = gb.execRES(reg.l, 5)
	case 0xAE:
		gb.Tick(4)
		v := gb.execRES(gb.Read(*reg.HL), 5)
		gb.Tick(4)
		gb.Write(*reg.HL, v)
	case 0xAF:
		reg.a = gb.execRES(reg.a, 5)

	case 0xB0:
		reg.b = gb.execRES(reg.b, 6)
	case 0xB1:
		reg.c = gb.execRES(reg.c, 6)
	case 0xB2:
		reg.d = gb.execRES(reg.d, 6)
	case 0xB3:
		reg.e = gb.execRES(reg.e, 6)
	case 0xB4:
		reg.h = gb.execRES(reg.h, 6)
	case 0xB5:
		reg.l = gb.execRES(reg.l, 6)
	case 0xB6:
		gb.Tick(4)
		v := gb.execRES(gb.Read(*reg.HL), 6)
		gb.Tick(4)
		gb.Write(*reg.HL, v)
	case 0xB7:
		reg.a = gb.execRES(reg.a, 6)

	case 0xB8:
		reg.b = gb.execRES(reg.b, 7)
	case 0xB9:
		reg.c = gb.execRES(reg.c, 7)
	case 0xBA:
		reg.d = gb.execRES(reg.d, 7)
	case 0xBB:
		reg.e = gb.execRES(reg.e, 7)
	case 0xBC:
		reg.h = gb.execRES(reg.h, 7)
	case 0xBD:
		reg.l = gb.execRES(reg.l, 7)
	case 0xBE:
		gb.Tick(4)
		v := gb.execRES(gb.Read(*reg.HL), 7)
		gb.Tick(4)
		gb.Write(*reg.HL, v)
	case 0xBF:
		reg.a = gb.execRES(reg.a, 7)

	case 0xC0:
		reg.b = gb.execSET(reg.b, 0)
	case 0xC1:
		reg.c = gb.execSET(reg.c, 0)
	case 0xC2:
		reg.d = gb.execSET(reg.d, 0)
	case 0xC3:
		reg.e = gb.execSET(reg.e, 0)
	case 0xC4:
		reg.h = gb.execSET(reg.h, 0)
	case 0xC5:
		reg.l = gb.execSET(reg.l, 0)
	case 0xC6:
		gb.Tick(4)
		v := gb.execSET(gb.Read(*reg.HL), 0)
		gb.Tick(4)
		gb.Write(*reg.HL, v)

	case 0xC7:
		reg.a = gb.execSET(reg.a, 0)

	case 0xC8:
		reg.b = gb.execSET(reg.b, 1)
	case 0xC9:
		reg.c = gb.execSET(reg.c, 1)
	case 0xCA:
		reg.d = gb.execSET(reg.d, 1)
	case 0xCB:
		reg.e = gb.execSET(reg.e, 1)
	case 0xCC:
		reg.h = gb.execSET(reg.h, 1)
	case 0xCD:
		reg.l = gb.execSET(reg.l, 1)
	case 0xCE:
		gb.Tick(4)
		v := gb.execSET(gb.Read(*reg.HL), 1)
		gb.Tick(4)
		gb.Write(*reg.HL, v)
	case 0xCF:
		reg.a = gb.execSET(reg.a, 1)

	case 0xD0:
		reg.b = gb.execSET(reg.b, 2)
	case 0xD1:
		reg.c = gb.execSET(reg.c, 2)
	case 0xD2:
		reg.d = gb.execSET(reg.d, 2)
	case 0xD3:
		reg.e = gb.execSET(reg.e, 2)
	case 0xD4:
		reg.h = gb.execSET(reg.h, 2)
	case 0xD5:
		reg.l = gb.execSET(reg.l, 2)
	case 0xD6:
		gb.Tick(4)
		v := gb.execSET(gb.Read(*reg.HL), 2)
		gb.Tick(4)
		gb.Write(*reg.HL, v)
	case 0xD7:
		reg.a = gb.execSET(reg.a, 2)

	case 0xD8:
		reg.b = gb.execSET(reg.b, 3)
	case 0xD9:
		reg.c = gb.execSET(reg.c, 3)
	case 0xDA:
		reg.d = gb.execSET(reg.d, 3)
	case 0xDB:
		reg.e = gb.execSET(reg.e, 3)
	case 0xDC:
		reg.h = gb.execSET(reg.h, 3)
	case 0xDD:
		reg.l = gb.execSET(reg.l, 3)
	case 0xDE:
		gb.Tick(4)
		v := gb.execSET(gb.Read(*reg.HL), 3)
		gb.Tick(4)
		gb.Write(*reg.HL, v)
	case 0xDF:
		reg.a = gb.execSET(reg.a, 3)

	case 0xE0:
		reg.b = gb.execSET(reg.b, 4)
	case 0xE1:
		reg.c = gb.execSET(reg.c, 4)
	case 0xE2:
		reg.d = gb.execSET(reg.d, 4)
	case 0xE3:
		reg.e = gb.execSET(reg.e, 4)
	case 0xE4:
		reg.h = gb.execSET(reg.h, 4)
	case 0xE5:
		reg.l = gb.execSET(reg.l, 4)
	case 0xE6:
		gb.Tick(4)
		v := gb.execSET(gb.Read(*reg.HL), 4)
		gb.Tick(4)
		gb.Write(*reg.HL, v)
	case 0xE7:
		reg.a = gb.execSET(reg.a, 4)

	case 0xE8:
		reg.b = gb.execSET(reg.b, 5)
	case 0xE9:
		reg.c = gb.execSET(reg.c, 5)
	case 0xEA:
		reg.d = gb.execSET(reg.d, 5)
	case 0xEB:
		reg.e = gb.execSET(reg.e, 5)
	case 0xEC:
		reg.h = gb.execSET(reg.h, 5)
	case 0xED:
		reg.l = gb.execSET(reg.l, 5)
	case 0xEE:
		gb.Tick(4)
		v := gb.execSET(gb.Read(*reg.HL), 5)
		gb.Tick(4)
		gb.Write(*reg.HL, v)
	case 0xEF:
		reg.a = gb.execSET(reg.a, 5)

	case 0xF0:
		reg.b = gb.execSET(reg.b, 6)
	case 0xF1:
		reg.c = gb.execSET(reg.c, 6)
	case 0xF2:
		reg.d = gb.execSET(reg.d, 6)
	case 0xF3:
		reg.e = gb.execSET(reg.e, 6)
	case 0xF4:
		reg.h = gb.execSET(reg.h, 6)
	case 0xF5:
		reg.l = gb.execSET(reg.l, 6)
	case 0xF6:
		gb.Tick(4)
		v := gb.execSET(gb.Read(*reg.HL), 6)
		gb.Tick(4)
		gb.Write(*reg.HL, v)
	case 0xF7:
		reg.a = gb.execSET(reg.a, 6)

	case 0xF8:
		reg.b = gb.execSET(reg.b, 7)
	case 0xF9:
		reg.c = gb.execSET(reg.c, 7)
	case 0xFA:
		reg.d = gb.execSET(reg.d, 7)
	case 0xFB:
		reg.e = gb.execSET(reg.e, 7)
	case 0xFC:
		reg.h = gb.execSET(reg.h, 7)
	case 0xFD:
		reg.l = gb.execSET(reg.l, 7)
	case 0xFE:
		gb.Tick(4)
		v := gb.execSET(gb.Read(*reg.HL), 7)
		gb.Tick(4)
		gb.Write(*reg.HL, v)
	case 0xFF:
		reg.a = gb.execSET(reg.a, 7)
	}
}

func (gb *GameBoy) execRot(v uint8, acc, right, throughCarry bool) uint8 {

	reg := gb.Cpu

	switch {
	case right && !throughCarry:

		reg.f.C = v&1 != 0
		v = (v >> 1) | ((v & 1) << 7)

	case right && throughCarry:

		carry := uint8(0)
		if reg.f.C {
			carry = 0x80
		}

		reg.f.C = v&1 != 0

		v = (v >> 1) | carry

	case !right && !throughCarry:
		reg.f.C = v&0x80 != 0
		v = (v << 1) | (v >> 7)

	case !right && throughCarry:

		carry := uint8(0)
		if reg.f.C {
			carry = 1
		}

		reg.f.C = v&0x80 != 0

		v = (v<<1)&0xFF | carry
	}

	reg.f.Z = v == 0 && !acc
	reg.f.S = false
	reg.f.H = false
	return v
}

func (gb *GameBoy) execSLA(v uint8) uint8 {
	gb.Cpu.f.C = v&0x80 != 0
	v <<= 1
	gb.Cpu.f.Z = v == 0
	gb.Cpu.f.S = false
	gb.Cpu.f.H = false
	return v
}

func (gb *GameBoy) execSRA(v uint8) uint8 {
	gb.Cpu.f.C = v&1 != 0
	v = (v & 128) | (v >> 1)
	gb.Cpu.f.Z = v == 0
	gb.Cpu.f.S = false
	gb.Cpu.f.H = false
	return v
}

func (gb *GameBoy) execSRL(v uint8) uint8 {
	gb.Cpu.f.C = v&1 != 0
	v >>= 1
	gb.Cpu.f.Z = v == 0
	gb.Cpu.f.S = false
	gb.Cpu.f.H = false
	return v
}

func (gb *GameBoy) execSWAP(v uint8) uint8 {
	v = uint8((v >> 4) | ((v << 4) & 0xF0))
	gb.Cpu.f.Z = v == 0
	gb.Cpu.f.S = false
	gb.Cpu.f.H = false
	gb.Cpu.f.C = false
	return v
}

func (gb *GameBoy) execBIT(v, bit uint8) {
	gb.Cpu.f.Z = (v>>bit)&1 == 0
	gb.Cpu.f.S = false
	gb.Cpu.f.H = true
}

func (gb *GameBoy) execRES(v, bit uint8) uint8 {
	return v &^ (1 << bit)
}

func (gb *GameBoy) execSET(v, bit uint8) uint8 {
	return v | (1 << bit)
}

func (gb *GameBoy) execDAA() {

	reg := gb.Cpu

	if !reg.f.S {
		if reg.f.C || reg.a > 0x99 {
			reg.a = reg.a + 0x60
			reg.f.C = true
		}

		if reg.f.H || reg.a&0xF > 0x9 {
			reg.a = reg.a + 0x06
			reg.f.H = false
		}

		reg.f.Z = reg.a == 0
		return
	}

	if reg.f.C && reg.f.H {
		reg.a += 0x9A
		reg.f.H = false

		reg.f.Z = reg.a == 0
		return
	}

	if reg.f.C {
		reg.a += 0xA0
		reg.f.Z = reg.a == 0
		return
	}

	if reg.f.H {
		reg.a += 0xFA
		reg.f.H = false
		reg.f.Z = reg.a == 0
		return
	}

	reg.f.Z = reg.a == 0
}

func (gb *GameBoy) execSCF() {
	// set carry flag
	gb.Cpu.f.S = false
	gb.Cpu.f.H = false
	gb.Cpu.f.C = true
}

func (gb *GameBoy) execCCF() {
	// compliment (invert) carry flag
	gb.Cpu.f.S = false
	gb.Cpu.f.H = false
	gb.Cpu.f.C = !gb.Cpu.f.C
}

func (gb *GameBoy) execCPL() {
	gb.Cpu.a = 0xFF ^ gb.Cpu.a
	gb.Cpu.f.S = true
	gb.Cpu.f.H = true
}

func (gb *GameBoy) execAddHl(a, b uint16) uint16 {
	res := a + b
	gb.Cpu.f.H = (a & 0xFFF) > (res & 0xFFF)
	gb.Cpu.f.C = uint(a)+uint(b) > 0xFFFF
	gb.Cpu.f.S = false
	return res
}

func (gb *GameBoy) execAddSp(a, b uint16) uint16 {
	res := uint16(int(a) + int(int8(b)))
	tmp := a ^ uint16(int8(b)) ^ res
	gb.Cpu.f.H = (tmp & 0x10) != 0
	gb.Cpu.f.Z = false
	gb.Cpu.f.C = (tmp & 0x100) != 0
	gb.Cpu.f.S = false
	return res
}

func (gb *GameBoy) execAdd(a, b uint8) uint8 {
	res := a + b
	gb.Cpu.f.H = (a&0xF)+(b&0xF) > 0xF
	gb.Cpu.f.Z = res == 0
	gb.Cpu.f.C = uint(a)+uint(b) > 0xFF
	gb.Cpu.f.S = false
	return res
}

func (gb *GameBoy) execAdc(a, b uint8) uint8 {
	carry := uint8(0)
	if gb.Cpu.f.C {
		carry = 1
	}

	res := a + b + carry
	gb.Cpu.f.H = (a&0xF)+(b&0xF)+carry > 0xF
	gb.Cpu.f.Z = res == 0
	gb.Cpu.f.C = uint16(a)+uint16(b)+uint16(carry) > 0xFF
	gb.Cpu.f.S = false
	return res
}

func (gb *GameBoy) execSub(a, b uint8) uint8 {
	res := a - b
	gb.Cpu.f.Z = res == 0
	gb.Cpu.f.S = true
	gb.Cpu.f.H = a&0xF < b&0xF
	gb.Cpu.f.C = a < b
	return res
}

func (gb *GameBoy) execSbc(a, b uint8) uint8 {
	carry := uint8(0)
	if gb.Cpu.f.C {
		carry = 1
	}

	res := a - b - carry
	gb.Cpu.f.Z = res == 0
	gb.Cpu.f.S = true
	gb.Cpu.f.H = a&0xF < b&0xF+carry
	gb.Cpu.f.C = uint(a) < uint(b)+uint(carry)
	return res
}

func (gb *GameBoy) execAnd(a, b uint8) uint8 {
	gb.Cpu.f.Z = a&b == 0
	gb.Cpu.f.S = false
	gb.Cpu.f.H = true
	gb.Cpu.f.C = false
	return a & b
}

func (gb *GameBoy) execXor(a, b uint8) uint8 {
	gb.Cpu.f.Z = a^b == 0
	gb.Cpu.f.S = false
	gb.Cpu.f.H = false
	gb.Cpu.f.C = false
	return a ^ b
}

func (gb *GameBoy) execOr(a, b uint8) uint8 {
	gb.Cpu.f.Z = a|b == 0
	gb.Cpu.f.S = false
	gb.Cpu.f.H = false
	gb.Cpu.f.C = false
	return a | b
}

func (gb *GameBoy) execCp(a, b uint8) {
	gb.Cpu.f.Z = a-b == 0
	gb.Cpu.f.S = true
	gb.Cpu.f.H = (b & 0xF) > (a & 0xF)
	gb.Cpu.f.C = b > a
}

func (gb *GameBoy) execInc(v uint8) uint8 {
	gb.Cpu.f.H = v&0xF == 0xF
	v++
	gb.Cpu.f.Z = v == 0
	gb.Cpu.f.S = false
	return v
}

func (gb *GameBoy) execDec(v uint8) uint8 {
	gb.Cpu.f.H = v&0xF == 0
	v--
	gb.Cpu.f.Z = v == 0
	gb.Cpu.f.S = true
	return v
}

func (gb *GameBoy) execJP(addr uint16, cond bool, cyclesIf, cycles int) (c int, pc uint16) {

	if cond {
		return cyclesIf, addr
	}

	return cycles, gb.Cpu.PC + 3
}

func (gb *GameBoy) execJR(addr uint8, cond bool, cyclesIf, cycles int) (c int, pc uint16) {

	if cond {
		return cyclesIf, uint16(int32(gb.Cpu.PC)+int32(int8(addr))) + 2
	}

	return cycles, gb.Cpu.PC + 2
}

func (gb *GameBoy) StackPop() uint16 {
	v := uint16(gb.Read(gb.Cpu.SP))
	gb.Cpu.SP++
	v |= uint16(gb.Read(gb.Cpu.SP)) << 8
	gb.Cpu.SP++
	return v
}

func (gb *GameBoy) StackPush(v uint16) {
	gb.Cpu.SP--
	gb.Write(gb.Cpu.SP, uint8(v>>8))
	gb.Cpu.SP--
	gb.Write(gb.Cpu.SP, uint8(v))
}

func (gb *GameBoy) execCall(addr uint16, cond bool, cyclesIf, cycles int) (c int, pc uint16) {
	if cond {
		gb.StackPush(gb.Cpu.PC + 3)
		return cyclesIf, addr
	}

	return cycles, gb.Cpu.PC + 3
}

func (gb *GameBoy) execRet(cond bool, cyclesIf, cycles int) (c int, pc uint16) {

	if cond {
		return cyclesIf, gb.StackPop()
	}

	return cycles, gb.Cpu.PC + 1
}
