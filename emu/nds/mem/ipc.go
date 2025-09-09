package mem

import (
	"fmt"

	"github.com/aabalke/guac/emu/nds/cpu"
	"github.com/aabalke/guac/emu/nds/utils"
)

type IPC struct {
	SYNC7 uint16
	SYNC9 uint16

    Fifo7to9 Fifo
    Fifo9to7 Fifo

    Irq7 *cpu.Irq
    Irq9 *cpu.Irq
}

func (i *IPC) Init(irq7, irq9 *cpu.Irq) {
    i.Fifo7to9.Buffer = [0x10]uint32{}
    i.Fifo7to9.Empty = true
    i.Fifo9to7.Buffer = [0x10]uint32{}
    i.Fifo9to7.Empty = true

    i.Irq7 = irq7
    i.Irq9 = irq9
}

type Fifo struct {
    Buffer [0x10]uint32
    Length, Head, Tail uint8
    Value uint32
    Empty, Full, Error bool

    IrqEmpty, IrqNotEmpty bool

    Enabled bool
}

func (f *Fifo) Write(v uint32) (ok bool) {

    if f.Full {
        return false
    }

    f.Buffer[f.Tail] = v
    f.Tail = (f.Tail + 1) & 0xF
    f.Length++


    f.Empty = f.Length == 0
    f.Full = f.Length >= 0x10

    //fmt.Printf("FIFO WRIT % X %d\n", f.Buffer, f.Length)

    return true
}

func (f *Fifo) Read() (v uint32, ok bool) {


    if f.Empty {
        return f.Value, false
    }

    f.Value = f.Buffer[f.Head]
    f.Head = (f.Head + 1) & 0xF
    f.Length--


    f.Empty = f.Length == 0
    f.Full = f.Length >= 0x10

    //fmt.Printf("FIFO READ % X %d\n", f.Buffer, f.Length)

    return f.Value, true
}

func (i *IPC) WriteSync(v, b uint8, isArm9 bool) {

    local := &i.SYNC9
    remote := &i.SYNC7

	if !isArm9 {
        local = &i.SYNC7
        remote = &i.SYNC9
	}

    if b == 0 {

        *local &^= 0xF0
        *local |= uint16(v & 0xF0)

        return
    }


    *local &^= 0b100_1111 << 8
    *local |= uint16(v & 0b100_1111) << 8
    *remote &^= 0xF
    *remote |= uint16(v & 0xF)



    if irq := utils.BitEnabled(uint32(v), 5) && utils.BitEnabled(uint32(*remote), 14); irq {
        if isArm9 {
            i.Irq7.SetIRQ(cpu.IRQ_IPC_SYNC)
            return
        }
        i.Irq9.SetIRQ(cpu.IRQ_IPC_SYNC)
    }
}

func (i *IPC) ReadSync(b uint8, arm9 bool) uint8 {


	if arm9 {
        v := uint8(i.SYNC9 >> (b * 8))
        //fmt.Printf("READ SYNC B %02X IS ARM %t %02X SYNC9 %04X SYNC7 %04X\n", b, arm9, v, i.SYNC9, i.SYNC7)
        return v
	}
    v := uint8(i.SYNC7 >> (b * 8))
    //fmt.Printf("READ SYNC B %02X IS ARM %t %02X SYNC9 %04X SYNC7 %04X\n", b, arm9, v, i.SYNC9, i.SYNC7)
    return v
}

func (i *IPC) WriteCnt(v, b uint8, isArm9 bool) {

    //if isArm9 {
    //    fmt.Printf("WRITE CNT B %d V %02X\n", b, v)
    //}

    local := &i.Fifo9to7
    remote := &i.Fifo7to9

    if !isArm9 {
        local = &i.Fifo7to9
        remote = &i.Fifo9to7
    }

    switch b {
    case 0:
        local.IrqEmpty = utils.BitEnabled(uint32(v), 2)
        if utils.BitEnabled(uint32(v), 3) { // flush

            if isArm9 {
                i.Fifo9to7 = Fifo{
                    Value: 0,
                    Buffer: [0x10]uint32{},
                    Empty: true,
                    IrqEmpty: local.IrqEmpty,
                    IrqNotEmpty: local.IrqNotEmpty,
                    Error: local.Error,
                    Enabled: local.Enabled,
                }
            } else {
                i.Fifo7to9 = Fifo{
                    Value: 0,
                    Buffer: [0x10]uint32{},
                    Empty: true,
                    IrqEmpty: local.IrqEmpty,
                    IrqNotEmpty: local.IrqNotEmpty,
                    Error: local.Error,
                    Enabled: local.Enabled,
                }
            }

            fmt.Printf("FLUSHED FIFO\n")
        }
    case 1:
        remote.IrqNotEmpty = utils.BitEnabled(uint32(v), 2)

        if ackErr := utils.BitEnabled(uint32(v), 6); ackErr {
            local.Error = false
        }

        local.Enabled = utils.BitEnabled(uint32(v), 7)
    }
    i.updateIRQs()
}

