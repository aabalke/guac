package gpu

import "encoding/binary"

type Vram struct {
	TextureSlots *[4]*[0x20000]uint8
	TexPalSlots  *[6]*[0x4000]uint8
}

func (vm *Vram) SyncTextures(textures *[4]*[0x20000]uint8, pals *[6]*[0x4000]uint8) {
	// TODO: temporary - making copied breaks first texture in mario kart and 3d pal test
	vm.TexPalSlots = pals
	vm.TextureSlots = textures
	return
	for i := range 4 {

		slot := (*textures)[i]

		if slot != nil {
			t := [0x20000]uint8{}
			copy(t[:], slot[:])
			vm.TextureSlots[i] = &t
		} else {
			vm.TextureSlots[i] = nil
		}
	}

	for i := range 6 {

		slot := (*pals)[i]

		if slot != nil {
			t := [0x4000]uint8{}
			copy(t[:], slot[:])
			vm.TexPalSlots[i] = &t
		} else {
			vm.TexPalSlots[i] = nil
		}
	}
}

func (vm *Vram) ReadTexture(addr uint32) uint8 {
	region := addr >> 17

	if region >= 4 || vm.TextureSlots[region] == nil {
		return 0
	}

	return vm.TextureSlots[region][addr&0x1FFFF]
}

func (vm *Vram) ReadPalTexture(addr uint32) uint16 {
	region := addr >> 14

	if region >= 6 || vm.TexPalSlots[region] == nil {
		return 0
	}

	return binary.LittleEndian.Uint16(vm.TexPalSlots[region][addr&0x3FFF:])
}
