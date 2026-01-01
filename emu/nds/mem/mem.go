package mem

import (
	_ "embed"
	"encoding/binary"
	"fmt"
	"time"
	"unsafe"

	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/emu/nds/cart"
	"github.com/aabalke/guac/emu/nds/cpu"
	"github.com/aabalke/guac/emu/nds/mem/dma"
	"github.com/aabalke/guac/emu/nds/mem/spi"
	"github.com/aabalke/guac/emu/nds/ppu"
	"github.com/aabalke/guac/emu/nds/snd"
)

//go:embed res_/bios7.bin
var arm7Bios []byte

//go:embed res_/bios9.bin
var arm9Bios []byte

type Mem struct {
	Tcm     Tcm
	MainRam [0x40_0000]uint8
	WRAM    WRAM
	Oam     [0x800]uint8

	Arm7Bios [0x4000]uint8
	Arm9Bios [0x1000]uint8

	// this size is temp
	IO [0x100_0000]uint8

	halted7, halted9 *bool
	irq7, irq9       *cpu.Irq
	dma7, dma9       *[4]dma.DMA

	arm7Pc *uint32

	Ppu       *ppu.PPU
	Cartridge *cart.Cartridge

	Vcount      uint32
	Dispstat    Dispstat
	Keypad      Keypad
	Gamecard    Gamecard
	div         Div
	sqrt        Sqrt
	Ipc         IPC
	Spi         spi.Spi
	Rtc         Rtc
	PostFlg     PostFlg
	PowCnt      PowCnt
	BiosProt    BiosProt
	WifiWaitCnt WifiWaitCnt
	Timers      [8]Timer

	Snd *snd.Snd

	Save bool

	Jit7, Jit9 Jit
}

type BiosProt uint16
type WifiWaitCnt uint8

func NewMemory(
	arm7Pc *uint32,
	halted7, halted9 *bool,
	dma7, dma9 *[4]dma.DMA,
	irq7, irq9 *cpu.Irq,
	jit7, jit9 Jit,
	c *cart.Cartridge,
	Ppu *ppu.PPU,
	snd *snd.Snd,
	savepath string) Mem {

	m := Mem{
		halted7:   halted7,
		halted9:   halted9,
		dma7:      dma7,
		dma9:      dma9,
		irq9:      irq9,
		irq7:      irq7,
		Cartridge: c,
		Ppu:       Ppu,
		arm7Pc:    arm7Pc,
		Snd:       snd,
		Jit7:      jit7,
		Jit9:      jit9,
	}

	// i believe this is default
	m.WRAM.WriteCNT(3)

	m.WriteArm9IO(0x304, 0x0F)
	m.WriteArm9IO(0x305, 0x82)

	m.BiosProt = 0x1204
	m.WifiWaitCnt = 0x30

	m.Keypad.KEYINPUT = 0x3FF
	m.Keypad.KEYINPUT2 = 0b100_0011

	m.Ipc.Init(irq7, irq9)

	m.LoadBios()

	m.Rtc.InitRtc()

	m.PowCnt.WriteCNT1(0, 0x0F, Ppu)
	m.PowCnt.WriteCNT1(1, 0x82, Ppu)

	m.Spi.Init()
	m.Gamecard.Init(irq7, irq9, dma7, dma9, c, savepath, &m.Save)

	texCache := &Ppu.Rasterizer.GeoEngine.TextureCache
	m.Ppu.Vram.Init(texCache)

	m.InitSaveLoop()

	return m
}

func (m *Mem) InitSaveLoop() {

	if config.Conf.Nds.DisableSaves {
		return
	}

	//println("SAVE DISABLED")

	//return

	saveTicker := time.Tick(time.Second)

	go func() {
		for range saveTicker {
			if m.Save {
				m.Gamecard.Backup.Save()
				m.Save = false
			}
		}
	}()
}

var lockWrites bool

func (mem *Mem) DirtyTransfer() {
	setBiosRam(mem)
	lockWrites = true
}

func (mem *Mem) LoadBios() {

	for i := range len(arm7Bios) {
		mem.Arm7Bios[i] = arm7Bios[i]
	}

	for i := range len(arm9Bios) {
		mem.Arm9Bios[i] = arm9Bios[i]
	}
}

func (mem *Mem) Read(addr uint32, arm9 bool) uint8 {

	if arm9 {

		if addr > 0xFF00_0000 {
			return mem.Arm9Bios[addr&0x0FFF]
		}

		if v, ok := mem.Tcm.ReadTcmWindow(addr); ok {
			return v
		}

		switch addr >> 24 {
		case 0x0, 0x1:
			v, _ := mem.Tcm.Read(addr)
			return v
		case 0x2:
			//ramUsageUnimplimented(addr)
			return mem.MainRam[addr&0x3F_FFFF]
		case 0x3:
			return mem.WRAM.Read(addr, true)
		case 0x4:
			//fmt.Printf("IO READ ARM9 %08X arm9 %t\n", addr, arm9)
			return mem.ReadArm9IO(addr - 0x400_0000)
		case 0x5:
			return mem.Ppu.Pram.Read(addr, mem.Ppu)
		case 0x6:
			return mem.Ppu.Vram.Read(addr, true)
		case 0x7:
			return mem.Oam[addr&0x7FF]
		case 0x8, 0x9, 0xA:
			return mem.ReadGbaSlot(addr, arm9)

		default:
			return 0
			//uhh.PrintPcs()
			panic(fmt.Sprintf("ARM9 ADDR %08X INVALID READ\n", addr))
		}
	}
	switch addr >> 24 {
	case 0x0, 0x1:

		if addr < 0x4000 && (*mem.arm7Pc) < 0x4000 {
			return mem.Arm7Bios[addr]
		}

		return 0xFF

	case 0x2:
		//ramUsageUnimplimented(addr)
		return mem.MainRam[addr&0x3F_FFFF]
	case 0x3:
		return mem.WRAM.Read(addr, false)
	case 0x4:
		//fmt.Printf("IO READ ARM9 %08X arm9 %t\n", addr, arm9)
		//printIO(addr, arm9, false)
		//fmt.Printf("RD IO %08X\n", addr)
		return mem.ReadArm7IO(addr - 0x400_0000)
	case 0x6:
		return mem.Ppu.Vram.Read(addr, false)
	case 0x8, 0x9, 0xA:
		return mem.ReadGbaSlot(addr, arm9)
	default:
		return 0
		//uhh.PrintPcs()
		panic(fmt.Sprintf("ARM7 ADDR %08X INVALID READ\n", addr))
		return 0
	}
}

