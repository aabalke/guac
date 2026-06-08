package gba

import (
	_ "embed"
)

//go:embed bios.bin
var biosFile []byte

func (gba *GBA) LoadBios() {
	for i := range len(biosFile) {

		if i >= len(gba.Mem.BIOS) {
			break
		}

		gba.Mem.BIOS[i] = biosFile[i]
	}
}
