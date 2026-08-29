package mem

import (
	"encoding/binary"
	"unsafe"
)

type Bus7 struct {
	M *Mem
}

func (b *Bus7) Read8(addr uint32) uint32 {
	switch addr >> 24 {
	case 0x0:

		if addr < 0x4000 && (*b.M.arm7Pc) < 0x4000 {
			return uint32((*b.M.Arm7Bios)[addr])
		}

		return uint32(0xFF)

	case 0x1:
		return uint32(0xFF)

	case 0x2:
		return uint32(b.M.MainRam[addr&0x3F_FFFF])
	case 0x3:
		return uint32(b.M.WRAM.Read7(addr))
	case 0x4:
		return uint32(b.ReadIO(addr & 0xFF_FFFF))
	case 0x6:
		return uint32(b.M.Ppu.Vram.Read7(addr))
	case 0x8, 0x9, 0xA:
		return uint32(b.M.Cartridge.ReadGbaSlot(addr, false))
	default:
		return uint32(0)
	}
}

func (b *Bus7) Read16(addr uint32) uint32 {
	addr &^= 1

	if io := addr>>24 == 4; io {

		switch addr {
		case 0x400_0100, 0x400_0102:
			return uint32(b.M.Timers7[0].Read16(int(addr & 3)))
		case 0x400_0104, 0x400_0106:
			return uint32(b.M.Timers7[1].Read16(int(addr & 3)))
		case 0x400_0108, 0x400_010A:
			return uint32(b.M.Timers7[2].Read16(int(addr & 3)))
		case 0x400_010c, 0x400_010E:
			return uint32(b.M.Timers7[3].Read16(int(addr & 3)))
		}
		if addr >= 0x480_0000 && addr < 0x490_0000 {
			return uint32(b.M.Wifi.Read16(addr))
		}
	}

	if ptr := b.ReadPtr(addr); ptr != nil {
		return uint32(binary.LittleEndian.Uint16((*[4]uint8)(ptr)[:]))
	}

	return b.Read8(addr) | (b.Read8(addr+1) << 8)
}

func (b *Bus7) Read32(addr uint32) uint32 {
	addr &^= 3
	switch addr {
	case 0x410_0000:
		return b.M.Ipc.ReadFifo(false)
	case 0x410_0010:
		return b.M.Cartridge.ReadCmdIn(false)
	default:

		if ptr := b.ReadPtr(addr); ptr != nil {
			return binary.LittleEndian.Uint32((*[4]uint8)(ptr)[:])
		}

		return b.Read16(addr) | (b.Read16(addr+2) << 16)
	}
}

func (b *Bus7) ReadPtr(addr uint32) unsafe.Pointer {
	switch addr >> 24 {
	case 0x0:

		if addr < 0x4000 { // do not limit based on pc, messes up jit arm7
			return unsafe.Add(unsafe.Pointer(&(*b.M.Arm7Bios)[0]), addr)
		}

		return nil

	case 0x2:
		return unsafe.Add(unsafe.Pointer(&b.M.MainRam), addr&0x3F_FFFF)
	case 0x3:
		return b.M.WRAM.ReadPtr7(addr)
	case 0x6:
		return b.M.Ppu.Vram.ReadPtr7(addr)
	default:
		return nil
	}
}

func (b *Bus7) Write8(addr uint32, v uint8) {
	// there may be the ability to invalidate only arm7 or arm9  -
	// for example wram is special

	switch addr >> 24 {
	case 0x2:
		//clearTempUnimplimented(addr)
		b.M.MainRam[addr&0x3F_FFFF] = v
	case 0x3:
		b.M.WRAM.Write7(addr, v)
	case 0x4:
		b.WriteIO(addr&0xFF_FFFF, v)
	case 0x6:
		b.M.Ppu.Vram.Write7(addr, v)
	}
}

func (b *Bus7) Write16(addr uint32, v uint16) {
	addr &^= 1
	if io := addr>>24 == 4; io {
		switch addr {
		case 0x400_0100, 0x400_0102:
			b.M.Timers7[0].Write16(addr&3, v)
			return
		case 0x400_0104, 0x400_0106:
			b.M.Timers7[1].Write16(addr&3, v)
			return
		case 0x400_0108, 0x400_010A:
			b.M.Timers7[2].Write16(addr&3, v)
			return
		case 0x400_010C, 0x400_010E:
			b.M.Timers7[3].Write16(addr&3, v)
			return
		}

		if addr >= 0x480_0000 && addr < 0x490_0000 {
			b.M.Wifi.Write16(addr, v)
			return
		}
	}

	if ptr := b.WritePtr(addr); ptr != nil {
		binary.LittleEndian.PutUint16((*[4]uint8)(ptr)[:], v)
		return
	}

	b.Write8(addr, uint8(v))
	b.Write8(addr+1, uint8(v>>8))
}

