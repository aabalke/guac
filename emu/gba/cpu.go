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
)

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
	Cpu  *Cpu
	R    [16]uint32
	CPSR Cond
	SPSR Cond
}

const (
	FLAG_N = 31
	FLAG_Z = 30
	FLAG_C = 29
	FLAG_V = 28
	FLAG_Q = 27
	FLAG_I = 7
	FLAG_F = 6
	FLAG_T = 5
)

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
