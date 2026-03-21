package gameboy

import (
	"unsafe"
)

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

	SpritePriorityOffset = 100

    UNPACKED_BG   = 0
    UNPACKED_OBJ0 = 1
    UNPACKED_OBJ1 = 2
)

func (gb *GameBoy) UpdateDisplay() {
	for y := range height {
        p32 := (*[width]uint32)(unsafe.Pointer(&gb.Pixels[(y*width)*4]))
        for x := range uint32(width) {
            p32[x] = gb.Screen[x][y]
        }
	}
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
		gb.RequestInterrupt(IRQ_VBL)
	}
}

func (gb *GameBoy) enableLCD() bool {
	return (gb.MemoryBus.Memory[LCDC]>>7)&1 != 0
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

	vBlank := currentLine >= height                   // mode 1
	oam := gb.Timer.ScanlineCounter >= 456-80         // mode 2
	drawing := gb.Timer.ScanlineCounter >= 456-80-172 // mode 3

	setStat := func(stat uint8, mode uint8) uint8 {
		return stat&^0b11 | mode
	}

	switch {
	case vBlank:
		newMode = 1
		stat = setStat(stat, newMode)
		modeSelected = (stat>>4)&1 != 0
	case oam:
		newMode = 2
		stat = setStat(stat, newMode)
		modeSelected = (stat>>5)&1 != 0
	case drawing:
		newMode = 3
		stat = setStat(stat, newMode)

		if newMode != currMode {
			gb.drawScanline(int32(currentLine))
		}

	default:
		newMode = 0
		stat = setStat(stat, newMode)
		modeSelected = (stat>>3)&1 != 0

		if currMode != newMode {
			gb.hdmaTransfer()
		}
	}

	enteredNewMode := modeSelected && (currMode != newMode)
	if enteredNewMode {
		gb.RequestInterrupt(IRQ_LCD)
	}

	currentLineCoin := gb.MemoryBus.Memory[LYC]
	if currentLine == currentLineCoin {
		stat |= 0b100
		if (stat>>6)&1 != 0 {
			gb.RequestInterrupt(IRQ_LCD)
		}

		gb.MemoryBus.Memory[STAT] = stat
		return
	}

	stat &^= 0b100
	gb.MemoryBus.Memory[STAT] = stat
}

func (gb *GameBoy) drawScanline(scanline int32) {

	lcdc := gb.MemoryBus.Memory[LCDC]

	if bgEnabled := (lcdc>>0)&1 != 0; bgEnabled || gb.Color {
		gb.renderTiles()
	}

	if objEnabled := (lcdc>>1)&1 != 0; objEnabled {
		gb.renderSprites(scanline)
	}
}