func (b *Bus7) Write32(addr, v uint32) {
	addr &^= 3
	if io := addr>>24 == 4; io {
		switch addr {
		case 0x400_0100:
			b.M.Timers7[0].Write32(v)
			return
		case 0x400_0104:
			b.M.Timers7[1].Write32(v)
			return
		case 0x400_0108:
			b.M.Timers7[2].Write32(v)
			return
		case 0x400_010c:
			b.M.Timers7[3].Write32(v)
			return
		case 0x400_0188:
			b.M.Ipc.WriteFifo(v, false)
			return
		}
	}

	if ptr := b.WritePtr(addr); ptr != nil {
		binary.LittleEndian.PutUint32((*[4]uint8)(ptr)[:], v)
		return
	}
	b.Write16(addr+0, uint16(v))
	b.Write16(addr+2, uint16(v>>16))
}

func (b *Bus7) WritePtr(addr uint32) unsafe.Pointer {
	switch addr >> 24 {
	case 0x2:
		return unsafe.Add(unsafe.Pointer(&b.M.MainRam), addr&0x3F_FFFF)
	case 0x3:
		return b.M.WRAM.ReadPtr7(addr)
	case 0x6:
		return b.M.Ppu.Vram.ReadPtr7(addr)
	default:
		return nil
	}
}
func (b *Bus7) WriteGXFIFO(v uint32) {}

type Bus9 struct {
	M *Mem
}

func (b *Bus9) Read8(addr uint32) uint32 {
	if v, ok := b.M.Tcm.Read(addr); ok {
		return uint32(v)
	}

	switch addr >> 24 {
	case 0x2:
		return uint32(b.M.MainRam[addr&0x3F_FFFF])
	case 0x3:
		return uint32(b.M.WRAM.Read9(addr))
	case 0x4:
		return uint32(b.ReadIO(addr & 0xFF_FFFF))
	case 0x5:
		return uint32(b.M.Ppu.ReadPram(addr))
	case 0x6:
		return uint32(b.M.Ppu.Vram.Read9(addr))
	case 0x7:
		return uint32(b.M.Oam[addr&0x7FF])
	case 0x8, 0x9, 0xA:
		return uint32(b.M.Cartridge.ReadGbaSlot(addr, true))
	case 0xFF:
		return uint32((*b.M.Arm9Bios)[addr&0x0FFF])
	default:
		return uint32(0)
	}
}

func (b *Bus9) Read16(addr uint32) uint32 {
	addr &^= 1
	if io := addr>>24 == 4; io {
		switch addr {
		case 0x400_0100, 0x400_0102:
			return uint32(b.M.Timers9[0].Read16(int(addr & 3)))
		case 0x400_0104, 0x400_0106:
			return uint32(b.M.Timers9[1].Read16(int(addr & 3)))
		case 0x400_0108, 0x400_010A:
			return uint32(b.M.Timers9[2].Read16(int(addr & 3)))
		case 0x400_010c, 0x400_010E:
			return uint32(b.M.Timers9[3].Read16(int(addr & 3)))
		}
	}

	if ptr := b.ReadPtr(addr); ptr != nil {
		return uint32(binary.LittleEndian.Uint16((*[4]uint8)(ptr)[:]))
	}

	return b.Read8(addr) | (b.Read8(addr+1) << 8)
}

func (b *Bus9) Read32(addr uint32) uint32 {
	addr &^= 3
	switch addr {
	case 0x410_0000:
		return b.M.Ipc.ReadFifo(true)
	case 0x410_0010:
		return b.M.Cartridge.ReadCmdIn(true)
	default:

		if ptr := b.ReadPtr(addr); ptr != nil {
			return binary.LittleEndian.Uint32((*[4]uint8)(ptr)[:])
		}

		return b.Read16(addr) | (b.Read16(addr+2) << 16)
	}
}

