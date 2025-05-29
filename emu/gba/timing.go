package gba

import "fmt"

type CycleTiming struct {
    instCycles int
    prevAddr uint32
}

func (c *CycleTiming) popSequential(newAddr uint32, isThumb bool) bool {

    fmt.Printf("PREV %08X, NEW %08X SEQ %t\n", c.prevAddr, newAddr, !isThumb && (newAddr == c.prevAddr + 4))

    isSequential := false

    switch {
    case !isThumb && (newAddr == c.prevAddr + 4):
        isSequential = true
    //case !isThumb && (newAddr == c.prevAddr + 8 || newAddr == c.prevAddr + 4):
    //    isSequential = true

    case isThumb && newAddr == c.prevAddr + 2:
        isSequential = true
    }

    return isSequential
}
