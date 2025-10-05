package mem

import (
	"encoding/binary"
	"fmt"
	"log"

	"github.com/aabalke/guac/emu/nds/cart"
	"github.com/aabalke/guac/emu/nds/cpu"
	"github.com/aabalke/guac/emu/nds/mem/dma"
	"github.com/aabalke/guac/emu/nds/utils"
)

const (
    GAMECARD_STAT_RAW = 0
    GAMECARD_STAT_K1A = 1
    GAMECARD_STAT_K1B = 2
    GAMECARD_STAT_KY2 = 3
)

type Gamecard struct {
    ExMem   *ExMem
    AuxSpi  *AuxSPI
    RomCtrl *RomCtrl

    Key2 Key2

    IsArm7 bool
    RomTransferIrq bool
    NDSSlotEnabled bool

    Status uint8
    Buffer []uint8

    Cartridge *cart.Cartridge
    Backup *cart.Backup

    irq7, irq9 *cpu.Irq
    dma7, dma9 *[4]dma.DMA

    ChipId [4]uint8
}

func (g *Gamecard) Init(irq7, irq9 *cpu.Irq, dma7, dma9 *[4]dma.DMA, c *cart.Cartridge) {

    g.irq7 = irq7
    g.irq9 = irq9
    g.dma7 = dma7
    g.dma9 = dma9
    g.Cartridge = c

    g.Backup = &cart.Backup{}
    g.Backup.Init()

    g.ExMem = &ExMem{Gamecard: g}
    g.AuxSpi = &AuxSPI{Gamecard: g}
    g.RomCtrl = &RomCtrl{Gamecard: g}

    g.ExMem.Gamecard   = g
    g.AuxSpi.Gamecard  = g
    g.RomCtrl.Gamecard = g

    // matches no cash nitrofs test
    g.ExMem.Write(0x80, 0)
    g.ExMem.Write(0xE8, 1)

    // always set
    g.ExMem.v |= 1 << 13

    g.RomCtrl.isReady = true
    g.RomCtrl.Key1Gap2Length = 0x18

    g.Key2 = NewDefaultKey2()

    // if skipping bios start in Key2
    g.Status = GAMECARD_STAT_KY2

    // if this gets changed, fix on init
    //g.ChipId = [4]uint8{0xFF, 0xFF, 0xFF, 0xFF}
    //g.ChipId = [4]uint8{0x00, 0x01, 0x02, 0x03}
    g.ChipId = [4]uint8{0xC2, 0x7F, 0x00, 0x80}
}

type ExMem struct {
    Gamecard *Gamecard
	isGBAAccessArm7         bool
	isCartAccessArm7         bool
	isMainMemorySync         bool
	isMainMemoryPriorityArm7 bool
	v                        uint16
}

// gba values are separate instances on arm7 and arm9

func (e *ExMem) Read(b uint8) uint8 {
    //fmt.Printf("R EXMEM %04X\n", e.v)
	return uint8(e.v >> (b << 3))
}

func (e *ExMem) Write(v uint8, b uint8) {

	// top bits can only be written by arm9

	e.v &^= 0xFF << (b << 3)
	e.v |= uint16(v) << (b << 3)
    e.v |= 1 << 13 // always set

    //fmt.Printf("W EXMEM %04X V %02X B %02d\n", e.v, v, b)

    switch b {
    case 0:
		e.isGBAAccessArm7 = utils.BitEnabled(uint32(v), 7)
    case 1:
		e.isCartAccessArm7 = utils.BitEnabled(uint32(v), 3)
		e.isMainMemorySync = utils.BitEnabled(uint32(v), 6)
		e.isMainMemoryPriorityArm7 = utils.BitEnabled(uint32(v), 7)
	}
}

type AuxSPI struct {
    Gamecard *Gamecard
    Baudrate uint8
    Hold bool
    IsBackup bool

    Value uint8
    Req, Res []uint8
}

func (a *AuxSPI) Read(b uint8) uint8 {

    switch b {
    case 0:

        v := a.Baudrate

        if a.Hold {
            v |= 0b100_0000
        }

        return v

    case 1:

        v := uint8(0)

        if a.IsBackup {
            v |= 0b10_0000
        }
        if a.Gamecard.RomTransferIrq {
            v |= 0b100_0000
        }
        if a.Gamecard.NDSSlotEnabled {
            v |= 0b1000_0000
        }

        return v

    case 2:
        return a.Value
    default:
        return 0
    }
}

