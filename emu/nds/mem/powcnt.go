package mem

import (
	"github.com/aabalke/guac/emu/nds/ppu"
)

type PowCnt struct {
	V  uint16
	V2 uint8
}

func (p *PowCnt) WriteCNT1(b, v uint32, ppu *ppu.PPU) {

	switch b {
	case 0:

		p.V &^= 0xFF
		p.V |= uint16(v & 0xF)

		ppu.LcdEnabled = (v>>0)&1 != 0
		ppu.EngineA2D = (v>>1)&1 != 0
		ppu.RenderingEngine = (v>>2)&1 != 0
		ppu.GeometryEngine = (v>>3)&1 != 0

	case 1:

		p.V &= 0xFF
		p.V |= uint16(v&0b1000_0010) << 8

		ppu.EngineB2D = (v>>1)&1 != 0
		ppu.TopA = (v>>7)&1 != 0
	}
}

func (p *PowCnt) WriteCNT2(v uint8) {
	p.V2 = v & 0b11
	// sound speakers, wifi
}
