package ppu

import (
	"encoding/binary"
	"fmt"

	"github.com/aabalke/guac/emu/cpu"
	"github.com/aabalke/guac/emu/nds/rast"
	"github.com/aabalke/guac/emu/nds/utils"
)

var _ = fmt.Sprintf

const (
	SCREEN_WIDTH  = 256
	SCREEN_HEIGHT = 192
)

type PPU struct {
	EngineA    Engine
	EngineB    Engine
	Rasterizer *rast.Rasterizer

	// these values are updated in PowCnt1
	// need to impliment port disabling
	LcdEnabled                      bool
	EngineA2D, EngineB2D            bool
	RenderingEngine, GeometryEngine bool
	TopA                            bool

	Vram VRAM

	Capture     Capture

	WHITE_SCANLINE []uint8
}

type Engine struct {
	Pixels []byte
	IsB    bool

    Pram PRAM

    Backdrop *uint16 // always first palette in engine pram

	Dispcnt Dispcnt

	MasterBright MasterBright

	Objects     [128]Object
	Backgrounds [4]Background
	Windows     Windows
	Blend       Blend
	Mosaic      Mosaic

    BgPriorities [4]struct {
        Idx [4]uint32
        Cnt int
    }
    ObjPriorities [4]struct {
        Idx [128]uint32
        Cnt int
    }

    ExtBgSlots [4]*[0x2000]uint8
    ExtObj     *[0x4000]uint8

    BgPalettes [4][SCREEN_WIDTH]uint32
    BgOks      [4][SCREEN_WIDTH]bool
    BgAlphas   [4][SCREEN_WIDTH]float32

    ObjPalettes [128][SCREEN_WIDTH]uint32
    ObjOks      [128][SCREEN_WIDTH]bool
}

type Dispcnt struct {
	Mode               uint32
	Is3D               bool
	TileObj1D          bool
	BitmapObj256       bool
	BitmapObj1D        bool
	ForcedBlank        bool
	DisplayObj         bool
	DisplayWin0        bool
	DisplayWin1        bool
	DisplayObjWin      bool
	DisplayMode        uint32
	VramBlock          uint32
	TileObjBoundary    uint32
	BitmapObjBoundary  bool
	HBlankIntervalFree bool
	CharBase           uint32
	ScreenBase         uint32
	BgExtPal           bool
	ObjExtPal          bool
}

// blends are [6]... because Bg0, Bg1, Bg2, Bg3, Obj, Bd
type Blend struct {
	Mode          uint32
	a, b          [6]bool
	aEv, bEv, yEv uint32
}

type Windows struct {
	Enabled            bool
	Win0, Win1, WinObj Window
	OutBg              [4]bool
	OutObj, OutBld     bool
}

type Window struct {
	Enabled        bool
	L, R, T, B     uint32
	oL, oR, oT, oB uint32
	InBg           [4]bool
	InObj, InBld   bool
}

type Mosaic struct {
	BgH, BgV, ObjH, ObjV uint32
}

type Background struct {

    // used for getPriority -> standard func
    MasterEnabled      bool

	Enabled            bool
	W, H               uint32
	Pa, Pb, Pc, Pd     uint32
	Priority           uint32
	CharBaseBlock      uint32
	Mosaic             bool
	Palette256         bool
	ScreenBaseBlock    uint32
	AffineWrap         bool
	Size               uint32
	XOffset, YOffset   uint32
	aXOffset, aYOffset uint32
	Affine             bool

	//PbCalc, PdCalc float64
	OutX, OutY float64

	Type uint8

	AltExtPalSlot bool
}

const (
	BG_TYPE_TEX = 0
	BG_TYPE_AFF = 1
	BG_TYPE_LAR = 2
	BG_TYPE_3D  = 3
	BG_TYPE_BGM = 4
	BG_TYPE_256 = 5
	BG_TYPE_DIR = 6
)

type Object struct {

    MasterEnabled  bool

	X, Y, W, H     uint32
	Pa, Pb, Pc, Pd float32
	RotScale       bool
	DoubleSize     bool
	Disable        bool
	Mode           uint32
	Mosaic         bool
	Palette256     bool
	Shape          uint32
	HFlip, VFlip   bool
	Size           uint32
	RotParams      uint32
	CharName       uint32
	Priority       uint32
	Palette        uint32
	OneDimensional bool

	ObjTileMapping uint8
	ObjBmpMapping  uint8

	TileBoundaryShift uint32
	BmpBoundaryShift  uint32
	BmpBoundaryMask   uint32
}

