package cart

import (
	"encoding/binary"
	"fmt"
)

type KEY1 struct {
    Bios *[0x4000]byte
    Cart *[]byte

    SecureArea [0x800/4]uint32
    Buf [0x412]uint32
}

func NewKey1(bios *[0x4000]uint8, cart *[]byte) *KEY1 {
    return &KEY1{
        Bios: bios,
        Cart: cart,
    }
}

func (k *KEY1) DecryptCard() {

    k.DecryptSecureArea()

    for i := range uint32(len(k.SecureArea)) {
        v := k.SecureArea[i]
        (*k.Cart)[0x4000+i*4+0] = byte(v>>0)
        (*k.Cart)[0x4000+i*4+1] = byte(v>>8)
        (*k.Cart)[0x4000+i*4+2] = byte(v>>16)
        (*k.Cart)[0x4000+i*4+3] = byte(v>>24)
    }
}

func (k *KEY1) DecryptSecureArea() {

    gamecode := binary.LittleEndian.Uint32((*k.Cart)[0xC:])
    arm9Base := binary.LittleEndian.Uint32((*k.Cart)[0x20:])

    for i := range uint32(len(k.SecureArea)) {
        k.SecureArea[i] = binary.LittleEndian.Uint32((*k.Cart)[arm9Base+i*4:])
    }

    k.InitKeycode(false, gamecode, 2, 2)

    k.SecureArea[0], k.SecureArea[1] = k.Decrypt(k.SecureArea[0], k.SecureArea[1])

    k.InitKeycode(false, gamecode, 3, 2)

    for i := 0; i < len(k.SecureArea); i += 2 {
        k.SecureArea[i], k.SecureArea[i+1] = k.Decrypt(k.SecureArea[i], k.SecureArea[i+1])
    }

    b := make([]byte, 8)
    binary.LittleEndian.PutUint32(b, k.SecureArea[0])
    binary.LittleEndian.PutUint32(b[4:], k.SecureArea[1])

    if decrypted := string(b) == "encryObj"; decrypted {
        fmt.Printf("Decrypted Secure Area\n")
        k.SecureArea[0] = 0xE7FFDEFF
        k.SecureArea[1] = 0xE7FFDEFF
    } else {
        fmt.Printf("Failed to Decrypt Secure Area\n")
        for i := range len(k.SecureArea) {
            k.SecureArea[i] = 0xE7FFDEFF
        }
    }
}

func (k *KEY1) InitKeycode(dsi bool, idcode, level, mod uint32) {

    k.LoadKeyBuf()

    keycode := []uint32{idcode, idcode>>1, idcode<<1}

    if level >= 1 {
        k.ApplyKeycode(keycode, mod)
    }

    if level >= 2 {
        k.ApplyKeycode(keycode, mod)
    }

    if level >= 3 {
        keycode[1] <<= 1
        keycode[2] >>= 1
        k.ApplyKeycode(keycode, mod)
    }
}

func (k *KEY1) LoadKeyBuf() {
    for i := range len(k.Buf) {
        k.Buf[i] = binary.LittleEndian.Uint32((*k.Bios)[0x30+i*4:])
    }
}

func (k *KEY1) ApplyKeycode(keycode []uint32, mod uint32) {

    keycode[1], keycode[2] = k.Encrypt(keycode[1], keycode[2])
    keycode[0], keycode[1] = k.Encrypt(keycode[0], keycode[1])

    for i := uint32(0); i <= 0x11; i++ {
        k.Buf[i] ^= k.ByteSwap(keycode[i % mod])
    }

    temp := [2]uint32{}
    for i := uint32(0); i <= 0x410; i += 2 {
        temp[0], temp[1] = k.Encrypt(temp[0], temp[1])
        k.Buf[i+0] = temp[1]
        k.Buf[i+1] = temp[0]
    }
}

func (k *KEY1) ByteSwap(v uint32) uint32 {
    return (v>>24)|((v>>8)&0xFF00)|((v<<8)&0xFF0000)|(v<<24)
}

func (k *KEY1) Encrypt(y, x uint32) (uint32, uint32) {

    for i, z := uint32(0), uint32(0); i <= 0xF; i++ {
        z =  k.Buf[i] ^ x
        x =  k.Buf[0x012 +  (z >> 24)        ]
        x += k.Buf[0x112 + ((z >> 16) & 0xFF)]
        x ^= k.Buf[0x212 + ((z >>  8) & 0xFF)]
        x += k.Buf[0x312 +  (z        & 0xFF)]
        x ^= y
        y = z
    }

    return x ^ k.Buf[0x10], y ^ k.Buf[0x11]
}

func (k *KEY1) Decrypt(y, x uint32) (uint32, uint32) {

    for i, z := uint32(0x11), uint32(0); i >= 0x2; i-- {
        z =  k.Buf[i] ^ x
        x =  k.Buf[0x012 +  (z >> 24)        ]
        x += k.Buf[0x112 + ((z >> 16) & 0xFF)]
        x ^= k.Buf[0x212 + ((z >>  8) & 0xFF)]
        x += k.Buf[0x312 +  (z        & 0xFF)]
        x ^= y
        y = z
    }

    return x ^ k.Buf[0x1], y ^ k.Buf[0x0]
}
