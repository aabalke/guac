package ui

import (
	"strconv"
	"strings"

	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/utils"
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

func KeyValidation() func(string) (bool, *string) {
	isValidPrefix := func(token string) bool {
		token = strings.ToLower(token)
		for _, a := range utils.KeyStrings {
			if strings.HasPrefix(strings.ToLower(a), token) {
				return true
			}
		}
		return false
	}

	return func(s string) (bool, *string) {
		out := strings.Map(func(r rune) rune {
			switch {
			case r >= 'A' && r <= 'Z':
				return r
			case r >= 'a' && r <= 'z':
				return r
			case r >= '0' && r <= '9':
				return r
			case r == ',', r == ' ':
				return r
			default:
				return -1
			}
		}, s)

		previous := ""
		r := []rune(s)
		if len(r) > 0 {
			previous = string(r[:len(r)-1])
		}

		// Second pass: only allow spaces after commas
		var b strings.Builder
		b.Grow(len(out))

		var prev rune
		for _, r := range out {
			if r == ' ' && prev != ',' {
				continue
			}
			if r == ',' && (prev == 0 || prev == ' ' || prev == ',') {
				continue
			}
			b.WriteRune(r)
			prev = r
		}

		out = b.String()

		// might be best to just check last bind not all

		for bind := range strings.SplitSeq(strings.ReplaceAll(out, " ", ""), ",") {
			if !isValidPrefix(bind) {
				return false, &previous
			}
		}

		previous = out

		return false, &out
	}
}

func ControllerValidation() func(string) (bool, *string) {
	isValidPrefix := func(token string) bool {
		token = strings.ToLower(token)
		for _, a := range utils.StdGamepadButtonStrings {
			if strings.HasPrefix(strings.ToLower(a), token) {
				return true
			}
		}
		return false
	}

	return func(s string) (bool, *string) {
		previous := ""
		if len(s)-1 > 0 {
			previous = s[:len(s)-1]
		}

		out := strings.Map(func(r rune) rune {
			switch r {
			case 'R', 'I', 'G', 'H', 'T', 'B', 'O', 'M', 'L', 'E', 'F', 'P', 'N', 'C', 'S', 'K':
				return r
			case 'r', 'i', 'g', 'h', 't', 'b', 'o', 'm', 'l', 'e', 'f', 'p', 'n', 'c', 's', 'k':
				return r
			case ',', ' ':
				return r
			default:
				return -1
			}
		}, s)

		// Second pass: only allow spaces after commas
		var b strings.Builder
		b.Grow(len(out))

		var prev rune
		for _, r := range out {
			if r == ' ' && prev != ',' {
				continue
			}
			if r == ',' && (prev == 0 || prev == ' ' || prev == ',') {
				continue
			}
			b.WriteRune(r)
			prev = r
		}

		out = b.String()

		// might be best to just check last

		for bind := range strings.SplitSeq(strings.ReplaceAll(out, " ", ""), ",") {
			if !isValidPrefix(bind) {
				return false, &previous
			}
		}

		previous = out

		return false, &out
	}
}

func StringValidation(maxLength int) func(string) (bool, *string) {
	return func(s string) (bool, *string) {
		if len(s) >= maxLength {
			s = s[:maxLength]
		}

		return false, &s
	}
}

func ColorNdsValidation() func(string) (bool, *string) {
	isValidPrefix := func(token string) bool {
		token = strings.ToLower(token)
		for _, a := range config.ColorNames {
			if strings.HasPrefix(strings.ToLower(a), token) {
				return true
			}
		}
		return false
	}

	return func(s string) (bool, *string) {
		previous := ""
		if len(s)-1 > 0 {
			previous = s[:len(s)-1]
		}

		out := strings.Map(func(r rune) rune {
			switch {
			case r >= 'A' && r <= 'Z':
				return r
			case r >= 'a' && r <= 'z':
				return r
			case r == ' ':
				return r
			default:
				return -1
			}
		}, s)

		// Second pass: only allow spaces after commas
		var b strings.Builder
		b.Grow(len(out))

		var prev rune
		for _, r := range out {
			if r == ' ' && prev == ' ' {
				continue
			}
			b.WriteRune(r)
			prev = r
		}

		out = b.String()

		if !isValidPrefix(out) {
			return false, &previous
		}

		previous = out

		return false, &out
	}
}
