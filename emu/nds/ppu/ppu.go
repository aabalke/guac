package ppu

import (
	"github.com/aabalke/guac/emu/nds/utils"
)

type PPU struct {
    EngineA Engine
    EngineB Engine
    PowCnt1 PowCnt1
}

type Engine struct {

    Pixels *[]byte

	Dispcnt Dispcnt

	Objects     [128]Object
	Backgrounds [4]Background
	Windows     Windows
	Blend       Blend
	Mosaic      Mosaic

	bgPriorities  [4][]uint32
	objPriorities [4][]uint32

}

type Dispcnt struct {
	Mode               uint32
	is3D               bool
	DisplayFrame1      bool
	HBlankIntervalFree bool
	OneDimensional     bool
	ForcedBlank        bool
	//DisplayBg [4]bool
	DisplayObj    bool
	DisplayWin0   bool
	DisplayWin1   bool
	DisplayObjWin bool

    DisplayMode uint32
    VramBlock uint32
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
}

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

	if bgs := addr >= 0x08 && addr < 0x40; bgs {
		e.UpdateBackgrounds(addr, v)
		return
	}

    switch addr {
	case 0x0:
	    e.Dispcnt.Mode = utils.GetVarData(v, 0, 2)

	case 0x1:
		e.Backgrounds[0].Enabled = utils.BitEnabled(v, 0)
		e.Backgrounds[1].Enabled = utils.BitEnabled(v, 1)
		e.Backgrounds[2].Enabled = utils.BitEnabled(v, 2)
		e.Backgrounds[3].Enabled = utils.BitEnabled(v, 3)
	case 0x2:
        e.Dispcnt.DisplayMode = utils.GetVarData(v, 0, 1)
        e.Dispcnt.VramBlock = utils.GetVarData(v, 2, 3)

	case 0x3:
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
		p.Backgrounds[0].AffineWrap = utils.BitEnabled(v, 5)
		p.Backgrounds[0].Size = utils.GetVarData(v, 6, 7)

	case 0x0A:
		p.Backgrounds[1].Priority = utils.GetVarData(v, 0, 1)
		p.Backgrounds[1].CharBaseBlock = utils.GetVarData(v, 2, 5) * 0x4000
		p.Backgrounds[1].Mosaic = utils.BitEnabled(v, 6)
		p.Backgrounds[1].Palette256 = utils.BitEnabled(v, 7)
	case 0x0B:
		p.Backgrounds[1].ScreenBaseBlock = utils.GetVarData(v, 0, 4) * 0x800
		p.Backgrounds[1].AffineWrap = utils.BitEnabled(v, 5)
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
		//p.Backgrounds[2].BgAffineReset()
	case 0x29:
		p.Backgrounds[2].aXOffset &^= 0xFF00
		p.Backgrounds[2].aXOffset |= v << 8
		//p.Backgrounds[2].BgAffineReset()
	case 0x2A:
		p.Backgrounds[2].aXOffset &^= 0xFF0000
		p.Backgrounds[2].aXOffset |= v << 16
		//p.Backgrounds[2].BgAffineReset()
	case 0x2B:
		p.Backgrounds[2].aXOffset &^= 0xFF000000
		p.Backgrounds[2].aXOffset |= v << 24
		//p.Backgrounds[2].BgAffineReset()

	case 0x2C:
		p.Backgrounds[2].aYOffset &^= 0xFF
		p.Backgrounds[2].aYOffset |= v
		//p.Backgrounds[2].BgAffineReset()
	case 0x2D:
		p.Backgrounds[2].aYOffset &^= 0xFF00
		p.Backgrounds[2].aYOffset |= v << 8
		//p.Backgrounds[2].BgAffineReset()
	case 0x2E:
		p.Backgrounds[2].aYOffset &^= 0xFF0000
		p.Backgrounds[2].aYOffset |= v << 16
		//p.Backgrounds[2].BgAffineReset()
	case 0x2F:
		p.Backgrounds[2].aYOffset &^= 0xFF000000
		p.Backgrounds[2].aYOffset |= v << 24
		//p.Backgrounds[2].BgAffineReset()

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
		//p.Backgrounds[3].BgAffineReset()
	case 0x39:
		p.Backgrounds[3].aXOffset &^= 0xFF00
		p.Backgrounds[3].aXOffset |= v << 8
		//p.Backgrounds[3].BgAffineReset()
	case 0x3A:
		p.Backgrounds[3].aXOffset &^= 0xFF0000
		p.Backgrounds[3].aXOffset |= v << 16
		//p.Backgrounds[3].BgAffineReset()
	case 0x3B:
		p.Backgrounds[3].aXOffset &^= 0xFF000000
		p.Backgrounds[3].aXOffset |= v << 24
		//p.Backgrounds[3].BgAffineReset()

	case 0x3C:
		p.Backgrounds[3].aYOffset &^= 0xFF
		p.Backgrounds[3].aYOffset |= v
		//p.Backgrounds[3].BgAffineReset()
	case 0x3D:
		p.Backgrounds[3].aYOffset &^= 0xFF00
		p.Backgrounds[3].aYOffset |= v << 8
		//p.Backgrounds[3].BgAffineReset()
	case 0x3E:
		p.Backgrounds[3].aYOffset &^= 0xFF0000
		p.Backgrounds[3].aYOffset |= v << 16
		//p.Backgrounds[3].BgAffineReset()
	case 0x3F:
		p.Backgrounds[3].aYOffset &^= 0xFF000000
		p.Backgrounds[3].aYOffset |= v << 24
		//p.Backgrounds[3].BgAffineReset()
	}
}

func (bg *Background) SetSize() {

	if bg.Affine {
		switch bg.Size {
		case 0:
			//bg.W, bg.H = 16, 16
			bg.W, bg.H = 128, 128
		case 1:
			//bg.W, bg.H = 32, 32
			bg.W, bg.H = 256, 256
		case 2:
			//bg.W, bg.H = 64, 64
			bg.W, bg.H = 512, 512
		case 3:
			//bg.W, bg.H = 128, 128
			bg.W, bg.H = 1024, 1024
		default:
			panic("PROHIBITTED AFFINE BG SIZE")
		}

		return
	}

	switch bg.Size {
	case 0:
		//bg.W, bg.H = 32, 32
		bg.W, bg.H = 256, 256
	case 1:
		//bg.W, bg.H = 64, 32
		bg.W, bg.H = 512, 256
	case 2:
		//bg.W, bg.H = 32, 64
		bg.W, bg.H = 256, 512
	case 3:
		//bg.W, bg.H = 64, 64
		bg.W, bg.H = 512, 512
	default:
		panic("PROHIBITTED BG SIZE")
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
