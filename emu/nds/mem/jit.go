package mem

type Jit interface {
	InvalidatePage(addr uint32)
}
