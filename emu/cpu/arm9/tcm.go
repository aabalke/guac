package arm9

import (
	"encoding/binary"
	"unsafe"
)

type Tcm struct {
	data              []uint8
	base, mask, size  uint32
	enabled, loadMode bool
}

func NewTcm(dataSize uint32) *Tcm {
	return &Tcm{
		data: make([]uint8, dataSize),
		mask: dataSize - 1,
	}
}

func (t *Tcm) Readable(addr, width uint32) bool {
	return t.enabled && !t.loadMode && addr >= t.base && addr+width-1 < t.base+t.size
}

func (t *Tcm) Writeable(addr, width uint32) bool {
	return t.enabled && addr >= t.base && addr+width-1 < t.base+t.size
}

func (t *Tcm) Read8(addr uint32) uint32 {
	return uint32(t.data[(addr-t.base)&t.mask])
}

func (t *Tcm) Read16(addr uint32) uint32 {
	addr &^= 1
	return uint32(binary.LittleEndian.Uint16(t.data[(addr-t.base)&t.mask:]))
}

func (t *Tcm) Read32(addr uint32) uint32 {
	addr &^= 3
	return binary.LittleEndian.Uint32(t.data[(addr-t.base)&t.mask:])
}

func (t *Tcm) ReadPtr(addr uint32) unsafe.Pointer {
	if t.enabled && !t.loadMode && addr >= t.base && addr < t.base+t.size {
		return unsafe.Add(unsafe.Pointer(&t.data[0]), (addr-t.base)&t.mask)
	}

	return nil
}

func (t *Tcm) Write8(addr uint32, v uint8) {
	t.data[(addr-t.base)&t.mask] = v
}

func (t *Tcm) Write16(addr uint32, v uint16) {
	addr &^= 1
	binary.LittleEndian.PutUint16(t.data[(addr-t.base)&t.mask:], v)
}

func (t *Tcm) Write32(addr, v uint32) {
	addr &^= 3
	binary.LittleEndian.PutUint32(t.data[(addr-t.base)&t.mask:], v)
}

func (t *Tcm) WritePtr(addr uint32) unsafe.Pointer {
	if t.enabled && addr >= t.base && addr < t.base+t.size {
		return unsafe.Add(unsafe.Pointer(&t.data[0]), (addr-t.base)&t.mask)
	}

	return nil
}
