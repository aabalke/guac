package mem

func (m *Mem) ReadGbaSlot(addr uint32) uint8 {

    //m.Gamecard.ExMem.

    // wrong cpu access == 0


    if sram := addr >= 0xA00_0000; sram {
        return 0xFF
    }

    return uint8(((addr >> 1) & 0xFFFF) >> ((addr & 1) * 8))

    //switch r := addr >> 24; r {
    //case 8, 9: return 0x
    //case 0xA, 0xB, 0xC, 0xD
    //}


}
