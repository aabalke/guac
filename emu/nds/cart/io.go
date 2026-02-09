package cart

import "log"

type ExMem struct {
	IsGBAAccessArm7          bool
	isCartAccessArm7         bool
	v                        uint16
}

// gba values are separate instances on arm7 and arm9

func (c *Cartridge) ReadExMem(b uint8) uint8 {
	return uint8(c.ExMem.v >> (b << 3))
}

func (c *Cartridge) WriteExMem(v uint8, b uint8) {

	// top bits can only be written by arm9
	c.ExMem.v &^= 0xFF << (b << 3)
	c.ExMem.v |= uint16(v) << (b << 3)
	c.ExMem.v |= 1 << 13 // always set

	switch b {
	case 0:
		c.ExMem.IsGBAAccessArm7 = (v>>7)&1 != 0
	case 1:
		c.ExMem.isCartAccessArm7 = (v>>3)&1 != 0
	}
}

type AuxSpi struct {
	Baudrate uint8
	Hold     bool
	IsBackup bool

	Value    uint8
	Req, Res []uint8
}

func (c *Cartridge) ReadAuxSpi(b uint8) uint8 {
	switch b {
	case 0:
		v := c.AuxSpi.Baudrate
		if c.AuxSpi.Hold {
			v |= 0b100_0000
		}
		return v

	case 1:

		v := uint8(0)
		if c.AuxSpi.IsBackup {
            v |= 1 << 5
		}
		if c.RomTransferIrq {
            v |= 1 << 6
		}
		if c.NDSSlotEnabled {
            v |= 1 << 7
		}

		return v
    default:
        panic("unknown byte cnt auxspi")
	}
}

func (c *Cartridge) WriteAuxSpi(v uint8, b uint8, arm9 bool) {

	a := &c.AuxSpi

	if ok := c.checkAccessRights(arm9); !ok {
		return
	}

	switch b {
	case 0:
		a.Baudrate = v & 0b11
		a.Hold = (v>>6)&1 != 0

		c.Backup.WrittenCnt = true
		return
	case 1:
		// top bits can only be written by arm9
		wasBackup := a.IsBackup
		a.IsBackup = (v>>5)&1 != 0
		c.RomTransferIrq = (v>>6)&1 != 0
		c.NDSSlotEnabled = (v>>7)&1 != 0

		if a.IsBackup && !wasBackup {
			if a.Req == nil {
				a.Req = make([]uint8, 16)
			}
			a.Req = a.Req[:0]
			a.Res = nil
		}

		c.Backup.WrittenCnt = true
		return
	case 2:

		if !a.IsBackup {
			log.Printf("Attempted to Write Data to Rom through AUXSPI.\n")
			return
		}

		c.WriteAuxSpiData(v)
		return
		//case 3:
		// do writes here transfer?
	}
}

func (c *Cartridge) ReadAuxSpiData() uint8 {

    if !c.NDSSlotEnabled {
        return 0
    }

    if !c.AuxSpi.IsBackup {
        return 0
    }

    // if busy return 0

    return c.AuxSpi.Value
}

var match = [4]uint8{
    0x02, 0x26, 0xd, 0x74,
}

func h(res []uint8) bool {

    if len(res) < 4 {
        return false
    }

    for i := range 4 {
        if res[i] != match[i] {
            return false
        }
    }

    return true
}

func (c *Cartridge) WriteAuxSpiData(v uint8) {

    if !c.NDSSlotEnabled {
        return
    }

    if !c.AuxSpi.IsBackup {
        return
    }

    // if busy return 0

	a := &c.AuxSpi

	var value uint8

	if len(a.Res) > 0 {
		value = a.Res[0]
		a.Res = a.Res[1:]
	}

	if len(a.Res) == 0 {
		var stat uint8
		a.Req = append(a.Req, v)

		a.Res, stat = c.Backup.Transfer(a.Req)

        if h(a.Res) {
            panic("pulled valid")
        }

		if stat == STAT_DONE {
			a.Req = a.Req[:0]
		}
	}

	a.Value = value

	if !a.Hold {
		//fmt.Println("FINISH TFX")
	}
}

