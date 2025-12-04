package amd64

import (
	"reflect"
	"unsafe"
)

func (a *Assembler) InternalCallFunc(f any) {

	if reflect.TypeOf(f).Kind() != reflect.Func {
		panic("CallFunc: Can't call non-func")
	}

	ival := *(*struct {
		typ uintptr
		fun uintptr
	})(unsafe.Pointer(&f))

	a.MovAbs(uint64(ival.fun), Rdx)
	a.Call(Indirect{Rdx, 0, 64})
}
