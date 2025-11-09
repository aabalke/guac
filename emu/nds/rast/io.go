package rast

import (
	"fmt"

	"github.com/aabalke/guac/emu/nds/utils"
)

func (r *Rasterizer) Read(addr uint32) uint8 {

    if addr >= 0x358 && addr < 0x380 {
        return r.ReadFog(addr)
    }

    switch {
    case addr < 0x620:
        // fall
    case addr < 0x630:
        return r.ReadPosTest(addr)
    case addr < 0x640:
        return r.ReadVecTest(addr)
    case addr < 0x680:
        return r.ReadClipMtx(addr)
    case addr < 0x6A4:
        return r.ReadVecMtx(addr)
    }

    //if addr & 0b11 == 0 { fmt.Printf("R ADDR %08X\n", addr) }
	switch addr {
	case 0x60:
		return r.GeoEngine.Disp3dCnt.Read(0)
	case 0x61:
		return r.GeoEngine.Disp3dCnt.Read(1)
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
            return uint8(min(2048, len(r.GeoEngine.Buffers.A.Polys)))
        } 

        return uint8(min(2048, len(r.GeoEngine.Buffers.B.Polys)))

    case 0x605:

        if r.GeoEngine.Buffers.BisRendering {
            return uint8(min(2048, len(r.GeoEngine.Buffers.A.Polys)) >> 8)
        } 

        return uint8(min(2048, len(r.GeoEngine.Buffers.B.Polys)) >> 8)

    case 0x606:

        polys := &r.GeoEngine.Buffers.B.Polys

        if r.GeoEngine.Buffers.BisRendering {
            polys = &r.GeoEngine.Buffers.A.Polys
        } 

        vertCnt := 0

        for _, v := range *polys {
            vertCnt += len(v.Vertices)
        }

        return uint8(min(6144, vertCnt))

    case 0x607:

        polys := &r.GeoEngine.Buffers.B.Polys

        if r.GeoEngine.Buffers.BisRendering {
            polys = &r.GeoEngine.Buffers.A.Polys
        } 

        vertCnt := 0

        for _, v := range *polys {
            vertCnt += len(v.Vertices)
        }
        return uint8(min(6144, vertCnt) >> 8)
	}

    //fmt.Printf("READ UNSETUP 3D IO %08X\n", addr)

    return 0

    panic(fmt.Sprintf("READ UNSETUP 3D IO %08X\n", addr))
}

func (r *Rasterizer) ReadPosTest(addr uint32) uint8 {

    d := &r.GeoEngine.PosTestData

    switch addr {
    case 0x620: return uint8(d[0] >> 0)
    case 0x621: return uint8(d[0] >> 8)
    case 0x622: return uint8(d[0] >> 16)
    case 0x623: return uint8(d[0] >> 24)
    case 0x624: return uint8(d[1] >> 0)
    case 0x625: return uint8(d[1] >> 8)
    case 0x626: return uint8(d[1] >> 16)
    case 0x627: return uint8(d[1] >> 24)
    case 0x628: return uint8(d[2] >> 0)
    case 0x629: return uint8(d[2] >> 8)
    case 0x62A: return uint8(d[2] >> 16)
    case 0x62B: return uint8(d[2] >> 24)
    case 0x62C: return uint8(d[3] >> 0)
    case 0x62D: return uint8(d[3] >> 8)
    case 0x62E: return uint8(d[3] >> 16)
    case 0x62F: return uint8(d[3] >> 24)
    }

    return 0
}

func (r *Rasterizer) ReadVecTest(addr uint32) uint8 {

    d := &r.GeoEngine.VecTestData

    switch addr {
    case 0x630: return uint8(d[0] >> 0)
    case 0x631: return uint8(d[0] >> 8)
    case 0x632: return uint8(d[1] >> 0)
    case 0x633: return uint8(d[1] >> 8)
    case 0x634: return uint8(d[2] >> 0)
    case 0x635: return uint8(d[2] >> 8)
    }

    return 0
}

