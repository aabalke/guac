package gba

import (
	"fmt"
	"image"
    "bufio"

	"image/color"
	"image/png"
	_ "image/png"
	"os"
)

type Debugger struct {
    gba *GBA
}

func (d *Debugger) print(i int) {
    reg := d.gba.Cpu.Reg
    p := func(a string, b uint32) { fmt.Printf("% 8s: % 9X\n", a, b)}
    s := func(a string) { fmt.Printf("%s\n", a)}

    s("--------  --------")
    p("inst", uint32(i))

    if d.gba.Cpu.Reg.CPSR.GetFlag(FLAG_T) {
        p("opcode", d.gba.Mem.Read16(reg.R[15]))
    } else {
        p("opcode", d.gba.Mem.Read32(reg.R[15]))
    }
    mode := d.gba.Cpu.Reg.getMode()
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

    //for i := 0x0300_7E98; i <= 0x0300_7ec0; i += 4 {
    ////for i := 0x0300_1B00; i <= 0x0300_1B40; i += 4 {
    //    p(fmt.Sprintf("VRAM %X", i), d.gba.Mem.Read32(uint32(i)))
    //}
    s("--------  --------")
    for i := 0x0300_1BD0; i <= 0x0300_1C20; i += 4 {
    //for i := 0x0300_1B08; i <= 0x0300_1B38; i += 4 {
        p(fmt.Sprintf("VRAM %X", i), d.gba.Mem.Read32(uint32(i)))
    }
}

func (d *Debugger) saveBg4() {

    Mem := d.gba.Mem

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

    Mem := d.gba.Mem

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
        tmp += fmt.Sprintf("%08X", d.gba.Mem.Read32(uint32(i)))
    }
    f, err := os.Create("./dump")
    if err != nil { panic(err) } 
    w := bufio.NewWriter(f)
    _, err = w.WriteString(tmp)

    if err != nil { panic(err) } 

    w.Flush()
}
