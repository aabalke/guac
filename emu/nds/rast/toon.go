package rast

import "github.com/aabalke/guac/emu/nds/rast/gl"

func WriteToonTbl(t *[32]gl.Color, addr uint32, v uint8) {
	addr -= 0x380
	idx := addr / 2
	hi := addr%2 == 1
	t[idx] = Convert15BitByte(t[idx], v, hi)
}
