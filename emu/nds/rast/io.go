package rast

import (
	"fmt"
	"unsafe"

	"github.com/aabalke/guac/emu/nds/rast/gl"
	"github.com/aabalke/guac/utils"
)

func (r *Rasterizer) Read(addr uint32) uint8 {
	if addr >= 0x358 && addr < 0x380 {
		return r.ReadFog(addr)
	}

	switch {
	case addr < 0x620:
		// fall
	case addr < 0x630:
		addr -= 0x620
		return uint8(r.GeoEngine.PosTestData[addr/4] >> ((addr & 3) * 8))
	case addr < 0x640:
		return r.ReadVecTest(addr)
	case addr < 0x680:
		return r.ReadClipMtx(addr)
	case addr < 0x6A4:
		return r.ReadVecMtx(addr)
	}

	//if addr & 0b11 == 0 { fmt.Printf("R ADDR %08X\n", addr) }
	switch addr {
	case 0x60, 0x61:
		return r.GeoEngine.Disp3dCnt.Read(uint8(addr & 1))
	case 0x62, 0x63:
		return 0

	case 0x600, 0x601, 0x602, 0x603:
		return r.GeoEngine.GxStat.Read(addr & 3)
	case 0x604:

		buf := &r.GeoEngine.Buffers.A
		if r.GeoEngine.Buffers.BisRendering {
			buf = &r.GeoEngine.Buffers.B
		}

		poly, _ := buf.GetCnts()
		return uint8(poly)

	case 0x605:

		buf := &r.GeoEngine.Buffers.A
		if r.GeoEngine.Buffers.BisRendering {
			buf = &r.GeoEngine.Buffers.B
		}

		poly, _ := buf.GetCnts()
		return uint8(poly >> 8)

	case 0x606:

		buf := &r.GeoEngine.Buffers.A
		if r.GeoEngine.Buffers.BisRendering {
			buf = &r.GeoEngine.Buffers.B
		}

		_, vert := buf.GetCnts()
		return uint8(vert)

	case 0x607:

		buf := &r.GeoEngine.Buffers.A
		if r.GeoEngine.Buffers.BisRendering {
			buf = &r.GeoEngine.Buffers.B
		}

		_, vert := buf.GetCnts()
		return uint8(vert >> 8)
	}

	return 0
}

func (r *Rasterizer) ReadVecTest(addr uint32) uint8 {
	d := &r.GeoEngine.VecTestData

	switch addr {
	case 0x630:
		return uint8(d[0] >> 0)
	case 0x631:
		return uint8(d[0] >> 8)
	case 0x632:
		return uint8(d[1] >> 0)
	case 0x633:
		return uint8(d[1] >> 8)
	case 0x634:
		return uint8(d[2] >> 0)
	case 0x635:
		return uint8(d[2] >> 8)
	default:
		return 0
	}
}

func (r *Rasterizer) ReadClipMtx(addr uint32) uint8 {
	arr := (*gl.MatrixArray)(unsafe.Pointer(&r.GeoEngine.ClipMatrix))
	i := addr - 0x680
	return uint8(utils.ConvertFromFloat(arr[i/4], 12) >> (i & 3))
}

func (r *Rasterizer) ReadVecMtx(addr uint32) uint8 {
	mtx := &r.GeoEngine.MtxStacks.Stacks[2].CurrMtx

	switch addr &^ 3 {
	case 0x680:
		return uint8(utils.ConvertFromFloat(mtx.X00, 12) >> (addr & 3))
	case 0x684:
		return uint8(utils.ConvertFromFloat(mtx.X01, 12) >> (addr & 3))
	case 0x688:
		return uint8(utils.ConvertFromFloat(mtx.X02, 12) >> (addr & 3))
	case 0x68C:
		return uint8(utils.ConvertFromFloat(mtx.X10, 12) >> (addr & 3))
	case 0x690:
		return uint8(utils.ConvertFromFloat(mtx.X11, 12) >> (addr & 3))
	case 0x694:
		return uint8(utils.ConvertFromFloat(mtx.X12, 12) >> (addr & 3))
	case 0x698:
		return uint8(utils.ConvertFromFloat(mtx.X20, 12) >> (addr & 3))
	case 0x69C:
		return uint8(utils.ConvertFromFloat(mtx.X21, 12) >> (addr & 3))
	case 0x6A0:
		return uint8(utils.ConvertFromFloat(mtx.X22, 12) >> (addr & 3))
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

	case 0x62, 0x63:
		return
	case 0x600, 0x601, 0x602, 0x603:
		r.GeoEngine.GxStat.Write(v, uint8(addr&3))

	case 0x610:
		r.Disp1Dot.param = (r.Disp1Dot.param & 0xFF00) | uint16(v)
		r.Disp1Dot.V = float64(r.Disp1Dot.param) / 8

	case 0x611:
		r.Disp1Dot.param = (r.Disp1Dot.param & 0xFF) | (uint16(v&0x7F) << 8)
		r.Disp1Dot.V = float64(r.Disp1Dot.param) / 8
	}
}

func (r *Rasterizer) WriteFog(addr uint32, v uint8) {
	f := &r.GeoEngine.Fog

	if addr >= 0x360 && addr < 0x380 {
		f.Density[addr-0x360] = v & 0x7F
		return
	}

	switch addr {
	case 0x358:
		f.Color = Convert15BitByte(f.Color, v, false)
	case 0x359:
		f.Color = Convert15BitByte(f.Color, v, true)
	case 0x35A:
		f.Color.A = float32(v&0x1F) / 0x1F

	case 0x35C:
		f.Offset &^= 0xFF
		f.Offset |= uint16(v)
		f.UpdateBoundaries()

	case 0x35D:
		v &= 0x7F
		f.Offset &^= 0xFF << 8
		f.Offset |= uint16(v) << 8
		f.UpdateBoundaries()
	}
}

func (r *Rasterizer) ReadFog(addr uint32) uint8 {
	f := &r.GeoEngine.Fog

	if addr >= 0x360 && addr < 0x380 {
		return f.Density[addr-0x360]
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

type addrRange struct {
	start, end, base uint32 // inclusive, addr steps by 4
}

var addrRanges = []addrRange{
	{0x440, 0x470, 0x10},
	{0x480, 0x4AC, 0x20},
	{0x4C0, 0x4D0, 0x30},
	{0x500, 0x504, 0x40},
	{0x540, 0x540, 0x50},
	{0x580, 0x580, 0x60},
	{0x5C0, 0x5C8, 0x70},
}

func (r *Rasterizer) GeoCmd(addr, v uint32) {
	d := &r.GeoEngine.Data

	if len(*d) == 0 {

		addr &= 0xFF_FFFF

		for _, r := range addrRanges {
			if addr >= r.start && addr <= r.end {
				v := r.base + ((addr - r.start) / 4)
				*d = append(*d, v)
				break
			}
		}
	}

	(*d) = append(*d, v)

	r.GeoEngine.Cmd(false, *d)
}
