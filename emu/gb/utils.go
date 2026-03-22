package gameboy

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
)

func LoadLine(gb *GameBoy, line int) {

	d, err := os.ReadFile(fmt.Sprintf("linedump_%d_01", line))
	if err != nil {
		panic(err)
	}

	start := 0
	lineCount := 0
	for i := range len(d) {
		if byte(d[i]) == byte('\n') {
			v := d[start:i]
			z, _ := strconv.Atoi(string(v))
			switch lineCount {
			case 0:
				gb.Cpu.PC = uint16(z)
			case 1:
				gb.Cpu.SP = uint16(z)
			case 2:
			case 3:
			case 4:
				gb.Cpu.a = uint8(z)
			case 5:
				gb.Cpu.b = uint8(z)
			case 6:
				gb.Cpu.c = uint8(z)
			case 7:
				gb.Cpu.d = uint8(z)
			case 8:
				gb.Cpu.e = uint8(z)
			case 9:
				//gb.Cpu.g = uint8(z)
			case 10:
				gb.Cpu.h = uint8(z)
			case 11:
				gb.Cpu.l = uint8(z)
			case 12:
				gb.Cpu.f.Z = v[0] == 't'
			case 13:
				gb.Cpu.f.S = v[0] == 't'
			case 14:
				gb.Cpu.f.H = v[0] == 't'
			case 15:
				gb.Cpu.f.C = v[0] == 't'
			case 16:
				gb.Cpu.IME = v[0] == 't'
			case 17:
				gb.Cpu.PendingInterrupt = v[0] == 't'
			case 18:
			}

			lineCount++
			start = i + 1
		}
	}

	d, err = os.ReadFile(fmt.Sprintf("linedump_%d_02", line))
	if err != nil {
		panic(err)
	}

	for i := range len(gb.MemoryBus.Memory) {
		gb.MemoryBus.Memory[i] = d[i]
	}
}

func DumpLine(gb *GameBoy, line int) {
	o := ""
	o += fmt.Sprintf("%d\n", gb.Cpu.PC)
	o += fmt.Sprintf("%d\n", gb.Cpu.SP)
	o += fmt.Sprintf("%d\n", 0)
	o += fmt.Sprintf("%d\n", 0)
	o += fmt.Sprintf("%d\n", gb.Cpu.a)
	o += fmt.Sprintf("%d\n", gb.Cpu.b)
	o += fmt.Sprintf("%d\n", gb.Cpu.c)
	o += fmt.Sprintf("%d\n", gb.Cpu.d)
	o += fmt.Sprintf("%d\n", gb.Cpu.e)
	//o += fmt.Sprintf("%d\n", gb.Cpu.g)
	o += fmt.Sprintf("%d\n", gb.Cpu.h)
	o += fmt.Sprintf("%d\n", gb.Cpu.l)
	o += fmt.Sprintf("%t\n", gb.Cpu.f.Z)
	o += fmt.Sprintf("%t\n", gb.Cpu.f.S)
	o += fmt.Sprintf("%t\n", gb.Cpu.f.H)
	o += fmt.Sprintf("%t\n", gb.Cpu.f.C)
	o += fmt.Sprintf("%t\n", gb.Cpu.IME)
	o += fmt.Sprintf("%t\n", gb.Cpu.PendingInterrupt)

	err := os.WriteFile(fmt.Sprintf("linedump_%d_01", line), []byte(o), 0644)
	if err != nil {
		panic(err)
	}

	var p []byte
	for i := range len(gb.MemoryBus.Memory) {
		p = append(p, gb.MemoryBus.Memory[i])
	}

	err = os.WriteFile(fmt.Sprintf("linedump_%d_02", line), p, 0644)
	if err != nil {
		panic(err)
	}
}

type SerialOutput struct {
	output   []byte
	Filename string
	prev     uint8
}

func (s *SerialOutput) DebugSerialOutput(gb *GameBoy) {
	// blarggs tests output to serial port if sc byte is 0x81, read prev byte

	if sc := gb.Read(uint16(0xFF02)); sc == 0x81 {
		sb := gb.Read(uint16(0xFF01))
		s.output = append(s.output, sb)
		gb.MemoryBus.Memory[0xFF02] = 0x80
	}
}

func (s *SerialOutput) Export() {

	err := os.WriteFile(s.Filename, s.output, 0644)
	if err != nil {
		panic(err)
	}
}

func (s *SerialOutput) Print() {

	for _, b := range s.output {
		fmt.Printf("%s", string(b))
	}
}

func DumpMemory(gb *GameBoy, filename string) {

	var o []byte

	for i := range len(gb.MemoryBus.Memory) {
		o = append(o, gb.MemoryBus.Memory[i])
	}

	err := os.WriteFile(filename, o, 0644)
	if err != nil {
		panic(err)
	}
}

func InitializeLog(filename string) (*os.File, *bufio.Writer) {

	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}

	// defer in parent func

	bufWriter := bufio.NewWriter(f)

	return f, bufWriter
}

func WriteLog(i int, gb GameBoy, opcode uint8, bufWriter *bufio.Writer) {

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

	fmt.Fprintf(bufWriter, "%s\n", s)

	BUF_SIZE := 10_000

	if i%BUF_SIZE == 0 {
		bufWriter.Flush()
	}
}
