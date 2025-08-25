package mem

import (
	_ "embed"

	"github.com/aabalke/guac/emu/nds/cart"
	"github.com/aabalke/guac/emu/nds/cpu"
	"github.com/aabalke/guac/emu/nds/mem/spi"
	"github.com/aabalke/guac/emu/nds/ppu"
)

//go:embed res_/bios7.bin
var arm7Bios []byte

//go:embed res_/bios9.bin
var arm9Bios []byte

type Mem struct {

    Tcm Tcm

	MainRam       [0x40_0000]uint8
	WRAM          WRAM
    Pram ppu.PRAM
    Vram ppu.VRAM
	OAMPal        [0x1000]uint8

	Arm7Bios [0x4000]uint8
	Arm9Bios [0x1000]uint8

	IO [0x100_0000]uint8

	arm9Irq *cpu.Irq
	arm7Irq *cpu.Irq

    arm9Dma *[4]DMA

    LowVector bool

    ppu *ppu.PPU

    exmem ExMem
    auxspi AuxSPI
    div Div
    sqrt Sqrt

    Cartridge *cart.Cartridge

    Vcount uint32
    Dispstat Dispstat
    Keypad Keypad

    Ipc IPC
    Spi spi.Spi
    Rtc Rtc
}

func NewMemory(dma9 *[4]DMA, irq7, irq9 *cpu.Irq, c *cart.Cartridge, ppu *ppu.PPU) Mem {
	m := Mem{
        arm9Dma: dma9,
        arm9Irq: irq9,
        arm7Irq: irq7,
        Cartridge: c,
        ppu: ppu,
    }

    // i believe this is default
    m.WRAM.WriteCNT(3)

    m.Keypad.KEYINPUT = 0x3FF
    m.Keypad.KEYINPUT2 = 0b100_0011

    m.Ipc.Init(irq7, irq9)

	m.LoadBios()

    m.Rtc.RegStatus1 = 0x02
    m.Rtc.RegStatus2 = 0x41

	return m
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

        if addr >= mem.Tcm.DtcmBase && addr < mem.Tcm.DtcmBase + mem.Tcm.DtcmSize {
            return mem.Tcm.Read(addr)
        }

		switch addr >> 24 {
		case 0x0, 0x1:
            return mem.Tcm.Read(addr)
		case 0x2:
			return mem.MainRam[addr & 0x3F_FFFF]
		case 0x3:
            return mem.WRAM.Read(addr, true)
		case 0x4:
			return mem.ReadArm9IO(addr - 0x400_0000)
		case 0x5:
            //panic("PRAM READ YAY")
            return mem.Pram.Read(addr)
		case 0x6:
            //panic("VRAM READ 9 YAY")
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
        return mem.WRAM.Read(addr, false)
    case 0x4:
        return mem.ReadArm7IO(addr-0x400_0000) //io
    case 0x5: // pal std
        return 0
    case 0x6:
        //panic("VRAM READ 7 YAY")
        return mem.Vram.Read(addr, false)
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
    switch addr {
    case 0x410_0000:
        return mem.Ipc.ReadFifo(arm9)
    default:
        a := uint32(mem.Read(addr+2, arm9)) | (uint32(mem.Read(addr+3, arm9)) << 8)
        b := uint32(mem.Read(addr, arm9)) | (uint32(mem.Read(addr+1, arm9)) << 8)
        return (a << 16) | b
    }
}

func (mem *Mem) Write(addr uint32, v uint8, arm9 bool) {

	if arm9 {

        if addr >= mem.Tcm.DtcmBase && addr < mem.Tcm.DtcmBase + mem.Tcm.DtcmSize {
            mem.Tcm.Write(addr, v)
            return
        }

		switch addr >> 24 {
		case 0x0, 0x1:
            mem.Tcm.Write(addr, v)
		case 0x2:
			mem.MainRam[addr&0x3F_FFFF] = v
		case 0x3:
            mem.WRAM.Write(addr, v, true)
		case 0x4:

            //fmt.Printf("WRITE IO %08X %02X\n", addr, v)

			mem.WriteArm9IO(addr-0x400_0000, v) //io
        case 0x5:
            //panic("PRAM WRITE YAY")
            mem.Pram.Write(addr, v)
		case 0x6:
            //panic("VRAM WRITE YAY")
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
        mem.WRAM.Write(addr, v, false)
    case 0x4:
        mem.WriteArm7IO(addr-0x400_0000, v) //io
    case 0x5: // pal std
    case 0x6:
        //panic("VRAM WRITE YAY")
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

    switch addr {
    case 0x400_0188: 
        mem.Ipc.WriteFifo(v, arm9)
    default:
        mem.Write(addr, uint8(v), arm9)
        mem.Write(addr+1, uint8(v>>8), arm9)
        mem.Write(addr+2, uint8(v>>16), arm9)
        mem.Write(addr+3, uint8(v>>24), arm9)
    }
}

func (mem *Mem) ReadArm9IO(addr uint32) uint8 {

	//if addr != 0x180 && addr != 0x181 && addr < 0x3000 {
	//	fmt.Printf("READ ADDR %08X\n", addr)
	//}

    if addr >= 0x188 && addr < 0x190 { panic("READ IPC FIFO FROM BYTE OR HALF")}

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

	case 0x00B8:
		return 0
	case 0x00B9:
		return 0
	case 0x00BA:
		return mem.arm9Dma[0].ReadControl(false)
	case 0x00BB:
		return mem.arm9Dma[0].ReadControl(true)
	case 0x00C4:
		return 0
	case 0x00C5:
		return 0
	case 0x00C6:
		return mem.arm9Dma[1].ReadControl(false)
	case 0x00C7:
		return mem.arm9Dma[1].ReadControl(true)
	case 0x00D0:
		return 0
	case 0x00D1:
		return 0
	case 0x00D2:
		return mem.arm9Dma[2].ReadControl(false)
	case 0x00D3:
		return mem.arm9Dma[2].ReadControl(true)
	case 0x00DC:
		return 0
	case 0x00DD:
		return 0
	case 0x00DE:
		return mem.arm9Dma[3].ReadControl(false)
	case 0x00DF:
		return mem.arm9Dma[3].ReadControl(true)

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

	//if addr >= 0xB0 && addr < 0xE0 {
	//	fmt.Printf("WRITE ADDR %08X V %02X\n", addr, v)
	//}
    if addr >= 0x188 && addr < 0x190 { panic("WRITE IPC FIFO FROM BYTE OR HALF")}

    if addr < 0x4 {
        mem.ppu.Update(addr, uint32(v))
    }

    if addr >= 0x280 && addr < 0x2B0 {
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
        mem.auxspi.Write(v, 0)
    case 0x1A1:
        mem.auxspi.Write(v, 1)
    case 0x1A2:
        mem.auxspi.WriteData(v, 0)
    case 0x1A3:
        mem.auxspi.WriteData(v, 1)

	case 0x00B0:
		mem.arm9Dma[0].WriteSrc(v, 0)
	case 0x00B1:
		mem.arm9Dma[0].WriteSrc(v, 1)
	case 0x00B2:
		mem.arm9Dma[0].WriteSrc(v, 2)
	case 0x00B3:
		mem.arm9Dma[0].WriteSrc(v, 3)
	case 0x00B4:
		mem.arm9Dma[0].WriteDst(v, 0)
	case 0x00B5:
		mem.arm9Dma[0].WriteDst(v, 1)
	case 0x00B6:
		mem.arm9Dma[0].WriteDst(v, 2)
	case 0x00B7:
		mem.arm9Dma[0].WriteDst(v, 3)
	case 0x00B8:
		mem.arm9Dma[0].WriteCount(v, false)
	case 0x00B9:
		mem.arm9Dma[0].WriteCount(v, true)
	case 0x00BA:
		mem.arm9Dma[0].WriteControl(v, false)
	case 0x00BB:
		mem.arm9Dma[0].WriteControl(v, true)
	case 0x00BC:
		mem.arm9Dma[1].WriteSrc(v, 0)
	case 0x00BD:
		mem.arm9Dma[1].WriteSrc(v, 1)
	case 0x00BE:
		mem.arm9Dma[1].WriteSrc(v, 2)
	case 0x00BF:
		mem.arm9Dma[1].WriteSrc(v, 3)
	case 0x00C0:
		mem.arm9Dma[1].WriteDst(v, 0)
	case 0x00C1:
		mem.arm9Dma[1].WriteDst(v, 1)
	case 0x00C2:
		mem.arm9Dma[1].WriteDst(v, 2)
	case 0x00C3:
		mem.arm9Dma[1].WriteDst(v, 3)
	case 0x00C4:
		mem.arm9Dma[1].WriteCount(v, false)
	case 0x00C5:
		mem.arm9Dma[1].WriteCount(v, true)
	case 0x00C6:
		mem.arm9Dma[1].WriteControl(v, false)
	case 0x00C7:
		mem.arm9Dma[1].WriteControl(v, true)
	case 0x00C8:
		mem.arm9Dma[2].WriteSrc(v, 0)
	case 0x00C9:
		mem.arm9Dma[2].WriteSrc(v, 1)
	case 0x00CA:
		mem.arm9Dma[2].WriteSrc(v, 2)
	case 0x00CB:
		mem.arm9Dma[2].WriteSrc(v, 3)
	case 0x00CC:
		mem.arm9Dma[2].WriteDst(v, 0)
	case 0x00CD:
		mem.arm9Dma[2].WriteDst(v, 1)
	case 0x00CE:
		mem.arm9Dma[2].WriteDst(v, 2)
	case 0x00CF:
		mem.arm9Dma[2].WriteDst(v, 3)
	case 0x00D0:
		mem.arm9Dma[2].WriteCount(v, false)
	case 0x00D1:
		mem.arm9Dma[2].WriteCount(v, true)
	case 0x00D2:
		mem.arm9Dma[2].WriteControl(v, false)
	case 0x00D3:
		mem.arm9Dma[2].WriteControl(v, true)
	case 0x00D4:
		mem.arm9Dma[3].WriteSrc(v, 0)
	case 0x00D5:
		mem.arm9Dma[3].WriteSrc(v, 1)
	case 0x00D6:
		mem.arm9Dma[3].WriteSrc(v, 2)
	case 0x00D7:
		mem.arm9Dma[3].WriteSrc(v, 3)
	case 0x00D8:
		mem.arm9Dma[3].WriteDst(v, 0)
	case 0x00D9:
		mem.arm9Dma[3].WriteDst(v, 1)
	case 0x00DA:
		mem.arm9Dma[3].WriteDst(v, 2)
	case 0x00DB:
		mem.arm9Dma[3].WriteDst(v, 3)
	case 0x00DC:
		mem.arm9Dma[3].WriteCount(v, false)
	case 0x00DD:
		mem.arm9Dma[3].WriteCount(v, true)
	case 0x00DE:
		mem.arm9Dma[3].WriteControl(v, false)
	case 0x00DF:
		mem.arm9Dma[3].WriteControl(v, true)

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
    case 0x240: mem.Vram.WriteCNT(addr, v)
    case 0x241: mem.Vram.WriteCNT(addr, v)
    case 0x242: mem.Vram.WriteCNT(addr, v)
    case 0x243: mem.Vram.WriteCNT(addr, v)
    case 0x244: mem.Vram.WriteCNT(addr, v)
    case 0x245: mem.Vram.WriteCNT(addr, v)
    case 0x246: mem.Vram.WriteCNT(addr, v)
    case 0x247:
        mem.WRAM.WriteCNT(v)
    case 0x248: mem.Vram.WriteCNT(addr, v)
    case 0x249: mem.Vram.WriteCNT(addr, v)
    case 0x24A: mem.Vram.WriteCNT(addr, v)
    case 0x24B: mem.Vram.WriteCNT(addr, v)
    case 0x24C: mem.Vram.WriteCNT(addr, v)
    case 0x24D: mem.Vram.WriteCNT(addr, v)
    case 0x24E: mem.Vram.WriteCNT(addr, v)
    case 0x24F: mem.Vram.WriteCNT(addr, v)

    case 0x304:
        mem.ppu.PowCnt1.V &^= 0xFF
        mem.ppu.PowCnt1.V |= uint16(v)
        mem.ppu.Update(addr, uint32(v))
    case 0x305:
        mem.ppu.PowCnt1.V &= 0xFF
        mem.ppu.PowCnt1.V |= uint16(v) << 8
        mem.ppu.Update(addr, uint32(v))
	default:
		mem.IO[addr] = v
	}
}

func (mem *Mem) ReadArm7IO(addr uint32) uint8 {

	//if addr != 0x180 && addr != 0x181 && addr < 0x3000 {
	//	fmt.Printf("READ ADDR %08X\n", addr)
	//}
    if addr >= 0x188 && addr < 0x190 { panic("READ IPC FIFO FROM BYTE OR HALF")}

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
        return mem.auxspi.Read(0)
    case 0x1A1:
        return mem.auxspi.Read(1)
    case 0x1A2:
        return mem.auxspi.Read(2)
    case 0x1A3:
        return mem.auxspi.Read(3)

    case 0x1C0:
        return mem.Spi.ReadCNT(0)
    case 0x1C1:
        return mem.Spi.ReadCNT(1)
    case 0x1C2:
        return mem.Spi.ReadData()
    case 0x1C3:
        return 0

    case 0x204:
        return mem.exmem.Read(0)
    case 0x205:
        return mem.exmem.Read(1)
	case 0x208:
		return mem.arm7Irq.ReadIME()
	case 0x210:
		return mem.arm7Irq.ReadIE(0)
	case 0x211:
		return mem.arm7Irq.ReadIE(1)
	case 0x212:
		return mem.arm7Irq.ReadIE(2)
	case 0x213:
		return mem.arm7Irq.ReadIE(3)
	case 0x214:
		return mem.arm7Irq.ReadIF(0)
	case 0x215:
		return mem.arm7Irq.ReadIF(1)
	case 0x216:
		return mem.arm7Irq.ReadIF(2)
	case 0x217:
		return mem.arm7Irq.ReadIF(3)
    case 0x240:
        return mem.Vram.CNT_7

    case 0x241:
        return mem.WRAM.ReadCNT()
	default:
		return mem.IO[addr]
	}
}

func (mem *Mem) WriteArm7IO(addr uint32, v uint8) {

	//if addr != 0x180 && addr != 0x181 && addr < 0x3000 {
	//	fmt.Printf("WRITE ADDR %08X V %02X\n", addr, v)
	//}

    if addr >= 0x188 && addr < 0x190 { panic("WRITE IPC FIFO FROM BYTE OR HALF")}

    if addr < 0x4 {
        mem.ppu.Update(addr, uint32(v))
    }

    if addr >= 0x240 && addr < 0x250 {
        //mem.Vram.WriteCNT(addr, v)
        //return
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
        mem.auxspi.Write(v, 0)
    case 0x1A1:
        mem.auxspi.Write(v, 1)
    case 0x1A2:
        mem.auxspi.WriteData(v, 0)
    case 0x1A3:
        mem.auxspi.WriteData(v, 1)
    case 0x1C0:
        mem.Spi.WriteCNT(0, v) 
    case 0x1C1:
        mem.Spi.WriteCNT(1, v) 
    case 0x1C2:
        mem.Spi.WriteData(v)
    case 0x1C3:
        return

    case 0x204:
        mem.exmem.Write(v, 0)
    case 0x205:
        mem.exmem.Write(v, 1)
	case 0x208:
		mem.arm7Irq.WriteIME(v)
	case 0x210:
		mem.arm7Irq.WriteIE(v, 0)
	case 0x211:
		mem.arm7Irq.WriteIE(v, 1)
	case 0x212:
		mem.arm7Irq.WriteIE(v, 2)
	case 0x213:
		mem.arm7Irq.WriteIE(v, 3)
	case 0x214:
		mem.arm7Irq.WriteIF(v, 0)
	case 0x215:
		mem.arm7Irq.WriteIF(v, 1)
	case 0x216:
		mem.arm7Irq.WriteIF(v, 2)
	case 0x217:
		mem.arm7Irq.WriteIF(v, 3)
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

    // temp attempt
    //mem.Write32(0x803FFC, 0x20057F4, true)
}