func NewPPU(irq *cpu.Irq) *PPU {

	p := &PPU{}

	p.EngineA.Pixels = make([]byte, SCREEN_WIDTH*SCREEN_HEIGHT*4)
	p.EngineB.Pixels = make([]byte, SCREEN_WIDTH*SCREEN_HEIGHT*4)
	p.EngineB.IsB = true

	p.Rasterizer = rast.NewRasterizer(&p.Vram, irq)
	p.Capture.Init(
		&p.Vram,
		p,
		&p.EngineA.Dispcnt.VramBlock,
		&p.EngineA.Pixels,
		&p.Rasterizer.Render.Pixels,
	)

	// screenoff optimization
	p.WHITE_SCANLINE = make([]uint8, SCREEN_WIDTH*4)
	for i := range len(p.WHITE_SCANLINE) {
		p.WHITE_SCANLINE[i] = 0xFF
	}

    p.EngineA.Backdrop = &p.EngineA.Pram.Bg[0]
    p.EngineB.Backdrop = &p.EngineB.Pram.Bg[0]

	return p
}

func (p *PPU) Update(addr, v uint32) {

	if engineA := addr < 0x60 || addr >= 0x6C && addr < 0x6E; engineA && p.EngineA2D {

		p.EngineA.UpdateEngine(addr, v)
		return
	}

	if capture := addr >= 0x60 && addr < 0x68; capture {
		//fmt.Printf("CAPTURE %08X %02X\n", addr, v)
		return
	}

	if engineRender := addr >= 0x320 && addr < 0x400; engineRender && p.RenderingEngine {
		return
	}

	if engineGeo := addr >= 0x400 && addr < 0x700; engineGeo && p.GeometryEngine {
		return
	}

	if engineB := addr >= 0x1000 && addr < 0x1060 || addr >= 0x106C && addr < 0x106E; engineB && p.EngineB2D {
		p.EngineB.UpdateEngine(addr&0xFF, v)
		return
	}
}

func (e *Engine) UpdateEngine(addr, v uint32) {

	if win := addr >= 0x40 && addr < 0x4C; win {
		e.UpdateWin(addr, v)
		return
	}

	if bgs := addr >= 0x08 && addr < 0x40; bgs {
		e.UpdateBackgrounds(addr, v)
		return
	}

	switch addr {
	case 0x0:

		e.Dispcnt.Mode = v & 0b111
		e.Dispcnt.Is3D = (v>>3)&1 != 0

		e.Dispcnt.TileObj1D = (v>>4)&1 != 0
		e.Dispcnt.BitmapObj256 = (v>>5)&1 != 0
		e.Dispcnt.BitmapObj1D = (v>>6)&1 != 0
		e.Dispcnt.ForcedBlank = (v>>7)&1 != 0

		e.UpdateObjMapping(&e.Dispcnt)
        e.setBgType(0)
        e.setBgType(1)
        e.setBgType(2)
        e.setBgType(3)

	case 0x1:
		e.Dispcnt.DisplayObj = (v>>4)&1 != 0
		e.Dispcnt.DisplayWin0 = (v>>5)&1 != 0
		e.Dispcnt.DisplayWin1 = (v>>6)&1 != 0
		e.Dispcnt.DisplayObjWin = (v>>7)&1 != 0

		e.Backgrounds[0].Enabled = (v>>0)&1 != 0
		e.Backgrounds[1].Enabled = (v>>1)&1 != 0
		e.Backgrounds[2].Enabled = (v>>2)&1 != 0
		e.Backgrounds[3].Enabled = (v>>3)&1 != 0

		wins := &e.Windows
		wins.Win0.Enabled = e.Dispcnt.DisplayWin0
		wins.Win1.Enabled = e.Dispcnt.DisplayWin1
		wins.WinObj.Enabled = e.Dispcnt.DisplayObjWin && e.Dispcnt.DisplayObj
		wins.Enabled = wins.Win0.Enabled || wins.Win1.Enabled || wins.WinObj.Enabled

		e.UpdateObjMapping(&e.Dispcnt)

	case 0x2:
		e.Dispcnt.DisplayMode = v & 0b11
		e.Dispcnt.VramBlock = (v >> 2) & 0b11
		e.Dispcnt.TileObjBoundary = (v >> 4) & 0b11

		e.Dispcnt.BitmapObjBoundary = (v>>6)&1 != 0
		e.Dispcnt.HBlankIntervalFree = (v>>7)&1 != 0
		e.UpdateObjMapping(&e.Dispcnt)

	case 0x3:

		e.Dispcnt.CharBase = (v & 0b111) * 0x1_0000
		e.Dispcnt.ScreenBase = ((v >> 3) & 0b111) * 0x1_0000

		e.Dispcnt.BgExtPal = (v>>6)&1 != 0
		e.Dispcnt.ObjExtPal = (v>>7)&1 != 0
		e.UpdateObjMapping(&e.Dispcnt)

	case 0x4C:

		e.Mosaic.BgH = v & 0xF
		e.Mosaic.BgV = (v >> 4) & 0xF

	case 0x4D:

		e.Mosaic.ObjH = v & 0xF
		e.Mosaic.ObjV = (v >> 4) & 0xF

	case 0x50:
		e.Blend.a[0] = (v>>0)&1 != 0
		e.Blend.a[1] = (v>>1)&1 != 0
		e.Blend.a[2] = (v>>2)&1 != 0
		e.Blend.a[3] = (v>>3)&1 != 0
		e.Blend.a[4] = (v>>4)&1 != 0
		e.Blend.a[5] = (v>>5)&1 != 0
		e.Blend.Mode = (v >> 6) & 0b11

	case 0x51:
		e.Blend.b[0] = (v>>0)&1 != 0
		e.Blend.b[1] = (v>>1)&1 != 0
		e.Blend.b[2] = (v>>2)&1 != 0
		e.Blend.b[3] = (v>>3)&1 != 0
		e.Blend.b[4] = (v>>4)&1 != 0
		e.Blend.b[5] = (v>>5)&1 != 0

	case 0x52:
		e.Blend.aEv = min(16, v&0x1F)

	case 0x53:
		e.Blend.bEv = min(16, v&0x1F)

	case 0x54:
		e.Blend.yEv = min(16, v&0x1F)
	case 0x6C:
		e.MasterBright.Write(uint8(v), 0)
	case 0x6D:
		e.MasterBright.Write(uint8(v), 1)
	}

}

