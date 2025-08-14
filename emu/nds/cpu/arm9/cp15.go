package arm9

import (
	"github.com/aabalke/guac/emu/nds/mem"
	"github.com/aabalke/guac/emu/nds/utils"
)

type CpRegister struct {
	op, cn, pn, cp, cm uint8
}

type Cp15 struct {
	R   map[CpRegister]uint32
	mem *mem.Mem
}

var (
    CTRL = CpRegister{op: 0, cn: 1, cm: 0, cp: 0, pn: 15}
    TCMP = CpRegister{op: 0, cn: 0, cm: 0, cp: 2, pn: 15}
    DTCM = CpRegister{op: 0, cn: 9, cm: 1, cp: 0, pn: 15}
    ITCM = CpRegister{op: 0, cn: 9, cm: 1, cp: 1, pn: 15}

)

func (c *Cp15) Init(mem *mem.Mem) {
	c.R = make(map[CpRegister]uint32)
    c.mem = mem

	// these register values match no$gba
	c.R[CTRL] = 0x00012078
	c.R[CpRegister{op: 0, cn: 2, cm: 0, cp: 0, pn: 15}] = 0x00000042
	c.R[CpRegister{op: 0, cn: 2, cm: 0, cp: 1, pn: 15}] = 0x00000042
	c.R[CpRegister{op: 0, cn: 3, cm: 0, cp: 0, pn: 15}] = 0x00000002

	c.R[CpRegister{op: 0, cn: 5, cm: 0, cp: 0, pn: 15}] = 0x00005545
	c.R[CpRegister{op: 0, cn: 5, cm: 0, cp: 1, pn: 15}] = 0x00001405
	c.R[CpRegister{op: 0, cn: 5, cm: 0, cp: 2, pn: 15}] = 0x15111011
	c.R[CpRegister{op: 0, cn: 5, cm: 0, cp: 3, pn: 15}] = 0x05100011

	c.R[CpRegister{op: 0, cn: 6, cm: 0, cp: 0, pn: 15}] = 0x04000033
	c.R[CpRegister{op: 0, cn: 6, cm: 1, cp: 0, pn: 15}] = 0x0200002B
	c.R[CpRegister{op: 0, cn: 6, cm: 2, cp: 0, pn: 15}] = 0x00000000
	c.R[CpRegister{op: 0, cn: 6, cm: 3, cp: 0, pn: 15}] = 0x08000035
	c.R[CpRegister{op: 0, cn: 6, cm: 4, cp: 0, pn: 15}] = 0x0300001B
	c.R[CpRegister{op: 0, cn: 6, cm: 5, cp: 0, pn: 15}] = 0x00000000
	c.R[CpRegister{op: 0, cn: 6, cm: 6, cp: 0, pn: 15}] = 0xFFFF001D
	c.R[CpRegister{op: 0, cn: 6, cm: 7, cp: 0, pn: 15}] = 0x027FF017

    c.R[TCMP] = 0x00140180
	c.R[DTCM] = 0x0300000A
	c.R[ITCM] = 0x00000020
}

func (c *Cp15) Read(reg CpRegister) uint32 {
	return c.R[reg]
}

func (c *Cp15) Write(v uint32, reg CpRegister) {

    switch reg {
    case CTRL:

        mask := uint32(0b11111111000010000101)
        v &= mask
        c.R[reg] &^= mask
        c.R[reg] |= v

        c.mem.Tcm.DtcmEnabled = utils.BitEnabled(v, 16)
        c.mem.Tcm.DtcmLoadMode = utils.BitEnabled(v, 17)
        c.mem.Tcm.ItcmEnabled = utils.BitEnabled(v, 18)
        c.mem.Tcm.ItcmLoadMode = utils.BitEnabled(v, 19)

        if v & 1 == 1 { panic("PU MODE")}

    case TCMP:

        return

    case DTCM:
        v &^= 0b1111_1100_0001

        c.mem.Tcm.DtcmSize = 512 << utils.GetVarData(v, 1, 6)
        c.mem.Tcm.DtcmBase = utils.GetVarData(v, 12, 31) << 12

        // base must be size aligned

    case ITCM:
        v &= 0b111110

        c.mem.Tcm.ItcmSize = 512 << utils.GetVarData(v, 1, 6)
    }

	c.R[reg] = v
}
