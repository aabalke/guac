package mem

import (
	"github.com/aabalke/guac/emu/nds/mem/spi"
)

// this sets the post bios ram usage

const (
    RAM_CART_CHIP_ID_1 = 0x27FF800
    RAM_CART_CHIP_ID_2 = 0x27FF804
    RAM_CART_HDR_CRC   = 0x27FF808
    RAM_CART_SEC_CRC   = 0x27FF80A
    RAM_BOOT_HANDLER   = 0x27FF810
    RAM_SEC_DISABLED   = 0x27FF812
    RAM_NDS7_BIOS_CRC  = 0x27FF850

    RAM_WIFI_USER_SET  = 0x27FF868
    RAM_WIFI_FLASH_P5  = 0x27FF874
    RAM_WIFI_FLASH_P3  = 0x27FF876

    RAM_MESSAGE        = 0x27FF880
    RAM_BOOT_TASK      = 0x27FF884

    RAM_CPY_CART_CHIP_ID_1 = 0x27FFC00
    RAM_CPY_CART_CHIP_ID_2 = 0x27FFC04
    RAM_CPY_CART_HDR_CRC   = 0x27FFC08
    RAM_CPY_CART_SEC_CRC   = 0x27FFC0A
    RAM_CPY_NDS7_BIOS_CRC  = 0x27FFC10
    RAM_CPY_SEC_DISABLED   = 0x27FFC12

    RAM_FRAME_CNT = 0x27FFC3C
    RAM_BOOT_IND  = 0x27FFC40
    RAM_WIFI_USER = 0x27FFC80
)

