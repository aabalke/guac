package gameboy

const (
	LCDC        = 0xFF40
	STAT        = 0xFF41
	SCY         = 0xFF42
	SCX         = 0xFF43
	LY          = 0xFF44
	LYC         = 0xFF45
	DMA         = 0xFF46
	BGPALETTE   = 0xFF47
	OBJ0PALETTE = 0xFF48
	OBJ1PALETTE = 0xFF49
	WY          = 0xFF4A
	WX          = 0xFF4B

    //GBC
    BCPS = 0xFF68
    BCPD = 0xFF69
    OCPS = 0xFF6A
    OCPD = 0xFF6B



    InterruptVBlank = 0b1
    InterruptLCD = 0b10
)

var last uint16

func (gb *GameBoy) flagEnabled(reg uint8, bit uint8) bool {
	mask := uint8(0b1) << bit
	return reg&mask == mask
}

func (gb *GameBoy) UpdateGraphics() {

	gb.setLCDStatus()

	if !gb.enableLCD() {
		return
	}

	gb.Timer.ScanlineCounter -= gb.Cycles

	if gb.Timer.ScanlineCounter <= 0 {

		gb.MemoryBus.Memory[LY]++
		currentLine := gb.MemoryBus.Memory[LY]

        speedMultipler := 1
        if gb.DoubleSpeed {
            speedMultipler = 2
        }

		gb.Timer.ScanlineCounter += 456 * speedMultipler

		switch {
		case currentLine == 144:
			gb.drawScanline()
			gb.RequestInterrupt(InterruptVBlank)
		case currentLine > 153:
			gb.MemoryBus.Memory[LY] = 0
			gb.drawScanline()
		case currentLine < 144:
			gb.drawScanline()
		}
	}
}

func (gb *GameBoy) enableLCD() bool {
	reg := gb.MemoryBus.Memory[LCDC]
	return gb.flagEnabled(reg, 7)
}

func (gb *GameBoy) setLCDStatus() {

	stat := gb.MemoryBus.Memory[STAT]

	if !gb.enableLCD() {
		gb.Timer.ScanlineCounter = 456
		gb.MemoryBus.Memory[LY] = 0                   // set y line
		gb.MemoryBus.Memory[STAT] = stat &^ 0b11 // clear ppu mode
		return
	}

	currentLine := gb.MemoryBus.Memory[LY]

	currMode := stat & 0b11
	var newMode uint8 = 0
	modeSelected := false

	vBlank := currentLine >= 144                      // mode 1
	oam := gb.Timer.ScanlineCounter >= 456-80         // mode 2
	drawing := gb.Timer.ScanlineCounter >= 456-80-172 // mode 3

	setStat := func(stat uint8, mode uint8) uint8 {
		return stat&^0b11 | mode
	}

	switch {
	case vBlank:
		newMode = 1
		stat = setStat(stat, newMode)
		modeSelected = gb.flagEnabled(stat, 4)
	case oam:
		newMode = 2
		stat = setStat(stat, newMode)
		modeSelected = gb.flagEnabled(stat, 5)
	case drawing:
		newMode = 3
		stat = setStat(stat, newMode)
	default:
		newMode = 0
		stat = setStat(stat, newMode)
		modeSelected = gb.flagEnabled(stat, 3)

        if currMode != newMode {
            gb.hdmaTransfer()
        }
	}

	enteredNewMode := modeSelected && (currMode != newMode)
	if enteredNewMode {
		gb.RequestInterrupt(InterruptLCD)
	}

	currentLineCoin := gb.MemoryBus.Memory[LYC]
	if currentLine == currentLineCoin {
		stat |= 0b100
		if gb.flagEnabled(stat, 6) {
			gb.RequestInterrupt(InterruptLCD)
		}

		gb.MemoryBus.Memory[STAT] = stat
		return
	}

	stat &^= 0b100
	gb.MemoryBus.Memory[STAT] = stat
}

