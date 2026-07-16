package gba

import (
	"encoding/binary"
)

type PPU struct {
	gba *GBA

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
	CGB                bool
	DisplayFrame1      bool
	HBlankIntervalFree bool
	OneDimensional     bool
	ForcedBlank        bool
	DisplayObj         bool
	DisplayWin0        bool
	DisplayWin1        bool
	DisplayObjWin      bool
}

// blends are [6]... because Bg0, Bg1, Bg2, Bg3, Obj, Bd
type Blend struct {
	aEv, bEv, yEv float32
	Mode          uint32
	a, b          [6]bool
}

type Windows struct {
	Win0, Win1, WinObj Window
	OutBg              [4]bool
	OutObj, OutBld     bool
	Enabled            bool
}

type Window struct {
	L, R, T, B     uint32
	oL, oR, oT, oB uint32
	InBg           [4]bool
	InObj, InBld   bool
	Enabled        bool
}

type Mosaic struct {
	BgH, BgV, ObjH, ObjV uint32
}

type Background struct {
	OutX, OutY         float32
	W, H               uint32
	Pa, Pb, Pc, Pd     uint32
	Priority           uint32
	CharBaseBlock      uint32
	ScreenBaseBlock    uint32
	Size               uint32
	XOffset, YOffset   uint32
	aXOffset, aYOffset uint32
	AffineWrap         bool
	Enabled            bool
	Invalid            bool
	Mosaic             bool
	Palette256         bool
	Affine             bool
}

type Object struct {
	Pa, Pb, Pc, Pd float32
	X, Y, W, H     uint32
	CharName       uint32
	Palette        uint32
	Mode           uint16
	RotParams      uint16
	Priority       uint16
	Shape          uint16
	Size           uint16
	RotScale       bool
	DoubleSize     bool
	Disable        bool
	Mosaic         bool
	Palette256     bool
	HFlip, VFlip   bool
	OneDimensional bool
}

func (p *PPU) UpdatePPU(addr, v uint32) {
	if win := addr >= 0x40 && addr < 0x4C; win {
		p.UpdateWin(addr, v)
		return
	}

	if bgs := addr >= 0x08 && addr < 0x40; bgs {
		p.UpdateBackgrounds(addr, v)
		return
	}

	switch addr {
	case 0x0:
		p.Dispcnt.Mode = v & 0b111
		p.Dispcnt.CGB = (v>>3)&1 != 0
		p.Dispcnt.DisplayFrame1 = (v>>4)&1 != 0
		p.Dispcnt.HBlankIntervalFree = (v>>5)&1 != 0
		p.Dispcnt.OneDimensional = (v>>6)&1 != 0
		p.Dispcnt.ForcedBlank = (v>>7)&1 != 0

	case 0x1:
		p.Dispcnt.DisplayObj = (v>>4)&1 != 0
		p.Dispcnt.DisplayWin0 = (v>>5)&1 != 0
		p.Dispcnt.DisplayWin1 = (v>>6)&1 != 0
		p.Dispcnt.DisplayObjWin = (v>>7)&1 != 0

		p.Backgrounds[0].Enabled = (v>>0)&1 != 0
		p.Backgrounds[1].Enabled = (v>>1)&1 != 0
		p.Backgrounds[2].Enabled = (v>>2)&1 != 0
		p.Backgrounds[3].Enabled = (v>>3)&1 != 0

		wins := &p.Windows
		wins.Win0.Enabled = p.Dispcnt.DisplayWin0
		wins.Win1.Enabled = p.Dispcnt.DisplayWin1
		wins.WinObj.Enabled = p.Dispcnt.DisplayObjWin && p.Dispcnt.DisplayObj
		wins.Enabled = wins.Win0.Enabled || wins.Win1.Enabled || wins.WinObj.Enabled

	case 0x4C:

		p.Mosaic.BgH = (v >> 0) & 0xF
		p.Mosaic.BgV = (v >> 4) & 0xF

	case 0x4D:

		p.Mosaic.ObjH = (v >> 0) & 0xF
		p.Mosaic.ObjV = (v >> 4) & 0xF

	case 0x50:
		p.Blend.a[0] = (v>>0)&1 != 0
		p.Blend.a[1] = (v>>1)&1 != 0
		p.Blend.a[2] = (v>>2)&1 != 0
		p.Blend.a[3] = (v>>3)&1 != 0
		p.Blend.a[4] = (v>>4)&1 != 0
		p.Blend.a[5] = (v>>5)&1 != 0
		p.Blend.Mode = (v >> 6) & 0b11

	case 0x51:
		p.Blend.b[0] = (v>>0)&1 != 0
		p.Blend.b[1] = (v>>1)&1 != 0
		p.Blend.b[2] = (v>>2)&1 != 0
		p.Blend.b[3] = (v>>3)&1 != 0
		p.Blend.b[4] = (v>>4)&1 != 0
		p.Blend.b[5] = (v>>5)&1 != 0

	case 0x52:
		p.Blend.aEv = float32(min(16, v&0x1F)) / 16

	case 0x53:
		p.Blend.bEv = float32(min(16, v&0x1F)) / 16

	case 0x54:
		p.Blend.yEv = float32(min(16, v&0x1F)) / 16

	}
}

