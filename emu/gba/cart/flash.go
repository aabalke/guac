package cart

const (
    FL_READ = 0
    FL_ID = 1
    FL_ERASE_ALL = 2
    FL_ERASE = 3
    FL_WRITE = 4
    FL_BANKSWITCH = 5
    FL_ERASE_SECTOR = 6
)

func (c *Cartridge) ReadFlash(addr uint32) uint8 {

    if c.FlashMode == FL_ID {
        switch addr {
        case 0: return uint8(c.Manufacturer)
        case 1: return uint8(c.Device)
        default: return 0xFF
        }
    }

    bankAddr := (c.FlashBank * 0x1_0000) + addr

    return c.Flash[bankAddr]
}

func (c *Cartridge) WriteFlash(addr uint32, v uint8) {

    if c.SetMode(addr, v) {
        return
    }

    switch c.FlashMode {
    case FL_ID, FL_READ:
        return

    case FL_WRITE:
        bankAddr := (c.FlashBank * 0x1_0000) + addr
        if c.Flash[bankAddr] == 0xFF {
            c.Flash[bankAddr] = v
        }

    case FL_ERASE:
        c.EraseFlash(addr, v)

    case FL_BANKSWITCH:
        if addr == 0x0000 && c.FlashType == FLASH128 && (v == 0 || v == 1) {
            c.FlashBank = uint32(v)
        }
    }

    c.FlashMode = FL_READ

}

func (c *Cartridge) SetMode(addr uint32, v uint8) bool {

    switch c.FlashStage {
    case 0: 
        if addr == 0x5555 && v == 0xAA {
            c.FlashStage = 1
            return true
        }
    case 1:
        if addr == 0x2AAA && v == 0x55 {
            c.FlashStage = 2
            return true
        }
    case 2:

        if c.FlashMode == FL_ERASE && v == 0x10 {
            c.FlashMode = FL_ERASE_ALL
            c.FlashStage = 0
            return true
        }

        if c.FlashMode == FL_ERASE && v == 0x30 {
            c.FlashMode = FL_ERASE_SECTOR
            c.FlashStage = 0
            return true
        }

        switch v {
        case 0x90:
            c.FlashMode = FL_ID
        case 0xF0:
            c.FlashMode = FL_READ
        case 0xA0:
            c.FlashMode = FL_WRITE
        case 0x80:
            c.FlashMode = FL_ERASE
        case 0xB0:
            c.FlashMode = FL_BANKSWITCH
        }
        c.FlashStage = 0
        return true

    default: c.FlashStage = 0
    }

    return false
}

func (c *Cartridge) EraseFlash(addr uint32, v uint8) {

    if c.FlashMode == FL_ERASE_ALL {
        for i := range len(c.Flash) {
            c.Flash[i] = 0xFF
        }

        c.FlashMode = FL_READ
        c.FlashStage = 0
        return

    }
 
    if c.FlashMode == FL_ERASE_SECTOR {

        if addr % 0x1000 != 0 {
            return
        }

        i := uint32(0)
        for i = range 0x1000 {
            c.Flash[(c.FlashBank * 0x1_0000) + addr + i] = 0xFF
        }

        c.FlashMode = FL_READ
        c.FlashStage = 0
    }
}
