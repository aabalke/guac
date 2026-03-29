package gameboy

import (
	"fmt"
	"time"
	"unsafe"

	"github.com/aabalke/guac/emu/gb/cartridge"
	"github.com/aabalke/guac/emu/gb/debug"
)

type MemoryBus struct {
	WRAM   [0x9000]uint8
	VRAM   [0x4000]uint8
	OAM    [0x100]uint8

    PROHIBITED [0x60]uint8
    IO         [0x80]uint8
	HRAM       [0x7F]uint8

	ramSaved bool

	VRAMBank uint8
	WRAMBank uint8

	HdmaLength uint8
	HdmaActive bool
}

func initMemory(gb *GameBoy) {

	gb.Write(0xFF04, 0x1E) // not sur eon this one
	gb.Write(0xFF05, 0x00)
	gb.Write(0xFF06, 0x00)
	gb.Write(0xFF07, 0x00)
	gb.Write(0xFF0F, 0xE1)
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
	gb.Write(0xFF42, 0x00)
	gb.Write(0xFF43, 0x00)
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
		return gb.Cartridge.Mbc.ReadPtr(gb.Cartridge, addr)
	case addr < 0x8000:
		return gb.Cartridge.Mbc.ReadRomPtr(gb.Cartridge, addr)
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

		addr -= 0xC000

		if addr+3 >= uint16(len(gb.MemoryBus.WRAM)) {
			return nil
		}

		return unsafe.Add(unsafe.Pointer(&gb.MemoryBus.WRAM), addr)

	case addr < 0xE000:

		addr = (addr - 0xC000) + (uint16(gb.MemoryBus.WRAMBank) * 0x1000)

		if addr+3 >= uint16(len(gb.MemoryBus.WRAM)) {
			return nil
		}

		return unsafe.Add(unsafe.Pointer(&gb.MemoryBus.WRAM), addr)

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
		return gb.Cartridge.Mbc.Read(gb.Cartridge, addr)
	case addr < 0x8000:
		return gb.Cartridge.Mbc.ReadRom(gb.Cartridge, addr)
	case addr < 0xA000:
		if gb.Color {
			offset := uint16(gb.MemoryBus.VRAMBank) * 0x2000
			return gb.MemoryBus.VRAM[addr-0x8000+offset]
		}
		return gb.MemoryBus.VRAM[addr-0x8000]

	case addr < 0xC000:
		return gb.Cartridge.Mbc.ReadRam(gb.Cartridge, addr)
	case addr < 0xD000:
		return gb.MemoryBus.WRAM[addr-0xC000]
	case addr < 0xE000:
		return gb.MemoryBus.WRAM[(addr-0xC000)+(uint16(gb.MemoryBus.WRAMBank)*0x1000)]
	case addr < 0xFE00:
		return gb.Cartridge.Mbc.ReadRam(gb.Cartridge, addr-0x2000)
	case addr < 0xFEA0:
		return gb.MemoryBus.OAM[addr-0xFE00]
	case addr < 0xFF00:
        return gb.MemoryBus.PROHIBITED[addr - 0xFEA0]
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

    if addr == 0xD880 && v == 3 {
        debug.B[1] = true
        fmt.Printf("Test %02d started...\n", v)
    }

	switch {
	case addr < 0x8000:
		gb.Cartridge.Mbc.Handle(addr, v)
		gb.Cpu.isBranching = true
	case addr < 0xA000:
		var offset uint16
		if gb.Color {
			offset = uint16(gb.MemoryBus.VRAMBank) * 0x2000
		}
		gb.MemoryBus.VRAM[addr-0x8000+offset] = v
	case addr < 0xC000:
		gb.Cartridge.Mbc.WriteRam(gb.Cartridge, addr, v)
		gb.MemoryBus.ramSaved = false
	case addr < 0xD000:
		gb.MemoryBus.WRAM[addr-0xC000] = v
	case addr < 0xE000:
		gb.MemoryBus.WRAM[(addr-0xC000)+(uint16(gb.MemoryBus.WRAMBank)*0x1000)] = v
	case addr < 0xFE00:
		gb.MemoryBus.WRAM[addr-0xE000] = v
	case addr < 0xFEA0:
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

func (gb *GameBoy) cgbDMATransfer(byte uint8) {

	if gb.MemoryBus.HdmaActive && !((byte>>7)&1 != 0) {
		gb.MemoryBus.HdmaActive = false
		gb.MemoryBus.IO[0x55] |= 0x80
		return
	}

	length := ((uint16(byte) & 0x7F) + 1) * 0x10

	if !((byte>>7)&1 != 0) {

		gb.performDMATransfer(length)
		gb.MemoryBus.IO[0x55] = 0xFF
		return
	}

	gb.MemoryBus.HdmaLength = byte
	gb.MemoryBus.HdmaActive = true
}

func (gb *GameBoy) hdmaTransfer() {

	MemBus := &gb.MemoryBus

	if !MemBus.HdmaActive {
		return
	}

	gb.performDMATransfer(0x10)
	if MemBus.HdmaLength > 0 {
		MemBus.HdmaLength--
		MemBus.IO[0x55] = MemBus.HdmaLength
		return
	}

	MemBus.IO[0x55] = 0xFF
	MemBus.HdmaActive = false
}

func (gb *GameBoy) performDMATransfer(length uint16) {

    io := &gb.MemoryBus.IO

	src := (uint16(io[0x51])<<8 | uint16(io[0x52])) & 0xFFF0
	dst := (uint16(io[0x53])<<8 | uint16(io[0x54])) & 0x1FF0
	dst += 0x8000

	for range length {
		b := gb.Read(src)
		gb.Write(dst, b)
		dst++
		src++
	}

	io[0x51] = uint8(src >> 8)
	io[0x52] = uint8(src & 0xFF)
	io[0x53] = uint8(dst >> 8)
	io[0x54] = uint8(dst & 0xF0)
}

func (gb *GameBoy) ReadIO(addr uint16) uint8 {

    if addr >= 0xFF10 && addr < 0xFF40 {
		return gb.ReadSound(uint32(addr&0xFF), gb.Apu)
    }

	switch addr {
	case 0xFF00:
		return gb.getJoypad()

	case 0xFF04: // DIV

        return uint8(gb.Timer.Div>>8)


	case 0xFF0F:
		return gb.Cpu.IF

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

		var b uint8 = 0

		if gb.DoubleSpeed {
			b |= 1 << 7
		}

		if gb.PrepareSpeedToggle {
			b |= 1
		}

		return b

	case 0xFF4F:
		return gb.MemoryBus.VRAMBank
	case 0xFF70:
		return gb.MemoryBus.WRAMBank
    default:
	    return gb.MemoryBus.IO[addr - 0xFF00]
	}
}

func (gb *GameBoy) WriteIO(addr uint16, v uint8) {

    io := &gb.MemoryBus.IO

    if addr >= 0xFF10 && addr < 0xFF40 {
		gb.WriteSound(uint32(addr&0xFF), v, gb.Apu)
        return
    }

	switch addr {
	case 0xFF04: // DIV
		gb.Timer.Counter = 0
        prev := gb.Timer.Div
		gb.Timer.Div = 0

        mask := uint16(1<<12)
        if gb.DoubleSpeed {
            mask = uint16(1<<13)
        }

        if prev & mask != 0 {
            gb.Apu.ClockFrameSequencer()
        } 

	case 0xFF07:
		currFreq := io[0x07] & 0x03
		io[0x07] = v | 0xF8
		newFreq := io[0x07] & 0x03

		if currFreq != newFreq {
			gb.Timer.Counter = 0
		}
        io[addr-0xFF00] = v

	case 0xFF0F:
		gb.Cpu.IF = v
        io[addr-0xFF00] = v

	case 0xFF44:
		io[0x44] = 0
        io[addr-0xFF00] = v

	case 0xFF46: // DMA
		address := uint16(v) << 8
        for i := range uint16(0xA0) {
            gb.Tick(4) // 160 m cycles = 4 tcycles per transfer
			a := gb.Read(address + i)

            gb.MemoryBus.OAM[i] = a
		}
        //io[addr-0xFF00] = v

	case 0xFF47: // bgpalette mono

		gb.UnpackedMonoPals[0][0] = *(*uint32)(unsafe.Pointer(&gb.Palette[(v>>0)&3][0])) | 0xFF00_0000
		gb.UnpackedMonoPals[0][1] = *(*uint32)(unsafe.Pointer(&gb.Palette[(v>>2)&3][0])) | 0xFF00_0000
		gb.UnpackedMonoPals[0][2] = *(*uint32)(unsafe.Pointer(&gb.Palette[(v>>4)&3][0])) | 0xFF00_0000
		gb.UnpackedMonoPals[0][3] = *(*uint32)(unsafe.Pointer(&gb.Palette[(v>>6)&3][0])) | 0xFF00_0000
        io[addr-0xFF00] = v

	case 0xFF48: // objpalette mono

		//gb.UnpackedMonoPals[1][0] = *(*uint32)(unsafe.Pointer(&gb.Palette[(v>>0) & 3][0])) | 0xFF00_0000
		gb.UnpackedMonoPals[1][1] = *(*uint32)(unsafe.Pointer(&gb.Palette[(v>>2)&3][0])) | 0xFF00_0000
		gb.UnpackedMonoPals[1][2] = *(*uint32)(unsafe.Pointer(&gb.Palette[(v>>4)&3][0])) | 0xFF00_0000
		gb.UnpackedMonoPals[1][3] = *(*uint32)(unsafe.Pointer(&gb.Palette[(v>>6)&3][0])) | 0xFF00_0000
        io[addr-0xFF00] = v

	case 0xFF49: // objpalette mono

		//gb.UnpackedMonoPals[2][0] = *(*uint32)(unsafe.Pointer(&gb.Palette[(v>>0) & 3][0])) | 0xFF00_0000
		gb.UnpackedMonoPals[2][1] = *(*uint32)(unsafe.Pointer(&gb.Palette[(v>>2)&3][0])) | 0xFF00_0000
		gb.UnpackedMonoPals[2][2] = *(*uint32)(unsafe.Pointer(&gb.Palette[(v>>4)&3][0])) | 0xFF00_0000
		gb.UnpackedMonoPals[2][3] = *(*uint32)(unsafe.Pointer(&gb.Palette[(v>>6)&3][0])) | 0xFF00_0000
        io[addr-0xFF00] = v

	case 0xFF4D:
		if gb.Color {
			gb.PrepareSpeedToggle = v & 1 != 0
            io[0x4D] &= 0x80
            io[0x4D] |= v & 1
		}
        io[addr-0xFF00] = v

	case 0xFF4F:
		if gb.Color && !gb.MemoryBus.HdmaActive {
			gb.MemoryBus.VRAMBank = v & 0x1
		}
        io[addr-0xFF00] = v

	case 0xFF55:
		if gb.Color {
			gb.cgbDMATransfer(v)
		}
        io[addr-0xFF00] = v

	case 0xFF68:
		if gb.Color {
			gb.bgPalette.Idx = v & 0b111111
			gb.bgPalette.Inc = (v>>7)&1 != 0
		}
        io[addr-0xFF00] = v

	case 0xFF69:
		if gb.Color {
			gb.bgPalette.Palette[gb.bgPalette.Idx] = v
			gb.bgPalette.update(gb.bgPalette.Idx)

			if gb.bgPalette.Inc {
				gb.bgPalette.Idx = (gb.bgPalette.Idx + 1) & 0b111111
			}
		}
        io[addr-0xFF00] = v

	case 0xFF6A:
		if gb.Color {
			gb.spPalette.Idx = v & 0b111111
			gb.spPalette.Inc = (v>>7)&1 != 0
		}
        io[addr-0xFF00] = v

	case 0xFF6B:
		if gb.Color {
			gb.spPalette.Palette[gb.spPalette.Idx] = v
			gb.spPalette.update(gb.spPalette.Idx)

			if gb.spPalette.Inc {
				gb.spPalette.Idx = (gb.spPalette.Idx + 1) & 0b111111
			}
		}
        io[addr-0xFF00] = v

	case 0xFF70:
		if gb.Color {
			gb.MemoryBus.WRAMBank = v & 0x7

			if gb.MemoryBus.WRAMBank == 0 {
				gb.MemoryBus.WRAMBank = 1
			}
		}
        io[addr-0xFF00] = v
    default:
        io[addr-0xFF00] = v
	}
}
