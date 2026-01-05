//go:build rc
package ppu

import (
    simd "simd/archsimd"
    "unsafe"
)

func ResetBlendPalettesSIMD(simd *Simd, backdrop uint32, bld *Blend) {

    simd.BlendModeUse = [5]bool{}
    simd.BlendMasks = [5][16]uint16{}

    clear(simd.BlendPalettes[:])

    for i := range SCREEN_WIDTH {

        bp := &simd.BlendPalettes[i]

        bp.NoBlendPalette = backdrop

        if bld.a[5] {
            bp.APalette = backdrop
            bp.hasA = true
            bp.targetATop = true
        }

        if bld.b[5] {
            bp.BPalette = backdrop
            bp.hasB = true
        }
    }
}

func (bp *BlendPalettes) GetBlendModeSIMD(simd *Simd, winBlend, objTrans bool, bld *Blend, x uint32) {

    switch {
    case !winBlend && bld.Mode == BLD_MODE_WHITE:
        simd.BlendModeUse[BLEND_WHITE] = true
        simd.BlendMasks[BLEND_WHITE][x / 16] |= 1 << (x & 0xF)
        return
    case !winBlend && bld.Mode == BLD_MODE_BLACK:
        simd.BlendModeUse[BLEND_BLACK] = true
        simd.BlendMasks[BLEND_BLACK][x / 16] |= 1 << (x & 0xF)
        return
    case (!winBlend && bld.Mode == BLD_MODE_STD) || objTrans:
        if !bp.hasA || !bp.hasB || !bp.targetATop || bp.alpha >= 1 {
            simd.BlendModeUse[BLEND_NONE] = true
            simd.BlendMasks[BLEND_NONE][x / 16] |= 1 << (x & 0xF)
            return
        }

        if bp.targetA3d {
            simd.BlendModeUse[BLEND_ALPHA_3D] = true
            simd.BlendMasks[BLEND_ALPHA_3D][x / 16] |= 1 << (x & 0xF)
            return
        }

        simd.BlendModeUse[BLEND_ALPHA] = true
        simd.BlendMasks[BLEND_ALPHA][x / 16] |= 1 << (x & 0xF)
        return
    }

    simd.BlendModeUse[BLEND_NONE] = true
    simd.BlendMasks[BLEND_NONE][x / 16] |= 1 << (x & 0xF)
    return
}

func BlendSIMD(simdData *Simd, bld *Blend) {

    if simdData.BlendModeUse[BLEND_NONE] {
        blend_none(simdData, bld)
    }

    if simdData.BlendModeUse[BLEND_ALPHA] {
        blend_alpha(simdData, bld)
    }

    if simdData.BlendModeUse[BLEND_ALPHA_3D] {
        blend_alpha_3d(simdData, bld)
    }

    if simdData.BlendModeUse[BLEND_WHITE] || simdData.BlendModeUse[BLEND_BLACK] {
        blend_grayscale(simdData, bld)
    }
}

//go:inline
func blend_none(simdData *Simd, bld *Blend) {

    bps := &simdData.BlendPalettes
    palPtr := unsafe.Pointer(&simdData.PalData)

	for x := uint32(0); x < SCREEN_WIDTH; x += 16 {

        mask := simdData.BlendMasks[BLEND_NONE][x / 16]
        if mask == 0 {
            palPtr = unsafe.Add(palPtr, 16 * 4)
            continue
        }

        var in [16]uint32
        for i := range uint32(16) {
            in[i] = bps[x+i].NoBlendPalette
        }

        maskVec := simd.Mask32x16FromBits(mask)
		noBlend := simd.LoadUint32x16(&in)
        noBlend.StoreMasked(((*[16]uint32)(palPtr)), maskVec)

        palPtr = unsafe.Add(palPtr, 16 * 4)
    }
}

