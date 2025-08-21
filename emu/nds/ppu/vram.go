package ppu

import (

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
}

type VramCnt struct {
	Mst, Ofs uint8
	Enabled  bool
}

func (vc *VramCnt) Write(v uint8) {
	vc.Mst = v & 0b111
    vc.Ofs = uint8(utils.GetVarData(uint32(v), 3, 4))
	vc.Enabled = utils.BitEnabled(uint32(v), 7)
}

func (vm *VRAM) WriteCNT(addr uint32, v uint8) {

    //fmt.Printf("WRITE VRAM CNT %08X %02X\n", addr, v)

	switch addr {
	case 0x240: vm.CNT_A.Write(v)
	case 0x241: vm.CNT_B.Write(v)
	case 0x242:
        vm.CNT_C.Write(v)
        vm.isCArm7 = v & 0b10000011 == 0b10000010

        vm.CNT_7 &^= 1

        if vm.isCArm7 {
            vm.CNT_7 |= 1
        }

	case 0x243:
        vm.CNT_D.Write(v)
        vm.isDArm7 = v & 0b10000011 == 0b10000010

        vm.CNT_7 &^= 0b10

        if vm.isDArm7 {
            vm.CNT_7 |= 0b10
        }
	case 0x244: vm.CNT_E.Write(v)
	case 0x245: vm.CNT_F.Write(v)
	case 0x246: vm.CNT_G.Write(v)
    // 0x247 is WRAMCNT
	case 0x248: vm.CNT_H.Write(v)
	case 0x249: vm.CNT_I.Write(v)
	}
}

func (vm *VRAM) Write(addr uint32, v uint8, arm9 bool) {

    addr &= 0xFF_FFFF

    // currently assuming lcdc mode

    if arm9 {

        base := uint32(0x100_0000) // make sure 0 does not grab everything
        if vm.CNT_A.Enabled {
            switch vm.CNT_A.Mst {
            case 0: base = uint32(0x80_0000)
            case 1: base = 0x20000 * uint32(vm.CNT_A.Ofs)
            case 2: base = 0x400000 + 0x20000 * uint32(vm.CNT_A.Ofs)
            case 3: // slot
            }
            if addr >= base && addr < base + 0x2_0000 {
                vm.A[addr - base] = v
            }
        }

        base = uint32(0x100_0000) // make sure 0 does not grab everything
        if vm.CNT_B.Enabled {
            switch vm.CNT_B.Mst {
            case 0: base = uint32(0x82_0000)
            case 1: base = 0x20000 * uint32(vm.CNT_B.Ofs)
            case 2: base = 0x400000 + 0x20000 * uint32(vm.CNT_B.Ofs)
            case 3: // slot
            }
            if addr >= base && addr < base + 0x2_0000 {
                vm.B[addr - base] = v
            }
        }

        base = uint32(0x100_0000) // make sure 0 does not grab everything
        if vm.CNT_C.Enabled {
            switch vm.CNT_C.Mst {
            case 0: base = uint32(0x84_0000)
            case 1: base = 0x20000 * uint32(vm.CNT_C.Ofs)
            case 2: // given to arm7
            case 3: // slot
            case 4: base = 0x20_0000
            }
            if addr >= base && addr < base + 0x2_0000 {
                vm.C[addr - base] = v
            }
        }

        base = uint32(0x100_0000) // make sure 0 does not grab everything
        if vm.CNT_D.Enabled {
            switch vm.CNT_D.Mst {
            case 0: base = uint32(0x86_0000)
            case 1: base = 0x20000 * uint32(vm.CNT_D.Ofs)
            case 2: // given to arm7
            case 3: // slot
            case 4: base = 0x60_0000
            }
            if addr >= base && addr < base + 0x2_0000 {
                vm.D[addr - base] = v
            }
        }

        base = uint32(0x100_0000) // make sure 0 does not grab everything
        if vm.CNT_E.Enabled {
            switch vm.CNT_E.Mst {
            case 0: base = uint32(0x88_0000)
            case 1: base = 0
            case 2: base = 0x40_0000
            case 3: // slot
            case 4: // slot
            }
            if addr >= base && addr < base + 0x1_0000 {
                vm.E[addr - base] = v
            }
        }

        base = uint32(0x100_0000) // make sure 0 does not grab everything
        if vm.CNT_F.Enabled {
            switch vm.CNT_F.Mst {
            case 0: base = uint32(0x89_0000)
            case 1: base = 0 // not sure
            case 2: base = 0 // not sure
            case 3: // slot
            case 4: // slot
            }
            if addr >= base && addr < base + 0x4000 {
                vm.F[addr - base] = v
            }
        }

        base = uint32(0x100_0000) // make sure 0 does not grab everything
        if vm.CNT_G.Enabled {
            switch vm.CNT_G.Mst {
            case 0: base = uint32(0x89_4000)
            case 1: base = 0 // not sure
            case 2: base = 0 // not sure
            case 3: // slot
            case 4: // slot
            }
            if addr >= base && addr < base + 0x4000 {
                vm.G[addr - base] = v
            }
        }

        base = uint32(0x100_0000) // make sure 0 does not grab everything
        if vm.CNT_H.Enabled {
            switch vm.CNT_H.Mst {
            case 0: base = uint32(0x89_8000)
            case 1: base = 0 // not sure
            case 2: base = 0 // not sure
            case 3: // slot
            case 4: // slot
            }
            if addr >= base && addr < base + 0x8000 {
                vm.H[addr - base] = v
            }
        }

        base = uint32(0x100_0000) // make sure 0 does not grab everything
        if vm.CNT_I.Enabled {
            switch vm.CNT_I.Mst {
            case 0: base = uint32(0x8A_0000)
            case 1: base = 0x208000 // not sure
            case 2: base = 0x600000 // not sure
            case 3: // slot
            }
            if addr >= base && addr < base + 0x8000 {
                vm.I[addr - base] = v
            }
        }

        return
    }

    if vm.isCArm7 && addr >= (uint32(vm.CNT_C.Ofs) * 0x20000) {
        vm.C[addr & 0x1FFFF] = v
    }

    if vm.isDArm7 && addr >= (uint32(vm.CNT_D.Ofs) * 0x20000) {
        vm.D[addr & 0x1FFFF] = v
    }
}