func (r *Rasterizer) ReadClipMtx(addr uint32) uint8 {

    mtx := &r.GeoEngine.ClipMatrix

    switch addr {
    case 0x640: return uint8(utils.ConvertFromFloat(mtx.X00, 12) >> 0)
    case 0x641: return uint8(utils.ConvertFromFloat(mtx.X00, 12) >> 8) 
    case 0x642: return uint8(utils.ConvertFromFloat(mtx.X00, 12) >> 16)
    case 0x643: return uint8(utils.ConvertFromFloat(mtx.X00, 12) >> 24)
    case 0x644: return uint8(utils.ConvertFromFloat(mtx.X01, 12) >> 0)
    case 0x645: return uint8(utils.ConvertFromFloat(mtx.X01, 12) >> 8) 
    case 0x646: return uint8(utils.ConvertFromFloat(mtx.X01, 12) >> 16)
    case 0x647: return uint8(utils.ConvertFromFloat(mtx.X01, 12) >> 24)
    case 0x648: return uint8(utils.ConvertFromFloat(mtx.X02, 12) >> 0)
    case 0x649: return uint8(utils.ConvertFromFloat(mtx.X02, 12) >> 8) 
    case 0x64A: return uint8(utils.ConvertFromFloat(mtx.X02, 12) >> 16)
    case 0x64B: return uint8(utils.ConvertFromFloat(mtx.X02, 12) >> 24)
    case 0x64C: return uint8(utils.ConvertFromFloat(mtx.X03, 12) >> 0)
    case 0x64D: return uint8(utils.ConvertFromFloat(mtx.X03, 12) >> 8) 
    case 0x64E: return uint8(utils.ConvertFromFloat(mtx.X03, 12) >> 16)
    case 0x64F: return uint8(utils.ConvertFromFloat(mtx.X03, 12) >> 24)

    case 0x650: return uint8(utils.ConvertFromFloat(mtx.X10, 12) >> 0)
    case 0x651: return uint8(utils.ConvertFromFloat(mtx.X10, 12) >> 8) 
    case 0x652: return uint8(utils.ConvertFromFloat(mtx.X10, 12) >> 16)
    case 0x653: return uint8(utils.ConvertFromFloat(mtx.X10, 12) >> 24)
    case 0x654: return uint8(utils.ConvertFromFloat(mtx.X11, 12) >> 0)
    case 0x655: return uint8(utils.ConvertFromFloat(mtx.X11, 12) >> 8) 
    case 0x656: return uint8(utils.ConvertFromFloat(mtx.X11, 12) >> 16)
    case 0x657: return uint8(utils.ConvertFromFloat(mtx.X11, 12) >> 24)
    case 0x658: return uint8(utils.ConvertFromFloat(mtx.X12, 12) >> 0)
    case 0x659: return uint8(utils.ConvertFromFloat(mtx.X12, 12) >> 8) 
    case 0x65A: return uint8(utils.ConvertFromFloat(mtx.X12, 12) >> 16)
    case 0x65B: return uint8(utils.ConvertFromFloat(mtx.X12, 12) >> 24)
    case 0x65C: return uint8(utils.ConvertFromFloat(mtx.X13, 12) >> 0)
    case 0x65D: return uint8(utils.ConvertFromFloat(mtx.X13, 12) >> 8) 
    case 0x65E: return uint8(utils.ConvertFromFloat(mtx.X13, 12) >> 16)
    case 0x65F: return uint8(utils.ConvertFromFloat(mtx.X13, 12) >> 24)

    case 0x660: return uint8(utils.ConvertFromFloat(mtx.X20, 12) >> 0)
    case 0x661: return uint8(utils.ConvertFromFloat(mtx.X20, 12) >> 8) 
    case 0x662: return uint8(utils.ConvertFromFloat(mtx.X20, 12) >> 16)
    case 0x663: return uint8(utils.ConvertFromFloat(mtx.X20, 12) >> 24)
    case 0x664: return uint8(utils.ConvertFromFloat(mtx.X21, 12) >> 0)
    case 0x665: return uint8(utils.ConvertFromFloat(mtx.X21, 12) >> 8) 
    case 0x666: return uint8(utils.ConvertFromFloat(mtx.X21, 12) >> 16)
    case 0x667: return uint8(utils.ConvertFromFloat(mtx.X21, 12) >> 24)
    case 0x668: return uint8(utils.ConvertFromFloat(mtx.X22, 12) >> 0)
    case 0x669: return uint8(utils.ConvertFromFloat(mtx.X22, 12) >> 8) 
    case 0x66A: return uint8(utils.ConvertFromFloat(mtx.X22, 12) >> 16)
    case 0x66B: return uint8(utils.ConvertFromFloat(mtx.X22, 12) >> 24)
    case 0x66C: return uint8(utils.ConvertFromFloat(mtx.X23, 12) >> 0)
    case 0x66D: return uint8(utils.ConvertFromFloat(mtx.X23, 12) >> 8) 
    case 0x66E: return uint8(utils.ConvertFromFloat(mtx.X23, 12) >> 16)
    case 0x66F: return uint8(utils.ConvertFromFloat(mtx.X23, 12) >> 24)

    case 0x670: return uint8(utils.ConvertFromFloat(mtx.X30, 12) >> 0)
    case 0x671: return uint8(utils.ConvertFromFloat(mtx.X30, 12) >> 8) 
    case 0x672: return uint8(utils.ConvertFromFloat(mtx.X30, 12) >> 16)
    case 0x673: return uint8(utils.ConvertFromFloat(mtx.X30, 12) >> 24)
    case 0x674: return uint8(utils.ConvertFromFloat(mtx.X31, 12) >> 0)
    case 0x675: return uint8(utils.ConvertFromFloat(mtx.X31, 12) >> 8) 
    case 0x676: return uint8(utils.ConvertFromFloat(mtx.X31, 12) >> 16)
    case 0x677: return uint8(utils.ConvertFromFloat(mtx.X31, 12) >> 24)
    case 0x678: return uint8(utils.ConvertFromFloat(mtx.X32, 12) >> 0)
    case 0x679: return uint8(utils.ConvertFromFloat(mtx.X32, 12) >> 8) 
    case 0x67A: return uint8(utils.ConvertFromFloat(mtx.X32, 12) >> 16)
    case 0x67B: return uint8(utils.ConvertFromFloat(mtx.X32, 12) >> 24)
    case 0x67C: return uint8(utils.ConvertFromFloat(mtx.X33, 12) >> 0)
    case 0x67D: return uint8(utils.ConvertFromFloat(mtx.X33, 12) >> 8) 
    case 0x67E: return uint8(utils.ConvertFromFloat(mtx.X33, 12) >> 16)
    case 0x67F: return uint8(utils.ConvertFromFloat(mtx.X33, 12) >> 24)
    }
    panic(fmt.Sprintf("CLIP MTX READ FROM NON CLIP MTX ADDR %08X", addr))
}

