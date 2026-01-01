package arm7

const (
	VEC_RESET         = 0x00
	VEC_UND           = 0x04
	VEC_SWI           = 0x08
	VEC_PREFETCHABORT = 0x0C
	VEC_DATAABORT     = 0x10
	VEC_ADDR26BIT     = 0x14
	VEC_IRQ           = 0x18
	VEC_FIQ           = 0x1C
)

func (cpu *Cpu) exception(addr uint32, mode uint32) {

	if mode != MODE_IRQ && mode != MODE_SWI {
		panic("UNKNOWN EXCEPTION MODE")
	}

	reg := &cpu.Reg
	r := &cpu.Reg.R

	curr := reg.CPSR.Mode

	if mode == curr {
		return
	}

	c := BANK_ID[reg.CPSR.Mode]
	i := BANK_ID[mode]
	reg.SP[c] = r[SP]
	reg.LR[c] = r[LR]
	r[SP] = reg.SP[i]
	r[LR] = reg.LR[i]
	reg.SPSR[i] = reg.CPSR

	switch {
	case mode == MODE_SWI && reg.CPSR.T:
		r[LR] = r[PC] + 2
		reg.LR[i] = r[PC] + 2
	default:
		r[LR] = r[PC] + 4
		reg.LR[i] = r[PC] + 4
	}

	reg.CPSR.Mode = mode
	reg.CPSR.T = false
	reg.CPSR.I = true

	r[PC] = addr
}

func (cpu *Cpu) ExitException(mode uint32) {

	reg := &cpu.Reg
	r := &cpu.Reg.R

	i := BANK_ID[mode]
	reg.CPSR = reg.SPSR[i]
	c := BANK_ID[cpu.Reg.CPSR.Mode]

	// if you set this up for fiq, get the special registers
	reg.LR[i] = r[LR]
	reg.SP[i] = r[SP]
	r[SP] = reg.SP[c]
	r[LR] = reg.LR[c]
}
