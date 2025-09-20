package rast

func (r *Rasterizer) DebugTexture(format uint8) {

    if format != 7 {
        panic("DEBUG TEXTURE UNKNOWN TYPE")
    }

    tex := &r.GeoEngine.Texture

    for i := uint32(0); i < tex.SizeS * tex.SizeT * 2; i += 2 {  

        data := uint32(r.VRAM.ReadTexture(i+0))
        data |= uint32(r.VRAM.ReadTexture(i+1)) << 8

        x := (i >> 1) % tex.SizeS
        y := (i >> 1) / tex.SizeS

        idx := x + (y * WIDTH)

        r.Render.PixelPalettes[idx] = data
    }
}
