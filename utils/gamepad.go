package utils

import (
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

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

	v = strings.ToLower(v)

	switch v {
	case "rightbottom":
		return ebiten.StandardGamepadButtonRightBottom
	case "rightright":
		return ebiten.StandardGamepadButtonRightRight
	case "rightleft":
		return ebiten.StandardGamepadButtonRightLeft
	case "righttop":
		return ebiten.StandardGamepadButtonRightTop
	case "fronttopleft":
		return ebiten.StandardGamepadButtonFrontTopLeft
	case "fronttopright":
		return ebiten.StandardGamepadButtonFrontTopRight
	case "frontbottomleft":
		return ebiten.StandardGamepadButtonFrontBottomLeft
	case "frontbottomright":
		return ebiten.StandardGamepadButtonFrontBottomRight
	case "centerleft":
		return ebiten.StandardGamepadButtonCenterLeft
	case "centerright":
		return ebiten.StandardGamepadButtonCenterRight
	case "leftstick":
		return ebiten.StandardGamepadButtonLeftStick
	case "rightstick":
		return ebiten.StandardGamepadButtonRightStick
	case "lefttop":
		return ebiten.StandardGamepadButtonLeftTop
	case "leftbottom":
		return ebiten.StandardGamepadButtonLeftBottom
	case "leftleft":
		return ebiten.StandardGamepadButtonLeftLeft
	case "leftright":
		return ebiten.StandardGamepadButtonLeftRight
	case "centercenter":
		return ebiten.StandardGamepadButtonCenterCenter
	}

	return -1
}
