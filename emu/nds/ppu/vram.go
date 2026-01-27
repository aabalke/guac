package ppu

import (
	"unsafe"

	"github.com/aabalke/guac/emu/nds/rast"
)

const (
	A = iota
	B
	C
	D
	E
	F
	G
	H
	I
)

type VRAM struct {
	a [0x2_0000]uint8
	b [0x2_0000]uint8
	c [0x2_0000]uint8
	d [0x2_0000]uint8
	e [0x1_0000]uint8
	f [0x0_4000]uint8
	g [0x0_4000]uint8
	h [0x0_8000]uint8
	i [0x0_4000]uint8

	Cnt   [9]VramCnt
	Cnt_7 uint8

	TextureSlots [4]*[0x2_0000]uint8
	TexPalSlots  [6]*[0x4000]uint8

	TextureCache *rast.TextureCache

	engineA *Engine
	engineB *Engine
}

func (v *VRAM) Init(t *rast.TextureCache, a, b *Engine) {

	v.engineA = a
	v.engineB = b
	v.TextureCache = t

	for i := range len(v.Cnt) {
		v.Cnt[i].Write(0x80)
	}

	v.Cnt[A].Size = 0x2_0000
	v.Cnt[B].Size = 0x2_0000
	v.Cnt[C].Size = 0x2_0000
	v.Cnt[D].Size = 0x2_0000
	v.Cnt[E].Size = 0x1_0000
	v.Cnt[F].Size = 0x0_4000
	v.Cnt[G].Size = 0x0_4000
	v.Cnt[H].Size = 0x0_8000
	v.Cnt[I].Size = 0x0_4000

	v.Cnt[A].bank = unsafe.Pointer(&v.a)
	v.Cnt[B].bank = unsafe.Pointer(&v.b)
	v.Cnt[C].bank = unsafe.Pointer(&v.c)
	v.Cnt[D].bank = unsafe.Pointer(&v.d)
	v.Cnt[E].bank = unsafe.Pointer(&v.e)
	v.Cnt[F].bank = unsafe.Pointer(&v.f)
	v.Cnt[G].bank = unsafe.Pointer(&v.g)
	v.Cnt[H].bank = unsafe.Pointer(&v.h)
	v.Cnt[I].bank = unsafe.Pointer(&v.i)
}

type VramCnt struct {
	V       uint8
	Mst     uint8
	Enabled bool
	Ofs     uint32
	Base    uint32
	Size    uint32
	bank    unsafe.Pointer

	arm7 bool
}

func (vc *VramCnt) Write(v uint8) {
	vc.V = v & 0b1001_1111
	vc.Mst = v & 0b111
	vc.Ofs = uint32(v>>3) & 0b11
	vc.Enabled = (v>>7)&1 != 0
}

