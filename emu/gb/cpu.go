package gameboy

import (
	"fmt"
)

// 4 cycles per m clock (if inst/opcode take 4 cycles, inc m by 1, 8 cycles, by 2

type Cpu struct {
	Registers           Registers

    InterruptMaster     bool
    PendingInterrupt    bool
    Halted              bool

	PC                  uint16
	SP                  uint16
}

func NewCpu() *Cpu {
    c := Cpu{
        Registers: Registers{
            a: 0x01,
            b: 0x00,
            c: 0x13,
            d: 0x00,
            e: 0xD8,
            h: 0x01,
            l: 0x4D,
            f: Flags{
                Zero: true,
                Subtraction: false,
                HalfCarry: true,
                Carry: true,
            },
        },
        InterruptMaster: false,
        PendingInterrupt: false,
		PC: 0x100,
        SP: 0xFFFE,
	}

    return &c
}

func (gb *GameBoy) Execute(opcode uint8) (cycles int) {

    reg := &gb.Cpu.Registers

    cycles = 1
    pc := gb.Cpu.PC + 1

    hl := gb.Cpu.Registers.hl()
    d8, d16 := gb.getImmediateData()
    hlValue, _ := gb.ReadByte(hl)

    switch opcode {
    case 0x00: // nop
    case 0x10: // stop / toggle speed

        if gb.Color && gb.PrepareSpeedToggle {
            gb.WriteByte(0xFF26, 0)
            gb.toggleDoubleSpeed()
        } else {
            gb.Cpu.Halted = true
        }

        pc++

    // Load State Move
    case 0x40: gb.execLd(&reg.b, reg.b)
    case 0x41: gb.execLd(&reg.b, reg.c)
    case 0x42: gb.execLd(&reg.b, reg.d)
    case 0x43: gb.execLd(&reg.b, reg.e)
    case 0x44: gb.execLd(&reg.b, reg.h)
    case 0x45: gb.execLd(&reg.b, reg.l)
    case 0x46: gb.execLd(&reg.b, hlValue); cycles = 2
    case 0x47: gb.execLd(&reg.b, reg.a)

    case 0x48: gb.execLd(&reg.c, reg.b)
    case 0x49: gb.execLd(&reg.c, reg.c)
    case 0x4A: gb.execLd(&reg.c, reg.d)
    case 0x4B: gb.execLd(&reg.c, reg.e)
    case 0x4C: gb.execLd(&reg.c, reg.h)
    case 0x4D: gb.execLd(&reg.c, reg.l)
    case 0x4E: gb.execLd(&reg.c, hlValue); cycles = 2
    case 0x4F: gb.execLd(&reg.c, reg.a)

    case 0x50: gb.execLd(&reg.d, reg.b)
    case 0x51: gb.execLd(&reg.d, reg.c)
    case 0x52: gb.execLd(&reg.d, reg.d)
    case 0x53: gb.execLd(&reg.d, reg.e)
    case 0x54: gb.execLd(&reg.d, reg.h)
    case 0x55: gb.execLd(&reg.d, reg.l)
    case 0x56: gb.execLd(&reg.d, hlValue); cycles = 2
    case 0x57: gb.execLd(&reg.d, reg.a)

    case 0x58: gb.execLd(&reg.e, reg.b)
    case 0x59: gb.execLd(&reg.e, reg.c)
    case 0x5A: gb.execLd(&reg.e, reg.d)
    case 0x5B: gb.execLd(&reg.e, reg.e)
    case 0x5C: gb.execLd(&reg.e, reg.h)
    case 0x5D: gb.execLd(&reg.e, reg.l)
    case 0x5E: gb.execLd(&reg.e, hlValue); cycles = 2
    case 0x5F: gb.execLd(&reg.e, reg.a)

    case 0x60: gb.execLd(&reg.h, reg.b)
    case 0x61: gb.execLd(&reg.h, reg.c)
    case 0x62: gb.execLd(&reg.h, reg.d)
    case 0x63: gb.execLd(&reg.h, reg.e)
    case 0x64: gb.execLd(&reg.h, reg.h)
    case 0x65: gb.execLd(&reg.h, reg.l)
    case 0x66: gb.execLd(&reg.h, hlValue); cycles = 2
    case 0x67: gb.execLd(&reg.h, reg.a)

    case 0x68: gb.execLd(&reg.l, reg.b)
    case 0x69: gb.execLd(&reg.l, reg.c)
    case 0x6A: gb.execLd(&reg.l, reg.d)
    case 0x6B: gb.execLd(&reg.l, reg.e)
    case 0x6C: gb.execLd(&reg.l, reg.h)
    case 0x6D: gb.execLd(&reg.l, reg.l)
    case 0x6E: gb.execLd(&reg.l, hlValue); cycles = 2
    case 0x6F: gb.execLd(&reg.l, reg.a)

    case 0x01: gb.execLd(RegisterBc, d16); pc = pc + 2; cycles = 3
    case 0x11: gb.execLd(RegisterDe, d16); pc = pc + 2; cycles = 3
    case 0x21: gb.execLd(RegisterHl, d16); pc = pc + 2; cycles = 3
    case 0x31: gb.execLd(&gb.Cpu.SP, d16); pc = pc + 2; cycles = 3

    case 0x70: gb.execLdMem(hl, reg.b, false); cycles = 2
    case 0x71: gb.execLdMem(hl, reg.c, false); cycles = 2
    case 0x72: gb.execLdMem(hl, reg.d, false); cycles = 2
    case 0x73: gb.execLdMem(hl, reg.e, false); cycles = 2
    case 0x74: gb.execLdMem(hl, reg.h, false); cycles = 2
    case 0x75: gb.execLdMem(hl, reg.l, false); cycles = 2
    case 0x76: gb.Cpu.Halted = true; 
        
    case 0x77: gb.execLdMem(hl, reg.a, false); cycles = 2

    case 0x0E: gb.execLd(&reg.c, d8); pc++; cycles = 2
    case 0x1E: gb.execLd(&reg.e, d8); pc++; cycles = 2
    case 0x2E: gb.execLd(&reg.l, d8); pc++; cycles = 2
    case 0x3E: gb.execLd(&reg.a, d8); pc++; cycles = 2

    case 0x78: gb.execLd(&reg.a, reg.b)
    case 0x79: gb.execLd(&reg.a, reg.c)
    case 0x7A: gb.execLd(&reg.a, reg.d)
    case 0x7B: gb.execLd(&reg.a, reg.e)
    case 0x7C: gb.execLd(&reg.a, reg.h)
    case 0x7D: gb.execLd(&reg.a, reg.l)
    case 0x7E: gb.execLd(&reg.a, hlValue); cycles = 2
    case 0x7F: gb.execLd(&reg.a, reg.a)

    case 0x80: gb.execAdd(&reg.a, reg.b)
    case 0x81: gb.execAdd(&reg.a, reg.c)
    case 0x82: gb.execAdd(&reg.a, reg.d)
    case 0x83: gb.execAdd(&reg.a, reg.e)
    case 0x84: gb.execAdd(&reg.a, reg.h)
    case 0x85: gb.execAdd(&reg.a, reg.l)
    case 0x86: gb.execAdd(&reg.a, hlValue); cycles = 2
    case 0x87: gb.execAdd(&reg.a, reg.a)

    case 0x09: gb.execAdd(RegisterHl, reg.bc()); cycles = 2
    case 0x19: gb.execAdd(RegisterHl, reg.de()); cycles = 2
    case 0x29: gb.execAdd(RegisterHl, reg.hl()); cycles = 2
    case 0x39: gb.execAdd(RegisterHl, gb.Cpu.SP); cycles = 2

    case 0x88: gb.execAdc(&reg.a, reg.b)
    case 0x89: gb.execAdc(&reg.a, reg.c)
    case 0x8A: gb.execAdc(&reg.a, reg.d)
    case 0x8B: gb.execAdc(&reg.a, reg.e)
    case 0x8C: gb.execAdc(&reg.a, reg.h)
    case 0x8D: gb.execAdc(&reg.a, reg.l)
    case 0x8E: gb.execAdc(&reg.a, hlValue); cycles = 2
    case 0x8F: gb.execAdc(&reg.a, reg.a)

    case 0x90: gb.execSub(&reg.a, reg.b)
    case 0x91: gb.execSub(&reg.a, reg.c)
    case 0x92: gb.execSub(&reg.a, reg.d)
    case 0x93: gb.execSub(&reg.a, reg.e)
    case 0x94: gb.execSub(&reg.a, reg.h)
    case 0x95: gb.execSub(&reg.a, reg.l)
    case 0x96: gb.execSub(&reg.a, hlValue); cycles = 2
    case 0x97: gb.execSub(&reg.a, reg.a)

    case 0x98: gb.execSbc(&reg.a, reg.b)
    case 0x99: gb.execSbc(&reg.a, reg.c)
    case 0x9A: gb.execSbc(&reg.a, reg.d)
    case 0x9B: gb.execSbc(&reg.a, reg.e)
    case 0x9C: gb.execSbc(&reg.a, reg.h)
    case 0x9D: gb.execSbc(&reg.a, reg.l)
    case 0x9E: gb.execSbc(&reg.a, hlValue); cycles = 2
    case 0x9F: gb.execSbc(&reg.a, reg.a)

    case 0xA0: gb.execAnd(&reg.a, reg.b)
    case 0xA1: gb.execAnd(&reg.a, reg.c)
    case 0xA2: gb.execAnd(&reg.a, reg.d)
    case 0xA3: gb.execAnd(&reg.a, reg.e)
    case 0xA4: gb.execAnd(&reg.a, reg.h)
    case 0xA5: gb.execAnd(&reg.a, reg.l)
    case 0xA6: gb.execAnd(&reg.a, hlValue); cycles = 2
    case 0xA7: gb.execAnd(&reg.a, reg.a)

    case 0xA8: gb.execXor(&reg.a, reg.b)
    case 0xA9: gb.execXor(&reg.a, reg.c)
    case 0xAA: gb.execXor(&reg.a, reg.d)
    case 0xAB: gb.execXor(&reg.a, reg.e)
    case 0xAC: gb.execXor(&reg.a, reg.h)
    case 0xAD: gb.execXor(&reg.a, reg.l)
    case 0xAE: gb.execXor(&reg.a, hlValue); cycles = 2
    case 0xAF: gb.execXor(&reg.a, reg.a)

    case 0xB0: gb.execOr(&reg.a, reg.b)
    case 0xB1: gb.execOr(&reg.a, reg.c)
    case 0xB2: gb.execOr(&reg.a, reg.d)
    case 0xB3: gb.execOr(&reg.a, reg.e)
    case 0xB4: gb.execOr(&reg.a, reg.h)
    case 0xB5: gb.execOr(&reg.a, reg.l)
    case 0xB6: gb.execOr(&reg.a, hlValue); cycles = 2
    case 0xB7: gb.execOr(&reg.a, reg.a)

    case 0xB8: gb.execCp(&reg.a, reg.b)
    case 0xB9: gb.execCp(&reg.a, reg.c)
    case 0xBA: gb.execCp(&reg.a, reg.d)
    case 0xBB: gb.execCp(&reg.a, reg.e)
    case 0xBC: gb.execCp(&reg.a, reg.h)
    case 0xBD: gb.execCp(&reg.a, reg.l)
    case 0xBE: gb.execCp(&reg.a, hlValue); cycles = 2
    case 0xBF: gb.execCp(&reg.a, reg.a)

    case 0x04: gb.execInc(&reg.b)
    case 0x14: gb.execInc(&reg.d)
    case 0x24: gb.execInc(&reg.h)
    case 0x34: gb.execIncDecHl(RegisterHl, false); cycles = 3
    case 0x0C: gb.execInc(&reg.c)
    case 0x1C: gb.execInc(&reg.e)
    case 0x2C: gb.execInc(&reg.l)
    case 0x3C: gb.execInc(&reg.a)
    case 0x03: gb.execInc(RegisterBc); cycles = 2
    case 0x13: gb.execInc(RegisterDe); cycles = 2
    case 0x23: gb.execInc(RegisterHl); cycles = 2
    case 0x33: gb.Cpu.SP += 1; cycles = 2

    case 0x05: gb.execDec(&reg.b)
    case 0x15: gb.execDec(&reg.d)
    case 0x25: gb.execDec(&reg.h)
    case 0x35: gb.execIncDecHl(RegisterHl, true); cycles = 3
    case 0x0D: gb.execDec(&reg.c)
    case 0x1D: gb.execDec(&reg.e)
    case 0x2D: gb.execDec(&reg.l)
    case 0x3D: gb.execDec(&reg.a)
    case 0x0B: gb.execDec(RegisterBc); cycles = 2
    case 0x1B: gb.execDec(RegisterDe); cycles = 2
    case 0x2B: gb.execDec(RegisterHl); cycles = 2
    case 0x3B: gb.Cpu.SP -= 1; cycles = 2

    //d8 arth
    case 0xC6: gb.execAdd(&reg.a, d8); pc++; cycles = 2
    case 0xD6: gb.execSub(&reg.a, d8); pc++; cycles = 2
    case 0xE6: gb.execAnd(&reg.a, d8); pc++; cycles = 2
    case 0xF6: gb.execOr(&reg.a, d8); pc++; cycles = 2
    case 0xCE: gb.execAdc(&reg.a, d8); pc++; cycles = 2
    case 0xDE: gb.execSbc(&reg.a, d8); pc++; cycles = 2
    case 0xEE: gb.execXor(&reg.a, d8); pc++; cycles = 2
    case 0xFE: gb.execCp(&reg.a, d8); pc++; cycles = 2
    case 0x06: gb.execLd(&reg.b, d8); pc++; cycles = 2
    case 0x16: gb.execLd(&reg.d, d8); pc++; cycles = 2
    case 0x26: gb.execLd(&reg.h, d8); pc++; cycles = 2
    case 0x36: gb.execLdMem(hl, d8, false); pc++; cycles = 3

    case 0x0A:
        bcValue, err := gb.ReadByte(reg.bc())
        if err != nil { panic(err) }
        gb.execLd(&reg.a, bcValue)
        cycles = 2
    case 0x1A:
        deValue, err := gb.ReadByte(reg.de())
        if err != nil { panic(err) }
        gb.execLd(&reg.a, deValue)
        cycles = 2
    case 0x2A:
        gb.execLd(&reg.a, hlValue)
        gb.execInc(RegisterHl)
        cycles = 2
    case 0x3A:
        gb.execLd(&reg.a, hlValue)
        gb.execDec(RegisterHl)
        cycles = 2

    case 0x02: gb.execLdMem(reg.bc(), reg.a, false); cycles = 2
    case 0x12: gb.execLdMem(reg.de(), reg.a, false); cycles = 2
    case 0x22:
        gb.WriteByte(hl, reg.a)
        reg.setHl(hl+1)
        cycles = 2

    case 0x32:
        gb.WriteByte(hl, reg.a)
        reg.setHl(hl-1)
        cycles = 2

    //other misc arth
    case 0x27: gb.execDAA()
    case 0x37: gb.execSCF()
    case 0x2F: gb.execCPL()
    case 0x3F: gb.execCCF()

    // Register A Rotations
    case 0x07: gb.execRotAcc(&reg.a, false, false)
    case 0x17: gb.execRotAcc(&reg.a, false, true)
    case 0x0F: gb.execRotAcc(&reg.a, true, false)
    case 0x1F: gb.execRotAcc(&reg.a, true, true)

    // CB
    case 0xCB: cycles = gb.execCB(d8); pc++

    // jump abs
    case 0xC2: cycles, pc = gb.execJP(d16, !reg.f.Zero,4,3)
    case 0xD2: cycles, pc = gb.execJP(d16, !reg.f.Carry,4,3)
    case 0xC3: cycles, pc = gb.execJP(d16, true,4, 4)
    case 0xCA: cycles, pc = gb.execJP(d16, reg.f.Zero,4,3)
    case 0xDA: cycles, pc = gb.execJP(d16, reg.f.Carry,4,3)
    case 0xE9: cycles, pc = gb.execJP(hl, true, 1, 1)

    // jump relative
    case 0x20: cycles, pc = gb.execJR(d8, !reg.f.Zero,3,2)
    case 0x30: cycles, pc = gb.execJR(d8, !reg.f.Carry,3,2)
    case 0x18: cycles, pc = gb.execJR(d8, true,3, 3)
    case 0x28: cycles, pc = gb.execJR(d8, reg.f.Zero,3,2)
    case 0x38: cycles, pc = gb.execJR(d8, reg.f.Carry,3,2)

    // Interrupts
    case 0xF3: gb.Cpu.InterruptMaster = false
    case 0xFB: gb.Cpu.PendingInterrupt = true

    // misc ld
    case 0x08: gb.execLdMem(d16, gb.Cpu.SP, false); pc = pc + 2; cycles = 5
    case 0xFA: gb.execLdFromMem(&reg.a, d16); pc = pc + 2; cycles = 4
    case 0xEA: gb.execLdMem(d16, reg.a, false); pc = pc + 2; cycles = 4
    case 0xF9: gb.execLd(&gb.Cpu.SP, hl); cycles = 2

    // push
    case 0xC5: gb.StackPush(reg.bc()); cycles = 4
    case 0xD5: gb.StackPush(reg.de()); cycles = 4
    case 0xE5: gb.StackPush(reg.hl()); cycles = 4
    case 0xF5: gb.StackPush(reg.af()); cycles = 4

    // pop
    case 0xC1: reg.setBc(gb.StackPop()); cycles = 3
    case 0xD1: reg.setDe(gb.StackPop()); cycles = 3
    case 0xE1: reg.setHl(gb.StackPop()); cycles = 3
    case 0xF1: reg.setAf(gb.StackPop() & 0xFFF0); cycles = 3

    // rst
    case 0xC7: pc = gb.execRst(0x00); cycles = 4
    case 0xD7: pc = gb.execRst(0x10); cycles = 4
    case 0xE7: pc = gb.execRst(0x20); cycles = 4
    case 0xF7: pc = gb.execRst(0x30); cycles = 4
    case 0xCF: pc = gb.execRst(0x08); cycles = 4
    case 0xDF: pc = gb.execRst(0x18); cycles = 4
    case 0xEF: pc = gb.execRst(0x28); cycles = 4
    case 0xFF: pc = gb.execRst(0x38); cycles = 4;

    // call
    case 0xC4: cycles, pc = gb.execCall(d16, !reg.f.Zero, 6, 3)
    case 0xD4: cycles, pc = gb.execCall(d16, !reg.f.Carry, 6, 3)
    case 0xCC: cycles, pc = gb.execCall(d16, reg.f.Zero, 6, 3)
    case 0xDC: cycles, pc = gb.execCall(d16, reg.f.Carry, 6, 3)
    case 0xCD: cycles, pc = gb.execCall(d16, true, 6, 6)

    // ret
    case 0xC0: cycles, pc = gb.execRet(!reg.f.Zero, 5, 2)
    case 0xD0: cycles, pc = gb.execRet(!reg.f.Carry, 5, 2)
    case 0xC8: cycles, pc = gb.execRet(reg.f.Zero, 5, 2)
    case 0xD8: cycles, pc = gb.execRet(reg.f.Carry, 5, 2)
    case 0xC9: cycles, pc = gb.execRet(true, 4, 4)
    case 0xD9:
        pc = gb.StackPop()
        gb.Cpu.InterruptMaster = true
        //gb.Cpu.PendingInterrupt = true
        cycles = 4

    case 0xE0: gb.execLdMem(uint16(d8), reg.a, true); pc++; cycles = 3
    case 0xF0: gb.execLdh(&reg.a, uint16(d8)); pc++; cycles = 3
    case 0xE8: gb.execAdd(&gb.Cpu.SP, d8); pc++; cycles = 4
    case 0xF8: gb.execLdSp(reg.setHl, &gb.Cpu.SP, d8); pc++; cycles = 3
    case 0xE2: gb.execLdMem(uint16(reg.c), reg.a, true); cycles = 2
    case 0xF2: gb.execLdh(&reg.a, uint16(reg.c)); cycles = 2

    // empty opcode
    case 0xD3, 0xE3, 0xE4, 0xF4, 0xDB, 0xEB, 0xEC, 0xFC, 0xDD, 0xED, 0xFD:
        panic(fmt.Sprintf("EMPTY OPCODE INSTRUCTION HIT %X", opcode))
    }

    gb.Cpu.PC = pc
    return cycles * 4
}

