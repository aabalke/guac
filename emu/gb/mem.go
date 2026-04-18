package gameboy

import (
	"time"
	"unsafe"

	"github.com/aabalke/guac/emu/gb/cartridge"
)

type MemoryBus struct {
	WRAM [9][0x1000]uint8
	VRAM [2][0x2000]uint8
	OAM  [0x100]uint8

	PROHIBITED [0x60]uint8
	IO         [0x80]uint8
	HRAM       [0x7F]uint8

	ramSaved bool

	VRAMBank   uint8
	WRAMBank   uint8
	WRAMBankV  uint8
	WRAMOffset uint16

	Hdma Hdma
    Oam OamDma

}

type OamDma struct {
    IsActive bool
    Pending, Pending2  bool
    OamValue uint8
    Idx uint16
    Base uint16
}

func (o *OamDma) Read() uint8 {
    return o.OamValue
}

func (o *OamDma) Write(gb *GameBoy, v uint8) {
    o.OamValue = v
    o.Base = uint16(v) << 8
    o.Idx = 0
    o.Pending = true
}

func (o *OamDma) Tick(gb *GameBoy, tcycles int) {

    for range tcycles / 4 {

        if o.Pending {
            o.Pending = false
            o.IsActive = true
            continue
        }

        if !o.IsActive {
            continue
        }

        a := uint8(0)

        // src of this behavior is sameboy - do not see any other ref
        // req mooneye source dma test
        if o.Base >= 0xE000 {
            a = gb.Read((o.Base + o.Idx) &^ 0x2000)
        } else {
            a = gb.Read(o.Base + o.Idx)
        }

        gb.MemoryBus.OAM[o.Idx] = a

        o.Idx++

        if o.Idx >= 0xA0 {
            o.IsActive = false
            o.Pending = false
            return
        }
    }
}

type Hdma struct {
	gb      *GameBoy
	Halted  bool
	Enabled bool
	Length  int
	Src     uint16
	Dst     uint16
	v       uint8
}

func (h *Hdma) Write(v uint8) {

	if terminate := h.Enabled && v&0x80 == 0; terminate {
		h.Enabled = false
		h.v = uint8(h.Length/0x10) - 1
		return
	}

	length := ((uint16(v) & 0x7F) + 1) * 0x10

	if hblank := v&0x80 != 0; hblank {
		h.Length = int(length)
		h.Enabled = true
		return
	}

	//~ 8 normal m cycles per 0x10 transfers
	tcycles := ((8 << h.gb.DoubleSpeedFlag) << 2)

	h.gb.Tick(int(length) * tcycles / 0x10)

	h.Transfer(length)
	h.Length = 0
	h.v = 0xFF
	h.Enabled = false
}

func (h *Hdma) HblankTransfer() {
	// should only be called from enabled check

	tcycles := (8 << h.gb.DoubleSpeedFlag) << 2
	h.gb.Tick(tcycles)

	h.Transfer(0x10)
	if h.Length > 0 {
		h.Length -= 0x10
		h.v = uint8(h.Length/0x10) - 1
		return
	}

	h.Length = 0
	h.v = 0xFF
	h.Enabled = false
}

func (h *Hdma) Transfer(length uint16) {

	src := h.Src & 0xFFF0
	dst := (h.Dst & 0x1FF0) | 0x8000

	for range length {
		b := h.gb.Read(src)
		h.gb.Write(dst, b)
		dst++
		src++
	}

	h.Src = src
	h.Dst = dst
}

