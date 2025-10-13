package rast

import (
	"fmt"

	"github.com/aabalke/guac/emu/nds/cpu"
	"github.com/aabalke/guac/emu/nds/rast/gl"
	"github.com/aabalke/guac/emu/nds/utils"
)

type GeoEngine struct {
    Irq *cpu.Irq
    Buffers *Buffers
    Data []uint32
    VRAM VRAM

    GxStat GXSTAT

    MtxStacks *MtxStacks
    Viewport Viewport

    PrepPoly Polygon
    ActivePoly Polygon

    Color gl.Color

    Texture Texture

    Vertex *gl.Vertex
    StoredNormal gl.Vector

    Packed bool
    PackedCmds [4]uint32
    PackedIdx uint8

    ClipMatrix gl.Matrix
    PosTestData [4]uint32

    Lights [4]gl.Light

    TextureCache TextureCache
}

func NewGeoEngine(buffers *Buffers, irq *cpu.Irq, vram VRAM) *GeoEngine {
    return &GeoEngine{
        VRAM: vram,
        Irq: irq,
        Buffers: buffers,
        MtxStacks: NewMtxStacks(),
        //Color: gl.Transparent,
        TextureCache: make(map[uint32]*[]gl.Color, 0),
        Vertex: &gl.Vertex{},
    }
}

func (g *GeoEngine) Fifo(v uint32) {

    //fmt.Printf("FIFO %08X\n", v)

    // this will be buggy, need to handle if packed cmd sets data to 0 ( add cmd???)
    if cmd := len(g.Data) == 0; cmd {

        if packed := v &^ 0xFF != 0; packed {

            g.PackedCmds = [4]uint32{}
            g.PackedIdx = 0
            g.Packed = true

            //fmt.Printf("Starting New Packed Command %08X\n", v)

            g.PackedCmds[0] = v & 0xFF
            g.PackedCmds[1] = (v >> 8)  & 0xFF
            g.PackedCmds[2] = (v >> 16) & 0xFF
            g.PackedCmds[3] = (v >> 24) & 0xFF

            v &= 0xFF
            g.Data = append(g.Data, v)
            // check if packed cmd has no params
            g.PackedFifo()
            return
        }

        g.Packed = false
        g.Data = append(g.Data, v)

        // check if packed cmd has no params
        g.Cmd(true, g.Data)
        return
    }

    if g.Packed {
        g.Data = append(g.Data, v)
        g.PackedFifo()
        return
    }

    g.Data = append(g.Data, v)
    g.Cmd(true, g.Data)
}

func (g *GeoEngine) PackedFifo() {

    g.Cmd(true, g.Data)

    if len(g.Data) != 0 {
        return
    }

    g.PackedIdx = (g.PackedIdx + 1) & 0b11

    //fmt.Printf("Updating PackedIdx %02d\n", g.PackedIdx)

    if finishedPacked := g.PackedIdx == 0; finishedPacked {
        g.Data = []uint32{}
        //fmt.Printf("Finished 4 Packed Commands\n")

        //g.Packed = false
        return
    }

    //fmt.Printf("Moving to Next Packed Command.\n")

    g.Data = append(g.Data, g.PackedCmds[g.PackedIdx])

    g.PackedFifo()
}

// packed cmds not implimented yet

