package rast

import "fmt"

func (r *Rasterizer) Read(addr uint32) uint8 {

    if addr >= 0x630 && addr < 0x636 {
        panic("READ VEC RESULT")
    }

    if addr >= 0x640 && addr < 0x6B0 {
        panic("READ CLIP MTX OR DIR MTX")
    }

    //if addr & 0b11 == 0 { fmt.Printf("R ADDR %08X\n", addr) }
	switch addr {
	case 0x60:
		return r.Disp3dCnt.Read(0)
	case 0x61:
		return r.Disp3dCnt.Read(1)
	case 0x62:
		return 0
	case 0x63:
		return 0
	case 0x600:
		return r.GeoEngine.GxStat.Read(0)
	case 0x601:
		return r.GeoEngine.GxStat.Read(1)
	case 0x602:
		return r.GeoEngine.GxStat.Read(2)
	case 0x603:
		return r.GeoEngine.GxStat.Read(3)
    case 0x604:

        if r.GeoEngine.Buffers.BisRendering {
            return uint8(min(2048, len(r.GeoEngine.Buffers.A)))
        } 

        return uint8(min(2048, len(r.GeoEngine.Buffers.B)))

    case 0x605:

        if r.GeoEngine.Buffers.BisRendering {
            return uint8(min(2048, len(r.GeoEngine.Buffers.A)) >> 8)
        } 

        return uint8(min(2048, len(r.GeoEngine.Buffers.B)) >> 8)

    case 0x606:

        polys := &r.GeoEngine.Buffers.B

        if r.GeoEngine.Buffers.BisRendering {
            polys = &r.GeoEngine.Buffers.A
        } 

        vertCnt := 0

        for _, v := range *polys {
            vertCnt += len(v.Vertices)
        }

        return uint8(min(6144, vertCnt))

    case 0x607:

        polys := &r.GeoEngine.Buffers.B

        if r.GeoEngine.Buffers.BisRendering {
            polys = &r.GeoEngine.Buffers.A
        } 

        vertCnt := 0

        for _, v := range *polys {
            vertCnt += len(v.Vertices)
        }

        return uint8(min(6144, vertCnt) >> 8)

    case 0x620: return uint8(r.GeoEngine.PosTestData[0] >> 0)
    case 0x621: return uint8(r.GeoEngine.PosTestData[0] >> 8)
    case 0x622: return uint8(r.GeoEngine.PosTestData[0] >> 16)
    case 0x623: return uint8(r.GeoEngine.PosTestData[0] >> 24)
    case 0x624: return uint8(r.GeoEngine.PosTestData[1] >> 0)
    case 0x625: return uint8(r.GeoEngine.PosTestData[1] >> 8)
    case 0x626: return uint8(r.GeoEngine.PosTestData[1] >> 16)
    case 0x627: return uint8(r.GeoEngine.PosTestData[1] >> 24)
    case 0x628: return uint8(r.GeoEngine.PosTestData[2] >> 0)
    case 0x629: return uint8(r.GeoEngine.PosTestData[2] >> 8)
    case 0x62A: return uint8(r.GeoEngine.PosTestData[2] >> 16)
    case 0x62B: return uint8(r.GeoEngine.PosTestData[2] >> 24)
    case 0x62C: return uint8(r.GeoEngine.PosTestData[3] >> 0)
    case 0x62D: return uint8(r.GeoEngine.PosTestData[3] >> 8)
    case 0x62E: return uint8(r.GeoEngine.PosTestData[3] >> 16)
    case 0x62F: return uint8(r.GeoEngine.PosTestData[3] >> 24)

	}

    //fmt.Printf("READ UNSETUP 3D IO %08X\n", addr)

    return 0

    panic(fmt.Sprintf("READ UNSETUP 3D IO %08X\n", addr))
}

