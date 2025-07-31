package gba

import "github.com/aabalke/guac/emu/gba/cart"

// ygba
var vsyncAddrMap = map[cart.Header]uint32{
    {Title: "ADVANCEWARS" , GameCode: "AWRE", Version: 0}: 0x80387EC, // Advance Wars (USA)
    {Title: "ADVANCEWARS" , GameCode: "AWRE", Version: 1}: 0x8038818, // Advance Wars (USA) (Rev 1)
    {Title: "DRILL DOZER" , GameCode: "V49E", Version: 0}: 0x80006BA, // Drill Dozer (USA)
    {Title: "FFTA_USVER." , GameCode: "AFXE", Version: 0}: 0x8000418, // Final Fantasy Tactics Advance (USA, Australia)
    {Title: "KURURIN"     , GameCode: "AKRP", Version: 0}: 0x800041A, // Kurukuru Kururin (Europe)
    {Title: "POKEMON EMER", GameCode: "BPEE", Version: 0}: 0x80008C6, // Pokemon - Emerald Version (USA, Europe)
    {Title: "POKEMON FIRE", GameCode: "BPRE", Version: 0}: 0x80008AA, // Pokemon - FireRed Version (USA)
    {Title: "POKEMON FIRE", GameCode: "BPRE", Version: 1}: 0x80008BE, // Pokemon - FireRed Version (USA, Europe) (Rev 1)
    {Title: "POKEMON LEAF", GameCode: "BPGE", Version: 0}: 0x80008AA, // Pokemon - LeafGreen Version (USA)
    {Title: "POKEMON LEAF", GameCode: "BPGE", Version: 1}: 0x80008BE, // Pokemon - LeafGreen Version (USA, Europe) (Rev 1)
}

func (gba *GBA) SetIdleAddr() {
    v, ok := vsyncAddrMap[*gba.Cartridge.Header]
    if !ok {
        return
    }

    gba.vsyncAddr = v
}
