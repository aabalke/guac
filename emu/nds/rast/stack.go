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
}

func NewMtxStacks() *MtxStacks {

    s := &MtxStacks{}

    psp := uint8(0)
    csp := uint8(0)
    tsp := uint8(0)

    s.Stacks[0].Init(1, 1, &psp)
    s.Stacks[1].Init(31, 63, &csp)
    s.Stacks[2].Init(31, 63, &csp)
    s.Stacks[3].Init(1, 1, &tsp)

    return s
}

func (m *MtxStacks) Push() {


    s := &m.Stacks[m.Mode]
    s1 := &m.Stacks[1]

    idx := int(*s.Pointer) % len(s.Mtxs)

    s.Mtxs[idx] = s.CurrMtx
    if m.Mode == 2 {
        s1.Mtxs[idx] = s1.CurrMtx
    }

    (*s.Pointer) = (*s.Pointer + 1) & s.MirrorMask

    //fmt.Printf("TOP %v %v\n", s.Mtxs[0], s1.Mtxs[0])
}

func (m *MtxStacks) Pop(param uint32) {

    s := &m.Stacks[m.Mode]
    s1 := &m.Stacks[1]

    offset := int(param & 0b1_1111)

    if neg := param & 0b10_0000 != 0; neg {
        offset = offset * -1
    }

    if m.Mode != 0 {
        (*s.Pointer) = uint8(int(*s.Pointer) - offset) & s.MirrorMask
    }

    idx := int(*s.Pointer) % len(s.Mtxs)
    //fmt.Printf("POP Mode %d %02d\n", m.Mode, idx)

    s.CurrMtx = s.Mtxs[idx]

    if m.Mode == 2 {
        s1.CurrMtx = s1.Mtxs[idx]
    }
}

func (m *MtxStacks) Store(param uint32) {

    s := &m.Stacks[m.Mode]
    s1 := &m.Stacks[1]

    idx := param & 0b1_1111

    if m.Mode == 0 {
        idx = 0
    }

    if idx == 31 {
        panic("NEED TO SETUP STACK OVERFLOW GX")
    }

    s.Mtxs[idx] = s.CurrMtx

    if m.Mode == 2 {
        s1.Mtxs[idx] = s1.CurrMtx
    }
}

func (m *MtxStacks) Restore(param uint32) {

    s := &m.Stacks[m.Mode]
    s1 := &m.Stacks[1]

    idx := param & 0b1_1111

    if m.Mode == 0 {
        idx = 0
    }

    if idx == 31 {
        panic("NEED TO SETUP STACK OVERFLOW GX")
    }

    s.CurrMtx = s.Mtxs[idx]

    if m.Mode == 2 {
        s1.CurrMtx = s1.Mtxs[idx]
    }
}

type MtxStack struct {
    CurrMtx gl.Matrix
	Mtxs []gl.Matrix
    Pointer *uint8
    MirrorMask uint8
}

func (m *MtxStack) Init(size, pointerMask uint8, pointer *uint8) {
    m.Mtxs = make([]gl.Matrix, size)
    m.MirrorMask = pointerMask
    m.Pointer = pointer
}

func (m *MtxStack) Store(mtx gl.Matrix, param uint32) {
    panic("NEED TO FIX STORE")
    // need mtx 0 to not use param
    idx := int(param & 0b1_1111)
    if idx == 31{
        panic("NEED TO SETUP STACK OVERFLOW GX")
    }
    m.Mtxs[idx] = mtx
}

func (m *MtxStack) Restore(param uint32) gl.Matrix {
    panic("NEED TO FIX RESTORE")
    // need mtx 0 to not use param
    idx := int(param & 0b1_1111)
    if idx == 31{
        panic("NEED TO SETUP STACK OVERFLOW GX")
    }
    return m.Mtxs[idx]
}

func (m *MtxStack) Curr() gl.Matrix {
    idx := int(*m.Pointer) % len(m.Mtxs)
    return m.Mtxs[idx]
}

