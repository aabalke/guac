package ppu

import (
	"fmt"

	"github.com/aabalke/guac/emu/nds/utils"
)

type PPU struct {

	Dispcnt Dispcnt
    PowCnt1 PowCnt1

	//Objects     [128]Object
	//Backgrounds [4]Background
	//Windows     Windows
	//Blend       Blend
	//Mosaic      Mosaic

	//bgPriorities  [4][]uint32
	//objPriorities [4][]uint32
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

type PowCnt1 struct {
    V uint16
    LcdEnabled bool
    EngineA2D, EngineB2D bool
    RenderingEngine, GeometryEngine bool
    TopA bool
}

func (p *PPU) Update(addr, v uint32) {

    switch addr {
	case 0x0:
        fmt.Printf("DISPCNT WRITE %08X %02X\n", addr, v)
		p.Dispcnt.Mode = utils.GetVarData(v, 0, 2)

	case 0x1:
        fmt.Printf("DISPCNT WRITE %08X %02X\n", addr, v)
	case 0x2:
        fmt.Printf("DISPCNT WRITE %08X %02X\n", addr, v)
        p.Dispcnt.DisplayMode = utils.GetVarData(v, 0, 1)
        p.Dispcnt.VramBlock = utils.GetVarData(v, 2, 3)

	case 0x3:
        fmt.Printf("DISPCNT WRITE %08X %02X\n", addr, v)

    case 0x304:

        p.PowCnt1.LcdEnabled      = utils.BitEnabled(v, 0)
        p.PowCnt1.EngineA2D       = utils.BitEnabled(v, 1)
        p.PowCnt1.RenderingEngine = utils.BitEnabled(v, 2)
        p.PowCnt1.GeometryEngine  = utils.BitEnabled(v, 3)

    case 0x305:
        p.PowCnt1.EngineB2D = utils.BitEnabled(v, 1)
        p.PowCnt1.TopA      = utils.BitEnabled(v, 7)
    }
}
