package gpu

import (
	"fmt"

	"github.com/aabalke/guac/emu/nds/gpu/gl"
)

var PARAM_LENGTH = [256]int{
	/*0x00*/ 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	/*0x10*/ 1, 1, 1, 1, 1, 1, 16, 12, 16, 12, 9, 3, 3, 0, 0, 0,
	/*0x20*/ 1, 1, 1, 2, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0,
	/*0x30*/ 1, 1, 1, 1, 32, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	/*0x40*/ 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	/*0x50*/ 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	/*0x60*/ 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	/*0x70*/ 3, 2, 1,
}

type GeoEngine struct {
	Irq     Irq
	Buffers *Buffers
	Data    []uint32

	Viewport  Viewport
	Disp3dCnt Disp3dCnt

	// actively edited structs
	ActivePoly Polygon
	Color      gl.Color
	Texture    Texture
	LightData  gl.LightData

	pendingPolyAttr    uint32
	lastVertexPosition gl.VectorW

	Packed     bool
	PackedCmds [4]uint32
	PackedIdx  uint8

	ClipMatrix  gl.Matrix
	WorldMatrix gl.Matrix // just for export
	PosTestData [4]uint32
	VecTestData [3]uint16
	ToonTbl     [32]gl.Color

	TextureCache TextureCache
	Vram         *Vram

	Fog gl.Fog

	UpdateClipMtx bool

	// GxStat
	StackOverflow bool
	TestInView    bool
	FifoIrq       uint8
	FifoEntries   uint16

	// mtx stacks
	Stacks        [4]MtxStack
	Mode          uint32
	psp, csp, tsp int // ptrs
}

type Texture struct {
	Sv, Tv             float32
	S, T               float32
	VramOffset         uint32
	SizeS, SizeT       uint32
	Format             uint32
	TransformationMode uint32
	PaletteBaseAddr    uint32
	PitchShift         uint32
	param              uint32
	RepeatS, RepeatT   bool
	FlipS, FlipT       bool
	TransparentZero    bool
}

type MtxStack struct {
	Curr gl.Matrix
	Mtxs []gl.Matrix
}

func NewGeoEngine(buffers *Buffers, irq Irq, vram *Vram) *GeoEngine {
	g := &GeoEngine{
		Vram:         vram,
		Irq:          irq,
		Buffers:      buffers,
		TextureCache: make(map[key]*[]gl.Color, 0),
	}

	g.Disp3dCnt.Fog = &g.Fog
	g.Stacks[0].Mtxs = make([]gl.Matrix, 1)
	g.Stacks[1].Mtxs = make([]gl.Matrix, 32)
	g.Stacks[2].Mtxs = make([]gl.Matrix, 32)
	g.Stacks[3].Mtxs = make([]gl.Matrix, 1)

	return g
}

func (g *GeoEngine) Fifo(v uint32) {
	// this will be buggy, need to handle if packed cmd sets data to 0 ( add cmd???)
	if cmd := len(g.Data) == 0; cmd {

		if packed := v&^0xFF != 0; packed {

			g.PackedCmds = [4]uint32{}
			g.PackedIdx = 0
			g.Packed = true

			g.PackedCmds[0] = v & 0xFF
			g.PackedCmds[1] = (v >> 8) & 0xFF
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

	if finishedPacked := g.PackedIdx == 0; finishedPacked {
		g.Data = []uint32{}
		//fmt.Printf("Finished 4 Packed Commands\n")

		//g.Packed = false
		return
	}

	g.Data = append(g.Data, g.PackedCmds[g.PackedIdx])

	g.PackedFifo()
}

func (g *GeoEngine) Cmd(fifo bool, data []uint32) {
	if !g.ValidParamCount(fifo) {
		return
	}

	if cmd := data[0]; len(data) > 1 {
		commands[cmd](g, data[1:])
	} else {
		commands[cmd](g, nil)
	}

	if g.UpdateClipMtx {
		// position is world space transform
		// perspective is perspecitve space transform

		pos := g.Stacks[1].Curr
		per := g.Stacks[0].Curr

		g.ClipMatrix = pos.Mul(per)
		g.WorldMatrix = pos
		g.UpdateClipMtx = false
	}

	g.Data = []uint32{}
}

func (g *GeoEngine) ValidParamCount(fifo bool) bool {
	cmd := g.Data[0]
	params := len(g.Data) - 1

	// when using fifo, sometimes no param provided, but when using io
	// dummy param is provided.
	if fifo && params == 0 {
		if cmd == 0x00 ||
			cmd == 0x11 ||
			cmd == 0x15 ||
			cmd == 0x41 {
			return true
		}
	}

	switch v := PARAM_LENGTH[cmd]; v {
	case 0:
		panic(fmt.Sprintf("UNKNOWN CMD GXFIFO % 2X", g.Data))
	default:
		return v == params
	}
}

//go:inline
func Write15BitColor(v uint32) gl.Color {
	return gl.MakeColorFrom15Bit(
		uint8(v&0x1F),
		uint8((v>>5)&0x1F),
		uint8((v>>10)&0x1F),
	)
}

//go:inline
func frustumPlaneTest(verts []gl.VectorW) bool {
	for plane := range 6 {
		allOutside := true

		for _, v := range verts {
			if insidePlane(v, plane) {
				allOutside = false
				break
			}
		}

		if allOutside {
			return false
		}
	}

	return true
}

//go:inline
func insidePlane(v gl.VectorW, plane int) bool {
	switch plane {
	case 0:
		return v.X >= -v.W // left
	case 1:
		return v.X <= v.W // right
	case 2:
		return v.Y >= -v.W // bottom
	case 3:
		return v.Y <= v.W // top
	case 4:
		return v.Z >= -v.W // near
	case 5:
		return v.Z <= v.W // far
	default:
		return false
	}
}
