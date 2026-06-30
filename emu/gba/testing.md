# Testing

[GBA Tests](https://emulation.gametechwiki.com/index.php/GBA_Tests)
### DenSinH / FuzzARM

👍 ARM_DataProcessing
👍 ARM_Any
👍 THUMB_DataProcessing
👍 THUMB_Any
👍 FuzzARM

### jsmolka / gba-tests

👍 arm
👍 thumb
👍 bios
👍 memory
👍 nes

   ppu
👍 hello
👍 shades
👍 stripes

   save
👍 flash64
❌ flash128
👍 none
👍 sram

note:
Flash128 fails if save exists. Test 009 requires cleared memory to pass.
This is because according to gbatek the target memory location must have been previously erased.

Nanoboy treats if as a free write, [addr] = v
Skyemu forces has to be erased addr &= v

not sure which is correct and if test has accounted for it.

### Arm Wrestler

[Link](https://github.com/destoer/armwrestler-gba-fixed/)

The standard version of arm wrestler is not for gba emulation.
Accurate GBA Emulators will fail on Ldm--! instructions, because of differences
in ARMv4 behavior.

(LDM opcodes with writeback: if the base register is included in the register list, writeback never happens)
Additionally, other ARMv5 instructions will fail.

👍 ARM ALU
👍 ARM LDR/STR
👍 ARM LDM/STM
👍 THUMB ALU
👍 THUMB LDR/STR
👍 THUMB LDM/STM

### hades-emu/Hades-Tests

[Link](https://github.com/hades-emu/Hades-Tests)

👍 Bios Open Bus [12/12]
❌ Dma Latch [3/4]
❌ Dma Start Delay [6/8]
👍 Timer Basic [10/10]

### Other
 
👍 deadbody Cpu Test

### MGBA Test Suite

Requires Official bios for passing tests

👍 Memory tests [1552/1552]
👍 I/O read tests [130/130]
👍 Timing tests [2020/2020]
❌ Timer count-up tests [903/936]
👍 Timer IRQ tests [90/90]
👍 Shifter tests [140/140]
👍 Carry tests [93/93]
❌ Multiply long tests [52/72] (matches mgba)
👍 BIOS math tests [615/615]
👍 DMA tests [1256/1256]
❌ SIO register R/W tests [25/90]
❌ SIO timing tests [0/4]
❌ Misc. edge case tests [3/10]
❌ Video tests
    👍 Basic Mode 3
    👍 Basic Mode 4
    👍 Degenerate OBJ transforms
    👍 Layer toggle
    ❌ Layer toggle 2
    ❌ OAM Update Delay
    ❌ Window offscreen reset (matches mgba)

### NBA-EMU Test Suite

❌ bus: 128kb Boundary
❌ dma: burst into tears [0/3]
👍 dma: force nseq access [2/2]
👍 dma: latch [3/3]
❌ dma: start delay [0/1]
❌ halt: halt cnt [0/6]
👍 irq: irq delay [3/3]
❌ ppu: bgpd
❌ ppu: bgx
❌ ppu: dispcnt-latch
👍 ppu: greenswap
❌ ppu: ram-access-timing
❌ ppu: sprite-hmosaic
❌ ppu: status-irq-dma
❌ ppu: vram-mirror [7/10]
👍 timer: start stop [2/2]
👍 timer: reload [7/7]

### AGS

👍 Memory
👍 Lcd
👍 Timer
❌ Dma (priority test)
👍 Key
👍 Irq

### Tonc

👍 bigmap
👍 bld_demo
👍 bm_modes

👍 brin_demo
   👍 move
   👍 screenblock
   👍 wrap

👍 cbb_demo
    👍 obj tile in top left (not sure if needed?)
    👍 0102/1011
    👍 2122/3031
    👍 no extra

👍 dma_demo
👍 first
👍 hello

👍 irq_demo
👍 key_demo
👍 m3_demo
👍 m7_demo
👍 m7_demo_mb
👍 m7_ex
❌ mos_demo
    👍 ObjH
    ❌ ObjV - final height is different, minor difference
    👍 BgH
    👍 BgV

👍 oacombo

❌ obj_aff
   👍 move
   👍 rotate
   👍 scale
   👍 shear
   👍 text
   👍 mask
   👍 double size
   👍 origin
   ❌ edge jerking / disappearing (normal and double mode also does work)
   👍 bg and obj layering

👍 obj_demo
    👍 move
    👍 palette change
    👍 hflip
    👍 vflip
    👍 decrease / increase starting tile
    👍 1d / 2d mappings

👍 octtest (blinks)
👍 pageflip
👍 prio_demo
👍 sbb_aff
👍 sbb_reg (has obj in top left, not sure if problem)
👍 second
👍 snd1_demo
👍 swi_demo
👍 swi_vsync
👍 tmr_demo
❌ tte_demo
❌ txt_bm
❌ txt_obj
👍 txt_se1
👍 txt_se2 (text has different amounts)
👍 win_demo

# Questions

1. Do all 72 long multiply tests pass on actual hardware?
2. Does Flash memory actually need to be zeroed (0xFF) before write? See Flash 128 problems
