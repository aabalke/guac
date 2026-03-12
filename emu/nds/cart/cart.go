package cart

import (
	"encoding/binary"
	"fmt"
	"log"
	"time"

	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/emu/cpu"
	"github.com/aabalke/guac/emu/nds/mem/dma"
)

type Cartridge struct {
	Rom     []uint8
	RomPath string
	RomLen  int

	Sav     []uint8
	SavPath string
	SavLen  int

	//io
	Header  Header
	ExMem   ExMem
	AuxSpi  AuxSpi
	RomCtrl RomCtrl

	irq7, irq9 *cpu.Irq
	dma7, dma9 *[4]dma.DMA
	Backup     *Backup

	// fields
	SaveFlag       bool
	RomTransferIrq bool
	NDSSlotEnabled bool
	Status         uint8
	Buffer         []uint8
	ChipId         [4]uint8
}

func NewCartridge(romPath, savPath string, bios *[0x4000]uint8, irq7, irq9 *cpu.Irq, dma7, dma9 *[4]dma.DMA) *Cartridge {

	c := &Cartridge{
		RomPath: romPath,
		SavPath: savPath,
		irq7:    irq7,
		irq9:    irq9,
		dma7:    dma7,
		dma9:    dma9,
	}

	var ok bool
	c.Rom, c.RomLen, ok = readFile(romPath)
	if !ok {
		panic("could not read rom path")
	}

	c.Header = NewHeader(c)

    if !c.Header.Decrypted {
        NewKey1(bios, &c.Rom).DecryptCard()
    }

	code := binary.LittleEndian.Uint32(c.Header.GameCode)
	c.Backup = NewBackup(c)
	c.readSave(savPath, code)
	c.Backup.setCartType()

	// matches no cash nitrofs test
	c.WriteExMem(0x80, 0)
	c.WriteExMem(0xE8, 1)
	// always set
	c.ExMem.v |= 1 << 13

	c.RomCtrl.isReady = true
	c.RomCtrl.Key1Gap2Length = 0x18

	// if skipping bios start in Key2
	c.Status = GAMECARD_STAT_KY2
	c.ChipId = createChipId(c.Backup)

	c.InitSaveLoop()

	return c
}

//go:inline
func (c *Cartridge) checkAccessRights(arm9 bool) bool {
	allowed9 := arm9 && !c.ExMem.isCartAccessArm7
	allowed7 := !arm9 && c.ExMem.isCartAccessArm7
	return allowed7 || allowed9
}

func (c *Cartridge) readSave(savPath string, gamecode uint32) {

	if romData, ok := roms[gamecode]; ok {
		// in db
		c.Backup.Size = romData.Size
		c.Backup.MemType = romData.BackupType
	} else {
		// i believe this are good values
		c.Backup.Size = 0x80_0000
		c.Backup.MemType = 0
	}

	c.Sav = make([]uint8, c.Backup.Size)
	for i := range c.Sav {
		c.Sav[i] = 0xFF
	}

	sav, length, ok := readFile(savPath)
	if ok {

		for i := range length {
			c.Sav[i] = sav[i]
		}
	}
}

const (
	GAMECARD_STAT_RAW = 0
	GAMECARD_STAT_K1A = 1
	GAMECARD_STAT_K1B = 2
	GAMECARD_STAT_KY2 = 3
)

func createChipId(b *Backup) [4]uint8 {

	var id [4]uint8

	id[0] = 0xC2

	const MB = 1024 * 1024

	if b.Size >= MB && b.Size <= 128*MB {
		v := ((b.Size >> 20) - 1)
		id[1] = uint8(v)
		id[2] = uint8(v >> 8)
	} else {
		v := (0x100 - (b.Size >> 28))
		id[1] = uint8(v)
		id[2] = uint8(v >> 8)
	}

	if nandFlag := b.MemType >= 8 && b.MemType < 11; nandFlag {
		id[3] = 0x08
	}

	// if dsi id[3] |= 0x4

	return id
}