func initMemory(gb *GameBoy) {

	gb.Write(0xFF04, 0x1E) // not sur eon this one
	gb.Write(0xFF05, 0x00)
	gb.Write(0xFF06, 0x00)
	gb.Write(0xFF07, 0x00)
    gb.Cpu.IF = 0xE1
	gb.Write(0xFF10, 0x80)
	gb.Write(0xFF11, 0xBF)
	gb.Write(0xFF12, 0xF3)
	gb.Write(0xFF14, 0xBF)
	gb.Write(0xFF16, 0x3F)
	gb.Write(0xFF17, 0x00)
	gb.Write(0xFF19, 0xBF)
	gb.Write(0xFF1A, 0x7F)
	gb.Write(0xFF1B, 0xFF)
	gb.Write(0xFF1C, 0x9F)
	gb.Write(0xFF1E, 0xBF)
	gb.Write(0xFF20, 0xFF)
	gb.Write(0xFF21, 0x00)
	gb.Write(0xFF22, 0x00)
	gb.Write(0xFF23, 0xBF)
	gb.Write(0xFF24, 0x77)
	gb.Write(0xFF25, 0xF3)

	gb.Write(0xFF26, 0xF1)

	gb.Write(0xFF40, 0x91)
	gb.Write(0xFF41, 0x81)
	gb.Write(0xFF42, 0x00)
	gb.Write(0xFF43, 0x00)
	//gb.Write(0xFF44, 0x90)
	gb.Write(0xFF45, 0x00)
	gb.Write(0xFF47, 0xFC)
	gb.Write(0xFF48, 0xFF)
	gb.Write(0xFF49, 0xFF)
	gb.Write(0xFF4A, 0x00)
	gb.Write(0xFF4B, 0x00)
	gb.Write(0xFFFF, 0x00)

	gb.MemoryBus.WRAMBank = 1

	gb.InitSaveLoop()
}

func (gb *GameBoy) SaveRam() {
	if !gb.MemoryBus.ramSaved {
		cartridge.WriteRam(gb.Cartridge.SavPath, gb.Cartridge.RamData)
        gb.Cartridge.Mbc.Save()
		gb.MemoryBus.ramSaved = true
	}
}

func (gb *GameBoy) InitSaveLoop() {

	saveTicker := time.Tick(time.Second)

	go func() {
		for range saveTicker {
			gb.SaveRam()
		}
	}()
}

func (gb *GameBoy) ReadPtr(addr uint16) unsafe.Pointer {

	switch {
	case addr < 0x4000:
		//return gb.Cartridge.Mbc.Read(gb.Cartridge, addr)
		//return gb.Cartridge.Mbc.ReadPtr(gb.Cartridge, addr)
		return nil
	case addr < 0x8000:
		//return gb.Cartridge.Mbc.ReadRomPtr(gb.Cartridge, addr)
		return nil
		//return gb.Cartridge.Mbc.ReadRom(gb.Cartridge, addr)
	case addr < 0xA000:

		//if gb.Color {
		//	offset := uint16(gb.MemoryBus.VRAMBank) * 0x2000
		//	return gb.MemoryBus.VRAM[addr-0x8000+offset]
		//}
		//return gb.MemoryBus.VRAM[addr-0x8000]

		return nil

	case addr < 0xC000:
		//return gb.Cartridge.Mbc.ReadRam(gb.Cartridge, addr)
		return nil
	case addr < 0xD000:

		addr &= 0xFFF

		if addr+3 >= uint16(len(gb.MemoryBus.WRAM[0])) {
			return nil
		}

		return unsafe.Add(unsafe.Pointer(&gb.MemoryBus.WRAM[0]), addr)

	case addr < 0xE000:

		addr &= 0xFFF

		if addr+3 >= uint16(len(gb.MemoryBus.WRAM[gb.MemoryBus.WRAMBank])) {
			return nil
		}

		return unsafe.Add(unsafe.Pointer(&gb.MemoryBus.WRAM[gb.MemoryBus.WRAMBank]), addr)

	case addr < 0xFF80:
		return nil

	case addr < 0xFFFE:
		return unsafe.Pointer(&gb.MemoryBus.HRAM[addr-0xFF80])

	default:
		return nil
	}
}

