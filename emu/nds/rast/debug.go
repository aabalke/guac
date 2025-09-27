package rast

func (r *Rasterizer) DebugTexture() {

    tex := &r.GeoEngine.Texture

    switch tex.Format {
    case 5:
        r.DebugTexture4x4()
        return
    case 7:
    default:
        return
    }

    for i := uint32(0); i < tex.SizeS * tex.SizeT * 2; i += 2 {  

        data := uint32(r.VRAM.ReadTexture(i+0))
        data |= uint32(r.VRAM.ReadTexture(i+1)) << 8

        x := (i >> 1) % tex.SizeS
        y := (i >> 1) / tex.SizeS

        idx := x + (y * WIDTH)

        r.Render.PixelPalettes[idx] = data
        r.Render.AlphaPixels[idx] = false
    }
}

func (r *Rasterizer) DebugTexture4x4() {

    //tex := &r.GeoEngine.Texture

    idx := 10

    if len(r.GeoEngine.Buffers.GetPolygons()) <= idx {
        return
    }

    tex := &r.GeoEngine.Buffers.GetPolygons()[idx].Texture

    for x := range tex.SizeT {
        for y := range tex.SizeS {

            vBase := tex.VramOffset
            pBase := tex.PaletteBaseAddr * 0x10

            data := r.getColor4x4Debug(vBase, pBase, tex.SizeS, x, y)
            i := x + y * WIDTH
            r.Render.PixelPalettes[i] = uint32(data)
            r.Render.AlphaPixels[i] = true
        }
    }
}

func (r *Rasterizer) getColor4x4Debug(vBase, pBase, w, x, y uint32) uint16 {

    t := r
    slot1Base := uint32(0x2_0000)

    slot0Base := vBase
    palBase := pBase

    blockX, blockY := x   >> 2, y   >> 2
    texelX, texelY := x & 0b11, y & 0b11

    blockIdx := uint32(blockY*(w/4) + blockX)

    blockData := uint32(t.VRAM.ReadTexture(slot0Base + blockIdx * 4))
    blockData |= uint32(t.VRAM.ReadTexture(slot0Base + blockIdx * 4 + 1)) << 8
    blockData |= uint32(t.VRAM.ReadTexture(slot0Base + blockIdx * 4 + 2)) << 16
    blockData |= uint32(t.VRAM.ReadTexture(slot0Base + blockIdx * 4 + 3)) << 24
    rowBits := (blockData >> (texelY*8)) & 0xFF
    texelVal := (rowBits >> (texelX*2)) & 0b11

    palInfo := uint32(t.VRAM.ReadTexture(slot1Base + blockIdx*2))
    palInfo |= uint32(t.VRAM.ReadTexture(slot1Base + blockIdx*2 + 1)) << 8
    palOffset := (palInfo & 0x3FFF) * 4
    mode := (palInfo >> 14) & 0b11

    color0 := uint16(t.VRAM.ReadPalTexture(palBase + palOffset + 0))
    color0 |= uint16(t.VRAM.ReadPalTexture(palBase + palOffset + 1)) << 8
    color1 := uint16(t.VRAM.ReadPalTexture(palBase + palOffset + 2))
    color1 |= uint16(t.VRAM.ReadPalTexture(palBase + palOffset + 3)) << 8
    color2 := uint16(t.VRAM.ReadPalTexture(palBase + palOffset + 4))
    color2 |= uint16(t.VRAM.ReadPalTexture(palBase + palOffset + 5)) << 8
    color3 := uint16(t.VRAM.ReadPalTexture(palBase + palOffset + 6))
    color3 |= uint16(t.VRAM.ReadPalTexture(palBase + palOffset + 7)) << 8

    blendMode1 := func (a, b uint16) uint16 {

        aR := uint16(a) & 0b11111
        aG := uint16(a >> 5) & 0b11111
        aB := uint16(a >> 10)& 0b11111

        bR := uint16(b) & 0b11111
        bG := uint16(b >> 5) & 0b11111
        bB := uint16(b >> 10)& 0b11111

        oR := (((aR + bR) / 2) & 0b11111)
        oG := (((aG + bG) / 2) & 0b11111) << 5
        oB := (((aB + bB) / 2) & 0b11111) << 10

        return oR | oG | oB
    }

    blendMode3 := func (a, b uint16) uint16 {

        aR := uint16(a) & 0b11111
        aG := uint16(a >> 5) & 0b11111
        aB := uint16(a >> 10)& 0b11111

        bR := uint16(b) & 0b11111
        bG := uint16(b >> 5) & 0b11111
        bB := uint16(b >> 10)& 0b11111

        oR := (((aR * 5 + bR * 3) / 8) & 0b11111)
        oG := (((aG * 5 + bG * 3) / 8) & 0b11111) << 5
        oB := (((aB * 5 + bB * 3) / 8) & 0b11111) << 10

        return oR | oG | oB
    }

    switch texelVal {
    case 0:
        return color0
    case 1:
        return color1
    case 2:
        switch mode {
        case 1:
            return blendMode1(color0, color1)
        case 3:
            return blendMode3(color0, color1)
        default:
            return color2
        }

    case 3:
        switch mode {
        case 2:
            return color3
        case 3:
            return blendMode3(color1, color0)
        default:
            return 0
        }
    }

    panic("UNKNOWN TEXEL VALUE OR MODE")
}
