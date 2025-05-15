# Testing

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

### Tonc

❌ bigmap
❌ bld_demo
👍 bm_modes

👍  brin_demo
   👍 move
   👍 screenblock
   👍 wrap

👍  cbb_demo
    ❌  obj tile in top left (not sure if needed?)
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
❌ oacombo
❌ obj_aff

👍 obj_demo
    👍 move
    👍 palette change
    👍 hflip
    👍 vflip
    👍 decrease / increase starting tile
    👍 1d / 2d mappings

❌ octtest
👍 pageflip
❌ prio_demo
❌ sbb_aff
👍 sbb_reg (has obj in top left, not sure if problem)
👍 second
❌ snd1_demo
❌ swi_demo
❌ swi_vsync
❌ tmr_demo
❌ tte_demo
❌ txt_bm
❌ txt_obj
❌ txt_se1
❌ txt_se2
❌ win_demo
