package cpu

import "github.com/aabalke33/guac/emu/gba"

type Cpu struct {
	GBA *gba.GBA
    Reg Reg
    Loc uint32
}

const (
    SP = 13
    LR = 14
    PC = 15
)

type Reg struct {
	Cpu  *Cpu
    R    [16]uint32
	CPSR Cond
	SPSR Cond
}

const (
    FLAG_N = 31
    FLAG_Z = 30
    FLAG_C = 29
    FLAG_V = 28
    FLAG_Q = 27
    //FLAG_J = 24
    //FLAG_E = 9
    //FLAG_A = 8
    FLAG_I = 7
    FLAG_F = 6
    FLAG_T = 5
)

type Cond uint32

func (c *Cond) GetFlag(flag uint32) bool {

    switch flag {
    case FLAG_N, FLAG_Z, FLAG_C, FLAG_V, FLAG_Q, FLAG_I, FLAG_F, FLAG_T:
        return (uint32(*c) >> flag) & 0b1 == 0b1
    }

    panic("Unknown Cond Flag Get")
}

func (c *Cond) SetFlag(flag uint32, value bool) {
    switch flag {
    case FLAG_N, FLAG_Z, FLAG_C, FLAG_V, FLAG_Q, FLAG_I, FLAG_F, FLAG_T:

        if value {
            *c |= (0b1 << flag)
            return
        }

        *c &^= (0b1 << flag)

        return
    }

    panic("Unknown Cond Flag Set")
}
