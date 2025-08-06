package gameboy

import (
	"errors"
	"fmt"
	"time"

	"github.com/aabalke/guac/emu/gb/cartridge"
)

type MemoryBus struct {
	Memory [0x10000]uint8
	WRAM   [0x9000]uint8
	VRAM   [0x4000]uint8
	OAM    [0x100]uint8

	ramSaved bool

	VRAMBank uint8
	WRAMBank uint8

	HdmaLength uint8
	HdmaActive bool
}

func initMemory(gb *GameBoy) {

	gb.WriteByte(0xFF04, 0x1E) // not sur eon this one
	gb.WriteByte(0xFF05, 0x00)
	gb.WriteByte(0xFF06, 0x00)
	gb.WriteByte(0xFF07, 0x00)
	gb.WriteByte(0xFF0F, 0xE1)
	gb.WriteByte(0xFF10, 0x80)
	gb.WriteByte(0xFF11, 0xBF)
	gb.WriteByte(0xFF12, 0xF3)
	gb.WriteByte(0xFF14, 0xBF)
	gb.WriteByte(0xFF16, 0x3F)
	gb.WriteByte(0xFF17, 0x00)
	gb.WriteByte(0xFF19, 0xBF)
	gb.WriteByte(0xFF1A, 0x7F)
	gb.WriteByte(0xFF1B, 0xFF)
	gb.WriteByte(0xFF1C, 0x9F)
	gb.WriteByte(0xFF1E, 0xBF)
	gb.WriteByte(0xFF20, 0xFF)
	gb.WriteByte(0xFF21, 0x00)
	gb.WriteByte(0xFF22, 0x00)
	gb.WriteByte(0xFF23, 0xBF)
	gb.WriteByte(0xFF24, 0x77)
	gb.WriteByte(0xFF25, 0xF3)

	gb.WriteByte(0xFF26, 0xF1)

	gb.WriteByte(0xFF40, 0x91)
	gb.WriteByte(0xFF42, 0x00)
	gb.WriteByte(0xFF43, 0x00)
	gb.WriteByte(0xFF45, 0x00)
	gb.WriteByte(0xFF47, 0xFC)
	gb.WriteByte(0xFF48, 0xFF)
	gb.WriteByte(0xFF49, 0xFF)
	gb.WriteByte(0xFF4A, 0x00)
	gb.WriteByte(0xFF4B, 0x00)
	gb.WriteByte(0xFFFF, 0x00)

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

func (gb *GameBoy) ReadByte(addr uint16) (uint8, error) {

	if addr == 0xFF26 {

		return 0xFA, nil

	}

	Memory := &gb.MemoryBus.Memory

	if int(addr) > len(Memory) {
		return 0, errors.New(fmt.Sprintf("Reading Memory Address %X > %X Capacity", addr, len(Memory)))
	}

	switch {
	case addr == 0xFF00:
		return gb.getJoypad(), nil
	case addr == 0xFF01:
		return Memory[addr], nil
	case addr == 0xFF68:

		if gb.Color {
			return gb.bgPalette.Idx, nil
		}

		return 0, nil

	case addr == 0xFF69:

		if gb.Color {
			return gb.bgPalette.Palette[gb.bgPalette.Idx], nil
		}

		return 0, nil
	case addr == 0xFF6A:

		if gb.Color {
			return gb.spPalette.Idx, nil
		}

		return 0, nil

	case addr == 0xFF6B:

		if gb.Color {
			return gb.spPalette.Palette[gb.spPalette.Idx], nil
		}

		return 0, nil

	case addr == 0xFF4D:

		var b uint8 = 0

		if gb.DoubleSpeed {
			b = 1 << 7
		}

		if gb.PrepareSpeedToggle {
			b |= 1
		}

		return b, nil

	case addr == 0xFF4F:
		return gb.MemoryBus.VRAMBank, nil
	case addr == 0xFF70:
		return gb.MemoryBus.WRAMBank, nil
	}

	switch {
	case addr < 0x4000:
		return gb.Cartridge.Mbc.Read(gb.Cartridge, addr), nil
	case addr < 0x8000:
		return gb.Cartridge.Mbc.ReadRom(gb.Cartridge, addr), nil
	case addr < 0xA000:
		var offset uint16
		if gb.Color {
			offset = uint16(gb.MemoryBus.VRAMBank) * 0x2000
		}
		return gb.MemoryBus.VRAM[addr-0x8000+offset], nil
	case addr < 0xC000:
		return gb.Cartridge.Mbc.ReadRam(gb.Cartridge, addr), nil
	case addr < 0xD000:
		return gb.MemoryBus.WRAM[addr-0xC000], nil
	case addr < 0xE000:
		return gb.MemoryBus.WRAM[(addr-0xC000)+(uint16(gb.MemoryBus.WRAMBank)*0x1000)], nil
	case addr < 0xFE00:
		return 0xFF, nil
	case addr < 0xFEA0:
		return gb.MemoryBus.OAM[addr-0xFE00], nil
	case addr < 0xFF00:
		return 0xFF, nil
	case addr < 0xFF10:
		return Memory[addr], nil
	case addr < 0xFF40:
		return Memory[addr], nil
	default:
		return Memory[addr], nil
	}
}

func (gb *GameBoy) WriteByte(addr uint16, byte uint8) error {

	//if addr == 0xFF26 {
	//    fmt.Printf("Wrote %X to %X\n", byte, addr)
	//}

	Memory := &gb.MemoryBus.Memory

	if int(addr) > len(Memory) {
		return errors.New(fmt.Sprintf("Writing to Memory Address %X > %X Capacity", addr, len(Memory)))
	}

	switch addr {
	case 0xFF04: // DIV
		gb.Timer.Counter = 0
		gb.Timer.DivReg = 0
		Memory[0xFF04] = 0
		return nil
	case 0xFF05:
		Memory[0xFF05] = byte
		return nil
	case 0xFF06:
		Memory[0xFF06] = byte
		return nil
	case 0xFF07:
		currFreq := Memory[0xFF07] & 0x03
		Memory[0xFF07] = byte | 0xF8
		newFreq := Memory[0xFF07] & 0x03

		if currFreq != newFreq {
			gb.Timer.Counter = 0
		}
		return nil
	case 0xFF26:

		// Only bit 7 of master volume is writeable
		bit := byte & 0x80
		v := (Memory[0xFF26] & 0x7F) | bit
		WriteSound(uint32(addr&0xFF), v, gb.Apu)
		//gb.Apu.Update(addr, v, gb)

	case 0xFF44:
		Memory[0xFF44] = 0
		return nil
	case 0xFF46: // DMA
		address := uint16(byte) << 8

		var i uint16 = 0
		for i = range 0xA0 {
			a, _ := gb.ReadByte(address + i)
			gb.WriteByte(0xFE00+i, a)
		}

		return nil

	case 0xFF4D:
		if gb.Color {
			gb.PrepareSpeedToggle = gb.flagEnabled(byte, 0)
			Memory[0xFF4D] = Memory[0xFF4D]*0x80 | (byte & 0x1)
		}

		return nil

	case 0xFF4F:
		if gb.Color && !gb.MemoryBus.HdmaActive {
			gb.MemoryBus.VRAMBank = byte & 0x1
		}

		return nil

	case 0xFF55:
		if gb.Color {
			gb.cgbDMATransfer(byte)
		}
		return nil
	case 0xFF68:
		if gb.Color {
			gb.bgPalette.Idx = byte & 0b111111
			gb.bgPalette.Inc = gb.flagEnabled(byte, 7)
		}
		return nil
	case 0xFF69:
		if gb.Color {
			gb.bgPalette.Palette[gb.bgPalette.Idx] = byte

			if gb.bgPalette.Inc {
				gb.bgPalette.Idx = (gb.bgPalette.Idx + 1) & 0b111111
			}
		}
		return nil
	case 0xFF6A:
		if gb.Color {
			gb.spPalette.Idx = byte & 0b111111
			gb.spPalette.Inc = gb.flagEnabled(byte, 7)
		}
		return nil
	case 0xFF6B:
		if gb.Color {
			gb.spPalette.Palette[gb.spPalette.Idx] = byte

			if gb.spPalette.Inc {
				gb.spPalette.Idx = (gb.spPalette.Idx + 1) & 0b111111
			}
		}
		return nil
	case 0xFF70:
		if gb.Color {
			gb.MemoryBus.WRAMBank = byte & 0x7

			if gb.MemoryBus.WRAMBank == 0 {
				gb.MemoryBus.WRAMBank = 1
			}
		}
		return nil
	}

	switch {
	case addr < 0x8000:
		gb.Cartridge.Mbc.Handle(addr, byte)
	case addr < 0xA000:
		var offset uint16
		if gb.Color {
			offset = uint16(gb.MemoryBus.VRAMBank) * 0x2000
		}
		gb.MemoryBus.VRAM[addr-0x8000+offset] = byte
	case addr < 0xC000:
		gb.Cartridge.Mbc.WriteRam(gb.Cartridge, addr, byte)
		gb.MemoryBus.ramSaved = false
	case addr < 0xD000:
		gb.MemoryBus.WRAM[addr-0xC000] = byte
	case addr < 0xE000:
		gb.MemoryBus.WRAM[(addr-0xC000)+(uint16(gb.MemoryBus.WRAMBank)*0x1000)] = byte
	case addr < 0xFE00:
		Memory[addr] = byte
		_ = gb.WriteByte(addr-0x2000, byte)
	case addr < 0xFEA0:
		// OAM
		gb.MemoryBus.OAM[addr-0xFE00] = byte
	case addr < 0xFEFF:
		// Prohibited Memory sometimes debugged in ROM only cartridges
		Memory[addr] = byte
	case addr < 0xFF10:
		// Misc IO
		Memory[addr] = byte
	case addr < 0xFF40:
		// Sound IO
		Memory[addr] = byte
		WriteSound(uint32(addr&0xFF), byte, gb.Apu)
		//gb.Apu.Update(addr, byte, gb)

	default:
		Memory[addr] = byte
	}

	return nil
}

func (gb *GameBoy) cgbDMATransfer(byte uint8) {

	if gb.MemoryBus.HdmaActive && !gb.flagEnabled(byte, 7) {
		gb.MemoryBus.HdmaActive = false
		gb.MemoryBus.Memory[0xFF55] |= 0x80
		return
	}

	length := ((uint16(byte) & 0x7F) + 1) * 0x10

	if !gb.flagEnabled(byte, 7) {

		gb.performDMATransfer(length)
		gb.MemoryBus.Memory[0xFF55] = 0xFF
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
		MemBus.Memory[0xFF55] = MemBus.HdmaLength
		return
	}

	MemBus.Memory[0xFF55] = 0xFF
	MemBus.HdmaActive = false
}

func (gb *GameBoy) performDMATransfer(length uint16) {

	Mem := &gb.MemoryBus.Memory

	src := (uint16(Mem[0xFF51])<<8 | uint16(Mem[0xFF52])) & 0xFFF0
	dst := (uint16(Mem[0xFF53])<<8 | uint16(Mem[0xFF54])) & 0x1FF0
	dst += 0x8000

	for range length {
		b, _ := gb.ReadByte(src)
		_ = gb.WriteByte(dst, b)
		dst++
		src++
	}

	Mem[0xFF51] = uint8(src >> 8)
	Mem[0xFF52] = uint8(src & 0xFF)
	Mem[0xFF53] = uint8(dst >> 8)
	Mem[0xFF54] = uint8(dst & 0xF0)
}
