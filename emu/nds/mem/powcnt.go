package mem

import (
	"github.com/aabalke/guac/emu/nds/ppu"
	"github.com/aabalke/guac/emu/nds/utils"
)

type PowCnt struct {
	V uint16
	V2 uint8
}

func (p *PowCnt) WriteCNT1(b, v uint32, ppu *ppu.PPU) {

    switch b {
    case 0:

        p.V &^= 0xFF
        p.V |= uint16(v & 0b1111)

        ppu.LcdEnabled = utils.BitEnabled(v, 0)
        ppu.EngineA2D = utils.BitEnabled(v, 1)
        ppu.RenderingEngine = utils.BitEnabled(v, 2)
        ppu.GeometryEngine = utils.BitEnabled(v, 3)

    case 1:

        p.V &= 0xFF
        p.V |= uint16(v&0b1000_0010) << 8

        ppu.EngineB2D = utils.BitEnabled(v, 1)


        prevTopA := ppu.TopA
        ppu.TopA = utils.BitEnabled(v, 7)

        if prevTopA != ppu.TopA {
            a := ppu.EngineA.Pixels
            ppu.EngineA.Pixels = ppu.EngineB.Pixels
            ppu.EngineB.Pixels = a
        }
    }
}

func (p *PowCnt) WriteCNT2(v uint8) {
    p.V2 = v & 0b11
    // sound speakers, wifi
}