func (gb *GameBoy) drawScanline() {

	lcdc := gb.MemoryBus.Memory[LCDC]

	if bgEnabled := gb.flagEnabled(lcdc, 0); bgEnabled || gb.Color {
		gb.renderBg()
	}

	if objEnabled := gb.flagEnabled(lcdc, 1); objEnabled {
		gb.renderObject()
	}
}

func (gb *GameBoy) renderBg() {

	Memory := &gb.MemoryBus.Memory
	lcdc := Memory[LCDC]
	scy := Memory[SCY]
	scx := Memory[SCX]
	wy := Memory[WY]
	wx := Memory[WX] - 7
	scanline := Memory[LY]

	winAddr := gb.flagEnabled(lcdc, 6)
	winEnabled := gb.flagEnabled(lcdc, 5)
	signedTiles := !gb.flagEnabled(lcdc, 4)
	bgAddr := gb.flagEnabled(lcdc, 3)

	useWindow := false
	scanLineInWindow := wy <= scanline
	if winEnabled && scanLineInWindow {
		useWindow = true
	}

	var tileData uint16 = 0x8000
	if signedTiles {
		tileData = 0x8800
	}

	var bgMemory uint16 = 0x9800
	if (!useWindow && bgAddr) || (useWindow && winAddr) {
		bgMemory = 0x9C00
	}

	yPos := scanline - wy
	if !useWindow {
		yPos = scy + scanline
	}

	row := uint16(yPos/8) * 32

	for pixel := uint8(0); pixel < 160; pixel++ {

		xPos := pixel + scx

		if useWindow && (pixel >= wx) {
			xPos = pixel - wx
		}

		col := uint16(xPos / 8)

		tileAddr := bgMemory + row + col
		var tileLocation uint16 = 0
		if signedTiles {
			//tileNum := int16(int8(Memory[tileAddr]))
			tileNum := int16(int8(gb.MemoryBus.VRAM[tileAddr-0x8000]))
			tileLocation = uint16(int32(tileData) + int32((tileNum+128)*16))
		} else {
			//tileNum := int16(Memory[tileAddr])
			tileNum := int16(gb.MemoryBus.VRAM[tileAddr-0x8000])
			tileLocation = tileData + uint16(tileNum*16)
		}

        if !gb.Color {

            line := (yPos % 8) * 2

            data1 := gb.MemoryBus.VRAM[tileLocation+uint16(line)-uint16(0x8000)]
            data2 := gb.MemoryBus.VRAM[tileLocation+uint16(line)+1-uint16(0x8000)]

            colorBit := -(int(xPos%8) - 7)

            colorNum := getVal(data2, uint8(colorBit))
            colorNum <<= 1
            colorNum |= getVal(data1, uint8(colorBit))
            color := gb.getColor(colorNum, BGPALETTE)

            if outOfBounds := (scanline < 0 ||
                scanline > 143 ||
                pixel < 0 ||
                pixel > 159); outOfBounds {
                continue
            }

            gb.ScanLineBG[pixel] = color == 0
            gb.Display.Screen[pixel][scanline] = uint32(color)
            continue
        }

        var bank uint16 = 0x8000
        tileAttr := gb.MemoryBus.VRAM[tileAddr-0x6000]

        if gb.flagEnabled(tileAttr, 3) {
            bank = 0x6000
        }

        priority := gb.flagEnabled(tileAttr, 7)

        var line uint8 = (yPos % 8) * 2
        if gb.flagEnabled(tileAttr, 6) {
            line = ((7 - yPos) % 8) * 2
        }

        data1 := gb.MemoryBus.VRAM[tileLocation+uint16(line)-bank]
        data2 := gb.MemoryBus.VRAM[tileLocation+uint16(line)+1-bank]

        if gb.flagEnabled(tileAttr, 5) {
            xPos = 7 - xPos
        }


        colorBit := -(int(xPos%8) - 7)

        colorNum := getVal(data2, uint8(colorBit))
        colorNum <<= 1
        colorNum |= getVal(data1, uint8(colorBit))

        cgbPalette := tileAttr & 0x7

        color := gb.bgPalette.get(cgbPalette, colorNum)

        if outOfBounds := (scanline < 0 ||
        scanline > 143 ||
        pixel < 0 ||
        pixel > 159); outOfBounds {
            continue
        }

        gb.ScanLineBG[pixel] = priority
        gb.Display.Screen[pixel][scanline] = color
	}
}