func (p *PPU) UpdateWin(addr, v uint32) {
	wins := &p.Windows
	win0 := &p.Windows.Win0
	win1 := &p.Windows.Win1
	winObj := &p.Windows.WinObj

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

func (p *PPU) UpdateOAM(relAddr uint32, v uint16) {
	switch idx := relAddr & 7; idx {
	case 0:
		obj := &p.Objects[relAddr>>3]
		attr := uint32(v)

		obj.Y = attr & 0xFF
		obj.RotScale = (attr>>8)&1 != 0

		if obj.RotScale {
			obj.DoubleSize = (attr>>9)&1 != 0
			UpdateAffineParams(obj, &p.gba.Mem.OAM)
		} else {
			obj.Disable = (attr>>9)&1 != 0
		}

		obj.Mode = (v >> 10) & 3
		obj.Mosaic = (attr>>12)&1 != 0
		obj.Palette256 = (attr>>13)&1 != 0

		obj.Shape = v >> 14
		obj.W = objSize[obj.Shape][obj.Size][0]
		obj.H = objSize[obj.Shape][obj.Size][1]

	case 2:
		obj := &p.Objects[relAddr>>3]
		attr := uint32(v)

		obj.X = attr & 0x1FF

		if obj.RotScale {
			obj.RotParams = (v >> 9) & 0x1F
			UpdateAffineParams(obj, &p.gba.Mem.OAM)
		} else {
			obj.HFlip = (attr>>12)&1 != 0
			obj.VFlip = (attr>>13)&1 != 0
		}

		obj.Size = v >> 14
		obj.W = objSize[obj.Shape][obj.Size][0]
		obj.H = objSize[obj.Shape][obj.Size][1]
	case 4:
		obj := &p.Objects[relAddr>>3]
		attr := uint32(v)

		obj.CharName = attr & 0x3FF
		obj.Priority = (v >> 10) & 3
		obj.Palette = attr >> 12

	case 6:
		paramIdx := uint16(relAddr / 0x20)

		v := float32(int16(v)) / 256

		for i := range 128 {

			obj := &p.Objects[i]

			if obj.RotScale && obj.RotParams == paramIdx {
				switch relAddr & 0x1F {
				case 0x06:
					obj.Pa = v
				case 0x0E:
					obj.Pb = v
				case 0x16:
					obj.Pc = v
				case 0x1E:
					obj.Pd = v
				}
				continue
			}
		}
	}
}

func UpdateAffineParams(obj *Object, oam *[0x400]uint8) {
	paramsAddr := obj.RotParams * 0x20
	obj.Pa = float32(int16(binary.LittleEndian.Uint16(oam[paramsAddr+0x06:]))) / 256
	obj.Pb = float32(int16(binary.LittleEndian.Uint16(oam[paramsAddr+0x0E:]))) / 256
	obj.Pc = float32(int16(binary.LittleEndian.Uint16(oam[paramsAddr+0x16:]))) / 256
	obj.Pd = float32(int16(binary.LittleEndian.Uint16(oam[paramsAddr+0x1E:]))) / 256
}

func (p *PPU) UpdateBackgrounds(addr, v uint32) {
	switch addr {
	case 0x08, 0xA, 0xC, 0xE:
		p.Backgrounds[(addr>>1)&3].Priority = v & 3
		p.Backgrounds[(addr>>1)&3].CharBaseBlock = ((v >> 2) & 0xF) * 0x4000
		p.Backgrounds[(addr>>1)&3].Mosaic = (v>>6)&1 != 0
		p.Backgrounds[(addr>>1)&3].Palette256 = (v>>7)&1 != 0
	case 0x09, 0xB, 0xD, 0xF:
		p.Backgrounds[(addr>>1)&3].ScreenBaseBlock = (v & 0x1F) * 0x800
		p.Backgrounds[(addr>>1)&3].AffineWrap = (v>>5)&1 != 0
		p.Backgrounds[(addr>>1)&3].Size = (v >> 6) & 3

	case 0x10, 0x14, 0x18, 0x1C, 0x11, 0x15, 0x19, 0x1D:
		p.Backgrounds[(addr>>2)&3].XOffset &^= 0xFF << ((addr & 1) << 3)
		p.Backgrounds[(addr>>2)&3].XOffset |= v << ((addr & 1) << 3)
	case 0x12, 0x16, 0x1A, 0x1E, 0x13, 0x17, 0x1B, 0x1F:
		p.Backgrounds[(addr>>2)&3].YOffset &^= 0xFF << ((addr & 1) << 3)
		p.Backgrounds[(addr>>2)&3].YOffset |= v << ((addr & 1) << 3)

	case 0x20, 0x21, 0x30, 0x31:
		p.Backgrounds[(addr>>4)&3].Pa &^= 0xFF << ((addr & 1) << 3)
		p.Backgrounds[(addr>>4)&3].Pa |= v << ((addr & 1) << 3)
	case 0x22, 0x23, 0x32, 0x33:
		p.Backgrounds[(addr>>4)&3].Pb &^= 0xFF << ((addr & 1) << 3)
		p.Backgrounds[(addr>>4)&3].Pb |= v << ((addr & 1) << 3)
	case 0x24, 0x25, 0x34, 0x35:
		p.Backgrounds[(addr>>4)&3].Pc &^= 0xFF << ((addr & 1) << 3)
		p.Backgrounds[(addr>>4)&3].Pc |= v << ((addr & 1) << 3)
	case 0x26, 0x27, 0x36, 0x37:
		p.Backgrounds[(addr>>4)&3].Pd &^= 0xFF << ((addr & 1) << 3)
		p.Backgrounds[(addr>>4)&3].Pd |= v << ((addr & 1) << 3)

	case 0x28, 0x29, 0x2A, 0x2B, 0x38, 0x39, 0x3A, 0x3B:
		p.Backgrounds[(addr>>4)&3].aXOffset &^= 0xFF << ((addr & 3) << 3)
		p.Backgrounds[(addr>>4)&3].aXOffset |= v << ((addr & 3) << 3)
		p.Backgrounds[(addr>>4)&3].BgAffineReset()

	case 0x2C, 0x2D, 0x2E, 0x2F, 0x3C, 0x3D, 0x3E, 0x3F:
		p.Backgrounds[(addr>>4)&3].aYOffset &^= 0xFF << ((addr & 3) << 3)
		p.Backgrounds[(addr>>4)&3].aYOffset |= v << ((addr & 3) << 3)
		p.Backgrounds[(addr>>4)&3].BgAffineReset()
	}
}

// w, h

var bgSize = [2][4][2]uint32{
	// std
	{{256, 256}, {512, 256}, {256, 512}, {512, 512}},
	// affine
	{{128, 128}, {256, 256}, {512, 512}, {1024, 1024}},
}

var objSize = [4][4][2]uint32{
	// SQUARE
	{{8, 8}, {16, 16}, {32, 32}, {64, 64}},
	// HORIZONTAL
	{{16, 8}, {32, 8}, {32, 16}, {64, 32}},
	// VERTICAL
	{{8, 16}, {8, 32}, {16, 32}, {32, 64}},
	// PROHIBITED
	{{8, 8}, {16, 16}, {32, 32}, {64, 64}},
}