func (vm *VRAM) Read(addr uint32, arm9 bool) uint8 {

    addr &= 0xFF_FFFF

    if arm9 {

        base := uint32(0x100_0000) // make sure 0 does not grab everything
        if vm.CNT_A.Enabled {
            switch vm.CNT_A.Mst {
            case 0: base = uint32(0x80_0000)
            case 1: base = 0x20000 * uint32(vm.CNT_A.Ofs)
            case 2: base = 0x400000 + 0x20000 * uint32(vm.CNT_A.Ofs)
            case 3: // slot
            }
            if addr >= base && addr < base + 0x2_0000 {
                return vm.A[addr - base]
            }
        }

        base = uint32(0x100_0000) // make sure 0 does not grab everything
        if vm.CNT_B.Enabled {
            switch vm.CNT_B.Mst {
            case 0: base = uint32(0x82_0000)
            case 1: base = 0x20000 * uint32(vm.CNT_B.Ofs)
            case 2: base = 0x400000 + 0x20000 * uint32(vm.CNT_B.Ofs)
            case 3: // slot
            }
            if addr >= base && addr < base + 0x2_0000 {
                return vm.B[addr - base]
            }
        }

        base = uint32(0x100_0000) // make sure 0 does not grab everything
        if vm.CNT_C.Enabled {
            switch vm.CNT_C.Mst {
            case 0: base = uint32(0x84_0000)
            case 1: base = 0x20000 * uint32(vm.CNT_C.Ofs)
            case 2: // given to arm7
            case 3: // slot
            case 4: base = 0x20_0000
            }
            if addr >= base && addr < base + 0x2_0000 {
                return vm.C[addr - base]
            }
        }

        base = uint32(0x100_0000) // make sure 0 does not grab everything
        if vm.CNT_D.Enabled {
            switch vm.CNT_D.Mst {
            case 0: base = uint32(0x86_0000)
            case 1: base = 0x20000 * uint32(vm.CNT_D.Ofs)
            case 2: // given to arm7
            case 3: // slot
            case 4: base = 0x60_0000
            }
            if addr >= base && addr < base + 0x2_0000 {
                return vm.D[addr - base]
            }
        }

        base = uint32(0x100_0000) // make sure 0 does not grab everything
        if vm.CNT_E.Enabled {
            switch vm.CNT_E.Mst {
            case 0: base = uint32(0x88_0000)
            case 1: base = 0
            case 2: base = 0x40_0000
            case 3: // slot
            case 4: // slot
            }
            if addr >= base && addr < base + 0x1_0000 {
                return vm.E[addr - base]
            }
        }

        base = uint32(0x100_0000) // make sure 0 does not grab everything
        if vm.CNT_F.Enabled {
            switch vm.CNT_F.Mst {
            case 0: base = uint32(0x89_0000)
            case 1: base = 0 // not sure
            case 2: base = 0 // not sure
            case 3: // slot
            case 4: // slot
            }
            if addr >= base && addr < base + 0x4000 {
                return vm.F[addr - base]
            }
        }

        base = uint32(0x100_0000) // make sure 0 does not grab everything
        if vm.CNT_G.Enabled {
            switch vm.CNT_G.Mst {
            case 0: base = uint32(0x89_4000)
            case 1: base = 0 // not sure
            case 2: base = 0 // not sure
            case 3: // slot
            case 4: // slot
            }
            if addr >= base && addr < base + 0x4000 {
                return vm.G[addr - base]
            }
        }

        base = uint32(0x100_0000) // make sure 0 does not grab everything
        if vm.CNT_H.Enabled {
            switch vm.CNT_H.Mst {
            case 0: base = uint32(0x89_8000)
            case 1: base = 0 // not sure
            case 2: base = 0 // not sure
            case 3: // slot
            case 4: // slot
            }
            if addr >= base && addr < base + 0x8000 {
                return vm.H[addr - base]
            }
        }

        base = uint32(0x100_0000) // make sure 0 does not grab everything
        if vm.CNT_I.Enabled {
            switch vm.CNT_I.Mst {
            case 0: base = uint32(0x8A_0000)
            case 1: base = 0x208000 // not sure
            case 2: base = 0x600000 // not sure
            case 3: // slot
            }
            if addr >= base && addr < base + 0x8000 {
                return vm.I[addr - base]
            }
        }

        return 0
    }

    if vm.isCArm7 && addr >= (uint32(vm.CNT_C.Ofs) * 0x20000) {
        return vm.C[addr & 0x1FFFF]
    }

    if vm.isDArm7 && addr >= (uint32(vm.CNT_D.Ofs) * 0x20000) {
        return vm.D[addr & 0x1FFFF]
    }

    return 0
}
