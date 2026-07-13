package gpio

type Gpio struct {
	Devices  []GpioDevice
	readable bool
	out      uint8
}

type GpioDevice interface {
	Read() uint8
	Write(v uint8)
}

func NewGpio() *Gpio {
	return &Gpio{
		Devices: []GpioDevice{},
	}
}

func (g *Gpio) Read(addr uint32) uint8 {
	if addr&1 == 1 {
		return 0
	}

	if !g.readable {
		return 0
	}

	switch addr {
	case 0x800_00C4:

		v := uint8(0)

		for _, device := range g.Devices {
			v |= device.Read()
		}

		return v

	case 0x800_00C6:
		return g.out

	case 0x800_00C8:

		if g.readable {
			return 1
		}

		return 0
	}

	return 0
}

func (g *Gpio) Write(addr uint32, v uint8) {
	if addr&1 == 1 {
		return
	}

	//fmt.Printf("Write %08X %02X\n", addr, v)
	switch addr {
	case 0x800_00C4:

		for _, device := range g.Devices {
			device.Write(v & g.out)
		}

	case 0x800_00C6:
		g.out = v & 0xF

	case 0x800_00C8:
		g.readable = v&1 != 0
	}
}
