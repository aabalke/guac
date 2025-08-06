package cartridge

import (
	"bufio"
	"errors"
	"os"
)

type Mbc interface {
	ReadRom(Cartridge, uint16) uint8
	ReadRam(Cartridge, uint16) uint8
	WriteRam(Cartridge, uint16, uint8)
	Read(Cartridge, uint16) uint8
	Handle(uint16, uint8)
}

func ReadRam(path string) ([]uint8, error) {

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {

			f, err2 := os.Create(path)
			if err2 != nil {
				panic(err2)
			}
			defer f.Close()

			return nil, err
		}

		return nil, err
	}

	defer f.Close()

	stats, err := f.Stat()
	if err != nil {
		return nil, err
	}

	if stats.Size() == 0 {
		return nil, errors.New("Save File has length zero")
	}

	data := make([]uint8, stats.Size())

	reader := bufio.NewReader(f)
	_, err = reader.Read(data)

	return data, nil
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
}
