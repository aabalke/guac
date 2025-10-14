package ppu

import (
	"unsafe"

	"github.com/aabalke/guac/emu/nds/rast"
	"github.com/aabalke/guac/emu/nds/utils"
)

type VRAM struct {
	A [0x2_0000]uint8
	B [0x2_0000]uint8
	C [0x2_0000]uint8
	D [0x2_0000]uint8
	E [0x1_0000]uint8
	F [0x0_4000]uint8
	G [0x0_4000]uint8
	H [0x0_8000]uint8
	I [0x0_4000]uint8

	CNT_A VramCnt
	CNT_B VramCnt
	CNT_C VramCnt
	CNT_D VramCnt
	CNT_E VramCnt
	CNT_F VramCnt
	CNT_G VramCnt
	CNT_H VramCnt
	CNT_I VramCnt

    CNT_7 uint8
    isCArm7, isDArm7 bool

    ExtABgSlots [4]*[0x2000]uint8
    ExtBBgSlots [4]*[0x2000]uint8
    ExtAPalObj *[0x4000]uint8
    ExtBPalObj *[0x4000]uint8
    TextureSlots [4]*[0x2_0000]uint8
    TexPalSlots  [6]*[0x4000]uint8

    TextureCache *rast.TextureCache

    banks [9]struct{
        bank unsafe.Pointer
        cnt *VramCnt
    }
}

func (v *VRAM) Init(t *rast.TextureCache) {
    v.TextureCache = t
    v.CNT_A.Write(0x80)
    v.CNT_B.Write(0x80)
    v.CNT_C.Write(0x80)
    v.CNT_D.Write(0x80)
    v.CNT_E.Write(0x80)
    v.CNT_F.Write(0x80)
    v.CNT_G.Write(0x80)
    v.CNT_H.Write(0x80)
    v.CNT_I.Write(0x80)

	v.CNT_A.Size  = 0x2_0000
	v.CNT_B.Size  = 0x2_0000
	v.CNT_C.Size  = 0x2_0000
	v.CNT_D.Size  = 0x2_0000
	v.CNT_E.Size  = 0x1_0000
	v.CNT_F.Size  = 0x0_4000
	v.CNT_G.Size  = 0x0_4000
	v.CNT_H.Size  = 0x0_8000
	v.CNT_I.Size  = 0x0_4000

    v.banks[0] = struct{bank unsafe.Pointer; cnt *VramCnt}{ bank: unsafe.Pointer(&v.A), cnt: &v.CNT_A}
    v.banks[1] = struct{bank unsafe.Pointer; cnt *VramCnt}{ bank: unsafe.Pointer(&v.B), cnt: &v.CNT_B}
    v.banks[2] = struct{bank unsafe.Pointer; cnt *VramCnt}{ bank: unsafe.Pointer(&v.C), cnt: &v.CNT_C}
    v.banks[3] = struct{bank unsafe.Pointer; cnt *VramCnt}{ bank: unsafe.Pointer(&v.D), cnt: &v.CNT_D}
    v.banks[4] = struct{bank unsafe.Pointer; cnt *VramCnt}{ bank: unsafe.Pointer(&v.E), cnt: &v.CNT_E}
    v.banks[5] = struct{bank unsafe.Pointer; cnt *VramCnt}{ bank: unsafe.Pointer(&v.F), cnt: &v.CNT_F}
    v.banks[6] = struct{bank unsafe.Pointer; cnt *VramCnt}{ bank: unsafe.Pointer(&v.G), cnt: &v.CNT_G}
    v.banks[7] = struct{bank unsafe.Pointer; cnt *VramCnt}{ bank: unsafe.Pointer(&v.H), cnt: &v.CNT_H}
    v.banks[8] = struct{bank unsafe.Pointer; cnt *VramCnt}{ bank: unsafe.Pointer(&v.I), cnt: &v.CNT_I}
}

type VramCnt struct {
    V uint8
	Mst uint8
	Enabled  bool
    Ofs uint32
    Base uint32
    Size uint32
}

