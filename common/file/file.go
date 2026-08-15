package file

import (
	"fmt"
	"os"
	"path/filepath"
)

const RW_R__R__ = 0o644

func Read(path string) (name string, rom, sav, rtc *[]byte) {
	switch ext := filepath.Ext(path); ext {
	case ".zip":
		name, rom, sav, rtc = ReadZip(path)

	case ".7z":
	case ".gb", ".gbc", ".gba", ".nds":

		name = filepath.Base(path)

		if data, err := os.ReadFile(path); err == nil {
			rom = &data
		}

		if data, err := os.ReadFile(path + ".save"); err == nil {
			sav = &data
		}

		if data, err := os.ReadFile(path + ".rtc"); err == nil {
			rtc = &data
		}
	}

	return name, rom, sav, rtc
}

func Write(path, name string, sav *[]byte) {
	switch ext := filepath.Ext(path); ext {
	case ".zip":
		WriteZip(path, name, sav)

	case ".7z":
	case ".gb", ".gbc", ".gba", ".nds":
		if err := os.WriteFile(path+".save", *sav, RW_R__R__); err != nil {
			panic(err)
		}
	}

	fmt.Printf("Saved\n")
}

func WriteRtc(path, name string, rtc *[]byte) {
	switch ext := filepath.Ext(path); ext {
	case ".zip":
		WriteRtcZip(path, name, rtc)

	case ".7z":
	case ".gb", ".gbc", ".gba", ".nds":
		if err := os.WriteFile(path+".rtc", *rtc, RW_R__R__); err != nil {
			panic(err)
		}
	}
}
