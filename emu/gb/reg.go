package gameboy

type Registers struct {
    a uint8
    b uint8
    c uint8
    d uint8
    e uint8
    g uint8
    h uint8
    l uint8
    f Flags
}

type Flags struct {
    Zero bool
    Subtraction bool
    HalfCarry bool
    Carry bool
}

const (
    RegisterAf = iota
    RegisterBc
    RegisterDe
    RegisterHl
)

func (r * Registers) setHl(v uint16) {
    r.h = uint8(v & 0xFF00 >> 8)
    r.l = uint8(v & 0xFF)
}

func (r * Registers) setDe(v uint16) {
    r.d = uint8(v & 0xFF00 >> 8)
    r.e = uint8(v & 0xFF)
}

func (r * Registers) setBc(v uint16) {
    r.b = uint8(v & 0xFF00 >> 8)
    r.c = uint8(v & 0xFF)
}

func (r * Registers) setAf(v uint16) {
    r.a = uint8(v & 0xFF00 >> 8)
    r.f.setFlags(uint8(v & 0xFF))
}

func (r *Registers) af() uint16 {
    return uint16(r.a) << 8 | uint16(r.f.getBits())
}

func (r *Registers) bc() uint16 {
    return uint16(r.b) << 8 | uint16(r.c)
}

func (r *Registers) de() uint16 {
    return uint16(r.d) << 8 | uint16(r.e)
}

func (r *Registers) hl() uint16 {
    return uint16(r.h) << 8 | uint16(r.l)
}

func (r *Registers) setCombinedRegister(v uint16, register int) {

    switch register {
    case RegisterAf: r.setAf(v)
    case RegisterBc: r.setBc(v)
    case RegisterDe: r.setDe(v)
    case RegisterHl: r.setHl(v)
    }
}

func (r *Registers) getCombinedRegister(register int) uint16 {
    switch register {
    case RegisterAf: return r.af()
    case RegisterBc: return r.bc()
    case RegisterDe: return r.de()
    case RegisterHl: return r.hl()
    }
    return 0
}

func (fr *Flags) getBits() uint8 {

    var v uint8

    if fr.Zero {
        v += 1 << 7
    }

    if fr.Subtraction {
        v += 1 << 6
    }

    if fr.HalfCarry {
        v += 1 << 5
    }

    if fr.Carry {
        v += 1 << 4
    }

    return v
}


func (fr *Flags) setFlags(bits uint8) {

    var zero bool = false
    var sub bool = false
    var half bool = false
    var carry bool = false

    if (bits >> 7) & 0b1 == 1 {
        zero = true
    }

    if (bits >> 6) & 0b1 == 1 {
        sub = true
    }

    if (bits >> 5) & 0b1 == 1 {
        half = true
    }

    if (bits >> 4) & 0b1 == 1 {
        carry = true
    }

    fr.Zero = zero
    fr.Subtraction = sub
    fr.HalfCarry = half
    fr.Carry = carry
}
