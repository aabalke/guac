package mem

import (
	"fmt"

	"github.com/aabalke/guac/emu/nds/ppu"
)

func (b *Bus9) ReadIO(addr uint32) uint8 {
	mem := b.M
	switch {
	case addr >= 0x280 && addr < 0x2B0:
		return mem.div.Read(addr)
	case addr >= 0x2B0 && addr < 0x2C0:
		return mem.sqrt.Read(addr)
	case addr >= 0xB0 && addr < 0xE0:
		addr -= 0xB0
		return mem.dma9[addr/12].Read(addr % 12)
	case (addr >= 0x320 && addr < 0x6A3) || (addr&^1 == 0x60):
		return mem.Ppu.Rasterizer.Read(addr)

	case addr >= 0x100 && addr < 0x110:
		addr -= 0x100
		return mem.Timers9[addr/4].Read(int(addr & 3))

	case addr >= 0x188 && addr < 0x190:
		panic("READ IPC FIFO FROM BYTE OR HALF")

	case addr >= 0x130 && addr < 0x134:
		return mem.Key.Read(addr)

	case addr >= 0x208 && addr < 0x218:
		return mem.irq9.Read(addr)
	}

	switch addr {
	case 0x4, 0x5:
		return mem.Dispstat.Read9(addr & 1)
	case 0x6:
		return uint8(mem.Vcount)
	case 0x7:
		return uint8(mem.Vcount >> 8)

	case 0x64, 0x65, 0x66, 0x67:
		return mem.Ppu.Capture.Read(addr)
	case 0x68, 0x69:
		return 0
	case 0x6C:
		return mem.Ppu.EngineA.MasterBright.Read(0)
	case 0x6D:
		return mem.Ppu.EngineA.MasterBright.Read(1)
	case 0x106C:
		return mem.Ppu.EngineB.MasterBright.Read(0)
	case 0x106D:
		return mem.Ppu.EngineB.MasterBright.Read(1)

	case 0x180, 0x81:
		return mem.Ipc.ReadSync(uint8(addr&1), true)
	case 0x184, 0x185, 0x186, 0x187:
		return mem.Ipc.ReadCnt(uint8(addr&3), true)
	case 0x1A0:
		return mem.Cartridge.ReadAuxSpi(0)
	case 0x1A1:
		return mem.Cartridge.ReadAuxSpi(1)
	case 0x1A2:
		return mem.Cartridge.ReadAuxSpiData()
	case 0x1A3:
		return 0
	case 0x1A4, 0x1A5, 0x1A6, 0x1A7:
		return mem.Cartridge.ReadRomCtrl(uint8(addr & 3))
	case 0x100010, 0x100011, 0x100012, 0x100013:
		panic("arm9 read io gamecard from read8 or read16")

	case 0x204:
		return mem.Cartridge.ReadExMem(0)
	case 0x205:
		return mem.Cartridge.ReadExMem(1)
	case 0x240:
		return mem.Ppu.Vram.Cnt[ppu.A].V
	case 0x241:
		return mem.Ppu.Vram.Cnt[ppu.B].V
	case 0x242:
		return mem.Ppu.Vram.Cnt[ppu.C].V
	case 0x243:
		return mem.Ppu.Vram.Cnt[ppu.D].V
	case 0x244:
		return mem.Ppu.Vram.Cnt[ppu.E].V
	case 0x245:
		return mem.Ppu.Vram.Cnt[ppu.F].V
	case 0x246:
		return mem.Ppu.Vram.Cnt[ppu.G].V
	case 0x248:
		return mem.Ppu.Vram.Cnt[ppu.H].V
	case 0x249:
		return mem.Ppu.Vram.Cnt[ppu.I].V

	case 0x247:
		return mem.WRAM.ReadCNT()
	case 0x300:
		return mem.PostFlg.Read(true)

	case 0x304:
		return uint8(mem.PowCnt.V)
	case 0x305:
		return uint8(mem.PowCnt.V >> 8)

	default:
		return mem.IO[addr]
	}
}

