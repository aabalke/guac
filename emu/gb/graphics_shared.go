package gameboy

import (
	"unsafe"
)

const (
	LCDC = 0x40
	//STAT        = 0x41
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

const (
	PPU_HBLANK = iota
	PPU_VBLANK
	PPU_OAM
	PPU_DRAW
)

type Stat struct {
	Mode      uint8
	Match     bool
	IrqHBlank bool
	IrqVBlank bool
	IrqOam    bool
	IrqLyc    bool
}

func (s *Stat) Read() uint8 {
	v := s.Mode
	if s.Match {
		v |= 1 << 2
	}
	if s.IrqHBlank {
		v |= 1 << 3
	}
	if s.IrqVBlank {
		v |= 1 << 4
	}
	if s.IrqOam {
		v |= 1 << 5
	}
	if s.IrqLyc {
		v |= 1 << 6
	}
	return v
}

func (s *Stat) Write(v uint8) {
	s.IrqHBlank = (v>>3)&1 != 0
	s.IrqVBlank = (v>>4)&1 != 0
	s.IrqOam = (v>>5)&1 != 0
	s.IrqLyc = (v>>6)&1 != 0
}

func (gb *GameBoy) UpdateDisplay() {
	for y := range height {
		p32 := (*[width]uint32)(unsafe.Pointer(&gb.Pixels[(y*width)*4]))
		for x := range uint32(width) {
			p32[x] = gb.Screen[x][y]
		}
	}
}

func (gb *GameBoy) UpdateGraphics() {

	if lcdDisabled := (gb.MemoryBus.IO[LCDC]>>7)&1 == 0; lcdDisabled {
		gb.Timer.ScanlineCounter = 456
		gb.MemoryBus.IO[LY] = 0
		gb.Stat.Mode = 0
		return
	}

	var (
		dot         = &gb.Timer.ScanlineCounter
		stat        = &gb.Stat
		currentLine = gb.MemoryBus.IO[LY]
	    prevMode    = gb.Stat.Mode
	)

	if vblank := currentLine >= height; vblank {
		stat.Mode = PPU_VBLANK
		if stat.IrqVBlank && prevMode != PPU_VBLANK {
			gb.SetIrq(IRQ_LCD)
		}
	} else if oam := *dot >= 456-80; oam {
		stat.Mode = PPU_OAM
		if stat.IrqOam && prevMode != PPU_OAM {
			gb.SetIrq(IRQ_LCD)
		}
	} else if drawing := *dot >= 456-80-172; drawing {
		stat.Mode = PPU_DRAW
		if PPU_DRAW != prevMode {
			gb.drawScanline(int32(currentLine))
		}
	} else if hblank := prevMode != PPU_HBLANK; hblank {
		gb.hdmaTransfer()
		if stat.IrqHBlank {
			gb.SetIrq(IRQ_LCD)
		}
	}

	stat.Match = currentLine == gb.MemoryBus.IO[LYC]
	if stat.Match && stat.IrqLyc {
		gb.SetIrq(IRQ_LCD)
	}

	*dot -= gb.Cycles
	if *dot > 0 {
		return
	}

	gb.MemoryBus.IO[LY]++

	speedMultipler := 1
	if gb.DoubleSpeed {
		speedMultipler = 2
	}

	*dot += 456 * speedMultipler

	switch currentLine {
	case height: // vblank
		gb.SetIrq(IRQ_VBL)
		gb.UpdateDisplay()
	case 153: // new frame
		gb.bgPriority = [width][height]bool{}
		gb.MemoryBus.IO[LY] = 0
	}
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
