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

func (cpu *Cpu) Reset() {
	cpu.exception(VEC_RESET, MODE_SWI)
}

func (cpu *Cpu) SoftReset() {
	cpu.exception(VEC_SWI, MODE_SWI)
}

func (cpu *Cpu) IRQException() {
	cpu.exception(VEC_IRQ, MODE_IRQ)
}

func (cpu *Cpu) exception(addr uint32, mode uint32) {

	reg := &cpu.Reg
	r := &cpu.Reg.R

	if mode == reg.CPSR.Mode {
		return
	}

    thumb := reg.CPSR.T

	c := BANK_ID[reg.CPSR.Mode]
	i := BANK_ID[mode]
	reg.SP[c] = r[SP]
	reg.LR[c] = r[LR]
	r[SP] = reg.SP[i]
	r[LR] = reg.LR[i]
	reg.SPSR[i] = reg.CPSR

    if thumb && (mode == MODE_SWI || mode == MODE_ABT) {
		r[LR] = r[PC] + 2
		reg.LR[i] = r[PC] + 2
    } else {
		r[LR] = r[PC] + 4
		reg.LR[i] = r[PC] + 4
    }

	reg.CPSR.Mode = mode
    reg.CPSR.T = false
    reg.CPSR.I = true

    if cpu.LowVector {
        r[PC] = addr & 0xFFFF
        return
    }

    r[PC] = addr
}

func (cpu *Cpu) ExitException(mode uint32) {

	reg := &cpu.Reg
	r := &cpu.Reg.R

	// PC is updated in final bios inst

	i := BANK_ID[mode]
	reg.CPSR = reg.SPSR[i]
	c := BANK_ID[cpu.Reg.CPSR.Mode]

	// if you set this up for fiq, get the special registers
	reg.LR[i] = r[LR]
	reg.SP[i] = r[SP]
	r[SP] = reg.SP[c]
	r[LR] = reg.LR[c]
}
