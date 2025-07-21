package gba

import (

	"github.com/aabalke33/guac/emu/gba/apu"
	"github.com/aabalke33/guac/emu/gba/utils"
)

func WriteSound(addr uint32, v uint8, a *apu.Apu) {

    if addr == 0x84 {

		//v &= 0x8F // should be 0x80 but setting channel bit does not work rn

		a.SoundCntX = uint16((uint8(a.SoundCntX) & 0x0F) | (v & 0x80))

		if disabled := !utils.BitEnabled(uint32(v), 7); disabled {
            a.Disable()
		}

        return
    }

	if disabled := !utils.BitEnabled(uint32(a.SoundCntX), 7); disabled {
		return
	}

    if wave := addr >= 0x90 && addr < 0xA0; wave {

        bank := (a.WaveChannel.CntL >> 2) & 0x10
        idx := (bank ^ 0x10) | uint16(addr) & 0xF
        a.WaveChannel.WaveRam[idx] = v
        return
    }

    if apu.IsResetSoundChan(addr) {
        a.ResetSoundChan(addr, v)
    }

	switch addr {

    case 0x60:
        a.ToneChannel1.CntL &^= 0x00FF
        a.ToneChannel1.CntL |= uint16(v)

    case 0x61:
        a.ToneChannel1.CntL &= 0x00FF
        a.ToneChannel1.CntL |= uint16(v) << 8

    case 0x62:
        a.ToneChannel1.CntH &^= 0x00FF
        a.ToneChannel1.CntH |= uint16(v)

    case 0x63:
        a.ToneChannel1.CntH &= 0x00FF
        a.ToneChannel1.CntH |= uint16(v) << 8

    case 0x64:
        a.ToneChannel1.CntX &^= 0x00FF
        a.ToneChannel1.CntX |= uint16(v)

    case 0x65:
        a.ToneChannel1.CntX &= 0x00FF
        a.ToneChannel1.CntX |= uint16(v) << 8

    case 0x66:
        return

    case 0x67:
        return

    case 0x68:
        a.ToneChannel2.CntH &^= 0x00FF
        a.ToneChannel2.CntH |= uint16(v)

    case 0x69:
        a.ToneChannel2.CntH &= 0x00FF
        a.ToneChannel2.CntH |= uint16(v) << 8

    case 0x6A:
        return

    case 0x6B:
        return

    case 0x6C:
        a.ToneChannel2.CntX &^= 0x00FF
        a.ToneChannel2.CntX |= uint16(v)

    case 0x6D:
        a.ToneChannel2.CntX &= 0x00FF
        a.ToneChannel2.CntX |= uint16(v) << 8

    case 0x6E:
        return

    case 0x6F:
        return

    case 0x70: a.WaveChannel.CntL = uint16(v)
    case 0x71:
        return
    case 0x72:
        a.WaveChannel.CntH &^= 0x00FF
        a.WaveChannel.CntH |= uint16(v)
    case 0x73:
        a.WaveChannel.CntH &= 0x00FF
        a.WaveChannel.CntH |= uint16(v) << 8
    case 0x74:
        a.WaveChannel.CntX &^= 0x00FF
        a.WaveChannel.CntX |= uint16(v)
    case 0x75:
        a.WaveChannel.CntX &= 0x00FF
        a.WaveChannel.CntX |= uint16(v) << 8

    case 0x76:
        return
    case 0x77:
        return

    case 0x78:
        a.NoiseChannel.CntL &^= 0x00FF
        a.NoiseChannel.CntL |= uint16(v)

    case 0x79:
        a.NoiseChannel.CntL &= 0x00FF
        a.NoiseChannel.CntL |= uint16(v) << 8
    case 0x7A:
        return
    case 0x7B:
        return
    case 0x7C:
        a.NoiseChannel.CntH &^= 0x00FF
        a.NoiseChannel.CntH |= uint16(v)
    case 0x7D:
        a.NoiseChannel.CntH &= 0x00FF
        a.NoiseChannel.CntH |= uint16(v) << 8
    case 0x7E:
        return
    case 0x7F:
        return

    case 0x80:
        a.SoundCntL &^= 0x00FF
        a.SoundCntL |= uint16(v)

    case 0x81:

        a.SoundCntL &= 0x00FF
        a.SoundCntL |= uint16(v) << 8


	case 0x82:

		a.SoundCntH &^= 0x00FF
		a.SoundCntH |= uint16(v)

	case 0x83:

		a.SoundCntH &= 0x00FF
		a.SoundCntH |= uint16(v) << 8

		if resetFifoA := utils.BitEnabled(uint32(a.SoundCntH), 11); resetFifoA {
			a.FifoA.Length = 0
		}

		if resetFifoB := utils.BitEnabled(uint32(a.SoundCntH), 15); resetFifoB {
			a.FifoB.Length = 0
		}

	case 0x85, 0x86, 0x87:
		return

    case 0x88:
        a.SoundBias &^= 0x00FF
        a.SoundBias |= uint16(v)

    case 0x89:

        a.SoundBias &= 0x00FF
        a.SoundBias |= uint16(v) << 8

	default:

        //fmt.Printf("SND WRITE AT ADDR %08X\n", addr)
		//a.IO[addr] = v
	}
}

