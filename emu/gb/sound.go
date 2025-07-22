package gameboy

import (

	"github.com/aabalke33/guac/emu/gba/apu"
	"github.com/aabalke33/guac/emu/gba/utils"
)

func WriteSound(addr uint32, v uint8, a *apu.Apu) {

    if addr < 0x10 || addr > 0x3F { panic("ADDRS ARE NOT PROPER") }

    if addr == 0x26 {

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

    if wave := addr >= 0x30 && addr < 0x40; wave {
        idx := uint16(addr) & 0xF
        a.WaveChannel.WaveRam[idx] = v
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
        a.ToneChannel2.CntH &= 0x00FF
        a.ToneChannel2.CntH |= uint16(v) << 8

    case 0x18:
        a.ToneChannel2.CntX &^= 0x00FF
        a.ToneChannel2.CntX |= uint16(v)

    case 0x19:
        a.ToneChannel2.CntX &= 0x00FF
        a.ToneChannel2.CntX |= uint16(v) << 8

    case 0x1A:
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
        idx := uint16(addr) & 0xF
        return a.WaveChannel.WaveRam[idx]
    }

    //if fifo := addr >= 0xA0 && addr < 0xB0; fifo {
    //    return 0
    //}

    switch addr {
    case 0x10: return uint8(a.ToneChannel1.CntL) &^ 0x80
    case 0x11: return uint8(a.ToneChannel1.CntH) & 0xC0
    case 0x12: return uint8(a.ToneChannel1.CntH >> 8)
    case 0x13: return 0
    case 0x14: return uint8(a.ToneChannel1.CntX >> 8) & 0x40
    case 0x15: return 0

    case 0x16: return uint8(a.ToneChannel2.CntH) & 0xC0
    case 0x17: return uint8(a.ToneChannel2.CntH >> 8)
    case 0x18: return 0
    case 0x19: return uint8(a.ToneChannel2.CntX >> 8) & 0x40

    case 0x1A: return uint8(a.WaveChannel.CntL) & 0xE0
    case 0x1B: return 0
    case 0x1C: return uint8(a.WaveChannel.CntH) & 0xE0
    case 0x1D: return 0
    case 0x1E: return uint8(a.WaveChannel.CntX) & 0x40
    case 0x1F: return 0

    case 0x20: return 0
    case 0x21: return uint8(a.NoiseChannel.CntL >> 8)
    case 0x22: return uint8(a.NoiseChannel.CntH)
    case 0x23: return uint8(a.NoiseChannel.CntH >> 8) & 0x40

    case 0x24: return uint8(a.SoundCntL) & 0x77
    case 0x25: return uint8(a.SoundCntL >> 8) & 0xFF
    //case 0x82: return uint8(a.SoundCntH) & 0x0F
    //case 0x83: return uint8(a.SoundCntH >> 8) & 0x77
    case 0x26: return uint8(a.SoundCntX) & 0x8F
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