func (a *AuxSPI) Write(v uint8, b uint8, arm9 bool) {

    if arm9 && a.Gamecard.ExMem.isCartAccessArm7 {
        return
    }

    if !arm9 && !a.Gamecard.ExMem.isCartAccessArm7 {
        return
    }

    switch b {
    case 0:
        //fmt.Printf("W AUXSPI V %02X B %02d\n", v, b)
        a.Baudrate = v & 0b11
        a.Hold = utils.BitEnabled(uint32(v), 6)

        a.Gamecard.Backup.WrittenCnt = true
        return
    case 1:
        //fmt.Printf("W AUXSPI V %02X B %02d\n", v, b)
        // top bits can only be written by arm9

        wasBackup := a.IsBackup
        a.IsBackup = utils.BitEnabled(uint32(v), 5)
        a.Gamecard.RomTransferIrq = utils.BitEnabled(uint32(v), 6)
        a.Gamecard.NDSSlotEnabled = utils.BitEnabled(uint32(v), 7)

        if a.IsBackup && !wasBackup {

            //fmt.Println("BEGIN TFX")

            if a.Req == nil {
                a.Req = make([]uint8, 16)
            }
            a.Req = a.Req[:0]
            a.Res = nil
        }

        a.Gamecard.Backup.WrittenCnt = true
        return
    case 2:

        if !a.IsBackup || a.Baudrate != 0 {
            log.Printf("Attempted to Write Data to Rom through AUXSPI.\n")
            //panic("AUXSPI")
            return
        }

        //fmt.Printf("W AUXDATA B %02d\n", b)
        //panic("AUXSPI DATA WRITE")

        a.WriteData(v)
        return
    //case 3:
        // do writes here transfer?
    }
}

func (a *AuxSPI) WriteData(v uint8) {

    var value uint8

    if len(a.Res) > 0 {
        value = a.Res[0]
        a.Res = a.Res[1:]
    }

    if len(a.Res) == 0 {
        var stat uint8
        a.Req = append(a.Req, v)

        a.Res, stat = a.Gamecard.Backup.Transfer(a.Req)

        if stat == cart.STAT_DONE {
            a.Req = a.Req[:0]
        }
    }

    a.Value = value

    if !a.Hold {
        //fmt.Println("FINISH TFX")
    }
}

type RomCtrl struct {
    Gamecard *Gamecard
    Key1GapLength uint32
    Key2EncryptionEnabled bool
    Key2ApplySeed bool
    Key1Gap2Length uint32
    Key2EncryptCmds bool
    isReady bool
    BlockSizeBits uint32
    CLKRate bool
    Key1GapCLK bool
    RESBRelease bool
    isWrite bool
    Active bool
    v uint32

    seed0, seed1 uint64

    BlockSize uint32

    Command [8]uint8
    DataOut uint32
}

func (r *RomCtrl) Read(b uint8) uint8 {

    //fmt.Println("READ  ROM CTRL")

    switch b {
    case 1:
        return uint8((r.v &^ (1 << 7)) >> (b * 8))
    case 2:
        v :=  uint8((r.v) >> (b * 8))

        if r.isReady {
            v |= 0b1000_0000
        } else {
            v &^= 0b1000_0000
        }

        return v
    }

    return uint8(r.v >> (b * 8))
}

func (r *RomCtrl) Write(v uint8, b uint8, arm9 bool) {

    if arm9 && r.Gamecard.ExMem.isCartAccessArm7 {
        return
    }

    if !arm9 && !r.Gamecard.ExMem.isCartAccessArm7 {
        return
    }

    // defaults in bios / etc key1 vs normal?

    switch b {
    case 0:
        r.v &^= (0xFF << (b * 8))
        r.v |= (uint32(v) << (b * 8))

        r.Key1GapLength &^= 0xFF
        r.Key1GapLength |= uint32(v)

    case 1:

        r.v &^= (0xFF << (b * 8))
        r.v |= (uint32(v) << (b * 8))
        r.Key1GapLength &= 0xFF
        r.Key1GapLength |= uint32(v) << 8

        r.Key2EncryptionEnabled = utils.BitEnabled(uint32(v), 5)
        r.Key2ApplySeed = utils.BitEnabled(uint32(v), 7)

        if r.Key2EncryptionEnabled {
            r.UpdateEncryption()
        }

    case 2:
        r.v &^= (0b0111_1111 << (b * 8))
        r.v |= (uint32(v & 0b0111_1111) << (b * 8))

        r.Key1Gap2Length = max(0x18, uint32(v & 0x1F))
        r.Key2EncryptCmds = utils.BitEnabled(uint32(v), 6)

    case 3:
        r.v &^= (0b1101_1111 << (b * 8))
        r.v |= (uint32(v) << (b * 8))

        r.BlockSizeBits = uint32(v & 0b111)
        r.CLKRate = utils.BitEnabled(uint32(v), 3)
        r.Key1GapCLK = utils.BitEnabled(uint32(v), 4)

        if !r.RESBRelease {
            r.v |= 1 << 29
            r.RESBRelease = utils.BitEnabled(uint32(v), 5)
        }

        r.isWrite = utils.BitEnabled(uint32(v), 6)
        r.Active = utils.BitEnabled(uint32(v), 7)

        //log.Printf("RomCtrl Write of V %02X at B %02X. Output %08X. Ready %t\n", v, b, r.v, r.isReady)
        if r.Active {
            r.Run(arm9)
        }
    }

}

