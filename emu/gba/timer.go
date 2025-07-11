package gba

import (
	"github.com/aabalke33/guac/emu/gba/utils"
)

type Timers [4]Timer

type Timer struct {
    Gba *GBA
    Idx int
	CNT, D uint32
    SavedInitialValue uint32
    SavedCycles uint32
    Elapsed uint32
}

func (tt *Timers) Update(cycles uint32) {

    overflow := false

    for i := range 4 {
        t := tt[i]

        overflow = t.Update(overflow, cycles)

        tt[i] = t
    }
}

func (t *Timer) Update(overflow bool, cycles uint32) bool {

    if !t.isEnabled() {
        return false
    }

    increment := uint32(0)
    if t.isCascade() && overflow {
        increment = 1
    }

    if !t.isCascade() {

        t.Elapsed += cycles

        if freq := t.getCycles(); t.Elapsed >= freq {
            increment = t.Elapsed / freq
            t.Elapsed %= freq
        }
    }

    overflow = false

    for range increment {
        tmp := t.D + 1
        if tmp > 0xFFFF {
            t.D = t.SavedInitialValue
            overflow = true
            continue
        }
        t.D = tmp
    }

    if !overflow {
        return false
    }

    if aTick := int((t.Gba.Mem.ReadIODirect(0x82, 2) >> 10) & 1) == t.Idx; aTick {

        fifo := t.Gba.DigitalApu.FifoA

        fifo.Load()

        if refill := fifo.Length <= 0x10; refill {
            t.Gba.Dma[1].transferFifo()
        }
    }

    if bTick := int((t.Gba.Mem.ReadIODirect(0x82, 2) >> 14) & 1) == t.Idx; bTick {

        fifo := t.Gba.DigitalApu.FifoB

        fifo.Load()

        if refill := fifo.Length <= 0x10; refill {
            t.Gba.Dma[2].transferFifo()
        }
    }

    if t.isOverflowIRQ() {
        t.Gba.Irq.setIRQ(3 + uint32(t.Idx))
    }

    return true
}

func (t *Timer) ReadCnt(hi bool) uint8 {

    if hi {
        return uint8(t.CNT >> 8)
    }

    return uint8(t.CNT)
}

func (t *Timer) WriteCnt(v uint8, hi bool) {

    if hi { return }

    oldValue := t.CNT & 0xC7
    t.CNT = uint32(v) & 0xC7

    if setEnabled := utils.BitEnabled(uint32(v), 7) && !utils.BitEnabled(oldValue, 7); setEnabled {
    //if setEnabled := utils.BitEnabled(uint32(v), 7); setEnabled {
        t.D = t.SavedInitialValue
        t.Elapsed = 0
    }
}

func (t *Timer) ReadD(hi bool) uint8 {

    if hi {
        return uint8(t.D >> 8)
    }

    return uint8(t.D)
}

func (t *Timer) WriteD(v uint8, hi bool) {

    if hi {
        t.SavedInitialValue = (t.SavedInitialValue & 0xFF) | (uint32(v) << 8)
        return
    }

    t.SavedInitialValue = (t.SavedInitialValue & 0xFF00) | uint32(v)
}

func (t *Timer) getCycles() uint32 {

	freq := utils.GetVarData(t.CNT, 0, 1)
	switch freq {
    case 0: return 1
    case 1: return 64
    case 2: return 256
    case 3: return 1024
	}

    return 1
}

func (t *Timer) isCascade() bool {
	return utils.BitEnabled(t.CNT, 2)
}

func (t *Timer) isOverflowIRQ() bool {
    return utils.BitEnabled(t.CNT, 6)
}

func (t *Timer) isEnabled() bool {
    return utils.BitEnabled(t.CNT, 7)
}
