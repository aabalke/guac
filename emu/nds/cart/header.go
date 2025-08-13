package cart

import (
	"encoding/binary"
	"fmt"
	"strings"
)

type Header struct {
	Title    string
	GameCode string
	//MakerCode

	UnitCode       uint8
	EncryptionSeed uint8
	CapacityShift  uint8
	//Region

	Version uint8

	Arm9Offset    uint32
	Arm9EntryAddr uint32
	Arm9RamAddr   uint32
	Arm9Size      uint32

	Arm7Offset    uint32
	Arm7EntryAddr uint32
	Arm7RamAddr   uint32
	Arm7Size      uint32
}

func NewHeader(c *Cartridge) Header {
	h := Header{
		Title:          strings.ToUpper(strings.ReplaceAll(string(c.Rom[:0xC]), "\x00", " ")),
		GameCode:       strings.ToUpper(string(c.Rom[0xC : 0xC+4])),
		UnitCode:       c.Rom[0x12],
		EncryptionSeed: c.Rom[0x13],
		CapacityShift:  c.Rom[0x14],
		Version:        c.Rom[0x1E],

		Arm9Offset:    binary.LittleEndian.Uint32(c.Rom[0x20:]),
		Arm9EntryAddr: binary.LittleEndian.Uint32(c.Rom[0x24:]),
		Arm9RamAddr:   binary.LittleEndian.Uint32(c.Rom[0x28:]),
		Arm9Size:      binary.LittleEndian.Uint32(c.Rom[0x2C:]),
		Arm7Offset:    binary.LittleEndian.Uint32(c.Rom[0x30:]),
		Arm7EntryAddr: binary.LittleEndian.Uint32(c.Rom[0x34:]),
		Arm7RamAddr:   binary.LittleEndian.Uint32(c.Rom[0x38:]),
		Arm7Size:      binary.LittleEndian.Uint32(c.Rom[0x3C:]),
	}

	h.validate()

	println(h.Title)
	println(h.GameCode)
	println(h.UnitCode)
	fmt.Printf("ARM9 OFF %08X\n", h.Arm9Offset)
	fmt.Printf("ARM9 ENT %08X\n", h.Arm9EntryAddr)
	fmt.Printf("ARM9 RAM %08X\n", h.Arm9RamAddr)
	fmt.Printf("ARM9 SIZ %08X\n", h.Arm9Size)

	fmt.Printf("ARM7 OFF %08X\n", h.Arm7Offset)
	fmt.Printf("ARM7 ENT %08X\n", h.Arm7EntryAddr)
	fmt.Printf("ARM7 RAM %08X\n", h.Arm7RamAddr)
	fmt.Printf("ARM7 SIZ %08X\n", h.Arm7Size)
	//fmt.Printf("ROM  VAL %08X\n", binary.LittleEndian.Uint32(c.Rom[0x4008:]))

	return h
}

func (h *Header) validate() {
	if dsiOnly := !(h.UnitCode == 0 || h.UnitCode == 2); dsiOnly {
		panic("CARTRIDGE IS DSI ONLY")
	}
}
