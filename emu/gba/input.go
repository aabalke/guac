package gba

import "github.com/hajimehoshi/ebiten/v2"

func (gba *GBA) InputHandler(keys []ebiten.Key, buttons []ebiten.GamepadButton) {

    tempJoypad := &gba.Keypad.KEYINPUT

    *tempJoypad = 0b11_1111_1111

    for _, key := range keys {
        switch key {
        case ebiten.KeyJ:
            *tempJoypad &^= 0b1
        case ebiten.KeyK:
            *tempJoypad &^= 0b10
        case ebiten.KeyL:
            *tempJoypad &^= 0b100
        case ebiten.KeySemicolon:
            *tempJoypad &^= 0b1000
        case ebiten.KeyD:
            *tempJoypad &^= 0b10000
        case ebiten.KeyA:
            *tempJoypad &^= 0b100000
        case ebiten.KeyW:
            *tempJoypad &^= 0b1000000
        case ebiten.KeyS:
            *tempJoypad &^= 0b10000000
        case ebiten.KeyY:
            *tempJoypad &^= 0b100000000
        case ebiten.KeyT:
            *tempJoypad &^= 0b1000000000
        }
    }

    for _, button := range buttons {
        switch button {
        case ebiten.GamepadButton2:
            *tempJoypad &^= 0b1
        case ebiten.GamepadButton1:
            *tempJoypad &^= 0b10
        case ebiten.GamepadButton0:
            *tempJoypad &^= 0b100
        case ebiten.GamepadButton3:
            *tempJoypad &^= 0b1000
        case ebiten.GamepadButton16:
            *tempJoypad &^= 0b10000
        case ebiten.GamepadButton18:
            *tempJoypad &^= 0b100000
        case ebiten.GamepadButton15:
            *tempJoypad &^= 0b1000000
        case ebiten.GamepadButton17:
            *tempJoypad &^= 0b10000000


        case ebiten.GamepadButton5:
            *tempJoypad &^= 0b100000000
        case ebiten.GamepadButton4:
            *tempJoypad &^= 0b1000000000
        }
    }

    if gba.Keypad.keyIRQ() {
        gba.Irq.setIRQ(12)
    }
}
//
//func (gba *GBA) UpdateControllerInput(controllerEvent *sdl.ControllerButtonEvent, tempJoypad *uint16) {
//
//    switch key := controllerEvent.Button; key {
//    case sdl.CONTROLLER_BUTTON_A:
//        handleButton(tempJoypad, 0b1, controllerEvent)
//    case sdl.CONTROLLER_BUTTON_B:
//        handleButton(tempJoypad, 0b10, controllerEvent)
//    case sdl.CONTROLLER_BUTTON_X:
//        handleButton(tempJoypad, 0b100, controllerEvent)
//    case sdl.CONTROLLER_BUTTON_Y:
//        handleButton(tempJoypad, 0b1000, controllerEvent)
//    case sdl.CONTROLLER_BUTTON_DPAD_RIGHT:
//        handleButton(tempJoypad, 0b1_0000, controllerEvent)
//    case sdl.CONTROLLER_BUTTON_DPAD_LEFT:
//        handleButton(tempJoypad, 0b10_0000, controllerEvent)
//    case sdl.CONTROLLER_BUTTON_DPAD_UP:
//        handleButton(tempJoypad, 0b100_0000, controllerEvent)
//    case sdl.CONTROLLER_BUTTON_DPAD_DOWN:
//        handleButton(tempJoypad, 0b1000_0000, controllerEvent)
//    case sdl.CONTROLLER_BUTTON_RIGHTSHOULDER:
//        handleButton(tempJoypad, 0b1_0000_0000, controllerEvent)
//    case sdl.CONTROLLER_BUTTON_LEFTSHOULDER:
//        handleButton(tempJoypad, 0b10_0000_0000, controllerEvent)
//    } 
//}
