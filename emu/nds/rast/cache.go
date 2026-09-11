package rast

import (
	"github.com/aabalke/guac/emu/nds/rast/gl"
)

type TextureCache map[key]*[]gl.Color

type key struct{ base, offset uint32 }

func (t *TextureCache) Reset() {
	clear(*t)
}

func (t *TextureCache) Add(vram VRAM, tex *Texture, key key) {
	switch tex.Format {
	case TEX_FMT_4_PAL:
		(*t)[key] = t.getPaletted(vram, tex, 2, 2)
	case TEX_FMT_16_PAL:
		(*t)[key] = t.getPaletted(vram, tex, 4, 1)
	case TEX_FMT_256_PAL:
		(*t)[key] = t.getPaletted(vram, tex, 8, 0)
	case TEX_FMT_A3I5:
		(*t)[key] = t.getTranslucent(vram, tex, 5)
	case TEX_FMT_A5I3:
		(*t)[key] = t.getTranslucent(vram, tex, 3)
	case TEX_FMT_4X4:
		(*t)[key] = t.getCompressed(vram, tex)
	case TEX_FMT_DIRECT:
		(*t)[key] = t.getDirect(vram, tex)
	default:
		panic("UNSETUP TEX CACHE METHOD")
	}
}

func (t *TextureCache) Get(vram VRAM, tex *Texture) *[]gl.Color {
	key := key{tex.PaletteBaseAddr, tex.VramOffset}
	if v, ok := (*t)[key]; ok {
		return v
	}
	t.Add(vram, tex, key)
	return (*t)[key]
}

func (t *TextureCache) getDirect(vram VRAM, tex *Texture) *[]gl.Color {
	out := make([]gl.Color, tex.SizeS*tex.SizeT)

	for y := range tex.SizeT {
		for x := range tex.SizeS {
			i := x + (y * tex.SizeS)
			addr := tex.VramOffset + (i * 2)
			data := uint32(vram.ReadTexture(addr + 0))
			data |= uint32(vram.ReadTexture(addr+1)) << 8

			if transparent := data&0x8000 == 0; transparent {
				out[i] = gl.Transparent
				continue
			}

			out[i] = gl.MakeColorFrom15Bit(
				uint8(data&0x1F),
				uint8(data>>5)&0x1F,
				uint8(data>>10)&0x1F,
			)
		}
	}

	return &out
}

func (t *TextureCache) getPaletted(vram VRAM, tex *Texture, bitsPerTexel, bitsPerTexelShift uint32) *[]gl.Color {
	out := make([]gl.Color, tex.SizeS*tex.SizeT)

	palBase := tex.PaletteBaseAddr

	if bitsPerTexel == 2 {
		palBase *= 0x8
	} else {
		palBase *= 0x10
	}

	for y := range tex.SizeT {
		for x := range tex.SizeS {
			i := uint32(x + (y * tex.SizeS))

			palIdx := uint32(vram.ReadTexture(tex.VramOffset + (i >> bitsPerTexelShift)))
			palIdx = (palIdx >> ((i & ((8 / bitsPerTexel) - 1)) * bitsPerTexel)) & ((1 << bitsPerTexel) - 1)

			if palIdx == 0 && tex.TransparentZero {
				out[i] = gl.Transparent
				continue
			}

			// palettes take up 2 bytes each
			palIdx *= 2

			data := vram.ReadPalTexture(palBase + palIdx)

			out[i] = gl.MakeColorFrom15Bit(
				uint8(data)&0x1F,
				uint8(data>>5)&0x1F,
				uint8(data>>10)&0x1F,
			)
		}
	}

	return &out
}

