# Testing

https://emulation.gametechwiki.com/index.php/GBA_Tests

### DenSinH / FuzzARM

👍 ARM_DataProcessing
👍 ARM_Any
👍 THUMB_DataProcessing
👍 THUMB_Any
👍 FuzzARM

### jsmolka / gba-tests

👍 arm
👍 thumb
❌ bios
❌ memory (passing mirroring, failing str video)
❌ nes

   ppu
👍 hello
👍 shades
👍 stripes

   save
❌ flash64
❌ flash128
👍 none
❌ sram

### Arm Wrestler

https://github.com/destoer/armwrestler-gba-fixed/

Preface: The Standard version of Arm Wrestler floating around is for NDS emulation.
Accurate GBA emulators will fail on LDM--! instructions, since ARMv4 behavior differs.
(LDM opcodes with writeback: if the base register is included in the register list, writeback never happens)
Additionally, other ARMv5 instructions will fail.

This emulator is tested against the destoer/armwrestler-gba-fixed version, which has fixed these problems.

👍 ARM ALU
👍 ARM LDR/STR
👍 ARM LDM/STM
👍 THUMB ALU
👍 THUMB LDR/STR
👍 THUMB LDM/STM

### Other
 
👍 deadbody Cpu Test

### MGBA Test Suite

❌ Memory tests [1153/1552]
❌ I/O read tests [12/130]
❌ Timing tests [240/2020]
❌ Timer count-up tests [crash]
❌ Timer IRQ tests [0/90]
👍 Shifter tests [140/140]
👍 Carry tests [93/93]
👍 Multiply long tests [52/72] (matches mgba)
👍 BIOS math tests [615/615]
❌ DMA tests [1018/1256]
❌ SIO register R/W tests [7/90]
❌ SIO timing tests [0/8]
❌ Misc. edge case tests [0/10]
❌ Video tests
    ❌ Basic Mode 3
    ❌ Basic Mode 4
    ❌ Degenerate OBJ transoforms
    ❌ Layer toggle
    ❌ Layer toggle 2
    ❌ OAM Update Delay
    ❌ Window offscreen reset

### Tonc

! Affine and Mode 1, Mode 2 are temp disabled !

👍 bigmap
❌ bld_demo (need to complete black and white blend)
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

❌ dma_demo
👍 first
👍 hello

❌ irq_demo
👍 key_demo
👍 m3_demo
❌ m7_demo
❌ m7_demo_mb
❌ m7_ex
❌ mos_demo
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

❌ octtest
👍 pageflip
👍 prio_demo
❌ sbb_aff
👍 sbb_reg (has obj in top left, not sure if problem)
👍 second
❌ snd1_demo
👍 swi_demo (1.3 works)
👍 swi_vsync (works but speed is off from cycle counts)
👍 tmr_demo (1.3 works) (uses faux cycle * 4)
❌ tte_demo
❌ txt_bm
❌ txt_obj
👍 txt_se1
👍 txt_se2 (text has different amounts)
👍 win_demo