func (engine *Engine) UpdateWin(addr uint32, v uint32) {

	wins := &engine.Windows
	win0 := &engine.Windows.Win0
	win1 := &engine.Windows.Win1
	winObj := &engine.Windows.WinObj

	const (
		WIN0Ha = 0x40
		WIN0Hb = 0x41
		WIN1Ha = 0x42
		WIN1Hb = 0x43
		WIN0Va = 0x44
		WIN0Vb = 0x45
		WIN1Va = 0x46
		WIN1Vb = 0x47
		WININ0 = 0x48
		WININ1 = 0x49
		WINOUT = 0x4A
		WINOBJ = 0x4B

		SCREEN_WIDTH  = 256
		SCREEN_HEIGHT = 192
	)

	switch addr {
	case WIN0Ha:
		win0.oR = v
		win0.R = v

		if win0.oR > SCREEN_WIDTH || win0.oL > win0.oR {
			win0.R = SCREEN_WIDTH
		}

	case WIN0Hb:
		win0.oL = v
		win0.L = v

		if win0.oR > SCREEN_WIDTH || win0.oL > win0.oR {
			win0.R = SCREEN_WIDTH
		}

	case WIN1Ha:
		win1.oR = v
		win1.R = v

		if win1.oR > SCREEN_WIDTH || win1.oL > win1.oR {
			win1.R = SCREEN_WIDTH
		}

	case WIN1Hb:
		win1.oL = v
		win1.L = v

		if win1.oR > SCREEN_WIDTH || win1.oL > win1.oR {
			win1.R = SCREEN_WIDTH
		}

	case WIN0Va:
		win0.oB = v
		win0.B = v

		if win0.oB > SCREEN_HEIGHT || win0.oT > win0.oB {
			win0.B = SCREEN_HEIGHT
		}

	case WIN0Vb:
		win0.oT = v
		win0.T = v

		if win0.oB > SCREEN_HEIGHT || win0.oT > win0.oB {
			win0.B = SCREEN_HEIGHT
		}

	case WIN1Va:
		win1.oB = v
		win1.B = v

		if win1.oB > SCREEN_HEIGHT || win1.oT > win1.oB {
			win1.B = SCREEN_HEIGHT
		}

	case WIN1Vb:
		win1.oT = v
		win1.T = v

		if win1.oB > SCREEN_HEIGHT || win1.oT > win1.oB {
			win1.B = SCREEN_HEIGHT
		}

	case WININ0:
		win0.InBg[0] = (v>>0)&1 != 0
		win0.InBg[1] = (v>>1)&1 != 0
		win0.InBg[2] = (v>>2)&1 != 0
		win0.InBg[3] = (v>>3)&1 != 0
		win0.InObj = (v>>4)&1 != 0
		win0.InBld = (v>>5)&1 != 0
	case WININ1:
		win1.InBg[0] = (v>>0)&1 != 0
		win1.InBg[1] = (v>>1)&1 != 0
		win1.InBg[2] = (v>>2)&1 != 0
		win1.InBg[3] = (v>>3)&1 != 0
		win1.InObj = (v>>4)&1 != 0
		win1.InBld = (v>>5)&1 != 0
	case WINOUT:
		wins.OutBg[0] = (v>>0)&1 != 0
		wins.OutBg[1] = (v>>1)&1 != 0
		wins.OutBg[2] = (v>>2)&1 != 0
		wins.OutBg[3] = (v>>3)&1 != 0
		wins.OutObj = (v>>4)&1 != 0
		wins.OutBld = (v>>5)&1 != 0
	case WINOBJ:
		winObj.InBg[0] = (v>>0)&1 != 0
		winObj.InBg[1] = (v>>1)&1 != 0
		winObj.InBg[2] = (v>>2)&1 != 0
		winObj.InBg[3] = (v>>3)&1 != 0
		winObj.InObj = (v>>4)&1 != 0
		winObj.InBld = (v>>5)&1 != 0
	}
}

