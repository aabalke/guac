package headless

import (
	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/emu/gb"
	"github.com/aabalke/guac/emu/gba"
	"github.com/aabalke/guac/emu/nds"
	"github.com/aabalke/guac/utils"
)

func StartHeadless() {
	path := config.Conf.General.RomPath

	switch romType := utils.GetRomType(path); romType {
	case utils.GB:
		gb := gb.NewGameBoy(nil, path)
		for {
			gb.Update(0x10000)
		}

	case utils.GBA:
		gba := gba.NewGBA(nil, path)
		for {
			gba.Update(0x10000)
		}
	case utils.NDS:
		nds := nds.NewNds(nil, path)
		for {
			nds.Update(false)
		}
	}
}