func (gb *GameBoy) Read(addr uint16) uint8 {

	switch {
	case addr < 0x4000:
		return gb.Cartridge.Mbc.Read(addr)
	case addr < 0x8000:

		return gb.Cartridge.Mbc.Read(addr)
	case addr < 0xA000:

		if drawing := gb.Stat.Mode == 3; drawing {
			return 0xFF
		}

		return gb.MemoryBus.VRAM[gb.MemoryBus.VRAMBank][addr&0x1FFF]

	case addr < 0xC000:
		return gb.Cartridge.Mbc.Read(addr)
	case addr < 0xD000:
		return gb.MemoryBus.WRAM[0][addr&0xFFF]
	case addr < 0xE000:
		return gb.MemoryBus.WRAM[gb.MemoryBus.WRAMBank][addr&0xFFF]
	case addr < 0xFE00:
		return gb.MemoryBus.WRAM[0][addr&0xFFF]
	case addr < 0xFEA0:

		if notAvailable := gb.Stat.Mode&0b10 != 0; notAvailable {
			return 0xFF
		}

        if dma := gb.MemoryBus.Oam.IsActive; dma {
			return 0xFF
        }

		return gb.MemoryBus.OAM[addr-0xFE00]
	case addr < 0xFF00:
		return gb.MemoryBus.PROHIBITED[addr-0xFEA0]
	case addr < 0xFF80:
		return gb.ReadIO(addr)
	case addr < 0xFFFF:
		return gb.MemoryBus.HRAM[addr-0xFF80]
	case addr == 0xFFFF:
		return gb.Cpu.IE
	default:
		panic("not possible read")
	}
}

func (gb *GameBoy) Write(addr uint16, v uint8) {

	//if addr == 0xD880 { // test addr for blargg
	//    fmt.Printf("\nTest %02d started...\n", v)
	//    debug.B[4] = true
	//}

	switch {
	case addr < 0x8000:
		gb.Cartridge.Mbc.Write(addr, v)
		gb.Cpu.isBranching = true
	case addr < 0xA000:

		if drawing := gb.Stat.Mode == 3; drawing {
			return
		}

		gb.MemoryBus.VRAM[gb.MemoryBus.VRAMBank][addr&0x1FFF] = v

	case addr < 0xC000:
		gb.Cartridge.Mbc.Write(addr, v)
		gb.MemoryBus.ramSaved = false
	case addr < 0xD000:
		gb.MemoryBus.WRAM[0][addr&0xFFF] = v
	case addr < 0xE000:
		gb.MemoryBus.WRAM[gb.MemoryBus.WRAMBank][addr&0xFFF] = v
	case addr < 0xFE00:
		gb.MemoryBus.WRAM[0][addr&0xFFF] = v
	case addr < 0xFEA0:
		if notAvailable := gb.Stat.Mode&0b10 != 0; notAvailable {
			return
		}

		gb.MemoryBus.OAM[addr-0xFE00] = v
	case addr < 0xFF00:
		gb.MemoryBus.PROHIBITED[addr-0xFEA0] = v
	case addr < 0xFF80:
		gb.WriteIO(addr, v)
	case addr < 0xFFFF:
		gb.MemoryBus.HRAM[addr-0xFF80] = v
	case addr == 0xFFFF:
		gb.Cpu.IE = v

	default:
		panic("not possible write")
	}
}

func (gb *GameBoy) ReadIO(addr uint16) uint8 {

	if addr >= 0xFF10 && addr < 0xFF40 {
		return gb.ReadSound(uint8(addr), gb.Apu)
	}

	switch addr {
	case 0xFF00:
		return gb.getJoypad() | 0xC0

	case 0xFF04: // DIV
		return uint8(gb.Timer.Div >> 8)

	case 0xFF05:
		return gb.Timer.TIMA

	case 0xFF06:
		return gb.Timer.TMA

	case 0xFF07:
		v := gb.Timer.FreqBits | 0xF8
		if gb.Timer.Enabled {
			v |= 4
		}

		return v

	case 0xFF0F:
		return gb.Cpu.IF | 0xE0

	case 0xFF40:
		return gb.Lcdc.Read()

	case 0xFF41:
		return gb.Stat.Read()

	case 0xFF44:
		return gb.MemoryBus.IO[uint8(addr)]

    case 0xFF46:
        return gb.MemoryBus.Oam.Read()

	case 0xFF4F:
		return gb.MemoryBus.VRAMBank | 0xFE

	case 0xFF55:
		return gb.MemoryBus.Hdma.v

	case 0xFF68:

		if gb.Color {
			return gb.bgPalette.Idx
		}

		return 0

	case 0xFF69:

		if gb.Color {
			return gb.bgPalette.Palette[gb.bgPalette.Idx]
		}

		return 0
	case 0xFF6A:

		if gb.Color {
			return gb.spPalette.Idx
		}

		return 0

	case 0xFF6B:

		if gb.Color {
			return gb.spPalette.Palette[gb.spPalette.Idx]
		}

		return 0

	case 0xFF4D:

		b := uint8(gb.DoubleSpeedFlag << 7)

		if gb.PrepareSpeedToggle {
			b |= 1
		}

		return b

	case 0xFF70:
		return gb.MemoryBus.WRAMBank
	default:

		return gb.MemoryBus.IO[uint8(addr)]
	}
}

