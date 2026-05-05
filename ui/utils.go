package ui

import (
	"image/color"
	"strconv"
	"strings"

	"github.com/aabalke/guac/utils"
	"github.com/hajimehoshi/ebiten/v2"
)

func toString(value any) string {

	switch v := value.(type) {
	case *int:
		return strconv.Itoa(*v)
	case *string:
		return *v
	case *[]string:
		return join(*v, ", ", func(s string) string { return s })

	case *[]int:
		return join(*v, ", ", strconv.Itoa)

	case *[]ebiten.StandardGamepadButton:
		return join(*v, ", ", utils.GamepadButtonToString)
    
    case *color.Color:
        return utils.ColorToHex(*v)

	default:
		panic("not supported text box input")
	}
}

func fromString(original any, value string) {
	switch v := original.(type) {
	case *int:
		*v, _ = strconv.Atoi(value)
	case *string:
		*v = value
	case *[]string:
		*v = strings.Split(strings.ReplaceAll(value, " ", ""), ",")
	case *[]int:
		a := strings.Split(strings.ReplaceAll(value, " ", ""), ",")
		nums := []int{}

		for _, num := range a {
			n, _ := strconv.Atoi(num)
			nums = append(nums, n)
		}

		*v = nums

	case *[]ebiten.StandardGamepadButton:
		strs := strings.Split(strings.ReplaceAll(value, " ", ""), ",")

		*v = []ebiten.StandardGamepadButton{}
		for i := range strs {
			*v = append(*v, utils.StringToGamepadButton(strs[i]))
		}

	default:
		panic("not supported text box input")
	}
}

func join[T any](vals []T, sep string, f func(T) string) string {
	out := make([]string, len(vals))
	for i, v := range vals {
		out[i] = f(v)
	}
	return strings.Join(out, sep)
}

const MAX_DIALOG_LEN = 24

func trim(s string, max int) string {
	r := []rune(s)

	if len(r) <= max {
		return s
	}

	return "..." + string(r[len(r)-(max-len([]rune("..."))):])
}
