package arm9

import (
	a "github.com/aabalke/gojit/arm64"
)

func (j *Jit) emitClz(op uint32) {
	j.LdrReg(a.R00, op & 0xF)
	j.Clz(a.R00, a.R00, false)
	j.StrReg(a.R00, (op >> 12) & 0xF)
}

func (j *Jit) emitMul(op uint32) {

	var (
		inst = (op >> 21) & 0xF
		set  = (op >> 20) & 1 != 0
		rd   = (op >> 16) & 0xF
		rn   = (op >> 12) & 0xF
		rs   = (op >> 8) & 0xF
		rm   = op & 0xF
	)

	switch inst {
	case MUL, MLA:

		j.LdrReg(a.R00, rs)
		j.LdrReg(a.R01, rm)

		acc := a.RZR

		if inst == MLA {
			j.LdrReg(a.R02, rn)
			acc = a.R02
		}

		j.MAdd(a.R00, a.R00, a.R01, acc, false)

		j.StrReg(a.R00, rd)

		if set {
			j.TstReg(a.R00, a.R00, 0, 0, false)
			j.Cset(a.R00, a.N, false)
			j.StrFlag(a.R00, N)

			j.Cset(a.R00, a.Z, false)
			j.StrFlag(a.R00, Z)
		}

		return

	case UMULL, UMLAL, SMULL, SMLAL:

		j.LdrReg(a.R00, rs)
		j.LdrReg(a.R01, rm)

		acc := a.RZR

		if inst == UMLAL || inst == SMLAL {
			j.LdrReg(a.R02, rd)
			j.LslImm(a.R02, a.R02, 32, true)
			j.LdrReg(a.R03, rn)
			j.Bfi(a.R02, a.R03, 32, 0, true)

			acc = a.R02
		}

		if inst == SMLAL || inst == SMULL {
			j.Smaddl(a.R00, a.R00, a.R01, acc)
		} else {
			j.Umaddl(a.R00, a.R00, a.R01, acc)
		}

		if set {
			j.TstReg(a.R00, a.R00, 0, 0, true)
			j.Cset(a.R01, a.N, false)
			j.StrFlag(a.R01, N)

			j.Cset(a.R01, a.Z, false)
			j.StrFlag(a.R01, Z)

			j.Movz(a.R01, 0, 0, false)
			j.StrFlag(a.R01, C)
		}

		j.StrReg(a.R00, rn)
		j.LsrImm(a.R00, a.R00, 32, true)
		j.StrReg(a.R00, rd)

		return
	}

	x := (op >> 5) & 1 != 0
	y := (op >> 6) & 1 != 0

	switch inst {
	case SMLAxy:

		// 32 bit rm and rs -> 32/16 bit signed
		j.LdrReg(a.R00, rm)

		if x {
			j.AsrImm(a.R00, a.R00, 16, false)
		}
		j.LslImm(a.R00, a.R00, 16, false)
		j.AsrImm(a.R00, a.R00, 16, false)

		j.LdrReg(a.R01, rs)

		if y {
			j.AsrImm(a.R01, a.R01, 16, false)
		}
		j.LslImm(a.R01, a.R01, 16, false)
		j.AsrImm(a.R01, a.R01, 16, false)

		// 32 bit rn -> 64 bit signed
		j.LdrReg(a.R02, rn)
		j.LslImm(a.R02, a.R02, 32, true)
		j.AsrImm(a.R02, a.R02, 32, true)

		j.Smaddl(a.R00, a.R00, a.R01, a.R02)

		// to check overflow, check if signed 64bit version == signed 32 bit version
		// if not equal, has to be overflow

		j.Sxtw(a.R01, a.R00)
		j.CmpReg(a.R00, a.R01, 0, 0, true)

		j.Cset(a.R01, a.NE, false)
		j.LdrFlag(a.R02, Q)
		j.OrrReg(a.R01, a.R01, a.R02, 0, 0, false)
		j.StrFlag(a.R01, Q)

		j.StrReg(a.R00, rd)

	case SMLAWySMLALWy:

		j.LdrReg(a.R00, rm)

		j.LdrReg(a.R01, rs)
		if y {
			j.AsrImm(a.R01, a.R01, 16, false)
		}
		j.LslImm(a.R01, a.R01, 16, false)
		j.AsrImm(a.R01, a.R01, 16, false)

		j.Smaddl(a.R00, a.R00, a.R01, a.RZR)

		j.AsrImm(a.R00, a.R00, 16, false)

		if smulwa := !x; smulwa {
			j.LdrReg(a.R01, rn)
			j.ADDReg(a.R00, a.R00, a.R01, 0, 0, true, false, false)

			j.Cset(a.R01, a.V, false)
			j.LdrFlag(a.R02, Q)
			j.OrrReg(a.R01, a.R01, a.R02, 0, 0, false)
			j.StrFlag(a.R01, Q)
		}

		j.StrReg(a.R00, rd)

	case SMLALxy:

		j.LdrReg(a.R00, rm)
		if x {
			j.AsrImm(a.R00, a.R00, 16, false)
		}
		j.LslImm(a.R00, a.R00, 16, false)
		j.AsrImm(a.R00, a.R00, 16, false)

		j.LdrReg(a.R01, rs)
		if y {
			j.AsrImm(a.R01, a.R01, 16, false)
		}
		j.LslImm(a.R01, a.R01, 16, false)
		j.AsrImm(a.R01, a.R01, 16, false)

		j.LdrReg(a.R02, rd)
		j.LslImm(a.R02, a.R02, 32, true)
		j.LdrReg(a.R03, rn)
		j.Bfi(a.R02, a.R03, 32, 0, true)

		j.Smaddl(a.R00, a.R00, a.R01, a.R02)

		j.StrReg(a.R00, rn)
		j.LsrImm(a.R00, a.R00, 32, true)
		j.StrReg(a.R00, rd)

	case SMULxy:

		j.LdrReg(a.R00, rm)
		if x {
			j.AsrImm(a.R00, a.R00, 16, false)
		}

		j.LdrReg(a.R01, rs)
		if y {
			j.AsrImm(a.R01, a.R01, 16, false)
		}

		j.Smaddl(a.R00, a.R00, a.R01, a.RZR)

		j.StrReg(a.R00, rd)
	}
}