func (r *Rasterizer) ReadVecMtx(addr uint32) uint8 {

    mtx := &r.GeoEngine.MtxStacks.Stacks[2].CurrMtx

    switch addr {
    case 0x680: return uint8(utils.ConvertFromFloat(mtx.X00, 12) >> 0)
    case 0x681: return uint8(utils.ConvertFromFloat(mtx.X00, 12) >> 8) 
    case 0x682: return uint8(utils.ConvertFromFloat(mtx.X00, 12) >> 16)
    case 0x683: return uint8(utils.ConvertFromFloat(mtx.X00, 12) >> 24)
    case 0x684: return uint8(utils.ConvertFromFloat(mtx.X01, 12) >> 0)
    case 0x685: return uint8(utils.ConvertFromFloat(mtx.X01, 12) >> 8) 
    case 0x686: return uint8(utils.ConvertFromFloat(mtx.X01, 12) >> 16)
    case 0x687: return uint8(utils.ConvertFromFloat(mtx.X01, 12) >> 24)
    case 0x688: return uint8(utils.ConvertFromFloat(mtx.X02, 12) >> 0)
    case 0x689: return uint8(utils.ConvertFromFloat(mtx.X02, 12) >> 8) 
    case 0x68A: return uint8(utils.ConvertFromFloat(mtx.X02, 12) >> 16)
    case 0x68B: return uint8(utils.ConvertFromFloat(mtx.X02, 12) >> 24)
    case 0x68C: return uint8(utils.ConvertFromFloat(mtx.X10, 12) >> 0)
    case 0x68D: return uint8(utils.ConvertFromFloat(mtx.X10, 12) >> 8) 
    case 0x68E: return uint8(utils.ConvertFromFloat(mtx.X10, 12) >> 16)
    case 0x68F: return uint8(utils.ConvertFromFloat(mtx.X10, 12) >> 24)

    case 0x690: return uint8(utils.ConvertFromFloat(mtx.X11, 12) >> 0)
    case 0x691: return uint8(utils.ConvertFromFloat(mtx.X11, 12) >> 8) 
    case 0x692: return uint8(utils.ConvertFromFloat(mtx.X11, 12) >> 16)
    case 0x693: return uint8(utils.ConvertFromFloat(mtx.X11, 12) >> 24)
    case 0x694: return uint8(utils.ConvertFromFloat(mtx.X12, 12) >> 0)
    case 0x695: return uint8(utils.ConvertFromFloat(mtx.X12, 12) >> 8) 
    case 0x696: return uint8(utils.ConvertFromFloat(mtx.X12, 12) >> 16)
    case 0x697: return uint8(utils.ConvertFromFloat(mtx.X12, 12) >> 24)
    case 0x698: return uint8(utils.ConvertFromFloat(mtx.X20, 12) >> 0)
    case 0x699: return uint8(utils.ConvertFromFloat(mtx.X20, 12) >> 8) 
    case 0x69A: return uint8(utils.ConvertFromFloat(mtx.X20, 12) >> 16)
    case 0x69B: return uint8(utils.ConvertFromFloat(mtx.X20, 12) >> 24)
    case 0x69C: return uint8(utils.ConvertFromFloat(mtx.X21, 12) >> 0)
    case 0x69D: return uint8(utils.ConvertFromFloat(mtx.X21, 12) >> 8) 
    case 0x69E: return uint8(utils.ConvertFromFloat(mtx.X21, 12) >> 16)
    case 0x69F: return uint8(utils.ConvertFromFloat(mtx.X21, 12) >> 24)

    case 0x6A0: return uint8(utils.ConvertFromFloat(mtx.X22, 12) >> 0)
    case 0x6A1: return uint8(utils.ConvertFromFloat(mtx.X22, 12) >> 8) 
    case 0x6A2: return uint8(utils.ConvertFromFloat(mtx.X22, 12) >> 16)
    case 0x6A3: return uint8(utils.ConvertFromFloat(mtx.X22, 12) >> 24)
    }

    panic(fmt.Sprintf("VEC MTX READ FROM NON VEC MTX ADDR %08X", addr))
}