func (gb *GameBoy) renderTiles() {

	scrollY := (gb.MemoryBus.Memory[0xFF42])
	scrollX := (gb.MemoryBus.Memory[0xFF43])
	windowY := (gb.MemoryBus.Memory[0xFF4A])
	windowX := int(gb.MemoryBus.Memory[0xFF4B]) - 7
	lcdc := gb.MemoryBus.Memory[LCDC]
	scanline := (gb.MemoryBus.Memory[LY])

	signedTiles := !((lcdc>>4)&1 != 0)
	scanLineInWindow := windowY <= (scanline)
	winEnabled := (lcdc>>5)&1 != 0
    useWindow := winEnabled && scanLineInWindow

    tileData := uint16(0x8000)
	if signedTiles {
		tileData = 0x8800
	}

    winMemory := uint16(0x9800)
	if winAddr := (lcdc>>6)&1 != 0; winAddr {
		winMemory = 0x9C00
	}

    bgMemory := uint16(0x9800)
	if bgAddr := (lcdc>>3)&1 != 0; bgAddr {
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

	for pixel := range width {
		xPos := (uint8(pixel) + scrollX)

		// Translate the current x pos to window space if necessary
		if useWindow && pixel >= windowX {
			xPos = uint8((int(pixel) - windowX))
		}

		// Which of the 32 horizontal tiles does this x_pox fall within?
		tileCol := uint16(xPos / 8)

		// Get the tile identity number

		// PER PIXEL OF SCAN LINE, NEED TO CHECK IF PIXEL >= WX AS WELL TO CHOOSE TILE ADDR (BG VS WIN)
		tileAddress := bgMemory + tileRow + tileCol
		if useWindow && pixel >= windowX {
			tileAddress = winMemory + tileRow + tileCol
		}

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
		if gb.Color && (tileAttr>>3)&1 != 0 {
			bankOffset = 0x6000
		}
		priority := (tileAttr>>7)&1 != 0

		var line byte
        if vFlip := gb.Color && (tileAttr>>6)&1 != 0; vFlip {
			// Vertical flip
			line = ((7 - yPos) & 7) * 2
		} else {
			line = (yPos & 7) * 2
		}
		// Get the tile data from memory
		data1 := gb.MemoryBus.VRAM[tileLocation+uint16(line)-bankOffset]
		data2 := gb.MemoryBus.VRAM[tileLocation+uint16(line)+1-bankOffset]

        if hFlip := gb.Color && (tileAttr>>5)&1 != 0; hFlip {
			xPos = 7 - xPos
		}

		colorBit := 7 - (xPos & 7)

		//colorNum := getVal(data2, uint8(colorBit))
		//colorNum <<= 1
		//colorNum |= getVal(data1, uint8(colorBit))
        colorNum := getColorVal(data2, data1, colorBit, colorBit)

		if gb.Color {
            if draw := !gb.bgPriority[pixel][scanline]; draw {
                cgbPalette := tileAttr & 0x7
                color := gb.bgPalette.get(cgbPalette, colorNum) | 0xFF00_0000
                gb.Screen[pixel][scanline] = color
            }

			gb.bgPriority[pixel][scanline] = priority

		} else {
            if draw := !gb.bgPriority[pixel][scanline]; draw {
                gb.Screen[pixel][scanline] = gb.UnpackedMonoPals[0][colorNum]
            }
		}

		gb.pixelDrawn[pixel] = colorNum != 0
	}
}

func (gb *GameBoy) renderSprites(scanline int32) {

	lcdControl := gb.Read(LCDC)

	var ySize int32 = 8
	if (lcdControl>>2)&1 != 0 {
		ySize = 16
	}

	var minx [width]int32
	var lineSprites = 0
	for sprite := range uint16(40) {
		index := sprite * 4

		yP := gb.MemoryBus.OAM[index]
		yPos := int32(yP) - 16

		if scanline < yPos || scanline >= (yPos+ySize) {
			continue
		}

		// Only 10 sprites are allowed to be displayed on each line
		if lineSprites >= 10 {
			break
		}
		lineSprites++

		xP := gb.MemoryBus.OAM[index + 1]
		xPos := int32(xP) - 8
		tileLocation := gb.MemoryBus.OAM[index + 2]
		attributes := gb.MemoryBus.OAM[index + 3]

		yFlip := (attributes>>6)&1 != 0
		xFlip := (attributes>>5)&1 != 0
		priority := !((attributes>>7)&1 != 0)

		// Bank the sprite data in is (CGB only)
		var bank uint16 = 0
		if gb.Color && (attributes>>3)&1 != 0 {
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

		for tilePixel := range uint8(8) {
			pixel := int16(xPos) + 7 - int16(tilePixel)
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

			//colorNum := getVal(data2, uint8(colorBit))
			//colorNum <<= 1
			//colorNum |= getVal(data1, uint8(colorBit))
            colorNum := getColorVal(data2, data1, colorBit, colorBit)

			// Colour 0 is transparent for sprites
			if colorNum == 0 {
				continue
			}

            if gb.Color {
                cgbPalette := attributes & 0x7
                color := gb.spPalette.get(cgbPalette, colorNum) | 0xFF00_0000

                drawPixel := (priority && !gb.bgPriority[pixel][scanline]) || !gb.pixelDrawn[pixel]
                if drawPixel {
                    gb.Screen[pixel][scanline] = color
                }

            } else {

                pal := UNPACKED_OBJ0
                if (attributes>>4)&1 != 0 {
                    pal = UNPACKED_OBJ1
                }

                drawPixel := (priority && !gb.bgPriority[pixel][scanline]) || !gb.pixelDrawn[pixel]
                if drawPixel {
                    gb.Screen[pixel][scanline] = gb.UnpackedMonoPals[pal][colorNum]
                }
            }

            minx[pixel] = int32(xPos) + SpritePriorityOffset
		}
	}
}

func getColorVal(val1, val2, pos1, pos2 uint8) uint8 {
	return ((val1 >> pos1) & 1) << 1 | ((val2 >> pos2) & 1)
}

