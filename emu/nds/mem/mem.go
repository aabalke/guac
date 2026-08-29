package mem

import (
	"os"

	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/emu/bios"
	"github.com/aabalke/guac/emu/gba/timer"
	"github.com/aabalke/guac/emu/nds/cart"
	"github.com/aabalke/guac/emu/nds/irq"
	"github.com/aabalke/guac/emu/nds/mem/dma"
	"github.com/aabalke/guac/emu/nds/mem/spi"
	"github.com/aabalke/guac/emu/nds/mem/wifi"
	"github.com/aabalke/guac/emu/nds/ppu"
	"github.com/aabalke/guac/emu/nds/snd"
)

type Irq interface {
	SetIRQ(irq uint32)
}

type Mem struct {
	Tcm     Tcm
	MainRam [0x40_0000]uint8
	WRAM    WRAM
	Oam     [0x800]uint8

	Arm7Bios *[]uint8
	Arm9Bios *[]uint8

	// this size is temp
	IO [0x100_0000]uint8

	halted7    *bool
	irq7, irq9 *irq.Irq
	dma7, dma9 *[4]dma.DMA

	arm7Pc *uint32

	Ppu       *ppu.PPU
	Cartridge *cart.Cartridge
	Wifi      *wifi.Wifi
	Snd       *snd.Snd

	Vcount      uint32
	Dispstat    Dispstat
	Key         *Key
	div         Div
	sqrt        Sqrt
	Ipc         *Ipc
	Spi         *spi.Spi
	Rtc         Rtc
	PostFlg     PostFlg
	PowCnt      *PowCnt
	BiosProt    BiosProt
	WifiWaitCnt WifiWaitCnt
	Timers7     [4]*timer.Timer
	Timers9     [4]*timer.Timer
	Bus7        Bus7
	Bus9        Bus9
}

type (
	BiosProt    uint16
	WifiWaitCnt uint8
)

func (m *Mem) InitMemory(
	arm7Pc *uint32,
	halted7 *bool,
	dma7, dma9 *[4]dma.DMA,
	irq7, irq9 *irq.Irq,
	c *cart.Cartridge,
	ppu *ppu.PPU,
	snd *snd.Snd,
) {
	m.halted7 = halted7
	m.dma7 = dma7
	m.dma9 = dma9
	m.irq9 = irq9
	m.irq7 = irq7
	m.Cartridge = c
	m.Ppu = ppu
	m.arm7Pc = arm7Pc
	m.Snd = snd

	// i believe this is default
	m.WRAM.WriteCNT(3)

	m.BiosProt = 0x1204
	m.WifiWaitCnt = 0x30

	m.Key = NewKey(irq7, irq9)
	m.Ipc = NewIpc(irq7, irq9)
	m.PowCnt = NewPowCnt(ppu)
	m.Spi = spi.NewSpi(&m.Key.Input2)
	m.Wifi = wifi.NewWifi()

	m.LoadBios()

	m.Rtc.InitRtc()

	m.Bus7 = Bus7{M: m}
	m.Bus9 = Bus9{M: m}
}

func (mem *Mem) DirectBootMemory() {
	setBiosRam(mem, mem.Cartridge.ChipId)
}

func (mem *Mem) LoadBios() {
	mem.Arm7Bios = &bios.BiosNtrArm7
	mem.Arm9Bios = &bios.BiosNtrArm9
	b := &config.Conf.Nds.Bios

	if b.Arm7Path != "" {
		if buf, err := os.ReadFile(b.Arm7Path); err == nil {
			mem.Arm7Bios = &buf
		}
	}

	if b.Arm9Path != "" {
		if buf, err := os.ReadFile(b.Arm9Path); err == nil {
			mem.Arm9Bios = &buf
		}
	}
}
