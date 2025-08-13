package mem

import (
	_ "embed"

	"github.com/aabalke/guac/emu/nds/cart"
	"github.com/aabalke/guac/emu/nds/cpu"
	"github.com/aabalke/guac/emu/nds/ppu"
)

//go:embed res/bios7.bin
var arm7Bios []byte

//go:embed res/bios9.bin
var arm9Bios []byte

type Mem struct {
	MainRam       [0x40_0000]uint8
	WRAM          WRAM
	TCMCache      [0xF000]uint8
    Vram VRAM
	OAMPal        [0x1000]uint8
	//ThreeMemory   [0x3_E000]uint8
	//WifiRam       [0x2000]uint8
	//FirmwareFlash [0x4_0000]uint8
	//BiosRom       [0x9000]uint8

	Arm7Bios [0x4000]uint8
	Arm9Bios [0x1000]uint8

	IO [0x100_0000]uint8

	arm9Irq *cpu.Irq
	arm7Irq *cpu.Irq
    ppu *ppu.PPU

    exmem ExMem
    auxspi AuxSPI
    div Div
    sqrt Sqrt

	ITCM [0x8000]uint8
	DTCM [0x4000]uint8

	TmpTCM [0x200_0000]uint8

    Cartridge *cart.Cartridge

    Vcount uint32
    Dispstat Dispstat
    Keypad Keypad

    WramArm7 [0x1_0000]uint8
}

//  4096KB Main RAM (8192KB in debug version)
//  96KB   WRAM (64K mapped to NDS7, plus 32K mappable to NDS7 or NDS9)
//  60KB   TCM/Cache (TCM: 16K Data, 32K Code) (Cache: 4K Data, 8K Code)
//  656KB  VRAM (allocateable as BG/OBJ/2D/3D/Palette/Texture/WRAM memory)
//  4KB    OAM/PAL (2K OBJ Attribute Memory, 2K Standard Palette RAM)
//  248KB  Internal 3D Memory (104K Polygon RAM, 144K Vertex RAM)
//  ?KB    Matrix Stack, 48 scanline cache
//  8KB    Wifi RAM
//  256KB  Firmware FLASH (512KB in iQue variant, with chinese charset)
//  36KB   BIOS ROM (4K NDS9, 16K NDS7, 16K GBA)

func NewMemory(irq *cpu.Irq, c *cart.Cartridge, ppu *ppu.PPU) Mem {
	m := Mem{
        arm9Irq: irq,
        Cartridge: c,
        ppu: ppu,
    }

    m.Keypad.KEYINPUT = 0x3FF

	m.LoadBios()

	return m
}

func (mem *Mem) LoadBios() {

	for i := range len(arm7Bios) {
		mem.Arm7Bios[i] = uint8(arm7Bios[i])
	}

	for i := range len(arm9Bios) {
		mem.Arm9Bios[i] = uint8(arm9Bios[i])
	}
}

func (mem *Mem) Read(addr uint32, arm9 bool) uint8 {

	if arm9 {

		if addr > 0xFF00_0000 {
			return mem.Arm9Bios[addr&0x0FFF]
		}

		switch addr >> 24 {
		case 0x0:

			return mem.TmpTCM[addr]

			if addr < 0x8000 {
				return mem.ITCM[addr]
			}

			return 0 // tcm
		case 0x1:
			return mem.TmpTCM[addr]
			return 0 // tcm
		case 0x2:
			return mem.MainRam[addr&0x3F_FFFF]
		case 0x3:
            return mem.WRAM.Read(addr, true)
		case 0x4:
			return mem.ReadArm9IO(addr - 0x400_0000)
		case 0x5:
			return 0 // pal std
		case 0x6:
            return mem.Vram.Read(addr, true)
		case 0x7:
			return 0 // oam
		case 0x8, 0x9:
			return 0 // gba rom
		case 0xA, 0xB, 0xC, 0xD, 0xE:
			return 0 // gba ram
		}

        return 0
	}
    switch addr >> 24 {
    case 0x0, 0x1:
        if addr < uint32(len(mem.Arm7Bios)) {
            return mem.Arm7Bios[addr]
        }

        return 0

    case 0x2:
			return mem.MainRam[addr&0x3F_FFFF]
    case 0x3:
        if addr < 0x380_0000 {
            return mem.WRAM.Read(addr, false)
        } else {
            return mem.WramArm7[(addr - 0x380_0000) % 0x1_0000]
        }
    case 0x4:
        //mem.WriteArm9IO(addr-0x400_0000, v) //io
    case 0x5: // pal std
    case 0x6:
        //mem.Vram.Write(addr, v, true)
    }

	return 0
}

