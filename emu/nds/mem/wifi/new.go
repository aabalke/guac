package wifi

import (
	"math/rand"
)

type Wifi struct {
	WRxBufBegin   uint16//hwio.Reg16 `hwio:"offset=0x50"`
	WRxBufEnd     uint16//hwio.Reg16 `hwio:"offset=0x52"`
	WRxBufRdAddr  uint16//hwio.Reg16 `hwio:"offset=0x58,rwmask=0x1FFF"`
	//WRxBufRdData hwio.Reg16 `hwio:"offset=0x60,readonly,rcb"`

	WTxBufWrAddr  uint16 //hwio.Reg16 `hwio:"offset=0x68,rwmask=0x1FFF"`
	//WTxBufWrData  hwio.Reg16 `hwio:"offset=0x70,writeonly,wcb"`
	WTxBufGapTop  uint16//hwio.Reg16 `hwio:"offset=0x74,rwmask=0x1FFF"`
	WTxBufGapDisp uint16//hwio.Reg16 `hwio:"offset=0x76,rwmask=0xFFF"`

	//BaseBandCnt   hwio.Reg16 `hwio:"offset=0x158,wcb"`
	BaseBandWrite uint16//hwio.Reg16 `hwio:"offset=0x15A,writeonly"`
	BaseBandRead  uint16//hwio.Reg16 `hwio:"offset=0x15C,readonly"`
	BaseBandBusy  uint16//hwio.Reg16 `hwio:"offset=0x15E,readonly"`
	BaseBandMode  uint16//hwio.Reg16 `hwio:"offset=0x160"`
	BaseBandPower uint16//hwio.Reg16 `hwio:"offset=0x168"`
	bbRegWritable [256]bool
	bbRegs        [256]uint8

	//Random hwio.Reg16 `hwio:"offset=0x044,readonly,rcb"`
	rand   *rand.Rand

    ram [0x2000>>1]uint16
	//WifiRam hwio.Mem `hwio:"bank=1,offset=0,size=0x2000,rw8=off,rw16,rw32"`

    io [0x8000>>1]uint16
}

func NewWifi() *Wifi {
    wf := &Wifi{}
	wf.rand = rand.New(rand.NewSource(0))
	wf.bbInit()
	return wf
}

func (wf *Wifi) Write16(addr uint32, v uint16) {

    addr &= 0x7FFF

    switch addr {
    case 0x50:
        wf.WRxBufBegin = v

    case 0x52:
        wf.WRxBufEnd = v

    case 0x58:
        wf.WRxBufRdAddr = v & 0x1FFF

    case 0x68:
        wf.WTxBufWrAddr = v & 0x1FFF

    case 0x70:
        wf.WriteWTXBUFWRDATA(v)

    case 0x74:
        wf.WTxBufGapTop = v & 0x1FFF

    case 0x76:
        wf.WTxBufGapDisp = v & 0xFFF

    case 0x158:
        wf.WriteBASEBANDCNT(v)

    case 0x15A:
        wf.BaseBandWrite = v

    case 0x160:
        wf.BaseBandMode = v

    case 0x168:
        wf.BaseBandPower = v

    }

    wf.io[addr>>1] = v
}

func (wf *Wifi) Read16(addr uint32) uint16 {

    addr &= 0x7FFF

    switch addr {
    case 0x44:
        return wf.ReadRANDOM()

    case 0x50:
        return wf.WRxBufBegin

    case 0x52:
        return wf.WRxBufEnd

    case 0x58:
        return wf.WRxBufRdAddr

    case 0x60:
        return wf.ReadWRXBUFRDDATA()
    case 0x68:
        return wf.WTxBufWrAddr

    case 0x74:
        return wf.WTxBufGapTop

    case 0x76:
        return wf.WTxBufGapDisp

    case 0x15C:
        return wf.BaseBandRead

    case 0x15E:
        return wf.BaseBandBusy

    case 0x160:
        return wf.BaseBandMode

    case 0x168:
        return wf.BaseBandPower

    }

    return wf.io[addr>>1]
}

func (wf *Wifi) bbInit() {
	// Initialize baseband registers
	wf.bbRegs[0x00] = 0x6D // Chip ID
	wf.bbRegs[0x5D] = 0x1

	for idx := range wf.bbRegWritable {
		if (idx >= 0x1 && idx <= 0xC) || (idx >= 0x13 && idx <= 0x15) ||
			(idx >= 0x1B && idx <= 0x26) || (idx >= 0x28 && idx <= 0x4C) ||
			(idx >= 0x4E && idx <= 0x5C) || (idx >= 0x62 && idx <= 0x63) ||
			idx == 0x65 || idx == 0x67 || idx == 0x68 {
			wf.bbRegWritable[idx] = true
		}
	}

}

func (wf *Wifi) WriteBASEBANDCNT(val uint16) {

	idx := val & 0xFF
    dir := val >> 12

    const (
        W = 5
        R = 6
    )

    if dir == R {
		wf.BaseBandRead = uint16(wf.bbRegs[idx])
        return
    }

    if dir == W && wf.bbRegWritable[idx] {
        wf.bbRegs[idx] = uint8(wf.BaseBandWrite & 0xFF)
        return
    }
}

func (wf *Wifi) ReadRANDOM() uint16 {
	return uint16(wf.rand.Uint32()) & 0x3FF
}

func (wf *Wifi) WriteWTXBUFWRDATA(val uint16) {
	off := wf.WTxBufWrAddr

    wf.ram[off>>1] = val
	//binary.LittleEndian.PutUint16(wf.ram[off:off+2], val)
	off += 2
	if off == wf.WTxBufGapTop {
		off += wf.WTxBufGapDisp * 2
	}
	off &= 0x1FFF
	wf.WTxBufWrAddr = off
}

func (wf *Wifi) ReadWRXBUFRDDATA() uint16 {
	off := wf.WRxBufRdAddr
    val := wf.ram[off>>1]
	//val := binary.LittleEndian.Uint16(wf.WifiRam[off : off+2])
	off += 2
	if off == wf.WRxBufEnd&0x1FFF {
		off = wf.WRxBufBegin
	}
	off &= 0x1FFF
	wf.WRxBufRdAddr = off
	return val
}