func (c *Cartridge) RunRom(arm9 bool) {

	r := &c.RomCtrl

	//log.Printf("RUNNING COMMAND %X %08X\n", r.Command, r.v)

	// Data Block size   (0=None, 1..6=100h SHL (1..6) bytes, 7=4 bytes)
	switch r.BlockSizeBits {
	case 0b0:
		r.BlockSize = 0
	case 0b111:
		r.BlockSize = 4
	default:
		r.BlockSize = 0x100 << r.BlockSizeBits
	}

	buffer := make([]uint8, r.BlockSize)

	switch c.Status {
	//case GAMECARD_STAT_RAW:
	//case GAMECARD_STAT_K1A:
	//case GAMECARD_STAT_K1B:
	case GAMECARD_STAT_KY2:

		const (
			DATA_READ    = 0xB7
			GET_CHIP_ID3 = 0xB8

			//NAND_STAT = 0xD6
		)

		switch r.Command[0] {
		case DATA_READ:

			addr := binary.BigEndian.Uint32(r.Command[1:5])

			// Addresses that do exceed the ROM size do mirror to the valid address range (that includes mirroring non-loadable regions like 0..7FFFh to "8000h+(addr AND 1FFh)"; some newer games are using this behaviour for some kind of anti-piracy checks).
			if addr >= uint32(c.RomLen) {
				addr %= uint32(c.RomLen)
			}

			if addr < 0x8000 {
				addr &= 0x01FF
				addr += 0x8000
			}

			for i := range uint32(len(buffer)) {
				buffer[i] = c.Rom[addr+i]
			}

			// todo
			// the datastream wraps to the begin of the current 4K block when address+length crosses a 4K boundary (1000h bytes)

		case GET_CHIP_ID3:

			buffer = c.ChipId[:]
			//fmt.Printf("CHIP ID = % X\n", r.Gamecard.Buffer)

		//case NAND_STAT, 0x94:

		//    //fmt.Printf("READING NAND STATUS ON Gamecard Key2\n")

		//    // 0x20 is value on startup

		//    // this is temp (0xFF) to force next
		//    //r.Gamecard.Buffer = []uint8{0x20, 0x20, 0x20, 0x20}
		//    //r.Gamecard.Buffer = []uint8{0x0, 0x0, 0x0, 0x0}
		//    buffer = r.Gamecard.ChipId[:]

		//case 0xB5:
		//    fmt.Printf("READING NAND HIGHZ ON Gamecard Key2\n")
		//    r.Gamecard.Buffer = nil //[]uint8{0,0,0,0}
		//    r.Gamecard.Transfer(true)

		default:
			panic(fmt.Sprintf("Unsupported Gamecard Key2 Cmd %02X", r.Command[0]))
			buffer = nil //[]uint8{0,0,0,0}
		}
		c.Buffer = buffer
		c.RomTransfer(true, arm9)
	default:
		panic("BAD GAMECARD STATUS")
	}
}

func (c *Cartridge) RomTransfer(initial bool, arm9 bool) {

	if len(c.Buffer) == 0 {

		c.RomCtrl.v &^= (1 << 31)

		c.RomCtrl.Active = false
		c.RomCtrl.isReady = false

		switch {
		case !c.RomTransferIrq:
			return
		case arm9:
			c.irq9.SetIRQ(cpu.IRQ_CARD_TRANS_COMPLETE)
			return
		default:
			c.irq7.SetIRQ(cpu.IRQ_CARD_TRANS_COMPLETE)
			return
		}
	}

	// calc accurate clkrate

	c.RomCtrl.DataOut = binary.LittleEndian.Uint32(c.Buffer[0:4])
	c.Buffer = c.Buffer[4:]

	c.RomCtrl.isReady = true

	for i := range 4 {
		c.dma7[i].GamecartTransfer(false, initial)
		c.dma9[i].GamecartTransfer(true, initial)
	}
}

func (c *Cartridge) InitSaveLoop() {

	if config.Conf.Nds.DisableSaves {
		return
	}

	saveTicker := time.Tick(time.Second)

	go func() {
		for range saveTicker {
			if c.SaveFlag {
				log.Printf("Saving Game Path: %s\n", c.SavPath)
				writeFile(c.SavPath, c.Sav[:])
				c.SaveFlag = false
			}
		}
	}()
}
