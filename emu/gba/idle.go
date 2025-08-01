package gba

import "github.com/aabalke/guac/emu/gba/cart"

// ygba
var vsyncAddrMap = map[cart.Header]uint32{
    {Title: "ADVANCEWARS" , GameCode: "AWRE", Version: 0}: 0x80387EC,
    {Title: "ADVANCEWARS" , GameCode: "AWRE", Version: 1}: 0x8038818,
    {Title: "DRILL DOZER" , GameCode: "V49E", Version: 0}: 0x80006BA,
    {Title: "FFTA_USVER." , GameCode: "AFXE", Version: 0}: 0x8000418,
    {Title: "KURURIN"     , GameCode: "AKRP", Version: 0}: 0x800041A,
    {Title: "POKEMON EMER", GameCode: "BPEE", Version: 0}: 0x80008C6,
    {Title: "POKEMON FIRE", GameCode: "BPRE", Version: 0}: 0x80008AA,
    {Title: "POKEMON FIRE", GameCode: "BPRE", Version: 1}: 0x80008BE,
    {Title: "POKEMON LEAF", GameCode: "BPGE", Version: 0}: 0x80008AA,
    {Title: "POKEMON LEAF", GameCode: "BPGE", Version: 1}: 0x80008BE,
}

func (gba *GBA) SetIdleAddr() {
    v, ok := vsyncAddrMap[*gba.Cartridge.Header]
    if !ok {
        return
    }

    gba.vsyncAddr = v
}