func (p *Engine) UpdateBackgrounds(addr, v uint32) {

	switch addr {
	case 0x08:
		p.Backgrounds[0].Priority = v & 0b11
		p.Backgrounds[0].CharBaseBlock = ((v >> 2) & 0xF) * 0x4000
		p.Backgrounds[0].Mosaic = (v>>6)&1 != 0
		p.Backgrounds[0].Palette256 = (v>>7)&1 != 0
	case 0x09:
		p.Backgrounds[0].ScreenBaseBlock = (v & 0x1F) * 0x800
		p.Backgrounds[0].AltExtPalSlot = (v>>5)&1 != 0
		p.Backgrounds[0].Size = (v >> 6) & 0b11

	case 0x0A:
		p.Backgrounds[1].Priority = v & 0b11
		p.Backgrounds[1].CharBaseBlock = ((v >> 2) & 0xF) * 0x4000
		p.Backgrounds[1].Mosaic = (v>>6)&1 != 0
		p.Backgrounds[1].Palette256 = (v>>7)&1 != 0
	case 0x0B:
		p.Backgrounds[1].ScreenBaseBlock = (v & 0x1F) * 0x800
		p.Backgrounds[1].AltExtPalSlot = (v>>5)&1 != 0
		p.Backgrounds[1].Size = (v >> 6) & 0b11

	case 0x0C:
		p.Backgrounds[2].Priority = v & 0b11
		p.Backgrounds[2].CharBaseBlock = ((v >> 2) & 0xF) * 0x4000
		p.Backgrounds[2].Mosaic = (v>>6)&1 != 0
		p.Backgrounds[2].Palette256 = (v>>7)&1 != 0
        p.setBgType(2)
	case 0x0D:
		p.Backgrounds[2].ScreenBaseBlock = (v & 0x1F) * 0x800
		p.Backgrounds[2].AffineWrap = (v>>5)&1 != 0
		p.Backgrounds[2].Size = (v >> 6) & 0b11

	case 0x0E:
		p.Backgrounds[3].Priority = v & 0b11
		p.Backgrounds[3].CharBaseBlock = ((v >> 2) & 0xF) * 0x4000
		p.Backgrounds[3].Mosaic = (v>>6)&1 != 0
		p.Backgrounds[3].Palette256 = (v>>7)&1 != 0
        p.setBgType(3)

	case 0x0F:
		p.Backgrounds[3].ScreenBaseBlock = (v & 0x1F) * 0x800
		p.Backgrounds[3].AffineWrap = (v>>5)&1 != 0
		p.Backgrounds[3].Size = (v >> 6) & 0b11

	case 0x10:
		p.Backgrounds[0].XOffset &^= 0xFF
		p.Backgrounds[0].XOffset |= v
	case 0x11:
		p.Backgrounds[0].XOffset &= 0xFF
		p.Backgrounds[0].XOffset |= v << 8
	case 0x12:
		p.Backgrounds[0].YOffset &^= 0xFF
		p.Backgrounds[0].YOffset |= v
	case 0x13:
		p.Backgrounds[0].YOffset &= 0xFF
		p.Backgrounds[0].YOffset |= v << 8

	case 0x14:
		p.Backgrounds[1].XOffset &^= 0xFF
		p.Backgrounds[1].XOffset |= v
	case 0x15:
		p.Backgrounds[1].XOffset &= 0xFF
		p.Backgrounds[1].XOffset |= v << 8
	case 0x16:
		p.Backgrounds[1].YOffset &^= 0xFF
		p.Backgrounds[1].YOffset |= v
	case 0x17:
		p.Backgrounds[1].YOffset &= 0xFF
		p.Backgrounds[1].YOffset |= v << 8

	case 0x18:
		p.Backgrounds[2].XOffset &^= 0xFF
		p.Backgrounds[2].XOffset |= v
	case 0x19:
		p.Backgrounds[2].XOffset &= 0xFF
		p.Backgrounds[2].XOffset |= v << 8
	case 0x1A:
		p.Backgrounds[2].YOffset &^= 0xFF
		p.Backgrounds[2].YOffset |= v
	case 0x1B:
		p.Backgrounds[2].YOffset &= 0xFF
		p.Backgrounds[2].YOffset |= v << 8

	case 0x1C:
		p.Backgrounds[3].XOffset &^= 0xFF
		p.Backgrounds[3].XOffset |= v
	case 0x1D:
		p.Backgrounds[3].XOffset &= 0xFF
		p.Backgrounds[3].XOffset |= v << 8
	case 0x1E:
		p.Backgrounds[3].YOffset &^= 0xFF
		p.Backgrounds[3].YOffset |= v
	case 0x1F:
		p.Backgrounds[3].YOffset &= 0xFF
		p.Backgrounds[3].YOffset |= v << 8

	case 0x20:
		p.Backgrounds[2].Pa &^= 0xFF
		p.Backgrounds[2].Pa |= v
	case 0x21:
		p.Backgrounds[2].Pa &= 0xFF
		p.Backgrounds[2].Pa |= v << 8
	case 0x22:
		p.Backgrounds[2].Pb &^= 0xFF
		p.Backgrounds[2].Pb |= v
	case 0x23:
		p.Backgrounds[2].Pb &= 0xFF
		p.Backgrounds[2].Pb |= v << 8
	case 0x24:
		p.Backgrounds[2].Pc &^= 0xFF
		p.Backgrounds[2].Pc |= v
	case 0x25:
		p.Backgrounds[2].Pc &= 0xFF
		p.Backgrounds[2].Pc |= v << 8
	case 0x26:
		p.Backgrounds[2].Pd &^= 0xFF
		p.Backgrounds[2].Pd |= v
	case 0x27:
		p.Backgrounds[2].Pd &= 0xFF
		p.Backgrounds[2].Pd |= v << 8

	case 0x28:
		p.Backgrounds[2].aXOffset &^= 0xFF
		p.Backgrounds[2].aXOffset |= v
		p.Backgrounds[2].BgAffineReset()
	case 0x29:
		p.Backgrounds[2].aXOffset &^= 0xFF00
		p.Backgrounds[2].aXOffset |= v << 8
		p.Backgrounds[2].BgAffineReset()
	case 0x2A:
		p.Backgrounds[2].aXOffset &^= 0xFF0000
		p.Backgrounds[2].aXOffset |= v << 16
		p.Backgrounds[2].BgAffineReset()
	case 0x2B:
		p.Backgrounds[2].aXOffset &^= 0xFF000000
		p.Backgrounds[2].aXOffset |= v << 24
		p.Backgrounds[2].BgAffineReset()

	case 0x2C:
		p.Backgrounds[2].aYOffset &^= 0xFF
		p.Backgrounds[2].aYOffset |= v
		p.Backgrounds[2].BgAffineReset()
	case 0x2D:
		p.Backgrounds[2].aYOffset &^= 0xFF00
		p.Backgrounds[2].aYOffset |= v << 8
		p.Backgrounds[2].BgAffineReset()
	case 0x2E:
		p.Backgrounds[2].aYOffset &^= 0xFF0000
		p.Backgrounds[2].aYOffset |= v << 16
		p.Backgrounds[2].BgAffineReset()
	case 0x2F:
		p.Backgrounds[2].aYOffset &^= 0xFF000000
		p.Backgrounds[2].aYOffset |= v << 24
		p.Backgrounds[2].BgAffineReset()

	case 0x30:
		p.Backgrounds[3].Pa &^= 0xFF
		p.Backgrounds[3].Pa |= v
	case 0x31:
		p.Backgrounds[3].Pa &= 0xFF
		p.Backgrounds[3].Pa |= v << 8
	case 0x32:
		p.Backgrounds[3].Pb &^= 0xFF
		p.Backgrounds[3].Pb |= v
	case 0x33:
		p.Backgrounds[3].Pb &= 0xFF
		p.Backgrounds[3].Pb |= v << 8
	case 0x34:
		p.Backgrounds[3].Pc &^= 0xFF
		p.Backgrounds[3].Pc |= v
	case 0x35:
		p.Backgrounds[3].Pc &= 0xFF
		p.Backgrounds[3].Pc |= v << 8
	case 0x36:
		p.Backgrounds[3].Pd &^= 0xFF
		p.Backgrounds[3].Pd |= v
	case 0x37:
		p.Backgrounds[3].Pd &= 0xFF
		p.Backgrounds[3].Pd |= v << 8

	case 0x38:
		p.Backgrounds[3].aXOffset &^= 0xFF
		p.Backgrounds[3].aXOffset |= v
		p.Backgrounds[3].BgAffineReset()
	case 0x39:
		p.Backgrounds[3].aXOffset &^= 0xFF00
		p.Backgrounds[3].aXOffset |= v << 8
		p.Backgrounds[3].BgAffineReset()
	case 0x3A:
		p.Backgrounds[3].aXOffset &^= 0xFF0000
		p.Backgrounds[3].aXOffset |= v << 16
		p.Backgrounds[3].BgAffineReset()
	case 0x3B:
		p.Backgrounds[3].aXOffset &^= 0xFF000000
		p.Backgrounds[3].aXOffset |= v << 24
		p.Backgrounds[3].BgAffineReset()

	case 0x3C:
		p.Backgrounds[3].aYOffset &^= 0xFF
		p.Backgrounds[3].aYOffset |= v
		p.Backgrounds[3].BgAffineReset()
	case 0x3D:
		p.Backgrounds[3].aYOffset &^= 0xFF00
		p.Backgrounds[3].aYOffset |= v << 8
		p.Backgrounds[3].BgAffineReset()
	case 0x3E:
		p.Backgrounds[3].aYOffset &^= 0xFF0000
		p.Backgrounds[3].aYOffset |= v << 16
		p.Backgrounds[3].BgAffineReset()
	case 0x3F:
		p.Backgrounds[3].aYOffset &^= 0xFF000000
		p.Backgrounds[3].aYOffset |= v << 24
		p.Backgrounds[3].BgAffineReset()
	}
}

