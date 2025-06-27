package gba

import "fmt"

func assert(b bool, s string) {
    if !b {
        panic(fmt.Sprintf("ASSERT FAIL: %s", s))
    }
}
