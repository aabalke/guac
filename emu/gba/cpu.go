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

	BIOS_STARTUP  = 0
	BIOS_SWI      = 1
	BIOS_IRQ      = 2
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
	BIOS_STARTUP:  0xE129F000,
	BIOS_SWI:      0xE3A02004,
	BIOS_IRQ:      0x03007FFC,
	BIOS_IRQ_POST: 0xE55EC002,
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

	c.Reg.FIQ[0] = 0x4098_8194
	c.Reg.FIQ[1] = 0x0410_0084
	c.Reg.FIQ[2] = 0x808C_1042
	c.Reg.FIQ[3] = 0x16A0_439B
	c.Reg.FIQ[4] = 0x4482_0443

	c.Reg.SP[BANK_ID[MODE_FIQ]] =   0x0041_0C81
	c.Reg.LR[BANK_ID[MODE_FIQ]] =   0xA928_314E
	c.Reg.SPSR[BANK_ID[MODE_FIQ]] = 0xF000_00FF

    c.Reg.SP[BANK_ID[MODE_IRQ]] =   0x0300_7FA0
    c.Reg.LR[BANK_ID[MODE_IRQ]] =   0x0000_0000
    c.Reg.SPSR[BANK_ID[MODE_IRQ]] = 0x0000_0010

	return c
}

func (c *Cpu) Execute(opcode uint32) int {

	if c.Reg.CPSR.GetFlag(FLAG_T) {
		return c.DecodeTHUMB(uint16(opcode))
	}

    return c.DecodeARM(opcode)
}

type Reg struct {
	Cpu  *Cpu
	R    [16]uint32
	SP   [6]uint32
	LR   [6]uint32
	FIQ  [5]uint32 // r8 - r12
	USR  [5]uint32 // r8 - r12 // tmp to restore after FIQ
	CPSR Cond
	SPSR [6]Cond
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

func (r *Reg) getMode() uint32 {
	return uint32(r.CPSR) & 0b11111
}

func (r *Reg) setMode(prev, curr uint32) {

	if prev == curr {
		return
	}

	r.CPSR.SetMode(curr)

	if BANK_ID[prev] == BANK_ID[curr] {
		return
	}



	r.switchRegisterBanks(prev, curr)
}

func (r *Reg) switchRegisterBanks(prev, curr uint32) {

	if prev != MODE_FIQ {
		for i := range 5 {
			r.USR[i] = r.R[8+i]
		}
	}

	r.SP[BANK_ID[prev]] = r.R[SP]
	r.LR[BANK_ID[prev]] = r.R[LR]

	if prev == MODE_FIQ {
		for i := range 5 {
			r.FIQ[i] = r.R[8+i]
		}
	}

	if curr != MODE_FIQ {
		for i := range 5 {
			r.R[8+i] = r.USR[i]
		}
	}

	r.R[SP] = r.SP[BANK_ID[curr]]
	r.R[LR] = r.LR[BANK_ID[curr]]

	if curr == MODE_FIQ {
		for i := range 5 {
			r.R[8+i] = r.FIQ[i]
		}
	}

}
