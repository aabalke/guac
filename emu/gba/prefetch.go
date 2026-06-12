package gba

type Prefetch struct {
	Enabled               bool
	Disabled              bool
	Active                bool
	Head, Addr            uint32
	Opcodes               uint8
	Thumb                 bool
	AccessTime, Countdown int64
	Ws                    *Waitstate
}

func NewPrefetch(ws *Waitstate) *Prefetch {
	return &Prefetch{
		Ws:         ws,
		AccessTime: 5,
	}
}

func (p *Prefetch) Cancel() {
	p.Active = false
}

func (p *Prefetch) Step(cycles int64) {
	if !p.Active {
		return
	}

	p.Countdown -= cycles

	size, capacity := uint32(4), uint8(4)
	if p.Thumb {
		size, capacity = 2, 8
	}

	for p.Countdown <= 0 {
		p.Opcodes++
		if !p.Enabled || (p.Opcodes >= capacity) {
			break
		}

		p.Addr += size
		p.Countdown += p.AccessTime
	}
}

func (p *Prefetch) Wait(addr uint32, cycles int64, thumb bool) int64 {
	size := uint32(4)
	if p.Thumb {
		size = 2
	}

	if p.Active {
		if p.Opcodes > 0 && addr == p.Head {
			p.Opcodes--
			p.Head += size
			p.Step(1)
			return 1
		}

		if p.Countdown > 0 && addr == p.Addr {
			cycles = p.Countdown
			p.Step(p.Countdown)
			p.Head = p.Addr
			p.Opcodes = 0
			return cycles
		}
	}

	p.Cancel()

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

	if p.Enabled {
		p.Restart(addr, thumb)
	}

	return cycles
}

func (p *Prefetch) Restart(addr uint32, thumb bool) {
	p.Active = true
	p.Opcodes, p.Thumb = 0, thumb
	if thumb {
		addr &^= 1
		p.AccessTime = int64(p.Ws.S[(addr>>25)&3])
		p.Head, p.Addr = addr+2, addr+2
	} else {
		addr &^= 3
		p.AccessTime = int64(p.Ws.S[(addr>>25)&3]) << 1
		p.Head, p.Addr = addr+4, addr+4
	}
	p.Countdown = p.AccessTime
}
