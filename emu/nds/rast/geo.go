package rast

import (
	"fmt"
	"image/color"

	"github.com/aabalke/guac/emu/nds/rast/gl"
	"github.com/aabalke/guac/emu/nds/utils"
)

type GeoEngine struct {
    Buffers *Buffers
    Data []uint32
    NextData uint32
    NextDataIdx uint8

    GxStat GXSTAT

    MtxStacks *MtxStacks
    Viewport Viewport

    PrepPoly Polygon
    ActivePoly Polygon

    Color gl.Color

    Texture Texture

    Vertex *gl.Vertex

    Packed bool
    PackedCmds [4]uint32
    PackedIdx uint8


    ClipMatrix gl.Matrix

}

func NewGeoEngine(buffers *Buffers) *GeoEngine {
    return &GeoEngine{
        Buffers: buffers,
        MtxStacks: NewMtxStacks(),
    }
}

func (g *GeoEngine) Fifo(v uint32) {

    // this will be buggy, need to handle if packed cmd sets data to 0 ( add cmd???)
    if cmd := len(g.Data) == 0; cmd {

        if packed := v &^ 0xFF != 0; packed {

            g.Packed = true

            g.PackedCmds[0] = v & 0xFF
            g.PackedCmds[1] = (v >> 8)  & 0xFF
            g.PackedCmds[2] = (v >> 16) & 0xFF
            g.PackedCmds[3] = (v >> 24) & 0xFF

            v &= 0xFF
            g.Data = append(g.Data, v)
            g.NextDataIdx = 0
            return
        }

        g.Packed = false

        g.Data = append(g.Data, v)
        g.NextDataIdx = 0
        return
    }

    if g.Packed {
        g.PackedFifo(v)
        return
    }

    g.NextData &^= 0xFF << (8 * g.NextDataIdx)
    g.NextData |= (v & 0xFF) << (8 * g.NextDataIdx)

    g.NextDataIdx = (g.NextDataIdx + 1) & 0b11

    if fullParam := g.NextDataIdx == 0; fullParam {
        g.Data = append(g.Data, g.NextData)
        g.Cmd(g.Data)
    }
}

func (g *GeoEngine) PackedFifo(v uint32) {

    g.Data = append(g.Data, v)
    g.Cmd(g.Data)

    if len(g.Data) == 0 {
        g.PackedIdx = (g.PackedIdx + 1) & 0b11
        g.Data = append(g.Data, g.PackedCmds[g.PackedIdx])
    }
}

// packed cmds not implimented yet