//go:inline
func blend_grayscale(simdData *Simd, bld *Blend) {

	var (
        bps = &simdData.BlendPalettes
		top = simd.BroadcastUint16x32(0x1F)
		factor    = simd.BroadcastUint16x32(uint16(bld.yEv))

        aPals [SCREEN_WIDTH]uint32
	)

    for i := range uint32(SCREEN_WIDTH) {
        aPals[i] = bps[i].APalette
    }

    for _, blend := range [...]BlendType{BLEND_WHITE, BLEND_BLACK} {

        palPtr := unsafe.Pointer(&simdData.PalData)
        aPtr := unsafe.Pointer(&aPals)

        for x := uint32(0); x < SCREEN_WIDTH; x += 16 {

            mask := simdData.BlendMasks[blend][x / 16]
            if mask == 0 {
                palPtr = unsafe.Add(palPtr, 16 * 4)
                aPtr = unsafe.Add(aPtr, 16 * 4)
                continue
            }

            maskVec := simd.Mask32x16FromBits(mask)
            in := Convert15To24Bit(simd.LoadUint32x16((*[16]uint32)(aPtr)))

            // convert uint32s into channels uint8
            // extend to uint16 since uint8 func are limited
            hi := in.GetHi().AsUint8x32().ExtendToUint16()
            lo := in.GetLo().AsUint8x32().ExtendToUint16()

            if blend == BLEND_WHITE {
                // ch += (31 - ch) * m.Factor >> 4
                hi = hi.Add(top.Sub(hi).Mul(factor).ShiftAllRight(4))
                lo = lo.Add(top.Sub(lo).Mul(factor).ShiftAllRight(4))
            } else {
                // ch -= ch * m.Factor >> 4
                hi = hi.Sub(hi.Mul(factor).ShiftAllRight(4))
                lo = lo.Sub(lo.Mul(factor).ShiftAllRight(4))
            }

            // channels: uint16 -> uint8 -> uint32 pixel
            in = in.SetHi(hi.SaturateToUint8().AsUint32x8())
            in = in.SetLo(lo.SaturateToUint8().AsUint32x8())

            Convert24To15Bit(in).StoreMasked(((*[16]uint32)(palPtr)), maskVec)

            palPtr = unsafe.Add(palPtr, 16 * 4)
            aPtr = unsafe.Add(aPtr, 16 * 4)
        }
    }
}

//go:inline
func blend_alpha(simdData *Simd, bld *Blend) {

	var (
        bps = &simdData.BlendPalettes
		factorA    = simd.BroadcastUint16x32(uint16(bld.aEv))
		factorB    = simd.BroadcastUint16x32(uint16(bld.bEv))

        aPals [SCREEN_WIDTH]uint32
        bPals [SCREEN_WIDTH]uint32

        palPtr = unsafe.Pointer(&simdData.PalData)
        aPtr = unsafe.Pointer(&aPals)
        bPtr = unsafe.Pointer(&bPals)
	)

    for i := range uint32(SCREEN_WIDTH) {
        aPals[i] = bps[i].APalette
        bPals[i] = bps[i].BPalette
    }

	for x := uint32(0); x < SCREEN_WIDTH; x += 16 {

        mask := simdData.BlendMasks[BLEND_ALPHA][x / 16]
        if mask == 0 {
            palPtr  = unsafe.Add(palPtr, 16 * 4)
            aPtr = unsafe.Add(aPtr, 16 * 4)
            bPtr = unsafe.Add(bPtr, 16 * 4)
            continue
        }

        maskVec := simd.Mask32x16FromBits(mask)
        a := Convert15To24Bit(simd.LoadUint32x16((*[16]uint32)(aPtr)))
        b := Convert15To24Bit(simd.LoadUint32x16((*[16]uint32)(bPtr)))

		// convert uint32s into channels uint8
		// extend to uint16 since uint8 func are limited
		ahi := a.GetHi().AsUint8x32().ExtendToUint16()
		alo := a.GetLo().AsUint8x32().ExtendToUint16()
		bhi := b.GetHi().AsUint8x32().ExtendToUint16()
		blo := b.GetLo().AsUint8x32().ExtendToUint16()

        // ch a * aEv + b * bEv
        ahi = (ahi.Mul(factorA).Add(bhi.Mul(factorB))).ShiftAllRight(4)
        alo = (alo.Mul(factorA).Add(blo.Mul(factorB))).ShiftAllRight(4)

		// channels: uint16 -> uint8 -> uint32 pixel
        a = a.SetHi(ahi.SaturateToUint8().AsUint32x8())
		a = a.SetLo(alo.SaturateToUint8().AsUint32x8())

        outVector := Convert24To15Bit(a)

        outVector.StoreMasked(((*[16]uint32)(palPtr)), maskVec)

        palPtr  = unsafe.Add(palPtr, 16 * 4)
        aPtr = unsafe.Add(aPtr, 16 * 4)
        bPtr = unsafe.Add(bPtr, 16 * 4)
    }
}