func (gb *GameBoy) WriteIO(addr uint16, v uint8) {

	io := &gb.MemoryBus.IO

	if addr >= 0xFF10 && addr < 0xFF40 {
		gb.WriteSound(uint8(addr), v, gb.Apu)
		return
	}

	switch addr {
	case 0xFF04: // DIV

		t := &gb.Timer
		prevDiv := t.Div
		t.Div = 0

		mask := uint16(1 << 12)
		mask <<= uint16(gb.DoubleSpeedFlag)

		if prevDiv&mask != 0 {
			gb.Apu.ClockFrameSequencer()
		}

		if !t.Enabled {
			return
		}

		if prevDiv&fallingEdgeBits[t.FreqBits] != 0 {
			if overflow := t.TIMA == 0xFF; overflow {
				t.TIMA = t.TMA
				gb.SetIrq(IRQ_TMR)
				return
			}

			t.TIMA++
		}

	case 0xFF05:

		gb.Timer.PendingOverflow = false
		if !gb.Timer.BCycle {
			gb.Timer.TIMA = v
		}

	case 0xFF06:
		gb.Timer.TMA = v

		if gb.Timer.BCycle {
			gb.Timer.TIMA = v
		}

	case 0xFF07:
		t := &gb.Timer

		if gb.Timer.FreqBits != v&3 {
			gb.Timer.FreqBits = v & 3
		}

		wasEnabled := gb.Timer.Enabled
		gb.Timer.Enabled = v&4 != 0

		if !wasEnabled || gb.Timer.Enabled {
			return
		}

		if gb.Timer.Div&fallingEdgeBits[t.FreqBits] != 0 {
			if overflow := t.TIMA == 0xFF; overflow {
				t.TIMA = t.TMA
				gb.SetIrq(IRQ_TMR)
				return
			}

			t.TIMA++
		}

	case 0xFF0F:
		gb.Cpu.IF = v & 0x1F

	case 0xFF40:
		gb.Lcdc.Write(v)

	case 0xFF41:
		gb.Stat.Write(v)

	case 0xFF44:
		io[uint8(addr)] = v

	case 0xFF46: // DMA

		if gb.Stat.Mode&0b10 != 0 {
			return
		}

        gb.MemoryBus.Oam.Write(gb, v)

	case 0xFF47: // bgpalette mono

		gb.UnpackedMonoPals[0][0] = *(*uint32)(unsafe.Pointer(&gb.Palette[(v>>0)&3][0])) | 0xFF00_0000
		gb.UnpackedMonoPals[0][1] = *(*uint32)(unsafe.Pointer(&gb.Palette[(v>>2)&3][0])) | 0xFF00_0000
		gb.UnpackedMonoPals[0][2] = *(*uint32)(unsafe.Pointer(&gb.Palette[(v>>4)&3][0])) | 0xFF00_0000
		gb.UnpackedMonoPals[0][3] = *(*uint32)(unsafe.Pointer(&gb.Palette[(v>>6)&3][0])) | 0xFF00_0000
		io[uint8(addr)] = v

	case 0xFF48: // objpalette mono

		//gb.UnpackedMonoPals[1][0] = *(*uint32)(unsafe.Pointer(&gb.Palette[(v>>0) & 3][0])) | 0xFF00_0000
		gb.UnpackedMonoPals[1][1] = *(*uint32)(unsafe.Pointer(&gb.Palette[(v>>2)&3][0])) | 0xFF00_0000
		gb.UnpackedMonoPals[1][2] = *(*uint32)(unsafe.Pointer(&gb.Palette[(v>>4)&3][0])) | 0xFF00_0000
		gb.UnpackedMonoPals[1][3] = *(*uint32)(unsafe.Pointer(&gb.Palette[(v>>6)&3][0])) | 0xFF00_0000
		io[uint8(addr)] = v

	case 0xFF49: // objpalette mono

		//gb.UnpackedMonoPals[2][0] = *(*uint32)(unsafe.Pointer(&gb.Palette[(v>>0) & 3][0])) | 0xFF00_0000
		gb.UnpackedMonoPals[2][1] = *(*uint32)(unsafe.Pointer(&gb.Palette[(v>>2)&3][0])) | 0xFF00_0000
		gb.UnpackedMonoPals[2][2] = *(*uint32)(unsafe.Pointer(&gb.Palette[(v>>4)&3][0])) | 0xFF00_0000
		gb.UnpackedMonoPals[2][3] = *(*uint32)(unsafe.Pointer(&gb.Palette[(v>>6)&3][0])) | 0xFF00_0000
		io[uint8(addr)] = v

	case 0xFF4D:
		if gb.Color {
			gb.PrepareSpeedToggle = v&1 != 0
			io[0x4D] &= 0x80
			io[0x4D] |= v & 1
		}
		io[uint8(addr)] = v

	case 0xFF4F:
		if gb.Color && !gb.MemoryBus.Hdma.Enabled {
			gb.MemoryBus.VRAMBank = v & 0x1
			return
		}
		io[uint8(addr)] = v

	case 0xFF51:
		if gb.Color {
			gb.MemoryBus.Hdma.Src &= 0xFF
			gb.MemoryBus.Hdma.Src |= uint16(v) << 8
		}

	case 0xFF52:
		if gb.Color {
			gb.MemoryBus.Hdma.Src &^= 0xFF
			gb.MemoryBus.Hdma.Src |= uint16(v)
		}

	case 0xFF53:
		if gb.Color {
			gb.MemoryBus.Hdma.Dst &= 0xFF
			gb.MemoryBus.Hdma.Dst |= uint16(v) << 8
		}

	case 0xFF54:
		if gb.Color {
			gb.MemoryBus.Hdma.Dst &^= 0xFF
			gb.MemoryBus.Hdma.Dst |= uint16(v)
		}

	case 0xFF55:
		if gb.Color {
			gb.MemoryBus.Hdma.Write(v)
		}
		io[uint8(addr)] = v

	case 0xFF68:
		if gb.Color {
			gb.bgPalette.Idx = v & 0b111111
			gb.bgPalette.Inc = (v>>7)&1 != 0
		}
		io[uint8(addr)] = v

	case 0xFF69:
		if gb.Color {
			gb.bgPalette.Palette[gb.bgPalette.Idx] = v
			gb.bgPalette.update(gb.bgPalette.Idx)

			if gb.bgPalette.Inc {
				gb.bgPalette.Idx = (gb.bgPalette.Idx + 1) & 0b111111
			}
		}
		io[uint8(addr)] = v

	case 0xFF6A:
		if gb.Color {
			gb.spPalette.Idx = v & 0b111111
			gb.spPalette.Inc = (v>>7)&1 != 0
		}
		io[uint8(addr)] = v

	case 0xFF6B:
		if gb.Color {
			gb.spPalette.Palette[gb.spPalette.Idx] = v
			gb.spPalette.update(gb.spPalette.Idx)

			if gb.spPalette.Inc {
				gb.spPalette.Idx = (gb.spPalette.Idx + 1) & 0b111111
			}
		}

		io[uint8(addr)] = v

	case 0xFF70:
		if gb.Color {
			gb.MemoryBus.WRAMBank = v & 7
			if v&7 == 0 {
				gb.MemoryBus.WRAMBank = 1
			}

			gb.MemoryBus.WRAMBankV = v & 7
		}

		io[uint8(addr)] = v
	default:

		io[uint8(addr)] = v
	}
}