func (g *GeoEngine) Cmd(fifo bool, data []uint32) {

    //fmt.Printf("DATA % X\n", data)

    if !g.ValidParamCount(fifo) {
        return
    }

    //fmt.Printf("C %05d %t PACKED %08X IDX %02d % 9X\n", cnt, fifo, g.PackedCmds, g.PackedIdx, data)
    //cnt++
    //fmt.Printf("DATA % X\n", data)


    if cmd := data[0]; cmd == 0x30 || cmd == 0x31 || cmd == 0x34 {
        g.Data = []uint32{}
        return
    }

    s := &g.MtxStacks.Stacks[g.MtxStacks.Mode]
    s1 := &g.MtxStacks.Stacks[1]
    sMode := g.MtxStacks.Mode

    switch cmd := data[0]; cmd {
    case 0x10:
        g.MtxStacks.Mode = data[1] & 0b11

    case 0x11:
        g.MtxStacks.Push()

    case 0x12:
        g.MtxStacks.Pop(data[1])

    case 0x13:
        g.MtxStacks.Store(data[1])

    case 0x14:
        g.MtxStacks.Restore(data[1])

    case 0x15:

        s.CurrMtx = gl.Identity()

        if sMode == 2 {
            s1.CurrMtx = gl.Identity()
        }

    case 0x16:

        m := gl.Matrix{
            X00: utils.ConvertToFloat(data[1], 12),
            X01: utils.ConvertToFloat(data[2], 12),
            X02: utils.ConvertToFloat(data[3], 12),
            X03: utils.ConvertToFloat(data[4], 12),
            X10: utils.ConvertToFloat(data[5], 12),
            X11: utils.ConvertToFloat(data[6], 12),
            X12: utils.ConvertToFloat(data[7], 12),
            X13: utils.ConvertToFloat(data[8], 12),
            X20: utils.ConvertToFloat(data[9], 12),
            X21: utils.ConvertToFloat(data[10], 12),
            X22: utils.ConvertToFloat(data[11], 12),
            X23: utils.ConvertToFloat(data[12], 12),
            X30: utils.ConvertToFloat(data[13], 12),
            X31: utils.ConvertToFloat(data[14], 12),
            X32: utils.ConvertToFloat(data[15], 12),
            X33: utils.ConvertToFloat(data[16], 12),
        }

        s.CurrMtx = m
        if sMode == 2 {
            s1.CurrMtx = m
        }


    case 0x17:

        m := gl.Matrix{
            X00: utils.ConvertToFloat(data[1], 12),
            X01: utils.ConvertToFloat(data[2], 12),
            X02: utils.ConvertToFloat(data[3], 12),
            X10: utils.ConvertToFloat(data[4], 12),
            X11: utils.ConvertToFloat(data[5], 12),
            X12: utils.ConvertToFloat(data[6], 12),
            X20: utils.ConvertToFloat(data[7], 12),
            X21: utils.ConvertToFloat(data[8], 12),
            X22: utils.ConvertToFloat(data[9], 12),
            X30: utils.ConvertToFloat(data[10], 12),
            X31: utils.ConvertToFloat(data[11], 12),
            X32: utils.ConvertToFloat(data[12], 12),
            X33: 1.0,
        }

        s.CurrMtx = m
        if sMode == 2 {
            s1.CurrMtx = m
        }

    case 0x18:

        m := gl.Matrix{
            X00: utils.ConvertToFloat(data[1], 12),
            X01: utils.ConvertToFloat(data[2], 12),
            X02: utils.ConvertToFloat(data[3], 12),
            X03: utils.ConvertToFloat(data[4], 12),
            X10: utils.ConvertToFloat(data[5], 12),
            X11: utils.ConvertToFloat(data[6], 12),
            X12: utils.ConvertToFloat(data[7], 12),
            X13: utils.ConvertToFloat(data[8], 12),
            X20: utils.ConvertToFloat(data[9], 12),
            X21: utils.ConvertToFloat(data[10], 12),
            X22: utils.ConvertToFloat(data[11], 12),
            X23: utils.ConvertToFloat(data[12], 12),
            X30: utils.ConvertToFloat(data[13], 12),
            X31: utils.ConvertToFloat(data[14], 12),
            X32: utils.ConvertToFloat(data[15], 12),
            X33: utils.ConvertToFloat(data[16], 12),
        }

        s.CurrMtx = m.Mul(s.CurrMtx)
        if sMode == 2 {
            s1.CurrMtx = m.Mul(s1.CurrMtx)
        }

    case 0x19:

        m := gl.Matrix{
            X00: utils.ConvertToFloat(data[1], 12),
            X01: utils.ConvertToFloat(data[2], 12),
            X02: utils.ConvertToFloat(data[3], 12),
            X10: utils.ConvertToFloat(data[4], 12),
            X11: utils.ConvertToFloat(data[5], 12),
            X12: utils.ConvertToFloat(data[6], 12),
            X20: utils.ConvertToFloat(data[7], 12),
            X21: utils.ConvertToFloat(data[8], 12),
            X22: utils.ConvertToFloat(data[9], 12),
            X30: utils.ConvertToFloat(data[10], 12),
            X31: utils.ConvertToFloat(data[11], 12),
            X32: utils.ConvertToFloat(data[12], 12),
            X33: 1.0,
        }

        s.CurrMtx = m.Mul(s.CurrMtx)
        if sMode == 2 {
            s1.CurrMtx = m.Mul(s1.CurrMtx)
        }

    case 0x1A:

        m := gl.Matrix{
            X00: utils.ConvertToFloat(data[1], 12),
            X01: utils.ConvertToFloat(data[2], 12),
            X02: utils.ConvertToFloat(data[3], 12),
            X10: utils.ConvertToFloat(data[4], 12),
            X11: utils.ConvertToFloat(data[5], 12),
            X12: utils.ConvertToFloat(data[6], 12),
            X20: utils.ConvertToFloat(data[7], 12),
            X21: utils.ConvertToFloat(data[8], 12),
            X22: utils.ConvertToFloat(data[9], 12),
            X33: 1.0,
        }

        s.CurrMtx = m.Mul(s.CurrMtx)
        if sMode == 2 {
            s1.CurrMtx = m.Mul(s1.CurrMtx)
        }

    case 0x1B:

        v := gl.Vector{
            X: utils.ConvertToFloat(data[1], 12),
            Y: utils.ConvertToFloat(data[2], 12),
            Z: utils.ConvertToFloat(data[3], 12),
        }

        // no effect on vector matrix - keeps light vector length intact
        if sMode != 2 {
            s.CurrMtx = s.CurrMtx.Scale(v)
        } else {
            s1.CurrMtx = s1.CurrMtx.Scale(v)
        }

    case 0x1C:

        v := gl.Vector{
            X: utils.ConvertToFloat(data[1], 12),
            Y: utils.ConvertToFloat(data[2], 12),
            Z: utils.ConvertToFloat(data[3], 12),
        }

        s.CurrMtx = s.CurrMtx.Translate(v)
        if sMode == 2 {
            s1.CurrMtx = s1.CurrMtx.Translate(v)
        }

    case 0x20:

        g.Color = Write15BitColor(data[1])

    case 0x21:

        convert := func(v uint16) float64 {
            v &= 0x3FF
            val := int16(v << 6) >> 6
            return float64(val) / 512
        }

        x := convert(uint16(data[1]))
        y := convert(uint16(data[1]>>10))
        z := convert(uint16(data[1]>>20))
        v := gl.Vector{X: x, Y: y, Z: z}

        directionalMtx := g.MtxStacks.Stacks[2].CurrMtx
        g.StoredNormal = directionalMtx.MulPosition(v)
        g.StoredNormal = g.StoredNormal.Normalize()

        g.StoredNormal = g.StoredNormal.MulScalar(-1)

    case 0x22:

        g.Texture.WriteCoord(data[1])

    case 0x23:
        g.Vertex = g.ActivePoly.WriteVertex(data, g, V_16)

    case 0x24:
        g.Vertex = g.ActivePoly.WriteVertex(data, g, V_10)

    case 0x25:
        g.Vertex = g.ActivePoly.WriteVertex(data, g, V_XY)

    case 0x26:
        g.Vertex = g.ActivePoly.WriteVertex(data, g, V_XZ)

    case 0x27:
        g.Vertex = g.ActivePoly.WriteVertex(data, g, V_YZ)

    case 0x28:
        g.Vertex = g.ActivePoly.WriteVertex(data, g, V_DF)

    case 0x29:

        g.PrepPoly.WriteAttrs(data[1])

    case 0x2A:

        g.Texture.WriteParam(data[1])

    case 0x2B:

        g.Texture.WritePalBase(data[1])

    case 0x32:

        x := utils.Convert10ToFloat(uint16(data[1]), 9)
        y := utils.Convert10ToFloat(uint16(data[1] >> 10), 9)
        z := utils.Convert10ToFloat(uint16(data[1] >> 20), 9)
        // not sure if need to use VectorW
        v := gl.Vector{X: x, Y: y, Z: z}
        idx := data[1] >> 30

        directionalMtx := g.MtxStacks.Stacks[2].CurrMtx
        g.Lights[idx].Direction = directionalMtx.MulPosition(v)
        g.Lights[idx].Direction = g.Lights[idx].Direction.Normalize()

        //fmt.Printf("UPDATING % .2f\n", g.Lights[idx].Direction)

    case 0x33:

        idx := data[1] >> 30
        g.Lights[idx].Color = Write15BitColor(data[1])

    case 0x40:

        if len(g.ActivePoly.Vertices) != 0 {

            //fmt.Printf("BAD ACTIVE POLYGON HAS VERTICIES WHEN SETTING NEW BEGIN. WAS END_VTXS NOT CALLED? VERTS LEN %d\n", len(g.ActivePoly.Vertices))

            g.AddPolygon()
        }

        g.ActivePoly = g.PrepPoly
        g.PrepPoly = Polygon{}

        g.ActivePoly.PrimitiveType = uint8(data[1] & 0b11)

    case 0x41:

        g.AddPolygon()

    case 0x50:

        if g.Buffers.BisRendering {
            g.Buffers.B = []Polygon{}
        } else {
            g.Buffers.A = []Polygon{}
        }

        g.Buffers.BisRendering = !g.Buffers.BisRendering

    case 0x60: 

        g.Viewport.X1 = uint8(data[1])
        g.Viewport.Y1 = uint8(data[1] >> 8)
        g.Viewport.X2 = uint8(data[1] >> 16)
        g.Viewport.Y2 = uint8(data[1] >> 24)

    case 0x70:
        g.BoxTest(data, &g.ClipMatrix)

    case 0x71:

        g.PosTestData = g.PosTest(data, &g.ClipMatrix)

    case 0x0:
        //fmt.Printf("UNSETUP GX CMD %02X\n", cmd)

    default:
        //panic(fmt.Sprintf("UNSETUP GX CMD %02X\n", cmd))
        fmt.Printf("UNSETUP GX CMD %02X\n", cmd)

    }

    g.Data = []uint32{}

    g.UpdateClipMtx()

    //fmt.Printf("STATUS %v\n", g.MtxStacks.ClipMatrix)
}

