package rast

import "encoding/binary"

type Vram struct {
	TextureSlots [4][0x20000]uint8
	TexPalSlots  [6][0x4000]uint8

	TextureSlotEnabled    [4]bool
	TexturePalSlotEnabled [6]bool
}

func (vm *Vram) SyncTextures(textures *[4]*[0x20000]uint8, pals *[6]*[0x4000]uint8) {
	for i := range 4 {

		slot := (*textures)[i]

		vm.TextureSlotEnabled[i] = slot != nil

		if slot != nil {
			copy(vm.TextureSlots[i][:], (*textures[i])[:])
		}
	}

	for i := range 6 {

		slot := (*pals)[i]

		vm.TexturePalSlotEnabled[i] = slot != nil

		if slot != nil {
			copy(vm.TexPalSlots[i][:], (*pals[i])[:])
		}
	}
}

func (vm *Vram) ReadTexture(addr uint32) uint8 {
	region := addr >> 17

	if !vm.TextureSlotEnabled[region] || region >= 4 {
		return 0
	}

	return vm.TextureSlots[region][addr&0x1FFFF]
}

func (vm *Vram) ReadPalTexture(addr uint32) uint16 {
	region := addr >> 14

	if region >= 6 || !vm.TexturePalSlotEnabled[region] {
		return 0
	}

	return binary.LittleEndian.Uint16(vm.TexPalSlots[region][addr&0x3FFF:])
}