func (mem *Mem) Read8(addr uint32, arm9 bool) uint32 {
	return uint32(mem.Read(addr, arm9))
}
func (mem *Mem) Read16(addr uint32, arm9 bool) uint32 {
	if ptr, ok := mem.ReadPtr(addr, arm9); ok {
		return uint32(binary.LittleEndian.Uint16((*[4]uint8)(ptr)[:]))
	}
	return uint32(mem.Read(addr, arm9)) | (uint32(mem.Read(addr+1, arm9)) << 8)
}
func (mem *Mem) Read32(addr uint32, arm9 bool) uint32 {

	switch addr {
	case 0x410_0000:
		return mem.Ipc.ReadFifo(arm9)
	case 0x410_0010:
		return mem.Gamecard.RomCtrl.ReadCmdIn(arm9)
	default:

		if ptr, ok := mem.ReadPtr(addr, arm9); ok {
			return binary.LittleEndian.Uint32((*[4]uint8)(ptr)[:])
		}

		a := uint32(mem.Read(addr+2, arm9)) | (uint32(mem.Read(addr+3, arm9)) << 8)
		b := uint32(mem.Read(addr, arm9)) | (uint32(mem.Read(addr+1, arm9)) << 8)
		return (a << 16) | b
	}
}

func (mem *Mem) WritePtr(addr uint32, arm9 bool) (unsafe.Pointer, bool) {
	//mem.Jit7.InvalidatePage(addr)
	mem.Jit9.InvalidatePage(addr)

	if arm9 {

		if v, ok := mem.Tcm.ReadTcmWindowPtr(addr); ok {
			return v, ok
		}

		switch addr >> 24 {
		case 0x0, 0x1:
			return mem.Tcm.ReadPtr(addr)
		case 0x2:
			return unsafe.Add(unsafe.Pointer(&mem.MainRam), addr&0x3F_FFFF), true
		case 0x3:
			return mem.WRAM.ReadPtr(addr, true)
		}

		return nil, false
	}

	switch addr >> 24 {
	case 0x0:

		if addr < 0x4000 && (*mem.arm7Pc) < 0x4000 {
			return unsafe.Add(unsafe.Pointer(&mem.Arm7Bios), addr), true
		}

		return nil, false

	case 0x2:
		return unsafe.Add(unsafe.Pointer(&mem.MainRam), addr&0x3F_FFFF), true
	case 0x3:
		return mem.WRAM.ReadPtr(addr, false)
	case 0x6:
		return mem.Ppu.Vram.ReadPtr(addr, false)
	default:
		return nil, false
	}
}

func (mem *Mem) ReadPtr(addr uint32, arm9 bool) (unsafe.Pointer, bool) {

	if arm9 {

		if addr > 0xFF00_0000 {
			return unsafe.Add(unsafe.Pointer(&mem.Arm9Bios), addr&0x0FFF), true
		}

		//if v, ok := mem.Tcm.ReadTcmWindow(addr); ok {
		if v, ok := mem.Tcm.ReadTcmWindowPtr(addr); ok {
			return v, ok
		}

		switch addr >> 24 {
		case 0x0, 0x1:
			return mem.Tcm.ReadPtr(addr)
		case 0x2:
			return unsafe.Add(unsafe.Pointer(&mem.MainRam), addr&0x3F_FFFF), true
		case 0x3:
			return mem.WRAM.ReadPtr(addr, true)
		case 0x5:
			return nil, false
			//return mem.Ppu.Pram.Read(addr, mem.ppu)
		case 0x6:
			return mem.Ppu.Vram.ReadPtr(addr, true)
		case 0x7:
			return unsafe.Add(unsafe.Pointer(&mem.Oam), addr&0x7FF), true
		}

		return nil, false
	}

	switch addr >> 24 {
	case 0x0:

		if addr < 0x4000 && (*mem.arm7Pc) < 0x4000 {
			return unsafe.Add(unsafe.Pointer(&mem.Arm7Bios), addr), true
		}

		return nil, false

	case 0x2:
		return unsafe.Add(unsafe.Pointer(&mem.MainRam), addr&0x3F_FFFF), true
	case 0x3:
		return mem.WRAM.ReadPtr(addr, false)
	case 0x6:
		return mem.Ppu.Vram.ReadPtr(addr, false)
	default:
		return nil, false
	}
}

func (mem *Mem) Write(addr uint32, v uint8, arm9 bool) {

	//mem.Jit7.InvalidatePage(addr)
	mem.Jit9.InvalidatePage(addr)

	if arm9 {

		if ok := mem.Tcm.WriteTcmWindow(addr, v); ok {
			return
		}

		switch addr >> 24 {
		case 0x0, 0x1:
			mem.Tcm.Write(addr, v)
		case 0x2:
			//clearTempUnimplimented(addr)
			mem.MainRam[addr&0x3F_FFFF] = v
		case 0x3:
			mem.WRAM.Write(addr, v, true)
		case 0x4:
			//printIO(addr, arm9, true)
			mem.WriteArm9IO(addr-0x400_0000, v)
		case 0x5:
			mem.Ppu.Pram.Write(addr, v, mem.Ppu)
		case 0x6:
			mem.Ppu.Vram.Write(addr, v, true)
		case 0x7:
			mem.Oam[addr&0x7FF] = v
			mem.Ppu.UpdateOAM(addr, v, &mem.Oam)
		}

		return
	}

	switch addr >> 24 {
	case 0x2:
		//clearTempUnimplimented(addr)
		mem.MainRam[addr&0x3F_FFFF] = v
	case 0x3:
		mem.WRAM.Write(addr, v, false)
	case 0x4:
		//printIO(addr, arm9, true)
		mem.WriteArm7IO(addr-0x400_0000, v)
	case 0x6:
		mem.Ppu.Vram.Write(addr, v, false)
	}
}

