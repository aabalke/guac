package gameboy

import (
	"fmt"

	"github.com/aabalke/guac/emu/gb/apu"
)

func (gb *GameBoy) WriteSound(addr uint32, v uint8, a *apu.Apu) {

	if addr == 0x26 {
		wasEnabled := a.Enabled
		a.Enabled = v&0x80 != 0
		if !a.Enabled && wasEnabled {
			a.PowerOff()
		}

		if !wasEnabled && a.Enabled {
			a.PowerOn()
		}

		return
	}

	if !gb.Apu.Enabled && !gb.Color {

		switch addr {
		case 0x11:
			a.ToneChannel1.ResetLength(v & 0x3F)
		case 0x16:
			a.ToneChannel2.ResetLength(v & 0x3F)
		case 0x1B:
			a.WaveChannel.ResetLength(v)
		case 0x20:
			a.NoiseChannel.ResetLength(v & 0x3F)
		}

		return
	}

	if tone := addr < 0x1A; tone {
		ch := &a.ToneChannel1
		if addr >= 0x16 {
			ch = &a.ToneChannel2
		}

		switch addr {
		case 0x10:

			ch.SweepStep = v & 7
			ch.SweepDecrease = (v>>3)&1 != 0
			ch.SweepPace = (v >> 4) & 7

            if ch.NegateLatch && !ch.SweepDecrease {
                ch.ChannelEnabled = false
            }

		case 0x11, 0x16:

			ch.Duty = v >> 6
			ch.ResetLength(v & 0x3F)

		case 0x12, 0x17:

			wasEnabled := ch.DACEnabled
			ch.DACEnabled = v&0xF8 != 0
			if wasEnabled && !ch.DACEnabled {
				ch.ChannelEnabled = false
			}

			ch.InitVolume = v >> 4
			ch.EnvEnabled = v&7 != 0
			ch.EnvIncrement = (v>>3)&1 != 0
			ch.EnvPace = v & 7

		case 0x13, 0x18:
			ch.Period &^= 0x00FF
			ch.Period |= uint16(v)

		case 0x14, 0x19:

			ch.Period &^= 0xFF00
			ch.Period |= uint16(v&7) << 8

			prev := ch.LenEnabled
			ch.LenEnabled = (v>>6)&1 != 0

			if !prev && ch.LenEnabled {
				ch.LengthTrigger()
			}

			if (v & 0x80) != 0 {
				ch.Trigger()
			}
		}

		return
	}

	if wave := addr < 0x20; wave {
		switch ch := &a.WaveChannel; addr {
		case 0x1A:

			wasEnabled := ch.DACEnabled
			ch.DACEnabled = v&0x80 != 0

			if wasEnabled && !ch.DACEnabled {
				ch.ChannelEnabled = false
			}

		case 0x1B:
			ch.ResetLength(v)

		case 0x1C:
			ch.OutputLevel = (v >> 5) & 3
		case 0x1D:
			ch.Period &^= 0x00FF
			ch.Period |= uint16(v)

		case 0x1E:

			ch.Period &^= 0xFF00
			ch.Period |= uint16(v&7) << 8

			prev := ch.LenEnabled
			ch.LenEnabled = (v>>6)&1 != 0
			if !prev && ch.LenEnabled {
				ch.LengthTrigger()
			}

			if v&0x80 != 0 {
				ch.Trigger()
			}
		}

		return
	}

	if noise := addr < 0x24; noise {

		switch ch := &a.NoiseChannel; addr {
		case 0x20: // 41
			ch.ResetLength(v & 0x3F)

		case 0x21: //42

			wasEnabled := ch.DACEnabled
			ch.DACEnabled = v&0xF8 != 0
			if wasEnabled && !ch.DACEnabled {
				ch.ChannelEnabled = false
			}

			ch.InitVolume = v >> 4
			ch.EnvEnabled = v&7 != 0
			ch.EnvIncrement = (v>>3)&1 != 0
			ch.EnvPace = v & 7

		case 0x22:

			ch.S = v >> 4
			ch.R = v & 7
			ch.Width7 = v&0x8 != 0

		case 0x23:

			prev := ch.LenEnabled
			ch.LenEnabled = (v>>6)&1 != 0

			if !prev && ch.LenEnabled {
				ch.LengthTrigger()
			}

			if v&0x80 != 0 {
				ch.Trigger()
			}
		}

		return
	}

	switch addr {
	case 0x24:
		a.Master = v
		return

	case 0x25:
		a.PanReg = v
		return
	}

	if addr < 0x30 {
		return
	}

	if wave := addr < 0x40; wave {
        a.WaveChannel.Ram[(addr - 0x30) & 0xF] = v
        a.WaveChannel.UpdateCachedRam(addr, v)
		return
	}
}

func (gb *GameBoy) ReadSound(addr uint32, a *apu.Apu) uint8 {

	//fmt.Printf("R ADDR %02X\n", addr)

	if wave := addr >= 0x30 && addr < 0x40; wave {

        return a.WaveChannel.Ram[(addr - 0x30)&0xF]
        //return 0xFF
        ch := &a.WaveChannel

        ////if ch.WavRamAccessClock <= 0 || !ch.ChannelEnabled {
        ////    return 0xFF
        ////}

        //return 0x00
        //return 0xDE

        addr -= 0x30
        addr >>= 1

        a := ch.Buffer[addr+0]
        b := ch.Buffer[addr+1] << 4
        return a | b

		////bank := (a.WaveChannel.CntL >> 2) & 0x10
        //return a.WaveChannel.WaveRam[((a.WaveChannel.WavePosition>>1)&0x1F)+0x40-addr]
	}

	if addr >= 0x27 && addr < 0x30 {
		return 0xFF
	}

	if tone := addr >= 0x10 && addr < 0x1A; tone {

		ch := &a.ToneChannel1
		if addr >= 0x16 {
			ch = &a.ToneChannel2
		}

		switch addr {
		case 0x10:

			v := ch.SweepStep

			if ch.SweepDecrease {
				v |= 1 << 3
			}

			v |= ch.SweepPace << 4

			return v | 0x80

		case 0x11, 0x16:
			return (ch.Duty << 6) | 0x3F

		case 0x12, 0x17:

			v := ch.EnvPace

			if ch.EnvIncrement {
				v |= 1 << 3
			}

			v |= ch.InitVolume << 4

			return v

		case 0x14, 0x19:

			if ch.LenEnabled {
				return 0xFF
			}

			return 0xBF

		default:
			return 0xFF
		}
	}

	if wave := addr >= 0x1A && addr < 0x20; wave {

		switch ch := &a.WaveChannel; addr {
		case 0x1A:

			if ch.DACEnabled {
				return 0xFF
			}

			return 0x7F

		case 0x1C:
			return (ch.OutputLevel << 5) | 0x9F

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

		switch ch := &a.NoiseChannel; addr {

		case 0x21:

			v := ch.EnvPace

			if ch.EnvIncrement {
				v |= 1 << 3
			}

			v |= ch.InitVolume << 4

			return v

		case 0x22:
			v := ch.R
			v |= (ch.S) << 4
			if ch.Width7 {
				v |= 1 << 3
			}

			return v

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
		return a.Master
	case 0x25:
		return a.PanReg
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
