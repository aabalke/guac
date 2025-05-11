package gba

import "github.com/aabalke33/guac/emu/gba/utils"

const (
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

func (b BgControl) getPriority() uint16 {
	return uint16(b & 0b11)
}

func (b BgControl) getCharacterBaseBlock() uint16 {
	return uint16((b >> 2) & 0b11)
}

func (b BgControl) isMosaic() bool {
	return utils.BitEnabled(uint32(b), 6)
}

func (b BgControl) is256() bool {
	return utils.BitEnabled(uint32(b), 7)
}

func (b BgControl) getScreenBaseBlock() uint16 {
	return uint16(utils.GetVarData(uint32(b), 8, 12))
}

func (b BgControl) isWraparound() bool {
	return utils.BitEnabled(uint32(b), 13)
}

func (b BgControl) getScreenSize() uint16 {
    // Value  Text Mode      Rotation/Scaling Mode
    // 0      256x256 (2K)   128x128   (256 bytes)
    // 1      512x256 (4K)   256x256   (1K)
    // 2      256x512 (4K)   512x512   (4K)
    // 3      512x512 (8K)   1024x1024 (16K)
    return uint16(utils.GetVarData(uint32(b), 14, 15))
}
