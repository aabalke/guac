package rast

//import "fmt"

type Rasterizer struct {
}

type GXSTAT uint32

func (g *GXSTAT) Write(v, b uint8) {

	(*g) |= 0x0600_0000

	if b == 3 {
		(*g) &^= 0xFF << 24
		(*g) |= GXSTAT((uint32(v) & 0b1100_0000) << 24)
	}

	//fmt.Printf("WRITING TO GX STAT %08X\n", *g)
}

func (g *GXSTAT) Read(b uint32) uint8 {

	(*g) |= 0x0600_0000

	//fmt.Printf("READING TO GX STAT %08X\n", *g)
	return uint8(uint32(*g) >> (8 * b))
}
