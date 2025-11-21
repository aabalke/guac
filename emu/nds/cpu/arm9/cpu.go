package arm9

import (
	"unsafe"

	"github.com/aabalke/guac/emu/nds/cpu"
	"github.com/aabalke/guac/emu/nds/cpu/cp15"
	"github.com/aabalke/guac/emu/nds/debug"
	"github.com/aabalke/guac/emu/nds/mem/dma"
)

var _ = debug.B

type Cpu struct {
	mem    cpu.MemoryInterface
	Irq    *cpu.Irq
	Reg    Reg
	Halted bool

    LowVector bool

    Cp15 *cp15.Cp15

    Dma [4]dma.DMA

    PcPtr unsafe.Pointer
    PcOff int
    isBranching bool
    BranchPc uint32
    LoopCnt uint32
    LoopLen uint32

    Jit *Jit
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

func NewCpu(m cpu.MemoryInterface, irq *cpu.Irq, cp15 *cp15.Cp15) *Cpu {

	c := &Cpu{
		mem: m,
		Irq: irq,
        Cp15: cp15,
	}

    // skip bios
    c.Irq.IME = true

    c.Jit = NewJit(c)
    c.Jit.CreateBlocks()

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
    Mode uint32
}

func (c *Cond) Get() uint32 {

    v := c.Mode

    if c.N { v |= 1 << FLAG_N}
	if c.Z { v |= 1 << FLAG_Z}
	if c.C { v |= 1 << FLAG_C}
	if c.V { v |= 1 << FLAG_V}
	if c.Q { v |= 1 << FLAG_Q}
	if c.I { v |= 1 << FLAG_I}
	if c.F { v |= 1 << FLAG_F}
	if c.T { v |= 1 << FLAG_T}

    return v
}

func (c *Cond) Set(v uint32) {

    c.N = (v >> FLAG_N) & 1 == 1
    c.Z = (v >> FLAG_Z) & 1 == 1
    c.C = (v >> FLAG_C) & 1 == 1
    c.V = (v >> FLAG_V) & 1 == 1
    c.Q = (v >> FLAG_Q) & 1 == 1
    c.I = (v >> FLAG_I) & 1 == 1
    c.F = (v >> FLAG_F) & 1 == 1
    c.T = (v >> FLAG_T) & 1 == 1

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

func (cpu *Cpu) GetOpArm() (uint32, int) {

    r := &cpu.Reg.R

    if cpu.isBranching {
        cpu.isBranching = false
        cpu.PcOff = 0
        if r[PC] != cpu.BranchPc {
            cpu.PcPtr = nil
            //if cpu.LoopCnt >= 10 && debug.B[0] {
            //    if _, ok := fastFuncs[cpu.BranchPc]; !ok {
            //        fmt.Printf(
            //        "LOOP OVER CURR PC %08X BR PC %08X OP %08X LEN %08d LOOP CNT %08d\n",
            //        r[15], cpu.BranchPc, cpu.mem.Read32(cpu.BranchPc, true), cpu.LoopLen, cpu.LoopCnt)
            //    }
            //}

        } else {
            cpu.LoopCnt++

            if f, ok := fastFuncs[r[PC]]; ok {
                return f(cpu)
            }
        }

        // this is here for debugging above, could probably move earlier
        cpu.LoopLen = 0
    }

    if cpu.PcPtr == nil {
        cpu.LoopCnt = 0
        cpu.BranchPc = r[PC]
        if p, ok := cpu.mem.ReadPtr(r[PC], true); ok {
            cpu.PcPtr = p
        } else {
            return cpu.mem.Read32(r[PC], true), 0
        }
    }

    op := *(*uint32)(unsafe.Add(cpu.PcPtr, cpu.PcOff))
    cpu.PcOff += 4
    cpu.LoopLen++
    cpu.isBranching = ((op >> 27) & 1 == 1) || (op >> 12) & 0xF == 0xF

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

            //if cpu.LoopCnt == 100 {
            //    fmt.Printf("LOOP OVER PC %08X LEN %08d\n", cpu.BranchPc, cpu.LoopLen)
            //}
        }
    }

    if cpu.PcPtr == nil {
        if p, ok := cpu.mem.ReadPtr(r[PC], true); ok {
            cpu.LoopCnt = 0
            cpu.LoopLen = 0
            cpu.BranchPc = r[PC]
            cpu.PcPtr = p
        } else {
            return uint16(cpu.mem.Read16(r[PC], true))
        }
    }

    op := *(*uint16)(unsafe.Add(cpu.PcPtr, cpu.PcOff))
    cpu.PcOff += 2
    cpu.LoopLen++
    cpu.isBranching = (op >> 14) != 0

    return op
}

