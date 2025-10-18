package utils

const (
	LSL = iota
	LSR
	ASR
	ROR
)

var ShiftFuncs = [4]func(val, is uint32, isCarry, imm, currCarry bool) (shift uint32, setCarry, carry bool) {
    func(val, is uint32, isCarry, imm, currCarry bool) (shift uint32, setCarry, carry bool) {

        // LSL

        if is == 0 && imm {
            return val, false, false
        }

        if is > 32 {
            return 0, isCarry, false
        }

        setCarry = is > 0 && isCarry
        carry = val&(1<<(32-is)) > 0
        return val << is, setCarry, carry
    },
    func(val, is uint32, isCarry, imm, currCarry bool) (shift uint32, setCarry, carry bool) {

        // LSR

        if is == 0 && imm {
            is = 32
        }

        setCarry = is > 0 && isCarry
        carry = val&(1<<(is-1)) > 0
        return val >> is, setCarry, carry
    },
    func(val, is uint32, isCarry, imm, currCarry bool) (shift uint32, setCarry, carry bool) {

        // ASR

        if (is == 0 && imm) || is > 32 {
            is = 32
        }

        setCarry = is > 0 && isCarry
        carry = val&(1<<(is-1)) > 0

        msb := val & 0x8000_0000
        for range is {
            val = (val >> 1) | msb
        }

        return val, setCarry, carry
    },
    func(val, is uint32, isCarry, imm, currCarry bool) (shift uint32, setCarry, carry bool) {

        // ROR

        return Ror(val, is, isCarry, imm, currCarry)
    },
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

		return getValue((val&^1)|c, 1), true, val&1 == 1
	}

	carry = (val>>((is-1)&31))&0b1 > 0
	setCarry = is > 0 && isCarry

	return getValue(val, is), setCarry, carry
}

func RorSimple(v, shift uint32) uint32 {
	shift &= 31
	return (v >> shift) | (v << (32 - shift))
}