func (t *TextureCache) getTranslucent(vram VRAM, tex *Texture, colorBits uint8) *[]gl.Color {
	out := make([]gl.Color, tex.SizeS*tex.SizeT)

	tex.PaletteBaseAddr *= 0x10

	for y := range tex.SizeT {
		for x := range tex.SizeS {

			var (
				i        = (x + (y * tex.SizeS))
				palIdx   = uint32(vram.ReadTexture(tex.VramOffset + i))
				colorIdx = (palIdx & ((1 << colorBits) - 1)) * 2
				data     = vram.ReadPalTexture(tex.PaletteBaseAddr + colorIdx)
			)

			out[i] = gl.MakeColorFrom15Bit(
				uint8(data&0x1F),
				uint8(data>>5)&0x1F,
				uint8(data>>10)&0x1F,
			)

			switch colorBits {
			case 3:
				out[i].A = float32(palIdx>>3) / 31
			case 5:
				out[i].A = float32(palIdx>>5) / 7
			}
		}
	}

	return &out
}

// rasky/ndsemu

func (t *TextureCache) getCompressed(vram VRAM, tex *Texture) *[]gl.Color {
	off := tex.VramOffset
	out := make([]gl.Color, tex.SizeS*tex.SizeT)

	const SLOT_SIZE = 128 * 1024

	var xtraoff uint32
	switch slot := off / SLOT_SIZE; slot {
	case 0:
		xtraoff = SLOT_SIZE + off/2
	case 2:
		xtraoff = SLOT_SIZE + (off-2*SLOT_SIZE)/2 + 0x10000
	default:
		panic("Invalid Slot 4x4 Compressed Texture")
	}

	for y := uint32(0); y < tex.SizeT; y += 4 {
		for x := uint32(0); x < tex.SizeS; x += 4 {

			var (
				xtra = (uint32(vram.ReadTexture(xtraoff+0)) |
					uint32(vram.ReadTexture(xtraoff+1))<<8)

				mode    = xtra >> 14
				paloff  = uint32(xtra & 0x3FFF)
				palAddr = (tex.PaletteBaseAddr * 0x10) + paloff*4

				colors = [4]uint16{
					vram.ReadPalTexture(palAddr + 0),
					vram.ReadPalTexture(palAddr + 2),
				}
			)
			xtraoff += 2

			switch mode {
			case 0:
				colors[2] = vram.ReadPalTexture(palAddr + 4)
			case 1:
				colors[2] = blendMode1(colors[0], colors[1])
			case 2:
				colors[2] = vram.ReadPalTexture(palAddr + 4)
				colors[3] = vram.ReadPalTexture(palAddr + 6)
			case 3:
				colors[2] = blendMode3(colors[0], colors[1])
				colors[3] = blendMode3(colors[1], colors[0])
			}

			for j := range uint32(4) {
				pack := vram.ReadTexture(off)
				off++
				for i := range uint32(4) {
					k := ((y+j)<<tex.PitchShift + (x + i))
					tex := (pack >> uint(i*2)) & 3

					if tex == 3 && mode <= 1 {
						out[k] = gl.Transparent
						continue
					}

					out[k] = gl.MakeColorFrom15Bit(
						uint8(colors[tex]&0x1F),
						uint8(colors[tex]>>5)&0x1F,
						uint8(colors[tex]>>10)&0x1F,
					)
				}
			}
		}
	}

	return &out
}

func blendMode1(a, b uint16) uint16 {
	aR := uint16(a) & 0x1F
	aG := uint16(a>>5) & 0x1F
	aB := uint16(a>>10) & 0x1F

	bR := uint16(b) & 0x1F
	bG := uint16(b>>5) & 0x1F
	bB := uint16(b>>10) & 0x1F

	oR := (((aR + bR) / 2) & 0x1F)
	oG := (((aG + bG) / 2) & 0x1F) << 5
	oB := (((aB + bB) / 2) & 0x1F) << 10

	return oR | oG | oB
}

func blendMode3(a, b uint16) uint16 {
	aR := uint16(a) & 0x1F
	aG := uint16(a>>5) & 0x1F
	aB := uint16(a>>10) & 0x1F

	bR := uint16(b) & 0x1F
	bG := uint16(b>>5) & 0x1F
	bB := uint16(b>>10) & 0x1F

	oR := (((aR*5 + bR*3) / 8) & 0x1F)
	oG := (((aG*5 + bG*3) / 8) & 0x1F) << 5
	oB := (((aB*5 + bB*3) / 8) & 0x1F) << 10

	return oR | oG | oB
}
