package gba

import (
	"bufio"
	"fmt"
	"image"

	"image/color"
	"image/png"
	_ "image/png"
	"os"
)

type Debugger struct {
    Gba *GBA
    Version int
}

func (d *Debugger) print(i int) {
    reg := d.Gba.Cpu.Reg
    p := func(a string, b uint32) { fmt.Printf("% 8s: % 9X\n", a, b)}
    s := func(a string) { fmt.Printf("%s\n", a)}

    s("--------  --------")
    fmt.Printf("inst dec %d\n", uint32(i))
    p("inst", uint32(i))

    if d.Gba.Cpu.Reg.CPSR.GetFlag(FLAG_T) {
        p("opcode", d.Gba.Mem.Read16(reg.R[15]))
    } else {
        p("opcode", d.Gba.Mem.Read32(reg.R[15]))
    }
    mode := d.Gba.Cpu.Reg.getMode()
    s("--------  --------")
    p("r00", reg.R[0])
    p("r01", reg.R[1])
    p("r02", reg.R[2])
    p("r03", reg.R[3])
    p("r04", reg.R[4])
    p("r05", reg.R[5])
    p("r06", reg.R[6])
    p("r07", reg.R[7])
    p("r08", reg.R[8])
    p("r09", reg.R[9])
    p("r10", reg.R[10])
    p("r11", reg.R[11])
    p("r12", reg.R[12])
    p("sp/r13", reg.R[13])
    p("lr/r14", reg.R[14])
    p("pc/r15", reg.R[15])
    s("--------  --------")
    p("cpsr", uint32(reg.CPSR))
    p("spsr", uint32(reg.SPSR[BANK_ID[mode]]))
    p("MODE", BANK_ID[mode])
    p("0x3007FFC", d.Gba.Mem.Read32(0x3007FFC))
    p("0x4000004", d.Gba.Mem.Read16(0x4000004))
    p("0x4000208", d.Gba.Mem.Read16(0x4000208))
    p("0x4000200", d.Gba.Mem.Read16(0x4000200))
    p("0x4000202", d.Gba.Mem.Read16(0x4000202))

    s("--------  --------")

    for i := range len(reg.LR) {
        p(fmt.Sprintf("LR %02d", i), uint32(reg.LR[uint32(i)]))
    }

    s("--------  --------")

    //j := uint32(0x4000208)
    //p(fmt.Sprintf("IME %04X", j), d.Gba.Mem.Read16(uint32(j)))
    //j = uint32(0x4000204)
    //p(fmt.Sprintf("WS  %04X", j), d.gba.Mem.Read16(uint32(j)))
    //j = uint32(0x4000202)
    //p(fmt.Sprintf("IF  %04X", j), d.gba.Mem.Read16(uint32(j)))
    //j = uint32(0x4000200)
    //p(fmt.Sprintf("IE  %04X", j), d.gba.Mem.Read16(uint32(j)))

    //s("\n\n")
    //p(fmt.Sprintf("STACK %X", 0x3007E2E), d.gba.Mem.Read32(0x3007E2E))
    //for i := 0x0400_00E0; i >= 0x0400_00D0; i -= 4 {

    //start := 0x40000E0
    //count := 0x10
    //for i := start; i >= start - (count * 4); i -= 4 {
    //    p(fmt.Sprintf("IO %X", i), d.gba.Mem.Read32(uint32(i)))
    //}
    //s("------")
    //start := 0xE00_00D0
    //count := 0x10
    //for i := start; i >= start - (count * 4); i -= 4 {
    //    p(fmt.Sprintf("IO %X", i), d.Gba.Mem.Read32(uint32(i)))
    //}
    start := 0xE00_00D0
    count := 0x10
    for i := start; i >= start - (count); i -- {
        p(fmt.Sprintf("IO %X", i), d.Gba.Mem.Read8(uint32(i)))
    }
}

func (d *Debugger) saveBg4() {

    Mem := d.Gba.Mem

    WIDTH_BG2 := 240
    HEIGHT_BG2 := 160

    img := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{WIDTH_BG2, HEIGHT_BG2}})

    for y := range HEIGHT_BG2 {
        for x := range WIDTH_BG2 {

            palIdx := Mem.VRAM[(y * WIDTH_BG2) + x]

            palData := uint32(Mem.PRAM[(palIdx * 2) + 1]) << 8 | uint32(Mem.PRAM[palIdx * 2])
            r := uint8((palData >> 10) & 0b11111)
            g := uint8((palData >> 5) & 0b11111)
            b := uint8(palData & 0b11111)

            c := convertTo24bit(r, g, b)
            img.Set(x, y, c)
        }
    }

    f, _ := os.Create("bg2.png")
    png.Encode(f, img)
}

