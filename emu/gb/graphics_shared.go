package gameboy

import (
	"unsafe"
)

const (
	LCDC        = 0x40
	STAT        = 0x41
	SCY         = 0x42
	SCX         = 0x43
	LY          = 0x44
	LYC         = 0x45
	DMA         = 0x46
	BGPALETTE   = 0x47
	OBJ0PALETTE = 0x48
	OBJ1PALETTE = 0x49
	WY          = 0x4A
	WX          = 0x4B

	//GBC
	BCPS = 0x68
	BCPD = 0x69
	OCPS = 0x6A
	OCPD = 0x6B

	SpritePriorityOffset = 100 // random to distiguish uninitialized from valid 0

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

	gb.MemoryBus.IO[LY]++
	currentLine := gb.MemoryBus.IO[LY]

	speedMultipler := 1
	if gb.DoubleSpeed {
		speedMultipler = 2
	}

	gb.Timer.ScanlineCounter += 456 * speedMultipler

	if currentLine > 153 {
		gb.bgPriority = [width][height]bool{}
		gb.MemoryBus.IO[LY] = 0
	}

	if currentLine == height {
		gb.RequestInterrupt(IRQ_VBL)
		gb.UpdateDisplay()
	}
}

func (gb *GameBoy) enableLCD() bool {
	return (gb.MemoryBus.IO[LCDC]>>7)&1 != 0
}

func (gb *GameBoy) setLCDStatus() {

	stat := gb.MemoryBus.IO[STAT]

	if !gb.enableLCD() {
		gb.Timer.ScanlineCounter = 456
		gb.MemoryBus.IO[LY] = 0              // set y line
		gb.MemoryBus.IO[STAT] = stat &^ 0b11 // clear ppu mode
		return
	}

	currentLine := gb.MemoryBus.IO[LY]

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

	currentLineCoin := gb.MemoryBus.IO[LYC]
	if currentLine == currentLineCoin {
		stat |= 0b100
		if (stat>>6)&1 != 0 {
			gb.RequestInterrupt(IRQ_LCD)
		}

		gb.MemoryBus.IO[STAT] = stat
		return
	}

	stat &^= 0b100
	gb.MemoryBus.IO[STAT] = stat
}

func (gb *GameBoy) drawScanline(scanline int32) {

	lcdc := gb.MemoryBus.IO[LCDC]

	if gb.Color {
		gb.renderTilesGBC()
	} else if bgEnabled := (lcdc>>0)&1 != 0; bgEnabled {
		gb.renderTilesDMG()
	}

	if objEnabled := (lcdc>>1)&1 != 0; objEnabled {
		if gb.Color {
			gb.renderSpritesGBC(scanline)
		} else {
			gb.renderSpritesDMG(scanline)
		}
	}
}

func getColorVal(val1, val2, pos1, pos2 uint8) uint8 {
	return ((val1>>pos1)&1)<<1 | ((val2 >> pos2) & 1)
}

//
//func (gb *GameBoy) renderSprites(scanline int32) {
//
//    lcdc := gb.MemoryBus.Memory[LCDC]
//
//	var ySize int32 = 8
//	if (lcdc>>2)&1 != 0 {
//		ySize = 16
//	}
//
//	gb.spMinx = [width]int32{}
//	var lineSprites = 0
//	for sprite := range uint16(40) {
//		index := sprite * 4
//
//		yP := gb.MemoryBus.OAM[index]
//		yPos := int32(yP) - 16
//
//		if scanline < yPos || scanline >= (yPos+ySize) {
//			continue
//		}
//
//		// Only 10 sprites are allowed to be displayed on each line
//		if lineSprites >= 10 {
//			break
//		}
//		lineSprites++
//
//		xP := gb.MemoryBus.OAM[index+1]
//		xPos := int32(xP) - 8
//		tileLocation := gb.MemoryBus.OAM[index+2]
//		attributes := gb.MemoryBus.OAM[index+3]
//
//		yFlip := (attributes>>6)&1 != 0
//		xFlip := (attributes>>5)&1 != 0
//		priority := !((attributes>>7)&1 != 0)
//
//		// Bank the sprite data in is (CGB only)
//		var bank uint16 = 0
//		if gb.Color && (attributes>>3)&1 != 0 {
//			bank = 1
//		}
//
//		// Set the line to draw based on if the sprite is flipped on the y
//		line := scanline - yPos
//		if yFlip {
//			line = ySize - line - 1
//		}
//
//		dataAddress := (uint16(tileLocation) * 0x10) + uint16(line*2) + (bank * 0x2000)
//		data1 := gb.MemoryBus.VRAM[dataAddress]
//		data2 := gb.MemoryBus.VRAM[dataAddress+1]
//
//		for tilePixel := range uint8(8) {
//			pixel := int16(xPos) + 7 - int16(tilePixel)
//			if pixel < 0 || pixel >= width {
//				continue
//			}
//
//			if gb.spMinx[pixel] != 0 && (gb.Color || gb.spMinx[pixel] <= int32(xPos)+SpritePriorityOffset) {
//				continue
//			}
//
//			colorBit := tilePixel
//			if xFlip {
//				colorBit = byte(int8(colorBit-7) * -1)
//			}
//
//			//colorNum := getVal(data2, uint8(colorBit))
//			//colorNum <<= 1
//			//colorNum |= getVal(data1, uint8(colorBit))
//			colorNum := getColorVal(data2, data1, colorBit, colorBit)
//
//			if transparent := colorNum == 0; transparent {
//				continue
//			}
//
//			drawPixel := (priority && !gb.bgPriority[pixel][scanline]) || !gb.pixelDrawn[pixel]
//			if drawPixel {
//				if gb.Color {
//					cgbPalette := attributes & 0x7
//					color := gb.spPalette.Unpacked[(cgbPalette<<2)+(colorNum)]
//					gb.Screen[pixel][scanline] = color
//
//				} else {
//
//					pal := UNPACKED_OBJ0
//					if (attributes>>4)&1 != 0 {
//						pal = UNPACKED_OBJ1
//					}
//
//					gb.Screen[pixel][scanline] = gb.UnpackedMonoPals[pal][colorNum]
//				}
//			}
//
//			gb.spMinx[pixel] = int32(xPos) + SpritePriorityOffset
//		}
//	}
//}