func (i *IPC) ReadCnt(b uint8, isArm9 bool) uint8 {

    local := &i.Fifo9to7
    remote := &i.Fifo7to9

    if !isArm9 {
        local = &i.Fifo7to9
        remote = &i.Fifo9to7
    }

    v := uint8(0)

    switch b {
    case 0: 

        if local.Empty { v |= 1 }
        if local.Full { v |= 1 << 1 }
        if remote.IrqEmpty { v |= 1 << 2 }

    case 1: 

        if remote.Empty { v |= 1 }
        if remote.Full { v |= 1 << 1 }
        if remote.IrqNotEmpty { v |= 1 << 2 }
        if local.Error { v |= 1 << 6 }
        if local.Enabled { v |= 1 << 7 }
    }

    return v
}

func (i *IPC) WriteFifo(v uint32, isArm9 bool) {

    local := &i.Fifo9to7

    if !isArm9 {
        local = &i.Fifo7to9
    }

    if local.Enabled {
        ok := local.Write(v)
        if !ok {
            local.Error = true
        }
    }

    //fmt.Printf("WRITING TO FIFO %08X ARM9 %t\n", v, isArm9)
    i.updateIRQs()

    //if local.IrqNotEmpty && !local.Empty {
    //    if isArm9 {
    //        i.Irq7.SetIRQ(cpu.IRQ_IPC_RECV_FIFO)
    //    } else {
    //        i.Irq9.SetIRQ(cpu.IRQ_IPC_RECV_FIFO)
    //    }
    //}
}

func (i *IPC) ReadFifo(isArm9 bool) uint32 {

    local := &i.Fifo9to7
    remote := &i.Fifo7to9

    if !isArm9 {
        local = &i.Fifo7to9
        remote = &i.Fifo9to7
    }


    if !local.Enabled {
        return remote.Value
    }

    v, ok := remote.Read()
    if !ok {
        local.Error = true
        return remote.Value
    }

    //if remote.IrqEmpty && remote.Empty {
    //    if isArm9 {
    //        i.Irq7.SetIRQ(cpu.IRQ_IPC_SEND_FIFO)
    //    } else {
    //        i.Irq9.SetIRQ(cpu.IRQ_IPC_SEND_FIFO)
    //    }
    //}

    i.updateIRQs()

    return v
}

var irqEmptyFlag [2]bool
var irqNotEmptyFlag [2]bool

func (i *IPC) updateIRQs() {
    i.updateIrqFlagsCpu(true)
    i.updateIrqFlagsCpu(false)
}

func (i *IPC) updateIrqFlagsCpu(isArm9 bool) {

    local := &i.Fifo9to7
    remote := &i.Fifo7to9

    idx := 0

    if !isArm9 {
        local = &i.Fifo7to9
        remote = &i.Fifo9to7
        idx = 1
    }

	newEmptyFlag := local.Empty && local.IrqEmpty
	newDataFlag := !remote.Empty && remote.IrqNotEmpty

	if !irqEmptyFlag[idx] && newEmptyFlag {
        if isArm9 {
            i.Irq9.SetIRQ(cpu.IRQ_IPC_SEND_FIFO)
        } else {
            i.Irq7.SetIRQ(cpu.IRQ_IPC_SEND_FIFO)
        }
	}

	if !irqNotEmptyFlag[idx] && newDataFlag {
        if isArm9 {
            i.Irq9.SetIRQ(cpu.IRQ_IPC_RECV_FIFO)
        } else {
            i.Irq7.SetIRQ(cpu.IRQ_IPC_RECV_FIFO)
        }
	}

    irqEmptyFlag[idx] = newEmptyFlag
    irqNotEmptyFlag[idx] = newDataFlag
}