func (mem *Mem) Write8(addr uint32, v uint8, arm9 bool) {
	mem.Write(addr, v, arm9)
}
func (mem *Mem) Write16(addr uint32, v uint16, arm9 bool) {

	if arm9 && addr >= 0x0400_0068 && addr < 0x0400_006C {
		mem.Ppu.DisplayFifo.FifoWrite(v)
		return
	}

	if ptr, ok := mem.WritePtr(addr, arm9); ok {
		binary.LittleEndian.PutUint16((*[4]uint8)(ptr)[:], v)
		return
	}

	mem.Write(addr, uint8(v), arm9)
	mem.Write(addr+1, uint8(v>>8), arm9)
}
func (mem *Mem) Write32(addr uint32, v uint32, arm9 bool) {

	if arm9 && addr == 0x400_0400 {
		mem.Ppu.Rasterizer.GeoEngine.Fifo(v)
		return
	}

	if arm9 {

		if geo := addr >= 0x4000440 && addr < 0x4000600; geo {
			mem.Ppu.Rasterizer.GeoCmd(addr, v)
			return
		}

		if gxfifo := addr >= 0x400_0400 && addr < 0x4000440; gxfifo {
			mem.Ppu.Rasterizer.GeoEngine.Fifo(v)
			//mem.Ppu.Rasterizer.GeoCmdFifo(v)
			return
		}

		if addr >= 0x0400_0068 && addr < 0x0400_006C {
			mem.Ppu.DisplayFifo.FifoWrite(uint16(v))
			mem.Ppu.DisplayFifo.FifoWrite(uint16(v >> 16))
			return
		}
	}

	switch addr {
	case 0x400_0188:
		mem.Ipc.WriteFifo(v, arm9)
	default:
		if ptr, ok := mem.WritePtr(addr, arm9); ok {
			binary.LittleEndian.PutUint32((*[4]uint8)(ptr)[:], v)
			return
		}
		mem.Write(addr+0, uint8(v), arm9)
		mem.Write(addr+1, uint8(v>>8), arm9)
		mem.Write(addr+2, uint8(v>>16), arm9)
		mem.Write(addr+3, uint8(v>>24), arm9)
	}
}

func (mem *Mem) WriteGXFIFO(v uint32) {
	mem.Ppu.Rasterizer.GeoEngine.Fifo(v)
}

func (mem *Mem) ReadArm9IO(addr uint32) uint8 {

	//if addr != 0x180 && addr != 0x181 && addr < 0x3000 {
	//	fmt.Printf("READ ADDR %08X\n", addr)
	//}

	if addr >= 0x188 && addr < 0x190 {
		panic("READ IPC FIFO FROM BYTE OR HALF")
	}

	switch {
	case addr >= 0x280 && addr < 0x2B0:
		return mem.div.Read(addr)
	case addr >= 0x2B0 && addr < 0x2C0:
		return mem.sqrt.Read(addr)
	case addr >= 0xB0 && addr < 0xE0:
		return mem.ReadDma(mem.dma9, addr)
	case (addr >= 0x320 && addr < 0x6A3) || (addr&^1 == 0x60):
		return mem.Ppu.Rasterizer.Read(addr)
	}

	switch addr {
	case 0x4:
		return uint8(mem.Dispstat.A9)
	case 0x5:
		return uint8(mem.Dispstat.A9 >> 8)
	case 0x6:
		return uint8(mem.Vcount)
	case 0x7:
		return uint8(mem.Vcount >> 8)

	case 0x64:
		return mem.Ppu.Capture.Read(addr)
	case 0x65:
		return mem.Ppu.Capture.Read(addr)
	case 0x66:
		return mem.Ppu.Capture.Read(addr)
	case 0x67:
		return mem.Ppu.Capture.Read(addr)
	case 0x68:
		return 0
	case 0x69:
		return 0
	case 0x6C:
		return mem.Ppu.EngineA.MasterBright.Read(0)
	case 0x6D:
		return mem.Ppu.EngineA.MasterBright.Read(1)
	case 0x106C:
		return mem.Ppu.EngineB.MasterBright.Read(0)
	case 0x106D:
		return mem.Ppu.EngineB.MasterBright.Read(1)

	case 0x100:
		return mem.Timers[0].ReadD(false)
	case 0x101:
		return mem.Timers[0].ReadD(true)
	case 0x102:
		return mem.Timers[0].ReadCnt(false)
	case 0x103:
		return mem.Timers[0].ReadCnt(true)
	case 0x104:
		return mem.Timers[1].ReadD(false)
	case 0x105:
		return mem.Timers[1].ReadD(true)
	case 0x106:
		return mem.Timers[1].ReadCnt(false)
	case 0x107:
		return mem.Timers[1].ReadCnt(true)
	case 0x108:
		return mem.Timers[2].ReadD(false)
	case 0x109:
		return mem.Timers[2].ReadD(true)
	case 0x10A:
		return mem.Timers[2].ReadCnt(false)
	case 0x10B:
		return mem.Timers[2].ReadCnt(true)
	case 0x10C:
		return mem.Timers[3].ReadD(false)
	case 0x10D:
		return mem.Timers[3].ReadD(true)
	case 0x10E:
		return mem.Timers[3].ReadCnt(false)
	case 0x10F:
		return mem.Timers[3].ReadCnt(true)

	case 0x130:
		return mem.Keypad.readINPUT(false)
	case 0x131:
		return mem.Keypad.readINPUT(true)
	case 0x132:
		return mem.Keypad.readCNT(false)
	case 0x133:
		return mem.Keypad.readCNT(true)

	case 0x180:
		return mem.Ipc.ReadSync(0, true)
	case 0x181:
		return mem.Ipc.ReadSync(1, true)

	case 0x184:
		return mem.Ipc.ReadCnt(0, true)
	case 0x185:
		return mem.Ipc.ReadCnt(1, true)
	case 0x186:
		return mem.Ipc.ReadCnt(2, true)
	case 0x187:
		return mem.Ipc.ReadCnt(3, true)

	case 0x1A0:
		return mem.Gamecard.AuxSpi.Read(0)
	case 0x1A1:
		return mem.Gamecard.AuxSpi.Read(1)
	case 0x1A2:
		return mem.Gamecard.AuxSpi.Read(2)
	case 0x1A3:
		return mem.Gamecard.AuxSpi.Read(3)
	case 0x1A4:
		return mem.Gamecard.RomCtrl.Read(0)
	case 0x1A5:
		return mem.Gamecard.RomCtrl.Read(1)
	case 0x1A6:
		return mem.Gamecard.RomCtrl.Read(2)
	case 0x1A7:
		return mem.Gamecard.RomCtrl.Read(3)

	case 0x100010:
		panic("READING GAMECARD READ IN FROM READ16 OR READ8")
	case 0x100011:
		panic("READING GAMECARD READ IN FROM READ16 OR READ8")
	case 0x100012:
		panic("READING GAMECARD READ IN FROM READ16 OR READ8")
	case 0x100013:
		panic("READING GAMECARD READ IN FROM READ16 OR READ8")

	case 0x204:
		return mem.Gamecard.ExMem.Read(0)
	case 0x205:
		return mem.Gamecard.ExMem.Read(1)
	case 0x208:
		return mem.irq9.ReadIME()
	case 0x209:
		return 0
	case 0x210:
		return mem.irq9.ReadIE(0)
	case 0x211:
		return mem.irq9.ReadIE(1)
	case 0x212:
		return mem.irq9.ReadIE(2)
	case 0x213:
		return mem.irq9.ReadIE(3)
	case 0x214:
		return mem.irq9.ReadIF(0)
	case 0x215:
		return mem.irq9.ReadIF(1)
	case 0x216:
		return mem.irq9.ReadIF(2)
	case 0x217:
		return mem.irq9.ReadIF(3)
	case 0x240:
		return mem.Ppu.Vram.CNT_A.V
	case 0x241:
		return mem.Ppu.Vram.CNT_B.V
	case 0x242:
		return mem.Ppu.Vram.CNT_C.V
	case 0x243:
		return mem.Ppu.Vram.CNT_D.V
	case 0x244:
		return mem.Ppu.Vram.CNT_E.V
	case 0x245:
		return mem.Ppu.Vram.CNT_F.V
	case 0x246:
		return mem.Ppu.Vram.CNT_G.V
	case 0x248:
		return mem.Ppu.Vram.CNT_H.V
	case 0x249:
		return mem.Ppu.Vram.CNT_I.V

	case 0x247:
		return mem.WRAM.ReadCNT()
	case 0x300:
		return mem.PostFlg.Read(true)

	case 0x304:
		return uint8(mem.PowCnt.V)
	case 0x305:
		return uint8(mem.PowCnt.V >> 8)

	default:
		//panic(fmt.Sprintf("READ UNKNOWN ARM9 IO ADDR %08X", addr))
		return mem.IO[addr]
	}
}

