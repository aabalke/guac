package cpu

func (c *Cpu) swi(op uint32) {
	//op &= 0xFFFFFF

	//switch op {
	//case 0xB:
	//	c.CpuSet()
	//	return
	//}

	c.Exception(VEC_SWI, MODE_SWI)
}

//// doesn't seem to speed up
//func (c *Cpu) CpuSet() {
//	var (
//		r      = &c.Reg.R
//		rs     = r[0]
//		rd     = r[1]
//		cnt    = r[2] & 0x1FFFFF
//		fill   = (r[2]>>24)&1 != 0
//		isWord = (r[2]>>26)&1 != 0
//	)
//
//	if rs < 0x200_0000 {
//		return
//	}
//
//	// not cycle accurate
//
//	if fill {
//		if isWord {
//
//			v := c.Read32(rs)
//
//			for range cnt {
//				c.Write32(rd, v)
//				rd += 4
//			}
//
//			r[0] += 4
//
//		} else {
//
//			v := c.Read16(rs)
//			is := (rs & 1) << 3
//			v = bits.RotateLeft32(v, -int(is))
//
//			for range cnt {
//				c.Write16(rd, uint16(v))
//				rd += 2
//			}
//
//			r[0] += 2
//		}
//
//		r[1] = rd
//	} else {
//		srcPtr := c.gba.Mem.ReadPtr(rs)
//		dstPtr := c.gba.Mem.ReadPtr(rd)
//
//		if isWord {
//			for range cnt {
//
//				var v uint32
//				if srcPtr == nil {
//					v = c.Read32(rs)
//				} else {
//					v = *(*uint32)(srcPtr)
//					srcPtr = unsafe.Add(srcPtr, 4)
//				}
//
//				if dstPtr == nil {
//					c.Write32(rd, v)
//				} else {
//					*(*uint32)(dstPtr) = v
//					dstPtr = unsafe.Add(dstPtr, 4)
//				}
//
//				rs += 4
//				rd += 4
//			}
//		} else {
//			for range cnt {
//
//				var v uint32
//				if srcPtr == nil {
//					v = c.Read16(rs)
//					is := (rs & 1) << 3
//					v = bits.RotateLeft32(v, -int(is))
//				} else {
//					v = *(*uint32)(srcPtr)
//					srcPtr = unsafe.Add(srcPtr, 2)
//				}
//
//				if dstPtr == nil {
//					c.Write16(rd, uint16(v))
//				} else {
//					*(*uint16)(dstPtr) = uint16(v)
//					dstPtr = unsafe.Add(dstPtr, 2)
//				}
//
//				rs += 2
//				rd += 2
//			}
//		}
//		r[0] = rs
//		r[1] = rd
//	}
//}
