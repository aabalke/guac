package gameboy

import (
	"fmt"

	"github.com/aabalke/guac/emu/gb/apu"
)

func (gb *GameBoy) WriteSound(addr uint32, v uint8, a *apu.Apu) {

	if addr == 0x26 {

        a.Enabled = v & 0x80 != 0
        if !a.Enabled {
            a.PowerOff()
        }

		return
	}

    if !gb.Apu.Enabled {
		return
	}

	if wave := addr >= 0x30 && addr < 0x40; wave {
		a.WaveChannel.WaveRam[addr & 0xF] = v
		return
	}

    if addr >= 0x27 && addr < 0x30 {
        return
    }

    if tone1 := addr >= 0x10 && addr < 0x16; tone1 {

        ch := &a.ToneChannel1

        switch addr {
        case 0x10:

            ch.SweepStep = v & 7
            ch.SweepDecrease = (v >> 3) & 1 != 0
            ch.SweepPace = (v >> 4) & 7

        case 0x11:

            ch.Duty = v >> 6
            //ch.ResetLength(v & 0x3F, false)

        case 0x12:

            wasEnabled := ch.DACEnabled
            ch.DACEnabled = v & 0xF8 != 0
            if wasEnabled && !ch.DACEnabled {
                ch.ChannelEnabled = false
            }

            ch.InitVolume = v >> 4
            ch.EnvEnabled = v & 7 != 0
            ch.EnvIncrement  = (v >> 3) & 1 != 0
            ch.EnvPace  = v & 7

            //fmt.Printf("W INIT VOL %X, PACE %X\n", ch.InitVolume, v & 7)

        case 0x13:
            ch.Period &^= 0x00FF
            ch.Period |= uint16(v)

        case 0x14:
            ch.Period &^= 0xFF00
            ch.Period |= uint16(v & 7) << 8
            ch.LenEnabled = (v >> 6) & 1 != 0
            ch.ChannelEnabled = (v >> 7) & 1 != 0
            if ch.ChannelEnabled {
                ch.Trigger()
            }
        }

        return
    }

    if tone2 := addr >= 0x16 && addr < 0x1A; tone2 {

        ch := &a.ToneChannel2

        switch addr {
        case 0x16:

            ch.Duty = v >> 6
            //ch.ResetLength(v & 0x3F, false)

        case 0x17:

            wasEnabled := ch.DACEnabled
            ch.DACEnabled = v & 0xF8 != 0
            if wasEnabled && !ch.DACEnabled {
                ch.ChannelEnabled = false
            }

            ch.InitVolume = v >> 4
            ch.EnvEnabled = v & 7 != 0
            ch.EnvIncrement  = (v >> 3) & 1 != 0
            ch.EnvPace  = v & 7

        case 0x18:
            ch.Period &^= 0x00FF
            ch.Period |= uint16(v)

        case 0x19:

            ch.Period &^= 0xFF00
            ch.Period |= uint16(v & 7) << 8
            ch.LenEnabled = (v >> 6) & 1 != 0

            if v & 0x80 != 0 {
                ch.Trigger()
            }
        }

        return
    }

    if wave := addr >= 0x1A && addr < 0x20; wave {

        ch := &a.WaveChannel

        switch addr {
        case 0x1A:

            wasEnabled := ch.DACEnabled
            ch.DACEnabled = v & 0x80 != 0

            if wasEnabled && !ch.DACEnabled {
                ch.ChannelEnabled = false
            }

        case 0x1B:
            ch.ResetLength(v, false)

        case 0x1C:
            ch.CntH &= 0x00FF
            ch.CntH |= uint16(v) << 8
        case 0x1D:
            ch.Period &^= 0x00FF
            ch.Period |= uint16(v)

        case 0x1E:

            ch.Period &^= 0xFF00
            ch.Period |= uint16(v & 7) << 8
            ch.LenEnabled = (v >> 6) & 1 != 0

            if v & 0x80 != 0 {
                ch.Trigger()
            }
        }

        return
    }

    if noise := addr >= 0x20 && addr < 0x24; noise {

        ch := &a.NoiseChannel

        switch addr {
        case 0x20:
            ch.ResetLength(v & 0x3F, false)

        case 0x21:
            wasEnabled := ch.DACEnabled
            ch.DACEnabled = v & 0xF8 != 0
            if wasEnabled && !ch.DACEnabled {
                ch.ChannelEnabled = false
            }

            ch.VolumeRegister = v

        case 0x22:
            ch.RandomRegister = v
        case 0x23:
            ch.LenEnabled = (v >> 6) & 1 != 0

            if v & 0x80 != 0 {
                ch.Trigger()
            }
        }

        return
    }

	switch addr {
	case 0x24:
		a.SoundCntL &^= 0x00FF
		a.SoundCntL |= uint16(v)

	case 0x25:
		a.SoundCntL &= 0x00FF
		a.SoundCntL |= uint16(v) << 8
	}
}