func (mem *Mem) WriteArm9IO(addr uint32, v uint8) {

	if addr >= 0x188 && addr < 0x190 {
		panic("WRITE IPC FIFO FROM BYTE OR HALF")
	}

	if ppu := addr < 0x70 || (addr >= 0x1000 && addr < 0x1070); ppu {
		mem.Ppu.Update(addr, uint32(v))
	}

	//if !(addr >= 0x208 && addr < 0x240) {
	//    fmt.Printf("ARM9 WRITE ADDR %08X V %02X\n", addr, v)
	//}

	switch {
	case addr >= 0x280 && addr < 0x2B0:
		mem.div.Write(addr, v)
		return
	case addr >= 0x2B0 && addr < 0x2C0:
		mem.sqrt.Write(addr, v)
		return
	case addr >= 0xB0 && addr < 0xE0:
		mem.WriteDma(mem.dma9, addr, v)
		return
	case (addr >= 0x320 && addr < 0x6A3) || (addr&^1 == 0x60):
		if addr >= 0x440 && addr < 0x600 {
			panic(fmt.Sprintf("WRITE HALF or BYTE TO 3D %08X\n", addr))
		}

		mem.Ppu.Rasterizer.Write(addr, v)
		return
	}

	switch addr {
	case 0x4:
		mem.Dispstat.Write(v, false, true)
	case 0x5:
		mem.Dispstat.Write(v, true, true)
	case 0x6:
		mem.Vcount &^= 0xFF
		mem.Vcount |= uint32(v)
	case 0x7:
		mem.Vcount &= 0xFF
		mem.Vcount |= uint32(v) << 8

	case 0x64:
		mem.Ppu.Capture.Write(addr, v)
	case 0x65:
		mem.Ppu.Capture.Write(addr, v)
	case 0x66:
		mem.Ppu.Capture.Write(addr, v)
	case 0x67:
		mem.Ppu.Capture.Write(addr, v)
	case 0x68:
		panic("ADDR WRITE 0x68 FIFO")
	case 0x69:
		panic("ADDR WRITE 0x69 FIFO")
	case 0x6A:
		panic("ADDR WRITE 0x6A FIFO")
	case 0x6B:
		panic("ADDR WRITE 0x6B FIFO")

	case 0x184:
		mem.Ipc.WriteCnt(v, 0, true)
	case 0x185:
		mem.Ipc.WriteCnt(v, 1, true)
	case 0x186:
		mem.Ipc.WriteCnt(v, 2, true)
	case 0x187:
		mem.Ipc.WriteCnt(v, 3, true)

	case 0x130:
		return
	case 0x131:
		return
	case 0x132:
		mem.Keypad.writeCNT(v, false)
	case 0x133:
		mem.Keypad.writeCNT(v, true)

	case 0x180:
		mem.Ipc.WriteSync(v, 0, true)
	case 0x181:
		mem.Ipc.WriteSync(v, 1, true)

	case 0x1A0:
		mem.Gamecard.AuxSpi.Write(v, 0, true)
	case 0x1A1:
		mem.Gamecard.AuxSpi.Write(v, 1, true)
	case 0x1A2:
		mem.Gamecard.AuxSpi.Write(v, 2, true)
	case 0x1A3:
		mem.Gamecard.AuxSpi.Write(v, 3, true)
	case 0x1A4:
		mem.Gamecard.RomCtrl.Write(v, 0, true)
	case 0x1A5:
		mem.Gamecard.RomCtrl.Write(v, 1, true)
	case 0x1A6:
		mem.Gamecard.RomCtrl.Write(v, 2, true)
	case 0x1A7:
		mem.Gamecard.RomCtrl.Write(v, 3, true)

	case 0x1A8:
		mem.Gamecard.RomCtrl.WriteCmdOut(v, 0, true)
	case 0x1A9:
		mem.Gamecard.RomCtrl.WriteCmdOut(v, 1, true)
	case 0x1AA:
		mem.Gamecard.RomCtrl.WriteCmdOut(v, 2, true)
	case 0x1AB:
		mem.Gamecard.RomCtrl.WriteCmdOut(v, 3, true)
	case 0x1AC:
		mem.Gamecard.RomCtrl.WriteCmdOut(v, 4, true)
	case 0x1AD:
		mem.Gamecard.RomCtrl.WriteCmdOut(v, 5, true)
	case 0x1AE:
		mem.Gamecard.RomCtrl.WriteCmdOut(v, 6, true)
	case 0x1AF:
		mem.Gamecard.RomCtrl.WriteCmdOut(v, 7, true)
	case 0x1B0:
		mem.Gamecard.RomCtrl.WriteSeed(v, 0, 0, true)
	case 0x1B1:
		mem.Gamecard.RomCtrl.WriteSeed(v, 1, 0, true)
	case 0x1B2:
		mem.Gamecard.RomCtrl.WriteSeed(v, 2, 0, true)
	case 0x1B3:
		mem.Gamecard.RomCtrl.WriteSeed(v, 3, 0, true)
	case 0x1B4:
		mem.Gamecard.RomCtrl.WriteSeed(v, 0, 1, true)
	case 0x1B5:
		mem.Gamecard.RomCtrl.WriteSeed(v, 1, 1, true)
	case 0x1B6:
		mem.Gamecard.RomCtrl.WriteSeed(v, 2, 1, true)
	case 0x1B7:
		mem.Gamecard.RomCtrl.WriteSeed(v, 3, 1, true)
	case 0x1B8:
		mem.Gamecard.RomCtrl.WriteSeed(v, 4, 0, true)
	case 0x1B9:
		mem.Gamecard.RomCtrl.WriteSeed(v, 5, 0, true)
	case 0x1BA:
		mem.Gamecard.RomCtrl.WriteSeed(v, 4, 1, true)
	case 0x1BB:
		mem.Gamecard.RomCtrl.WriteSeed(v, 5, 1, true)

	case 0x100010:
		mem.Gamecard.RomCtrl.WriteCmdIn(v, 0, true)
	case 0x100011:
		mem.Gamecard.RomCtrl.WriteCmdIn(v, 1, true)
	case 0x100012:
		mem.Gamecard.RomCtrl.WriteCmdIn(v, 2, true)
	case 0x100013:
		mem.Gamecard.RomCtrl.WriteCmdIn(v, 3, true)

	case 0x100:
		mem.Timers[0].WriteD(v, false)
	case 0x101:
		mem.Timers[0].WriteD(v, true)
	case 0x102:
		mem.Timers[0].WriteCnt(v, false)
	case 0x103:
		mem.Timers[0].WriteCnt(v, true)
	case 0x104:
		mem.Timers[1].WriteD(v, false)
	case 0x105:
		mem.Timers[1].WriteD(v, true)
	case 0x106:
		mem.Timers[1].WriteCnt(v, false)
	case 0x107:
		mem.Timers[1].WriteCnt(v, true)
	case 0x108:
		mem.Timers[2].WriteD(v, false)
	case 0x109:
		mem.Timers[2].WriteD(v, true)
	case 0x10A:
		mem.Timers[2].WriteCnt(v, false)
	case 0x10B:
		mem.Timers[2].WriteCnt(v, true)
	case 0x10C:
		mem.Timers[3].WriteD(v, false)
	case 0x10D:
		mem.Timers[3].WriteD(v, true)
	case 0x10E:
		mem.Timers[3].WriteCnt(v, false)
	case 0x10F:
		mem.Timers[3].WriteCnt(v, true)

	case 0x204:
		mem.Gamecard.ExMem.Write(v, 0)
	case 0x205:
		mem.Gamecard.ExMem.Write(v, 1)

	case 0x208:
		mem.irq9.WriteIME(v)
	case 0x209:
		return
	case 0x210:
		mem.irq9.WriteIE(v, 0)
	case 0x211:
		mem.irq9.WriteIE(v, 1)
	case 0x212:
		mem.irq9.WriteIE(v, 2)
	case 0x213:
		mem.irq9.WriteIE(v, 3)
	case 0x214:
		mem.irq9.WriteIF(v, 0)
	case 0x215:
		mem.irq9.WriteIF(v, 1)
	case 0x216:
		mem.irq9.WriteIF(v, 2)
	case 0x217:
		mem.irq9.WriteIF(v, 3)

	// vram reads - gbatek says read only, needed to match no$gba
	case 0x240:
		mem.Ppu.Vram.WriteCNT(addr, v)
	case 0x241:
		mem.Ppu.Vram.WriteCNT(addr, v)
	case 0x242:
		mem.Ppu.Vram.WriteCNT(addr, v)
	case 0x243:
		mem.Ppu.Vram.WriteCNT(addr, v)
	case 0x244:
		mem.Ppu.Vram.WriteCNT(addr, v)
	case 0x245:
		mem.Ppu.Vram.WriteCNT(addr, v)
	case 0x246:
		mem.Ppu.Vram.WriteCNT(addr, v)
	case 0x247:
		mem.WRAM.WriteCNT(v)
	case 0x248:
		mem.Ppu.Vram.WriteCNT(addr, v)
	case 0x249:
		mem.Ppu.Vram.WriteCNT(addr, v)

	case 0x300:
		mem.PostFlg.Write(v, true)

	case 0x304:
		mem.PowCnt.WriteCNT1(0, uint32(v), mem.Ppu)
	case 0x305:
		mem.PowCnt.WriteCNT1(1, uint32(v), mem.Ppu)

	default:
		//panic(fmt.Sprintf("WRTE UNKNOWN ARM9 IO ADDR %08X", addr))
		mem.IO[addr] = v
	}

	//if addr >= 0x1000 && addr <= 0x1004 {
	//    fmt.Printf("DISPSTAT %08X\n", binary.LittleEndian.Uint32(mem.IO[0x1000:]))
	//}
}