func setBiosRam(mem *Mem) {

    c := &mem.Cartridge.Rom
    f := &spi.FirmwareData

    h := &mem.Cartridge.Header

    for i := range h.Arm9Size {
        v := mem.Cartridge.Rom[h.Arm9Offset + i]
        mem.Write(h.Arm9RamAddr + i, v, true)
    }

    for i := range h.Arm7Size {
        v := mem.Cartridge.Rom[h.Arm7Offset + i]

        mem.Write(h.Arm7RamAddr + i, v, false)
    }

    //27FF800h 4     NDS Gamecart Chip ID 1
    //27FF804h 4     NDS Gamecart Chip ID 2

    //27FF808h 2     NDS Cart Header CRC (verified)            ;hdr[15Eh]
    mem.Write(RAM_CART_HDR_CRC, c[0x15E], true)
    mem.Write(RAM_CART_HDR_CRC + 1, c[0x15F], true)

    //27FF80Ah 2     NDS Cart Secure Area CRC (not verified ?) ;hdr[06Ch]
    mem.Write(RAM_CART_SEC_CRC, c[0x6C], true)
    mem.Write(RAM_CART_SEC_CRC + 1, c[0x6D], true)

    //27FF810h 2     Boot handler task number (usually FFFFh at cart boot time)
    mem.Write(RAM_BOOT_HANDLER, 0xFF, true)
    mem.Write(RAM_BOOT_HANDLER + 1, 0xFF, true)

    //27FF812h 2     Secure disable (0=Normal, 1=Disable; Cart[078h]=BIOS[1088h])

    //if secDisabled := c[0x78] == b9[0x1088]; secDisabled {
    //    mem.Write(RAM_SEC_DISABLED, 1, true)
    //}

    //27FF850h 2     NDS7 BIOS CRC (5835h)

    //27FF868h 4     Wifi FLASH User Settings FLASH Address (fmw[20h]*8)
    v := uint32((*f)[0x20]) * 8
    mem.Write32(RAM_WIFI_USER_SET, v, true)

    //27FF874h 2     Wifi FLASH firmware part5 crc16 (359Ah) (fmw[026h])
    mem.Write(RAM_WIFI_FLASH_P3, (*f)[0x26], true)

    //27FF876h 2     Wifi FLASH firmware part3/part4 crc16 (fmw[004h] or ZERO) usually zero
    mem.Write(RAM_WIFI_FLASH_P3, 0x0, true)

    //27FF880h 4     Message from NDS9 to NDS7  (=7 at cart boot time)
    mem.Write(RAM_MESSAGE, 7, true)

    //27FF884h 4     NDS7 Boot Task (also checked by NDS9) (=6 at cart boot time)
    mem.Write(RAM_BOOT_TASK, 6, true)

    //27FFC00h 4     NDS Gamecart Chip ID 1   (copy of 27FF800h)
    //27FFC04h 4     NDS Gamecart Chip ID 2   (copy of 27FF804h)
    //27FFC08h 2     NDS Cart Header CRC      (copy of 27FF808h)
    //27FFC0Ah 2     NDS Cart Secure Area CRC (copy of 27FF80Ah)
    //27FFC0Ch 2     NDS Cart Missing/Bad CRC (copy of 27FF80Ch)
    //27FFC0Eh 2     NDS Cart Secure Area Bad (copy of 27FF80Eh)
    //27FFC10h 2     NDS7 BIOS CRC (5835h)    (copy of <27FF850h>)
    //27FFC12h 2     Secure Disable           (copy of 27FF812h)


    //27FFC3Ch 4     Frame Counter (eg. 00000332h in no$gba with original firmware)
    mem.Write32(RAM_FRAME_CNT, 0x332, true)

    //27FFC40h 2     Boot Indicator (1=normal; required for some NDS games, 2=wifi)
    mem.Write(RAM_BOOT_IND, 1, true)

    //27FFC80h 70h   Wifi FLASH User Settings (fmw[newest_user_settings])

    const (
        USER_SETTING_RAM = 0x27FFC80
        USER_SETTING_0   = 0x3FE00
    )

    for i := range uint32(0x100) {
        v := spi.FirmwareData[USER_SETTING_0 + i]
        mem.Write(USER_SETTING_RAM + i, v, true)
    }


    //27FFE00h 170h  NDS Cart Header at 27FFE00h+0..16Fh
    const CART_HEADER_RAM = 0x27FFE00

    for i := range uint32(0x170) {
        v := mem.Cartridge.Rom[i]
        mem.Write(CART_HEADER_RAM + i, v, true)
    }




    // temp adc calibration to match nocash
    //mem.Write(USER_SETTING_RAM + 0x58, 0xDF, true)
    //mem.Write(USER_SETTING_RAM + 0x59, 0x02, true)
    //mem.Write(USER_SETTING_RAM + 0x5A, 0x2C, true)
    //mem.Write(USER_SETTING_RAM + 0x5B, 0x03, true)
    //mem.Write(USER_SETTING_RAM + 0x5C, 0x20, true)
    //mem.Write(USER_SETTING_RAM + 0x5D, 0x20, true)

    //mem.Write(USER_SETTING_RAM + 0x5E, 0x3B, true)
    //mem.Write(USER_SETTING_RAM + 0x5F, 0x0D, true)
    //mem.Write(USER_SETTING_RAM + 0x60, 0xE7, true)
    //mem.Write(USER_SETTING_RAM + 0x61, 0x0C, true)
    //mem.Write(USER_SETTING_RAM + 0x62, 0xE0, true)
    //mem.Write(USER_SETTING_RAM + 0x63, 0xA0, true)
}


//27FFFF8h 2     NDS9 Scratch addr for SWI IsDebugger check
//27FFFFAh 2     NDS7 Scratch addr for SWI IsDebugger check
//27FFFFEh 2     Main Memory Control (on-chip power-down I/O port)
//DTCM+3FF8h 4   NDS9 IRQ IF Check Bits (hardcoded RAM address)
//DTCM+3FFCh 4   NDS9 IRQ Handler (hardcoded RAM address)
//37F8000h FE00h ARM7 bootcode can be loaded here (37F8000h..3807DFFh)
//380FFF8h 4     NDS7 IRQ IF Check Bits (hardcoded RAM address)
//380FFFCh 4     NDS7 IRQ Handler (hardcoded RAM address)
