package gba

const (
    DISPCNT  = 0x0000
    DISPSTAT  = 0x0004
	KEYINPUT = 0x0130
	KEYCNT   = 0x0132
	BG0CNT   = 0x0008
	BG1CNT   = 0x000A
	BG2CNT   = 0x000C
	BG3CNT   = 0x000E
)

var (
	VCOUNT = uint8(0)

    Bg0Control = BgControl(0)
    Bg1Control = BgControl(0)
    Bg2Control = BgControl(0)
    Bg3Control = BgControl(0)
)

type BgControl uint16
