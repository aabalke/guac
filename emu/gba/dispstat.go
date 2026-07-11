package gba

const (
	DISP_VBL = 0b001
	DISP_HBL = 0b010
	DISP_VCF = 0b100
)

type Dispstat uint16

func (d *Dispstat) Write(v uint8, hi bool) {
	if hi {
		*d = Dispstat((uint16(*d) & 0xFF) | (uint16(v) << 8))
		return
	}

	*d = Dispstat((uint16(*d) &^ 0b0011_1000) | uint16(v&^7))
}

func (d *Dispstat) GetLYC() uint8 {
	return uint8(*d >> 8)
}