func (mem *Mem) ReadArm7IO(addr uint32) uint8 {

	//if addr != 0x180 && addr != 0x181 && addr < 0x3000 {
	//	fmt.Printf("READ ADDR %08X\n", addr)
	//}
	if addr >= 0x188 && addr < 0x190 {
		panic("READ IPC FIFO FROM BYTE OR HALF")
	}

	switch {
	case addr >= 0xB0 && addr < 0xE0:
		return mem.ReadDma(mem.dma7, addr)
	case addr >= 0x400 && addr < 0x600:
		return mem.Snd.Read(addr)
	}

	switch addr {
	case 0x4:
		return uint8(mem.Dispstat.A7)
	case 0x5:
		return uint8(mem.Dispstat.A7 >> 8)
	case 0x6:
		return uint8(mem.Vcount)
	case 0x7:
		return uint8(mem.Vcount >> 8)

	case 0x100:
		return mem.Timers[4].ReadD(false)
	case 0x101:
		return mem.Timers[4].ReadD(true)
	case 0x102:
		return mem.Timers[4].ReadCnt(false)
	case 0x103:
		return mem.Timers[4].ReadCnt(true)
	case 0x104:
		return mem.Timers[5].ReadD(false)
	case 0x105:
		return mem.Timers[5].ReadD(true)
	case 0x106:
		return mem.Timers[5].ReadCnt(false)
	case 0x107:
		return mem.Timers[5].ReadCnt(true)
	case 0x108:
		return mem.Timers[6].ReadD(false)
	case 0x109:
		return mem.Timers[6].ReadD(true)
	case 0x10A:
		return mem.Timers[6].ReadCnt(false)
	case 0x10B:
		return mem.Timers[6].ReadCnt(true)
	case 0x10C:
		return mem.Timers[7].ReadD(false)
	case 0x10D:
		return mem.Timers[7].ReadD(true)
	case 0x10E:
		return mem.Timers[7].ReadCnt(false)
	case 0x10F:
		return mem.Timers[7].ReadCnt(true)

	case 0x130:
		return mem.Keypad.readINPUT(false)
	case 0x131:
		return mem.Keypad.readINPUT(true)
	case 0x132:
		return mem.Keypad.readCNT(false)
	case 0x133:
		return mem.Keypad.readCNT(true)
	case 0x136:
		return mem.Keypad.readINPUT2()

	case 0x138:
		return mem.Rtc.Read()
	case 0x139:
		return 0
	case 0x13A:
		return 0
	case 0x13B:
		return 0

	case 0x180:
		return mem.Ipc.ReadSync(0, false)
	case 0x181:
		return mem.Ipc.ReadSync(1, false)
	case 0x184:
		return mem.Ipc.ReadCnt(0, false)
	case 0x185:
		return mem.Ipc.ReadCnt(1, false)
	case 0x186:
		return mem.Ipc.ReadCnt(2, false)
	case 0x187:
		return mem.Ipc.ReadCnt(3, false)

	case 0x1A0:
		return mem.Gamecard.AuxSpi.Read(0)
	case 0x1A1:
		return mem.Gamecard.AuxSpi.Read(1)
	case 0x1A2:
		return mem.Gamecard.AuxSpi.Read(2)
	case 0x1A3:
		return mem.Gamecard.AuxSpi.Read(3)
	case 0x1A4:
		return mem.Gamecard.RomCtrl.Read(0)
	case 0x1A5:
		return mem.Gamecard.RomCtrl.Read(1)
	case 0x1A6:
		return mem.Gamecard.RomCtrl.Read(2)
	case 0x1A7:
		return mem.Gamecard.RomCtrl.Read(3)

	case 0x100010:
		panic("READING GAMECARD READ IN FROM READ16 OR READ8")
	case 0x100011:
		panic("READING GAMECARD READ IN FROM READ16 OR READ8")
	case 0x100012:
		panic("READING GAMECARD READ IN FROM READ16 OR READ8")
	case 0x100013:
		panic("READING GAMECARD READ IN FROM READ16 OR READ8")

	case 0x1C0:
		return mem.Spi.ReadCNT(0)
	case 0x1C1:
		return mem.Spi.ReadCNT(1)
	case 0x1C2:
		return mem.Spi.ReadData()
	case 0x1C3:
		return 0

	case 0x204:
		return mem.Gamecard.ExMem.Read(0)
	case 0x205:
		return mem.Gamecard.ExMem.Read(1)

	case 0x206:
		return uint8(mem.WifiWaitCnt)
	case 0x207:
		return 0

	case 0x208:
		return mem.irq7.ReadIME()
	case 0x209:
		return 0
	case 0x210:
		return mem.irq7.ReadIE(0)
	case 0x211:
		return mem.irq7.ReadIE(1)
	case 0x212:
		return mem.irq7.ReadIE(2)
	case 0x213:
		return mem.irq7.ReadIE(3)
	case 0x214:
		return mem.irq7.ReadIF(0)
	case 0x215:
		return mem.irq7.ReadIF(1)
	case 0x216:
		return mem.irq7.ReadIF(2)
	case 0x217:
		return mem.irq7.ReadIF(3)

	case 0x240:
		return mem.Ppu.Vram.CNT_7
	case 0x241:
		return mem.WRAM.ReadCNT()

	case 0x300:
		return mem.PostFlg.Read(false)

	case 0x301:

		if *mem.halted7 {
			return 0b1000_0000
		} else {
			return 0b0000_0000
		}

	case 0x304:
		return mem.PowCnt.V2

	case 0x308:
		return uint8(mem.BiosProt)
	case 0x309:
		return uint8(mem.BiosProt >> 8)

	case 0x808000:
		return 0x40
	case 0x808001:
		return 0xC3

	default:
		//panic(fmt.Sprintf("READ UNKNOWN ARM7 IO ADDR %08X", addr))
		return mem.IO[addr]
	}
}