//go:inline
func blend_alpha_3d(simdData *Simd, bld *Blend) {

    panic("untested 3d target blend simd")

	//var (
    //    bps = &simdData.BlendPalettes
	//	bit5      = simd.BroadcastUint32x16(0x1F)

    //    aPals [SCREEN_WIDTH]uint32
    //    bPals [SCREEN_WIDTH]uint32
    //    alpha [SCREEN_WIDTH]uint8
    //    alphaB [SCREEN_WIDTH]uint8

    //    palPtr = unsafe.Pointer(&simdData.PalData)
    //    aPtr = unsafe.Pointer(&aPals)
    //    bPtr = unsafe.Pointer(&bPals)

    //    alphaPtr = unsafe.Pointer(&alpha)
    //    alphaBPtr = unsafe.Pointer(&alphaB)
	//)

    //for i := range uint32(SCREEN_WIDTH) {
    //    aPals[i] = bps[i].APalette
    //    bPals[i] = bps[i].BPalette
    //    alpha[i] = uint8(bps[i].alpha * 16)
    //    alphaB[i] = uint8((1 - bps[i].alpha) * 16)
    //}

	//for x := uint32(0); x < SCREEN_WIDTH; x += 16 {

    //    mask := simdData.BlendMasks[BLEND_ALPHA_3D][x / 16]
    //    if mask == 0 {
    //        palPtr  = unsafe.Add(palPtr, 16 * 4)
    //        aPtr = unsafe.Add(aPtr, 16 * 4)
    //        bPtr = unsafe.Add(bPtr, 16 * 4)
    //        alphaPtr = unsafe.Add(alphaPtr, 16)
    //        alphaBPtr = unsafe.Add(alphaBPtr, 16)
    //        continue
    //    }

    //    maskVec := simd.Mask32x16FromBits(mask)
	//	apals := simd.LoadUint32x16((*[16]uint32)(aPtr))
	//	bpals := simd.LoadUint32x16((*[16]uint32)(bPtr))
	//	alphaVec := simd.LoadUint8x32((*[16]uint8)(alphaPtr))
	//	alphaBVec := simd.LoadUint8x32((*[16]uint8)(alphaBPtr))

    //    // convert
    //    aVector := Convert15To24Bit(apals)
    //    bVector := Convert15To24Bit(bpals)

    //    alphaVec = alphaVec.ExtendToUint16()
    //    alphaBVec = alphaBVec.ExtendToUint16()

    //    // calc

	//	// convert uint32s into channels uint8
	//	// extend to uint16 since uint8 func are limited
	//	ahi := aVector.GetHi().AsUint8x32().ExtendToUint16()
	//	alo := aVector.GetLo().AsUint8x32().ExtendToUint16()
	//	bhi := bVector.GetHi().AsUint8x32().ExtendToUint16()
	//	blo := bVector.GetLo().AsUint8x32().ExtendToUint16()

    //    // ch a * aEv + b * bEv
    //    ahi = ahi.Mul(alphaVec).ShiftAllRight(4)
    //    alo = alo.Mul(alphaVec).ShiftAllRight(4)

    //    ahi = ahi.Add(bhi.Mul(alphaBVec).ShiftAllRight(4))
    //    alo = alo.Add(blo.Mul(alphaBVec).ShiftAllRight(4))

	//	// channels: uint16 -> uint8 -> uint32 pixel
    //    aVector = aVector.SetHi(ahi.SaturateToUint8().AsUint32x8())
	//	aVector = aVector.SetLo(alo.SaturateToUint8().AsUint32x8())

    //    // convert
    //    outVector = Convert24To15Bit(aVector)

    //    outVector.StoreMasked(((*[16]uint32)(palPtr)), maskVec)

    //    palPtr  = unsafe.Add(palPtr, 16 * 4)
    //    aPtr = unsafe.Add(aPtr, 16 * 4)
    //    bPtr = unsafe.Add(bPtr, 16 * 4)
    //    alphaPtr = unsafe.Add(alphaPtr, 16)
    //    alphaBPtr = unsafe.Add(alphaBPtr, 16)
    //}
}