func (b *Bus9) WriteIO(addr uint32, v uint8) {
	mem := b.M

	if ppu := addr < 0x70 || (addr >= 0x1000 && addr < 0x1070); ppu {
		mem.Ppu.Update(addr, uint32(v))
	}

	switch {
	case addr >= 0x280 && addr < 0x2B0:
		mem.div.Write(addr, v)
		return
	case addr >= 0x2B0 && addr < 0x2C0:
		mem.sqrt.Write(addr, v)
		return
	case addr >= 0xB0 && addr < 0xE0:
		addr -= 0xB0
		mem.dma9[addr/12].Write(addr%12, v)
		return
	case addr >= 0x100 && addr < 0x110:
		addr -= 0x100
		mem.Timers9[addr/4].Write(int(addr&3), v)
		return
	case addr >= 0x130 && addr < 0x134:
		mem.Key.Write(addr, v)
		return

	case addr >= 0x188 && addr < 0x190:
		panic("WRITE IPC FIFO FROM BYTE OR HALF")

	case (addr >= 0x320 && addr < 0x6A3) || (addr&^1 == 0x60):
		if addr >= 0x440 && addr < 0x600 {
			panic(fmt.Sprintf("WRITE HALF or BYTE TO 3D %08X\n", addr))
		}

		mem.Ppu.Rasterizer.Write(addr, v)
		return
	case addr >= 0x208 && addr < 0x218:
		mem.irq9.Write8(addr, v)
		return
	}

	switch addr {
	case 0x4, 0x5:
		mem.Dispstat.Write9(addr&1, v)
	case 0x6:
		mem.Vcount = (mem.Vcount & 0xFF00) | uint32(v)
	case 0x7:
		mem.Vcount = (mem.Vcount & 0x00FF) | (uint32(v) << 8)

	case 0x64, 0x65, 0x66, 0x67:
		mem.Ppu.Capture.Write(addr, v)
	case 0x68, 0x69, 0x6A, 0x6B:

	case 0x184, 0x185, 0x186, 0x187:
		mem.Ipc.WriteCnt(v, uint8(addr&3), true)

	case 0x180, 0x181:
		mem.Ipc.WriteSync(v, uint8(addr&1), true)

	case 0x1A0:
		mem.Cartridge.WriteAuxSpi(v, 0, true)
	case 0x1A1:
		mem.Cartridge.WriteAuxSpi(v, 1, true)
	case 0x1A2:
		mem.Cartridge.WriteAuxSpiData(v)
	case 0x1A3:
		return
	case 0x1A4, 0x1A5, 0x1A6, 0x1A7:
		mem.Cartridge.WriteRomCtrl(v, uint8(addr&3), true)

	case 0x1A8, 0x1A9, 0x1AA, 0x1AB, 0x1AC, 0x1AD, 0x1AE, 0x1AF:
		mem.Cartridge.WriteCmdOut(v, uint8(addr&7), true)
	case 0x1B0, 0x1B1, 0x1B2, 0x1B3:
		mem.Cartridge.WriteSeed(v, uint8(addr&3), 0, true)
	case 0x1B4, 0x1B5, 0x1B6, 0x1B7:
		mem.Cartridge.WriteSeed(v, uint8(addr&3), 1, true)
	case 0x1B8:
		mem.Cartridge.WriteSeed(v, 4, 0, true)
	case 0x1B9:
		mem.Cartridge.WriteSeed(v, 5, 0, true)
	case 0x1BA:
		mem.Cartridge.WriteSeed(v, 4, 1, true)
	case 0x1BB:
		mem.Cartridge.WriteSeed(v, 5, 1, true)

	case 0x100010, 0x100011, 0x100012, 0x100013:
		mem.Cartridge.WriteCmdIn(v, uint8(addr&3), true)

	case 0x204:
		mem.Cartridge.WriteExMem(v, 0)
	case 0x205:
		mem.Cartridge.WriteExMem(v, 1)

	// vram reads - gbatek says read only, needed to match no$gba
	case 0x240, 0x241, 0x242, 0x243, 0x244, 0x245, 0x246, 0x248, 0x249:
		mem.Ppu.Vram.WriteCnt(addr, v)
	case 0x247:
		mem.WRAM.WriteCNT(v)

	case 0x300:
		mem.PostFlg.Write(v, true)

	case 0x304:
		mem.PowCnt.WriteCnt1(0, v)
	case 0x305:
		mem.PowCnt.WriteCnt1(1, v)

	default:
		mem.IO[addr] = v
	}
}

func (b *Bus7) ReadIO(addr uint32) uint8 {
	mem := b.M

	switch {
	case addr >= 0xB0 && addr < 0xE0:
		addr -= 0xB0
		return mem.dma7[addr/12].Read(addr % 12)

	case addr >= 0x400 && addr < 0x600:
		return mem.Snd.Read(addr)
	case addr >= 0x100 && addr < 0x110:
		addr -= 0x100
		return mem.Timers7[addr/4].Read(int(addr & 3))
	case addr >= 0x188 && addr < 0x190:
		panic("READ IPC FIFO FROM BYTE OR HALF")
	case addr >= 0x130 && addr < 0x134:
		return mem.Key.Read(addr)
	case addr >= 0x208 && addr < 0x218:
		return mem.irq7.Read(addr)
	}

	switch addr {
	case 0x4, 0x5:
		return mem.Dispstat.Read7(addr & 1)
	case 0x6:
		return uint8(mem.Vcount)
	case 0x7:
		return uint8(mem.Vcount >> 8)

	case 0x134:
		return 0x0F
	case 0x135:
		return 0x80

	case 0x136:
		return mem.Key.Input2

	case 0x138:
		return mem.Rtc.Read()
	case 0x139, 0x13A, 0x13B:
		return 0

	case 0x180, 0x81:
		return mem.Ipc.ReadSync(uint8(addr&1), false)
	case 0x184, 0x185, 0x186, 0x187:
		return mem.Ipc.ReadCnt(uint8(addr&3), false)
	case 0x1A0:
		return mem.Cartridge.ReadAuxSpi(0)
	case 0x1A1:
		return mem.Cartridge.ReadAuxSpi(1)
	case 0x1A2:
		return mem.Cartridge.ReadAuxSpiData()
	case 0x1A3:
		return 0
	case 0x1A4, 0x1A5, 0x1A6, 0x1A7:
		return mem.Cartridge.ReadRomCtrl(uint8(addr & 3))
	case 0x100010, 0x100011, 0x100012, 0x100013:
		panic("arm7 read io gamecard from read8 or read16")

	case 0x1C0, 0x1C1, 0x1C2, 0x1C3:
		return mem.Spi.Read(addr & 3)

	case 0x204, 0x205:
		return mem.Cartridge.ReadExMem(uint8(addr & 1))

	case 0x206:
		return uint8(mem.WifiWaitCnt)
	case 0x207:
		return 0

	case 0x240:
		return mem.Ppu.Vram.Cnt_7
	case 0x241:
		return mem.WRAM.ReadCNT()

	case 0x300:
		return mem.PostFlg.Read(false)

	case 0x301:

		if *mem.halted7 {
			return 0b1000_0000
		}

		return 0

	case 0x304:
		return mem.PowCnt.V2

	case 0x308:
		return uint8(mem.BiosProt)
	case 0x309:
		return uint8(mem.BiosProt >> 8)

	default:
		return mem.IO[addr]
	}
}