func (gb *GameBoy) execCB(cbOpcode uint8) (cycles int) {

    right := true
    throughCarry := true

    cycles = 2
    reg := &gb.Cpu.Registers

    switch cbOpcode {
    case 0x00: gb.execRot(&reg.b, !right, !throughCarry)
    case 0x01: gb.execRot(&reg.c, !right, !throughCarry)
    case 0x02: gb.execRot(&reg.d, !right, !throughCarry)
    case 0x03: gb.execRot(&reg.e, !right, !throughCarry)
    case 0x04: gb.execRot(&reg.h, !right, !throughCarry)
    case 0x05: gb.execRot(&reg.l, !right, !throughCarry)
    case 0x06: gb.execRot(RegisterHl, !right, !throughCarry); cycles = 4
    case 0x07: gb.execRot(&reg.a, !right, !throughCarry)
        
    case 0x10: gb.execRot(&reg.b, !right, throughCarry)
    case 0x11: gb.execRot(&reg.c, !right, throughCarry)
    case 0x12: gb.execRot(&reg.d, !right, throughCarry)
    case 0x13: gb.execRot(&reg.e, !right, throughCarry)
    case 0x14: gb.execRot(&reg.h, !right, throughCarry)
    case 0x15: gb.execRot(&reg.l, !right, throughCarry)
    case 0x16: gb.execRot(RegisterHl, !right, throughCarry); cycles = 4
    case 0x17: gb.execRot(&reg.a, !right, throughCarry)

    case 0x08: gb.execRot(&reg.b, right, !throughCarry)
    case 0x09: gb.execRot(&reg.c, right, !throughCarry)
    case 0x0A: gb.execRot(&reg.d, right, !throughCarry)
    case 0x0B: gb.execRot(&reg.e, right, !throughCarry)
    case 0x0C: gb.execRot(&reg.h, right, !throughCarry)
    case 0x0D: gb.execRot(&reg.l, right, !throughCarry)
    case 0x0E: gb.execRot(RegisterHl, right, !throughCarry); cycles = 4
    case 0x0F: gb.execRot(&reg.a, right, !throughCarry)
        
    case 0x18: gb.execRot(&reg.b, right, throughCarry)
    case 0x19: gb.execRot(&reg.c, right, throughCarry)
    case 0x1A: gb.execRot(&reg.d, right, throughCarry)
    case 0x1B: gb.execRot(&reg.e, right, throughCarry)
    case 0x1C: gb.execRot(&reg.h, right, throughCarry)
    case 0x1D: gb.execRot(&reg.l, right, throughCarry)
    case 0x1E: gb.execRot(RegisterHl, right, throughCarry); cycles = 4
    case 0x1F: gb.execRot(&reg.a, right, throughCarry)

    case 0x20: gb.execSLA(&reg.b)
    case 0x21: gb.execSLA(&reg.c)
    case 0x22: gb.execSLA(&reg.d)
    case 0x23: gb.execSLA(&reg.e)
    case 0x24: gb.execSLA(&reg.h)
    case 0x25: gb.execSLA(&reg.l)
    case 0x26: gb.execSLA(RegisterHl); cycles = 4
    case 0x27: gb.execSLA(&reg.a)

    case 0x28: gb.execSRA(&reg.b)
    case 0x29: gb.execSRA(&reg.c)
    case 0x2A: gb.execSRA(&reg.d)
    case 0x2B: gb.execSRA(&reg.e)
    case 0x2C: gb.execSRA(&reg.h)
    case 0x2D: gb.execSRA(&reg.l)
    case 0x2E: gb.execSRA(RegisterHl); cycles = 4
    case 0x2F: gb.execSRA(&reg.a)

    case 0x30: gb.execSWAP(&reg.b)
    case 0x31: gb.execSWAP(&reg.c)
    case 0x32: gb.execSWAP(&reg.d)
    case 0x33: gb.execSWAP(&reg.e)
    case 0x34: gb.execSWAP(&reg.h)
    case 0x35: gb.execSWAP(&reg.l)
    case 0x36: gb.execSWAP(RegisterHl); cycles = 4
    case 0x37: gb.execSWAP(&reg.a)

    case 0x38: gb.execSRL(&reg.b)
    case 0x39: gb.execSRL(&reg.c)
    case 0x3A: gb.execSRL(&reg.d)
    case 0x3B: gb.execSRL(&reg.e)
    case 0x3C: gb.execSRL(&reg.h)
    case 0x3D: gb.execSRL(&reg.l)
    case 0x3E: gb.execSRL(RegisterHl); cycles = 4
    case 0x3F: gb.execSRL(&reg.a)

    case 0x40: gb.execBIT(&reg.b, 0)
    case 0x41: gb.execBIT(&reg.c, 0)
    case 0x42: gb.execBIT(&reg.d, 0)
    case 0x43: gb.execBIT(&reg.e, 0)
    case 0x44: gb.execBIT(&reg.h, 0)
    case 0x45: gb.execBIT(&reg.l, 0)
    case 0x46: gb.execBIT(RegisterHl, 0); cycles = 3
    case 0x47: gb.execBIT(&reg.a, 0)

    case 0x48: gb.execBIT(&reg.b, 1)
    case 0x49: gb.execBIT(&reg.c, 1)
    case 0x4A: gb.execBIT(&reg.d, 1)
    case 0x4B: gb.execBIT(&reg.e, 1)
    case 0x4C: gb.execBIT(&reg.h, 1)
    case 0x4D: gb.execBIT(&reg.l, 1)
    case 0x4E: gb.execBIT(RegisterHl, 1); cycles = 3
    case 0x4F: gb.execBIT(&reg.a, 1)

    case 0x50: gb.execBIT(&reg.b, 2)
    case 0x51: gb.execBIT(&reg.c, 2)
    case 0x52: gb.execBIT(&reg.d, 2)
    case 0x53: gb.execBIT(&reg.e, 2)
    case 0x54: gb.execBIT(&reg.h, 2)
    case 0x55: gb.execBIT(&reg.l, 2)
    case 0x56: gb.execBIT(RegisterHl, 2); cycles = 3
    case 0x57: gb.execBIT(&reg.a, 2)

    case 0x58: gb.execBIT(&reg.b, 3)
    case 0x59: gb.execBIT(&reg.c, 3)
    case 0x5A: gb.execBIT(&reg.d, 3)
    case 0x5B: gb.execBIT(&reg.e, 3)
    case 0x5C: gb.execBIT(&reg.h, 3)
    case 0x5D: gb.execBIT(&reg.l, 3)
    case 0x5E: gb.execBIT(RegisterHl, 3); cycles = 3
    case 0x5F: gb.execBIT(&reg.a, 3)

    case 0x60: gb.execBIT(&reg.b, 4)
    case 0x61: gb.execBIT(&reg.c, 4)
    case 0x62: gb.execBIT(&reg.d, 4)
    case 0x63: gb.execBIT(&reg.e, 4)
    case 0x64: gb.execBIT(&reg.h, 4)
    case 0x65: gb.execBIT(&reg.l, 4)
    case 0x66: gb.execBIT(RegisterHl, 4); cycles = 3
    case 0x67: gb.execBIT(&reg.a, 4)

    case 0x68: gb.execBIT(&reg.b, 5)
    case 0x69: gb.execBIT(&reg.c, 5)
    case 0x6A: gb.execBIT(&reg.d, 5)
    case 0x6B: gb.execBIT(&reg.e, 5)
    case 0x6C: gb.execBIT(&reg.h, 5)
    case 0x6D: gb.execBIT(&reg.l, 5)
    case 0x6E: gb.execBIT(RegisterHl, 5); cycles = 3
    case 0x6F: gb.execBIT(&reg.a, 5)

    case 0x70: gb.execBIT(&reg.b, 6)
    case 0x71: gb.execBIT(&reg.c, 6)
    case 0x72: gb.execBIT(&reg.d, 6)
    case 0x73: gb.execBIT(&reg.e, 6)
    case 0x74: gb.execBIT(&reg.h, 6)
    case 0x75: gb.execBIT(&reg.l, 6)
    case 0x76: gb.execBIT(RegisterHl, 6); cycles = 3
    case 0x77: gb.execBIT(&reg.a, 6)

    case 0x78: gb.execBIT(&reg.b, 7)
    case 0x79: gb.execBIT(&reg.c, 7)
    case 0x7A: gb.execBIT(&reg.d, 7)
    case 0x7B: gb.execBIT(&reg.e, 7)
    case 0x7C: gb.execBIT(&reg.h, 7)
    case 0x7D: gb.execBIT(&reg.l, 7)
    case 0x7E: gb.execBIT(RegisterHl, 7); cycles = 3
    case 0x7F: gb.execBIT(&reg.a, 7)

    case 0x80: gb.execRES(&reg.b, 0)
    case 0x81: gb.execRES(&reg.c, 0)
    case 0x82: gb.execRES(&reg.d, 0)
    case 0x83: gb.execRES(&reg.e, 0)
    case 0x84: gb.execRES(&reg.h, 0)
    case 0x85: gb.execRES(&reg.l, 0)
    case 0x86: gb.execRES(RegisterHl, 0); cycles = 4
    case 0x87: gb.execRES(&reg.a, 0)

    case 0x88: gb.execRES(&reg.b, 1)
    case 0x89: gb.execRES(&reg.c, 1)
    case 0x8A: gb.execRES(&reg.d, 1)
    case 0x8B: gb.execRES(&reg.e, 1)
    case 0x8C: gb.execRES(&reg.h, 1)
    case 0x8D: gb.execRES(&reg.l, 1)
    case 0x8E: gb.execRES(RegisterHl, 1); cycles = 4
    case 0x8F: gb.execRES(&reg.a, 1)

    case 0x90: gb.execRES(&reg.b, 2)
    case 0x91: gb.execRES(&reg.c, 2)
    case 0x92: gb.execRES(&reg.d, 2)
    case 0x93: gb.execRES(&reg.e, 2)
    case 0x94: gb.execRES(&reg.h, 2)
    case 0x95: gb.execRES(&reg.l, 2)
    case 0x96: gb.execRES(RegisterHl, 2); cycles = 4
    case 0x97: gb.execRES(&reg.a, 2)

    case 0x98: gb.execRES(&reg.b, 3)
    case 0x99: gb.execRES(&reg.c, 3)
    case 0x9A: gb.execRES(&reg.d, 3)
    case 0x9B: gb.execRES(&reg.e, 3)
    case 0x9C: gb.execRES(&reg.h, 3)
    case 0x9D: gb.execRES(&reg.l, 3)
    case 0x9E: gb.execRES(RegisterHl, 3); cycles = 4
    case 0x9F: gb.execRES(&reg.a, 3)

    case 0xA0: gb.execRES(&reg.b, 4)
    case 0xA1: gb.execRES(&reg.c, 4)
    case 0xA2: gb.execRES(&reg.d, 4)
    case 0xA3: gb.execRES(&reg.e, 4)
    case 0xA4: gb.execRES(&reg.h, 4)
    case 0xA5: gb.execRES(&reg.l, 4)
    case 0xA6: gb.execRES(RegisterHl, 4); cycles = 4
    case 0xA7: gb.execRES(&reg.a, 4)

    case 0xA8: gb.execRES(&reg.b, 5)
    case 0xA9: gb.execRES(&reg.c, 5)
    case 0xAA: gb.execRES(&reg.d, 5)
    case 0xAB: gb.execRES(&reg.e, 5)
    case 0xAC: gb.execRES(&reg.h, 5)
    case 0xAD: gb.execRES(&reg.l, 5)
    case 0xAE: gb.execRES(RegisterHl, 5); cycles = 4
    case 0xAF: gb.execRES(&reg.a, 5)

    case 0xB0: gb.execRES(&reg.b, 6)
    case 0xB1: gb.execRES(&reg.c, 6)
    case 0xB2: gb.execRES(&reg.d, 6)
    case 0xB3: gb.execRES(&reg.e, 6)
    case 0xB4: gb.execRES(&reg.h, 6)
    case 0xB5: gb.execRES(&reg.l, 6)
    case 0xB6: gb.execRES(RegisterHl, 6); cycles = 4
    case 0xB7: gb.execRES(&reg.a, 6)

    case 0xB8: gb.execRES(&reg.b, 7)
    case 0xB9: gb.execRES(&reg.c, 7)
    case 0xBA: gb.execRES(&reg.d, 7)
    case 0xBB: gb.execRES(&reg.e, 7)
    case 0xBC: gb.execRES(&reg.h, 7)
    case 0xBD: gb.execRES(&reg.l, 7)
    case 0xBE: gb.execRES(RegisterHl, 7); cycles = 4
    case 0xBF: gb.execRES(&reg.a, 7)

    case 0xC0: gb.execSET(&reg.b, 0)
    case 0xC1: gb.execSET(&reg.c, 0)
    case 0xC2: gb.execSET(&reg.d, 0)
    case 0xC3: gb.execSET(&reg.e, 0)
    case 0xC4: gb.execSET(&reg.h, 0)
    case 0xC5: gb.execSET(&reg.l, 0)
    case 0xC6: gb.execSET(RegisterHl, 0); cycles = 4
    case 0xC7: gb.execSET(&reg.a, 0)

    case 0xC8: gb.execSET(&reg.b, 1)
    case 0xC9: gb.execSET(&reg.c, 1)
    case 0xCA: gb.execSET(&reg.d, 1)
    case 0xCB: gb.execSET(&reg.e, 1)
    case 0xCC: gb.execSET(&reg.h, 1)
    case 0xCD: gb.execSET(&reg.l, 1)
    case 0xCE: gb.execSET(RegisterHl, 1); cycles = 4
    case 0xCF: gb.execSET(&reg.a, 1)

    case 0xD0: gb.execSET(&reg.b, 2)
    case 0xD1: gb.execSET(&reg.c, 2)
    case 0xD2: gb.execSET(&reg.d, 2)
    case 0xD3: gb.execSET(&reg.e, 2)
    case 0xD4: gb.execSET(&reg.h, 2)
    case 0xD5: gb.execSET(&reg.l, 2)
    case 0xD6: gb.execSET(RegisterHl, 2); cycles = 4
    case 0xD7: gb.execSET(&reg.a, 2)

    case 0xD8: gb.execSET(&reg.b, 3)
    case 0xD9: gb.execSET(&reg.c, 3)
    case 0xDA: gb.execSET(&reg.d, 3)
    case 0xDB: gb.execSET(&reg.e, 3)
    case 0xDC: gb.execSET(&reg.h, 3)
    case 0xDD: gb.execSET(&reg.l, 3)
    case 0xDE: gb.execSET(RegisterHl, 3); cycles = 4
    case 0xDF: gb.execSET(&reg.a, 3)

    case 0xE0: gb.execSET(&reg.b, 4)
    case 0xE1: gb.execSET(&reg.c, 4)
    case 0xE2: gb.execSET(&reg.d, 4)
    case 0xE3: gb.execSET(&reg.e, 4)
    case 0xE4: gb.execSET(&reg.h, 4)
    case 0xE5: gb.execSET(&reg.l, 4)
    case 0xE6: gb.execSET(RegisterHl, 4); cycles = 4
    case 0xE7: gb.execSET(&reg.a, 4)

    case 0xE8: gb.execSET(&reg.b, 5)
    case 0xE9: gb.execSET(&reg.c, 5)
    case 0xEA: gb.execSET(&reg.d, 5)
    case 0xEB: gb.execSET(&reg.e, 5)
    case 0xEC: gb.execSET(&reg.h, 5)
    case 0xED: gb.execSET(&reg.l, 5)
    case 0xEE: gb.execSET(RegisterHl, 5); cycles = 4
    case 0xEF: gb.execSET(&reg.a, 5)

    case 0xF0: gb.execSET(&reg.b, 6)
    case 0xF1: gb.execSET(&reg.c, 6)
    case 0xF2: gb.execSET(&reg.d, 6)
    case 0xF3: gb.execSET(&reg.e, 6)
    case 0xF4: gb.execSET(&reg.h, 6)
    case 0xF5: gb.execSET(&reg.l, 6)
    case 0xF6: gb.execSET(RegisterHl, 6); cycles = 4
    case 0xF7: gb.execSET(&reg.a, 6)

    case 0xF8: gb.execSET(&reg.b, 7)
    case 0xF9: gb.execSET(&reg.c, 7)
    case 0xFA: gb.execSET(&reg.d, 7)
    case 0xFB: gb.execSET(&reg.e, 7)
    case 0xFC: gb.execSET(&reg.h, 7)
    case 0xFD: gb.execSET(&reg.l, 7)
    case 0xFE: gb.execSET(RegisterHl, 7); cycles = 4
    case 0xFF: gb.execSET(&reg.a, 7)
    }

    return cycles
}

