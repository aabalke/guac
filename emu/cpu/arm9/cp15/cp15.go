package cp15

import (
	"github.com/aabalke/guac/emu/nds/mem"
)

type CpRegister struct {
	Op, Cn, Cm, Cp, Pn uint8
}

type Cp15 struct {
	R   map[CpRegister]uint32
	mem *mem.Mem
}

var (
	// id codes
	MAIN = CpRegister{Op: 0, Cn: 0, Cm: 0, Cp: 0, Pn: 15}
	CACH = CpRegister{Op: 0, Cn: 0, Cm: 0, Cp: 1, Pn: 15}
	TCMP = CpRegister{Op: 0, Cn: 0, Cm: 0, Cp: 2, Pn: 15}

	CTRL = CpRegister{Op: 0, Cn: 1, Cm: 0, Cp: 0, Pn: 15}

	// tcm
	DTCM = CpRegister{Op: 0, Cn: 9, Cm: 1, Cp: 0, Pn: 15}
	ITCM = CpRegister{Op: 0, Cn: 9, Cm: 1, Cp: 1, Pn: 15}

	// Cache Control
	HALT  = CpRegister{Op: 0, Cn: 7, Cm: 0, Cp: 4, Pn: 15}
	HALT2 = CpRegister{Op: 0, Cn: 7, Cm: 8, Cp: 2, Pn: 15}
)

func (c *Cp15) Init(mem *mem.Mem) {
	c.R = make(map[CpRegister]uint32)
	c.mem = mem

	// these register values match no$gba
	c.R[CTRL] = 0x00012078
	c.R[CpRegister{Op: 0, Cn: 2, Cm: 0, Cp: 0, Pn: 15}] = 0x00000042
	c.R[CpRegister{Op: 0, Cn: 2, Cm: 0, Cp: 1, Pn: 15}] = 0x00000042
	c.R[CpRegister{Op: 0, Cn: 3, Cm: 0, Cp: 0, Pn: 15}] = 0x00000002

	c.R[CpRegister{Op: 0, Cn: 5, Cm: 0, Cp: 0, Pn: 15}] = 0x00005545
	c.R[CpRegister{Op: 0, Cn: 5, Cm: 0, Cp: 1, Pn: 15}] = 0x00001405
	c.R[CpRegister{Op: 0, Cn: 5, Cm: 0, Cp: 2, Pn: 15}] = 0x15111011
	c.R[CpRegister{Op: 0, Cn: 5, Cm: 0, Cp: 3, Pn: 15}] = 0x05100011

	c.R[CpRegister{Op: 0, Cn: 6, Cm: 0, Cp: 0, Pn: 15}] = 0x04000033
	c.R[CpRegister{Op: 0, Cn: 6, Cm: 1, Cp: 0, Pn: 15}] = 0x0200002B
	c.R[CpRegister{Op: 0, Cn: 6, Cm: 2, Cp: 0, Pn: 15}] = 0x00000000
	c.R[CpRegister{Op: 0, Cn: 6, Cm: 3, Cp: 0, Pn: 15}] = 0x08000035
	c.R[CpRegister{Op: 0, Cn: 6, Cm: 4, Cp: 0, Pn: 15}] = 0x0300001B
	c.R[CpRegister{Op: 0, Cn: 6, Cm: 5, Cp: 0, Pn: 15}] = 0x00000000
	c.R[CpRegister{Op: 0, Cn: 6, Cm: 6, Cp: 0, Pn: 15}] = 0xFFFF001D
	c.R[CpRegister{Op: 0, Cn: 6, Cm: 7, Cp: 0, Pn: 15}] = 0x027FF017

	c.R[MAIN] = 0x41059461
	c.R[CACH] = 0x0F0D2112
	c.R[TCMP] = 0x00140180
	c.R[DTCM] = 0x0300000A
	c.R[ITCM] = 0x00000020
}

func (c *Cp15) Read(reg CpRegister) uint32 {
	return c.R[reg]
}

func (c *Cp15) Write(v uint32, reg CpRegister, lowVector *bool) {

	if reg.Cn == 6 {
		return
	}

	switch reg {
	case TCMP, MAIN, CACH:
		return
	case CTRL:

		mask := uint32(0b1111_1111_0000_1000_0101)
		v &= mask

		c.R[reg] &^= mask
		c.R[reg] |= v

		*lowVector = (c.R[reg] >> 13) == 0

		c.mem.Tcm.DtcmEnabled = (c.R[reg]>>16)&1 != 0
		c.mem.Tcm.DtcmLoadMode = (c.R[reg]>>17)&1 != 0
		c.mem.Tcm.ItcmEnabled = (c.R[reg]>>18)&1 != 0
		c.mem.Tcm.ItcmLoadMode = (c.R[reg]>>19)&1 != 0

		//if v & 1 == 1 { panic("PU MODE")}

	case DTCM:
		v &^= 0b1111_1100_0001
		c.mem.Tcm.DtcmSize = 512 << ((v >> 1) & 0x3F)
		c.mem.Tcm.DtcmBase = v & 0xFFFF_F000

		// base must be size aligned

	case ITCM:
		v &= 0b111110
		c.mem.Tcm.ItcmSize = 512 << ((v >> 1) & 0x3F)
	}

	c.R[reg] = v
}