func (vm *VRAM) WriteCnt(addr uint32, v uint8) {

	switch addr {
	case 0x240:

		cnt := &vm.Cnt[A]
		cnt.Write(v)
		cnt.Base = 0x100_0000

		switch cnt.Mst {
		case 0:
			cnt.Base = 0x80_0000
		case 1:
			cnt.Base = 0x20000 * cnt.Ofs
		case 2:
			cnt.Base = 0x400000 + 0x20000*cnt.Ofs
		case 3:
			vm.TextureCache.Reset()
			vm.TextureSlots[cnt.Ofs] = (*[0x20000]uint8)(cnt.bank)
		}

	case 0x241:

		cnt := &vm.Cnt[B]
		cnt.Write(v)
		cnt.Base = 0x100_0000

		switch cnt.Mst {
		case 0:
			cnt.Base = 0x82_0000
		case 1:
			cnt.Base = 0x20000 * cnt.Ofs
		case 2:
			cnt.Base = 0x400000 + 0x20000*cnt.Ofs
		case 3:
			vm.TextureCache.Reset()
			vm.TextureSlots[cnt.Ofs] = (*[0x20000]uint8)(cnt.bank)
		}

	case 0x242:

		cnt := &vm.Cnt[C]
		cnt.Write(v)
		cnt.Base = 0x100_0000

		cnt.arm7 = false
		vm.Cnt_7 &^= 1

		switch cnt.Mst {
		case 0:
			cnt.Base = 0x84_0000
		case 1:
			cnt.Base = 0x20000 * cnt.Ofs
		case 2:

			if cnt.Ofs >= 2 {
				panic("INVALID ARM7 Cnt C OFS")
			}

			cnt.arm7 = true
			vm.Cnt_7 |= 1

		case 3:
			vm.TextureCache.Reset()
			vm.TextureSlots[cnt.Ofs] = (*[0x20000]uint8)(cnt.bank)

		case 4:
			cnt.Base = 0x20_0000
		}

	case 0x243:

		cnt := &vm.Cnt[D]
		cnt.Write(v)
		cnt.Base = 0x100_0000

		cnt.arm7 = false
		vm.Cnt_7 &^= 0b10

		switch cnt.Mst {
		case 0:
			cnt.Base = 0x86_0000
		case 1:
			cnt.Base = 0x20000 * cnt.Ofs
		case 2:

			if cnt.Ofs >= 2 {
				panic("INVALID ARM7 Cnt D OFS")
			}

			cnt.arm7 = true
			vm.Cnt_7 |= 0b10

		case 3:
			vm.TextureCache.Reset()
			vm.TextureSlots[cnt.Ofs] = (*[0x20000]uint8)(cnt.bank)

		case 4:
			cnt.Base = 0x60_0000
		}

	case 0x244:

		cnt := &vm.Cnt[E]
		cnt.Write(v)
		cnt.Base = 0x100_0000

		switch cnt.Mst {
		case 0:
			cnt.Base = 0x88_0000
		case 1:
			cnt.Base = 0
		case 2:
			cnt.Base = 0x40_0000
		case 3:
			vm.TextureCache.Reset()
			vm.TexPalSlots[0] = (*[0x4000]uint8)(cnt.bank)
			vm.TexPalSlots[1] = (*[0x4000]uint8)(unsafe.Add(cnt.bank, 0x4000))
			vm.TexPalSlots[2] = (*[0x4000]uint8)(unsafe.Add(cnt.bank, 0x8000))
			vm.TexPalSlots[3] = (*[0x4000]uint8)(unsafe.Add(cnt.bank, 0xC000))

		case 4:
			vm.engineA.ExtBgSlots[0] = (*[0x2000]uint8)(cnt.bank)
			vm.engineA.ExtBgSlots[1] = (*[0x2000]uint8)(unsafe.Add(cnt.bank, 0x2000))
			vm.engineA.ExtBgSlots[2] = (*[0x2000]uint8)(unsafe.Add(cnt.bank, 0x4000))
			vm.engineA.ExtBgSlots[3] = (*[0x2000]uint8)(unsafe.Add(cnt.bank, 0x6000))
		}

	case 0x245:

		cnt := &vm.Cnt[F]
		cnt.Write(v)
		cnt.Base = 0x100_0000

		switch cnt.Mst {
		case 0:
			cnt.Base = 0x89_0000
		case 1:
			cnt.Base = (0x4000 * (cnt.Ofs & 1)) + (0x10000 * (cnt.Ofs >> 1))
		case 2:
			cnt.Base = 0x40_0000 + (0x4000 * (cnt.Ofs & 1)) + (0x10000 * (cnt.Ofs >> 1))
		case 3:
			vm.TextureCache.Reset()
			idx := (cnt.Ofs & 1) + (cnt.Ofs>>1)*4
			vm.TexPalSlots[idx] = (*[0x4000]uint8)(cnt.bank)
		case 4:

			if cnt.Ofs == 0 {
				vm.engineA.ExtBgSlots[0] = (*[0x2000]uint8)(cnt.bank)
				vm.engineA.ExtBgSlots[1] = (*[0x2000]uint8)(unsafe.Add(cnt.bank, 0x2000))
			} else {
				vm.engineA.ExtBgSlots[2] = (*[0x2000]uint8)(cnt.bank)
				vm.engineA.ExtBgSlots[3] = (*[0x2000]uint8)(unsafe.Add(cnt.bank, 0x2000))
			}

		case 5:
			vm.engineA.ExtObj = (*[0x4000]uint8)(cnt.bank)
		}

	case 0x246:
		cnt := &vm.Cnt[G]
		cnt.Write(v)
		cnt.Base = 0x100_0000

		switch cnt.Mst {
		case 0:
			cnt.Base = 0x89_4000
		case 1:
			cnt.Base = (0x4000 * (cnt.Ofs & 1)) + (0x10000 * (cnt.Ofs >> 1))
		case 2:
			cnt.Base = 0x40_0000 + (0x4000 * (cnt.Ofs & 1)) + (0x10000 * (cnt.Ofs >> 1))
		case 3:
			vm.TextureCache.Reset()
			idx := (cnt.Ofs & 1) + (cnt.Ofs>>1)*4
			vm.TexPalSlots[idx] = (*[0x4000]uint8)(cnt.bank)
		case 4:
			if cnt.Ofs == 0 {
				vm.engineA.ExtBgSlots[0] = (*[0x2000]uint8)(cnt.bank)
				vm.engineA.ExtBgSlots[1] = (*[0x2000]uint8)(unsafe.Add(cnt.bank, 0x2000))
			} else {
				vm.engineA.ExtBgSlots[2] = (*[0x2000]uint8)(cnt.bank)
				vm.engineA.ExtBgSlots[3] = (*[0x2000]uint8)(unsafe.Add(cnt.bank, 0x2000))
			}

		case 5:
			vm.engineA.ExtObj = (*[0x4000]uint8)(cnt.bank)
		}

		// 0x247 is WRAMCnt
	case 0x248:
		cnt := &vm.Cnt[H]
		cnt.Write(v)
		cnt.Base = 0x100_0000

		switch cnt.Mst {
		case 0:
			cnt.Base = 0x89_8000
		case 1:
			cnt.Base = 0x20_0000
		case 2:
			vm.engineB.ExtBgSlots[0] = (*[0x2000]uint8)(cnt.bank)
			vm.engineB.ExtBgSlots[1] = (*[0x2000]uint8)(unsafe.Add(cnt.bank, 0x2000))
			vm.engineB.ExtBgSlots[2] = (*[0x2000]uint8)(unsafe.Add(cnt.bank, 0x4000))
			vm.engineB.ExtBgSlots[3] = (*[0x2000]uint8)(unsafe.Add(cnt.bank, 0x6000))
		}

	case 0x249:
		cnt := &vm.Cnt[I]
		cnt.Write(v)
		cnt.Base = 0x100_0000

		switch cnt.Mst {
		case 0:
			cnt.Base = 0x8A_0000
		case 1:
			cnt.Base = 0x20_8000
		case 2:
			cnt.Base = 0x60_0000
		case 3:
			vm.engineB.ExtObj = (*[0x4000]uint8)(cnt.bank)
		}
	}
}

