package gba

type Prefetch struct {
	Tick                  func(cycles int)
	Ws                    *Waitstate
	AccessTime, Countdown int64
	Head, Addr            uint32
	Width                 uint32
	Capacity              uint32
	Opcodes               uint32
	Enabled               bool
	Disabled              bool
	Active                bool
}

func NewPrefetch(ws *Waitstate, Tick func(int)) *Prefetch {
	return &Prefetch{
		Ws:         ws,
		AccessTime: 5,
		Tick:       Tick,
	}
}

func (p *Prefetch) Cancel(r15 uint32) {
	if !p.Active {
		return
	}

	p.Active = false

	if r15 < 0x800_0000 || r15 >= 0xE00_0000 {
		return
	}

	halfPlusOne := (p.AccessTime >> 1) + 1
	if p.Countdown == 1 || (p.Width == 4 && p.Countdown == halfPlusOne) {
		p.Tick(1)
	}
}

func (p *Prefetch) Step(cycles int64) {
	p.Countdown -= cycles

	if !p.Enabled {
		p.Opcodes++
		return
	}

	for p.Countdown <= 0 {
		if p.Opcodes >= p.Capacity {
			p.Opcodes++
			break
		}

		p.Opcodes++
		p.Addr += p.Width
		p.Countdown += p.AccessTime
	}
}

func (p *Prefetch) Wait(r15, addr, width uint32, cycles int64, code bool) {
	if !code {
		p.Cancel(r15)
		p.Tick(int(cycles))
		return
	}

	// width 2 = cap 8, width 4 = cap 4
	p.Capacity = (1 << ((width & 3) >> 1)) << 2

	if p.Active {
		if p.Opcodes != 0 && addr == p.Head {
			p.Opcodes--
			p.Head += width
			p.Tick(1)
			return
		}

		if p.Countdown > 0 && addr == p.Addr {
			p.Tick(int(p.Countdown))
			p.Head = p.Addr
			p.Opcodes = 0
			return
		}
	}

	p.Cancel(r15)

	if p.Disabled {
		p.Disabled = false

		region := (addr >> 25) & 3

		if region == 3 {
			panic("sram")
		}

		switch cycles {
		case int64(p.Ws.S[region]):
			cycles = int64(p.Ws.N[0])
		case int64(p.Ws.S[region] << 1):
			cycles = int64(p.Ws.N[0]) + int64(p.Ws.S[0])
		}

	}

	p.Tick(int(cycles))

	if !p.Enabled {
		return
	}

	p.Active = true
	p.Opcodes = 0
	p.Width = width
	p.AccessTime = int64(p.Ws.S[(addr>>25)&3]) << (width >> 2) // * 2 if 32bit addr
	p.Countdown = p.AccessTime
	p.Addr = addr + width
	p.Head = p.Addr
}
