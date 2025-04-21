package utils

import (
	"testing"
)

func TestBitEnabled(t *testing.T) {

    tests := []struct{
        value uint32
        bit uint8
        result bool
    }{
        {0b0000_0001_1110, 0, false},
        {0b0000_1001_1110, 4, true},
        {0b0000_1001_1110, 7, true},
        {0b1_0000_1001_1110, 11, false},
    }

    for _, tt := range tests {

        out := bitEnabled(tt.value, tt.bit)

        if out != tt.result {
            t.Errorf("BitEnabled Failed. Expected %t, got=%t", tt.result, out)
        }
    }
}
