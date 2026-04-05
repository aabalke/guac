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

const (
    R8_B = iota
    R8_C
    R8_D
    R8_E
    R8_H
    R8_L
    R8_HL
    R8_A
)

func (gb *GameBoy) Block1(op uint8) {
	reg := gb.Cpu
    var dst *uint8
    var src uint8

    switch (op >> 3) & 7 {
    case R8_B:      dst = &reg.b
    case R8_C:      dst = &reg.c
    case R8_D:      dst = &reg.d
    case R8_E:      dst = &reg.e
    case R8_H:      dst = &reg.h
    case R8_L:      dst = &reg.l
    case R8_HL:     dst = nil
    case R8_A:      dst = &reg.a
    }

    switch op & 7 {
    case R8_B:      src = reg.b
    case R8_C:      src = reg.c
    case R8_D:      src = reg.d
    case R8_E:      src = reg.e
    case R8_H:      src = reg.h
    case R8_L:      src = reg.l
    case R8_HL:     gb.Tick(4); src = gb.Read(*reg.HL)
    case R8_A:      src = reg.a
    }

    if hl := dst == nil; hl {
        if op == 0x76 {
            gb.Cpu.Halted = true
            return
        }

        gb.Tick(4)
        gb.Write(*reg.HL, src)
        return
    }

    *dst = src
}

const (
    ARTH_ADD = iota
    ARTH_ADC
    ARTH_SUB
    ARTH_SBC
    ARTH_AND
    ARTH_XOR
    ARTH_OR
    ARTH_CP
)

func (gb * GameBoy) Block2(op uint8) {

	reg := gb.Cpu
    var src uint8

    switch op & 7 {
    case R8_B:      src = reg.b
    case R8_C:      src = reg.c
    case R8_D:      src = reg.d
    case R8_E:      src = reg.e
    case R8_H:      src = reg.h
    case R8_L:      src = reg.l
    case R8_HL:     gb.Tick(4); src = gb.Read(*reg.HL)
    case R8_A:      src = reg.a
    }

    switch (op >> 3) & 0xF {
    case ARTH_ADD: reg.a = gb.execAdd(reg.a, src)
    case ARTH_ADC: reg.a = gb.execAdc(reg.a, src)
    case ARTH_SUB: reg.a = gb.execSub(reg.a, src)
    case ARTH_SBC: reg.a = gb.execSbc(reg.a, src)
    case ARTH_AND: reg.a = gb.execAnd(reg.a, src)
    case ARTH_XOR: reg.a = gb.execXor(reg.a, src)
    case ARTH_OR:  reg.a = gb.execOr(reg.a, src)
    case ARTH_CP:  gb.execCp(reg.a, src)
    }
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

    if block1 := op & 0xC0 == 0x40; block1 {
		gb.Block1(op)
		reg.PC = pc
        return
	}

    if block2 := op & 0xC0 == 0x80; block2 {
		gb.Block2(op)
		reg.PC = pc
        return
	}

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

const (
    CB_OTR = 0b00
    CB_BIT = 0b01
    CB_RES = 0b10
    CB_SET = 0b11
)

func (gb *GameBoy) execCB(op uint8) {

    reg := gb.Cpu

    var src *uint8

    bit := (op >> 3) & 7

    switch op & 7 {
    case R8_B:      src = &reg.b
    case R8_C:      src = &reg.c
    case R8_D:      src = &reg.d
    case R8_E:      src = &reg.e
    case R8_H:      src = &reg.h
    case R8_L:      src = &reg.l
    case R8_HL:
        gb.Tick(4)
        v := gb.Read(*reg.HL)

        switch op >> 6 {
        case 0:
            switch inst := (op >> 3); inst {
                case 0: v = gb.execRot(v, !acc, !right, !throughCarry)
                case 1: v = gb.execRot(v, !acc, right, !throughCarry)
                case 2: v = gb.execRot(v, !acc, !right, throughCarry)
                case 3: v = gb.execRot(v, !acc, right, throughCarry)
                case 4: v = gb.execSLA(v)
                case 5: v = gb.execSRA(v)
                case 6: v = gb.execSWAP(v)
                case 7: v = gb.execSRL(v)
            }

            gb.Tick(4)
            gb.Write(*reg.HL, v)

        case 1:
            gb.execBIT(v, bit)
        case 2:
            v = gb.execRES(v, bit)
            gb.Tick(4)
            gb.Write(*reg.HL, v)
        case 3:
            v = gb.execSET(v, bit)
            gb.Tick(4)
            gb.Write(*reg.HL, v)
        }
        return

    case R8_A:      src = &reg.a
    }

    switch op >> 6 {
    case 0:
        switch inst := (op >> 3); inst {
        case 0: *src = gb.execRot(*src, !acc, !right, !throughCarry)
        case 1: *src = gb.execRot(*src, !acc, right, !throughCarry)
        case 2: *src = gb.execRot(*src, !acc, !right, throughCarry)
        case 3: *src = gb.execRot(*src, !acc, right, throughCarry)
        case 4: *src = gb.execSLA(*src)
        case 5: *src = gb.execSRA(*src)
        case 6: *src = gb.execSWAP(*src)
        case 7: *src = gb.execSRL(*src)
        }
    case 1:
        gb.execBIT(*src, bit)
    case 2:
        *src = gb.execRES(*src, bit)
    case 3:
        *src = gb.execSET(*src, bit)
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
