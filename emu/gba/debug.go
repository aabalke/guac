package gba

import (
	"bufio"
	"fmt"

	_ "image/png"
	"os"
)

type Debugger struct {
    Gba *GBA
    Version int
}

func (d *Debugger) print(i int) {
    reg := &d.Gba.Cpu.Reg
    p := func(a string, b uint32) { fmt.Printf("% 8s: % 9X\n", a, b)}
    s := func(a string) { fmt.Printf("%s\n", a)}

    s("--------  --------")
    fmt.Printf("inst dec %d\n", uint32(i))
    p("inst", uint32(i))

    if d.Gba.Cpu.Reg.CPSR.GetFlag(FLAG_T) {
        p("opcode", d.Gba.Mem.Read16(reg.R[15]))
    } else {
        p("opcode", d.Gba.Mem.Read32(reg.R[15]))
    }
    mode := d.Gba.Cpu.Reg.getMode()
    s("--------  --------")
    p("r00", reg.R[0])
    p("r01", reg.R[1])
    p("r02", reg.R[2])
    p("r03", reg.R[3])
    p("r04", reg.R[4])
    p("r05", reg.R[5])
    p("r06", reg.R[6])
    p("r07", reg.R[7])
    p("r08", reg.R[8])
    p("r09", reg.R[9])
    p("r10", reg.R[10])
    p("r11", reg.R[11])
    p("r12", reg.R[12])
    p("sp/r13", reg.R[13])
    p("lr/r14", reg.R[14])
    p("pc/r15", reg.R[15])
    s("--------  --------")
    p("cpsr", uint32(reg.CPSR))
    p("spsr", uint32(reg.SPSR[BANK_ID[mode]]))
    p("MODE", BANK_ID[mode])
    //p("0x3007FFC", d.Gba.Mem.Read32(0x3007FFC))
    //p("0x4000004", d.Gba.Mem.Read16(0x4000004))
    //p("0x4000208", d.Gba.Mem.Read16(0x4000208))
    //p("0x4000200", d.Gba.Mem.Read16(0x4000200))
    //p("0x4000202", d.Gba.Mem.Read16(0x4000202))

    s("--------  --------")

    for i := range len(reg.LR) {
        p(fmt.Sprintf("LR %02d", i), uint32(reg.LR[uint32(i)]))
    }

    s("--------  --------")
    //p(fmt.Sprintf("4744 %08X", i), d.Gba.Mem.Read32(0x802E7A4))
    count := 0x20
    start := 0x6003800 + count * 4
    for i := start; i >= start - (count * 4); i -= 4 {
        p(fmt.Sprintf("IO %X", i), d.Gba.Mem.Read32(uint32(i)))
    }

    s("--------  --------")

    //j := uint32(0x4000208)
    //p(fmt.Sprintf("IME %04X", j), d.Gba.Mem.Read16(uint32(j)))
    //j = uint32(0x4000204)
    //p(fmt.Sprintf("WS  %04X", j), d.gba.Mem.Read16(uint32(j)))
    //j = uint32(0x4000202)
    //p(fmt.Sprintf("IF  %04X", j), d.gba.Mem.Read16(uint32(j)))
    //j = uint32(0x4000200)
    //p(fmt.Sprintf("IE  %04X", j), d.gba.Mem.Read16(uint32(j)))

    //s("\n\n")
    //p(fmt.Sprintf("STACK %X", 0x3007E2E), d.gba.Mem.Read32(0x3007E2E))
    //for i := 0x0400_00E0; i >= 0x0400_00D0; i -= 4 {

    //start := 0x40000E0
    //count := 0x10
    //for i := start; i >= start - (count * 4); i -= 4 {
    //    p(fmt.Sprintf("IO %X", i), d.gba.Mem.Read32(uint32(i)))
    //}
    //s("------")

    //start := 0x3007EB0
    //start := 0x30014B0
    //start := 0xE000080
    //count := 0x20
    //for i := start; i >= start - (count * 4); i -= 4 {
    //    p(fmt.Sprintf("IO %X", i), d.Gba.Mem.Read32(uint32(i)))
    //}
}

func (d *Debugger) dump(s, e uint32) {

    // fix to buffer some day
    tmp := ""

    for i := s; i <= e; i += 4 {
        tmp += fmt.Sprintf("%08X", d.Gba.Mem.Read32(uint32(i)))
    }
    f, err := os.Create("./dump")
    if err != nil { panic(err) } 
    w := bufio.NewWriter(f)
    _, err = w.WriteString(tmp)

    if err != nil { panic(err) } 

    w.Flush()
}

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

    BUF_SIZE := uint64(10_000)

    if CURR_INST%BUF_SIZE == 0 {
        l.bufWriter.Flush()
    }
}

var cpuDurations [100]int64

var aveAverages []int64

//st := time.Now()
//durations[count % 10] = time.Since(st).Milliseconds()

func getProfilerTimes(frame int) bool {
    if frame % 100 == 0 {
        //fmt.Printf("\rCPU %02dms GFX %02dms SND %02dms", average(cpuDurations), average(gfxDurations), average(sndDurations))

        i := average(cpuDurations)

        aveAverages = append(aveAverages, i)
        fmt.Printf("\rCPU %02dms", i)
    }

    if frame == 1000 {
        fmt.Printf("\rCPU %02dms", averageBetter(aveAverages))
        //os.Exit(0)
    }

    return false
}

func average(xs[100]int64)int64 {
	total := int64(0)
	for _,v:=range xs {
		total +=v
	}
	return total/int64(len(xs))
}

func averageBetter(xs[]int64)int64 {
	total := int64(0)
	for _, v :=range xs {
		total +=v
	}
	return total/int64(len(xs))
}
