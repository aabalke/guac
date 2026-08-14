package file

import (
	"os"
	"path/filepath"
)

const RW_R__R__ = 0o644

func Read(path string) (rom, sav, rtc *[]byte) {
	switch ext := filepath.Ext(path); ext {
	case ".zip":
	case ".7z":
	case ".gb", ".gba", ".nds":

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

	return rom, sav, rtc
}

func Write(path string, sav *[]byte) {
	switch ext := filepath.Ext(path); ext {
	case ".zip":
	case ".7z":
	case ".gb", ".gba", ".nds":
		if err := os.WriteFile(path+".save", *sav, RW_R__R__); err != nil {
			panic(err)
		}
	}
}

func WriteRtc(path string, rtc *[]byte) {
	switch ext := filepath.Ext(path); ext {
	case ".zip":
	case ".7z":
	case ".gb", ".gba", ".nds":
		if err := os.WriteFile(path+".rtc", *rtc, RW_R__R__); err != nil {
			panic(err)
		}
	}
}
