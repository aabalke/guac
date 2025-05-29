package gba

import (
    "os"
    "bufio"
    "fmt"
)

type Logger struct {
	Instruction    int
	MaxInstruction int

    gba *GBA
    file *os.File
    bufWriter *bufio.Writer
}

func NewLogger(path string, gba *GBA) *Logger {

    l := Logger{}
    f, err := os.Create(path)
    if err != nil { panic(err) }

    l.file = f
    l.bufWriter = bufio.NewWriter(f)
    l.gba = gba

    return &l
}

func (l *Logger) Close() {
    l.bufWriter.Flush()
    l.file.Close()
}

func (l *Logger) WriteLog() {

    gba := l.gba

    s := fmt.Sprintf(
        "CURR %08X INST %08X MODE %08X CPSR %08X SPSR %08X R00 %08X R01 %08X R02 %08X R03 %08X R04 %08X R05 %08X R06 %08X R07 %08X R08 %08X R09 %08X R10 %08X R11 %08X R12 %08X R13 %08X R14 %08X R15 %08X R14B0 %08X IME %08X, IE %08X, IF %08X",
        CURR_INST, gba.Mem.Read32(gba.Cpu.Reg.R[15]), gba.Cpu.Reg.getMode(), gba.Cpu.Reg.CPSR, gba.Cpu.Reg.SPSR[BANK_ID[gba.Cpu.Reg.getMode()]],
        gba.Cpu.Reg.R[0],
        gba.Cpu.Reg.R[1],
        gba.Cpu.Reg.R[2],
        gba.Cpu.Reg.R[3],
        gba.Cpu.Reg.R[4],
        gba.Cpu.Reg.R[5],
        gba.Cpu.Reg.R[6],
        gba.Cpu.Reg.R[7],
        gba.Cpu.Reg.R[8],
        gba.Cpu.Reg.R[9],
        gba.Cpu.Reg.R[10],
        gba.Cpu.Reg.R[11],
        gba.Cpu.Reg.R[12],
        gba.Cpu.Reg.R[13],
        gba.Cpu.Reg.R[14],
        gba.Cpu.Reg.R[15],
        gba.Mem.Read32(0x400_0208),
        gba.Mem.Read32(0x400_0200),
        gba.Mem.Read32(0x400_0202),
        gba.Cpu.Reg.LR[0],
    )

    fmt.Fprintf(l.bufWriter, "%s\n", s)

    BUF_SIZE := 10_000

    if CURR_INST%BUF_SIZE == 0 {
        l.bufWriter.Flush()
    }
}

