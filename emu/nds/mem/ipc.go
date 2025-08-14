package mem

import (

	"github.com/aabalke/guac/emu/nds/cpu"
	"github.com/aabalke/guac/emu/nds/utils"
)

type IPC struct {
	SYNC7 uint16
	SYNC9 uint16

    Fifo7 Fifo
    Fifo9 Fifo

    Irq7 *cpu.Irq
    Irq9 *cpu.Irq
}

func (i *IPC) Init(irq7, irq9 *cpu.Irq) {
    i.Fifo7.Buffer = [0x10]uint32{}
    i.Fifo7.Empty = true
    i.Fifo7.Cnt = 1
    i.Fifo9.Buffer = [0x10]uint32{}
    i.Fifo9.Empty = true
    i.Fifo9.Cnt = 1

    i.Irq7 = irq7
    i.Irq9 = irq9
}

type Fifo struct {
    Buffer [0x10]uint32
    Length, Head, Tail uint8
    Value uint32
    Empty, Full, Error bool
    Cnt uint8

    Irq, IrqEmpty, IrqNotEmpty bool
}

func (f *Fifo) Write(v uint32) (ok bool) {

    if f.Length >= 0x10 {
        return false
    }

    f.Buffer[f.Tail] = v
    f.Tail = (f.Tail + 1) & 0xF
    f.Length++

    f.Empty = f.Length <= 0
    f.Full = f.Length >= 0x10

    if f.Empty {
        f.Cnt |= 1
    } else {
        f.Cnt &^= 1
    }

    if f.Full {
        f.Cnt |= 0b10
    } else {
        f.Cnt &^= 0b10
    }

    return true
}

func (f *Fifo) Read() (v uint32, ok bool) {

    if f.Length == 0 {
        return f.Value, false
    }

    f.Value = f.Buffer[f.Head]
    f.Head = (f.Head + 1) & 0xF
    f.Length--

    f.Empty = f.Length <= 0
    f.Full = f.Length >= 0x10

    if f.Empty {
        f.Cnt |= 1
    } else {
        f.Cnt &^= 1
    }

    if f.Full {
        f.Cnt |= 0b10
    } else {
        f.Cnt &^= 0b10
    }

    return f.Value, true
}

func (i *IPC) WriteSync(v uint8, isArm9 bool, arm7Irq, arm9Irq *cpu.Irq) {

    local := &i.SYNC9
    remote := &i.SYNC7

	if !isArm9 {
        local = &i.SYNC7
        remote = &i.SYNC9
	}

    *local &^= 0b110_1111 << 8
    *local |= uint16(v&0b110_1111) << 8
    *remote &^= 0xF
    *remote |= uint16(v & 0xF)

    if irq := utils.BitEnabled(uint32(v), 5) && utils.BitEnabled(uint32(*remote), 14); irq {
        if isArm9 {
            arm7Irq.SetIRQ(cpu.IRQ_IPC_SYNC)
            return
        }
        arm9Irq.SetIRQ(cpu.IRQ_IPC_SYNC)
    }
}

func (i *IPC) ReadSync(b uint8, arm9 bool) uint8 {

	if arm9 {
		return uint8(i.SYNC9 >> (b * 8))
	}

	return uint8(i.SYNC7 >> (b * 8))
}

func (i *IPC) WriteCnt(v, b uint8, isArm9 bool) {

    if isArm9 {

        switch b {
        case 0:
            i.Fifo9.IrqEmpty = utils.BitEnabled(uint32(v), 2)

        case 1:
            if ackErr := i.Fifo9.Error && utils.BitEnabled(uint32(v), 6); ackErr {
                i.Fifo9.Error = false
            }

            i.Fifo9.IrqNotEmpty = utils.BitEnabled(uint32(v), 2)
            i.Fifo9.Irq = utils.BitEnabled(uint32(v), 7)
        }

        return
    }

    switch b {
    case 0:
        i.Fifo7.IrqEmpty = utils.BitEnabled(uint32(v), 2)
    case 1:
        if ackErr := i.Fifo7.Error && utils.BitEnabled(uint32(v), 6); ackErr {
            i.Fifo7.Error = false
        }

        i.Fifo7.IrqNotEmpty = utils.BitEnabled(uint32(v), 2)
        i.Fifo7.Irq = utils.BitEnabled(uint32(v), 7)
    }

    return
}

func (i *IPC) ReadCnt(b uint8, isArm9 bool) uint8 {

    if isArm9 {

        //println("READING CNT ARM 9", i.Fifo9.Cnt)

        switch b {
        case 0: 
            v := i.Fifo9.Cnt

            if i.Fifo7.IrqEmpty {
                v |= 1 << 2
            }

            return v
        case 1: 
            v := i.Fifo7.Cnt
            if i.Fifo9.Error {
                v |= 1 << 6
            }

            if i.Fifo9.IrqNotEmpty {
                v |= 1 << 2
            }

            if i.Fifo9.Irq {
                v |= 1 << 7
            }

            return v
        }

        return 0
    }
    switch b {
    case 0:
        v := i.Fifo7.Cnt

        if i.Fifo9.IrqEmpty {
            v |= 1 << 2
        }

        return v
    case 1:
        v := i.Fifo9.Cnt

        if i.Fifo7.Error {
            v |= 1 << 6
        }
        if i.Fifo9.IrqNotEmpty {
            v |= 1 << 2
        }

        if i.Fifo9.Irq {
            v |= 1 << 7
        }

        return v
    }

    return 0
}

func (i *IPC) WriteFifo(v uint32, isArm9 bool) {

    if isArm9 {
        //if // disabled {}

        i.Fifo9.Write(v)
        //if !ok {
        //    i.Fifo7.Error = true
        //}

        //fmt.Printf("9 WRITE FIFO %v\n", i.Fifo9.Buffer)

        if i.Fifo7.Irq && i.Fifo7.IrqNotEmpty && !i.Fifo7.Empty {
            i.Irq7.SetIRQ(cpu.IRQ_IPC_RECV_FIFO)
        }

        return
    }

    //if // disabled {}
    //fmt.Printf("7 WRITE FIFO %v\n", i.Fifo7.Buffer)
    i.Fifo7.Write(v)

    if i.Fifo9.Irq && i.Fifo9.IrqNotEmpty && !i.Fifo7.Empty {
        i.Irq9.SetIRQ(cpu.IRQ_IPC_RECV_FIFO)
    }
}

func (i *IPC) ReadFifo(isArm9 bool) uint32 {

    if isArm9 {
        //if // disabled {}
        v, ok := i.Fifo7.Read()
        if !ok {
            i.Fifo9.Error = true
        }

        if i.Fifo7.Irq && i.Fifo7.IrqEmpty && i.Fifo7.Empty {
            i.Irq7.SetIRQ(cpu.IRQ_IPC_SEND_FIFO)
        }

        return v
    }

    v, ok := i.Fifo9.Read()
    if !ok {
        i.Fifo7.Error = true
    }

    if i.Fifo9.Irq && i.Fifo9.IrqEmpty && i.Fifo9.Empty {
        i.Irq9.SetIRQ(cpu.IRQ_IPC_SEND_FIFO)
    }

    return v
}
