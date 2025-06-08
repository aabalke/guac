package gba

import (
    "fmt"
	"github.com/aabalke33/guac/emu/gba/utils"
)

type Timers [4]Timer

func (tt *Timers) Increment(newCycles uint32) {

    for i := range tt {

        t := &tt[i]

        if !t.isEnabled() || t.isCascade() {
            continue
        }
        t.SavedCycles += newCycles

        //fmt.Printf("Incrementing Timer %d, count=%04X\n", t.Idx, t.D)

        for range t.SavedCycles / t.getCycles() {
            overflow := t.Increment(false)
            if overflow {
                    //fmt.Printf("Timer %d overflow: D=0x%04X, reload=0x%04X\n", t.Idx, t.D, t.SavedInitialValue)

                if i < 3 && tt[i+1].isCascade() {
                    tt.cascade(i)
                }

                if t.isOverflowIRQ() {
                    t.raiseIRQ()
                }
            }
        }

        t.SavedCycles %= t.getCycles()
    }
}

func (tt *Timers) cascade(overflowTimerIdx int) {

    if notPossible := overflowTimerIdx == 4; notPossible {
        return
    }

    cascadeIdx := overflowTimerIdx + 1

    if !tt[cascadeIdx].isEnabled() {
        return
    }

    overflow := tt[cascadeIdx].Increment(true)

    if overflow {
        tt.cascade(cascadeIdx)
        if tt[cascadeIdx].isOverflowIRQ() {
            tt[cascadeIdx].raiseIRQ()
        }
    }
}

type Timer struct {
    Gba *GBA
    Idx int
	CNT, D uint32
    SavedInitialValue uint32
    SavedCycles uint32
}

func (t *Timer) raiseIRQ() {
    cause := fmt.Sprintf("TIMER%d", t.Idx)
    t.Gba.triggerIRQ(0x3 + uint32(t.Idx), cause)
}

func (t *Timer) ReadCnt(hi bool) uint8 {
    if hi {
        return 0x0
    }

    return uint8(t.CNT) & 0xFF
}

func (t *Timer) WriteCnt(v uint8, hi bool) {

    if hi { return }

    if setEnabled := utils.BitEnabled(uint32(v), 7); setEnabled {
        t.D = t.SavedInitialValue
        t.SavedInitialValue = 0
    }

    t.CNT = uint32(v)
}

func (t *Timer) ReadD(hi bool) uint8 {

    if hi {
        return uint8(t.D >> 8) & 0xFF
    }

    return uint8(t.D) & 0xFF
}

func (t *Timer) WriteD(v uint8, hi bool) {

    if hi {
        //t.SavedInitialValue |= (uint32(v) << 8)
        t.SavedInitialValue = (t.SavedInitialValue & 0x00FF) | (uint32(v) << 8)
        return
    }

    t.SavedInitialValue = (t.SavedInitialValue & 0xFF00) | uint32(v)
}

func (t *Timer) Increment(cascade bool) bool {

    if t.isCascade() && !cascade {
        return false
    }
    if !t.isCascade() && cascade {
        panic("NON-CASCADE TImER INCREMENTING IN CASCADE")
    }

    overflow := t.D == 0xFFFF

    t.D++

    if overflow {
        t.D = t.SavedInitialValue
        //t.D = 0
    }

    return overflow
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
