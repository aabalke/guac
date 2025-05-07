package gba

func (gba *GBA) updateDisplay() {

	Mem := gba.Mem

	index := 0
	for y := range SCREEN_HEIGHT {
		for x := range SCREEN_WIDTH {

			palIdx := Mem.VRAM[(y*SCREEN_WIDTH)+x]

			palData := uint32(Mem.PRAM[(palIdx*2)+1])<<8 | uint32(Mem.PRAM[palIdx*2])

			r := uint8((palData) & 0b11111)
			g := uint8((palData >> 5) & 0b11111)
			b := uint8((palData >> 10) & 0b11111)

			c := convertTo24bit(r, g, b)

			(*gba.Pixels)[index] = c.R
			(*gba.Pixels)[index+1] = c.G
			(*gba.Pixels)[index+2] = c.B
			(*gba.Pixels)[index+3] = c.A
			index += 4
		}
	}
}