func (vm *VRAM) Write7(addr uint32, v uint8) {

	addr &= 0xFF_FFFF

	cnt := &vm.Cnt[C]

	if cnt.Enabled && cnt.arm7 && addr >= (cnt.Ofs*cnt.Size) {
		vm.c[addr&0x1FFFF] = v
	}

	cnt = &vm.Cnt[D]

	if cnt.Enabled && cnt.arm7 && addr >= (cnt.Ofs*cnt.Size) {
		vm.d[addr&0x1FFFF] = v
	}
}

func (vm *VRAM) Write9(addr uint32, v uint8) {

	addr &= 0xFF_FFFF

	for i := range len(vm.Cnt) {

		cnt := &vm.Cnt[i]

		if !cnt.Enabled {
			continue
		}

		if addr >= cnt.Base && addr < cnt.Base+cnt.Size {

			if i < 3 || (i >= 4 && i < 7) {
				vm.TextureCache.Reset()
			}

			(*[0x2_0000]uint8)(cnt.bank)[addr-cnt.Base] = v
			// return ???
		}
	}
}

func (vm *VRAM) Read7(addr uint32) uint8 {

	addr &= 0xFF_FFFF

	cnt := &vm.Cnt[C]
	if cnt.Enabled && cnt.arm7 && addr >= (cnt.Ofs*cnt.Size) {
		return vm.c[addr&0x1FFFF]
	}

	cnt = &vm.Cnt[D]
	if cnt.Enabled && cnt.arm7 && addr >= (cnt.Ofs*cnt.Size) {
		return vm.d[addr&0x1FFFF]
	}

	return 0
}

