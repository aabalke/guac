package rast

import (
	"fmt"
	"math"

	"github.com/aabalke/guac/emu/nds/cpu"
	"github.com/aabalke/guac/emu/nds/rast/gl"
	"github.com/aabalke/guac/emu/nds/utils"
)

type GeoEngine struct {
    Irq *cpu.Irq
    Buffers *Buffers
    Data []uint32

    GxStat GXSTAT

    MtxStacks *MtxStacks
    Viewport Viewport

    PrepPoly Polygon
    ActivePoly Polygon

    Color gl.Color
    Texture Texture
    LightData gl.LightData
    Vertex *gl.Vertex

    Packed bool
    PackedCmds [4]uint32
    PackedIdx uint8

    ClipMatrix gl.Matrix
    PosTestData [4]uint32
    VecTestData [3]uint16

    TextureCache TextureCache
    Vram VRAM
}

func NewGeoEngine(buffers *Buffers, irq *cpu.Irq, vram VRAM) *GeoEngine {
    return &GeoEngine{
        Vram: vram,
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
        if sMode == 2 {
            s1.CurrMtx = s1.CurrMtx.Scale(v)
        } else {
            s.CurrMtx = s.CurrMtx.Scale(v)
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

        
        //IF TexCoordTransformMode=2 THEN TexCoord=NormalVector*Matrix (see TexCoord)
        //NormalVector=NormalVector*DirectionalMatrix
        //VertexColor = EmissionColor
        //FOR i=0 to 3
        //IF PolygonAttrLight[i]=enabled THEN
        //DiffuseLevel = max(0,-(LightVector[i]*NormalVector))
        //ShininessLevel = max(0,(-HalfVector[i])*(NormalVector))^2
        //IF TableEnabled THEN ShininessLevel = ShininessTable[ShininessLevel]
        //;note: below processed separately for the R,G,B color components...
        //VertexColor = VertexColor + SpecularColor*LightColor[i]*ShininessLevel
        //VertexColor = VertexColor + DiffuseColor*LightColor[i]*DiffuseLevel
        //VertexColor = VertexColor + AmbientColor*LightColor[i]
        //ENDIF
        //NEXT i

        x := utils.Convert10ToFloat(uint16(data[1]), 9)
        y := utils.Convert10ToFloat(uint16(data[1]>>10), 9)
        z := utils.Convert10ToFloat(uint16(data[1]>>20), 9)

        // I do not believe normal vector for lighting needs scaling

        if tex := &g.Texture; tex.TransformationMode == 2 {

            // divide 16 fixes fixed point scaling (normal .9, mtx .12)

            vtx := gl.VectorW{
                X: x/16,
                Y: y/16,
                Z: z/16,
            }

            mtx := &g.MtxStacks.Stacks[3].CurrMtx
            tex.S = tex.Sv + vtx.Dot3(mtx.Col(0))
            tex.T = tex.Tv + vtx.Dot3(mtx.Col(1))
        }

        directionalMtx := g.MtxStacks.Stacks[2].CurrMtx
        v := gl.Vector{
            X: x,
            Y: y,
            Z: z,
        }

        n := &g.LightData.Normal
        *n = directionalMtx.MulPosition(v)
        //*n = n.Normalize()
        //*n = n.MulScalar(-1)

        g.Color = g.LightData.EmissionColor

        for i, v := range g.LightData.Lights {

            if !g.ActivePoly.LightsEnabled[i] {
                continue
            }

            // stuff
            diffuseLevel := max(0, -(v.Vector.Dot(*n)))
            shininessLevel := math.Pow(max(0, -(v.HalfVector.Dot(*n))), 2)

            if g.LightData.UseSpecularTbl {
                shininessLevel = g.LightData.ShininessTbl[uint32(shininessLevel)]
            }

            g.Color.R += g.LightData.SpecularColor.R * v.Color.R * shininessLevel
            g.Color.R += g.LightData.DiffuseColor.R * v.Color.R * diffuseLevel
            g.Color.R += g.LightData.AmbientColor.R * v.Color.R

            g.Color.G += g.LightData.SpecularColor.G * v.Color.G * shininessLevel
            g.Color.G += g.LightData.DiffuseColor.G * v.Color.G * diffuseLevel
            g.Color.G += g.LightData.AmbientColor.G * v.Color.G

            g.Color.B += g.LightData.SpecularColor.B * v.Color.B * shininessLevel
            g.Color.B += g.LightData.DiffuseColor.B * v.Color.B * diffuseLevel
            g.Color.B += g.LightData.AmbientColor.B * v.Color.B
        }

        g.Color.A = 1

    case 0x22:

        g.Texture.WriteCoord(data[1], g)

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

    case 0x30:

        g.LightData.DiffuseColor = Write15BitColor(data[1])
        g.LightData.AmbientColor = Write15BitColor(data[1]>>16)

        if setVertex := data[1] >> 15 != 0; setVertex {
            g.Color = g.LightData.DiffuseColor
        }

    case 0x31:

        g.LightData.SpecularColor = Write15BitColor(data[1])
        g.LightData.EmissionColor = Write15BitColor(data[1]>>16)
        g.LightData.UseSpecularTbl = utils.BitEnabled(data[1], 15)

    case 0x32:

        x := utils.Convert10ToFloat(uint16(data[1]), 9)
        y := utils.Convert10ToFloat(uint16(data[1] >> 10), 9)
        z := utils.Convert10ToFloat(uint16(data[1] >> 20), 9)
        v := gl.Vector{X: x, Y: y, Z: z}
        directionalMtx := g.MtxStacks.Stacks[2].CurrMtx

        idx := data[1] >> 30
        light := &g.LightData.Lights[idx]
        light.Vector = directionalMtx.MulPosition(v)
        light.Vector = light.Vector.Normalize()

        LINE_OF_SIGHT_VEC := gl.Vector{X: 0, Y: -1}
        light.HalfVector = light.Vector.Add(LINE_OF_SIGHT_VEC)
        light.HalfVector = light.HalfVector.Mul(gl.Vector{X: 0.5, Y: 0.5, Z: 0.5})

    case 0x33:

        idx := data[1] >> 30
        g.LightData.Lights[idx].Color = Write15BitColor(data[1])

    case 0x34:

        sTbl := &g.LightData.ShininessTbl

        var i uint32
        for _, v := range data[1:] {
            sTbl[i+0] = float64((v)       & 0xFF) / 256
            sTbl[i+1] = float64((v >> 8)  & 0xFF) / 256
            sTbl[i+2] = float64((v >> 16) & 0xFF) / 256
            sTbl[i+3] = float64((v >> 24) & 0xFF) / 256
            i += 4
        }

    case 0x40:

        if endPoly := len(g.ActivePoly.Vertices) != 0; endPoly {
            g.AddPolygon()
        }

        g.ActivePoly = g.PrepPoly

        // do not clear poly - need state of params for next
        //g.PrepPoly = Polygon{}

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

    case 0x72:

        g.VecTestData = g.VecTest(data, &g.MtxStacks.Stacks[2].CurrMtx)

    case 0x0:
    default:
        fmt.Printf("UNSETUP GX CMD %02X\n", cmd)
    }

    g.Data = []uint32{}
    g.UpdateClipMtx()
}

func (g *GeoEngine) UpdateClipMtx() {
    pos := &g.MtxStacks.Stacks[1].CurrMtx
    per := &g.MtxStacks.Stacks[0].CurrMtx
    g.ClipMatrix = pos.Mul(*per)
}

var paramCnt = map[uint32]int {
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
        if (
            cmd == 0x00 ||
            cmd == 0x11 ||
            cmd == 0x15 ||
            cmd == 0x41) {
                return true
            }
    }

    v, ok := paramCnt[cmd]
    if !ok {
        panic(fmt.Sprintf("UNKNOWN CMD GXFIFO % 2X", g.Data))
    }

    return v == params
}

func (g *GeoEngine) AddPolygon() {
    g.ActivePoly.Texture = g.Texture
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
