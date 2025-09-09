package mem


type PostFlg uint8

func (p *PostFlg) Write(v uint8, arm9 bool) {
    if !arm9 {
        return
    }

    *p &^= 0b10
    *p |= PostFlg(v & 0b10)
}

func (p *PostFlg) Read(arm9 bool) uint8 {

    if !arm9 {
        return 0b1
    }

    return uint8(*p) | 0b1
}