func (gb *GameBoy) getImmediateData() (d8 uint8, d16 uint16) {
    // first byte is lower (0-7) and second is higher (8-15)
    d8, _ = gb.ReadByte(gb.Cpu.PC+1)
    d8b, _ := gb.ReadByte(gb.Cpu.PC+2)
    return d8, uint16(d8b) << 8 + uint16(d8)
}

func (gb *GameBoy) execRotAcc(target *uint8, right bool, throughCarry bool) {

    v := gb.getCbValue(target)
    reg := &gb.Cpu.Registers
    var res uint8
    var carry uint8 = 0
    var carryValue bool = false

    switch {
    case right && !throughCarry:
        res = (v>>1) | ((v&1)<<7)

    case right && throughCarry:
        if reg.f.Carry {
            carry = 0x80
        }

        res = (v>>1) | carry
        carryValue = (1 & v) == 1

    case !right && !throughCarry:
        res = (v<<1) | (v>>7)
        carryValue = v > 0x7F

    case !right && throughCarry:
        var carry uint8 = 0
        if reg.f.Carry {
            carry = 1
        }
        
        res = (v << 1) & 0xFF | carry
        carryValue = (v & 0x80) == 0x80
    }

    gb.setCbValue(target, res)

    if right && !throughCarry {
        carryValue = *target > 0x7F
    }

    reg.f.Carry = carryValue
    reg.f.Zero = false
    reg.f.Subtraction = false
    reg.f.HalfCarry = false
}

