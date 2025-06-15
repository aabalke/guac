package gba
//
//
//type Cache struct {
//    Arm [0x80_0000]CacheInstArm
//    Thumb [0x100_0000]CacheInstThumb
//}
//
//type CacheInstArm struct {
//    Cached bool
//    Exec func (cpu *Cpu, opcode uint32)
//}
//
//type CacheInstThumb struct {
//    Cached bool
//    Exec func (cpu *Cpu, opcode uint16)
//}
//
//
//func (c *Cache) Build(gba *GBA) {
//
//    // CURRENTLY JUST ROM AND NOT BLOCK BASED
//
//    mem := gba.Mem
//
//    s := 0x800_0000
//    e := 0xA00_0000
//
//    j := 0
//    k := 0
//    for i := 0; i < e - s; i += 4 {
//        opcode := mem.Read32(uint32(s + i))
//        cached, exec := c.DecodeARM(gba.Cpu, opcode)
//
//        c.Arm[j] = CacheInstArm{
//            Cached: cached,
//            Exec: exec,
//        }
//        j++
//
//        cachedT, execT := c.DecodeThumb(gba.Cpu, uint16(opcode))
//        c.Thumb[k] = CacheInstThumb{
//            Cached: cachedT,
//            Exec: execT,
//        }
//        k++
//
//        cachedT, execT = c.DecodeThumb(gba.Cpu, uint16(opcode >> 16))
//        c.Thumb[k] = CacheInstThumb{
//            Cached: cachedT,
//            Exec: execT,
//        }
//        k++
//    }
//}
//
//func (c *Cache) CheckArm(gba *GBA, opcode uint32, addr uint32) bool {
//
//    if addr < 0x800_0000 || addr >= 0xA00_0000 {
//        return false
//    }
//
//    romAddr := (addr - 0x800_0000) / 4
//    cached := c.Arm[romAddr].Cached
//
//    if cached {
//        c.Arm[romAddr].Exec(gba.Cpu, opcode)
//    }
//
//    return cached
//}
//
//func (c *Cache) CheckThumb(gba *GBA, opcode uint16, addr uint32) bool {
//
//    if addr < 0x800_0000 || addr >= 0xA00_0000 {
//        return false
//    }
//
//    romAddr := (addr - 0x800_0000) / 2
//    cached := c.Thumb[romAddr].Cached
//
//    if cached {
//        c.Thumb[romAddr].Exec(gba.Cpu, opcode)
//    }
//
//    return cached
//}
//
//func (c *Cache) DecodeARM(cpu *Cpu, opcode uint32) (Cached bool, Exec func (cpu *Cpu, opcode uint32)) {
//
//	switch {
//	case isB(opcode):
//		return true, func(cpu *Cpu, opcode uint32) { cpu.B(opcode) }
//	case isBX(opcode):
//		return true, func(cpu *Cpu, opcode uint32) { cpu.BX(opcode) }
//	case isSDT(opcode):
//        return true, func(cpu *Cpu, opcode uint32) { cpu.Sdt(opcode) }
//	case isBlock(opcode):
//		return true, func(cpu *Cpu, opcode uint32) { cpu.Block(opcode) }
//	case isHalf(opcode):
//		return true, func(cpu *Cpu, opcode uint32) { cpu.Half(opcode) }
//	case isPSR(opcode):
//		return true, func(cpu *Cpu, opcode uint32) { cpu.Psr(opcode) }
//	case isSWP(opcode):
//		return true, func(cpu *Cpu, opcode uint32) { cpu.Swp(opcode) }
//    case isM(opcode):
//        return true, func(cpu *Cpu, opcode uint32) { cpu.Mul(opcode) }
//	case isALU(opcode):
//		return true, func(cpu *Cpu, opcode uint32) { cpu.Alu(opcode) }
//	}
//
//    return false, func(cpu *Cpu, opcode uint32) {}
//}
//
//func (c *Cache) DecodeThumb(cpu *Cpu, opcode uint16) (Cached bool, Exec func (cpu *Cpu, opcode uint16)) {
//
//	switch {
//	case isThumbAddSub(opcode):
//		return true, func(cpu *Cpu, opcode uint16) { cpu.ThumbAddSub(opcode) }
//	case isThumbShift(opcode):
//        return true, func(cpu *Cpu, opcode uint16) { cpu.thumbShifted(opcode) }
//	case isThumbImm(opcode):
//        return true, func(cpu *Cpu, opcode uint16) { cpu.thumbImm(opcode) }
//	case isThumbAlu(opcode):
//		return true, func(cpu *Cpu, opcode uint16) { cpu.ThumbAlu(opcode) }
//	case isThumbHiReg(opcode):
//        return true, func(cpu *Cpu, opcode uint16) { cpu.HiRegBX(opcode) }
//	case isLSHalf(opcode):
//		return true, func(cpu *Cpu, opcode uint16) { cpu.thumbLSHalf(opcode) }
//	case isLSSigned(opcode):
//		return true, func(cpu *Cpu, opcode uint16) { cpu.thumbLSSigned(opcode) }
//	case isLPC(opcode):
//		return true, func(cpu *Cpu, opcode uint16) { cpu.thumbLPC(opcode) }
//	case isLSR(opcode):
//		return true, func(cpu *Cpu, opcode uint16) { cpu.thumbLSR(opcode) }
//	case isLSImm(opcode):
//		return true, func(cpu *Cpu, opcode uint16) { cpu.thumbLSImm(opcode) }
//	case isPushPop(opcode):
//		return true, func(cpu *Cpu, opcode uint16) { cpu.thumbPushPop(opcode) }
//    case isRelative(opcode):
//        return true, func(cpu *Cpu, opcode uint16) { cpu.thumbRelative(opcode) }
//    case isThumbB(opcode):
//        return true, func(cpu *Cpu, opcode uint16) { cpu.thumbB(opcode) }
//    case isJumpCall(opcode):
//        return true, func(cpu *Cpu, opcode uint16) { cpu.thumbJumpCalls(opcode) }
//    case isStack(opcode):
//        return true, func(cpu *Cpu, opcode uint16) { cpu.thumbStack(opcode) }
//    case isLongBranch(opcode):
//        return true, func(cpu *Cpu, opcode uint16) { cpu.thumbLongBranch(opcode) }
//    case isLSSP(opcode):
//        return true, func(cpu *Cpu, opcode uint16) { cpu.thumbLSSP(opcode) }
//    case isMulti(opcode):
//        return true, func(cpu *Cpu, opcode uint16) { cpu.thumbMulti(opcode) }
//	}
//
//    return false, func(cpu *Cpu, opcode uint16) {}
//}
