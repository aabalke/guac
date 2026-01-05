//go:build rc
package ppu

import (
    "unsafe"
	simd "simd/archsimd"
)

// func (m *MasterBright) ApplySIMD(palettes *[SCREEN_WIDTH]uint32, outPtr unsafe.Pointer) *[SCREEN_WIDTH]uint32 {
func (m *MasterBright) ApplySIMD(palettes *[SCREEN_WIDTH]uint32, outPtr unsafe.Pointer) {

	var (
		outVector = simd.BroadcastUint32x16(0x0000_0000)
		alpha     = simd.BroadcastUint32x16(0xFF00_0000)
		mask      = simd.BroadcastUint32x16(0x1F)
		threshold = simd.BroadcastUint16x32(0x1F)
		factor    = simd.BroadcastUint16x32(uint16(m.Factor))
		//out       = [SCREEN_WIDTH * 4]uint8{}
		//outPtr    = unsafe.Pointer(&out)
		palPtr = unsafe.Pointer(palettes)
	)

	for x := uint32(0); x < SCREEN_WIDTH; x += SCREEN_WIDTH / 16 {

		// clear
		outVector = simd.BroadcastUint32x16(0x0000_0000)
		pals := simd.LoadUint32x16((*[16]uint32)(palPtr))

		// r
		outVector = outVector.Or(pals.And(mask))

		// g
		pals = pals.ShiftAllRight(5)
		outVector = outVector.Or(pals.And(mask).ShiftAllLeft(8))

		// b
		pals = pals.ShiftAllRight(5)
		outVector = outVector.Or(pals.And(mask).ShiftAllLeft(16))

		// convert uint32s into channels uint8
		// extend to uint16 since uint8 func are limited
		hi := outVector.GetHi().AsUint8x32().ExtendToUint16()
		lo := outVector.GetLo().AsUint8x32().ExtendToUint16()

		switch m.Mode {
		case MB_UP:
			// ch += (31 - ch) * m.Factor >> 4
			hi = hi.Add(threshold.Sub(hi).Mul(factor).ShiftAllRight(4))
			lo = lo.Add(threshold.Sub(lo).Mul(factor).ShiftAllRight(4))

		case MB_DOWN:
			// ch -= ch * m.Factor >> 4
			hi = hi.Sub(hi.Mul(factor).ShiftAllRight(4))
			lo = lo.Sub(lo.Mul(factor).ShiftAllRight(4))
		}

		// 5bit to 8bit conversion
		// ch = (ch << 3) | (ch >> 2)
		hi = hi.ShiftAllLeft(3).Or(hi.ShiftAllRight(2))
		lo = lo.ShiftAllLeft(3).Or(lo.ShiftAllRight(2))

		// channels: uint16 -> uint8 -> uint32 pixel
		outVector = outVector.SetHi(hi.SaturateToUint8().AsUint32x8())
		outVector = outVector.SetLo(lo.SaturateToUint8().AsUint32x8())

		// alpha
		outVector = outVector.Or(alpha)

		outVector.Store(((*[16]uint32)(outPtr)))

		outPtr = unsafe.Add(outPtr, 16*4)
		palPtr = unsafe.Add(palPtr, 16*4)
	}

	//return (*[SCREEN_WIDTH]uint32)(unsafe.Pointer(&out))
}
