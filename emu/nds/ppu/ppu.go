package ppu

import (
	"github.com/aabalke/guac/emu/nds/utils"
)

type PPU struct {

	Dispcnt Dispcnt

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

func (p *PPU) Update(addr, v uint32) {

    switch addr {
	case 0x0:
		p.Dispcnt.Mode = utils.GetVarData(v, 0, 2)

	case 0x1:
	case 0x2:
        p.Dispcnt.DisplayMode = utils.GetVarData(v, 0, 1)
        p.Dispcnt.VramBlock = utils.GetVarData(v, 2, 3)

	case 0x3:
    }

}
