package arm7gba

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

func (cpu *Cpu) Exception(addr uint32, mode uint32) {

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
	thumb := reg.isThumb

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
	//	cpu.Mem.BIOS_MODE = BIOS_IRQ_POST
	//}

	reg := &cpu.Reg
	r := &cpu.Reg.R

	// PC is updated in final bios inst

	i := BANK_ID[mode]
	reg.CPSR = reg.SPSR[i]
	reg.isThumb = reg.CPSR.GetFlag(FLAG_T)
	c := BANK_ID[cpu.Reg.getMode()]

	// if you set this up for fiq, get the special registers
	reg.LR[i] = r[LR]
	reg.SP[i] = r[SP]
	r[SP] = reg.SP[c]
	r[LR] = reg.LR[c]
}
