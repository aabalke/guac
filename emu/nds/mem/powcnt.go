package mem

import (
	"github.com/aabalke/guac/emu/nds/ppu"
)

type PowCnt struct {
	ppu *ppu.PPU
	V   uint16
	V2  uint8
}

func NewPowCnt(ppu *ppu.PPU) *PowCnt {
	p := &PowCnt{
		ppu: ppu,
	}

	p.WriteCnt1(0, 0x0F)
	p.WriteCnt1(1, 0x82)
	return p
}

func (p *PowCnt) WriteCnt1(b, v uint8) {
	if b == 0 {
		p.V = (p.V & 0xFF00) | uint16(v&0xF)
		p.ppu.LcdEnabled = (v>>0)&1 != 0
		p.ppu.EngineA2D = (v>>1)&1 != 0
		p.ppu.RenderingEngine = (v>>2)&1 != 0
		p.ppu.GeometryEngine = (v>>3)&1 != 0
		return
	}

	p.V = (p.V & 0x00FF) | ((uint16(v) << 8) & 0x8200)
	p.ppu.EngineB2D = (v>>1)&1 != 0
	p.ppu.TopA = v&0x80 != 0
}

func (p *PowCnt) WriteCnt2(v uint8) {
	p.V2 = v & 3
	// sound speakers, wifi
}