func (gb *GameBoy) execRot(target interface{}, right bool, throughCarry bool) {

    v := gb.getCbValue(target)
    reg := &gb.Cpu.Registers
    var res uint8
    var carry uint8 = 0
    var carryValue = false

    switch {
    case right && !throughCarry:

        res = (v>>1) | ((v & 1) <<7)
        carryValue = v & 1 == 1

    case right && throughCarry:
        if reg.f.Carry {
            carry = 0x80
        }

        res = (v>>1) | carry
        carryValue = v & 1 == 1

    case !right && !throughCarry:
        res = (v<<1) | (v>>7)
        carryValue = v > 0x7F

    case !right && throughCarry:
        if reg.f.Carry {
            carry = 1
        }

        res = (v << 1) & 0xFF | carry
        carryValue = (v & 0x80) == 0x80
    }

    gb.setCbValue(target, res)
    reg.f.Carry = carryValue
    reg.f.Zero = res == 0
    reg.f.Subtraction = false
    reg.f.HalfCarry = false
}

func (gb *GameBoy) execSLA(target interface{}) {

    v := gb.getCbValue(target)
    carry := v >> 7
    res := (v<<1)&0xFF
    gb.setCbValue(target, res)

    gb.Cpu.Registers.f.Zero = res == 0
    gb.Cpu.Registers.f.Subtraction = false
    gb.Cpu.Registers.f.HalfCarry = false
    gb.Cpu.Registers.f.Carry = carry == 1
}