func (g *GeoEngine) UpdateClipMtx() {
    pos := &g.MtxStacks.Stacks[1].CurrMtx
    per := &g.MtxStacks.Stacks[0].CurrMtx

    g.ClipMatrix = pos.Mul(*per)
}

var paramCnt = map[uint32]int{
    0x00: 1,
    0x10: 1,
    0x11: 1,
    0x12: 1,
    0x13: 1,
    0x14: 1,
    0x15: 1,
    0x16: 16,
    0x17: 12,
    0x18: 16,
    0x19: 12,
    0x1A: 9,
    0x1B: 3,
    0x1C: 3,
    0x20: 1,
    0x21: 1,
    0x22: 1,
    0x23: 2,
    0x24: 1,
    0x25: 1,
    0x26: 1,
    0x27: 1,
    0x28: 1,
    0x29: 1,
    0x2A: 1,
    0x2B: 1,
    0x30: 1,
    0x31: 1,
    0x32: 1,
    0x33: 1,
    0x34: 32,
    0x40: 1,
    0x41: 1,
    0x50: 1,
    0x60: 1,
    0x70: 3,
    0x71: 2,
    0x72: 1,
}

func (g *GeoEngine) ValidParamCount(fifo bool) bool {

    cmd := g.Data[0]
    params := len(g.Data) - 1

    // when using fifo, sometimes no param provided, but when using io
    // dummy param is provided.
    if fifo && params == 0 {
        if cmd == 0x00 || cmd == 0x11 || cmd == 0x15 || cmd == 0x41 {
            return true
        }
    }

    cnt, ok := paramCnt[cmd]
    if !ok {
        panic(fmt.Sprintf("UNKNOWN CMD GXFIFO % 2X", g.Data))
    }

    return cnt == params
}

func (g *GeoEngine) AddPolygon() {


    //if shadow := g.ActivePoly.Mode == 3; shadow {
    //    //g.ActivePoly.Vertices = []gl.Vertex{}
    //    //return
    //}

    //g.ActivePoly.Texture = g.Texture
    g.Buffers.Append(g.ActivePoly)
    g.ActivePoly.Vertices = []gl.Vertex{}

}

func Write15BitColor(v uint32) gl.Color {
    return gl.MakeColorFrom15Bit(
	uint8((v) & 0b11111),
	uint8((v >> 5) & 0b11111),
	uint8((v >> 10) & 0b11111),
    )
}
