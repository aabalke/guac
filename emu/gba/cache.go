package gba

type Cache struct {
    Bios [0x4000 >> 2]InstructionARM
    Rom [0x200_0000 >> 2]InstructionARM

    BiosThumb [0x4000 >> 1]InstructionThumb
    RomThumb [0x200_0000 >> 1]InstructionThumb
}

type InstructionARM struct{
    Inst func(cpu *Cpu, opcode uint32)
    Invalid bool
}

type InstructionThumb struct{
    Inst func(cpu *Cpu, opcode uint16)
    Invalid bool
}

func (c *Cache) BuildCache(gba *GBA) {

    j := uint32(0)
    for i := uint32(0); i < 0x4000; i += 4 {
        opcode := gba.Mem.Read32(i)
        c.CacheInstructionARM(j, gba.Cpu, opcode, true)
        j++
    }

    j = uint32(0)
    for i := uint32(0x800_0000); i < 0xA00_0000; i += 4 {
        opcode := gba.Mem.Read32(i)
        c.CacheInstructionARM(j, gba.Cpu, opcode, false)
        j++
    }

    j = uint32(0)
    for i := uint32(0); i < 0x4000; i += 2 {
        opcode := uint16(gba.Mem.Read16(i))
        c.CacheInstructionTHUMB(j, gba.Cpu, opcode, true)
        j++
    }

    j = uint32(0)
    for i := uint32(0x800_0000); i < 0xA00_0000; i += 2 {
        opcode := uint16(gba.Mem.Read16(i))
        c.CacheInstructionTHUMB(j, gba.Cpu, opcode, false)
        j++
    }

}

func (c *Cache) CacheInstructionARM(i uint32, cpu *Cpu, opcode uint32, bios bool) {

    v := InstructionARM{}

    switch {
    case isB(opcode):
        v.Inst = func(cpu *Cpu, opcode uint32) { cpu.B(opcode) }
    case isBX(opcode):
        v.Inst = func(cpu *Cpu, opcode uint32) { cpu.BX(opcode) }
    case isSDT(opcode):
        v.Inst = func(cpu *Cpu, opcode uint32) { cpu.Sdt(opcode) }
    case isBlock(opcode):
        v.Inst = func(cpu *Cpu, opcode uint32) { cpu.Block(opcode) }
    case isHalf(opcode):
        v.Inst = func(cpu *Cpu, opcode uint32) { cpu.Half(opcode) }
    case isPSR(opcode):
        v.Inst = func(cpu *Cpu, opcode uint32) { cpu.Psr(opcode) }
    case isSWP(opcode):
        v.Inst = func(cpu *Cpu, opcode uint32) { cpu.Swp(opcode) }
    case isM(opcode):
        v.Inst = func(cpu *Cpu, opcode uint32) { cpu.Mul(opcode) }
    case isALU(opcode):
        v.Inst = func(cpu *Cpu, opcode uint32) { cpu.Alu(opcode) }
    default:
        v.Invalid = true
	}

    if bios {
        c.Bios[i] = v
    } else {
        c.Rom[i] = v
    }
}

func (c *Cache) CacheInstructionTHUMB(i uint32, cpu *Cpu, opcode uint16, bios bool) {

    v := InstructionThumb{}

	switch {
    case isthumbSWI(opcode):
        v.Invalid = true
	case isThumbAddSub(opcode):
		v.Inst = func(cpu *Cpu, opcode uint16) { cpu.ThumbAddSub(opcode) }
	case isThumbShift(opcode):
        v.Inst = func(cpu *Cpu, opcode uint16) { cpu.thumbShifted(opcode) }
	case isThumbImm(opcode):
        v.Inst = func(cpu *Cpu, opcode uint16) { cpu.thumbImm(opcode) }
	case isThumbAlu(opcode):
		v.Inst = func(cpu *Cpu, opcode uint16) { cpu.ThumbAlu(opcode) }
	case isThumbHiReg(opcode):
        v.Inst = func(cpu *Cpu, opcode uint16) { cpu.HiRegBX(opcode) }
	case isLSHalf(opcode):
		v.Inst = func(cpu *Cpu, opcode uint16) { cpu.thumbLSHalf(opcode) }
	case isLSSigned(opcode):
		v.Inst = func(cpu *Cpu, opcode uint16) { cpu.thumbLSSigned(opcode) }
	case isLPC(opcode):
		v.Inst = func(cpu *Cpu, opcode uint16) { cpu.thumbLPC(opcode) }
	case isLSR(opcode):
		v.Inst = func(cpu *Cpu, opcode uint16) { cpu.thumbLSR(opcode) }
	case isLSImm(opcode):
		v.Inst = func(cpu *Cpu, opcode uint16) { cpu.thumbLSImm(opcode) }
	case isPushPop(opcode):
		v.Inst = func(cpu *Cpu, opcode uint16) { cpu.thumbPushPop(opcode) }
    case isRelative(opcode):
        v.Inst = func(cpu *Cpu, opcode uint16) { cpu.thumbRelative(opcode) }
    case isThumbB(opcode):
        v.Inst = func(cpu *Cpu, opcode uint16) { cpu.thumbB(opcode) }
    case isJumpCall(opcode):
        v.Inst = func(cpu *Cpu, opcode uint16) { cpu.thumbJumpCalls(opcode) }
    case isStack(opcode):
        v.Inst = func(cpu *Cpu, opcode uint16) { cpu.thumbStack(opcode) }
    case isLongBranch(opcode):
        v.Inst = func(cpu *Cpu, opcode uint16) { cpu.thumbLongBranch(opcode) }
    case isShortLongBranch(opcode):
        v.Inst = func(cpu *Cpu, opcode uint16) { cpu.thumbShortLongBranch(opcode) }
    case isLSSP(opcode):
        v.Inst = func(cpu *Cpu, opcode uint16) { cpu.thumbLSSP(opcode) }
    case isMulti(opcode):
        v.Inst = func(cpu *Cpu, opcode uint16) { cpu.thumbMulti(opcode) }
	default:
        v.Invalid = true
	}

    if bios {
        c.BiosThumb[i] = v
    } else {
        c.RomThumb[i] = v
    }
}

func (c *Cache) runCache(gba *GBA, pc, opcode uint32) bool {


    if !(pc < 0x4000 || (pc >= 0x800_0000 && pc < 0xE00_0000)) {
        return false
    }

    if thumb := gba.Cpu.Reg.CPSR.GetFlag(FLAG_T); thumb {

        if pc & 0b1 != 0 {
            return false
        }

        var v InstructionThumb

        if pc <= 0x4000 {
            v = c.BiosThumb[pc >> 1]
        } else {
            v = c.RomThumb[((pc - 0x800_0000) % 0x200_0000) >> 1]
        }

        if v.Invalid {
            return false
        }

        v.Inst(gba.Cpu, uint16(opcode))

        return true
    }

    if pc & 0b11 != 0 {
        return false
    }

    var v InstructionARM

    if pc <= 0x4000 {
        v = c.Bios[pc >> 2]
    } else {
        v = c.Rom[((pc - 0x800_0000) % 0x200_0000) >> 2]
    }

    if v.Invalid {
        return false
    }

    v.Inst(gba.Cpu, opcode)

    return true
}