func (vc *VramCnt) Write(v uint8) {
    vc.V = v & 0b1001_1111
	vc.Mst = v & 0b111
    vc.Ofs = uint32(utils.GetVarData(uint32(v), 3, 4))
	vc.Enabled = utils.BitEnabled(uint32(v), 7)
}

func (vm *VRAM) WriteCNT(addr uint32, v uint8) {

	switch addr {
	case 0x240:

        cnt := &vm.CNT_A
        bank := &vm.A

        cnt.Write(v)

        cnt.Base = 0x100_0000

        switch cnt.Mst {
        case 0: cnt.Base = uint32(0x80_0000)
        case 1: cnt.Base = 0x20000 * cnt.Ofs
        case 2: cnt.Base = 0x400000 + 0x20000 * cnt.Ofs
        case 3:
            vm.TextureCache.Reset()
            vm.TextureSlots[cnt.Ofs] = bank
        }


	case 0x241:

        cnt := &vm.CNT_B
        bank := &vm.B

        cnt.Write(v)

        cnt.Base = 0x100_0000

        switch cnt.Mst {
        case 0: cnt.Base = uint32(0x82_0000)
        case 1: cnt.Base = 0x20000 * cnt.Ofs
        case 2: cnt.Base = 0x400000 + 0x20000 * cnt.Ofs
        case 3:
            vm.TextureCache.Reset()
            vm.TextureSlots[cnt.Ofs] = bank
        }


	case 0x242:

        cnt := &vm.CNT_C
        bank := &vm.C

        cnt.Write(v)
        cnt.Base = 0x100_0000

        vm.isCArm7 = false
        vm.CNT_7 &^= 1

        switch cnt.Mst {
        case 0: cnt.Base = uint32(0x84_0000)
        case 1: cnt.Base = 0x20000 * cnt.Ofs
        case 2:

            if cnt.Ofs >= 2 { panic("INVALID ARM7 CNT C OFS")}

            vm.isCArm7 = true
            vm.CNT_7 |= 1

        case 3:
            vm.TextureCache.Reset()
            vm.TextureSlots[cnt.Ofs] = bank
        case 4: cnt.Base = 0x20_0000
        }

	case 0x243:

        cnt := &vm.CNT_D
        bank := &vm.D

        cnt.Write(v)
        cnt.Base = 0x100_0000

        vm.isDArm7 = false
        vm.CNT_7 &^= 0b10

        switch cnt.Mst {
        case 0: cnt.Base = uint32(0x86_0000)
        case 1: cnt.Base = 0x20000 * cnt.Ofs
        case 2:

            if cnt.Ofs >= 2 { panic("INVALID ARM7 CNT D OFS")}

            vm.isDArm7 = true
            vm.CNT_7 |= 0b10

        case 3:
            vm.TextureCache.Reset()
            vm.TextureSlots[cnt.Ofs] = bank
        case 4: cnt.Base = 0x60_0000
        }

	case 0x244:

        cnt := &vm.CNT_E
        bank := &vm.E
        cnt.Write(v)
        cnt.Base = 0x100_0000

        switch cnt.Mst {
        case 0: cnt.Base = uint32(0x88_0000)
        case 1: cnt.Base = 0
        case 2: cnt.Base = 0x40_0000
        case 3:
            vm.TextureCache.Reset()
            vm.TexPalSlots[0] = (*[0x4000]uint8)(unsafe.Pointer(bank))
            vm.TexPalSlots[1] = (*[0x4000]uint8)(unsafe.Add(unsafe.Pointer(bank), 0x4000))
            vm.TexPalSlots[2] = (*[0x4000]uint8)(unsafe.Add(unsafe.Pointer(bank), 0x8000))
            vm.TexPalSlots[3] = (*[0x4000]uint8)(unsafe.Add(unsafe.Pointer(bank), 0xC000))

        case 4:
            vm.ExtABgSlots[0] = (*[0x2000]uint8)(unsafe.Pointer(bank))
            vm.ExtABgSlots[1] = (*[0x2000]uint8)(unsafe.Add(unsafe.Pointer(bank), 0x2000))
            vm.ExtABgSlots[2] = (*[0x2000]uint8)(unsafe.Add(unsafe.Pointer(bank), 0x4000))
            vm.ExtABgSlots[3] = (*[0x2000]uint8)(unsafe.Add(unsafe.Pointer(bank), 0x6000))
        }

	case 0x245:

        cnt := &vm.CNT_F
        bank := &vm.F
        cnt.Write(v)
        cnt.Base = 0x100_0000

        switch cnt.Mst {
        case 0: cnt.Base = uint32(0x89_0000)
        case 1: cnt.Base = (0x4000 * uint32(cnt.Ofs & 1)) + (0x10000 * uint32(cnt.Ofs >> 1))
        case 2: cnt.Base = 0x40_0000 + (0x4000 * uint32(cnt.Ofs & 1)) + (0x10000 * uint32(cnt.Ofs >> 1))
        case 3:
            vm.TextureCache.Reset()
            idx := (cnt.Ofs & 1) + (cnt.Ofs >> 1) * 4
            vm.TexPalSlots[idx] = (*[0x4000]uint8)(unsafe.Pointer(bank))
        case 4:

            if cnt.Ofs == 0 {
                vm.ExtABgSlots[0] = (*[0x2000]uint8)(unsafe.Pointer(bank))
                vm.ExtABgSlots[1] = (*[0x2000]uint8)(unsafe.Add(unsafe.Pointer(bank), 0x2000))
            } else {            
                vm.ExtABgSlots[2] = (*[0x2000]uint8)(unsafe.Pointer(bank))
                vm.ExtABgSlots[3] = (*[0x2000]uint8)(unsafe.Add(unsafe.Pointer(bank), 0x2000))
            }

        case 5:
            vm.ExtAPalObj = bank
        }

	case 0x246:
        cnt := &vm.CNT_G
        bank := &vm.G
        cnt.Write(v)
        cnt.Base = 0x100_0000

        switch cnt.Mst {
        case 0: cnt.Base = uint32(0x89_4000)
        case 1: cnt.Base = (0x4000 * uint32(cnt.Ofs & 1)) + (0x10000 * uint32(cnt.Ofs >> 1))
        case 2: cnt.Base = 0x40_0000 + (0x4000 * uint32(cnt.Ofs & 1)) + (0x10000 * uint32(cnt.Ofs >> 1))
        case 3:
            vm.TextureCache.Reset()
            idx := (cnt.Ofs & 1) + (cnt.Ofs >> 1) * 4
            vm.TexPalSlots[idx] = (*[0x4000]uint8)(unsafe.Pointer(bank))
        case 4:
            if cnt.Ofs == 0 {
                vm.ExtABgSlots[0] = (*[0x2000]uint8)(unsafe.Pointer(bank))
                vm.ExtABgSlots[1] = (*[0x2000]uint8)(unsafe.Add(unsafe.Pointer(bank), 0x2000))
            } else {            
                vm.ExtABgSlots[2] = (*[0x2000]uint8)(unsafe.Pointer(bank))
                vm.ExtABgSlots[3] = (*[0x2000]uint8)(unsafe.Add(unsafe.Pointer(bank), 0x2000))
            }

        case 5:
            vm.ExtAPalObj = bank
        }

    // 0x247 is WRAMCNT
	case 0x248:
        cnt := &vm.CNT_H
        bank := &vm.H
        cnt.Write(v)
        cnt.Base = 0x100_0000

        switch cnt.Mst {
        case 0: cnt.Base = 0x89_8000
        case 1: cnt.Base = 0x20_0000
        case 2:
            vm.ExtBBgSlots[0] = (*[0x2000]uint8)(unsafe.Pointer(bank))
            vm.ExtBBgSlots[1] = (*[0x2000]uint8)(unsafe.Add(unsafe.Pointer(bank), 0x2000))
            vm.ExtBBgSlots[2] = (*[0x2000]uint8)(unsafe.Add(unsafe.Pointer(bank), 0x4000))
            vm.ExtBBgSlots[3] = (*[0x2000]uint8)(unsafe.Add(unsafe.Pointer(bank), 0x6000))
        }

	case 0x249:
        cnt := &vm.CNT_I
        bank := &vm.I
        cnt.Write(v)
        cnt.Base = 0x100_0000

        switch cnt.Mst {
        case 0: cnt.Base = uint32(0x8A_0000)
        case 1: cnt.Base = 0x20_8000
        case 2: cnt.Base = 0x60_0000
        case 3: vm.ExtBPalObj = bank
        }
	}
}

