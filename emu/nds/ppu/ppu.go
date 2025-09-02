package ppu

import (
	"github.com/aabalke/guac/emu/nds/utils"
    "encoding/binary"
)

type PPU struct {
    EngineA Engine
    EngineB Engine
    PowCnt1 PowCnt1
}

type Engine struct {

    Pixels *[]byte
    IsB bool

	Dispcnt Dispcnt

	Objects     [128]Object
	Backgrounds [4]Background
	Windows     Windows
	Blend       Blend
	Mosaic      Mosaic

	BgPriorities  [4][]uint32
	ObjPriorities [4][]uint32

}

type Dispcnt struct {
	Mode               uint32
	Is3D               bool
    TileObj1D bool
    BitmapObj256 bool
    BitmapObj1D bool
	ForcedBlank        bool
	DisplayObj    bool
	DisplayWin0   bool
	DisplayWin1   bool
	DisplayObjWin bool
    DisplayMode uint32
    VramBlock uint32
    TileObjBoundary uint32
    BitmapObjBoundary bool
	HBlankIntervalFree bool
    CharBase uint32
    ScreenBase uint32
    BgExtPal bool
    ObjExtPal bool
}

// blends are [6]... because Bg0, Bg1, Bg2, Bg3, Obj, Bd
type Blend struct {
	Mode          uint32
	a, b          [6]bool
	aEv, bEv, yEv float32
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
	Enabled            bool
	Invalid            bool
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
    ObjBmpMapping uint8

    TileBoundaryShift uint32
    BmpBoundaryShift uint32
}

type PowCnt1 struct {
    V uint16
    LcdEnabled bool
    EngineA2D, EngineB2D bool
    RenderingEngine, GeometryEngine bool
    TopA bool
}

