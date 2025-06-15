package gba

import (
	"github.com/aabalke33/guac/emu/gba/utils"
)

type Timers [4]Timer

//func (tt *Timers) Increment(newCycles uint32) {
//
//    for i := range tt {
//
//        t := &tt[i]
//
//        if !t.isEnabled() || t.isCascade() {
//            continue
//        }
//        t.SavedCycles += newCycles
//
//        //fmt.Printf("Incrementing Timer %d, count=%04X Saved Cycles %08X\n", t.Idx, t.D, t.SavedCycles)
//
//        for range t.SavedCycles / t.getCycles() {
//
//            overflow := t.Increment(false)
//            if overflow {
//                if AAAA {
//                    fmt.Printf("Timer %d overflow: D=0x%04X, reload=0x%04X\n", t.Idx, t.D, t.SavedInitialValue)
//                }
//
//                if i < 3 && tt[i+1].isCascade() {
//                    tt.cascade(i)
//                } else {
//                    t.CNT = t.CNT &^ 0b1000_0000 // this breaks some fixes others
//                }
//
//                if t.isOverflowIRQ() {
//                    t.raiseIRQ()
//                }
//            }
//        }
//
//        t.SavedCycles %= t.getCycles()
//    }
//}
//
//func (tt *Timers) cascade(overflowTimerIdx int) {
//
//    if notPossible := overflowTimerIdx == 4; notPossible {
//        panic("OVERFLOW TIMER I N$")
//        return
//    }
//
//    cascadeIdx := overflowTimerIdx + 1
//
//    if !tt[cascadeIdx].isEnabled() {
//        return
//    }
//
//    overflow := tt[cascadeIdx].Increment(true)
//
//    if overflow {
//        tt.cascade(cascadeIdx)
//        if tt[cascadeIdx].isOverflowIRQ() {
//            tt[cascadeIdx].raiseIRQ()
//        }
//    }
//}

type Timer struct {
    Gba *GBA
    Idx int
	CNT, D uint32
    SavedInitialValue uint32
    SavedCycles uint32
    Elapsed uint32
}

func (t *Timer) raiseIRQ() {
    t.Gba.triggerIRQ(0x3 + uint32(t.Idx))
}

func (t *Timer) ReadCnt(hi bool) uint8 {

    if hi {
        return uint8(t.CNT >> 8)
    }

    return uint8(t.CNT)
}

func (t *Timer) WriteCnt(v uint8, hi bool) {

    if hi { return }

    //oldValue := t.CNT & 0xFF
    t.CNT = uint32(v) & 0xC7

    //if setEnabled := utils.BitEnabled(uint32(v), 7) && !utils.BitEnabled(oldValue, 7); setEnabled {
    if setEnabled := utils.BitEnabled(uint32(v), 7); setEnabled {
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
        //t.SavedInitialValue |= (uint32(v) << 8)
        t.SavedInitialValue = (t.SavedInitialValue & 0x00FF) | (uint32(v) << 8)
        return
    }

    t.SavedInitialValue = (t.SavedInitialValue & 0xFF00) | uint32(v)
}

//func (t *Timer) Increment(cascade bool) bool {
//
//    if t.isCascade() && !cascade {
//        return false
//    }
//    if !t.isCascade() && cascade {
//        panic("NON-CASCADE TIMER INCREMENTING IN CASCADE")
//    }
//
//    overflow := t.D == 0xFFFF
//
//    t.D++
//
//    if overflow {
//        t.D = t.SavedInitialValue
//
//
//        //t.CNT = 0
//    }
//
//    return overflow
//}

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

    increment := 0
    if t.isCascade() && overflow {
        increment = 1
    }

    if !t.isCascade() {
        t.Elapsed += cycles
        freq := t.getCycles()

        if t.Elapsed >= freq {
            increment = int(t.Elapsed / freq)
            t.Elapsed = t.Elapsed % freq
        }
    }

    overflow = false

    for range increment {
        tmp := t.D + 1
        if tmp > 0xFFFF {
            t.D = 0
        } else {
            t.D = tmp
        }

        if t.D == 0 {
            t.D = t.SavedInitialValue
            overflow = true
        }
    }

    if !overflow {
        return false
    }

    apu := t.Gba.Apu

    if aTick := utils.BitEnabled(t.Gba.Mem.Read16(0x400_0082), 10); aTick {

        channel := &apu.ChannelA

        channel.Ticks = (channel.Ticks + 1) % 16
        if channel.Ticks == 0 {
            channel.Refill = true
        }
    }
    if bTick := utils.BitEnabled(t.Gba.Mem.Read16(0x400_0082), 14); bTick {

        channel := &apu.ChannelB

        channel.Ticks = (channel.Ticks + 1) % 16
        if channel.Ticks == 0 {
            channel.Refill = true
        }
    }

    if apu.ChannelA.Refill || apu.ChannelB.Refill {
        t.Gba.DmaOnRefresh = true
        apu.ChannelA.Refill = false
        apu.ChannelB.Refill = false
    }

    if t.isOverflowIRQ() {
        t.raiseIRQ()
    }

    return true
}
