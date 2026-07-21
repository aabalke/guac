package gba

func (gba *GBA) WriteSound(addr uint32, v uint8) {
	a := gba.Apu

	switch addr {
	case 0x82:
		a.SoundCntH &^= 0x00FF
		a.SoundCntH |= uint16(v)
		return

	case 0x83:

		a.SoundCntH &= 0x00FF
		a.SoundCntH |= uint16(v) << 8

		if resetFifoA := (a.SoundCntH>>11)&1 != 0; resetFifoA {
			a.FifoA.Reset()
		}

		if resetFifoB := (a.SoundCntH>>15)&1 != 0; resetFifoB {
			a.FifoB.Reset()
		}

		return

	case 0x84:

		//v &= 0x8F // should be 0x80 but setting channel bit does not work rn

		wasEnabled := a.Enabled
		a.Enabled = v&0x80 != 0
		if !a.Enabled && wasEnabled {
			a.PowerOff()
		}

		if !wasEnabled && a.Enabled {
			a.PowerOn()
		}

		return
	case 0x85, 0x86, 0x87:
		return

	case 0x88:
		a.SoundBias &^= 0x00FF
		a.SoundBias |= uint16(v)
		return

	case 0x89:
		a.SoundBias &= 0x00FF
		a.SoundBias |= uint16(v) << 8
		return
	}

	if !a.Enabled {
		return
	}

	if wave := addr >= 0x90 && addr < 0xA0; wave {

		ch := &a.WaveChannel
		// invert bankIdx, ADD addr & 0xF
		idx := (uint8(ch.BankIdx<<4) ^ 0x10) | uint8(addr&0xF)
		ch.Ram[idx] = v
		return
	}

	if tone := addr < 0x70; tone {

		ch := &a.ToneChannel1
		if addr >= 0x68 {
			ch = &a.ToneChannel2
		}

		switch addr {
		case 0x60:

			ch.SweepStep = v & 7
			ch.SweepDecrease = (v>>3)&1 != 0
			ch.SweepPace = (v >> 4) & 7

			if ch.NegateLatch && !ch.SweepDecrease {
				ch.ChannelEnabled = false
			}

		case 0x62, 0x68:
			ch.Duty = v >> 6
			ch.ResetLength(v & 0x3F)

		case 0x63, 0x69:

			wasEnabled := ch.DACEnabled
			ch.DACEnabled = v&0xF8 != 0
			if wasEnabled && !ch.DACEnabled {
				ch.ChannelEnabled = false
			}

			ch.InitVolume = v >> 4
			ch.EnvEnabled = v&7 != 0
			ch.EnvIncrement = (v>>3)&1 != 0
			ch.EnvPace = v & 7
		case 0x64, 0x6C:
			ch.Period &^= 0x00FF
			ch.Period |= uint16(v)
		case 0x65, 0x6D:

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

	if wave := addr < 0x78; wave {
		switch ch := &a.WaveChannel; addr {
		case 0x70:

			ch.DoubleSize = (v>>5)&1 != 0
			ch.BankIdx = (v >> 6) & 1

			wasEnabled := ch.DACEnabled
			ch.DACEnabled = v&0x80 != 0

			if wasEnabled && !ch.DACEnabled {
				ch.ChannelEnabled = false
			}

		case 0x72:
			ch.ResetLength(v)

		case 0x73:
			ch.OutputLevel = (v >> 5) & 7

		case 0x74:
			ch.Period &^= 0x00FF
			ch.Period |= uint16(v)

		case 0x75:

			ch.Period &^= 0xFF00
			ch.Period |= uint16(v&7) << 8

			prev := ch.LenEnabled
			ch.LenEnabled = (v>>6)&1 != 0
			if !prev && ch.LenEnabled {
				ch.LengthTrigger()
			}

			if v&0x80 != 0 {
				ch.Trigger()
				gba.Scheduler.cancel(EVENT_SND_WAVE_CLK)
				if ch.ChannelEnabled {
					// period * 4 since gba is 4x speed
					period := (int64(2048-ch.ActivePeriod) << 1) << 2
					gba.Scheduler.schedule(EVENT_SND_WAVE_CLK, 1, period, gba.WaveRamClock, nil)
				}
			}
		}

		return
	}

	if noise := addr < 0x80; noise {

		switch ch := &a.NoiseChannel; addr {
		case 0x78: // 41
			ch.ResetLength(v & 0x3F)

		case 0x79: //42

			wasEnabled := ch.DACEnabled
			ch.DACEnabled = v&0xF8 != 0
			if wasEnabled && !ch.DACEnabled {
				ch.ChannelEnabled = false
			}

			ch.InitVolume = v >> 4
			ch.EnvEnabled = v&7 != 0
			ch.EnvIncrement = (v>>3)&1 != 0
			ch.EnvPace = v & 7

		case 0x7C:

			ch.S = v >> 4
			ch.R = v & 7
			ch.Width7 = v&0x8 != 0
			r := float64(ch.R)
			if r == 0 {
				r = 0.5
			}

			div := 1 << (ch.S + 1)
			frequency := (524288 / r) / float64(div)
			ch.CycleSamples = float64(gba.Apu.Ctx.SampleRate()) / frequency

		case 0x7D:

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
	case 0x80:
		a.SoundCntL &^= 0x00FF
		a.SoundCntL |= uint16(v)
	case 0x81:
		a.SoundCntL &= 0x00FF
		a.SoundCntL |= uint16(v) << 8

	}
}

func (gba *GBA) ReadSound(addr uint32) uint8 {
	a := gba.Apu

	if wave := addr >= 0x90 && addr < 0xA0; wave {
		bank := uint16(a.WaveChannel.BankIdx) << 4
		idx := (bank ^ 0x10) | uint16(addr)&0xF
		return a.WaveChannel.Ram[idx]
	}

	if tone := addr < 0x70; tone {

		ch := &a.ToneChannel1
		if addr >= 0x68 {
			ch = &a.ToneChannel2
		}

		switch addr {
		case 0x60:

			v := ch.SweepStep

			if ch.SweepDecrease {
				v |= 1 << 3
			}

			v |= ch.SweepPace << 4

			return v

		case 0x62, 0x68:
			return (ch.Duty << 6)

		case 0x63, 0x69:

			v := ch.EnvPace

			if ch.EnvIncrement {
				v |= 1 << 3
			}

			v |= ch.InitVolume << 4

			return v

		case 0x65, 0x6D:

			if ch.LenEnabled {
				return 0x40
			}

			return 0

		default:
			return 0
		}
	}

	if wave := addr < 0x78; wave {
		switch ch := &a.WaveChannel; addr {
		case 0x70:

			v := ch.BankIdx << 6

			if ch.DoubleSize {
				v |= 1 << 5
			}

			if ch.DACEnabled {
				v |= 0x80
			}

			return v

		case 0x73:
			return ch.OutputLevel << 5

		case 0x75:

			if ch.LenEnabled {
				return 0x40
			}

			return 0
		default:
			return 0
		}
	}

	if noise := addr < 0x80; noise {
		switch ch := &a.NoiseChannel; addr {

		case 0x79:

			v := ch.EnvPace

			if ch.EnvIncrement {
				v |= 1 << 3
			}

			v |= ch.InitVolume << 4

			return v

		case 0x7C:
			v := ch.R
			v |= ch.S << 4
			if ch.Width7 {
				v |= 1 << 3
			}

			return v

		case 0x7D:

			if ch.LenEnabled {
				return 0x40
			}

			return 0

		default:
			return 0
		}
	}

	switch addr {
	case 0x80:
		return uint8(a.SoundCntL) & 0x77
	case 0x81:
		return uint8(a.SoundCntL>>8) & 0xFF
	case 0x82:
		return uint8(a.SoundCntH) & 0x0F
	case 0x83:
		return uint8(a.SoundCntH>>8) & 0x77
	case 0x84:

		v := uint8(0)

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
	case 0x88:
		return uint8(a.SoundBias) & 0xFE
	case 0x89:
		return uint8(a.SoundBias>>8) & 0xC3
	default:
		return 0
	}
}
