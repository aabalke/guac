//go:build rc
package ppu

import simd "simd/archsimd"

var bit5 = simd.BroadcastUint32x16(0x1F)

func Convert15To24Bit(in simd.Uint32x16) simd.Uint32x16 {
    out := in.And(bit5)
    in = in.ShiftAllRight(5)
    out = out.Or(in.And(bit5).ShiftAllLeft(8))
    in = in.ShiftAllRight(5)
    out = out.Or(in.And(bit5).ShiftAllLeft(16))
    return out
}

func Convert24To15Bit(in simd.Uint32x16) simd.Uint32x16 {
    out := in.And(bit5)
    in = in.ShiftAllRight(8)
    out = out.Or(in.And(bit5).ShiftAllLeft(5))
    in = in.ShiftAllRight(8)
    out = out.Or(in.And(bit5).ShiftAllLeft(10))
    return out
}