type RomCtrl struct {
	Key1GapLength         uint32
	Key2EncryptionEnabled bool
	Key2ApplySeed         bool
	Key1Gap2Length        uint32
	Key2EncryptCmds       bool
	isReady               bool
	BlockSizeBits         uint32
	CLKRate               bool
	Key1GapCLK            bool

	RESBRelease bool

	isWrite bool
	Active  bool
	v       uint32

	seed0, seed1 uint64

	BlockSize uint32

	Command [8]uint8
	DataOut uint32
}

func (c *Cartridge) ReadRomCtrl(b uint8) uint8 {

	r := &c.RomCtrl

	switch b {
	case 1:
		return uint8((r.v &^ (1 << 7)) >> (b * 8))
	case 2:
		v := uint8((r.v) >> (b * 8))

		if r.isReady {
            v |= 1 << 7
		} else {
            v &^= 1 << 7
		}

		return v
	}

	return uint8(r.v >> (b * 8))
}

func (c *Cartridge) WriteRomCtrl(v uint8, b uint8, arm9 bool) {

	r := &c.RomCtrl

	if ok := c.checkAccessRights(arm9); !ok {
		return
	}

	// defaults in bios / etc key1 vs normal?

	switch b {
	case 0:
		r.v &^= (0xFF << (b * 8))
		r.v |= (uint32(v) << (b * 8))

		r.Key1GapLength &^= 0xFF
		r.Key1GapLength |= uint32(v)

	case 1:

		r.v &^= (0xFF << (b * 8))
		r.v |= (uint32(v) << (b * 8))
		r.Key1GapLength &= 0xFF
		r.Key1GapLength |= uint32(v) << 8

		r.Key2EncryptionEnabled = (v>>5)&1 != 0
		r.Key2ApplySeed = (v>>7)&1 != 0

		if r.Key2EncryptionEnabled {
			//log.Printf("Key2Encryption is being updated\n")
		}

	case 2:
		r.v &^= (0b0111_1111 << (b * 8))
		r.v |= (uint32(v&0b0111_1111) << (b * 8))

		r.Key1Gap2Length = max(0x18, uint32(v&0x1F))
		r.Key2EncryptCmds = (v>>6)&1 != 0

	case 3:
		r.v &^= (0b1101_1111 << (b * 8))
		r.v |= (uint32(v) << (b * 8))

		r.BlockSizeBits = uint32(v & 0b111)
		r.CLKRate = (v>>3)&1 != 0
		r.Key1GapCLK = (v>>4)&1 != 0

		if !r.RESBRelease {
			r.v |= 1 << 29
			r.RESBRelease = (v>>5)&1 != 0
		}

		r.isWrite = (v>>6)&1 != 0
		r.Active = (v>>7)&1 != 0

		if r.Active {
			c.RunRom(arm9)
		}
	}
}

func (c *Cartridge) WriteCmdOut(v, b uint8, arm9 bool) {

	if ok := c.checkAccessRights(arm9); !ok {
		return
	}

	c.RomCtrl.Command[b] = v
}

func (c *Cartridge) WriteCmdIn(v, b uint8, arm9 bool) {
	if ok := c.checkAccessRights(arm9); !ok {
		return
	}
}
func (c *Cartridge) ReadCmdIn(arm9 bool) uint32 {

	r := &c.RomCtrl
	v := r.DataOut

	if r.isReady {
		r.isReady = false
		c.RomTransfer(false, arm9)

	} else {
		log.Printf("WARNING GAMECARD ROM READ WITHOUT PENDING DATA\n")
	}
	return v

	return r.DataOut
}

func (c *Cartridge) WriteSeed(v, b, seed uint8, arm9 bool) {

	if ok := c.checkAccessRights(arm9); !ok {
		return
	}

	r := &c.RomCtrl

	log.Printf("W SEED    V %02X, B %02X\n", v, b)

	s := &r.seed0

	if seed == 1 {
		s = &r.seed1
	}

	(*s) &^= 0xFF << (b * 8)
	(*s) |= uint64(v) << (b * 8)
}