func (mem *Mem) Read8(addr uint32, arm9 bool) uint32 {
	return uint32(mem.Read(addr, arm9))
}
func (mem *Mem) Read16(addr uint32, arm9 bool) uint32 {
	return uint32(mem.Read(addr, arm9)) | (uint32(mem.Read(addr+1, arm9)) << 8)
}
func (mem *Mem) Read32(addr uint32, arm9 bool) uint32 {
	a := uint32(mem.Read(addr+2, arm9)) | (uint32(mem.Read(addr+3, arm9)) << 8)
	b := uint32(mem.Read(addr, arm9)) | (uint32(mem.Read(addr+1, arm9)) << 8)

	return (a << 16) | b
}

func (mem *Mem) Write(addr uint32, v uint8, arm9 bool) {

	if arm9 {

		switch addr >> 24 {
		case 0x0: // tcm

			mem.TmpTCM[addr] = v
		case 0x1: // tcm
			mem.TmpTCM[addr] = v
		case 0x2:
			mem.MainRam[addr&0x3F_FFFF] = v
		case 0x3:
            mem.WRAM.Write(addr, v, true)
		case 0x4:
			mem.WriteArm9IO(addr-0x400_0000, v) //io
		case 0x5: // pal std
		case 0x6:
            mem.Vram.Write(addr, v, true)
		case 0x7: // oam
		case 0x8, 0x9: // gba rom
		case 0xA, 0xB, 0xC, 0xD, 0xE: // gba ram
		case 0xF:
		}

        return
	}

    switch addr >> 24 {
    case 0x2:
        mem.MainRam[addr&0x3F_FFFF] = v
    case 0x3:
        if addr < 0x380_0000 {
            mem.WRAM.Write(addr, v, false)
        } else {
            mem.WramArm7[(addr - 0x380_0000) % 0x1_0000] = v
        }
    case 0x4:
        mem.WriteArm9IO(addr-0x400_0000, v) //io
    case 0x5: // pal std
    case 0x6:
        mem.Vram.Write(addr, v, true)
    }
}

func (mem *Mem) Write8(addr uint32, v uint8, arm9 bool) {
	mem.Write(addr, v, arm9)
}
func (mem *Mem) Write16(addr uint32, v uint16, arm9 bool) {
	mem.Write(addr, uint8(v), arm9)
	mem.Write(addr+1, uint8(v>>8), arm9)

}
func (mem *Mem) Write32(addr uint32, v uint32, arm9 bool) {
	mem.Write(addr, uint8(v), arm9)
	mem.Write(addr+1, uint8(v>>8), arm9)
	mem.Write(addr+2, uint8(v>>16), arm9)
	mem.Write(addr+3, uint8(v>>24), arm9)

}

