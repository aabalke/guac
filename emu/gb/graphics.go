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
	InterruptLCD    = 0b10

	SpritePriorityOffset = 100
)

var tileScanline [width]uint8

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

    if gb.Timer.ScanlineCounter > 0 {
        return
    }

	gb.MemoryBus.Memory[LY]++
	currentLine := gb.MemoryBus.Memory[LY]

	speedMultipler := 1
	if gb.DoubleSpeed {
		speedMultipler = 2
	}

	gb.Timer.ScanlineCounter += 456 * speedMultipler

	if currentLine > 153 {
        gb.bgPriority = [width][height]bool{}
		gb.MemoryBus.Memory[LY] = 0
	}

	if currentLine == height {
		gb.RequestInterrupt(InterruptVBlank)
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
		gb.MemoryBus.Memory[LY] = 0              // set y line
		gb.MemoryBus.Memory[STAT] = stat &^ 0b11 // clear ppu mode
		return
	}

	currentLine := gb.MemoryBus.Memory[LY]

	currMode := stat & 0b11
	var newMode uint8 = 0
	modeSelected := false

	vBlank := currentLine >= height                      // mode 1
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

        if newMode != currMode {
            gb.drawScanline(int32(currentLine))
        }

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

func (gb *GameBoy) drawScanline(scanline int32) {

	lcdc := gb.MemoryBus.Memory[LCDC]

	if bgEnabled := gb.flagEnabled(lcdc, 0); bgEnabled || gb.Color {
		gb.renderTiles()
	}

	if objEnabled := gb.flagEnabled(lcdc, 1); objEnabled {
		gb.renderSprites(scanline)
	}
}

func (gb *GameBoy) renderTiles() {

    //Mem := &gb.MemoryBus.Memory

	scrollY := (gb.MemoryBus.Memory[0xFF42])
	scrollX := (gb.MemoryBus.Memory[0xFF43])
	windowY := (gb.MemoryBus.Memory[0xFF4A])
	windowX := int(gb.MemoryBus.Memory[0xFF4B]) - 7
	lcdc := gb.MemoryBus.Memory[LCDC]
	scanline := (gb.MemoryBus.Memory[LY])

	winAddr := gb.flagEnabled(lcdc, 6)
	winEnabled := gb.flagEnabled(lcdc, 5)
	signedTiles := !gb.flagEnabled(lcdc, 4)
	bgAddr := gb.flagEnabled(lcdc, 3)

	useWindow := false
	scanLineInWindow := windowY <= (scanline)
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

	// yPos is used to calc which of 32 v-lines the current scanline is drawing
	var yPos uint8
	if !useWindow {
		yPos = uint8((scrollY + scanline))
	} else {
		yPos = uint8((scanline - windowY))
	}

	// which of the 8 vertical pixels of the current tile is the scanline on?
	var tileRow = uint16(yPos/8) * 32

    tileScanline = [width]uint8{}
	for pixel := range width {
		xPos := (uint8(pixel) + scrollX)

		// Translate the current x pos to window space if necessary
		if useWindow && pixel >= windowX {
			xPos = uint8((int(pixel) - windowX))
		}

		// Which of the 32 horizontal tiles does this x_pox fall within?
		tileCol := uint16(xPos / 8)

		// Get the tile identity number
		tileAddress := bgMemory + tileRow + tileCol

		// Deduce where this tile id is in memory
		tileLocation := tileData
		if !signedTiles {
			tileNum := int16(gb.MemoryBus.VRAM[tileAddress-0x8000])
			tileLocation = tileLocation + uint16(tileNum*16)
		} else {
			tileNum := int16(int8(gb.MemoryBus.VRAM[tileAddress-0x8000]))
			tileLocation = uint16(int32(tileLocation) + int32((tileNum+128)*16))
		}

		bankOffset := uint16(0x8000)

		// Attributes used in CGB mode TODO: check in CGB mode
		//
		//    Bit 0-2  Background Palette number  (BGP0-7)
		//    Bit 3    Tile VRAM Bank number      (0=Bank 0, 1=Bank 1)
		//    Bit 5    Horizontal Flip            (0=Normal, 1=Mirror horizontally)
		//    Bit 6    Vertical Flip              (0=Normal, 1=Mirror vertically)
		//    Bit 7    BG-to-OAM Priority         (0=Use OAM priority bit, 1=BG Priority)
		//
		tileAttr := gb.MemoryBus.VRAM[tileAddress-0x6000]
		if gb.Color && gb.flagEnabled(tileAttr, 3) {
			bankOffset = 0x6000
		}
		priority := gb.flagEnabled(tileAttr, 7)

		var line byte
		if gb.Color && gb.flagEnabled(tileAttr, 6) {
			// Vertical flip
			line = ((7 - yPos) % 8) * 2
		} else {
			line = (yPos % 8) * 2
		}
		// Get the tile data from memory
		data1 := gb.MemoryBus.VRAM[tileLocation+uint16(line)-bankOffset]
		data2 := gb.MemoryBus.VRAM[tileLocation+uint16(line)+1-bankOffset]

		if gb.Color && gb.flagEnabled(tileAttr, 5) {
			// Horizontal flip
			xPos = 7 - xPos
		}

		colorBit := -(int(xPos%8) - 7)

		colorNum := getVal(data2, uint8(colorBit))
		colorNum <<= 1
		colorNum |= getVal(data1, uint8(colorBit))

		var color uint32
		if gb.Color {
			cgbPalette := tileAttr & 0x7
			color = gb.bgPalette.get(cgbPalette, colorNum)
		} else {
			color = uint32(gb.getColor(colorNum, BGPALETTE))
		}

		if outOfBounds := (scanline < 0 ||
			scanline > 143 ||
			pixel < 0 ||
			pixel > 159); outOfBounds {
			continue
		}

        if !gb.bgPriority[pixel][scanline] || tileScanline[scanline] == 0 {
            gb.Screen[pixel][scanline] = color
        }

        if gb.Color {
            gb.bgPriority[pixel][scanline] = priority
        }

        tileScanline[pixel] = colorNum
	}
}