func (vm *VRAM) Write(addr uint32, v uint8, arm9 bool) {

    addr &= 0xFF_FFFF

    if !arm9 {

        cnt := &vm.CNT_C
        bank := &vm.C

        if vm.isCArm7 && addr >= (cnt.Ofs * cnt.Size) {
            bank[addr & 0x1FFFF] = v
        }

        cnt = &vm.CNT_D
        bank = &vm.D

        if vm.isDArm7 && addr >= (cnt.Ofs * cnt.Size) {
            bank[addr & 0x1FFFF] = v
        }

        return
    }

    for i, vb := range &vm.banks {

        if !vb.cnt.Enabled {
            continue
        }


        if addr >= vb.cnt.Base && addr < vb.cnt.Base + vb.cnt.Size {

            if i < 3 || (i >= 4 && i < 7) {
                vm.TextureCache.Reset()
            }

            (*[0x2_0000]uint8)(vb.bank)[addr - vb.cnt.Base] = v
            // return ???
        }
    }
}

func (vm *VRAM) Read(addr uint32, arm9 bool) uint8 {

    addr &= 0xFF_FFFF

    if !arm9 {

        cnt := &vm.CNT_C
        bank := &vm.C

        if vm.isCArm7 && addr >= (cnt.Ofs * cnt.Size) {
            return bank[addr & 0x1FFFF]
        }

        cnt = &vm.CNT_D
        bank = &vm.D

        if vm.isDArm7 && addr >= (cnt.Ofs * cnt.Size) {
            return bank[addr & 0x1FFFF]
        }

        return 0
    }

    for _, v := range &vm.banks {

        if !v.cnt.Enabled {
            continue
        }

        if addr >= v.cnt.Base && addr < v.cnt.Base + v.cnt.Size {
            return (*[0x2_0000]uint8)(v.bank)[addr - v.cnt.Base]
        }
    }

    return 0
}

