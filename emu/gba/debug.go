package gba

import "fmt"

type Debugger struct {
    gba *GBA
}




func (d *Debugger) print(i int) {
    reg := d.gba.Cpu.Reg
    p := func(a string, b uint32) { fmt.Printf("% 8s: %08X\n", a, b)}
    s := func(a string) { fmt.Printf("%s\n", a)}

    s("--------  --------")
    p("inst", uint32(i))
    p("opcode", d.gba.Mem.Read32(reg.R[15]))
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
    p("spsr", uint32(reg.SPSR))
    //p("mem5", uint32(d.gba.Mem.Read32(0x500_0000)))

    //for i := 0x300_7E90; i <= 0x300_7EC0; i += 4 {
    //    p(fmt.Sprintf("mem4 %X", i), d.gba.Mem.Read32(uint32(i)))
    //}
}
