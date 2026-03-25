package gameboy

import "github.com/aabalke/guac/emu/gb/apu"

func (gb *GameBoy) WriteSound(addr uint32, v uint8, a *apu.Apu) {

	if addr < 0x10 || addr > 0x3F {
		panic("ADDRS ARE NOT PROPER")
	}

	if addr == 0x26 {
        a.SoundCntX &^= 0x80
        a.SoundCntX |= uint16(v & 0x80)

        if v & 0x80 == 0 {
            a.PowerOff()
        }
		return
	}

	if disabled := (a.SoundCntX>>7)&1 == 0; disabled {
		return
	}

	if wave := addr >= 0x30 && addr < 0x40; wave {
		a.WaveChannel.WaveRam[addr & 0xF] = v
		return
	}

    if addr >= 0x27 && addr < 0x30 {
        return
    }

	switch addr {

	case 0x10:
		a.ToneChannel1.CntL &^= 0x00FF
		a.ToneChannel1.CntL |= uint16(v)

	case 0x11:

        a.ToneChannel1.Duty = v >> 6
        a.ToneChannel1.ResetLength(v & 0x3F, false)

	case 0x12:

        wasEnabled := a.ToneChannel1.DACEnabled
        a.ToneChannel1.DACEnabled = v & 0xF8 != 0
        if wasEnabled && !a.ToneChannel1.DACEnabled {
            a.ToneChannel1.ChannelEnabled = false
        }

        a.ToneChannel1.VolumeRegister = v

	case 0x13:
        a.ToneChannel1.Period &^= 0x00FF
        a.ToneChannel1.Period |= uint16(v)

	case 0x14:

        a.ToneChannel1.Period &^= 0xFF00
        a.ToneChannel1.Period |= uint16(v & 7) << 8
        a.ToneChannel1.LenEnabled = (v >> 6) & 1 != 0

        if v & 0x80 != 0 {
            a.ToneChannel1.Trigger()
        }

	case 0x15:
		return

	case 0x16:

        a.ToneChannel2.Duty = v >> 6
        a.ToneChannel2.ResetLength(v & 0x3F, false)

	case 0x17:


        wasEnabled := a.ToneChannel2.DACEnabled
        a.ToneChannel2.DACEnabled = v & 0xF8 != 0
        if wasEnabled && !a.ToneChannel2.DACEnabled {
            a.ToneChannel2.ChannelEnabled = false
        }

        a.ToneChannel2.VolumeRegister = v

	case 0x18:
        a.ToneChannel2.Period &^= 0x00FF
        a.ToneChannel2.Period |= uint16(v)

	case 0x19:

        a.ToneChannel2.Period &^= 0xFF00
        a.ToneChannel2.Period |= uint16(v & 7) << 8
        a.ToneChannel2.LenEnabled = (v >> 6) & 1 != 0

        if v & 0x80 != 0 {
            a.ToneChannel2.Trigger()
        }

	case 0x1A:

        wasEnabled := a.WaveChannel.DACEnabled
        a.WaveChannel.DACEnabled = v & 0x80 != 0

        if wasEnabled && !a.WaveChannel.DACEnabled {
            a.WaveChannel.ChannelEnabled = false
        }

	case 0x1B:
        a.WaveChannel.ResetLength(v, false)

	case 0x1C:
		a.WaveChannel.CntH &= 0x00FF
		a.WaveChannel.CntH |= uint16(v) << 8
	case 0x1D:
        a.WaveChannel.Period &^= 0x00FF
        a.WaveChannel.Period |= uint16(v)

	case 0x1E:

        a.WaveChannel.Period &^= 0xFF00
        a.WaveChannel.Period |= uint16(v & 7) << 8
        a.WaveChannel.LenEnabled = (v >> 6) & 1 != 0

        if v & 0x80 != 0 {
            a.WaveChannel.Trigger()
        }

	case 0x1F:
		return
	case 0x20:
		a.NoiseChannel.CntL &^= 0x00FF
		a.NoiseChannel.CntL |= uint16(v)

	case 0x21:
        wasEnabled := a.NoiseChannel.DACEnabled
        a.NoiseChannel.DACEnabled = v & 0xF8 != 0
        if wasEnabled && !a.NoiseChannel.DACEnabled {
            a.NoiseChannel.ChannelEnabled = false
        }

		a.NoiseChannel.CntL &= 0x00FF
		a.NoiseChannel.CntL |= uint16(v) << 8
	case 0x22:
		a.NoiseChannel.CntH &^= 0x00FF
		a.NoiseChannel.CntH |= uint16(v)
	case 0x23:
        if apu.IsResetSoundChan(addr, true) {
            a.ResetSoundChan(addr, v, true)
        }


		a.NoiseChannel.CntH &= 0x00FF
		a.NoiseChannel.CntH |= uint16(v) << 8

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

	switch addr {
	case 0x10:
		return uint8(a.ToneChannel1.CntL) | 0x80
	case 0x11:
		return (a.ToneChannel1.Duty << 6) | 0x3F
	case 0x12:
		return a.ToneChannel1.VolumeRegister
	case 0x13:
		return 0xFF
	case 0x14:

        if a.ToneChannel1.LenEnabled {
            return 0xFF
        }

        return 0xBF

	case 0x15:
		return 0xFF

	case 0x16:
		return (a.ToneChannel2.Duty << 6) | 0x3F
	case 0x17:
		return a.ToneChannel2.VolumeRegister
	case 0x18:
		return 0xFF
	case 0x19:

        if a.ToneChannel2.LenEnabled {
            return 0xFF
        }

        return 0xBF

	case 0x1A:

        if a.WaveChannel.DACEnabled {
            return 0xFF
        }

        return 0x7F

	case 0x1B:
		return 0xFF
	case 0x1C:
		return uint8(a.WaveChannel.CntH) | 0x9F
	case 0x1D:
		return 0xFF
	case 0x1E:

        if a.WaveChannel.LenEnabled {
            return 0xFF
        }

        return 0xBF

	case 0x1F:
		return 0xFF

	case 0x20:
		return 0xFF
	case 0x21:
		return uint8(a.NoiseChannel.CntL >> 8)
	case 0x22:
		return uint8(a.NoiseChannel.CntH)
	case 0x23:
		return uint8(a.NoiseChannel.CntH>>8) | 0xBF

	case 0x24:
		return uint8(a.SoundCntL)
	case 0x25:
		return uint8(a.SoundCntL>>8)
	case 0x26:

        v := uint8(a.SoundCntX) & 0x80 | 0x70

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

	default:
		return 0
	}
}