var fastFuncs = map[uint32]func(cpu *Cpu) (uint32, int) {
    0x02039A88:
    func(cpu *Cpu) (uint32, int) {

        r := &cpu.Reg.R
        cpsr := &cpu.Reg.CPSR
        r[0] = cpu.mem.Read32(0x21755F4, true)
        r[0] = cpu.mem.Read32(0x2, true)
        cpsr.Z = r[0] - 2 == 0
        r[PC] = 0x2039AA0

        //cpu.Jit.Blocks[0].f()

        return 0x1AFF_FFF8, 8
    },
    0x0214D000:
    func(cpu *Cpu) (uint32, int) {

        r := &cpu.Reg.R
        cpsr := &cpu.Reg.CPSR

        cpsr.Z = r[12] - r[2] == 0
        if cpsr.V != cpsr.N {
            r[3] = cpu.mem.Read32(r[0] + r[12], true)
            cpu.mem.Write32(r[1] + r[12], r[3], true)
            r[12] += 2
        }

        r[PC] = 0x214D010
        return 0xBAFF_FFFA, 4
    },
    0x0203DA3C:

    func(cpu *Cpu) (uint32, int) {

        r := &cpu.Reg.R
        cpsr := &cpu.Reg.CPSR

        //add r0, r8, r1, lsl #2
        //str r5, [r0, #0x4c]
        //add r1, r1, #1
        //cmp r1, #0x10

        r[0] = r[8] + (r[1] << 2)
        cpu.mem.Write32(r[0] + 0x4C, r[5], true)
        r[1] += 1

        cpsr.Z = r[1] - 0x10 == 0
        r[15] = 0x203DA4C
        return 0xBAFF_FFFA, 4
    },
    0x01FFDD04:

    func(cpu *Cpu) (uint32, int) {

        //add r2, r2, 0x1
        //ldrb r0, [r1, r2, lsl 0x2]
        //cmp r0, r14
        //0x1FF_DD00 (0xBAFFFFFB)

        r := &cpu.Reg.R
        cpsr := &cpu.Reg.CPSR

        r[2] += 1

        cpu.mem.Write8(r[1] + r[2] << 2, uint8(r[0]), true)

        cpsr.Z = r[0] - r[14] == 0


        r[15] = 0x1FF_DD00 
        return 0xBAFFFFFB, 3
    },
    0x020F8A98:

    func(cpu *Cpu) (uint32, int) {
        //mov r7, 0x44
        //mla r11, r3, r7, r2
        //mov r7, 0xA8
        //mla r10, r2, r7, r1
        //ldr r9, [r11, 0x1C]
        //ldr r7, [r11, 0x18]
        //add r3, r3, 0x1
        //ldr r10, [r10, 0x90]
        //sub r7, r9, r7
        //mul r7, r10, r7
        //mov r7, r7, lsl 0x10
        //cmp r3, 0xE
        //add r6, r6, r7, lsr 0x10
        //return 0x20F_8ACC, 0xBAFF_FFF1

        r := &cpu.Reg.R
        cpsr := &cpu.Reg.CPSR

        r[7] = 0x44

        r[11] = r[3] * r[7] + r[2]

        r[7] = 0xA8

        r[10] = r[2] * r[7] + r[1]

        r[9] = cpu.mem.Read32(r[11] + 0x1C, true)
        r[7] = cpu.mem.Read32(r[11] + 0x18, true)

        r[3] += 1

        r[10] = cpu.mem.Read32(r[10] + 0x90, true)

        r[7] = r[9] - r[7]

        r[7] = r[10] * r[7]

        r[7] = r[7] << 0x10

        cpsr.Z = r[3] - 0xE == 0

        r[6] = r[6] + r[7] >> 0x10

        r[15] =  0x20F_8ACC

        return 0xBAFF_FFF1, 13
    },

    0x0210E294:
    func(cpu *Cpu) (uint32, int) {
        //ldr r1, [r3]
        //add r12, r12, 0x1
        //cmp r1, 0x0
        //ldrne r1, [r3, 0x4]
        //addne r1, r1, 0x1
        //strne r1, [r3, 0x4]
        //ldr r2, [r0, 0x10]
        //ldr r1, [r0]
        //add r3, r3, r2
        //cmp r12, r1

        r := &cpu.Reg.R
        cpsr := &cpu.Reg.CPSR

        r[1] = cpu.mem.Read32(r[3], true)

        r[12] += 1

        cpsr.Z = r[1] == 0

        if r[1] != 0 {
            r[1] = cpu.mem.Read32(r[3] + 4, true)
            r[1] += 1
            cpu.mem.Write32(r[3] + 4, r[1], true)
        }

        r[2] = cpu.mem.Read32(r[0] + 0x10, true)
        r[1] = cpu.mem.Read32(r[0], true)
        r[3] += r[2]

        cpsr.Z = r[12] - r[1] == 0

        r[15] = 0x0210_E2C0

        return 0xE12FFF1E, 10
    },

    0x01FFDCF4:
    func(cpu *Cpu) (uint32, int) {
        //add r2, r2 0x1
        //ldrb r0, [r1, r2, lsl 0x2]
        //cmp r0, r14

        r := &cpu.Reg.R
        cpsr := &cpu.Reg.CPSR

        r[2] += 1
        r[0] = cpu.mem.Read8(r[1] + r[2] << 0x2, true)
        cpsr.Z = r[0] - r[14] == 0

        r[15] = 0x01FF_DD00

        return 0xBAFFFFFB, 3
    },

}

func (cpu *Cpu) Read32(addr uint32) uint32 {
    return cpu.mem.Read32(addr, true)
}

func (cpu *Cpu) Read16(addr uint32) uint32 {
    return cpu.mem.Read16(addr, true)
}

func (cpu *Cpu) Read8(addr uint32) uint32 {
    return cpu.mem.Read8(addr, true)
}

func (cpu *Cpu) Write32(addr, v uint32) {
    cpu.mem.Write32(addr, v, true)
}

func (cpu *Cpu) Write16(addr uint32, v uint16) {
    cpu.mem.Write16(addr, v, true)
}

func (cpu *Cpu) Write8(addr uint32, v uint8) {
    cpu.mem.Write8(addr, v, true)
}