func (bg *Background) SetSize() {

	switch bg.Type {
	case BG_TYPE_TEX:
		switch bg.Size {
		case 0:
			bg.W, bg.H = 256, 256
		case 1:
			bg.W, bg.H = 512, 256
		case 2:
			bg.W, bg.H = 256, 512
		case 3:
			bg.W, bg.H = 512, 512
		default:
			panic("PROHIBITTED BG SIZE")
		}
	case BG_TYPE_AFF, BG_TYPE_BGM:
		switch bg.Size {
		case 0:
			bg.W, bg.H = 128, 128
		case 1:
			bg.W, bg.H = 256, 256
		case 2:
			bg.W, bg.H = 512, 512
		case 3:
			bg.W, bg.H = 1024, 1024
		default:
			panic("PROHIBITTED AFFINE BG SIZE")
		}
	case BG_TYPE_LAR:
		switch bg.Size {
		case 0:
			bg.W, bg.H = 512, 1024
		case 1:
			bg.W, bg.H = 1024, 512
		default:
			panic("PROHIBITTED LARGE BITMAP BG SIZE")
		}
	case BG_TYPE_256, BG_TYPE_DIR:
		switch bg.Size {
		case 0:
			bg.W, bg.H = 128, 128
		case 1:
			bg.W, bg.H = 256, 256
		case 2:
			bg.W, bg.H = 512, 256
		case 3:
			bg.W, bg.H = 512, 512
		default:
			panic("PROHIBITTED AFFINE BG SIZE")
		}
	}
}

