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
		gb := gb.NewGameBoy(nil, path, true)
		for {
			gb.Update()
		}

	case utils.GBA:
		gba := gba.NewGBA(nil, path, true)
		for {
			gba.Update()
		}
	case utils.NDS:
		nds := nds.NewNds(nil, path, true)
		for {
			nds.Update(false)
		}
	}
}
