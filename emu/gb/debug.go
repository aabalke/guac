package gameboy

func (gb *GameBoy) GetOAMTile() [] byte {

	pixels := make([][]byte, 8)

	for i := range len(pixels) {
		pixels[i] = make([]byte, 8)
	}

    o := 0x0
    mem := gb.MemoryBus.VRAM[(0x10*o):]
	xOffset, yOffset := 0, 0

	//for i := 0; i < len(mem); i += 0x10 {
	for i := 0; i < 0x10; i += 0x10 {

		tile := RenderTile(mem[i : i+0x10])

		for y := range len(tile) {
			for x := range len(tile[0]) {
				yIdx := yOffset*tileWidth + y
				xIdx := xOffset*tileWidth + x
				//pixels[yIdx][xIdx] = tile[y][x]
				pixels[yIdx][xIdx] = tile[x][y]
			}
		}

		xOffset++
		if xOffset == 20 {
			yOffset++
			xOffset = 0
		}
	}

	output := gb.ConvertPixels(pixels)

	return output


    return nil
}

func (gb *GameBoy) GetBgTiles() []byte {

	pixels := make([][]byte, 8)

	for i := range len(pixels) {
		pixels[i] = make([]byte, 8)
	}

    o := 0x31
    mem := gb.MemoryBus.VRAM[(0x10*o):]
	xOffset, yOffset := 0, 0

	//for i := 0; i < len(mem); i += 0x10 {
	for i := 0; i < 0x10; i += 0x10 {

		tile := RenderTile(mem[i : i+0x10])

		for y := range len(tile) {
			for x := range len(tile[0]) {
				yIdx := yOffset*tileWidth + y
				xIdx := xOffset*tileWidth + x
				//pixels[yIdx][xIdx] = tile[y][x]
				pixels[yIdx][xIdx] = tile[x][y]
			}
		}

		xOffset++
		if xOffset == 20 {
			yOffset++
			xOffset = 0
		}
	}

	output := gb.ConvertPixels(pixels)

	return output
}

func RenderTile(data []uint8) [8][8]uint8 {

	if len(data) != 16 {
		panic("Pixel data != 16")
	}

	var pixels [8][8]uint8

	x := 0
	for i := 0; i < len(data); i = i + 2 {

		a := data[i]
		b := data[i+1]

		//fmt.Printf("a: %X b: %X\n", a, b)

		var mask uint8 = 0x80 //0b10000000
		for y := range tileWidth {

			av := (a & mask) == mask
			bv := (b & mask) == mask
			mask = mask >> 1

			switch true {
			case !av && !bv:
				pixels[x][y] = 0
			case av && !bv:
				pixels[x][y] = 1
			case !av && bv:
				pixels[x][y] = 2
			case av && bv:
				pixels[x][y] = 3
			}
		}
		x++
	}

	return pixels
}

func (gb *GameBoy) ConvertPixels(pixels [][]uint8) []byte {

	output := make([]byte, len(pixels[0])*len(pixels)*4)

	index := 0
	for y := range len(pixels[0]) {
		for x := range len(pixels) {
			v := pixels[x][y]
			if !gb.Color {
				switch v {
				case 0:
					ApplyColor(gb.Palette[0], &output, index)
				case 1:
					ApplyColor(gb.Palette[1], &output, index)
				case 2:
					ApplyColor(gb.Palette[2], &output, index)
				case 3:
					ApplyColor(gb.Palette[3], &output, index)
				}

				index += 4
				continue
			}

			//ApplyCGBColor(v, gb.Pixels, index)
			//index += 4
		}
	}

	return output
}