func (obj *Object) SetSize(shape, size uint32) {

	const (
		SQUARE     = 0
		HORIZONTAL = 1
		VERTICAL   = 2
	)

	switch shape {
	case SQUARE:
		switch size {
		case 0:
			obj.H, obj.W = 8, 8
			return
		case 1:
			obj.H, obj.W = 16, 16
			return
		case 2:
			obj.H, obj.W = 32, 32
			return
		case 3:
			obj.H, obj.W = 64, 64
			return
		}
	case HORIZONTAL:
		switch size {
		case 0:
			obj.H, obj.W = 8, 16
			return
		case 1:
			obj.H, obj.W = 8, 32
			return
		case 2:
			obj.H, obj.W = 16, 32
			return
		case 3:
			obj.H, obj.W = 32, 64
			return
		}
	case VERTICAL:
		switch size {
		case 0:
			obj.H, obj.W = 16, 8
			return
		case 1:
			obj.H, obj.W = 32, 8
			return
		case 2:
			obj.H, obj.W = 32, 16
			return
		case 3:
			obj.H, obj.W = 64, 32
			return
		}
	}

}

func (bg *Background) BgAffineReset() {
	bg.OutX = float64(utils.Convert28ToFloat(bg.aXOffset, 8))
	bg.OutY = float64(utils.Convert28ToFloat(bg.aYOffset, 8))
}

