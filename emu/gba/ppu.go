package gba

import (

	"github.com/aabalke33/guac/emu/gba/utils"
)

type PPU struct {

    gba *GBA

    Dispcnt Dispcnt

    Objects [128]Object
    Backgrounds [4]Background
    Windows Windows
    Blend Blend
}

type Dispcnt struct {
    Mode uint32
    CGB bool
    DisplayFrame1 bool
    HBlankIntervalFree bool
    OneDimensional bool
    ForcedBlank bool
    DisplayBg0 bool
    DisplayBg1 bool
    DisplayBg2 bool
    DisplayBg3 bool
    DisplayObj bool
    DisplayWin0 bool
    DisplayWin1 bool
    DisplayObjWin bool
}

// blends are [6]... because Bg0, Bg1, Bg2, Bg3, Obj, Bd
type Blend struct {
    Mode uint32
    a, b [6]bool
    aEv, bEv, yEv uint32
}

type Windows struct {
    Enabled bool
    Win0, Win1, WinObj Window
    OutBg0, OutBg1, OutBg2, OutBg3, OutObj, OutBld bool
}

type Window struct {
    Enabled bool
    L, R, T, B uint32
    oL, oR, oT, oB uint32
    InBg0, InBg1, InBg2, InBg3, InObj, InBld bool
}

type Background struct {
    Enabled bool
    Invalid bool
    W, H uint32
    Pa, Pb, Pc, Pd uint32
    Priority uint32
    CharBaseBlock uint32
    Mosaic bool
    Palette256 bool
    ScreenBaseBlock uint32
    AffineWrap bool
    Size uint32
    XOffset, YOffset uint32
    aXOffset, aYOffset uint32
    Affine bool
}

type Object struct {
    X, Y, W, H uint32
    Pa, Pb, Pc, Pd float32
    RotScale bool
    DoubleSize bool
    Disable bool
    Mode uint32
    Mosaic bool
    Palette256 bool
    Shape uint32
    HFlip, VFlip bool
    Size uint32
    RotParams uint32
    CharName uint32
    Priority uint32
    Palette uint32
    OneDimensional bool
}

func (p *PPU) UpdatePPU(addr uint32, v uint32) {

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
        p.Dispcnt.Mode = utils.GetVarData(v, 0, 2)
        p.Dispcnt.CGB = utils.BitEnabled(v, 3)
        p.Dispcnt.DisplayFrame1 = utils.BitEnabled(v, 4)
        p.Dispcnt.HBlankIntervalFree = utils.BitEnabled(v, 5)
        p.Dispcnt.OneDimensional = utils.BitEnabled(v, 6)
        p.Dispcnt.ForcedBlank = utils.BitEnabled(v, 7)

    case 0x1:
        p.Dispcnt.DisplayBg0 = utils.BitEnabled(v, 0)
        p.Dispcnt.DisplayBg1 = utils.BitEnabled(v, 1)
        p.Dispcnt.DisplayBg2 = utils.BitEnabled(v, 2)
        p.Dispcnt.DisplayBg3 = utils.BitEnabled(v, 3)
        p.Dispcnt.DisplayObj = utils.BitEnabled(v, 4)
        p.Dispcnt.DisplayWin0 = utils.BitEnabled(v, 5)
        p.Dispcnt.DisplayWin1 = utils.BitEnabled(v, 6)
        p.Dispcnt.DisplayObjWin = utils.BitEnabled(v,7)

        //p.Backgrounds[0].Enabled = p.Dispcnt.DisplayBg0
        //p.Backgrounds[1].Enabled = p.Dispcnt.DisplayBg1
        //p.Backgrounds[2].Enabled = p.Dispcnt.DisplayBg2
        //p.Backgrounds[3].Enabled = p.Dispcnt.DisplayBg3

        wins := &p.Windows
        wins.Win0.Enabled = p.Dispcnt.DisplayWin0
        wins.Win1.Enabled = p.Dispcnt.DisplayWin1
        wins.WinObj.Enabled = p.Dispcnt.DisplayObjWin && p.Dispcnt.DisplayObj
        wins.Enabled = wins.Win0.Enabled || wins.Win1.Enabled || wins.WinObj.Enabled

    case 0x50:
        p.Blend.a[0] = utils.BitEnabled(v, 0)
        p.Blend.a[1] = utils.BitEnabled(v, 1)
        p.Blend.a[2] = utils.BitEnabled(v, 2)
        p.Blend.a[3] = utils.BitEnabled(v, 3)
        p.Blend.a[4] = utils.BitEnabled(v, 4)
        p.Blend.a[5] = utils.BitEnabled(v, 5)
        p.Blend.Mode = utils.GetVarData(v, 6, 7)

    case 0x51:
        p.Blend.b[0] = utils.BitEnabled(v, 0)
        p.Blend.b[1] = utils.BitEnabled(v, 1)
        p.Blend.b[2] = utils.BitEnabled(v, 2)
        p.Blend.b[3] = utils.BitEnabled(v, 3)
        p.Blend.b[4] = utils.BitEnabled(v, 4)
        p.Blend.b[5] = utils.BitEnabled(v, 5)

    case 0x52:
        p.Blend.aEv = min(16, utils.GetVarData(v, 0, 4))

    case 0x53:
        p.Blend.bEv = min(16, utils.GetVarData(v, 0, 4))

    case 0x54:
        p.Blend.yEv = min(16, utils.GetVarData(v, 0, 4))

    }
}