func (mem *Mem) WriteArm7IO(addr uint32, v uint8) {

	if addr >= 0x188 && addr < 0x190 {
		panic("WRITE IPC FIFO FROM BYTE OR HALF")
	}

	//if !(addr >= 0x208 && addr < 0x240) {
	//    fmt.Printf("ARM7 WRITE ADDR %08X V %02X\n", addr, v)
	//}

	switch {
	case addr < 0x4:
		mem.Ppu.Update(addr, uint32(v))

	case addr >= 0xB0 && addr < 0xE0:
		mem.WriteDma(mem.dma7, addr, v)
		return

	case addr >= 0x400 && addr < 0x600:
		mem.Snd.Write(addr, v)
		return
	}

	switch addr {
	case 0x4:
		mem.Dispstat.Write(v, false, false)
	case 0x5:
		mem.Dispstat.Write(v, true, false)

	case 0x6:
		mem.Vcount &^= 0xFF
		mem.Vcount |= uint32(v)
	case 0x7:
		mem.Vcount &= 0xFF
		mem.Vcount |= uint32(v) << 8

	case 0x100:
		mem.Timers[4].WriteD(v, false)
	case 0x101:
		mem.Timers[4].WriteD(v, true)
	case 0x102:
		mem.Timers[4].WriteCnt(v, false)
	case 0x103:
		mem.Timers[4].WriteCnt(v, true)
	case 0x104:
		mem.Timers[5].WriteD(v, false)
	case 0x105:
		mem.Timers[5].WriteD(v, true)
	case 0x106:
		mem.Timers[5].WriteCnt(v, false)
	case 0x107:
		mem.Timers[5].WriteCnt(v, true)
	case 0x108:
		mem.Timers[6].WriteD(v, false)
	case 0x109:
		mem.Timers[6].WriteD(v, true)
	case 0x10A:
		mem.Timers[6].WriteCnt(v, false)
	case 0x10B:
		mem.Timers[6].WriteCnt(v, true)
	case 0x10C:
		mem.Timers[7].WriteD(v, false)
	case 0x10D:
		mem.Timers[7].WriteD(v, true)
	case 0x10E:
		mem.Timers[7].WriteCnt(v, false)
	case 0x10F:
		mem.Timers[7].WriteCnt(v, true)

	case 0x130:
		return
	case 0x131:
		return
	case 0x132:
		mem.Keypad.writeCNT(v, false)
	case 0x133:
		mem.Keypad.writeCNT(v, true)

	case 0x138:
		mem.Rtc.Write(v)
	case 0x139:
		return
	case 0x13A:
		return
	case 0x13B:
		return

	case 0x180:
		mem.Ipc.WriteSync(v, 0, false)
	case 0x181:
		mem.Ipc.WriteSync(v, 1, false)

	case 0x184:
		mem.Ipc.WriteCnt(v, 0, false)
	case 0x185:
		mem.Ipc.WriteCnt(v, 1, false)
	case 0x186:
		mem.Ipc.WriteCnt(v, 2, false)
	case 0x187:
		mem.Ipc.WriteCnt(v, 3, false)

	case 0x1A0:
		mem.Gamecard.AuxSpi.Write(v, 0, false)
	case 0x1A1:
		mem.Gamecard.AuxSpi.Write(v, 1, false)
	case 0x1A2:
		mem.Gamecard.AuxSpi.Write(v, 2, false)
	case 0x1A3:
		mem.Gamecard.AuxSpi.Write(v, 3, false)
	case 0x1A4:
		mem.Gamecard.RomCtrl.Write(v, 0, false)
	case 0x1A5:
		mem.Gamecard.RomCtrl.Write(v, 1, false)
	case 0x1A6:
		mem.Gamecard.RomCtrl.Write(v, 2, false)
	case 0x1A7:
		mem.Gamecard.RomCtrl.Write(v, 3, false)

	case 0x1A8:
		mem.Gamecard.RomCtrl.WriteCmdOut(v, 0, false)
	case 0x1A9:
		mem.Gamecard.RomCtrl.WriteCmdOut(v, 1, false)
	case 0x1AA:
		mem.Gamecard.RomCtrl.WriteCmdOut(v, 2, false)
	case 0x1AB:
		mem.Gamecard.RomCtrl.WriteCmdOut(v, 3, false)
	case 0x1AC:
		mem.Gamecard.RomCtrl.WriteCmdOut(v, 4, false)
	case 0x1AD:
		mem.Gamecard.RomCtrl.WriteCmdOut(v, 5, false)
	case 0x1AE:
		mem.Gamecard.RomCtrl.WriteCmdOut(v, 6, false)
	case 0x1AF:
		mem.Gamecard.RomCtrl.WriteCmdOut(v, 7, false)
	case 0x1B0:
		mem.Gamecard.RomCtrl.WriteSeed(v, 0, 0, false)
	case 0x1B1:
		mem.Gamecard.RomCtrl.WriteSeed(v, 1, 0, false)
	case 0x1B2:
		mem.Gamecard.RomCtrl.WriteSeed(v, 2, 0, false)
	case 0x1B3:
		mem.Gamecard.RomCtrl.WriteSeed(v, 3, 0, false)
	case 0x1B4:
		mem.Gamecard.RomCtrl.WriteSeed(v, 0, 1, false)
	case 0x1B5:
		mem.Gamecard.RomCtrl.WriteSeed(v, 1, 1, false)
	case 0x1B6:
		mem.Gamecard.RomCtrl.WriteSeed(v, 2, 1, false)
	case 0x1B7:
		mem.Gamecard.RomCtrl.WriteSeed(v, 3, 1, false)
	case 0x1B8:
		mem.Gamecard.RomCtrl.WriteSeed(v, 4, 0, false)
	case 0x1B9:
		mem.Gamecard.RomCtrl.WriteSeed(v, 5, 0, false)
	case 0x1BA:
		mem.Gamecard.RomCtrl.WriteSeed(v, 4, 1, false)
	case 0x1BB:
		mem.Gamecard.RomCtrl.WriteSeed(v, 5, 1, false)

	case 0x100010:
		mem.Gamecard.RomCtrl.WriteCmdIn(v, 0, false)
	case 0x100011:
		mem.Gamecard.RomCtrl.WriteCmdIn(v, 1, false)
	case 0x100012:
		mem.Gamecard.RomCtrl.WriteCmdIn(v, 2, false)
	case 0x100013:
		mem.Gamecard.RomCtrl.WriteCmdIn(v, 3, false)

	case 0x1C0:
		mem.Spi.WriteCNT(0, v)
	case 0x1C1:
		mem.Spi.WriteCNT(1, v)
	case 0x1C2:
		mem.Spi.WriteData(v)
	case 0x1C3:
		return

	case 0x204:
		mem.Gamecard.ExMem.Write(v, 0)

	case 0x208:
		mem.irq7.WriteIME(v)
	case 0x209:
		return
	case 0x210:
		mem.irq7.WriteIE(v, 0)
	case 0x211:
		mem.irq7.WriteIE(v, 1)
	case 0x212:
		mem.irq7.WriteIE(v, 2)
	case 0x213:
		mem.irq7.WriteIE(v, 3)
	case 0x214:
		mem.irq7.WriteIF(v, 0)
	case 0x215:
		mem.irq7.WriteIF(v, 1)
	case 0x216:
		mem.irq7.WriteIF(v, 2)
	case 0x217:
		mem.irq7.WriteIF(v, 3)

	case 0x300:
		mem.PostFlg.Write(v, false)

	case 0x301:

		v >>= 6

		switch v {
		case 0:
			(*mem.halted7) = false
		case 2:
			(*mem.halted7) = true
		default:
			panic(fmt.Sprintf("UNKNOWN HALTCNT VALUE ARM7 %d", v))
		}
	case 0x304:
		mem.PowCnt.WriteCNT2(v)

	case 0x308:
		return

	default:
		//panic(fmt.Sprintf("WRTE UNKNOWN ARM7 IO ADDR %08X", addr))
		mem.IO[addr] = v
	}
}

