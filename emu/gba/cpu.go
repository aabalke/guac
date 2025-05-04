package gba

type Cpu struct {
	Gba *GBA
	Reg Reg
	Loc uint32
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

    BIOS_STARTUP = 0
    BIOS_SWI = 1
    BIOS_IRQ_PRE = 2
    BIOS_IRQ_POST = 3
)

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
    BIOS_STARTUP:   0xE129F000,
    BIOS_SWI:       0xE3A02004,
    BIOS_IRQ_PRE:   0x03007FFC,
    BIOS_IRQ_POST:  0xE55EC002,
}

func NewCpu(gba *GBA) *Cpu {

	c := &Cpu{
		Reg: Reg{},
		Gba: gba,
	}

	c.Reg.R[0] = 0x0000_0CA5
	c.Reg.R[SP] = 0x0300_7F00
	c.Reg.R[LR] = 0x0800_0000
	c.Reg.R[PC] = 0x0800_0000
	c.Reg.CPSR = 0x0000_001F
	return c
}

func (c *Cpu) Execute(opcode uint32) {

    if c.Reg.CPSR.GetFlag(FLAG_T) {
        c.DecodeTHUMB(uint16(opcode))
    } else {
        c.DecodeARM(opcode)
    }

}

type Reg struct {
    // FIQ is ignored
	Cpu  *Cpu
	R    [16]uint32
    SP   [5]uint32
    LR   [5]uint32
	CPSR Cond
	SPSR [5]Cond
}

type Cond uint32

func (c *Cond) GetFlag(flag uint32) bool {

	switch flag {
	case FLAG_N, FLAG_Z, FLAG_C, FLAG_V, FLAG_Q, FLAG_I, FLAG_F, FLAG_T:
		return (uint32(*c)>>flag)&0b1 == 0b1
	}

	panic("Unknown Cond Flag Get")
}

func (c *Cond) SetFlag(flag uint32, value bool) {
	switch flag {
	case FLAG_N, FLAG_Z, FLAG_C, FLAG_V, FLAG_Q, FLAG_I, FLAG_F, FLAG_T:

		if value {
			*c |= (0b1 << flag)
			return
		}

		*c &^= (0b1 << flag)

		return
	}

	panic("Unknown Cond Flag Set")
}

func (c *Cond) SetField(loBit uint32, value uint32) {
	mask := 0b1111_1111 << loBit
	*c &^= Cond(mask)
    value <<= loBit
	*c |= Cond(value)
}

func (c *Cond) SetMode(mode uint32) {
    *c = Cond((uint32(*c) &^ 0b11111) | mode)
}

func (r *Reg) getMode() uint32{
    return uint32(r.CPSR) & 0b11111
}

func (r *Reg) setMode(mode uint32) {

    curr := r.getMode()

    if curr == mode {
        return
    }

    if mode == MODE_FIQ {
        panic("Gba has been set to unsupported FIQ Mode")
    }

    r.CPSR.SetMode(mode)

    if BANK_ID[curr] == BANK_ID[mode] {
        return
    }

    r.SP[BANK_ID[curr]] = r.R[SP]
    r.LR[BANK_ID[curr]] = r.R[LR]

    r.R[SP] = r.SP[BANK_ID[mode]]
    r.R[LR] = r.LR[BANK_ID[mode]]
}