func (p *PPU) UpdateWin(addr uint32, v uint32) {

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
        win0.R = v
    case WIN0Hb:
        win0.L = v
    case WIN1Ha:
        win1.R = v
        win1.oR = v

        if win1.oR == 0 && win1.oL == 0 {
            win1.R = 240
        }

    case WIN1Hb:
        win1.L = v
        win1.oL = v

        if win1.oR == 0 && win1.oL == 0 {
            win1.R = 240
        }

    case WIN0Va:
        win0.B = v
    case WIN0Vb:
        win0.T = v
    case WIN1Va:
        win1.B = v
        win1.oB = v
        if win1.T == 0 && win1.B == 0 {
            win1.B = 160
        }

    case WIN1Vb:
        win1.T = v
        win1.oT = v
        if win1.T == 0 && win1.B == 0 {
            win1.B = 160
        }

    case WININ0:
        win0.InBg0 = utils.BitEnabled(v, 0)
        win0.InBg1 = utils.BitEnabled(v, 1)
        win0.InBg2 = utils.BitEnabled(v, 2)
        win0.InBg3 = utils.BitEnabled(v, 3)
        win0.InObj = utils.BitEnabled(v, 4)
        win0.InBld = utils.BitEnabled(v, 5)
    case WININ1:
        win1.InBg0 = utils.BitEnabled(v, 0)
        win1.InBg1 = utils.BitEnabled(v, 1)
        win1.InBg2 = utils.BitEnabled(v, 2)
        win1.InBg3 = utils.BitEnabled(v, 3)
        win1.InObj = utils.BitEnabled(v, 4)
        win1.InBld = utils.BitEnabled(v, 5)
    case WINOUT:
        wins.OutBg0 = utils.BitEnabled(v, 0)
        wins.OutBg1 = utils.BitEnabled(v, 1)
        wins.OutBg2 = utils.BitEnabled(v, 2)
        wins.OutBg3 = utils.BitEnabled(v, 3)
        wins.OutObj = utils.BitEnabled(v, 4)
        wins.OutBld = utils.BitEnabled(v, 5)
    case WINOBJ:
        winObj.InBg0 = utils.BitEnabled(v, 0)
        winObj.InBg1 = utils.BitEnabled(v, 1)
        winObj.InBg2 = utils.BitEnabled(v, 2)
        winObj.InBg3 = utils.BitEnabled(v, 3)
        winObj.InObj = utils.BitEnabled(v, 4)
        winObj.InBld = utils.BitEnabled(v, 5)
    }
}

func (p *PPU) UpdateAffine(relAddr uint32) {

    paramIdx := (relAddr &^ 0b1) / 0x20

    for i := range 128 {

        obj := &p.Objects[i]

        if !obj.RotScale {
            continue
        }

        if obj.RotParams != paramIdx {
            continue
        }

        UpdateAffineParams(obj, &p.gba.Mem)
    }
}

