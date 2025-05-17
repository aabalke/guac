package gba

import (
	"fmt"
	"github.com/aabalke33/guac/emu/gba/utils"
)

type Timers [4]Timer

func (tt *Timers) Increment(newCycles uint32) {

    //println(2, newCycles, tt[2].getCycles(), tt[2].SavedCycles, tt[2].SavedCycles / tt[2].getCycles())

    for i := range tt {

        t := &tt[i]


        if !t.isEnabled() || t.isCascade() {
            continue
        }
        t.SavedCycles += newCycles

        for range t.SavedCycles / t.getCycles() {
            overflow := t.Increment()
            if overflow {

                if tt[i+1].isCascade() {
                    tt.cascade(i)
                }

                if t.isOverflowIRQ() {
                    tt.raiseIRQ()
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

    overflow := tt[cascadeIdx].CascadeIncrement()

    if overflow {
        tt.cascade(cascadeIdx)
        if tt[cascadeIdx].isOverflowIRQ() {
            tt.raiseIRQ()
        }
    }
}

func (tt *Timers) raiseIRQ() {

    fmt.Printf("Timer Interrupt Raised\n")
}

type Timer struct {
	CNT, D uint32
    SavedInitialValue uint32
    SavedCycles uint32
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
        t.SavedInitialValue |= (uint32(v) << 8)
        return
    }

    t.SavedInitialValue |= uint32(v)
}

func (t *Timer) CascadeIncrement() bool {

    if !t.isCascade() {
        panic("NON-CASCADE TImER INCREMENTING IN CASCADE")
    }

    cnt := t.D + 1

    if overflow := cnt > 0xFFFF; overflow {
        t.D = 0
        return true
    }

    t.D = cnt

    return false
}

func (t *Timer) Increment() bool {

    if t.isCascade() {
        return false
    }

    t.D++

    if overflow := t.D > 0xFFFF; overflow {
        t.D = 0
    }

    return t.D == 0
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
