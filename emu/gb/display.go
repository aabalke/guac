package gameboy

func (gb *GameBoy) GetPixels() []byte {
    return *gb.Pixels
}

func (gb *GameBoy) UpdateDisplay() {
	index := 0

    for y := range height {
        for x := range width {

			v := gb.Screen[x][y]

            if !gb.Color {
                switch v {
                case 0: ApplyColor(gb.Palette[0], gb.Pixels, index)
                case 1: ApplyColor(gb.Palette[1], gb.Pixels, index)
                case 2: ApplyColor(gb.Palette[2], gb.Pixels, index)
                case 3: ApplyColor(gb.Palette[3], gb.Pixels, index)
                }

                index += 4
                continue
            }

            ApplyCGBColor(v, gb.Pixels, index)

			index += 4
		}
	}
}

func ApplyColor(color []uint8, pixels *[]byte, i int) {
    (*pixels)[i] = color[0]
    (*pixels)[i+1] = color[1]
    (*pixels)[i+2] = color[2]
    (*pixels)[i+3] = 255
}

func ApplyCGBColor(colorCombined uint32, pixels *[]byte, i int) {

    color := []uint8{
        uint8(colorCombined & 0xFF0000 >> 16),
        uint8(colorCombined & 0xFF00 >> 8),
        uint8(colorCombined & 0xFF),
    }

    ApplyColor(color, pixels, i)
}