func (gb *GameBoy) execSRA(target interface{}) {

    v := gb.getCbValue(target)
    res := (v & 128) | (v >> 1)
    gb.setCbValue(target, res)

    gb.Cpu.Registers.f.Carry = (v&1) == 1
    gb.Cpu.Registers.f.Zero = res == 0
    gb.Cpu.Registers.f.Subtraction = false
    gb.Cpu.Registers.f.HalfCarry = false
}

func (gb *GameBoy) execSRL(target interface{}) {

    v := gb.getCbValue(target)
    v16 := uint16(v)
    carry := v16 & 1
    res16 := v16 >> 1
    res := uint8(res16)
    gb.setCbValue(target, res)

    gb.Cpu.Registers.f.Carry = carry == 1
    gb.Cpu.Registers.f.Zero = res16 == 0
    gb.Cpu.Registers.f.Subtraction = false
    gb.Cpu.Registers.f.HalfCarry = false
}

func (gb *GameBoy) execSWAP(target interface{}) {

    v := gb.getCbValue(target)
    a := v >> 4
    b := (v << 4) & 0xF0
    res := uint8(uint16(a | b))
    gb.setCbValue(target, res)

    gb.Cpu.Registers.f.Zero = res == 0
    gb.Cpu.Registers.f.Subtraction = false
    gb.Cpu.Registers.f.HalfCarry = false
    gb.Cpu.Registers.f.Carry = false
}

