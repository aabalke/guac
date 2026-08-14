package cart

import (
	"bufio"
	"os"
	"unsafe"
)

type Mbc interface {
	ReadPtr(uint16) unsafe.Pointer
	Read(uint16) uint8
	Write(uint16, uint8)
	Save()
}

func WriteRam(path string, data []uint8) {
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	writer := bufio.NewWriter(f)
	_, err = writer.Write(data)
	if err != nil {
		panic(err)
	}

	writer.Flush()
}
