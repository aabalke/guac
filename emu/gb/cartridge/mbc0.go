package cartridge

import "fmt"

type Mbc0 struct {
	Cartridge *Cartridge
}

func NewMbc0(c *Cartridge) *Mbc0 {

	fmt.Printf("Cartridge ROM ONLY\n")

	if c.Type != 0 {
		// Type 8 and Type 9
		// No licensed cartridge makes use of this option. The exact behavior is unknown
		panic("unsupported gb/gbc cartridge\n")
	}

	return &Mbc0{Cartridge: c}
}

func (m *Mbc0) Read(addr uint16) uint8 {

	if addr < 0x8000 {
		return m.Cartridge.Data[addr]
	}

	return 0
}

func (m *Mbc0) Write(addr uint16, v uint8) {}
func (m *Mbc0) Save()                      {}