func (gb *GameBoy) ReadSound(addr uint32, a *apu.Apu) uint8 {

	if wave := addr >= 0x30 && addr < 0x40; wave {
		//bank := (a.WaveChannel.CntL >> 2) & 0x10
		return a.WaveChannel.WaveRam[addr & 0xF]
	}

    if addr >= 0x27 && addr < 0x30 {
        return 0xFF
    }

    if tone1 := addr >= 0x10 && addr < 0x16; tone1 {

        ch := &a.ToneChannel1

        switch addr {
        case 0x10:

            v := ch.SweepStep

            if ch.SweepDecrease {
                v |= 1 << 3
            }

            v |= ch.SweepPace << 4

            return v | 0x80

        case 0x11:
            return (ch.Duty << 6) | 0x3F

        case 0x12:

            v := ch.EnvPace

            if ch.EnvIncrement {
                v |= 1 << 3
            }

            v |= ch.InitVolume << 4

            return v

        case 0x14:

            if ch.LenEnabled {
                return 0xFF
            }

            return 0xBF

        default:
            return 0xFF
        }
    }

    if tone2 := addr >= 0x16 && addr < 0x1A; tone2 {

        ch := &a.ToneChannel2

        switch addr {
        case 0x16:
            return (ch.Duty << 6) | 0x3F

        case 0x17:

            v := ch.EnvPace

            if ch.EnvIncrement {
                v |= 1 << 3
            }

            v |= ch.InitVolume << 4

            return v

        case 0x19:

            if ch.LenEnabled {
                return 0xFF
            }

            return 0xBF

        default:
            return 0xFF
        }
    }

    if wave := addr >= 0x1A && addr < 0x20; wave {

        ch := &a.WaveChannel

        switch addr {
        case 0x1A:

            if ch.DACEnabled {
                return 0xFF
            }

            return 0x7F

        case 0x1C:
            return uint8(ch.CntH>>8) | 0x9F

        case 0x1E:

            if ch.LenEnabled {
                return 0xFF
            }

            return 0xBF
        default:
            return 0xFF
        }
    }

    if noise := addr >= 0x20 && addr < 0x24; noise {

        ch := &a.NoiseChannel

        switch addr {
        case 0x21:
            return ch.VolumeRegister

        case 0x22:
            return ch.RandomRegister

        case 0x23:

            if ch.LenEnabled {
                return 0xFF
            }

            return 0xBF

        default:
            return 0xFF
        }
    }

	switch addr {
	case 0x24:
		return uint8(a.SoundCntL)
	case 0x25:
		return uint8(a.SoundCntL>>8)
	case 0x26:

        v := uint8(0x70)

        if a.Enabled {
            v |= 1 << 7
        }

        if a.ToneChannel1.ChannelEnabled {
            v |= 1 << 0
        }

        if a.ToneChannel2.ChannelEnabled {
            v |= 1 << 1
        }

        if a.WaveChannel.ChannelEnabled {
            v |= 1 << 2
        }

        if a.NoiseChannel.ChannelEnabled {
            v |= 1 << 3
        }

        return v
	}

    panic(fmt.Sprintf("not possible read sound %04X", addr))
}
