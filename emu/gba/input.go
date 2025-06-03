package gba

import (
	"github.com/veandco/go-sdl2/sdl"
)

func (gba *GBA) InputHandler(event sdl.Event) {

    var tempJoypad uint16 = gba.Joypad
    reqInterrupt := false

    switch e := event.(type) {
    case *sdl.KeyboardEvent:
        gba.UpdateKeyboardInput(e, &tempJoypad, &reqInterrupt)
    case *sdl.ControllerButtonEvent:
        gba.UpdateControllerInput(e, &tempJoypad, &reqInterrupt)
    }

    if reqInterrupt {
        gba.triggerIRQ(12)
    }

    gba.Joypad = tempJoypad

    return
}

func (gba *GBA) UpdateKeyboardInput(keyEvent *sdl.KeyboardEvent, tempJoypad *uint16, reqInterrupt *bool) {

    switch key := keyEvent.Keysym.Sym; key {
    case sdl.K_j: // A
        handleKey(tempJoypad, 0b1, reqInterrupt, keyEvent)
    case sdl.K_k: // B
        handleKey(tempJoypad, 0b10, reqInterrupt, keyEvent)
    case sdl.K_l: // SELECT
        handleKey(tempJoypad, 0b100, reqInterrupt, keyEvent)
    case sdl.K_SEMICOLON: // START
        handleKey(tempJoypad, 0b1000, reqInterrupt, keyEvent)
    case sdl.K_d: //
        handleKey(tempJoypad, 0b1_0000, reqInterrupt, keyEvent)
    case sdl.K_a: //
        handleKey(tempJoypad, 0b10_0000, reqInterrupt, keyEvent)
    case sdl.K_w: //
        handleKey(tempJoypad, 0b100_0000, reqInterrupt, keyEvent)
    case sdl.K_s: //
        handleKey(tempJoypad, 0b1000_0000, reqInterrupt, keyEvent)
    case sdl.K_y: //
        handleKey(tempJoypad, 0b1_0000_0000, reqInterrupt, keyEvent)
    case sdl.K_t: //
        handleKey(tempJoypad, 0b10_0000_0000, reqInterrupt, keyEvent)
    } 

}

func (gba *GBA) UpdateControllerInput(controllerEvent *sdl.ControllerButtonEvent, tempJoypad *uint16, reqInterrupt *bool) {

    switch key := controllerEvent.Button; key {
    case sdl.CONTROLLER_BUTTON_A: // A //ps x
        handleButton(tempJoypad, 0b1, reqInterrupt, controllerEvent)
    case sdl.CONTROLLER_BUTTON_B: // B
        handleButton(tempJoypad, 0b10, reqInterrupt, controllerEvent)
    case sdl.CONTROLLER_BUTTON_X: // SELECT
        handleButton(tempJoypad, 0b100, reqInterrupt, controllerEvent)
    case sdl.CONTROLLER_BUTTON_Y: // START
        handleButton(tempJoypad, 0b1000, reqInterrupt, controllerEvent)
    case sdl.CONTROLLER_BUTTON_DPAD_RIGHT:
        handleButton(tempJoypad, 0b1_0000, reqInterrupt, controllerEvent)
    case sdl.CONTROLLER_BUTTON_DPAD_LEFT:
        handleButton(tempJoypad, 0b10_0000, reqInterrupt, controllerEvent)
    case sdl.CONTROLLER_BUTTON_DPAD_UP:
        handleButton(tempJoypad, 0b100_0000, reqInterrupt, controllerEvent)
    case sdl.CONTROLLER_BUTTON_DPAD_DOWN:
        handleButton(tempJoypad, 0b1000_0000, reqInterrupt, controllerEvent)
    case sdl.CONTROLLER_BUTTON_RIGHTSHOULDER:
        handleButton(tempJoypad, 0b1_0000_0000, reqInterrupt, controllerEvent)
    case sdl.CONTROLLER_BUTTON_LEFTSHOULDER:
        handleButton(tempJoypad, 0b10_0000_0000, reqInterrupt, controllerEvent)
    } 
}

func handleKey(tempJoypad *uint16, mask uint16, reqInterrupt *bool, keyEvent *sdl.KeyboardEvent) {

    if keyEvent.State == sdl.PRESSED {
    //if keyEvent.Type == sdl.KEYDOWN || keyEvent.State == sdl.PRESSED {
        *tempJoypad &^= mask

        if !*reqInterrupt {
            *reqInterrupt = true
        }

        return
    }

    if keyEvent.State == sdl.RELEASED {
    //if keyEvent.Type == sdl.KEYUP || keyEvent.State == sdl.RELEASED{
        *tempJoypad |= mask
    }

}

func handleButton(tempJoypad *uint16, mask uint16, reqInterrupt *bool, controllerEvent *sdl.ControllerButtonEvent) {

    if controllerEvent.Type == sdl.CONTROLLERBUTTONDOWN {

        *tempJoypad &^= mask

        if !*reqInterrupt {
            *reqInterrupt = true
        }
    }

    if controllerEvent.Type == sdl.CONTROLLERBUTTONUP {
        *tempJoypad |= mask
    }
}