func (bg *Background) BgAffineUpdate() {
	bg.OutX += float64(utils.Convert16ToFloat(uint16(bg.Pb), 8))
	bg.OutY += float64(utils.Convert16ToFloat(uint16(bg.Pd), 8))
}

func (p *PPU) UpdateOAM(relAddr uint32, v uint8, oam *[0x800]uint8) {

	relAddr &= 0x7FF

	engine := &p.EngineA
	if relAddr >= 0x400 {
		engine = &p.EngineB
		//relAddr -= 0x400
	}

	attrIdx := relAddr % 8

	if affineParam := attrIdx == 6 || attrIdx == 7; affineParam {
		p.UpdateAffine(relAddr, engine, oam)
		return
	}

	objIdx := (relAddr & 0x3FF) / 8

	obj := &engine.Objects[objIdx]

	attr := uint32(oam[relAddr])

	switch attrIdx {
	case 0:
		obj.Y = attr
	case 1:

		obj.RotScale = (attr>>0)&1 != 0
		obj.Mode = (attr >> 2) & 0b11
		obj.Mosaic = (attr>>4)&1 != 0
		obj.Palette256 = (attr>>5)&1 != 0
		obj.Shape = (attr >> 6) & 0b11
		obj.SetSize(obj.Shape, obj.Size)

		if obj.RotScale {
			obj.DoubleSize = (attr>>1)&1 != 0
			UpdateAffineParams(obj, oam, engine.IsB)
		} else {
			obj.Disable = (attr>>1)&1 != 0
		}

	case 2:
		obj.X &^= 0xFF
		obj.X |= attr
	case 3:
		obj.X &= 0xFF
		obj.X |= (attr & 0b1) << 8
		obj.Size = (attr >> 6) & 0b11
		obj.SetSize(obj.Shape, obj.Size)

		if obj.RotScale {
			obj.RotParams = (attr >> 1) & 0x1F
			UpdateAffineParams(obj, oam, engine.IsB)
		}
		obj.HFlip = (attr>>4)&1 != 0
		obj.VFlip = (attr>>5)&1 != 0

	case 4:
		obj.CharName &^= 0xFF
		obj.CharName |= attr
	case 5:
		obj.CharName &= 0xFF
		obj.CharName |= (attr & 0b11) << 8
		obj.Priority = (attr >> 2) & 0b11
		obj.Palette = (attr >> 4) & 0xF
	}
}

func UpdateAffineParams(obj *Object, oam *[0x800]uint8, isB bool) {
	paramsAddr := obj.RotParams * 0x20

	if isB {
		paramsAddr += 0x400
	}

	obj.Pa = float32(int16(binary.LittleEndian.Uint16(oam[paramsAddr+0x06:]))) / 256
	obj.Pb = float32(int16(binary.LittleEndian.Uint16(oam[paramsAddr+0x0E:]))) / 256
	obj.Pc = float32(int16(binary.LittleEndian.Uint16(oam[paramsAddr+0x16:]))) / 256
	obj.Pd = float32(int16(binary.LittleEndian.Uint16(oam[paramsAddr+0x1E:]))) / 256
}