func (gb *GameBoy) execBIT(target interface{}, bit int) {

    v := gb.getCbValue(target)
    gb.Cpu.Registers.f.Zero = (v>>bit)&1 == 0
    gb.Cpu.Registers.f.Subtraction = false
    gb.Cpu.Registers.f.HalfCarry = true
}

func (gb *GameBoy) execRES(target interface{}, bit int) {

    v := gb.getCbValue(target)
    res := v &^ (0b1 << bit)
    gb.setCbValue(target, res)
}

func (gb *GameBoy) execSET(target interface{}, bit int) {

    v := gb.getCbValue(target)
    res := v | (0b1 << bit)
    gb.setCbValue(target, res)
}

func (gb *GameBoy) execDAA() {

    reg := &gb.Cpu.Registers

    if !reg.f.Subtraction {
        if reg.f.Carry || reg.a > 0x99 {
            reg.a = reg.a + 0x60
            reg.f.Carry = true
        }

        if reg.f.HalfCarry || reg.a & 0xF > 0x9 {
            reg.a = reg.a + 0x06
            reg.f.HalfCarry = false
        }

        reg.f.Zero = reg.a == 0
        return
    }

    if reg.f.Carry && reg.f.HalfCarry {
        reg.a += 0x9A
        reg.f.HalfCarry = false

        reg.f.Zero = reg.a == 0
        return
    } 

    if reg.f.Carry {
        reg.a +=  0xA0
        reg.f.Zero = reg.a == 0
        return
    }

    if reg.f.HalfCarry {
        reg.a += 0xFA
        reg.f.HalfCarry = false
        reg.f.Zero = reg.a == 0
        return
    }

    reg.f.Zero = reg.a == 0
}