func (g *GeoEngine) Cmd(data []uint32) {

    if !g.ValidParamCount() {
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

        g.UpdateClipMtx()

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

        g.UpdateClipMtx()

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

        g.UpdateClipMtx()

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

        g.UpdateClipMtx()

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

        g.UpdateClipMtx()

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

        g.UpdateClipMtx()

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

        g.UpdateClipMtx()

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

        g.UpdateClipMtx()

    case 0x20:

        g.WriteColor(data[1])

    case 0x22:

        g.Texture.WriteCoord(data[1])

    case 0x23:

        g.Vertex = g.ActivePoly.WriteVtx16(
            data,
            &g.ClipMatrix,
            g.Color,
            g.Texture.S,
            g.Texture.T)

    case 0x24:

        g.Vertex = g.ActivePoly.WriteVtx10(
            data,
            &g.ClipMatrix,
            g.Color,
            g.Texture.S,
            g.Texture.T)

    case 0x25:

        g.Vertex = g.ActivePoly.WriteVtxRelative(
            data,
            &g.ClipMatrix,
            g.Color,
            g.Texture.S,
            g.Texture.T,
            g.Vertex,
            REL_XY,
        )

    case 0x26:

        g.Vertex = g.ActivePoly.WriteVtxRelative(
            data,
            &g.ClipMatrix,
            g.Color,
            g.Texture.S,
            g.Texture.T,
            g.Vertex,
            REL_XZ,
        )

    case 0x27:

        g.Vertex = g.ActivePoly.WriteVtxRelative(
            data,
            &g.ClipMatrix,
            g.Color,
            g.Texture.S,
            g.Texture.T,
            g.Vertex,
            REL_YZ,
        )

    case 0x29:

        g.PrepPoly.WriteAttrs(data[1])

    case 0x2A:

        g.Texture.WriteParam(data[1])

    case 0x2B:

        g.Texture.WritePalBase(data[1])

    case 0x40:

        if len(g.ActivePoly.Vertices) != 0 {
            fmt.Printf("BAD ACTIVE POLYGON HAS VERTICIES WHEN SETTING NEW BEGIN. WAS END_VTXS NOT CALLED? VERTS LEN %d\n", len(g.ActivePoly.Vertices))
            g.ActivePoly.Texture = g.Texture
            g.Buffers.Append(g.ActivePoly)
            g.ActivePoly.Vertices = []gl.Vertex{}
        }

        g.ActivePoly = g.PrepPoly
        g.PrepPoly = Polygon{}

        g.ActivePoly.PrimitiveType = uint8(data[1] & 0b11)

    case 0x41:

        g.ActivePoly.Texture = g.Texture
        g.Buffers.Append(g.ActivePoly)
        g.ActivePoly.Vertices = []gl.Vertex{}

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


    default:
        //panic(fmt.Sprintf("UNSETUP GX CMD %02X\n", cmd))
        fmt.Printf("UNSETUP GX CMD %02X\n", cmd)
    }

    g.Data = []uint32{}

    //fmt.Printf("STATUS %v\n", g.MtxStacks.ClipMatrix)
}

func (g *GeoEngine) UpdateClipMtx() {
    pos := g.MtxStacks.Stacks[1].CurrMtx
    per := g.MtxStacks.Stacks[0].CurrMtx
    g.ClipMatrix = pos.Mul(per)
}

func (g *GeoEngine) ValidParamCount() bool {

    cmd := g.Data[0]
    params := len(g.Data) - 1

    switch cmd {
    case 0x00: return params == 1 
    case 0x10: return params == 1 
    case 0x11: return params == 1 
    case 0x12: return params == 1 
    case 0x13: return params == 1 
    case 0x14: return params == 1 
    case 0x15: return params == 1 
    case 0x16: return params == 16
    case 0x17: return params == 12
    case 0x18: return params == 16
    case 0x19: return params == 12
    case 0x1A: return params == 9 
    case 0x1B: return params == 3 
    case 0x1C: return params == 3 
    case 0x20: return params == 1 
    case 0x21: return params == 1 
    case 0x22: return params == 1 
    case 0x23: return params == 2 
    case 0x24: return params == 1 
    case 0x25: return params == 1 
    case 0x26: return params == 1 
    case 0x27: return params == 1 
    case 0x28: return params == 1 
    case 0x29: return params == 1 
    case 0x2A: return params == 1 
    case 0x2B: return params == 1 
    case 0x30: return params == 1 
    case 0x31: return params == 1 
    case 0x32: return params == 1 
    case 0x33: return params == 1 
    case 0x34: return params == 32
    case 0x40: return params == 1 
    case 0x41: return params == 1 
    case 0x50: return params == 1 
    case 0x60: return params == 1 
    case 0x70: return params == 3 
    case 0x71: return params == 2 
    case 0x72: return params == 1 
    }

    panic(fmt.Sprintf("UNKNOWN CMD GXFIFO % 2X", g.Data))
}

func (geo *GeoEngine) WriteColor(v uint32) {

	r := uint8((v) & 0b11111)
	g := uint8((v >> 5) & 0b11111)
	b := uint8((v >> 10) & 0b11111)

	r = (r << 3) | (r >> 2)
	g = (g << 3) | (g >> 2)
	b = (b << 3) | (b >> 2)

    c := color.RGBA{
        R: r,
        G: g,
        B: b,
        A: 0xFF,
    }

    geo.Color = gl.MakeColor(c)
}
