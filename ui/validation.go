package ui

import (
	"strconv"
	"strings"
)

func NumberValidation(maxValue int) func(string) (bool, *string) {
	return func(s string) (bool, *string) {
		digits := strings.Map(func(r rune) rune {
			if r >= '0' && r <= '9' {
				return r
			}
			return -1
		}, s)

		if digits == "" {
			return false, &digits
		}

		v, _ := strconv.Atoi(digits)

		if v > maxValue {
			digits = strconv.Itoa(maxValue)
			return false, &digits
		}

		return false, &digits
	}
}

func HexValidation(maxValue int) func(string) (bool, *string) {
	return func(s string) (bool, *string) {

		s = strings.ToUpper(s)

		hex := strings.Map(func(r rune) rune {
			switch {
			case r >= '0' && r <= '9':
				return r
			case r >= 'a' && r <= 'f':
				return r
			case r >= 'A' && r <= 'F':
				return r
			default:
				return -1
			}
		}, s)

		if hex == "" {
			return false, &hex
		}

		v, _ := strconv.ParseInt(hex, 16, 64)

		if int(v) > maxValue {
			hex = strconv.FormatInt(int64(maxValue), 16)
			return false, &hex
		}

		return false, &hex
	}
}

func NoValidation() func(string) (bool, *string) {
	return func(s string) (bool, *string) {
		return true, &s
	}
}