func (gb *GameBoy) execSCF() {

    // set carry flag
    gb.Cpu.Registers.f.Subtraction = false
    gb.Cpu.Registers.f.HalfCarry = false
    gb.Cpu.Registers.f.Carry = true
}

func (gb *GameBoy) execCCF() {

    // compliment (invert) carry flag
    gb.Cpu.Registers.f.Subtraction = false
    gb.Cpu.Registers.f.HalfCarry = false
    gb.Cpu.Registers.f.Carry = !gb.Cpu.Registers.f.Carry
}

func (gb *GameBoy) execCPL() {
    gb.Cpu.Registers.a = 0xFF ^ gb.Cpu.Registers.a
    gb.Cpu.Registers.f.Subtraction = true
    gb.Cpu.Registers.f.HalfCarry = true
}


func (gb *GameBoy) execLdh(target *uint8, from uint16) {


    f := 0xFF00+from
    v, err := gb.ReadByte(f)

    if err != nil {
        panic(err)
    }

    *target = v
}


func (gb *GameBoy) execLd(target interface{}, from interface{}) {

    switch f := from.(type) {
    case uint8:
        switch t := target.(type) {
        case *uint8: *t = f
        default: panic("execLD target from combo unknown from = uint8")
        }
    case uint16:
        switch t := target.(type) {
        case *uint16: *t = f
        case int:
            gb.Cpu.Registers.setCombinedRegister(f, t)
        default: panic("execLD target from combo unknown from = uint16")
        }
    }
}

func (gb *GameBoy) execLdSp(setter func(uint16), sp *uint16, d8 uint8) {

    a := int32(*sp)
    b := int32(int8(d8))
    newValue := a + b
    temp := a ^ b ^ newValue

    setter(uint16(newValue))

    gb.Cpu.Registers.f.Zero = false
    gb.Cpu.Registers.f.Subtraction = false
    gb.Cpu.Registers.f.HalfCarry = (temp&0x10) == 0x10
    gb.Cpu.Registers.f.Carry = (temp&0x100) == 0x100
}


func (gb *GameBoy) execLdFromMem(target *uint8, fromAddr uint16) {

    v, err := gb.ReadByte(fromAddr)
    if err != nil {
        panic(err)
    }

    *target = v
}

func (gb *GameBoy) execLdMem(addr uint16, from any, half bool) {

    v := addr

    if half {
        v = 0xFF00+addr
    }

    switch f := from.(type) {
    case uint8:
        gb.WriteByte(v, f)
    case uint16:
        gb.WriteByte(v, uint8(f & 0xFF))
        gb.WriteByte(v+1, uint8((f & 0xFF00)>> 8))
    default:
        panic("execLDMem from type unknown")
    }
}

func (gb *GameBoy) execAdd(target interface{}, from interface{}) {

    switch f := from.(type) {
    case uint8:
        switch t := target.(type) {
        case *uint8:
            a := uint16(*t)
            b := uint16(f)
            newValue := a + b

            *t = uint8(newValue & 0xFF)
            gb.Cpu.Registers.f.HalfCarry = ((a&0xF)+(b&0xF)) > 0xF
            gb.Cpu.Registers.f.Zero = uint8(newValue) == 0
            gb.Cpu.Registers.f.Carry = newValue > 0xFF

    case *uint16: //gb.Cpu.SP, s8
            a := *t
            b := int8(f)
            newValue := uint16(int32(a) + int32(b))
            tmp := a ^ uint16(b) ^ newValue

            *t = newValue
            gb.Cpu.Registers.f.HalfCarry = (tmp & 0x10) == 0x10
            gb.Cpu.Registers.f.Zero = false
            gb.Cpu.Registers.f.Carry = (tmp & 0x100) == 0x100
        }

    case uint16:
        switch t := target.(type) {
        case int:
            a := gb.Cpu.Registers.getCombinedRegister(t)
            newValue := int32(a) + int32(f)
            gb.Cpu.Registers.setCombinedRegister(uint16(newValue), t)
            gb.Cpu.Registers.f.HalfCarry = int32(a&0xFFF) > (newValue&0xFFF)
            gb.Cpu.Registers.f.Carry = newValue > 0xFFFF
        }
    }

    gb.Cpu.Registers.f.Subtraction = false
}

func (gb *GameBoy) execAdc(to *uint8, from uint8) {
    a := uint16(*to)
    b := uint16(from)
    newValue := a + b

    halfCarry := (a&0xF)+(b&0xF)

    if gb.Cpu.Registers.f.Carry {
        newValue++
        halfCarry++
    }

    *to = uint8(newValue & 0xFF)
    gb.Cpu.Registers.f.HalfCarry = halfCarry > 0xF
    gb.Cpu.Registers.f.Zero = uint8(newValue) == 0
    gb.Cpu.Registers.f.Carry = newValue > 0xFF
    gb.Cpu.Registers.f.Subtraction = false
}

func (gb *GameBoy) execSub(to *uint8, from uint8) {
    a := uint16(*to)
    b := uint16(from)
    newValue := int16(a) - int16(b)

    *to = uint8(newValue & 0xFF)
    gb.Cpu.Registers.f.Zero = uint8(newValue) == 0
    gb.Cpu.Registers.f.Subtraction = true
    gb.Cpu.Registers.f.HalfCarry = int16(a&0x0F)-int16(b&0xF) < 0
    gb.Cpu.Registers.f.Carry = newValue < 0
}