func (gb *GameBoy) renderSprites(scanline int32) {

    lcdControl, _ := gb.ReadByte(LCDC)

	var ySize int32 = 8
	if gb.flagEnabled(lcdControl, 2) {
		ySize = 16
	}

	var minx [width]int32
	var lineSprites = 0
	for sprite := uint16(0); sprite < 40; sprite++ {
		index := sprite * 4

		yP, _ := gb.ReadByte(0xFE00+index)
        yPos := int32(yP) - 16

		if scanline < yPos || scanline >= (yPos+ySize) {
			continue
		}

		// Only 10 sprites are allowed to be displayed on each line
		if lineSprites >= 10 {
			break
		}
		lineSprites++

		xP, _ := gb.ReadByte(uint16(0xFE00+index+1))
        xPos := int32(xP) - 8
		tileLocation, _ := gb.ReadByte(uint16(0xFE00 + index + 2))
		attributes, _ := gb.ReadByte(uint16(0xFE00 + index + 3))

		yFlip := gb.flagEnabled(attributes, 6)
		xFlip := gb.flagEnabled(attributes, 5)
		priority := !gb.flagEnabled(attributes, 7)

		// Bank the sprite data in is (CGB only)
		var bank uint16 = 0
		if gb.Color && gb.flagEnabled(attributes, 3) {
			bank = 1
		}

		// Set the line to draw based on if the sprite is flipped on the y
		line := scanline - yPos
		if yFlip {
			line = ySize - line - 1
		}


		dataAddress := (uint16(tileLocation) * 0x10) + uint16(line*2) + (bank * 0x2000)
		data1 := gb.MemoryBus.VRAM[dataAddress]
		data2 := gb.MemoryBus.VRAM[dataAddress+1]

		for tilePixel := byte(0); tilePixel < 8; tilePixel++ {
			pixel := int16(xPos) + 7-int16(tilePixel)
			if pixel < 0 || pixel >= width {
				continue
			}

			if minx[pixel] != 0 && (gb.Color || minx[pixel] <= int32(xPos)+SpritePriorityOffset) {
				continue
			}

			colorBit := tilePixel
			if xFlip {
				colorBit = byte(int8(colorBit-7) * -1)
			}

			// Find the colour value by combining the data bits

			colorNum := getVal(data2, uint8(colorBit))
			colorNum <<= 1
			colorNum |= getVal(data1, uint8(colorBit))

			// Colour 0 is transparent for sprites
            if colorNum == 0 {
                continue
            }

            var color uint32
            if gb.Color {
                cgbPalette := attributes & 0x7
                color = gb.spPalette.get(cgbPalette, colorNum)

            } else {
                colorAddr := uint16(OBJ0PALETTE)
                if gb.flagEnabled(attributes, 4) {
                    colorAddr = OBJ1PALETTE
                }

                color = uint32(gb.getColor(colorNum, colorAddr))
            }

            drawPixel := (priority && !gb.bgPriority[pixel][scanline]) || tileScanline[pixel] == 0 
            if drawPixel {
                gb.Screen[pixel][scanline] = color
            }

            minx[pixel] = int32(xPos) + SpritePriorityOffset
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

func getVal(val uint8, pos uint8) uint8 {
	return (val >> pos) & 1
}
