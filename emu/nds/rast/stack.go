package rast

import (

	"github.com/aabalke/guac/emu/nds/rast/gl"
)

const (
    ProjectionStack = 0
    CoordinateStack = 1
    DirectionalStack = 2
    TextureStack = 3
)

type MtxStacks struct {
    Mode uint32
    Stacks [4]MtxStack
    Overflow bool
}

type MtxStack struct {
    CurrMtx gl.Matrix
	Mtxs []gl.Matrix
    Pointer *int
    MirrorMask int
    Size int
}

func (m *MtxStack) Init(size, pointerMask int, pointer *int) {
    m.Mtxs = make([]gl.Matrix, size)
    m.MirrorMask = pointerMask
    m.Pointer = pointer
    m.Size = size
}

func NewMtxStacks() *MtxStacks {

    s := &MtxStacks{}

    // csp is shared
    psp, csp, tsp := 0, 0, 0
    s.Stacks[0].Init(1, 0, &psp)
    s.Stacks[1].Init(31, 63, &csp)
    s.Stacks[2].Init(31, 63, &csp)
    s.Stacks[3].Init(1, 0, &tsp)

    return s
}

func (m *MtxStacks) Push() {

    s := &m.Stacks[m.Mode]
    idx := int(*s.Pointer) % len(s.Mtxs)

    switch m.Mode {
    case 1, 2:
        s1 := &m.Stacks[1]
        s2 := &m.Stacks[2]
        s1.Mtxs[idx] = s1.CurrMtx
        s2.Mtxs[idx] = s2.CurrMtx
    default:
        s.Mtxs[idx] = s.CurrMtx
    }

    (*s.Pointer) = (*s.Pointer + 1) & s.MirrorMask
}

func (m *MtxStacks) Pop(param uint32) {

    switch m.Mode {
    case 1, 2:

        s := &m.Stacks[m.Mode]

        offset := int8((param&0x3F)<<2) >> 2
        *s.Pointer -= int(offset)

        if *s.Pointer >= s.Size {
            m.Overflow = true
        }

        if *s.Pointer < 0 {
            *s.Pointer += len(s.Mtxs)
        }

        idx := int(*s.Pointer) % len(s.Mtxs)

        s1 := &m.Stacks[1]
        s2 := &m.Stacks[2]
        s1.CurrMtx = s1.Mtxs[idx]
        s2.CurrMtx = s2.Mtxs[idx]

    default:
        s := &m.Stacks[m.Mode]

        *s.Pointer -= 1
        *s.Pointer &= 1

        if *s.Pointer >= s.Size {
            m.Overflow = true
        }

        s.CurrMtx = s.Mtxs[0]
    }
}

func (m *MtxStacks) Store(param uint32) {

    switch m.Mode {
    case 1, 2:
        idx := int(param & 0x1F)

        s1 := &m.Stacks[1]
        s2 := &m.Stacks[2]

        if idx >= s1.Size {
            m.Overflow = true
        }

        s1.Mtxs[idx] = s1.CurrMtx
        s2.Mtxs[idx] = s2.CurrMtx
    default:
        s := &m.Stacks[m.Mode]
        s.Mtxs[0] = s.CurrMtx
    }
}

func (m *MtxStacks) Restore(param uint32) {

    switch m.Mode {
    case 1, 2:
        idx := int(param & 0x1F)

        s1 := &m.Stacks[1]
        s2 := &m.Stacks[2]

        if idx >= s1.Size {
            m.Overflow = true
        }

        s1.CurrMtx = s1.Mtxs[idx]
        s2.CurrMtx = s2.Mtxs[idx]

    default:
        s := &m.Stacks[m.Mode]
        s.CurrMtx = s.Mtxs[0]
    }
}