func (gb *GameBoy) execSbc(to *uint8, from uint8) {
    a := uint16(*to)
    b := uint16(from)
    newValue := int16(a) - int16(b)
    halfCarry := int16(a&0xF)-int16(b&0xF)

    if gb.Cpu.Registers.f.Carry {
        newValue--
        halfCarry--
    }

    *to = uint8(newValue & 0xFF)

    gb.Cpu.Registers.f.Zero = uint8(newValue) == 0
    gb.Cpu.Registers.f.Subtraction = true
    gb.Cpu.Registers.f.HalfCarry = halfCarry < 0
    gb.Cpu.Registers.f.Carry = newValue < 0
}

func (gb *GameBoy) execAnd(to *uint8, from uint8) {

    newValue := *to & from

    *to = uint8(newValue & 0xFF)
    gb.Cpu.Registers.f.Zero = newValue == 0
    gb.Cpu.Registers.f.Subtraction = false
    gb.Cpu.Registers.f.HalfCarry = true
    gb.Cpu.Registers.f.Carry = false
}

func (gb *GameBoy) execXor(to *uint8, from uint8) {

    newValue := *to ^ from

    *to = uint8(newValue & 0xFF)
    gb.Cpu.Registers.f.Zero = newValue == 0
    gb.Cpu.Registers.f.Subtraction = false
    gb.Cpu.Registers.f.HalfCarry = false
    gb.Cpu.Registers.f.Carry = false
}

func (gb *GameBoy) execOr(to *uint8, from uint8) {

    newValue := *to | from

    *to = uint8(newValue & 0xFF)
    gb.Cpu.Registers.f.Zero = newValue == 0
    gb.Cpu.Registers.f.Subtraction = false
    gb.Cpu.Registers.f.HalfCarry = false
    gb.Cpu.Registers.f.Carry = false
}

func (gb *GameBoy) execCp(to *uint8, from uint8) {

    newValue := *to - from

    gb.Cpu.Registers.f.Zero = newValue == 0
    gb.Cpu.Registers.f.Subtraction = true
    gb.Cpu.Registers.f.HalfCarry = ((from & 0xF) > (*to & 0xF))
    gb.Cpu.Registers.f.Carry = from > *to
}

func (gb *GameBoy) execInc(target interface{}) {

    switch t := target.(type) {
    case *uint8:
        var v uint16 = uint16(*t)
        newValue := uint8(v+1 & 0xFF)
        *t = newValue
        gb.Cpu.Registers.f.Zero = newValue == 0
        gb.Cpu.Registers.f.Subtraction = false
        gb.Cpu.Registers.f.HalfCarry = ((v&0xF)+(1&0xF) > 0xF)
    case int:
        v := gb.Cpu.Registers.getCombinedRegister(t)
        newValue := v+1
        gb.Cpu.Registers.setCombinedRegister(newValue, t)
    }
}

func (gb *GameBoy) execDec(target interface{}) {

    switch t := target.(type) {
    case *uint8:
        var v uint16 = uint16(*t)
        newValue := uint8(v-1 & 0xFF)
        *t = newValue
        gb.Cpu.Registers.f.Zero = newValue == 0
        gb.Cpu.Registers.f.Subtraction = true
        gb.Cpu.Registers.f.HalfCarry = ((v&0xF)-(1&0xF) > 0xF)
    case int:
        v := gb.Cpu.Registers.getCombinedRegister(t)
        newValue := v-1
        gb.Cpu.Registers.setCombinedRegister(uint16(newValue & 0xFFFF), t)
    }
}

func (gb *GameBoy) execIncDecHl(register int, decrement bool) {

    v := gb.getCbValue(register)

    res := v+1
    if decrement {
        gb.Cpu.Registers.f.Subtraction = true
        gb.Cpu.Registers.f.HalfCarry = (v&0x0F) == 0
        res = v-1
    } else {
        gb.Cpu.Registers.f.Subtraction = false
        gb.Cpu.Registers.f.HalfCarry = (v&0xF)+(1&0xF) > 0xF
    }

    gb.setCbValue(register, res)
    gb.Cpu.Registers.f.Zero = res == 0
}

func (gb *GameBoy) execJP(addr uint16, condition bool, cyclesIf int, cycles int) (c int, pc uint16) {

    if condition {
        return cyclesIf, addr
    }

    return cycles, gb.Cpu.PC + 3
}

func (gb *GameBoy) execJR(addr uint8, condition bool, cyclesIf int, cycles int) (c int, pc uint16) {

    if condition {
        return cyclesIf, uint16(int32(gb.Cpu.PC) + int32(int8(addr))) + 2
    }

    return cycles, gb.Cpu.PC + 2
}


func (gb *GameBoy) StackPop() uint16 {

    lo, _ := gb.ReadByte(gb.Cpu.SP)
    gb.Cpu.SP++
    hi, _ := gb.ReadByte(gb.Cpu.SP)
    gb.Cpu.SP++
	return uint16(lo) | (uint16(hi) << 8)
}

func (gb *GameBoy) StackPush(value uint16) {

    hi := uint8(value >> 8 & 0xFF)
    lo := uint8(value & 0xFF)

    gb.Cpu.SP--
    gb.WriteByte(gb.Cpu.SP, hi)
    gb.Cpu.SP--
    gb.WriteByte(gb.Cpu.SP, lo)
}

func (gb *GameBoy) execRst(addr uint16) (pc uint16) {
    gb.StackPush(gb.Cpu.PC+1)
    return addr
}

func (gb *GameBoy) execCall(addr uint16, condition bool, cyclesIf int, cycles int) (c int, pc uint16) {
    if condition {
        gb.StackPush(gb.Cpu.PC+3)
        return cyclesIf, addr
    }

    return cycles, gb.Cpu.PC+3
}

func (gb *GameBoy) execRet(condition bool, cyclesIf int, cycles int) (c int, pc uint16) {

    if condition {
        return cyclesIf, gb.StackPop()
    }

    return cycles, gb.Cpu.PC+1
}

func (gb *GameBoy) getCbValue(target interface{}) uint8 {

    //Cb only needs register or hlvalue memory location

    var v uint8

    switch t := target.(type) {
    case *uint8: v = *t
    case int:
        var hlValue uint16 = gb.Cpu.Registers.getCombinedRegister(t)
        var err error
        v, err = gb.ReadByte(hlValue)
        if err != nil {
            panic(err)
        }
    }

    return v
}

func (gb *GameBoy) setCbValue(target interface{}, res uint8) {

    //Cb only needs register or hlvalue memory location

    switch t := target.(type) {
    case *uint8: *t = res
    case int:
        var hlValue uint16 = gb.Cpu.Registers.getCombinedRegister(t)
        gb.WriteByte(hlValue, uint8(res))
    }
}
