package gameboy

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
)

func LoadLine(gb *GameBoy, line int) {

    d, err := os.ReadFile(fmt.Sprintf("linedump_%d_01",line))
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
                gb.Cpu.Registers.a = uint8(z)
            case 5:
                gb.Cpu.Registers.b = uint8(z)
            case 6:
                gb.Cpu.Registers.c = uint8(z)
            case 7:
                gb.Cpu.Registers.d = uint8(z)
            case 8:
                gb.Cpu.Registers.e = uint8(z)
            case 9:
                gb.Cpu.Registers.g = uint8(z)
            case 10:
                gb.Cpu.Registers.h = uint8(z)
            case 11:
                gb.Cpu.Registers.l = uint8(z)
            case 12:
                gb.Cpu.Registers.f.Zero = v[0] == 't'
            case 13:
                gb.Cpu.Registers.f.Subtraction = v[0] == 't'
            case 14:
                gb.Cpu.Registers.f.HalfCarry = v[0] == 't'
            case 15:
                gb.Cpu.Registers.f.Carry = v[0] == 't'
            case 16:
                gb.Cpu.InterruptMaster = v[0] == 't'
            case 17:
                gb.Cpu.PendingInterrupt = v[0] == 't'
            case 18:
            }

            lineCount++
            start = i+1
        }
    }

    d, err = os.ReadFile(fmt.Sprintf("linedump_%d_02",line))
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
    o += fmt.Sprintf("%d\n", gb.Cpu.Registers.a)
    o += fmt.Sprintf("%d\n", gb.Cpu.Registers.b)
    o += fmt.Sprintf("%d\n", gb.Cpu.Registers.c)
    o += fmt.Sprintf("%d\n", gb.Cpu.Registers.d)
    o += fmt.Sprintf("%d\n", gb.Cpu.Registers.e)
    o += fmt.Sprintf("%d\n", gb.Cpu.Registers.g)
    o += fmt.Sprintf("%d\n", gb.Cpu.Registers.h)
    o += fmt.Sprintf("%d\n", gb.Cpu.Registers.l)
    o += fmt.Sprintf("%t\n", gb.Cpu.Registers.f.Zero)
    o += fmt.Sprintf("%t\n", gb.Cpu.Registers.f.Subtraction)
    o += fmt.Sprintf("%t\n", gb.Cpu.Registers.f.HalfCarry)
    o += fmt.Sprintf("%t\n", gb.Cpu.Registers.f.Carry)
    o += fmt.Sprintf("%t\n", gb.Cpu.InterruptMaster)
    o += fmt.Sprintf("%t\n", gb.Cpu.PendingInterrupt)

    err := os.WriteFile(fmt.Sprintf("linedump_%d_01",line), []byte(o), 0644)
    if err != nil {
        panic(err)
    }

    var p []byte
    for i := range len(gb.MemoryBus.Memory) {
        p = append(p, gb.MemoryBus.Memory[i])
    }

    err = os.WriteFile(fmt.Sprintf("linedump_%d_02",line), p, 0644)
    if err != nil {
        panic(err)
    }
}

type SerialOutput struct {
    output []byte
    Filename string
    prev uint8
}

func (s *SerialOutput) DebugSerialOutput(gb *GameBoy) {
    // blarggs tests output to serial port if sc byte is 0x81, read prev byte

    sc, err := gb.ReadByte(uint16(0xFF02))
    if err != nil {
        panic(err)
    }

    if sc == 0x81 {
        sb, err := gb.ReadByte(uint16(0xFF01))
        if err != nil {
            panic(err)
        }

        s.output = append(s.output, sb)
        gb.MemoryBus.Memory[0xFF02] = 0x80
    }
}

func (s *SerialOutput) Export() {

    err := os.WriteFile(s.Filename, s.output, 0644)
    if err != nil { panic(err) }
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

func InitializeLog(filename string) (*os.File, *bufio.Writer){

    f, err := os.Create(filename)
    if err != nil { panic(err) }

    // defer in parent func

    bufWriter := bufio.NewWriter(f)

    return f, bufWriter
}

func WriteLog(i int, gb GameBoy, opcode uint8, bufWriter *bufio.Writer) {

    pc0, _ := gb.ReadByte(gb.Cpu.PC)
    pc1, _ := gb.ReadByte(gb.Cpu.PC+1)
    pc2, _ := gb.ReadByte(gb.Cpu.PC+2)
    pc3, _ := gb.ReadByte(gb.Cpu.PC+3)

    s := fmt.Sprintf(
        "A:%02X F:%02X B:%02X C:%02X D:%02X E:%02X H:%02X L:%02X SP:%04X PC:%04X PCMEM:%02X,%02X,%02X,%02X",
        gb.Cpu.Registers.a,
        gb.Cpu.Registers.f.getBits(),
        gb.Cpu.Registers.b,
        gb.Cpu.Registers.c,
        gb.Cpu.Registers.d,
        gb.Cpu.Registers.e,
        gb.Cpu.Registers.h,
        gb.Cpu.Registers.l,
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