func (vm *VRAM) ReadPtr(addr uint32, arm9 bool) (unsafe.Pointer, bool) {

    addr &= 0xFF_FFFF

    if !arm9 {

        cnt := &vm.CNT_C
        bank := &vm.C

        if vm.isCArm7 && addr >= (cnt.Ofs * cnt.Size) {
            return unsafe.Add(unsafe.Pointer(bank), addr & 0x1FFFF), true
        }

        cnt = &vm.CNT_D
        bank = &vm.D

        if vm.isDArm7 && addr >= (cnt.Ofs * cnt.Size) {
            return unsafe.Add(unsafe.Pointer(bank), addr & 0x1FFFF), true
        }

        return nil, false
    }

    for _, v := range &vm.banks {

        if !v.cnt.Enabled {
            continue
        }

        if addr >= v.cnt.Base && addr < v.cnt.Base + v.cnt.Size {
            return unsafe.Add(v.bank, addr - v.cnt.Base), true
        }
    }

    return nil, false
}

func (vm *VRAM) ReadGraphical(addr uint32) uint16 {

    for _, v := range &vm.banks {

        if !v.cnt.Enabled {
            continue
        }

        if !(addr >= v.cnt.Base && addr < v.cnt.Base + v.cnt.Size) {
            continue
        }

        if addr + 1 >= v.cnt.Base + v.cnt.Size {
            return uint16((*[1]uint8)(unsafe.Add(v.bank, addr - v.cnt.Base))[0])
        }

        return (*[1]uint16)(unsafe.Add(v.bank, addr - v.cnt.Base))[0]
    }

    return 0
}

func (vm *VRAM) ReadTexture(addr uint32) uint8 {
    return vm.TextureSlots[addr >> 17][addr & 0x1FFFF]
}

func (vm *VRAM) ReadPalTexture(addr uint32) uint8 {
    return vm.TexPalSlots[addr >> 14][addr & 0x3FFF]
}