func (vm *VRAM) Read9(addr uint32) uint8 {

	addr &= 0xFF_FFFF

	for i := range len(vm.Cnt) {

		cnt := &vm.Cnt[i]

		if !cnt.Enabled || cnt.arm7 {
			continue
		}

		if addr >= cnt.Base && addr < cnt.Base+cnt.Size {
			return (*[0x2_0000]uint8)(cnt.bank)[addr-cnt.Base]
		}
	}

	return 0
}

func (vm *VRAM) ReadPtr7(addr uint32) (unsafe.Pointer, bool) {

	addr &= 0xFF_FFFF

	cnt := &vm.Cnt[C]
	if cnt.Enabled && cnt.arm7 && addr >= (cnt.Ofs*cnt.Size) {
		return unsafe.Add(cnt.bank, addr&0x1FFFF), true
	}

	cnt = &vm.Cnt[D]
	if cnt.Enabled && cnt.arm7 && addr >= (cnt.Ofs*cnt.Size) {
		return unsafe.Add(cnt.bank, addr&0x1FFFF), true
	}

	return nil, false
}

func (vm *VRAM) ReadPtr9(addr uint32) (unsafe.Pointer, bool) {

	addr &= 0xFF_FFFF

	for i := range len(vm.Cnt) {

		cnt := &vm.Cnt[i]

		if !cnt.Enabled || cnt.arm7 {
			continue
		}

		if addr >= cnt.Base && addr < cnt.Base+cnt.Size {
			return unsafe.Add(cnt.bank, addr-cnt.Base), true
		}
	}

	return nil, false
}

func (vm *VRAM) ReadGraphicalPtr(addr uint32) unsafe.Pointer {

	for i := range len(vm.Cnt) {

		cnt := &vm.Cnt[i]

		if !cnt.Enabled || cnt.arm7 {
			continue
		}

		end := cnt.Base + cnt.Size

		if addr < cnt.Base || addr >= end {
			continue
		}

		return unsafe.Add(cnt.bank, addr-cnt.Base)
	}

	return nil
}

func (vm *VRAM) Read16(addr uint32) uint16 {

	// only should be used in 2d graphics

	for i := range len(vm.Cnt) {

		cnt := &vm.Cnt[i]

		if !cnt.Enabled || cnt.arm7 {
			continue
		}

		end := cnt.Base + cnt.Size

		if addr < cnt.Base || addr+1 >= end {
			continue
		}

		return *(*uint16)(unsafe.Add(cnt.bank, addr-cnt.Base))
	}

	return 0
}

func (vm *VRAM) ReadTexture(addr uint32) uint8 {

	region := addr >> 17

	if region >= 4 {
		return 0
	}

	if vm.TextureSlots[region] == nil {
		return 0
	}

	return vm.TextureSlots[region][addr&0x1FFFF]
}

func (vm *VRAM) ReadPalTexture(addr uint32) uint8 {

	region := addr >> 14

	if region >= 6 {
		return 0
	}

	if vm.TexPalSlots[region] == nil {
		return 0
	}

	return vm.TexPalSlots[region][addr&0x3FFF]
}