func (gb *GameBoy) renderObject() {

    Mem := &gb.MemoryBus



	lcdc := Mem.Memory[LCDC]
	use8x16 := gb.flagEnabled(lcdc, 2)

	for sprite := range 40 {

		index := sprite * 4
		objAddr := 0xFE00 + uint16(index)
		yPos := Mem.Memory[objAddr] - 16
		xPos := Mem.Memory[objAddr+1] - 8
		tileLocation := Mem.Memory[objAddr+2]
		attributes := Mem.Memory[objAddr+3]

		yFlip := gb.flagEnabled(attributes, 6)
		xFlip := gb.flagEnabled(attributes, 5)
		priority := !gb.flagEnabled(attributes, 7)
		scanline := Mem.Memory[LY]

		var ysize uint8 = 8
		if use8x16 {
			ysize = 16
		}

		// does this sprite intercept with the scanline?
		intercept := (scanline >= yPos) && (scanline < (yPos + ysize))

		if intercept {
			line := int(scanline - yPos)
			if yFlip {
				line -= int(ysize)
				line *= -1
			}
			line *= 2
			dataAddress := (uint16(int(tileLocation)*16 + line))
			data1 := Mem.VRAM[dataAddress]
			data2 := Mem.VRAM[dataAddress+1]

			for tilePixel := 7; tilePixel >= 0; tilePixel-- {
				colorBit := tilePixel

				if xFlip {
					colorBit -= 7
					colorBit *= -1
				}

				colorNum := getVal(data2, uint8(colorBit))
				colorNum <<= 1
				colorNum |= getVal(data1, uint8(colorBit))

                if colorNum == 0 {
                    continue
                }

                xPix := 0 - tilePixel + 7
                pixel := int(xPos) + xPix
                final := Mem.Memory[LY]

                if final < 0 || final > 143 || pixel < 0 || pixel > 159 {
                    continue
                }

                if !gb.Color {

                    colorAddr := uint16(OBJ0PALETTE)
                    if gb.flagEnabled(attributes, 4) {
                        colorAddr = OBJ1PALETTE
                    }

                    color := gb.getColor(colorNum, colorAddr)

                    if gb.ScanLineBG[pixel] || priority {
                        gb.Display.Screen[pixel][final] = uint32(color)
                    }
                    continue
                }

                cgbPalette := attributes & 0x7
                color := gb.spPalette.get(cgbPalette, colorNum)

                if gb.ScanLineBG[pixel] || priority {
                    gb.Display.Screen[pixel][final] = uint32(color)
                }
			}
		}
	}
}

func (gb *GameBoy) getColor(colorNum uint8, addr uint16) uint8 {

	palette := gb.MemoryBus.Memory[addr]

	var hi uint8 = 1
	var lo uint8 = 0
	switch colorNum {
	case 0:
		hi = 1
		lo = 0
	case 1:
		hi = 3
		lo = 2
	case 2:
		hi = 5
		lo = 4
	case 3:
		hi = 7
		lo = 6
	}

	var color uint8 = getVal(palette, hi) << 1
	color |= getVal(palette, lo)
	return color
}

//func getVal(val uint8, pos uint8) uint8 {
func getVal(val uint8, pos uint8) uint8 {
	return (val >> pos) & 1
}
