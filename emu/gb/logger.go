package gameboy

import (
	"bufio"
	"fmt"
	"os"
)

var L *Logger

type Logger struct {
	Instruction    int
	MaxInstruction int

	gb        *GameBoy
	file      *os.File
	bufWriter *bufio.Writer
}

func NewLogger(path string, gb *GameBoy) *Logger {

	l := &Logger{}
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}

	l.file = f
	l.bufWriter = bufio.NewWriter(f)
	l.gb = gb

	return l
}

func (l *Logger) Close() {
	l.bufWriter.Flush()
	l.file.Close()
}

func (l *Logger) WriteLog(i int, opcode uint8) {

	gb := l.gb

	pc0 := gb.Read(gb.Cpu.PC)
	pc1 := gb.Read(gb.Cpu.PC + 1)
	pc2 := gb.Read(gb.Cpu.PC + 2)
	pc3 := gb.Read(gb.Cpu.PC + 3)

	s := fmt.Sprintf(
		"A:%02X F:%02X B:%02X C:%02X D:%02X E:%02X H:%02X L:%02X SP:%04X PC:%04X PCMEM:%02X,%02X,%02X,%02X",
		gb.Cpu.a,
		gb.Cpu.f.Get(),
		gb.Cpu.b,
		gb.Cpu.c,
		gb.Cpu.d,
		gb.Cpu.e,
		gb.Cpu.h,
		gb.Cpu.l,
		gb.Cpu.SP,
		gb.Cpu.PC,
		pc0,
		pc1,
		pc2,
		pc3,
	)

	fmt.Fprintf(l.bufWriter, "%s\n", s)

	BUF_SIZE := 10_000

	if i%BUF_SIZE == 0 {
		l.bufWriter.Flush()
	}
}
