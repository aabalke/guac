package cpu

import "unsafe"

type MemoryInterface interface {
	Write8(addr uint32, v uint8, arm9 bool)
	Write16(addr uint32, v uint16, arm9 bool)
	Write32(addr uint32, v uint32, arm9 bool)
	WritePtr(addr uint32, arm9 bool) (unsafe.Pointer, bool)

	Read8(addr uint32, arm9 bool) uint32
	Read16(addr uint32, arm9 bool) uint32
	Read32(addr uint32, arm9 bool) uint32
	ReadPtr(addr uint32, arm9 bool) (unsafe.Pointer, bool)
}