func (r *Rasterizer) Write(addr uint32, v uint8) {

    if addr >= 0x330 && addr < 0x400 {
        return
    }

    //if addr & 0b11 == 0 { fmt.Printf("W ADDR %08X V %02X\n" ,addr ,v) }

    switch {
    case addr >= 0x350 && addr < 0x358:
        r.RearPlane.Write(addr, v)
        return
    }

	switch addr {
	case 0x60:
		r.Disp3dCnt.Write(v, 0)
	case 0x61:
		r.Disp3dCnt.Write(v, 1)
    case 0x62:
        return
    case 0x63:
        return
	case 0x600:
		r.GeoEngine.GxStat.Write(v, 0)
	case 0x601:
		r.GeoEngine.GxStat.Write(v, 1)
	case 0x602:
		r.GeoEngine.GxStat.Write(v, 2)
	case 0x603:
		r.GeoEngine.GxStat.Write(v, 3)

    default:
        //fmt.Printf("WRITE UNSETUP 3D IO %08X\n", addr)
        //panic(fmt.Sprintf("WRITES UNSETUP 3D IO %08X %02X\n", addr, v))
	}
}

func (r *Rasterizer) GeoCmdFifo(v uint32) {
    //fmt.Printf("WRITING FIFO CMD ADDR V %08X\n", v)
    r.GeoEngine.Fifo(v)
}

func (r *Rasterizer) GeoCmd(addr, v uint32) {

    d := &r.GeoEngine.Data

    addr &= 0xFF_FFFF

    //fmt.Printf("WRITING CMD %08X ADDR V %08X\n", addr, v)

    if len(*d) == 0 {
        switch addr {
        case 0x440: (*d) = append(*d, 0x10)
        case 0x444: (*d) = append(*d, 0x11)
        case 0x448: (*d) = append(*d, 0x12)
        case 0x44C: (*d) = append(*d, 0x13)
        case 0x450: (*d) = append(*d, 0x14)
        case 0x454: (*d) = append(*d, 0x15)
        case 0x458: (*d) = append(*d, 0x16)
        case 0x45C: (*d) = append(*d, 0x17)
        case 0x460: (*d) = append(*d, 0x18)
        case 0x464: (*d) = append(*d, 0x19)
        case 0x468: (*d) = append(*d, 0x1A)
        case 0x46C: (*d) = append(*d, 0x1B)
        case 0x470: (*d) = append(*d, 0x1C)
        case 0x480: (*d) = append(*d, 0x20)
        case 0x484: (*d) = append(*d, 0x21)
        case 0x488: (*d) = append(*d, 0x22)
        case 0x48C: (*d) = append(*d, 0x23)
        case 0x490: (*d) = append(*d, 0x24)
        case 0x494: (*d) = append(*d, 0x25)
        case 0x498: (*d) = append(*d, 0x26)
        case 0x49C: (*d) = append(*d, 0x27)
        case 0x4A0: (*d) = append(*d, 0x28)
        case 0x4A4: (*d) = append(*d, 0x29)
        case 0x4A8: (*d) = append(*d, 0x2A)
        case 0x4AC: (*d) = append(*d, 0x2B)
        case 0x4C0: (*d) = append(*d, 0x30)
        case 0x4C4: (*d) = append(*d, 0x31)
        case 0x4C8: (*d) = append(*d, 0x32)
        case 0x4CC: (*d) = append(*d, 0x33)
        case 0x4D0: (*d) = append(*d, 0x34)
        case 0x500: (*d) = append(*d, 0x40)
        case 0x504: (*d) = append(*d, 0x41)
        case 0x540: (*d) = append(*d, 0x50)
        case 0x580: (*d) = append(*d, 0x60)
        case 0x5C0: (*d) = append(*d, 0x70)
        case 0x5C4: (*d) = append(*d, 0x71)
        case 0x5C8: (*d) = append(*d, 0x72)
        }
    }

    (*d) = append(*d, v)

    r.GeoEngine.Cmd(false, *d)
}
