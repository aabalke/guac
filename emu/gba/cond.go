package gba

//const (
//    EQ = iota
//    NE
//    CS
//    CC
//    MI
//    PL
//    VS
//    VC
//    HI
//    LS
//    GE
//    LT
//    GT
//    LE
//    AL
//    NV
//) 

var masks = [16]uint16{
    0:  0xF0F0,
    1:  0x0F0F,
    2:  0xCCCC,
    3:  0x3333,
    4:  0xFF00,
    5:  0x00FF,
    6:  0xAAAA,
    7:  0x5555,
    8:  0x0C0C,
    9:  0xF3F3,
    10: 0xAA55,
    11: 0x55AA,
    12: 0x0A05,
    13: 0xF5FA,
    14: 0xFFFF,
    15: 0x0000,
}

func (cpu *Cpu) CheckCond(cond uint32) bool {
    return (masks[cond] & (1 << (cpu.Reg.CPSR >> 28))) != 0
}
