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
	//pc2 := gb.Read(gb.Cpu.PC + 2)
	//pc3 := gb.Read(gb.Cpu.PC + 3)

	//s := fmt.Sprintf(
	//	"FFFC %02X %02X SP:%04X PC:%04X PCMEM:%02X,%02X,%02X,%02X SCX %02X LY %02X IF %003X IE %03X",
	//    gb.Read(0xFFFC),
	//    gb.Read(0xFFFD),
	//	gb.Cpu.SP,
	//	gb.Cpu.PC,
	//	pc0,
	//	pc1,
	//	pc2,
	//	pc3,
	//    gb.MemoryBus.IO[0x43],
	//    gb.MemoryBus.IO[0x44],
	//    gb.Cpu.IF,
	//    gb.Cpu.IE,
	//)

	s := fmt.Sprintf(
		"AF=%02X%02X BC=%04X DE=%04X HL=%04X SP=%04X PC=%04X PCMEM=%02X,%02X IF=%02X IE=%02X",
		gb.Cpu.a,
		gb.Cpu.f.Get(),
		*gb.Cpu.BC,
		*gb.Cpu.DE,
		*gb.Cpu.HL,
		gb.Cpu.SP,
		gb.Cpu.PC,
		pc0,
		pc1,
		gb.Cpu.IF,
		gb.Cpu.IE,
	)

	fmt.Fprintf(l.bufWriter, "%s\n", s)

	fmt.Printf("%s\n", s)

	BUF_SIZE := 10_000

	if i%BUF_SIZE == 0 {
		l.bufWriter.Flush()
	}
}
