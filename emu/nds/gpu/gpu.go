package gpu

import (
	"github.com/aabalke/guac/emu/nds/gpu/gl"
)

const (
	SIZE_FIFO  = 256
	SIZE_PIPE  = 4
	SIZE_FLOOD = 64
)

var PARAM_LENGTH = [256]int{
	/* 0x00 */ 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	/* 0x10 */ 1, 0, 1, 1, 1, 0, 16, 12, 16, 12, 9, 3, 3, 0, 0, 0,
	/* 0x20 */ 1, 1, 1, 2, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0,
	/* 0x30 */ 1, 1, 1, 1, 32, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	/* 0x40 */ 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	/* 0x50 */ 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	/* 0x60 */ 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	/* 0x70 */ 3, 2, 1,
}

type GeoEngine struct {
	doDmas  func()
	Irq     Irq
	Buffers *Buffers

	Fifo, pipe, flood *Queue

	Viewport  Viewport
	Disp3dCnt Disp3dCnt

	// actively edited structs
	ActivePoly Polygon
	Color      gl.Color
	Texture    Texture
	LightData  gl.LightData

	pendingPolyAttr    uint32
	lastVertexPosition gl.VectorW

	pendingCmdsCnt  int
	pendingParamCnt int
	packedCmds      uint32

	params   [32]uint32
	paramCnt int

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

	// mtx stacks
	Stacks        [4]MtxStack
	Mode          uint32
	psp, csp, tsp int // ptrs

	Stalled bool
	Busy    bool
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

type FifoEntry struct {
	Cmd   uint32
	Param uint32
}

func NewGeoEngine(buffers *Buffers, irq Irq, vram *Vram, doDmas func()) *GeoEngine {
	g := &GeoEngine{
		Vram:         vram,
		Irq:          irq,
		doDmas:       doDmas,
		Buffers:      buffers,
		TextureCache: make(map[key]*[]gl.Color, 0),
		Fifo:         NewQueue(SIZE_FIFO),
		pipe:         NewQueue(SIZE_PIPE),
		flood:        NewQueue(SIZE_FLOOD),
	}

	g.Disp3dCnt.Fog = &g.Fog
	g.Stacks[0].Mtxs = make([]gl.Matrix, 1)
	g.Stacks[1].Mtxs = make([]gl.Matrix, 32)
	g.Stacks[2].Mtxs = make([]gl.Matrix, 32)
	g.Stacks[3].Mtxs = make([]gl.Matrix, 1)

	return g
}

func (g *GeoEngine) Step() {
	if g.pipe.IsEmpty() {
		return
	}

	if !g.Stalled {
		g.RunCommand()
		return
	}

	for g.Stalled {
		g.RunCommand()
	}
}

func (g *GeoEngine) WriteGX(v FifoEntry) {
	if g.Fifo.IsEmpty() && !g.pipe.IsFull() {
		g.pipe.Enqueue(v)
	} else {
		if g.Fifo.IsFull() {
			g.Stalled = true
			g.flood.Enqueue(v)
			return
		}
		g.Fifo.Enqueue(v)
	}

	g.Busy = true
}

func (g *GeoEngine) ReadGX() FifoEntry {
	v := g.pipe.Dequeue()

	if g.pipe.Length() <= (SIZE_PIPE / 2) {

		for range SIZE_PIPE / 2 {
			if !g.Fifo.IsEmpty() {
				g.pipe.Enqueue(g.Fifo.Dequeue())
			}
		}

		for !g.flood.IsEmpty() {
			if g.Fifo.IsFull() {
				break
			}

			g.Fifo.Enqueue(g.flood.Dequeue())

			if g.flood.IsEmpty() {
				g.Stalled = false
			}
		}

		if g.Fifo.Length() < (SIZE_FIFO / 2) {
			g.doDmas()
		}

		g.doIrq()
	}

	return v
}

func (g *GeoEngine) WriteFifo(v uint32) {
	if packedParams := g.pendingCmdsCnt != 0; packedParams {
		cmd := g.packedCmds & 0xFF
		g.WriteGX(FifoEntry{cmd, v})

		g.pendingParamCnt--
		if doNextPackedCmd := g.pendingParamCnt == 0; doNextPackedCmd {
			g.packedCmds >>= 8
			g.pendingCmdsCnt--

			if nop := g.packedCmds == 0; nop {
				g.pendingCmdsCnt = 0
			}

			g.unpack()
		}

		return
	}

	// new set of packed commands
	g.pendingCmdsCnt = 4
	g.packedCmds = v
	g.pendingParamCnt = 0

	g.unpack()
}

func (g *GeoEngine) unpack() {
	for g.pendingCmdsCnt > 0 {
		cmd := g.packedCmds & 0xFF
		g.pendingParamCnt = PARAM_LENGTH[cmd]

		if g.pendingParamCnt > 0 {
			return
		}

		g.WriteGX(FifoEntry{Cmd: cmd})
		g.packedCmds >>= 8
		g.pendingCmdsCnt--
		if g.packedCmds == 0 {
			g.pendingCmdsCnt = 0
		}
	}
}

func (g *GeoEngine) RunCommand() {
	v := g.ReadGX()
	g.Execute(v)

	g.Busy = !g.pipe.IsEmpty()
}

func (g *GeoEngine) Execute(v FifoEntry) {
	g.params[g.paramCnt] = v.Param
	g.paramCnt++

	if g.paramCnt < PARAM_LENGTH[v.Cmd] {
		return
	}

	if noParams := g.paramCnt <= 0; noParams {
		commands[v.Cmd](g, nil)
	} else {
		commands[v.Cmd](g, g.params[:g.paramCnt])
	}

	g.paramCnt = 0

	if g.UpdateClipMtx {
		// position is world space transform
		// perspective is perspecitve space transform

		pos := g.Stacks[1].Curr
		per := g.Stacks[0].Curr

		g.ClipMatrix = pos.Mul(per)
		g.WorldMatrix = pos
		g.UpdateClipMtx = false
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

func (g *GeoEngine) doIrq() {
	if g.FifoIrq == 1 && g.Fifo.Length() < (SIZE_FIFO/2) {
		g.Irq.SetIRQ(21)
	}

	if g.FifoIrq == 2 && g.Fifo.IsEmpty() {
		g.Irq.SetIRQ(21)
	}
}
