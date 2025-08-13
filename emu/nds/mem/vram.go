package mem

import "github.com/aabalke/guac/emu/nds/utils"

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
}

type VramCnt struct {
	Mst, Ofs uint8
	Enabled  bool
}

func (vc *VramCnt) Write(v uint8) {
	vc.Mst = v & 0b111
	vc.Ofs = (v >> 3) & 0b11
	vc.Enabled = utils.BitEnabled(uint32(v), 7)
}

func (vm *VRAM) WriteCNT(addr uint32, v uint8) {
	switch addr {
	case 0x240: vm.CNT_A.Write(v)
	case 0x241: vm.CNT_B.Write(v)
	case 0x242: vm.CNT_C.Write(v)
	case 0x243: vm.CNT_D.Write(v)
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
        switch {
        case addr < 0x80_0000: return
        case addr < 0x82_0000: vm.A[addr - 0x80_0000] = v
        case addr < 0x84_0000: vm.B[addr - 0x82_0000] = v
        case addr < 0x86_0000: vm.C[addr - 0x84_0000] = v
        case addr < 0x88_0000: vm.D[addr - 0x86_0000] = v
        case addr < 0x89_0000: vm.E[addr - 0x88_0000] = v
        case addr < 0x89_4000: vm.F[addr - 0x89_0000] = v
        case addr < 0x89_8000: vm.G[addr - 0x89_4000] = v
        case addr < 0x8A_0000: vm.H[addr - 0x89_8000] = v
        case addr < 0x8A_4000: vm.I[addr - 0x8A_0000] = v
        }

        return
    }
}

func (vm *VRAM) Read(addr uint32, arm9 bool) uint8 {

    addr &= 0xFF_FFFF

    if arm9 {
        switch {
        case addr < 0x80_0000: return 0
        case addr < 0x82_0000: return vm.A[addr - 0x80_0000]
        case addr < 0x84_0000: return vm.B[addr - 0x82_0000]
        case addr < 0x86_0000: return vm.C[addr - 0x84_0000]
        case addr < 0x88_0000: return vm.D[addr - 0x86_0000]
        case addr < 0x89_0000: return vm.E[addr - 0x88_0000]
        case addr < 0x89_4000: return vm.F[addr - 0x89_0000]
        case addr < 0x89_8000: return vm.G[addr - 0x89_4000]
        case addr < 0x8A_0000: return vm.H[addr - 0x89_8000]
        case addr < 0x8A_4000: return vm.I[addr - 0x8A_0000]
        }
    }

    return 0
}
