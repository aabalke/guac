package gba

type CycleTiming struct {
    instCycles int
    prevAddr uint32
}

func (c *CycleTiming) popSequential(newAddr uint32, isThumb bool) bool {

    isSequential := false

    switch {
    case !isThumb && (newAddr == c.prevAddr + 8 || newAddr == c.prevAddr + 4):
        isSequential = true

    case isThumb && newAddr == c.prevAddr + 4:
        isSequential = true
    }

    c.prevAddr = newAddr

    return isSequential
}
