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

	if mode != MODE_IRQ && mode != MODE_SWI {
		panic("UNKNOWN EXCEPTION MODE")
	}

	reg := &cpu.Reg
	r := &cpu.Reg.R

	curr := reg.getMode()

	if mode == curr {
		return
	}

	//switch mode {
	//case MODE_IRQ:
	//	cpu.mem.BIOS_MODE = BIOS_IRQ
	//case MODE_SWI:
	//	gba.Mem.BIOS_MODE = BIOS_SWI
	//}

	//thumb := reg.CPSR.GetFlag(FLAG_T)
	thumb := reg.IsThumb

	c := BANK_ID[reg.getMode()]
	i := BANK_ID[mode]
	reg.SP[c] = r[SP]
	reg.LR[c] = r[LR]
	r[SP] = reg.SP[i]
	r[LR] = reg.LR[i]
	reg.SPSR[i] = reg.CPSR

	switch {
	case mode == MODE_SWI && thumb:
		r[LR] = r[PC] + 2
		reg.LR[i] = r[PC] + 2
	default:
		r[LR] = r[PC] + 4
		reg.LR[i] = r[PC] + 4
	}

	reg.CPSR.SetMode(mode)
	reg.CPSR.SetThumb(false, cpu)
	reg.CPSR.SetFlag(FLAG_I, true)

	r[PC] = addr
	return
}

func (cpu *Cpu) ExitException(mode uint32) {

	//if mode == MODE_IRQ {
	//	cpu.mem.BIOS_MODE = BIOS_IRQ_POST
	//}

	reg := &cpu.Reg
	r := &cpu.Reg.R

	// PC is updated in final bios inst

	i := BANK_ID[mode]
	reg.CPSR = reg.SPSR[i]
	reg.IsThumb = reg.CPSR.GetFlag(FLAG_T)
	c := BANK_ID[cpu.Reg.getMode()]

	// if you set this up for fiq, get the special registers
	reg.LR[i] = r[LR]
	reg.SP[i] = r[SP]
	r[SP] = reg.SP[c]
	r[LR] = reg.LR[c]
}