func (d *Debugger) saveBg2() {

    Mem := d.Gba.Mem

    WIDTH_BG2 := 240
    HEIGHT_BG2 := 160

    img := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{WIDTH_BG2, HEIGHT_BG2}})

    for y := range HEIGHT_BG2 {
        for x := range WIDTH_BG2 {
            palIdx := Mem.VRAM[(y * WIDTH_BG2) + x]

            palData := uint32(Mem.PRAM[(palIdx * 2) + 1]) << 8 | uint32(Mem.PRAM[palIdx * 2])
            r := uint8((palData >> 10) & 0b11111)
            g := uint8((palData >> 5) & 0b11111)
            b := uint8(palData & 0b11111)

            c := convertTo24bit(r, g, b)
            img.Set(x, y, c)
        }
    }

    f, _ := os.Create("bg2.png")
    png.Encode(f, img)
}


func convertTo24bit(r, g, b uint8) color.RGBA {
    return color.RGBA{
        R: (r << 3) | (r >> 2),
        G: (g << 3) | (g >> 2),
        B: (b << 3) | (b >> 2),
        A: 0xFF,
    }
}

func (d *Debugger) dump(s, e uint32) {

    // fix to buffer some day
    tmp := ""

    for i := s; i <= e; i += 4 {
        tmp += fmt.Sprintf("%08X", d.Gba.Mem.Read32(uint32(i)))
    }
    f, err := os.Create("./dump")
    if err != nil { panic(err) } 
    w := bufio.NewWriter(f)
    _, err = w.WriteString(tmp)

    if err != nil { panic(err) } 

    w.Flush()
}

func (gba *GBA) debugGraphics() {

    const (
        DEBUG_WIDTH = 1080
        DEBUG_HEIGHT = 1080
        palette256 = true
        baseAddr = 0x600_0000
        count = 0x08
    )

    tileSize := 0x20
    if palette256 {
        tileSize = 0x40
    }

	// base addr usually inc of 0x4000 over 0x0600_0000
	// count is # of tiles to view

	for offset := range count {
		tileOffset := offset * tileSize
		tileAddr := baseAddr + tileOffset
		debugTile(gba, uint(tileAddr), tileSize, offset, 0, false, palette256)
	}

}

func debugTile(gba *GBA, tileAddr uint, tileSize, xOffset, yOffset int, obj, palette256 bool) {
    const (
        DEBUG_WIDTH = 120
        DEBUG_HEIGHT = 600
    )

	xOffset *= tileSize
	yOffset *= tileSize

	indexOffset := xOffset + (yOffset * DEBUG_WIDTH)

	mem := gba.Mem
	index := 0
	byteOffset := 0

	for y := range 8 {

		iY := DEBUG_WIDTH * y

		for x := range 8 {

			tileData := mem.Read16(uint32(tileAddr) + uint32(byteOffset))

            //fmt.Printf("%08X %08X\n", tileAddr, mem.VRAM[0x20])

            var palIdx uint32
            if !palette256 {
                bitDepth := 4
                palIdx = (tileData >> uint32(bitDepth)) & 0b1111
                if x%2 == 0 {
                    palIdx = tileData & 0b1111
                }
            } else {
                palIdx = tileData & 0b1111_1111
            }


			palData := gba.getPalette(uint32(palIdx), 0, obj)
			index = (iY + x + indexOffset) * 4

			gba.applyDebugColor(palData, uint32(index))

            if !palette256 {

                if x%2 == 1 {
                    byteOffset += 1
                }

            } else {
                byteOffset += 1
            }
		}
	}
}

