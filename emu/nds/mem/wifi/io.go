package wifi

import (
	"encoding/binary"
	"fmt"
)

const (
    CHIP_ID = 0x1440

    Wep64  = 1
    Wep128 = 2
    Wep152 = 3
)

type Wifi struct {
    WifiCtrl WifiCtrl
    ConfigPorts ConfigPorts
    PowerDown PowerDown
    IO [0xFFFF]uint8
}

type WifiCtrl struct {
    SoftwareMode uint8
    WepMode uint8

    MacAddr [6]uint8
}

type ConfigPorts struct {
    // ports r/w not setup
    ports [0x200]uint16
    rx_len_crop uint16
}

type PowerDown struct {

    //PowerTx
    AutoWakeup bool
    AutoSleep bool
    UnknownPowerTx uint8
}


func NewWifi() *Wifi {
    return &Wifi{}
}

func (w *Wifi) InitWifi(f *[]byte) {

    // init from firmware

    w.WifiCtrl.MacAddr = ([6]uint8)((*f)[0x36:])

    w.PowerDown.AutoWakeup = ((*f)[0x5C]) & 1 != 0
    w.PowerDown.AutoSleep = ((*f)[0x5C]>>1) & 1 != 0
    w.PowerDown.UnknownPowerTx = ((*f)[0x5C]>>2) & 0b11

    b16 := binary.LittleEndian.Uint16
    w.ConfigPorts.ports[0x120] = b16((*f)[0x4C:])
    w.ConfigPorts.ports[0x122] = b16((*f)[0x4E:])
    w.ConfigPorts.ports[0x124] = b16((*f)[0x5E:])
    w.ConfigPorts.ports[0x128] = b16((*f)[0x60:])
    w.ConfigPorts.ports[0x130] = b16((*f)[0x54:])
    w.ConfigPorts.ports[0x132] = b16((*f)[0x56:])
	w.ConfigPorts.ports[0x140] = b16((*f)[0x58:])
	w.ConfigPorts.ports[0x142] = b16((*f)[0x5A:])
	w.ConfigPorts.ports[0x144] = b16((*f)[0x52:])
	w.ConfigPorts.ports[0x146] = b16((*f)[0x44:])
	w.ConfigPorts.ports[0x148] = b16((*f)[0x46:])
	w.ConfigPorts.ports[0x14a] = b16((*f)[0x48:])
	w.ConfigPorts.ports[0x14c] = b16((*f)[0x4A:])
	w.ConfigPorts.ports[0x150] = b16((*f)[0x62:])
	w.ConfigPorts.ports[0x154] = b16((*f)[0x50:])
}

const PRINT_IO = false

func (w *Wifi) Read(addr uint32) uint8 {

    if PRINT_IO {
        fmt.Printf("WIFI R %08X\n", addr)
    }

	addr &= 0xFFFF

	switch addr {
	case 0x8000:
        return uint8(CHIP_ID & 0xFF)
	case 0x8001:
        return uint8((CHIP_ID >> 8) & 0xFF)

    case 0x8006:
        v := w.WifiCtrl.SoftwareMode
        v |= w.WifiCtrl.WepMode << 3
        return v

    case 0x8007:
        return 0

    case 0x8018:
        return w.WifiCtrl.MacAddr[0]
    case 0x8019:
        return w.WifiCtrl.MacAddr[1]
    case 0x801A:
        return w.WifiCtrl.MacAddr[2]
    case 0x801B:
        return w.WifiCtrl.MacAddr[3]
    case 0x801C:
        return w.WifiCtrl.MacAddr[4]
    case 0x801D:
        return w.WifiCtrl.MacAddr[5]


    default:
        //panic(fmt.Sprintf("WIFI R %04X\n", addr))
		return w.IO[addr]
	}
}

func (w *Wifi) Write16(addr uint32, v uint16) {
    w.Write(addr, uint8(v))
    w.Write(addr+1, uint8(v>>8))
}

func (w *Wifi) Write(addr uint32, v uint8) {

    if PRINT_IO {
        fmt.Printf("WIFI W %08X %02X\n", addr, v)
    }

	addr &= 0xFFFF

    switch addr {
    case 0x8000:
        return
    case 0x8001:
        return
    case 0x8006:
        w.WifiCtrl.SoftwareMode = v & 7
        w.WifiCtrl.WepMode = (v >> 3) & 7
        return

    case 0x8007:
        return

    case 0x8018:
        w.WifiCtrl.MacAddr[0] = v
        return
    case 0x8019:
        w.WifiCtrl.MacAddr[1] = v
        return
    case 0x801A:
        w.WifiCtrl.MacAddr[2] = v
        return
    case 0x801B:
        w.WifiCtrl.MacAddr[3] = v
        return
    case 0x801C:
        w.WifiCtrl.MacAddr[4] = v
        return
    case 0x801D:
        w.WifiCtrl.MacAddr[5] = v
        return
    default:
        //panic(fmt.Sprintf("WIFI W %04X %02X\n", addr, v))
    }

	w.IO[addr] = v
}
