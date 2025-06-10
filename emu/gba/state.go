package gba

import (
	"encoding/gob"
	"fmt"
	"os"
)

func LoadState(gba *GBA, path string) {

    breakState(gba)

	f, err := os.Open(path)
    if err != nil {
        panic(err)
    }

    err = gob.NewDecoder(f).Decode(gba)
    if err != nil {
        panic(err)
    }

    f.Close()

    returnState(gba)

    fmt.Printf("State File Decoded\n")
    gba.GBA_LOCK = false

}

func SaveState(gba *GBA, path string) {

    breakState(gba)

    err := os.Rename(path, path + ".backup")
    if err != nil {
        fmt.Printf("There was no original State File to backup\n")
    }

	f, err := os.Create(path)
    if err != nil {
        panic(err)
    }

    err = gob.NewEncoder(f).Encode(gba)
    if err != nil {
        panic(err)
    }

    f.Close()

    returnState(gba)

    fmt.Printf("State File Encoded\n")

}

func breakState(gba *GBA) {
    gba.Cartridge.Header.Cartridge = nil
    gba.Cpu.Gba = nil
    gba.Cpu.Reg.Cpu = nil
    gba.Dma[0].Gba = nil
    gba.Dma[1].Gba = nil
    gba.Dma[2].Gba = nil
    gba.Dma[3].Gba = nil
    gba.Timers[0].Gba = nil
    gba.Timers[1].Gba = nil
    gba.Timers[2].Gba = nil
    gba.Timers[3].Gba = nil
    gba.Mem.GBA = nil
    gba.Debugger.Gba = nil
    gba.InterruptStack.Gba = nil
}

func returnState(gba *GBA) {
    gba.Cartridge.Header.Cartridge = gba.Cartridge
    gba.Cpu.Gba = gba
    gba.Cpu.Reg.Cpu = gba.Cpu
    gba.Dma[0].Gba = gba
    gba.Dma[1].Gba = gba
    gba.Dma[2].Gba = gba
    gba.Dma[3].Gba = gba
    gba.Timers[0].Gba = gba
    gba.Timers[1].Gba = gba
    gba.Timers[2].Gba = gba
    gba.Timers[3].Gba = gba
    gba.Mem.GBA = gba
    gba.Debugger.Gba = gba
    gba.InterruptStack.Gba = gba
}
