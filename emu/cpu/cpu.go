package cpu

import "unsafe"

type MemoryInterface interface {
	Write8(addr uint32, v uint8)
	Write16(addr uint32, v uint16)
	Write32(addr uint32, v uint32)
	WritePtr(addr uint32) (unsafe.Pointer, bool)

	Read8(addr uint32) uint32
	Read16(addr uint32) uint32
	Read32(addr uint32) uint32
	ReadPtr(addr uint32) (unsafe.Pointer, bool)
}