func (d *Debugger) debugIRQ() {

    gba := d.Gba
    mem := gba.Mem
    reg := gba.Cpu.Reg
    r := gba.Cpu.Reg.R

    t := reg.CPSR.GetFlag(FLAG_T)
    opcode := uint32(0)
    if !t {
        opcode = mem.Read32(r[PC])
    } else {
        opcode = mem.Read16(r[PC])
    }

    usrBank := BANK_ID[MODE_USR]
    irqBank := BANK_ID[MODE_IRQ]

    fmt.Printf("-----------------------------------------------------------\n")
    fmt.Printf("IRQ CURR INST %d\n", CURR_INST)
    fmt.Printf("PC %08X T %t OPCODE %08X CPSR %08X\n", r[PC], t, opcode, reg.CPSR)
    fmt.Printf("R0 %08X R1 %08X R2 %08X R3 %08X R4 %08X\n", r[0], r[1], r[2], r[3], r[4])
    fmt.Printf("R5 %08X R6 %08X R7 %08X R8 %08X R9 %08X\n", r[5], r[6], r[7], r[8], r[9])
    fmt.Printf("R10 %08X R11 %08X R12 %08X\n", r[10], r[11], r[12])
    fmt.Printf("IRQ STACK ADDR (0x03007FFC) %08X\n", mem.Read32(0x03007FFC))
    fmt.Printf("STACK ADDR %08X, VALUE %08X\n", r[SP]+20, mem.Read32(r[SP]+20))
    fmt.Printf("MODE %02X (If Mode %02X IRQ,if %02X USR)\n", reg.getMode(), MODE_IRQ, MODE_USR)
    fmt.Printf("CURR SP %08X, LR %08X SPSR %08X\n", r[SP], r[LR], reg.SPSR[BANK_ID[reg.getMode()]])
    fmt.Printf("USR  SP %08X, LR %08X SPSR %08X\n", reg.SP[usrBank], reg.LR[usrBank], reg.SPSR[usrBank])
    fmt.Printf("IRQ  SP %08X, LR %08X SPSR %08X\n", reg.SP[irqBank], reg.LR[irqBank], reg.SPSR[irqBank])
    fmt.Printf("-----------------------------------------------------------\n")
}

type Logger struct {
	Instruction    int
	MaxInstruction int

    gba *GBA
    file *os.File
    bufWriter *bufio.Writer
}

func NewLogger(path string, gba *GBA) *Logger {

    l := Logger{}
    f, err := os.Create(path)
    if err != nil { panic(err) }

    l.file = f
    l.bufWriter = bufio.NewWriter(f)
    l.gba = gba

    return &l
}

func (l *Logger) Close() {
    l.bufWriter.Flush()
    l.file.Close()
}

func (l *Logger) WriteLog() {

    gba := l.gba

    s := fmt.Sprintf(
        "CURR %08X INST %08X MODE %08X CPSR %08X SPSR %08X R00 %08X R01 %08X R02 %08X R03 %08X R04 %08X R05 %08X R06 %08X R07 %08X R08 %08X R09 %08X R10 %08X R11 %08X R12 %08X R13 %08X R14 %08X R15 %08X R14B0 %08X IME %08X, IE %08X, IF %08X",
        CURR_INST, gba.Mem.Read32(gba.Cpu.Reg.R[15]), gba.Cpu.Reg.getMode(), gba.Cpu.Reg.CPSR, gba.Cpu.Reg.SPSR[BANK_ID[gba.Cpu.Reg.getMode()]],
        gba.Cpu.Reg.R[0],
        gba.Cpu.Reg.R[1],
        gba.Cpu.Reg.R[2],
        gba.Cpu.Reg.R[3],
        gba.Cpu.Reg.R[4],
        gba.Cpu.Reg.R[5],
        gba.Cpu.Reg.R[6],
        gba.Cpu.Reg.R[7],
        gba.Cpu.Reg.R[8],
        gba.Cpu.Reg.R[9],
        gba.Cpu.Reg.R[10],
        gba.Cpu.Reg.R[11],
        gba.Cpu.Reg.R[12],
        gba.Cpu.Reg.R[13],
        gba.Cpu.Reg.R[14],
        gba.Cpu.Reg.R[15],
        gba.Mem.Read32(0x400_0208),
        gba.Mem.Read32(0x400_0200),
        gba.Mem.Read32(0x400_0202),
        gba.Cpu.Reg.LR[0],
    )

    fmt.Fprintf(l.bufWriter, "%s\n", s)

    BUF_SIZE := 10_000

    if CURR_INST%BUF_SIZE == 0 {
        l.bufWriter.Flush()
    }
}