func (p *PPU) Update(addr, v uint32) {

    if engineA := addr < 0x60; engineA {
        p.EngineA.UpdateEngine(addr, v)
        return
    } else if engineB := addr >= 0x1000 && addr < 0x1070; engineB {
        p.EngineB.UpdateEngine(addr & 0xFF, v)
        return
    }

    switch addr {
    case 0x304:

        p.PowCnt1.LcdEnabled      = utils.BitEnabled(v, 0)
        p.PowCnt1.EngineA2D       = utils.BitEnabled(v, 1)
        p.PowCnt1.RenderingEngine = utils.BitEnabled(v, 2)
        p.PowCnt1.GeometryEngine  = utils.BitEnabled(v, 3)

    case 0x305:
        p.PowCnt1.EngineB2D = utils.BitEnabled(v, 1)

        prevTopA := p.PowCnt1.TopA
        p.PowCnt1.TopA = utils.BitEnabled(v, 7)

        if prevTopA != p.PowCnt1.TopA {
            a := p.EngineA.Pixels
            b := p.EngineB.Pixels
            p.EngineA.Pixels = b
            p.EngineB.Pixels = a
        }
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
	    e.Dispcnt.Mode = utils.GetVarData(v, 0, 2)
        e.Dispcnt.Is3D = utils.BitEnabled(v, 3)
        e.Dispcnt.TileObj1D = utils.BitEnabled(v, 4)
        e.Dispcnt.BitmapObj256 = utils.BitEnabled(v, 5)
        e.Dispcnt.BitmapObj1D = utils.BitEnabled(v, 6)
        e.Dispcnt.ForcedBlank = utils.BitEnabled(v, 7)

        e.UpdateObjMapping(&e.Dispcnt)

	case 0x1:
		e.Dispcnt.DisplayObj = utils.BitEnabled(v, 4)
		e.Dispcnt.DisplayWin0 = utils.BitEnabled(v, 5)
		e.Dispcnt.DisplayWin1 = utils.BitEnabled(v, 6)
		e.Dispcnt.DisplayObjWin = utils.BitEnabled(v, 7)

		e.Backgrounds[0].Enabled = utils.BitEnabled(v, 0)
		e.Backgrounds[1].Enabled = utils.BitEnabled(v, 1)
		e.Backgrounds[2].Enabled = utils.BitEnabled(v, 2)
		e.Backgrounds[3].Enabled = utils.BitEnabled(v, 3)

		wins := &e.Windows
		wins.Win0.Enabled = e.Dispcnt.DisplayWin0
		wins.Win1.Enabled = e.Dispcnt.DisplayWin1
		wins.WinObj.Enabled = e.Dispcnt.DisplayObjWin && e.Dispcnt.DisplayObj
		wins.Enabled = wins.Win0.Enabled || wins.Win1.Enabled || wins.WinObj.Enabled
        e.UpdateObjMapping(&e.Dispcnt)

	case 0x2:

        e.Dispcnt.DisplayMode = utils.GetVarData(v, 0, 1)
        e.Dispcnt.VramBlock = utils.GetVarData(v, 2, 3)
        e.Dispcnt.TileObjBoundary = utils.GetVarData(v, 4, 5)
		e.Dispcnt.BitmapObjBoundary = utils.BitEnabled(v, 6)
		e.Dispcnt.HBlankIntervalFree = utils.BitEnabled(v, 7)
        e.UpdateObjMapping(&e.Dispcnt)

	case 0x3:

        e.Dispcnt.CharBase = utils.GetVarData(v, 0, 2) * 0x1_0000
        e.Dispcnt.ScreenBase = utils.GetVarData(v, 3, 5) * 0x1_0000
        e.Dispcnt.BgExtPal = utils.BitEnabled(v, 6)
        e.Dispcnt.ObjExtPal = utils.BitEnabled(v, 7)
        e.UpdateObjMapping(&e.Dispcnt)

	case 0x4C:

		e.Mosaic.BgH = utils.GetVarData(v, 0, 3)
		e.Mosaic.BgV = utils.GetVarData(v, 4, 7)

	case 0x4D:

		e.Mosaic.ObjH = utils.GetVarData(v, 0, 3)
		e.Mosaic.ObjV = utils.GetVarData(v, 4, 7)

	case 0x50:
		e.Blend.a[0] = utils.BitEnabled(v, 0)
		e.Blend.a[1] = utils.BitEnabled(v, 1)
		e.Blend.a[2] = utils.BitEnabled(v, 2)
		e.Blend.a[3] = utils.BitEnabled(v, 3)
		e.Blend.a[4] = utils.BitEnabled(v, 4)
		e.Blend.a[5] = utils.BitEnabled(v, 5)
		e.Blend.Mode = utils.GetVarData(v, 6, 7)

	case 0x51:
		e.Blend.b[0] = utils.BitEnabled(v, 0)
		e.Blend.b[1] = utils.BitEnabled(v, 1)
		e.Blend.b[2] = utils.BitEnabled(v, 2)
		e.Blend.b[3] = utils.BitEnabled(v, 3)
		e.Blend.b[4] = utils.BitEnabled(v, 4)
		e.Blend.b[5] = utils.BitEnabled(v, 5)

	case 0x52:
		e.Blend.aEv = float32(min(16, utils.GetVarData(v, 0, 4))) / 16

	case 0x53:
		e.Blend.bEv = float32(min(16, utils.GetVarData(v, 0, 4))) / 16

	case 0x54:
		e.Blend.yEv = float32(min(16, utils.GetVarData(v, 0, 4))) / 16
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

        SCREEN_WIDTH = 256
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
		win0.InBg[0] = utils.BitEnabled(v, 0)
		win0.InBg[1] = utils.BitEnabled(v, 1)
		win0.InBg[2] = utils.BitEnabled(v, 2)
		win0.InBg[3] = utils.BitEnabled(v, 3)
		win0.InObj = utils.BitEnabled(v, 4)
		win0.InBld = utils.BitEnabled(v, 5)
	case WININ1:
		win1.InBg[0] = utils.BitEnabled(v, 0)
		win1.InBg[1] = utils.BitEnabled(v, 1)
		win1.InBg[2] = utils.BitEnabled(v, 2)
		win1.InBg[3] = utils.BitEnabled(v, 3)
		win1.InObj = utils.BitEnabled(v, 4)
		win1.InBld = utils.BitEnabled(v, 5)
	case WINOUT:
		wins.OutBg[0] = utils.BitEnabled(v, 0)
		wins.OutBg[1] = utils.BitEnabled(v, 1)
		wins.OutBg[2] = utils.BitEnabled(v, 2)
		wins.OutBg[3] = utils.BitEnabled(v, 3)
		wins.OutObj = utils.BitEnabled(v, 4)
		wins.OutBld = utils.BitEnabled(v, 5)
	case WINOBJ:
		winObj.InBg[0] = utils.BitEnabled(v, 0)
		winObj.InBg[1] = utils.BitEnabled(v, 1)
		winObj.InBg[2] = utils.BitEnabled(v, 2)
		winObj.InBg[3] = utils.BitEnabled(v, 3)
		winObj.InObj = utils.BitEnabled(v, 4)
		winObj.InBld = utils.BitEnabled(v, 5)
	}
}

