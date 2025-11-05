package nds

import (
	"fmt"
	"os"
	"github.com/aabalke/guac/emu/nds/debug"
)

func Log(nds *Nds, start, end uint64, arm9 bool) {
    switch {
    case debug.CURR_INST < start:
        return
    case debug.CURR_INST < end:
        nds.LogCpu(arm9)
        return
    default:
        debug.L.Close()
        os.Exit(0)
    }
}

func (nds *Nds) LogCpu(arm9 bool) {

    var (
        r       [16]uint32
        opcode  uint32
        cpsr    uint32
        ie      uint32
        ime     bool
    )

    if arm9 {
        cpu := &nds.arm9
        r = cpu.Reg.R
        cpsr = uint32(cpu.Reg.CPSR)
        if cpu.Reg.IsThumb {
            opcode = nds.mem.Read16(r[15], true)
        } else {
            opcode = nds.mem.Read32(r[15], true)
        }
    } else {
        cpu := &nds.arm7
        r = cpu.Reg.R
        cpsr = uint32(cpu.Reg.CPSR)
        if cpu.Reg.IsThumb {
            opcode = nds.mem.Read16(r[15], false)
        } else {
            opcode = nds.mem.Read32(r[15], false)
        }
    }

    s := fmt.Sprintf("R %08X ", r)
    s += fmt.Sprintf("OP %08X ", opcode)
    s += fmt.Sprintf("CPSR %08X ", cpsr)
    s += fmt.Sprintf("IME %t ", ime)
    s += fmt.Sprintf("IE %08X ", ie)
    s += fmt.Sprintf("CURR %08X ", debug.CURR_INST)
    debug.L.Write(s)
}
