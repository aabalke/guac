package gba

import "fmt"

func printer(args map[string]any) {

    out := ""

    for k, v := range args {
        out += fmt.Sprintf("%s ", k)

        switch val := v.(type) {
        case int: out += fmt.Sprintf("%X", val)
        case int8: out += fmt.Sprintf("%X", val)
        case int16: out += fmt.Sprintf("%X", val)
        case int32: out += fmt.Sprintf("%X", val)
        case int64: out += fmt.Sprintf("%X", val)
        case uint: out += fmt.Sprintf("%X", val)
        case uint8: out += fmt.Sprintf("%X", val)
        case uint16: out += fmt.Sprintf("%X", val)
        case uint32: out += fmt.Sprintf("%X", val)
        case uint64: out += fmt.Sprintf("%X", val)
        case string: out += fmt.Sprintf("%s", val)
        case bool: out += fmt.Sprintf("%t", val)
        }

        out += " "
    }


    out += "\n"

    fmt.Print(out)
}