func (p *Engine) UpdateBackgrounds(addr, v uint32) {

	switch addr {
	case 0x08:
		p.Backgrounds[0].Priority = utils.GetVarData(v, 0, 1)
		p.Backgrounds[0].CharBaseBlock = utils.GetVarData(v, 2, 5) * 0x4000
		p.Backgrounds[0].Mosaic = utils.BitEnabled(v, 6)
		p.Backgrounds[0].Palette256 = utils.BitEnabled(v, 7)
	case 0x09:
		p.Backgrounds[0].ScreenBaseBlock = utils.GetVarData(v, 0, 4) * 0x800
        p.Backgrounds[0].AltExtPalSlot = utils.BitEnabled(v, 5)
		p.Backgrounds[0].Size = utils.GetVarData(v, 6, 7)

	case 0x0A:
		p.Backgrounds[1].Priority = utils.GetVarData(v, 0, 1)
		p.Backgrounds[1].CharBaseBlock = utils.GetVarData(v, 2, 5) * 0x4000
		p.Backgrounds[1].Mosaic = utils.BitEnabled(v, 6)
		p.Backgrounds[1].Palette256 = utils.BitEnabled(v, 7)
	case 0x0B:
		p.Backgrounds[1].ScreenBaseBlock = utils.GetVarData(v, 0, 4) * 0x800
        p.Backgrounds[1].AltExtPalSlot = utils.BitEnabled(v, 5)
		p.Backgrounds[1].Size = utils.GetVarData(v, 6, 7)

	case 0x0C:
		p.Backgrounds[2].Priority = utils.GetVarData(v, 0, 1)
		p.Backgrounds[2].CharBaseBlock = utils.GetVarData(v, 2, 5) * 0x4000
		p.Backgrounds[2].Mosaic = utils.BitEnabled(v, 6)
		p.Backgrounds[2].Palette256 = utils.BitEnabled(v, 7)
	case 0x0D:
		p.Backgrounds[2].ScreenBaseBlock = utils.GetVarData(v, 0, 4) * 0x800
        p.Backgrounds[2].AffineWrap = utils.BitEnabled(v, 5)
		p.Backgrounds[2].Size = utils.GetVarData(v, 6, 7)

	case 0x0E:
		p.Backgrounds[3].Priority = utils.GetVarData(v, 0, 1)
		p.Backgrounds[3].CharBaseBlock = utils.GetVarData(v, 2, 5) * 0x4000
		p.Backgrounds[3].Mosaic = utils.BitEnabled(v, 6)
		p.Backgrounds[3].Palette256 = utils.BitEnabled(v, 7)

	case 0x0F:
		p.Backgrounds[3].ScreenBaseBlock = utils.GetVarData(v, 0, 4) * 0x800
        p.Backgrounds[3].AffineWrap = utils.BitEnabled(v, 5)
		p.Backgrounds[3].Size = utils.GetVarData(v, 6, 7)

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
    case BG_TYPE_AFF:
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
    default:
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
		case 1:
			obj.H, obj.W = 16, 16
		case 2:
			obj.H, obj.W = 32, 32
		case 3:
			obj.H, obj.W = 64, 64
		}
	case HORIZONTAL:
		switch size {
		case 0:
			obj.H, obj.W = 8, 16
		case 1:
			obj.H, obj.W = 8, 32
		case 2:
			obj.H, obj.W = 16, 32
		case 3:
			obj.H, obj.W = 32, 64
		}
	case VERTICAL:
		switch size {
		case 0:
			obj.H, obj.W = 16, 8
		case 1:
			obj.H, obj.W = 32, 8
		case 2:
			obj.H, obj.W = 32, 16
		case 3:
			obj.H, obj.W = 64, 32
		}
	}
}

