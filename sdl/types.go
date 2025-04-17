package sdl

// I32 is a type allowing for int32 | *int32

type I32 any

func GetI32(u I32) int32 {
	switch v := u.(type) {
	case int32:
		return v
	case *int32:
		return *v
	case int:
		panic("invalid I32Union int")
	case float32:
		panic("invalid I32Union float32")
	case float64:
		panic("invalid I32Union float64")
	default:
		panic("invalid I32Union")
	}
}

func SetI32(u *I32, val any) {
	switch v := val.(type) {
	case int:
		*u = int32(v)
	case float64:
		*u = int32(v)
	case float32:
		*u = int32(v)
	case int32:
		*u = v
	case *int32:
		*u = v
	default:
		panic("SetI32 only works on int32 | *int32")
	}
}
