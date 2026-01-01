package mem

// souce rasky/ndsemu

// Key2 is a simple stream cipher that uses 2 39-bit LSFRs to generate
// the PRNG to encrypt the ciphertext. The LSFRs have the following polynomials:
// 	 L1 = x^5+x^17+x^18+x^31
// 	 L2 = x^5+x^18+x^23+x^31

type Key2 struct {
	x, y uint64
}

func NewDefaultKey2() Key2 {
	return NewKey2(0x58_C56D_E0E8, 0x5C_879B_9B05)
}

func NewKey2(x, y uint64) Key2 {
	return Key2{
		x: br39(x),
		y: br39(y),
	}
}

func (k *Key2) Encrypt(output, input []uint8) {
	for idx, v := range input {
		x := uint8((k.x >> 5) ^ (k.x >> 17) ^ (k.x >> 18) ^ (k.x >> 31))
		y := uint8((k.y >> 5) ^ (k.y >> 23) ^ (k.y >> 18) ^ (k.y >> 31))
		output[idx] = v ^ x ^ y
		k.x = k.x<<8 | uint64(x)
		k.y = k.y<<8 | uint64(y)
	}
}

func br39(val uint64) uint64 {
	ret := uint64(0)
	for i := range 39 {
		ret |= (val & 1) << uint(38-i)
		val >>= 1
	}
	return ret
}
