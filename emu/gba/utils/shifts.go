package utils

const (
	LSL = iota
	LSR
	ASR
	ROR
)

type ShiftArgs struct {
	SType, Val, Is                uint32
	IsCarry, Immediate, CurrCarry bool
}

func Shift(s *ShiftArgs) (shift uint32, setCarry, carry bool) {
	switch s.SType {
	case LSL:
		return Lsl(s.Val, s.Is, s.IsCarry, s.Immediate)
	case LSR:
		return Lsr(s.Val, s.Is, s.IsCarry, s.Immediate)
	case ASR:
		return Asr(s.Val, s.Is, s.IsCarry, s.Immediate)
	case ROR:
		return Ror(s.Val, s.Is, s.IsCarry, s.Immediate, s.CurrCarry)
	default:
		panic("Unknown Shift Type")
	}
}

func Lsl(val, is uint32, isCarry, immediate bool) (shift uint32, setCarry, carry bool) {

	if is == 0 && immediate {
		return val, false, false
	}

	if is > 32 {
		return 0, isCarry, false
	}

	setCarry = is > 0 && isCarry
	carry = val&(1<<(32-is)) > 0
	return val << uint(is), setCarry, carry
}

func Lsr(val, is uint32, isCarry, immediate bool) (shift uint32, setCarry, carry bool) {

	if is == 0 && immediate {
		is = 32
	}

	setCarry = is > 0 && isCarry
	carry = val&(1<<(is-1)) > 0
	return val >> uint(is), setCarry, carry
}

func Asr(val, is uint32, isCarry, immediate bool) (shift uint32, setCarry, carry bool) {

	if (is == 0 && immediate) || is > 32 {
		is = 32
	}

	setCarry = is > 0 && isCarry
	carry = val&(1<<(is-1)) > 0

	msb := val & 0x8000_0000
	for range uint(is) {
		val = (val >> 1) | msb
	}

	return val, setCarry, carry
}

func Ror(val, is uint32, isCarry, immediate bool, currCarry bool) (shift uint32, setCarry, carry bool) {

	getValue := func(v, shift uint32) uint32 {
		shift &= 31
		tmp0 := v >> shift
		tmp1 := v << (32 - shift)
		return tmp0 | tmp1
	}

	if is == 0 && immediate {
		c := uint32(0)
		if currCarry {
			c = 1
		}

		return getValue((val&^1)|c, 1), true, BitEnabled(val, 0)
	}

	carry = (val>>((is-1)&31))&0b1 > 0
	setCarry = is > 0 && isCarry

	return getValue(val, is), setCarry, carry
}
