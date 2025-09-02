package nds

import (
	"bufio"
	"fmt"
	"os"
    "os/exec"
)

const(
    BUF_SIZE = 0xFFF
)

type Logger struct {
	nds  *Nds
	file *os.File
    bufWriter *bufio.Writer
    path string
}

func NewLogger(path string, nds *Nds) *Logger {

    l := Logger{}

    f, err := os.Create(path)
    if err != nil {
        panic(err)
    }

    l.path = path
    l.file = f
    l.bufWriter = bufio.NewWriter(f)
    l.nds = nds

    return &l
}

func (l *Logger) Update(start, end, curr uint64, arm9 bool) {

    if CURR_INST < start {
        return
    }

    if CURR_INST > end {
        return
    }

    logger.Write(CURR_INST, arm9)

    if CURR_INST == end {
        logger.Close()
        os.Exit(0)
    }
}

func (l *Logger) Close() {

    l.bufWriter.Flush()
    l.file.Close()

    fmt.Printf("Closing Logger\n")

    var cmd *exec.Cmd
    cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", l.path)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    err := cmd.Start()
    if err != nil {
        panic(err)
    }
}

func (l *Logger) Write(i uint64, arm9 bool) {

    if !arm9 && i & 1 == 1 {
        return
    }

    nds := l.nds

    var s string

    if arm9 {
        cpu := nds.arm9
        r := &cpu.Reg.R
        cpsr := cpu.Reg.CPSR
        opcode := nds.mem.Read32(r[15], true)
        op := fmt.Sprintf("%08X", opcode)
        if cpu.Reg.IsThumb {
            op = fmt.Sprintf("%04X", opcode)
        }
        s = fmt.Sprintf("%08X,%s,R0 %08X R1 %08X R2 %08X R3 %08X R4 %08X R5 %08X R6 %08X R7 %08X R8 %08X R9 %08X R10 %08X R11 %08X R12 %08X SP %08X LR %08X CPSR %08X CURR %d",
        r[15], op, r[0], r[1], r[2], r[3], r[4], r[5], r[6], r[7], r[8], r[9], r[10], r[11], r[12], r[13], r[14], cpsr, CURR_INST)
    } else {
        cpu := nds.arm7
        r := &cpu.Reg.R
        cpsr := cpu.Reg.CPSR
        opcode := nds.mem.Read32(r[15], false)
        op := fmt.Sprintf("%08X", opcode)
        if cpu.Reg.IsThumb {
            op = fmt.Sprintf("%04X", opcode)
        }
        s = fmt.Sprintf("%08X,%s,R0 %08X R1 %08X R2 %08X R3 %08X R4 %08X R5 %08X R6 %08X R7 %08X R8 %08X R9 %08X R10 %08X R11 %08X R12 %08X SP %08X LR %08X CPSR %08X CURR %d",
        r[15], op, r[0], r[1], r[2], r[3], r[4], r[5], r[6], r[7], r[8], r[9], r[10], r[11], r[12], r[13], r[14], cpsr, CURR_INST)
    }

    fmt.Fprintf(l.bufWriter, "%s\n", s)

    if i & BUF_SIZE == 0 {
        l.bufWriter.Flush()
    }
}
