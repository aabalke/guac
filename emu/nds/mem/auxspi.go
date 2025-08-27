package mem

import "github.com/aabalke/guac/emu/nds/utils"

type ExMem struct {
	isCartAccessArm7         bool
	isMainMemorySync         bool
	isMainMemoryPriorityArm7 bool
	v                        uint16
}

// gba values are separate instances on arm7 adn arm9

func (e *ExMem) Read(b uint8) uint8 {
	return uint8(e.v >> (b << 3))
}

func (e *ExMem) Write(v uint8, b uint8) {

	// top bits can only be written by arm9

	e.v &^= 0xFF << (b << 3)
	e.v |= uint16(v) << (b << 3)

	if b == 1 {
		e.isCartAccessArm7 = utils.BitEnabled(uint32(v), 3)
		e.isMainMemorySync = utils.BitEnabled(uint32(v), 6)
		e.isMainMemoryPriorityArm7 = utils.BitEnabled(uint32(v), 7)
	}
}

type AuxSPI struct {
    Baudrate uint8
    Hold bool
    Busy bool
    SerialMode bool
    Irq bool
    Enabled bool

    Data uint8
}

func (a *AuxSPI) Read(b uint8) uint8 {


    switch b {
    case 0:

        v := a.Baudrate

        if a.Hold {
            v |= 0b1000000
        }

        if a.Busy {
            v |= 0b10000000
        }

        return v

    case 1:

        v := uint8(0)

        if a.SerialMode {
            v |= 0b100000
        }
        if a.SerialMode {
            v |= 0b1000000
        }
        if a.SerialMode {
            v |= 0b10000000
        }

        return v

    case 2:
        return a.Data
    default:
        return 0
    }
}

func (a *AuxSPI) Write(v uint8, b uint8) {

	// top bits can only be written by arm9

    if b == 0 {
        a.Baudrate = v & 0b11
        a.Hold = utils.BitEnabled(uint32(v), 6)
        a.Busy = utils.BitEnabled(uint32(v), 7)
        return
    }

    a.SerialMode = utils.BitEnabled(uint32(v), 5)
    a.Irq = utils.BitEnabled(uint32(v), 6)
    a.Enabled = utils.BitEnabled(uint32(v), 7)
}

func (a *AuxSPI) WriteData(_ uint8, b uint8) {

    // start transfer with write but value does not matter
    a.Busy = true

    // spi stuff
    a.Busy = false
}