func (b *Bus9) ReadPtr(addr uint32) unsafe.Pointer {
	if ptr := b.M.Tcm.ReadPtr(addr); ptr != nil {
		return ptr
	}

	switch addr >> 24 {
	case 0x2:
		return unsafe.Add(unsafe.Pointer(&b.M.MainRam), addr&0x3F_FFFF)
	case 0x3:
		return b.M.WRAM.ReadPtr9(addr)
	case 0x5:
		return nil
	case 0x6:
		return b.M.Ppu.Vram.ReadPtr9(addr)
	case 0x7:
		return unsafe.Add(unsafe.Pointer(&b.M.Oam), addr&0x7FF)
	case 0xFF:
		return unsafe.Add(unsafe.Pointer(&(*b.M.Arm9Bios)[0]), addr&0x0FFF)
	}

	return nil
}

func (b *Bus9) Write8(addr uint32, v uint8) {
	if ok := b.M.Tcm.Write(addr, v); ok {
		return
	}

	switch addr >> 24 {
	case 0x2:
		//clearTempUnimplimented(addr)
		b.M.MainRam[addr&0x3F_FFFF] = v
	case 0x3:
		b.M.WRAM.Write9(addr, v)
	case 0x4:
		b.WriteIO(addr&0xFF_FFFF, v)
	case 0x5:
		b.M.Ppu.WritePram(addr, v)
	case 0x6:
		b.M.Ppu.Vram.Write9(addr, v)
	case 0x7:
		b.M.Oam[addr&0x7FF] = v
		b.M.Ppu.UpdateOAM(addr, v, &b.M.Oam)
	}
}

func (b *Bus9) Write16(addr uint32, v uint16) {
	addr &^= 1

	if io := addr>>24 == 4; io {
		switch addr {
		case 0x400_0100, 0x400_0102:
			b.M.Timers9[0].Write16(addr&3, v)
			return
		case 0x400_0104, 0x400_0106:
			b.M.Timers9[1].Write16(addr&3, v)
			return
		case 0x400_0108, 0x400_010A:
			b.M.Timers9[2].Write16(addr&3, v)
			return
		case 0x400_010C, 0x400_010E:
			b.M.Timers9[3].Write16(addr&3, v)
			return
		}
	}

	if ptr := b.WritePtr(addr); ptr != nil {
		binary.LittleEndian.PutUint16((*[4]uint8)(ptr)[:], v)
		return
	}

	b.Write8(addr, uint8(v))
	b.Write8(addr+1, uint8(v>>8))
}

func (b *Bus9) Write32(addr, v uint32) {
	addr &^= 3

	if io := addr>>24 == 4; io {

		if geo := addr >= 0x4000440 && addr < 0x4000600; geo {
			b.M.Ppu.Rasterizer.GeoCmd(addr, v)
			return
		}

		if gxfifo := addr >= 0x400_0400 && addr < 0x4000440; gxfifo {
			b.M.Ppu.Rasterizer.GeoEngine.Fifo(v)
			return
		}

		switch addr {
		case 0x400_0100:
			b.M.Timers9[0].Write32(v)
			return
		case 0x400_0104:
			b.M.Timers9[1].Write32(v)
			return
		case 0x400_0108:
			b.M.Timers9[2].Write32(v)
			return
		case 0x400_010c:
			b.M.Timers9[3].Write32(v)
			return
		case 0x400_0188:
			b.M.Ipc.WriteFifo(v, true)
			return
		}
	}

	if ptr := b.WritePtr(addr); ptr != nil {
		binary.LittleEndian.PutUint32((*[4]uint8)(ptr)[:], v)
		return
	}

	b.Write16(addr+0, uint16(v))
	b.Write16(addr+2, uint16(v>>16))
}

func (b *Bus9) WritePtr(addr uint32) unsafe.Pointer {
	if ptr := b.M.Tcm.WritePtr(addr); ptr != nil {
		return ptr
	}

	switch addr >> 24 {
	case 0x2:
		return unsafe.Add(unsafe.Pointer(&b.M.MainRam), addr&0x3F_FFFF)
	case 0x3:
		return b.M.WRAM.ReadPtr9(addr)
	case 0x6:
		return b.M.Ppu.Vram.ReadPtr9(addr)
	}

	return nil
}
func (b *Bus9) WriteGXFIFO(v uint32) { b.M.Ppu.Rasterizer.GeoEngine.Fifo(v) }
