package arm9

// Register index is 0x123, 1 = Cn, 2 = Cm, 3 = Cp

type Cp15 struct {
	cpu            *Cpu
	R              map[uint32]uint32
	ProtectionUnit ProtectionUnit

	Ctrl uint32
}

var (
	MAIN  = uint32(0x000)
	CACH  = uint32(0x001)
	TCMP  = uint32(0x002)
	CTRL  = uint32(0x100)
	DTCM  = uint32(0x910)
	ITCM  = uint32(0x911)
	HALT  = uint32(0x704)
	HALT2 = uint32(0x782)
)

func NewCp15(cpu *Cpu) *Cp15 {
	c := &Cp15{
		R:   make(map[uint32]uint32),
		cpu: cpu,
	}

	c.Write(CTRL, 0x00012078)
	c.Write(DTCM, 0x0300000A)
	c.Write(ITCM, 0x00000020)

	c.Write(0x200, 0x00000042)
	c.Write(0x201, 0x00000042)
	c.Write(0x300, 0x00000002)
	c.Write(0x500, 0x00005545)
	c.Write(0x501, 0x00001405)
	c.Write(0x502, 0x15111011)
	c.Write(0x503, 0x05100011)

	c.Write(0x600, 0x04000033)
	c.Write(0x610, 0x0200002B)
	c.Write(0x620, 0x00000000)
	c.Write(0x630, 0x08000035)
	c.Write(0x640, 0x0300001B)
	c.Write(0x650, 0x00000000)
	c.Write(0x660, 0xFFFF001D)
	c.Write(0x670, 0x027FF017)

	return c
}

func (c *Cp15) Read(idx uint32) uint32 {
	switch idx {
	case MAIN:
		return 0x41059461
	case CACH:
		return 0x0F0D2112
	case TCMP:
		return 0x00140180
	case CTRL:
		return c.Ctrl
	case 0x500, 0x501:

		permissions := &c.ProtectionUnit.DataCache.Permissions
		if idx == 0x501 {
			permissions = &c.ProtectionUnit.InstCache.Permissions
		}

		val := uint32(0)
		for i := range 8 {
			val |= ((*permissions >> (i * 4)) & 3) << (i * 2)
		}
		return val

	case 0x502:
		return c.ProtectionUnit.DataCache.Permissions
	case 0x503:
		return c.ProtectionUnit.InstCache.Permissions
	case 0x200:
		return uint32(c.ProtectionUnit.DataCache.Ctrl)
	case 0x201:
		return uint32(c.ProtectionUnit.InstCache.Ctrl)

	case 0x600, 0x610, 0x620, 0x630, 0x640, 0x650, 0x660, 0x670,
		0x601, 0x611, 0x621, 0x631, 0x641, 0x651, 0x661, 0x671:
		return c.ProtectionUnit.Regions[(idx>>4)&7]

	default:
		return c.R[idx]
	}
}

func (c *Cp15) Write(idx uint32, v uint32) {
	//fmt.Printf("ID %03X V %08X\n", idx, v)
	switch idx {
	case CTRL:

		prev := c.Ctrl

		c.Ctrl = (v & 0xFF085) | 0x78

		c.ProtectionUnit.Enabled = (c.Ctrl>>0)&1 != 0
		c.ProtectionUnit.DataCache.Enabled = (c.Ctrl>>2)&1 != 0
		c.ProtectionUnit.InstCache.Enabled = (c.Ctrl>>12)&1 != 0

		c.cpu.LowVector = (c.Ctrl>>13)&1 == 0
		c.cpu.Dtcm.enabled = (c.Ctrl>>16)&1 != 0
		c.cpu.Dtcm.loadMode = (c.Ctrl>>17)&1 != 0
		c.cpu.Itcm.enabled = (c.Ctrl>>18)&1 != 0
		c.cpu.Itcm.loadMode = (c.Ctrl>>19)&1 != 0

		if prev&0x1005 != c.Ctrl&0x1005 {
			c.ProtectionUnit.RefreshMPU()
		}

	case DTCM:
		v &= 0xFFFF_F07E
		sizeShift := max(3, min(((v>>1)&0x1F), 23))
		c.cpu.Dtcm.size = 512 << sizeShift
		c.cpu.Dtcm.base = v & 0xFFFF_F000
		c.R[idx] = v

	case ITCM:
		v &= 0xFFFF_F07E
		sizeShift := max(3, min(((v>>1)&0x1F), 23))
		c.cpu.Itcm.size = 512 << sizeShift
		c.R[idx] = v

	case HALT, HALT2:
		c.cpu.Halted = true

	case 0x600, 0x610, 0x620, 0x630, 0x640, 0x650, 0x660, 0x670,
		0x601, 0x611, 0x621, 0x631, 0x641, 0x651, 0x661, 0x671:

		n := int((idx >> 4) & 7)
		c.ProtectionUnit.Regions[n] = v
		c.ProtectionUnit.RefreshMPU()

	case 0x200, 0x201:

		ctrl := &c.ProtectionUnit.DataCache.Ctrl
		if idx == 0x201 {
			ctrl = &c.ProtectionUnit.InstCache.Ctrl
		}

		diff := *ctrl ^ uint8(v)
		*ctrl = uint8(v)

		for i := range 8 {
			if diff&(1<<i) != 0 {
				c.ProtectionUnit.UpdateRegion(i)
			}
		}

	case 0x500, 0x501:

		permissions := &c.ProtectionUnit.DataCache.Permissions
		if idx == 0x501 {
			permissions = &c.ProtectionUnit.InstCache.Permissions
		}

		prev := *permissions
		*permissions = 0

		for i := range 8 {
			*permissions |= (v & 3) << (i * 4)
		}

		diff := *permissions ^ prev

		for i := range 8 {
			if (diff & (0xF << (i * 4))) != 0 {
				c.ProtectionUnit.UpdateRegion(i)
			}
		}

	case 0x502, 0x503:

		permissions := &c.ProtectionUnit.DataCache.Permissions
		if idx == 0x503 {
			permissions = &c.ProtectionUnit.InstCache.Permissions
		}

		diff := *permissions ^ v
		*permissions = v

		for i := range 8 {
			if (diff & (0xF << (i * 4))) != 0 {
				c.ProtectionUnit.UpdateRegion(i)
			}
		}

	default:
		c.R[idx] = v
	}
}