func (p *PPU) UpdateOAM(relAddr uint32) {

    attrIdx := relAddr % 8

    m := &p.gba.Mem

    if affineParam := attrIdx == 6 || attrIdx == 7; affineParam {
        p.UpdateAffine(relAddr)
        return
    }

    objIdx := relAddr / 8

    obj := &p.Objects[objIdx]

    attr := uint32(m.OAM[relAddr])

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
            UpdateAffineParams(obj, m)
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
            UpdateAffineParams(obj, m)
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

func UpdateAffineParams(obj *Object, m *Memory) {
    paramsAddr := 0x700_0000 + (obj.RotParams * 0x20)
    obj.Pa = float32(int16(m.Read16(paramsAddr + 0x06))) / 256
    obj.Pb = float32(int16(m.Read16(paramsAddr + 0x0E))) / 256
    obj.Pc = float32(int16(m.Read16(paramsAddr + 0x16))) / 256
    obj.Pd = float32(int16(m.Read16(paramsAddr + 0x1E))) / 256
}

func (p *PPU) UpdateBackgrounds(addr, v uint32) {

    return

    switch addr {
    case 0x08:
        p.Backgrounds[0].Priority = utils.GetVarData(v, 0, 1)
        p.Backgrounds[0].CharBaseBlock = utils.GetVarData(v, 2, 3)
        p.Backgrounds[0].Mosaic = utils.BitEnabled(v, 6)
        p.Backgrounds[0].Palette256 = utils.BitEnabled(v, 7)
    case 0x09:
        p.Backgrounds[0].ScreenBaseBlock = utils.GetVarData(v, 0, 4)
        p.Backgrounds[0].AffineWrap = utils.BitEnabled(v, 5)
        p.Backgrounds[0].Size = utils.GetVarData(v, 6, 7)

    case 0x0A:
        p.Backgrounds[1].Priority = utils.GetVarData(v, 0, 1)
        p.Backgrounds[1].CharBaseBlock = utils.GetVarData(v, 2, 3)
        p.Backgrounds[1].Mosaic = utils.BitEnabled(v, 6)
        p.Backgrounds[1].Palette256 = utils.BitEnabled(v, 7)
    case 0x0B:
        p.Backgrounds[1].ScreenBaseBlock = utils.GetVarData(v, 0, 4)
        p.Backgrounds[1].AffineWrap = utils.BitEnabled(v, 5)
        p.Backgrounds[1].Size = utils.GetVarData(v, 6, 7)

    case 0x0C:
        p.Backgrounds[2].Priority = utils.GetVarData(v, 0, 1)
        p.Backgrounds[2].CharBaseBlock = utils.GetVarData(v, 2, 3)
        p.Backgrounds[2].Mosaic = utils.BitEnabled(v, 6)
        p.Backgrounds[2].Palette256 = utils.BitEnabled(v, 7)
    case 0x0D:
        p.Backgrounds[2].ScreenBaseBlock = utils.GetVarData(v, 0, 4)
        p.Backgrounds[2].AffineWrap = utils.BitEnabled(v, 5)
        p.Backgrounds[2].Size = utils.GetVarData(v, 6, 7)

    case 0x0E:
        p.Backgrounds[3].Priority = utils.GetVarData(v, 0, 1)
        p.Backgrounds[3].CharBaseBlock = utils.GetVarData(v, 2, 3)
        p.Backgrounds[3].Mosaic = utils.BitEnabled(v, 6)
        p.Backgrounds[3].Palette256 = utils.BitEnabled(v, 7)
    case 0x0F:
        p.Backgrounds[3].ScreenBaseBlock = utils.GetVarData(v, 0, 4)
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
    case 0x29:
        p.Backgrounds[2].aXOffset &^= 0xFF00
        p.Backgrounds[2].aXOffset |= v << 8
    case 0x2A:
        p.Backgrounds[2].aXOffset &^= 0xFF0000
        p.Backgrounds[2].aXOffset |= v << 16
    case 0x2B:
        p.Backgrounds[2].aXOffset &^= 0xFF000000
        p.Backgrounds[2].aXOffset |= v << 24

    case 0x2C:
        p.Backgrounds[2].aYOffset &^= 0xFF
        p.Backgrounds[2].aYOffset |= v
    case 0x2D:
        p.Backgrounds[2].aYOffset &^= 0xFF00
        p.Backgrounds[2].aYOffset |= v << 8
    case 0x2E:
        p.Backgrounds[2].aYOffset &^= 0xFF0000
        p.Backgrounds[2].aYOffset |= v << 16
    case 0x2F:
        p.Backgrounds[2].aYOffset &^= 0xFF000000
        p.Backgrounds[2].aYOffset |= v << 24

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
    case 0x39:
        p.Backgrounds[3].aXOffset &^= 0xFF00
        p.Backgrounds[3].aXOffset |= v << 8
    case 0x3A:
        p.Backgrounds[3].aXOffset &^= 0xFF0000
        p.Backgrounds[3].aXOffset |= v << 16
    case 0x3B:
        p.Backgrounds[3].aXOffset &^= 0xFF000000
        p.Backgrounds[3].aXOffset |= v << 24

    case 0x3C:
        p.Backgrounds[3].aYOffset &^= 0xFF
        p.Backgrounds[3].aYOffset |= v
    case 0x3D:
        p.Backgrounds[3].aYOffset &^= 0xFF00
        p.Backgrounds[3].aYOffset |= v << 8
    case 0x3E:
        p.Backgrounds[3].aYOffset &^= 0xFF0000
        p.Backgrounds[3].aYOffset |= v << 16
    case 0x3F:
        p.Backgrounds[3].aYOffset &^= 0xFF000000
        p.Backgrounds[3].aYOffset |= v << 24
    }
}

func (bg *Background) setSize() {

    if bg.Affine {
        switch bg.Size {
        case 0: bg.W, bg.H = 16, 16
        case 1: bg.W, bg.H = 32, 32
        case 2: bg.W, bg.H = 64, 64
        case 3: bg.W, bg.H = 128, 128
        default: panic("PROHIBITTED AFFINE BG SIZE")
        }

        return
    }

    switch bg.Size {
    case 0: bg.W, bg.H = 32, 32
    case 1: bg.W, bg.H = 64, 32
    case 2: bg.W, bg.H = 32, 64
    case 3: bg.W, bg.H = 64, 64
    default: panic("PROHIBITTED BG SIZE")
    }
}
