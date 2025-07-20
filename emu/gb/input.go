package gameboy

import "github.com/hajimehoshi/ebiten/v2"

func (gb *GameBoy) InputHandler(keys []ebiten.Key, buttons []ebiten.GamepadButton) {

    tempJoypad := &gb.Joypad

    reqInterrupt := false

    *tempJoypad = 0b1111_1111

    for _, key := range keys {
        switch key {
        case ebiten.KeyJ:
            *tempJoypad &^= 0b10000
            reqInterrupt = true
        case ebiten.KeyK:
            *tempJoypad &^= 0b100000
            reqInterrupt = true
        case ebiten.KeyL:
            *tempJoypad &^= 0b1000000
            reqInterrupt = true
        case ebiten.KeySemicolon:
            *tempJoypad &^= 0b10000000
            reqInterrupt = true
        case ebiten.KeyD:
            *tempJoypad &^= 0b1
            reqInterrupt = true
        case ebiten.KeyA:
            *tempJoypad &^= 0b10
            reqInterrupt = true
        case ebiten.KeyW:
            *tempJoypad &^= 0b100
            reqInterrupt = true
        case ebiten.KeyS:
            *tempJoypad &^= 0b1000
            reqInterrupt = true
        }
    }

    for _, button := range buttons {
        switch button {
        case ebiten.GamepadButton2:
            *tempJoypad &^= 0b10000
            reqInterrupt = true
        case ebiten.GamepadButton1:
            *tempJoypad &^= 0b100000
            reqInterrupt = true
        case ebiten.GamepadButton0:
            *tempJoypad &^= 0b1000000
            reqInterrupt = true
        case ebiten.GamepadButton3:
            *tempJoypad &^= 0b10000000
            reqInterrupt = true
        case ebiten.GamepadButton16:
            *tempJoypad &^= 0b1
            reqInterrupt = true
        case ebiten.GamepadButton18:
            *tempJoypad &^= 0b10
            reqInterrupt = true
        case ebiten.GamepadButton15:
            *tempJoypad &^= 0b100
            reqInterrupt = true
        case ebiten.GamepadButton17:
            *tempJoypad &^= 0b1000
            reqInterrupt = true
        }
    }

    if reqInterrupt {
        const JOYPAD uint8 = 0b10000
        gb.RequestInterrupt(JOYPAD)
    }
}
