package gba

import (
	"github.com/veandco/go-sdl2/sdl"
)

func (gba *GBA) InputHandler(event sdl.Event) {

    tempJoypad := &gba.Keypad.KEYINPUT

    switch e := event.(type) {
    case *sdl.KeyboardEvent:
        gba.UpdateKeyboardInput(e, tempJoypad)
    case *sdl.ControllerButtonEvent:
        gba.UpdateControllerInput(e, tempJoypad)
    }

    //gba.Keypad.KEYINPUT = tempJoypad

    if gba.Keypad.keyIRQ() {
        gba.Irq.setIRQ(12)
    }

    return
}

func (gba *GBA) UpdateKeyboardInput(keyEvent *sdl.KeyboardEvent, tempJoypad *uint16) {

    switch key := keyEvent.Keysym.Sym; key {
    case sdl.K_j: // A
        handleKey(tempJoypad, 0b1, keyEvent)
    case sdl.K_k: // B
        handleKey(tempJoypad, 0b10, keyEvent)
    case sdl.K_l: // SELECT
        handleKey(tempJoypad, 0b100, keyEvent)
    case sdl.K_SEMICOLON: // START
        handleKey(tempJoypad, 0b1000, keyEvent)
    case sdl.K_d: //
        handleKey(tempJoypad, 0b1_0000, keyEvent)
    case sdl.K_a: //
        handleKey(tempJoypad, 0b10_0000, keyEvent)
    case sdl.K_w: //
        handleKey(tempJoypad, 0b100_0000, keyEvent)
    case sdl.K_s: //
        handleKey(tempJoypad, 0b1000_0000, keyEvent)
    case sdl.K_y: //
        handleKey(tempJoypad, 0b1_0000_0000, keyEvent)
    case sdl.K_t: //
        handleKey(tempJoypad, 0b10_0000_0000, keyEvent)
    } 
}

func (gba *GBA) UpdateControllerInput(controllerEvent *sdl.ControllerButtonEvent, tempJoypad *uint16) {

    switch key := controllerEvent.Button; key {
    case sdl.CONTROLLER_BUTTON_A:
        handleButton(tempJoypad, 0b1, controllerEvent)
    case sdl.CONTROLLER_BUTTON_B:
        handleButton(tempJoypad, 0b10, controllerEvent)
    case sdl.CONTROLLER_BUTTON_X:
        handleButton(tempJoypad, 0b100, controllerEvent)
    case sdl.CONTROLLER_BUTTON_Y:
        handleButton(tempJoypad, 0b1000, controllerEvent)
    case sdl.CONTROLLER_BUTTON_DPAD_RIGHT:
        handleButton(tempJoypad, 0b1_0000, controllerEvent)
    case sdl.CONTROLLER_BUTTON_DPAD_LEFT:
        handleButton(tempJoypad, 0b10_0000, controllerEvent)
    case sdl.CONTROLLER_BUTTON_DPAD_UP:
        handleButton(tempJoypad, 0b100_0000, controllerEvent)
    case sdl.CONTROLLER_BUTTON_DPAD_DOWN:
        handleButton(tempJoypad, 0b1000_0000, controllerEvent)
    case sdl.CONTROLLER_BUTTON_RIGHTSHOULDER:
        handleButton(tempJoypad, 0b1_0000_0000, controllerEvent)
    case sdl.CONTROLLER_BUTTON_LEFTSHOULDER:
        handleButton(tempJoypad, 0b10_0000_0000, controllerEvent)
    } 
}

func handleKey(tempJoypad *uint16, mask uint16, keyEvent *sdl.KeyboardEvent) {

    if keyEvent.State == sdl.PRESSED {
    //if keyEvent.Type == sdl.KEYDOWN || keyEvent.State == sdl.PRESSED {
        *tempJoypad &^= mask

        return
    }

    if keyEvent.State == sdl.RELEASED {
    //if keyEvent.Type == sdl.KEYUP || keyEvent.State == sdl.RELEASED{
        *tempJoypad |= mask
    }

}

func handleButton(tempJoypad *uint16, mask uint16, controllerEvent *sdl.ControllerButtonEvent) {

    if controllerEvent.Type == sdl.CONTROLLERBUTTONDOWN {
        *tempJoypad &^= mask
        return
    }

    if controllerEvent.Type == sdl.CONTROLLERBUTTONUP {
        *tempJoypad |= mask
    }
}