func (r *RomCtrl) WriteCmdOut(v, b uint8, arm9 bool) {

    if arm9 && r.Gamecard.ExMem.isCartAccessArm7 {
        return
    }

    if !arm9 && !r.Gamecard.ExMem.isCartAccessArm7 {
        return
    }
    //log.Printf("W CMD OUT V %02X, B %02X\n", v, b)
    r.Command[b] = v
}
func (r *RomCtrl) WriteCmdIn(v, b uint8, arm9 bool) {

    if arm9 && r.Gamecard.ExMem.isCartAccessArm7 {
        return
    }

    if !arm9 && !r.Gamecard.ExMem.isCartAccessArm7 {
        return
    }
    //log.Printf("W CMD IN  V %02X, B %02X\n", v, b)
}
func (r *RomCtrl) ReadCmdIn(arm9 bool) uint32 {


    v := r.DataOut

    //fmt.Printf("READ CMD IN V %08X\n", v)

    if r.isReady {
        r.isReady = false
        r.Gamecard.Transfer(false, arm9)

    } else {
        fmt.Printf("WARNING GAMECARD ROM READ WITHOUT PENDING DATA\n")
    }

    //log.Printf("R CMD IN B %08X CTRL %08X\n", v, r.v)

    return v

    return r.DataOut
}

func (r *RomCtrl) WriteSeed(v, b, seed uint8, arm9 bool) {

    if arm9 && r.Gamecard.ExMem.isCartAccessArm7 {
        return
    }

    if !arm9 && !r.Gamecard.ExMem.isCartAccessArm7 {
        return
    }
    log.Printf("W SEED    V %02X, B %02X\n", v, b)

    s := &r.seed0

    if seed == 1 {
        s = &r.seed1
    }

    (*s) &^= 0xFF << (b * 8)
    (*s) |=  uint64(v) << (b * 8)
}

func (r *RomCtrl) UpdateEncryption() {
    //log.Printf("Updating Encyption Key2\n")
    r.Gamecard.Key2 = NewKey2(r.seed0, r.seed1)
}

func (r *RomCtrl) Run(arm9 bool) {

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

    switch r.Gamecard.Status {
    //case GAMECARD_STAT_RAW:
    //case GAMECARD_STAT_K1A:
    //case GAMECARD_STAT_K1B:
    case GAMECARD_STAT_KY2:

        const(
            DATA_READ = 0xB7
            GET_CHIP_ID3 = 0xB8

            //NAND_STAT = 0xD6
        )

        switch r.Command[0] {
        case DATA_READ:

            addr := binary.BigEndian.Uint32(r.Command[1:5])

            // Addresses that do exceed the ROM size do mirror to the valid address range (that includes mirroring non-loadable regions like 0..7FFFh to "8000h+(addr AND 1FFh)"; some newer games are using this behaviour for some kind of anti-piracy checks).
            if addr >= uint32(r.Gamecard.Cartridge.RomLength) {
                addr %= r.Gamecard.Cartridge.RomLength
            }

            if addr < 0x8000 {
                addr &= 0x01FF
                addr += 0x8000
            }

            for i := range uint32(len(buffer)) {
                buffer[i] = r.Gamecard.Cartridge.Rom[addr + i]
            }

            // todo 
            // the datastream wraps to the begin of the current 4K block when address+length crosses a 4K boundary (1000h bytes)

        case GET_CHIP_ID3:

            buffer = r.Gamecard.ChipId[:]
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
        r.Gamecard.Buffer = buffer
        r.Gamecard.Transfer(true, arm9)
    default: panic("BAD GAMECARD STATUS")
    }
}

func (g *Gamecard) Transfer(initial bool, arm9 bool) {

    //fmt.Printf("INIT %t LEN %d\n", initial, len(g.Buffer))

    if len(g.Buffer) == 0 {

        g.RomCtrl.v &^= (1 << 31)

        g.RomCtrl.Active = false
        g.RomCtrl.isReady = false

        if g.RomTransferIrq {
            if arm9 {

                g.irq9.SetIRQ(cpu.IRQ_CARD_TRANS_COMPLETE)
            } else {
                g.irq7.SetIRQ(cpu.IRQ_CARD_TRANS_COMPLETE)
            }
        }

        //log.Printf("FINISHED GAMECARD %08X\n", g.RomCtrl.v)

        return
    }

    // calc accurate clkrate

    g.RomCtrl.DataOut = binary.LittleEndian.Uint32(g.Buffer[0:4])
    g.Buffer = g.Buffer[4:]

    g.RomCtrl.isReady = true

    for i := range 4 {
        g.dma7[i].GamecartTransfer(false, initial)
        g.dma9[i].GamecartTransfer(true, initial)
    }
}
