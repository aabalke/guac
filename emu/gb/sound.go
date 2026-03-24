package gameboy

import (
	"github.com/aabalke/guac/emu/apu"
)

func WriteSound(addr uint32, v uint8, a *apu.Apu) {

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

	if apu.IsResetSoundChan(addr, true) {
		a.ResetSoundChan(addr, v, true)
	}

	switch addr {

	case 0x10:
		a.ToneChannel1.CntL &^= 0x00FF
		a.ToneChannel1.CntL |= uint16(v)

	case 0x11:
		a.ToneChannel1.CntH &^= 0x00FF
		a.ToneChannel1.CntH |= uint16(v)

	case 0x12:

        wasEnabled := a.ToneChannel1.DACEnabled
        a.ToneChannel1.DACEnabled = v & 0xF8 != 0
        if wasEnabled && !a.ToneChannel1.DACEnabled {
            a.ToneChannel1.ChannelEnabled = false
        }

		a.ToneChannel1.CntH &= 0x00FF
		a.ToneChannel1.CntH |= uint16(v) << 8

	case 0x13:
		a.ToneChannel1.CntX &^= 0x00FF
		a.ToneChannel1.CntX |= uint16(v)

	case 0x14:
		a.ToneChannel1.CntX &= 0x00FF
		a.ToneChannel1.CntX |= uint16(v) << 8

	case 0x15:
		return

	case 0x16:
		a.ToneChannel2.CntH &^= 0x00FF
		a.ToneChannel2.CntH |= uint16(v)

	case 0x17:
        wasEnabled := a.ToneChannel2.DACEnabled
        a.ToneChannel2.DACEnabled = v & 0xF8 != 0
        if wasEnabled && !a.ToneChannel2.DACEnabled {
            a.ToneChannel2.ChannelEnabled = false
        }

		a.ToneChannel2.CntH &= 0x00FF
		a.ToneChannel2.CntH |= uint16(v) << 8

	case 0x18:
		a.ToneChannel2.CntX &^= 0x00FF
		a.ToneChannel2.CntX |= uint16(v)

	case 0x19:
		a.ToneChannel2.CntX &= 0x00FF
		a.ToneChannel2.CntX |= uint16(v) << 8

	case 0x1A:

        wasEnabled := a.WaveChannel.DACEnabled
        a.WaveChannel.DACEnabled = v & 0x80 != 0

        if wasEnabled && !a.WaveChannel.DACEnabled {
            a.WaveChannel.ChannelEnabled = false
        }

		a.WaveChannel.CntL = uint16(v)
	case 0x1B:
		a.WaveChannel.CntH &^= 0x00FF
		a.WaveChannel.CntH |= uint16(v)
	case 0x1C:
		a.WaveChannel.CntH &= 0x00FF
		a.WaveChannel.CntH |= uint16(v) << 8
	case 0x1D:
		a.WaveChannel.CntX &^= 0x00FF
		a.WaveChannel.CntX |= uint16(v)
	case 0x1E:
		a.WaveChannel.CntX &= 0x00FF
		a.WaveChannel.CntX |= uint16(v) << 8

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

		a.NoiseChannel.CntH &= 0x00FF
		a.NoiseChannel.CntH |= uint16(v) << 8

	case 0x24:
		a.SoundCntL &^= 0x00FF
		a.SoundCntL |= uint16(v)

	case 0x25:

		a.SoundCntL &= 0x00FF
		a.SoundCntL |= uint16(v) << 8

	//case 0x82:

	//	a.SoundCntH &^= 0x00FF
	//	a.SoundCntH |= uint16(v)

	//case 0x83:

	//	a.SoundCntH &= 0x00FF
	//	a.SoundCntH |= uint16(v) << 8

	//	if resetFifoA := utils.BitEnabled(uint32(a.SoundCntH), 11); resetFifoA {
	//		a.FifoA.Length = 0
	//	}

	//	if resetFifoB := utils.BitEnabled(uint32(a.SoundCntH), 15); resetFifoB {
	//		a.FifoB.Length = 0
	//	}

	case 0x85, 0x86, 0x87:
		return

		//case 0x88:
		//    a.SoundBias &^= 0x00FF
		//    a.SoundBias |= uint16(v)

		//case 0x89:

		//    a.SoundBias &= 0x00FF
		//    a.SoundBias |= uint16(v) << 8

	default:

		//fmt.Printf("SND WRITE AT ADDR %08X\n", addr)
		//a.IO[addr] = v
	}
}

func ReadSound(addr uint32, a *apu.Apu) uint8 {

	if wave := addr >= 0x30 && addr < 0x40; wave {
		//bank := (a.WaveChannel.CntL >> 2) & 0x10
		return a.WaveChannel.WaveRam[addr & 0xF]
	}

    if addr >= 0x27 && addr < 0x30 {
        return 0xFF
    }


	//if fifo := addr >= 0xA0 && addr < 0xB0; fifo {
	//    return 0
	//}

	switch addr {
	case 0x10:
		return uint8(a.ToneChannel1.CntL) | 0x80
	case 0x11:
		return uint8(a.ToneChannel1.CntH) | 0x3F
	case 0x12:
		return uint8(a.ToneChannel1.CntH >> 8)
	case 0x13:
		return 0xFF
	case 0x14:
		return uint8(a.ToneChannel1.CntX>>8) | 0xBF
	case 0x15:
		return 0xFF

	case 0x16:
		return uint8(a.ToneChannel2.CntH) | 0x3F
	case 0x17:
		return uint8(a.ToneChannel2.CntH >> 8)
	case 0x18:
		return 0xFF
	case 0x19:
		return uint8(a.ToneChannel2.CntX>>8) | 0xBF

	case 0x1A:
		return uint8(a.WaveChannel.CntL) | 0x7F
	case 0x1B:
		return 0xFF
	case 0x1C:
		return uint8(a.WaveChannel.CntH) | 0x9F
	case 0x1D:
		return 0xFF
	case 0x1E:
		return uint8(a.WaveChannel.CntX) | 0xBF
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
	//case 0x82: return uint8(a.SoundCntH) & 0x0F
	//case 0x83: return uint8(a.SoundCntH >> 8) & 0x77
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
	//case 0x85: return 0
	//case 0x86: return 0
	//case 0x87: return 0

	//case 0x88: return uint8(a.SoundBias) &^ 0x1
	//case 0x89: return uint8(a.SoundBias >> 8) &^ 0xC3
	//case 0x8A: return 0
	//case 0x8B: return 0

	default:
		return 0
		//fmt.Printf("SND READ AT ADDR %08X\n", addr)
		//return a.IO[addr]
	}
}