func (mem *Mem) ReadArm9IO(addr uint32) uint8 {

	//if addr != 0x180 && addr != 0x181 && addr < 0x3000 {
	//	fmt.Printf("READ ADDR %08X\n", addr)
	//}

    if addr >= 0x280 && addr < 0x2B0 {
        return mem.div.Read(addr)
    } else if addr >= 0x2B0 && addr < 0x2C0 {
        return mem.sqrt.Read(addr)
    }

	switch addr {
	case 0x4:
		return uint8(mem.Dispstat)
	case 0x5:
		return uint8(mem.Dispstat >> 8)
    case 0x6:
        return uint8(mem.Vcount)
    case 0x7:
        return uint8(mem.Vcount >> 8)

	case 0x130:
		return mem.Keypad.readINPUT(false)
	case 0x131:
		return mem.Keypad.readINPUT(true)
	case 0x132:
		return mem.Keypad.readCNT(false)
	case 0x133:
		return mem.Keypad.readCNT(true)
    case 0x1A0:
        return mem.auxspi.Read(0)
    case 0x1A1:
        return mem.auxspi.Read(1)
    case 0x1A2:
        return mem.auxspi.Read(2)
    case 0x1A3:
        return mem.auxspi.Read(3)

    case 0x204:
        return mem.exmem.Read(0)
    case 0x205:
        return mem.exmem.Read(1)
	case 0x208:
		return mem.arm9Irq.ReadIME()
	case 0x210:
		return mem.arm9Irq.ReadIE(0)
	case 0x211:
		return mem.arm9Irq.ReadIE(1)
	case 0x212:
		return mem.arm9Irq.ReadIE(2)
	case 0x213:
		return mem.arm9Irq.ReadIE(3)
	case 0x214:
		return mem.arm9Irq.ReadIF(0)
	case 0x215:
		return mem.arm9Irq.ReadIF(1)
	case 0x216:
		return mem.arm9Irq.ReadIF(2)
	case 0x217:
		return mem.arm9Irq.ReadIF(3)
    case 0x247:
        return mem.WRAM.ReadCNT()
	default:
		return mem.IO[addr]
	}
}

func (mem *Mem) WriteArm9IO(addr uint32, v uint8) {

	//if addr != 0x180 && addr != 0x181 && addr < 0x3000 {
	//	fmt.Printf("WRITE ADDR %08X V %02X\n", addr, v)
	//}

    if addr < 0x4 {
        mem.ppu.Update(addr, uint32(v))
    }

    if addr >= 0x240 && addr < 0x250 {
        //mem.Vram.WriteCNT(addr, v)
        //return
    } else if addr >= 0x280 && addr < 0x2B0 {
        mem.div.Write(addr, v)
        return
    } else if addr >= 0x2B0 && addr < 0x2C0 {
        mem.sqrt.Write(addr, v)
        return
    }

	switch addr {
    case 0x4:
		mem.Dispstat.Write(v, false)
    case 0x5:
		mem.Dispstat.Write(v, true)

    case 0x6:
        mem.Vcount &^= 0xFF
        mem.Vcount |= uint32(v)
    case 0x7:
        mem.Vcount &= 0xFF
        mem.Vcount |= uint32(v) << 8

	case 0x130:
		return
	case 0x131:
		return
	case 0x132:
		mem.Keypad.writeCNT(v, false)
	case 0x133:
		mem.Keypad.writeCNT(v, true)
    case 0x1A0:
        mem.auxspi.Write(v, 0)
    case 0x1A1:
        mem.auxspi.Write(v, 1)
    case 0x1A2:
        mem.auxspi.WriteData(v, 0)
    case 0x1A3:
        mem.auxspi.WriteData(v, 1)

    case 0x204:
        mem.exmem.Write(v, 0)
    case 0x205:
        mem.exmem.Write(v, 1)
	case 0x208:
		mem.arm9Irq.WriteIME(v)
	case 0x210:
		mem.arm9Irq.WriteIE(v, 0)
	case 0x211:
		mem.arm9Irq.WriteIE(v, 1)
	case 0x212:
		mem.arm9Irq.WriteIE(v, 2)
	case 0x213:
		mem.arm9Irq.WriteIE(v, 3)
	case 0x214:
		mem.arm9Irq.WriteIF(v, 0)
	case 0x215:
		mem.arm9Irq.WriteIF(v, 1)
	case 0x216:
		mem.arm9Irq.WriteIF(v, 2)
	case 0x217:
		mem.arm9Irq.WriteIF(v, 3)
    case 0x247:
        mem.WRAM.WriteCNT(v)
	default:
		mem.IO[addr] = v
	}
}

// this is temp while I get gamecard transfers and spi figured out
func (mem *Mem) DirtyTransfer() {

    h := &mem.Cartridge.Header

    for i := range h.Arm9Size {
        v := mem.Cartridge.Rom[h.Arm9Offset + i]
        mem.Write(h.Arm9RamAddr + i, v, true)
    }

    for i := range h.Arm7Size {
        v := mem.Cartridge.Rom[h.Arm7Offset + i]
        mem.Write(h.Arm7RamAddr + i, v, false)
    }
}
