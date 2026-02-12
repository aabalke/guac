package wifi

//import (
//	"encoding/binary"
//	"fmt"
//	"math/rand"
//)
//
//const (
//    CHIP_ID = 0x1440
//
//    Wep64  = 1
//    Wep128 = 2
//    Wep152 = 3
//)
//
//type Wifi struct {
//    WifiCtrl    WifiCtrl
//    ConfigPorts ConfigPorts
//    PowerDown   PowerDown
//    BaseBand    *BaseBand
//    Rf Rf
//
//    IO [0xFFFF >> 1]uint16
//
//    random      *rand.Rand
//}
//
//type WifiCtrl struct {
//    SoftwareMode uint8
//    WepMode uint8
//
//    MacAddr [6]uint8
//
//    PrevRst bool
//}
//
//type ConfigPorts struct {
//    // ports r/w not setup
//    ports [0x200 >> 1]uint16
//    rx_len_crop uint16
//}
//
//type PowerDown struct {
//
//    UsCountDisabled bool
//
//    //PowerTx
//    AutoWakeup bool
//    AutoSleep bool
//    UnknownPowerTx uint8
//}
//
//type Rf struct {
//    Type3 bool
//
//    Data uint32
//    Index uint16
//    isRead bool
//
//    TransferLength uint16
//    Unknown uint16
//}
//
//func NewWifi() *Wifi {
//
//    bb := NewBaseBand()
//
//    return &Wifi{
//        BaseBand: bb,
//        random: rand.New(rand.NewSource(0)),
//    }
//}
//
//func (w *Wifi) InitWifi(f *[]byte) {
//
//    // init from firmware
//
//    w.WifiCtrl.MacAddr = ([6]uint8)((*f)[0x36:])
//
//    w.PowerDown.AutoWakeup = ((*f)[0x5C]) & 1 != 0
//    w.PowerDown.AutoSleep = ((*f)[0x5C]>>1) & 1 != 0
//    w.PowerDown.UnknownPowerTx = ((*f)[0x5C]>>2) & 0b11
//
//    b16 := binary.LittleEndian.Uint16
//    w.ConfigPorts.ports[0x120>>1] = b16((*f)[0x4C:])
//    w.ConfigPorts.ports[0x122>>1] = b16((*f)[0x4E:])
//    w.ConfigPorts.ports[0x124>>1] = b16((*f)[0x5E:])
//    w.ConfigPorts.ports[0x128>>1] = b16((*f)[0x60:])
//    w.ConfigPorts.ports[0x130>>1] = b16((*f)[0x54:])
//    w.ConfigPorts.ports[0x132>>1] = b16((*f)[0x56:])
//	w.ConfigPorts.ports[0x140>>1] = b16((*f)[0x58:])
//	w.ConfigPorts.ports[0x142>>1] = b16((*f)[0x5A:])
//	w.ConfigPorts.ports[0x144>>1] = b16((*f)[0x52:])
//	w.ConfigPorts.ports[0x146>>1] = b16((*f)[0x44:])
//	w.ConfigPorts.ports[0x148>>1] = b16((*f)[0x46:])
//	w.ConfigPorts.ports[0x14a>>1] = b16((*f)[0x48:])
//	w.ConfigPorts.ports[0x14c>>1] = b16((*f)[0x4A:])
//	w.ConfigPorts.ports[0x150>>1] = b16((*f)[0x62:])
//	w.ConfigPorts.ports[0x154>>1] = b16((*f)[0x50:])
//
//    w.PowerDown.UsCountDisabled = true
//    w.PowerDown.AutoWakeup = true
//    w.PowerDown.AutoSleep = true
//
//    w.Rf.TransferLength = b16((*f)[0x41:]) & 0x3F
//    w.Rf.Unknown = (b16((*f)[0x41:])) & 0b0100_0001_0000_0000
//
//    w.Rf.Type3 = ((*f)[0x40]) & 3 == 3
//    w.Rf.Data = 0xC008
//    w.Rf.isRead = true
//
//    if w.Rf.Type3 {
//        panic("unsupported Type3 Wifi firmware")
//    }
//}
//
//const PRINT_IO = true
//
//func (w *Wifi) Read(addr uint32) uint8 {
//
//    if PRINT_IO {
//        fmt.Printf("WIFI R %08X\n", addr)
//    }
//
//	addr &= 0xFFFF
//
//	switch addr {
//
//    case 0x8038:
//
//        v := w.PowerDown.UnknownPowerTx << 2
//
//        if w.PowerDown.AutoWakeup {
//            v |= 1
//        }
//
//        if w.PowerDown.AutoSleep {
//            v |= (1 << 1)
//        }
//
//        return v
//
//    case 0x815A:
//        return 0
//
//    case 0x815B:
//        return 0
//
//
//    case 0x815C:
//        return w.BaseBand.ReadData
//
//    case 0x815D:
//        return 0
//
//    case 0x815E:
//        return 0
//
//    case 0x815F:
//        return 0
//
//    case 0x8180:
//        //rf busy
//        return 0
//
//    case 0x8181:
//        return 0
//
//	//case 0x8000:
//    //    return uint8(CHIP_ID & 0xFF)
//	//case 0x8001:
//    //    return uint8((CHIP_ID >> 8) & 0xFF)
//
//    //case 0x8006:
//    //    v := w.WifiCtrl.SoftwareMode
//    //    v |= w.WifiCtrl.WepMode << 3
//    //    return v
//
//    //case 0x8007:
//    //    return 0
//
//    //case 0x8018:
//    //    return w.WifiCtrl.MacAddr[0]
//    //case 0x8019:
//    //    return w.WifiCtrl.MacAddr[1]
//    //case 0x801A:
//    //    return w.WifiCtrl.MacAddr[2]
//    //case 0x801B:
//    //    return w.WifiCtrl.MacAddr[3]
//    //case 0x801C:
//    //    return w.WifiCtrl.MacAddr[4]
//    //case 0x801D:
//    //    return w.WifiCtrl.MacAddr[5]
//
//    //case 0x8036:
//
//    //    if w.PowerDown.UsCountDisabled {
//    //        return 1
//    //    }
//
//    //    return 0
//
//    //case 0x8168:
//
//    //    return uint8(w.BaseBand.OddDisabled)
//
//    //case 0x8169:
//
//    //    if w.BaseBand.DisablePorts {
//    //        return 0x80
//    //    }
//
//    //    return 0x00
//
//    default:
//        panic(fmt.Sprintf("WIFI R %04X\n", addr))
//		return uint8(w.IO[addr>>1])
//	}
//}
//
//func (w *Wifi) Write16(addr uint32, v uint16) {
//
//    if PRINT_IO {
//        fmt.Printf("WIFI W %08X %02X\n", addr, v)
//    }
//
//	addr &= 0xFFFF
//
//    switch addr {
//    case 0x8004:
//
//        println("v", v)
//
//        rst := v & 1 != 0
//
//        switch {
//        case !w.WifiCtrl.PrevRst && rst:
//
//            w.Write16(0x8034, 0x02)
//            w.Write16(0x819C, 0x46)
//            w.Write16(0x8214, 0x09)
//            w.Write16(0x827C, 0x05)
//
//        case w.WifiCtrl.PrevRst && !rst:
//
//            w.Write16(0x827C, 0xA)
//        }
//
//        if (v >>13) & 1 != 0 {
//
//            w.Write16(0x8056, 0x0000)
//            w.Write16(0x80C0, 0x0000)
//            w.Write16(0x80C4, 0x0000)
//            w.Write16(0x81A4, 0x0000)
//            w.Write16(0x8278, 0x000F)
//        }
//
//        if (v >>14) & 1 != 0 {
//            w.Write16(0x8006, 0x0000)
//            w.Write16(0x8008, 0x0000)
//            w.Write16(0x800A, 0x0000)
//            w.Write16(0x8018, 0x0000)
//            w.Write16(0x801A, 0x0000)
//            w.Write16(0x801C, 0x0000)
//            w.Write16(0x8020, 0x0000)
//            w.Write16(0x8022, 0x0000)
//            w.Write16(0x8024, 0x0000)
//            w.Write16(0x8028, 0x0000)
//            w.Write16(0x802A, 0x0000)
//            w.Write16(0x802C, 0x0707)
//            w.Write16(0x802E, 0x0000)
//            w.Write16(0x8050, 0x4000)
//            w.Write16(0x8052, 0x4800)
//            w.Write16(0x8084, 0x0000)
//            w.Write16(0x80BC, 0x0001)
//            w.Write16(0x80D0, 0x0401)
//            w.Write16(0x80D4, 0x0001)
//            w.Write16(0x80E0, 0x0008)
//            w.Write16(0x80EC, 0x3F03)
//            w.Write16(0x8194, 0x0000)
//            w.Write16(0x8198, 0x0000)
//            w.Write16(0x81A2, 0x0001)
//            w.Write16(0x8224, 0x0003)
//            w.Write16(0x8230, 0x0047)
//        }
//
//    //case 0x8000:
//    //    return
//    //case 0x8006:
//    //    w.WifiCtrl.SoftwareMode = v & 7
//    //    w.WifiCtrl.WepMode = (v >> 3) & 7
//    //    return
//
//    //case 0x8007:
//    //    return
//
//    //case 0x8018:
//    //    w.WifiCtrl.MacAddr[0] = v
//    //    return
//    //case 0x8019:
//    //    w.WifiCtrl.MacAddr[1] = v
//    //    return
//    //case 0x801A:
//    //    w.WifiCtrl.MacAddr[2] = v
//    //    return
//    //case 0x801B:
//    //    w.WifiCtrl.MacAddr[3] = v
//    //    return
//    //case 0x801C:
//    //    w.WifiCtrl.MacAddr[4] = v
//    //    return
//    //case 0x801D:
//    //    w.WifiCtrl.MacAddr[5] = v
//    //    return
//
//    case 0x8036:
//        w.PowerDown.UsCountDisabled = v & 1 != 0
//
//    case 0x8038:
//        w.PowerDown.AutoWakeup = v & 1 != 0
//        w.PowerDown.AutoSleep = (v>>1) & 1 != 0
//        w.PowerDown.UnknownPowerTx = uint8(v>>2) & 3
//
//    case 0x8120: w.ConfigPorts.ports[0x120>>1] = v
//    case 0x8122: w.ConfigPorts.ports[0x122>>1] = v
//    case 0x8124: w.ConfigPorts.ports[0x124>>1] = v
//    case 0x8126: w.ConfigPorts.ports[0x126>>1] = v
//    case 0x8128: w.ConfigPorts.ports[0x128>>1] = v
//    case 0x812A: w.ConfigPorts.ports[0x12A>>1] = v
//    case 0x812C: w.ConfigPorts.ports[0x12C>>1] = v
//    case 0x812E: w.ConfigPorts.ports[0x12E>>1] = v
//    case 0x8130: w.ConfigPorts.ports[0x130>>1] = v
//    case 0x8132: w.ConfigPorts.ports[0x132>>1] = v
//    case 0x8134: w.ConfigPorts.ports[0x134>>1] = v
//    case 0x8136: w.ConfigPorts.ports[0x136>>1] = v
//    case 0x8138: w.ConfigPorts.ports[0x138>>1] = v
//    case 0x813A: w.ConfigPorts.ports[0x13A>>1] = v
//    case 0x813C: w.ConfigPorts.ports[0x13C>>1] = v
//    case 0x813E: w.ConfigPorts.ports[0x13E>>1] = v
//    case 0x8140: w.ConfigPorts.ports[0x140>>1] = v
//    case 0x8142: w.ConfigPorts.ports[0x142>>1] = v
//    case 0x8144: w.ConfigPorts.ports[0x144>>1] = v
//    case 0x8146: w.ConfigPorts.ports[0x146>>1] = v
//    case 0x8148: w.ConfigPorts.ports[0x148>>1] = v
//    case 0x814A: w.ConfigPorts.ports[0x14A>>1] = v
//    case 0x814C: w.ConfigPorts.ports[0x14C>>1] = v
//    case 0x814E: w.ConfigPorts.ports[0x14E>>1] = v
//    case 0x8150: w.ConfigPorts.ports[0x150>>1] = v
//    case 0x8152: w.ConfigPorts.ports[0x152>>1] = v
//    case 0x8154: w.ConfigPorts.ports[0x154>>1] = v
//
//    case 0x8158:
//
//        w.BaseBand.Index = v & 0xFF
//        w.BaseBand.Direction = (v >> 12)
//
//    case 0x815A:
//        w.BaseBand.WriteData = uint8(v)
//
//    case 0x8168:
//        w.BaseBand.OddDisabled = v & 0xFF
//        w.BaseBand.DisablePorts = v & 0x8000 != 0
//
//    case 0x817C:
//
//        w.Rf.Data &= 0xFFFF
//        w.Rf.Data |= uint32(v & 0b11)<< 16
//
//        w.Rf.Index = (v >> 2) & 0x1F
//        w.Rf.isRead = (v >> 7) & 1 != 0
//
//    case 0x817E:
//
//        w.Rf.Data &^= 0xFFFF
//        w.Rf.Data |= uint32(v)
//
//    case 0x8184:
//
//        w.Rf.TransferLength = v & 0x3F
//        w.Rf.Unknown = v & 0b0100_0001_0000_0000
//
//    default:
//        panic(fmt.Sprintf("WIFI W %04X %02X\n", addr, v))
//    }
//
//	w.IO[addr>>1] = v
//}
//
//
//func (w *Wifi) getRandom() uint16 {
//
//    return uint16(w.random.Uint32()) & 0x3FF
//
//}
//
//
//type BaseBand struct {
//
//    Index     uint16
//    Direction uint16
//
//    WriteData uint8
//    ReadData uint8
//
//    //Power uint16
//
//    OddDisabled  uint16
//    DisablePorts bool
//
//    reg[0x100] uint8
//    writeable[0x100] bool
//}
//
//func NewBaseBand() *BaseBand {
//
//    b := &BaseBand{}
//
//    b.reg[0x00] = 0x6D
//    b.reg[0x5D] = 0x01
//
//    for i := range b.reg {
//        switch {
//        case
//            i >= 0x01 && i <= 0x0C,
//            i >= 0x13 && i <= 0x15,
//            i >= 0x1B && i <= 0x26,
//            i >= 0x28 && i <= 0x4C,
//            i >= 0x4E && i <= 0x5C,
//            i >= 0x62 && i <= 0x63,
//            i == 0x65,
//            i == 0x67,
//            i == 0x68:
//            b.writeable[i] = true
//        }
//    }
//
//    b.DisablePorts = true
//    b.OddDisabled  = 0xD0
//
//    return b
//}