func (r *Rasterizer) Write(addr uint32, v uint8) {

    switch {
    case addr >= 0x350 && addr < 0x358:
        r.RearPlane.Write(addr, v)
        return
    case addr >= 0x380 && addr < 0x3C0:
        WriteToonTbl(&r.GeoEngine.ToonTbl, addr, v)
        return
    case addr >= 0x358 && addr < 0x380:
        r.WriteFog(addr, v)
        return
    case addr >= 0x330 && addr < 0x340:
        r.Edge.Write(addr, v)
        return
    }

	switch addr {
	case 0x60:
		r.GeoEngine.Disp3dCnt.Write(v, 0)
	case 0x61:

        prevRear := r.GeoEngine.Disp3dCnt.RearPlaneBitmapEnabled

		r.GeoEngine.Disp3dCnt.Write(v, 1)

        if r.GeoEngine.Disp3dCnt.RearPlaneBitmapEnabled && !prevRear {
            r.RearPlane.Cache()
        }


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

    case 0x610:
        r.Disp1Dot.param &^= 0xFF
        r.Disp1Dot.param |= uint16(v)
        r.Disp1Dot.V = float64(r.Disp1Dot.param) / 8

    case 0x611:
        v &= 0b0111_1111
        r.Disp1Dot.param &^= 0xFF << 8
        r.Disp1Dot.param |= uint16(v) << 8
        r.Disp1Dot.V = float64(r.Disp1Dot.param) / 8

    default:
        //fmt.Printf("WRITE UNSETUP 3D IO %08X\n", addr)
        //panic(fmt.Sprintf("WRITES UNSETUP 3D IO %08X %02X\n", addr, v))
	}
}

func (r *Rasterizer) WriteFog(addr uint32, v uint8) {

    f := &r.GeoEngine.Fog

    if addr >= 0x360 && addr < 0x380 {
        addr -= 0x360
        f.Density[addr] = v & 0b0111_1111
        return
    }

    switch addr {
    case 0x358:
        f.Color = Convert15BitByte(f.Color, v, false)
    case 0x359:
        f.Color = Convert15BitByte(f.Color, v, true)
    case 0x35A:
        f.Color.A = float64(v & 0x1F) / 0x1F

    case 0x35C:
        f.Offset &^= 0xFF
        f.Offset |= uint16(v)
        f.UpdateBoundaries()

    case 0x35D:
        v &= 0b0111_1111
        f.Offset &^= 0xFF << 8
        f.Offset |= uint16(v) << 8
        f.UpdateBoundaries()
    }
}

func (r *Rasterizer) ReadFog(addr uint32) uint8 {

    f := &r.GeoEngine.Fog

    if addr >= 0x360 && addr < 0x380 {
        addr -= 0x360
        return f.Density[addr]
    }

    switch addr {
    case 0x35A:
        return uint8(f.Color.A * 0x1F)

    case 0x35C:
        return uint8(f.Offset)

    case 0x35D:
        return uint8(f.Offset >> 8)
    }

    return 0
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