func (b *Bus7) WriteIO(addr uint32, v uint8) {
	mem := b.M

	switch {
	case addr < 0x4:
		mem.Ppu.Update(addr, uint32(v))

	case addr >= 0xB0 && addr < 0xE0:
		addr -= 0xB0
		mem.dma7[addr/12].Write(addr%12, v)
		return

	case addr >= 0x400 && addr < 0x600:
		mem.Snd.Write(addr, v)
		return

	case addr >= 0x100 && addr < 0x110:
		addr -= 0x100
		mem.Timers7[addr/4].Write(int(addr&3), v)
		return
	case addr >= 0x130 && addr < 0x134:
		mem.Key.Write(addr, v)
		return
	case addr >= 0x188 && addr < 0x190:
		panic("WRITE IPC FIFO FROM BYTE OR HALF")

	case addr >= 0x208 && addr < 0x218:
		mem.irq7.Write8(addr, v)
		return
	}

	switch addr {
	case 0x4, 0x5:
		mem.Dispstat.Write7(addr&1, v)
	case 0x6:
		mem.Vcount = (mem.Vcount & 0xFF00) | uint32(v)
	case 0x7:
		mem.Vcount = (mem.Vcount & 0x00FF) | (uint32(v) << 8)

	case 0x138:
		mem.Rtc.Write(v)
	case 0x139, 0x13A, 0x13B:
		return

	case 0x180, 0x181:
		mem.Ipc.WriteSync(v, uint8(addr&1), false)
	case 0x184, 0x185, 0x186, 0x187:
		mem.Ipc.WriteCnt(v, uint8(addr&3), false)
	case 0x1A0, 0x1A1:
		mem.Cartridge.WriteAuxSpi(v, uint8(addr&1), false)
	case 0x1A2:
		mem.Cartridge.WriteAuxSpiData(v)
	case 0x1A3:
		return
	case 0x1A4, 0x1A5, 0x1A6, 0x1A7:
		mem.Cartridge.WriteRomCtrl(v, uint8(addr&3), false)
	case 0x1A8, 0x1A9, 0x1AA, 0x1AB, 0x1AC, 0x1AD, 0x1AE, 0x1AF:
		mem.Cartridge.WriteCmdOut(v, uint8(addr&7), false)
	case 0x1B0, 0x1B1, 0x1B2, 0x1B3:
		mem.Cartridge.WriteSeed(v, uint8(addr&3), 0, false)
	case 0x1B4, 0x1B5, 0x1B6, 0x1B7:
		mem.Cartridge.WriteSeed(v, uint8(addr&3), 1, false)
	case 0x1B8:
		mem.Cartridge.WriteSeed(v, 4, 0, false)
	case 0x1B9:
		mem.Cartridge.WriteSeed(v, 5, 0, false)
	case 0x1BA:
		mem.Cartridge.WriteSeed(v, 4, 1, false)
	case 0x1BB:
		mem.Cartridge.WriteSeed(v, 5, 1, false)

	case 0x100010, 0x100011, 0x100012, 0x100013:
		mem.Cartridge.WriteCmdIn(v, uint8(addr&3), false)

	case 0x1C0, 0x1C1, 0x1C2, 0x1C3:
		mem.Spi.Write(addr&3, v)

	case 0x204:
		mem.Cartridge.WriteExMem(v, 0)

	case 0x300:
		mem.PostFlg.Write(v, false)

	case 0x301:

		switch v >> 6 {
		case 0:
			(*mem.halted7) = false
		case 2:
			(*mem.halted7) = true
		default:
			panic(fmt.Sprintf("UNKNOWN HALTCNT VALUE ARM7 %d", v))
		}

	case 0x304:
		mem.PowCnt.WriteCnt2(v)

	case 0x308:
		return

	default:
		mem.IO[addr] = v
	}
}
