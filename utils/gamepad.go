package utils

import "github.com/hajimehoshi/ebiten/v2"

func GamepadButtonToString(v ebiten.StandardGamepadButton) string {
	switch v {
	case ebiten.StandardGamepadButtonRightBottom:
		return "RightBottom"
	case ebiten.StandardGamepadButtonRightRight:
		return "RightRight"
	case ebiten.StandardGamepadButtonRightLeft:
		return "RightLeft"
	case ebiten.StandardGamepadButtonRightTop:
		return "RightTop"
	case ebiten.StandardGamepadButtonFrontTopLeft:
		return "FrontTopLeft"
	case ebiten.StandardGamepadButtonFrontTopRight:
		return "FrontTopRight"
	case ebiten.StandardGamepadButtonFrontBottomLeft:
		return "FrontBottomLeft"
	case ebiten.StandardGamepadButtonFrontBottomRight:
		return "FrontBottomRight"
	case ebiten.StandardGamepadButtonCenterLeft:
		return "CenterLeft"
	case ebiten.StandardGamepadButtonCenterRight:
		return "CenterRight"
	case ebiten.StandardGamepadButtonLeftStick:
		return "LeftStick"
	case ebiten.StandardGamepadButtonRightStick:
		return "RightStick"
	case ebiten.StandardGamepadButtonLeftTop:
		return "LeftTop"
	case ebiten.StandardGamepadButtonLeftBottom:
		return "LeftBottom"
	case ebiten.StandardGamepadButtonLeftLeft:
		return "LeftLeft"
	case ebiten.StandardGamepadButtonLeftRight:
		return "LeftRight"
	case ebiten.StandardGamepadButtonCenterCenter:
		return "CenterCenter"
	}

	return ""
}

func StringToGamepadButton(v string) ebiten.StandardGamepadButton {
	switch v {
	case "RightBottom":
		return ebiten.StandardGamepadButtonRightBottom
	case "RightRight":
		return ebiten.StandardGamepadButtonRightRight
	case "RightLeft":
		return ebiten.StandardGamepadButtonRightLeft
	case "RightTop":
		return ebiten.StandardGamepadButtonRightTop
	case "FrontTopLeft":
		return ebiten.StandardGamepadButtonFrontTopLeft
	case "FrontTopRight":
		return ebiten.StandardGamepadButtonFrontTopRight
	case "FrontBottomLeft":
		return ebiten.StandardGamepadButtonFrontBottomLeft
	case "FrontBottomRight":
		return ebiten.StandardGamepadButtonFrontBottomRight
	case "CenterLeft":
		return ebiten.StandardGamepadButtonCenterLeft
	case "CenterRight":
		return ebiten.StandardGamepadButtonCenterRight
	case "LeftStick":
		return ebiten.StandardGamepadButtonLeftStick
	case "RightStick":
		return ebiten.StandardGamepadButtonRightStick
	case "LeftTop":
		return ebiten.StandardGamepadButtonLeftTop
	case "LeftBottom":
		return ebiten.StandardGamepadButtonLeftBottom
	case "LeftLeft":
		return ebiten.StandardGamepadButtonLeftLeft
	case "LeftRight":
		return ebiten.StandardGamepadButtonLeftRight
	case "CenterCenter":
		return ebiten.StandardGamepadButtonCenterCenter
	}

	return -1
}