func (bg *Background) BgAffineReset() {
	bg.OutX = utils.Convert20_8Float(int32(bg.aXOffset))
	bg.OutY = utils.Convert20_8Float(int32(bg.aYOffset))
}

func (bg *Background) BgAffineUpdate() {
	bg.OutX += utils.Convert8_8Float(int16(bg.Pb))
	bg.OutY += utils.Convert8_8Float(int16(bg.Pd))
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
		obj.Y = attr & 0b1111_1111
	case 1:

		obj.RotScale = utils.BitEnabled(attr, 0)
		obj.Mode = utils.GetVarData(attr, 2, 3)
		obj.Mosaic = utils.BitEnabled(attr, 4)
		obj.Palette256 = utils.BitEnabled(attr, 5)
		obj.Shape = utils.GetVarData(attr, 6, 7)
		obj.setSize(obj.Shape, obj.Size)

		if obj.RotScale {
			obj.DoubleSize = utils.BitEnabled(attr, 1)
			UpdateAffineParams(obj, oam, engine.IsB)
		} else {
			obj.Disable = utils.BitEnabled(attr, 1)
		}

	case 2:
		obj.X &^= 0xFF
		obj.X |= attr
	case 3:
		obj.X &= 0xFF
		obj.X |= (attr & 0b1) << 8
		obj.Size = utils.GetVarData(attr, 6, 7)
		obj.setSize(obj.Shape, obj.Size)

		if obj.RotScale {
			obj.RotParams = utils.GetVarData(attr, 1, 5)
			UpdateAffineParams(obj, oam, engine.IsB)
		}
		obj.HFlip = utils.BitEnabled(attr, 4)
		obj.VFlip = utils.BitEnabled(attr, 5)
	case 4:
		obj.CharName &^= 0xFF
		obj.CharName |= attr
	case 5:
		obj.CharName &= 0xFF
		obj.CharName |= (attr & 0b11) << 8
		obj.Priority = utils.GetVarData(attr, 2, 3)
		obj.Palette = utils.GetVarData(attr, 4, 7)
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

func (obj *Object) setSize(shape, size uint32) {

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
		case 1:
			obj.H, obj.W = 16, 16
		case 2:
			obj.H, obj.W = 32, 32
		case 3:
			obj.H, obj.W = 64, 64
		}
	case HORIZONTAL:
		switch size {
		case 0:
			obj.H, obj.W = 8, 16
		case 1:
			obj.H, obj.W = 8, 32
		case 2:
			obj.H, obj.W = 16, 32
		case 3:
			obj.H, obj.W = 32, 64
		}
	case VERTICAL:
		switch size {
		case 0:
			obj.H, obj.W = 16, 8
		case 1:
			obj.H, obj.W = 32, 8
		case 2:
			obj.H, obj.W = 32, 16
		case 3:
			obj.H, obj.W = 64, 32
		}
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
            //panic("NEED TO SET UP 2D BITMAP OBJ")
        case !d.BitmapObj1D && d.BitmapObj256:
            obj.ObjBmpMapping = OBJ_BMP_256_2D
            //panic("NEED TO SET UP 2D BITMAP OBJ")
        case d.BitmapObj1D && !d.BitmapObj256 && !d.BitmapObjBoundary:
            obj.ObjBmpMapping = OBJ_BMP_128_1D
            obj.TileBoundaryShift = 7 // 256
        case d.BitmapObj1D && !d.BitmapObj256 && d.BitmapObjBoundary:
            obj.ObjBmpMapping = OBJ_BMP_256_1D
            obj.TileBoundaryShift = 8 // 256
        case d.BitmapObj1D && d.BitmapObj256:
            panic("DISPCNT HAS BOTH BITMAP 1D AND 256 SET")
        }
    }
}
