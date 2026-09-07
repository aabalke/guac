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

func (t *Tcm) Readable(addr uint32) bool {
	return t.enabled && !t.loadMode && addr >= t.base && addr < t.base+t.size
}

func (t *Tcm) Read(addr uint32) (uint32, bool) {
	if t.enabled && !t.loadMode && addr >= t.base && addr+0 < t.base+t.size {
		return uint32(t.data[(addr-t.base)&t.mask]), true
	}

	return 0, false
}

func (t *Tcm) Read16(addr uint32) (uint32, bool) {
	addr &^= 1
	if t.enabled && !t.loadMode && addr >= t.base && addr+1 < t.base+t.size {
		return uint32(binary.LittleEndian.Uint32(t.data[(addr-t.base)&t.mask:])), true
	}

	return 0, false
}

func (t *Tcm) Read32(addr uint32) (uint32, bool) {
	addr &^= 3
	if t.enabled && !t.loadMode && addr >= t.base && addr+3 < t.base+t.size {
		return binary.LittleEndian.Uint32(t.data[(addr-t.base)&t.mask:]), true
	}

	return 0, false
}

func (t *Tcm) ReadPtr(addr uint32) unsafe.Pointer {
	if t.enabled && !t.loadMode && addr >= t.base && addr < t.base+t.size {
		return unsafe.Add(unsafe.Pointer(&t.data[0]), (addr-t.base)&t.mask)
	}

	return nil
}

func (t *Tcm) Write(addr uint32, v uint8) bool {
	if t.enabled && addr >= t.base && addr+0 < t.base+t.size {
		t.data[(addr-t.base)&t.mask] = v
		return true
	}

	return false
}

func (t *Tcm) Write16(addr uint32, v uint16) bool {
	addr &^= 1
	if t.enabled && addr >= t.base && addr+1 < t.base+t.size {
		binary.LittleEndian.PutUint16(t.data[(addr-t.base)&t.mask:], v)
		return true
	}

	return false
}

func (t *Tcm) Write32(addr, v uint32) bool {
	addr &^= 3
	if t.enabled && addr >= t.base && addr+3 < t.base+t.size {
		binary.LittleEndian.PutUint32(t.data[(addr-t.base)&t.mask:], v)
		return true
	}

	return false
}

func (t *Tcm) WritePtr(addr uint32) unsafe.Pointer {
	if t.enabled && addr >= t.base && addr < t.base+t.size {
		return unsafe.Add(unsafe.Pointer(&t.data[0]), (addr-t.base)&t.mask)
	}

	return nil
}
