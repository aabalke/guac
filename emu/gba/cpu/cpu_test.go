package cpu

import (
	"testing"
)

func TestCond(t *testing.T) {

    tests := []struct{
        original uint32
        flag uint32
        value bool
        result uint32
    }{
        {0b1111_1111, 5, true, 0b1111_1111},
        {0b1100_1111, 5, true, 0b1110_1111},
        {0b1101_0111, 5, true, 0b1111_0111},
        {0b0010_0000, 5, false, 0b0},
    }

    for _, tt := range tests {

        c := Cond(tt.original)

        c.SetFlag(tt.flag, tt.value)

        if uint32(c) != tt.result {
            t.Errorf("GetByte Failed. Expected %b, got=%b", tt.result, uint32(c))
        }
    }
}