func (m *Mem) ReadDma(dmas *[4]dma.DMA, addr uint32) uint8 {

	switch addr {
	case 0x00B8:
		return 0
	case 0x00B9:
		return 0
	case 0x00BA:
		return dmas[0].ReadControl(false)
	case 0x00BB:
		return dmas[0].ReadControl(true)
	case 0x00C4:
		return 0
	case 0x00C5:
		return 0
	case 0x00C6:
		return dmas[1].ReadControl(false)
	case 0x00C7:
		return dmas[1].ReadControl(true)
	case 0x00D0:
		return 0
	case 0x00D1:
		return 0
	case 0x00D2:
		return dmas[2].ReadControl(false)
	case 0x00D3:
		return dmas[2].ReadControl(true)
	case 0x00DC:
		return 0
	case 0x00DD:
		return 0
	case 0x00DE:
		return dmas[3].ReadControl(false)
	case 0x00DF:
		return dmas[3].ReadControl(true)
	}

	return 0
}

func (m *Mem) WriteDma(dmas *[4]dma.DMA, addr uint32, v uint8) {
	switch addr {
	case 0x00B0:
		dmas[0].WriteSrc(v, 0)
	case 0x00B1:
		dmas[0].WriteSrc(v, 1)
	case 0x00B2:
		dmas[0].WriteSrc(v, 2)
	case 0x00B3:
		dmas[0].WriteSrc(v, 3)
	case 0x00B4:
		dmas[0].WriteDst(v, 0)
	case 0x00B5:
		dmas[0].WriteDst(v, 1)
	case 0x00B6:
		dmas[0].WriteDst(v, 2)
	case 0x00B7:
		dmas[0].WriteDst(v, 3)
	case 0x00B8:
		dmas[0].WriteCount(v, false)
	case 0x00B9:
		dmas[0].WriteCount(v, true)
	case 0x00BA:
		dmas[0].WriteControl(v, false)
	case 0x00BB:
		dmas[0].WriteControl(v, true)
	case 0x00BC:
		dmas[1].WriteSrc(v, 0)
	case 0x00BD:
		dmas[1].WriteSrc(v, 1)
	case 0x00BE:
		dmas[1].WriteSrc(v, 2)
	case 0x00BF:
		dmas[1].WriteSrc(v, 3)
	case 0x00C0:
		dmas[1].WriteDst(v, 0)
	case 0x00C1:
		dmas[1].WriteDst(v, 1)
	case 0x00C2:
		dmas[1].WriteDst(v, 2)
	case 0x00C3:
		dmas[1].WriteDst(v, 3)
	case 0x00C4:
		dmas[1].WriteCount(v, false)
	case 0x00C5:
		dmas[1].WriteCount(v, true)
	case 0x00C6:
		dmas[1].WriteControl(v, false)
	case 0x00C7:
		dmas[1].WriteControl(v, true)
	case 0x00C8:
		dmas[2].WriteSrc(v, 0)
	case 0x00C9:
		dmas[2].WriteSrc(v, 1)
	case 0x00CA:
		dmas[2].WriteSrc(v, 2)
	case 0x00CB:
		dmas[2].WriteSrc(v, 3)
	case 0x00CC:
		dmas[2].WriteDst(v, 0)
	case 0x00CD:
		dmas[2].WriteDst(v, 1)
	case 0x00CE:
		dmas[2].WriteDst(v, 2)
	case 0x00CF:
		dmas[2].WriteDst(v, 3)
	case 0x00D0:
		dmas[2].WriteCount(v, false)
	case 0x00D1:
		dmas[2].WriteCount(v, true)
	case 0x00D2:
		dmas[2].WriteControl(v, false)
	case 0x00D3:
		dmas[2].WriteControl(v, true)
	case 0x00D4:
		dmas[3].WriteSrc(v, 0)
	case 0x00D5:
		dmas[3].WriteSrc(v, 1)
	case 0x00D6:
		dmas[3].WriteSrc(v, 2)
	case 0x00D7:
		dmas[3].WriteSrc(v, 3)
	case 0x00D8:
		dmas[3].WriteDst(v, 0)
	case 0x00D9:
		dmas[3].WriteDst(v, 1)
	case 0x00DA:
		dmas[3].WriteDst(v, 2)
	case 0x00DB:
		dmas[3].WriteDst(v, 3)
	case 0x00DC:
		dmas[3].WriteCount(v, false)
	case 0x00DD:
		dmas[3].WriteCount(v, true)
	case 0x00DE:
		dmas[3].WriteControl(v, false)
	case 0x00DF:
		dmas[3].WriteControl(v, true)
	}
}
