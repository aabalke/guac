package gba

import "github.com/aabalke/guac/emu/gba/utils"

type Keypad struct {
	KEYINPUT uint16
	KEYCNT   uint16
}

func (k *Keypad) readINPUT(hi bool) uint8 {

	if hi {
		return uint8(k.KEYINPUT >> 8)
	}

	return uint8(k.KEYINPUT)
}

func (k *Keypad) readCNT(hi bool) uint8 {

	if hi {
		return uint8(k.KEYCNT >> 8)
	}

	return uint8(k.KEYCNT)
}

func (k *Keypad) writeCNT(v uint8, hi bool) {

	if hi {
		k.KEYCNT = k.KEYCNT&0xFF | (uint16(v) << 8)
		return
	}

	k.KEYCNT = k.KEYCNT&^0xFF | uint16(v)
}

func (k *Keypad) keyIRQ() bool {

	if disabled := utils.BitEnabled(uint32(k.KEYCNT), 14); disabled {
		return false
	}

	andFlag := utils.BitEnabled(uint32(k.KEYCNT), 15)

	if or := !andFlag && ^(k.KEYCNT)&k.KEYINPUT != 0; or {
		return true
	}

	if and := andFlag && ^(k.KEYCNT)&^k.KEYINPUT == 0; and {
		return true
	}

	return false
}