func (p *PPU) UpdateAffine(relAddr uint32, engine *Engine, oam *[0x800]uint8) {

	paramIdx := (relAddr &^ 0b1) / 0x20

	for i := range 128 {

		obj := &engine.Objects[i]

		if !obj.RotScale {
			continue
		}

		if obj.RotParams != paramIdx {
			continue
		}

		UpdateAffineParams(obj, oam, engine.IsB)
	}
}

const (
	OBJ_TIL_STD_2D = 0
	OBJ_TIL_STD_1D = 1
	OBJ_TIL_064_1D = 2
	OBJ_TIL_128_1D = 3
	OBJ_TIL_256_1D = 4

	OBJ_BMP_128_2D = 0
	OBJ_BMP_256_2D = 1
	OBJ_BMP_128_1D = 2
	OBJ_BMP_256_1D = 3
)

func (e *Engine) UpdateObjMapping(d *Dispcnt) {

	for i := range e.Objects {

		obj := &e.Objects[i]

		obj.ObjTileMapping = 0
		obj.TileBoundaryShift = 0
		obj.ObjBmpMapping = 0
		obj.BmpBoundaryShift = 0
		obj.BmpBoundaryMask = 0

		switch {
		case !d.TileObj1D:
			obj.ObjTileMapping = OBJ_TIL_STD_2D
			obj.TileBoundaryShift = 5 // 32
		case d.TileObj1D && d.TileObjBoundary == 0:
			obj.ObjTileMapping = OBJ_TIL_STD_1D
			obj.TileBoundaryShift = 5
		case d.TileObj1D && d.TileObjBoundary == 1:
			obj.ObjTileMapping = OBJ_TIL_064_1D
			obj.TileBoundaryShift = 6 // 64
		case d.TileObj1D && d.TileObjBoundary == 2:
			obj.ObjTileMapping = OBJ_TIL_128_1D
			obj.TileBoundaryShift = 7 // 128
		case d.TileObj1D && d.TileObjBoundary == 3:
			obj.ObjTileMapping = OBJ_TIL_256_1D
			obj.TileBoundaryShift = 8 // 256
		}

		switch {
		case !d.BitmapObj1D && !d.BitmapObj256:
			obj.ObjBmpMapping = OBJ_BMP_128_2D
			obj.BmpBoundaryMask = 0x0F
		case !d.BitmapObj1D && d.BitmapObj256:
			obj.ObjBmpMapping = OBJ_BMP_256_2D
			obj.BmpBoundaryMask = 0x1F
		case d.BitmapObj1D && !d.BitmapObj256 && !d.BitmapObjBoundary:
			obj.ObjBmpMapping = OBJ_BMP_128_1D
			obj.BmpBoundaryShift = 7 //128
		case d.BitmapObj1D && !d.BitmapObj256 && d.BitmapObjBoundary:
			obj.ObjBmpMapping = OBJ_BMP_256_1D
			obj.BmpBoundaryShift = 8 //256
		case d.BitmapObj1D && d.BitmapObj256:
			panic("DISPCNT HAS BOTH BITMAP 1D AND 256 SET")
		}
	}
}

func (e *Engine) setBgType(bgIdx uint32) {

    bg := &e.Backgrounds[bgIdx]

    switch bgIdx {
    case 0:
        if !e.IsB && e.Dispcnt.Is3D {
            bg.Type = BG_TYPE_3D
            return
        }

    case 2:
        switch e.Dispcnt.Mode {
        case 2, 4:
            bg.Type = BG_TYPE_AFF
            return
        case 5:
            switch {
            case !bg.Palette256:
                bg.Type = BG_TYPE_BGM
            case (bg.CharBaseBlock>>14)&1 == 1:
                bg.Type = BG_TYPE_DIR
            default:
                bg.Type = BG_TYPE_256
            }
            return
        case 6:
            bg.Type = BG_TYPE_LAR
            return
        }
    case 3:
        switch e.Dispcnt.Mode {
        case 0:
            bg.Type = BG_TYPE_TEX
        case 1, 2:
            bg.Type = BG_TYPE_AFF
        default:
            switch {
            case !bg.Palette256:
                bg.Type = BG_TYPE_BGM
            case (bg.CharBaseBlock>>14)&1 == 1:
                bg.Type = BG_TYPE_DIR
            default:
                bg.Type = BG_TYPE_256
            }
        }
        return
    }

    bg.Type = BG_TYPE_TEX
}
