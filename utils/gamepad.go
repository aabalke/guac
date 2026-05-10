package utils

import (
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

func GamepadButtonToString(v ebiten.StandardGamepadButton) string {
	return StdGamepadButtonStrings[v]
}

func StringToGamepadButton(v string) (ebiten.StandardGamepadButton, bool) {
	switch v := strings.ToLower(v); v {
	case "rightbottom":
		return ebiten.StandardGamepadButtonRightBottom, true
	case "rightright":
		return ebiten.StandardGamepadButtonRightRight, true
	case "rightleft":
		return ebiten.StandardGamepadButtonRightLeft, true
	case "righttop":
		return ebiten.StandardGamepadButtonRightTop, true
	case "fronttopleft":
		return ebiten.StandardGamepadButtonFrontTopLeft, true
	case "fronttopright":
		return ebiten.StandardGamepadButtonFrontTopRight, true
	case "frontbottomleft":
		return ebiten.StandardGamepadButtonFrontBottomLeft, true
	case "frontbottomright":
		return ebiten.StandardGamepadButtonFrontBottomRight, true
	case "centerleft":
		return ebiten.StandardGamepadButtonCenterLeft, true
	case "centerright":
		return ebiten.StandardGamepadButtonCenterRight, true
	case "leftstick":
		return ebiten.StandardGamepadButtonLeftStick, true
	case "rightstick":
		return ebiten.StandardGamepadButtonRightStick, true
	case "lefttop":
		return ebiten.StandardGamepadButtonLeftTop, true
	case "leftbottom":
		return ebiten.StandardGamepadButtonLeftBottom, true
	case "leftleft":
		return ebiten.StandardGamepadButtonLeftLeft, true
	case "leftright":
		return ebiten.StandardGamepadButtonLeftRight, true
	case "centercenter":
		return ebiten.StandardGamepadButtonCenterCenter, true
	}

	return -1, false
}

func StringToKey(v string) (ebiten.Key, bool) {
	var k ebiten.Key

	if err := k.UnmarshalText([]byte(v)); err != nil {
		return 0, false
	}

	return k, true
}

func KeyToString(k ebiten.Key) string {
	return k.String()
}

var StdGamepadButtonStrings = []string{
	"RightBottom",
	"RightRight",
	"RightLeft",
	"RightTop",
	"FrontTopLeft",
	"FrontTopRight",
	"FrontBottomLeft",
	"FrontBottomRight",
	"CenterLeft",
	"CenterRight",
	"LeftStick",
	"RightStick",
	"LeftTop",
	"LeftBottom",
	"LeftLeft",
	"LeftRight",
	"CenterCenter",
}

var KeyStrings = []string{
	"A",
	"B",
	"C",
	"D",
	"E",
	"F",
	"G",
	"H",
	"I",
	"J",
	"K",
	"L",
	"M",
	"N",
	"O",
	"P",
	"Q",
	"R",
	"S",
	"T",
	"U",
	"V",
	"W",
	"X",
	"Y",
	"Z",
	"Alt",
	"AltLeft",
	"AltRight",
	"ArrowDown",
	"ArrowLeft",
	"ArrowRight",
	"ArrowUp",
	"Backquote",
	"Backslash",
	"Backspace",
	"BracketLeft",
	"BracketRight",
	"CapsLock",
	"Comma",
	"ContextMenu",
	"Control",
	"ControlLeft",
	"ControlRight",
	"Delete",
	"Digit0",
	"Digit1",
	"Digit2",
	"Digit3",
	"Digit4",
	"Digit5",
	"Digit6",
	"Digit7",
	"Digit8",
	"Digit9",
	"End",
	"Enter",
	"Equal",
	"Escape",
	"F1",
	"F2",
	"F3",
	"F4",
	"F5",
	"F6",
	"F7",
	"F8",
	"F9",
	"F10",
	"F11",
	"F12",
	"F13",
	"F14",
	"F15",
	"F16",
	"F17",
	"F18",
	"F19",
	"F20",
	"F21",
	"F22",
	"F23",
	"F24",
	"Home",
	"Insert",
	"IntlBackslash",
	"Meta",
	"MetaLeft",
	"MetaRight",
	"Minus",
	"NumLock",
	"Numpad0",
	"Numpad1",
	"Numpad2",
	"Numpad3",
	"Numpad4",
	"Numpad5",
	"Numpad6",
	"Numpad7",
	"Numpad8",
	"Numpad9",
	"NumpadAdd",
	"NumpadDecimal",
	"NumpadDivide",
	"NumpadEnter",
	"NumpadEqual",
	"NumpadMultiply",
	"NumpadSubtract",
	"PageDown",
	"PageUp",
	"Pause",
	"Period",
	"PrintScreen",
	"Quote",
	"ScrollLock",
	"Semicolon",
	"Shift",
	"ShiftLeft",
	"ShiftRight",
	"Slash",
	"Space",
	"Tab",
}