func ReadSound(addr uint32, a *apu.Apu) uint8 {

    if wave := addr >= 0x90 && addr < 0xA0; wave {
        bank := (a.WaveChannel.CntL >> 2) & 0x10
        idx := (bank ^ 0x10) | uint16(addr) & 0xF
        return a.WaveChannel.WaveRam[idx]
    }

    if fifo := addr >= 0xA0 && addr < 0xB0; fifo {
        return 0
    }

    switch addr {
    case 0x60: return uint8(a.ToneChannel1.CntL) &^ 0x80
    case 0x61: return 0
    case 0x62: return uint8(a.ToneChannel1.CntH) & 0xC0
    case 0x63: return uint8(a.ToneChannel1.CntH >> 8)
    case 0x64: return 0
    case 0x65: return uint8(a.ToneChannel1.CntX >> 8) & 0x40
    case 0x66: return 0
    case 0x67: return 0

    case 0x68: return uint8(a.ToneChannel2.CntH) & 0xC0
    case 0x69: return uint8(a.ToneChannel2.CntH >> 8)
    case 0x6A: return 0
    case 0x6B: return 0
    case 0x6C: return 0
    case 0x6D: return uint8(a.ToneChannel2.CntX >> 8) & 0x40
    case 0x6E: return 0
    case 0x6F: return 0

    case 0x70: return uint8(a.WaveChannel.CntL) & 0xE0
    case 0x71: return 0
    case 0x72: return 0
    case 0x73: return uint8(a.WaveChannel.CntH) & 0xE0
    case 0x74: return 0
    case 0x75: return uint8(a.WaveChannel.CntX) & 0x40
    case 0x76: return 0
    case 0x77: return 0

    case 0x78: return 0
    case 0x79: return uint8(a.NoiseChannel.CntL >> 8)
    case 0x7A: return 0
    case 0x7B: return 0
    case 0x7C: return uint8(a.NoiseChannel.CntH)
    case 0x7D: return uint8(a.NoiseChannel.CntH >> 8) & 0x40
    case 0x7E: return 0
    case 0x7F: return 0

    case 0x80: return uint8(a.SoundCntL) & 0x77
    case 0x81: return uint8(a.SoundCntL >> 8) & 0xFF
    case 0x82: return uint8(a.SoundCntH) & 0x0F
    case 0x83: return uint8(a.SoundCntH >> 8) & 0x77
    case 0x84: return uint8(a.SoundCntX) & 0x8F
    case 0x85: return 0
    case 0x86: return 0
    case 0x87: return 0

    case 0x88: return uint8(a.SoundBias) &^ 0x1
    case 0x89: return uint8(a.SoundBias >> 8) &^ 0xC3
    case 0x8A: return 0
    case 0x8B: return 0

    default:
        return 0
        //fmt.Printf("SND READ AT ADDR %08X\n", addr)
        //return a.IO[addr]
    }
}
