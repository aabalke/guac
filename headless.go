package main

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
		gb := gb.NewGameBoy(path, nil)
		for i := 0; true; i++ {
			gb.Update(false)
		}

	case utils.GBA:
		gba := gba.NewGBA(path, nil)
		for i := 0; true; i++ {
			gba.Update(false)
		}
	case utils.NDS:
		nds := nds.NewNds(path, nil)
		for i := 0; true; i++ {
			nds.Update(false)
		}
	}
}
