package arm9

const (
	VEC_RESET         = 0xFFFF_0000
	VEC_UND           = 0xFFFF_0004
	VEC_SWI           = 0xFFFF_0008
	VEC_PREFETCHABORT = 0xFFFF_000C
	VEC_DATAABORT     = 0xFFFF_0010
	VEC_ADDR26BIT     = 0xFFFF_0014
	VEC_IRQ           = 0xFFFF_0018
	VEC_FIQ           = 0xFFFF_001C
)

func (cpu *Cpu) Exception(addr uint32, mode uint32) {

	var (
		cpsr = &cpu.Reg.CPSR
		reg  = &cpu.Reg
		r    = &cpu.Reg.R
	)

	if mode == cpsr.Mode {
		return
	}

	thumb := cpsr.T
	c := BANK_ID[cpsr.Mode]
	i := BANK_ID[mode]
	reg.SP[c] = r[SP]
	reg.LR[c] = r[LR]
	r[SP] = reg.SP[i]
	r[LR] = reg.LR[i]
	reg.SPSR[i] = *cpsr

	if thumb && (mode == MODE_SWI || mode == MODE_ABT) {
		r[LR] = r[PC] + 2
		reg.LR[i] = r[PC] + 2
	} else {
		r[LR] = r[PC] + 4
		reg.LR[i] = r[PC] + 4
	}

	cpsr.Mode = mode
	cpsr.T = false
	cpsr.I = true

	if cpu.LowVector {
		r[PC] = addr & 0xFFFF
		return
	}

	r[PC] = addr
}

func (cpu *Cpu) ExitException(mode uint32) {

	var (
		cpsr = &cpu.Reg.CPSR
		reg  = &cpu.Reg
		r    = &cpu.Reg.R
	)

	i := BANK_ID[mode]
	*cpsr = reg.SPSR[i]
	c := BANK_ID[cpsr.Mode]

	reg.LR[i] = r[LR]
	reg.SP[i] = r[SP]
	r[SP] = reg.SP[c]
	r[LR] = reg.LR[c]
}
