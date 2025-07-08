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
👍 bios
👍 memory
❌ nes

   ppu
👍 hello
👍 shades
👍 stripes

   save
❌ flash64 (matches mgba)
❌ flash128 (matches mgba)
👍 none
👍 sram (matches mgba)

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

❌ Memory tests [1552/1552]
❌ I/O read tests [129/130] (Final on is related to channel bits not being properly set off and on)
❌ Timing tests [228/2020]
❌ Timer count-up tests [186/936]
❌ Timer IRQ tests [1/90]
👍 Shifter tests [140/140]
👍 Carry tests [93/93]
👍 Multiply long tests [52/72] (matches mgba)
👍 BIOS math tests [615/615]
❌ DMA tests [1240/1256]
❌ SIO register R/W tests [25/90]
❌ SIO timing tests [0/4]
❌ Misc. edge case tests [3/10]
❌ Video tests
    ❌ Basic Mode 3
    ❌ Basic Mode 4
    👍 Degenerate OBJ transforms
    👍 Layer toggle
    👍 Layer toggle 2
    ❌ OAM Update Delay
    👍 Window offscreen reset (matches mgba)

### NBA-EMU Test Suite

❌ bus: 128kb Boundary
❌ dma: burst into tears[0/3]
❌ dma: force nseq access
❌ dma: latch
❌ dma: start delay
❌ halt: halt cnt[0/6]
❌ irq: irq delay
❌ ppu: bgpd
❌ ppu: bgx
❌ ppu: dispcnt-latch
❌ ppu: greenswap
❌ ppu: ram-access-timing
❌ ppu: sprite-hmosaic
❌ ppu: status-irq-dma
❌ ppu: vram-mirror [7/10]
❌ timer: start stop [0/2]
❌ timer: reload [0/7]

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

👍 octtest (blinks)
👍 pageflip
👍 prio_demo
❌ sbb_aff (does not hide at edges)
👍 sbb_reg (has obj in top left, not sure if problem)
👍 second
👍 snd1_demo
👍 swi_demo
👍 swi_vsync
👍 tmr_demo (1.3 works) (uses faux cycle * 4)
❌ tte_demo
❌ txt_bm
❌ txt_obj
👍 txt_se1
👍 txt_se2 (text has different amounts)
👍 win_demo

### Games

All games require Digital Sound

Advance Wars
    - No known errors (maybe menu)
Fire Emblem
    - No known errors
Fire Emblem Sacred Stones
    - No known errors
Golden Sun
    - No known errors
Drill Dozer
    - Objects not appearing, is affine at top of screen
Harvest Moon Friends of Mineral Town
    - Blending of Green
Hello Kitty Happy Party Pals
    - Some Mini games do not load
Kirby Nightmare in Dream Land
    - No known errors
Lord of The Rings Fellowship
    - No known errors
Lord of The Rings Two Towers
    - No known errors
Mario Kart Super Circuit
    - Mode 7
Mega Man Zero
    - Graphics
Metroid Fusion
    - No known errors
Mother 12
    - No known errors
Mother 3
    - No known errors
Pokémon Mystery Dungeon Red Rescue Team
    - Graphics Windows
Pokémon Firered / LeafGreen
    - No known errors
Pokémon Emerald
    - No known errors
Pokémon Ruby / Sapphire
    - No known errors
Sonic Advance
    - No known errors
Spyro Season of Ice
    - No known errors
Superstar Saga
    - No known errors
Super Dodge Ball Advance
    - No known errors
Super Mario Advance
    - No known errors
Tetris Worlds
    - No known errors
The Minish Cap
    - No known errors
Ultimate Puzzle Games
    - No known errors
Warioware Twisted
    - I believe needs Mode 7
Wolfenstein 3D
    - Does not save (verify)
Doom
    - Does not save (verify)
Doom II
    - Does not boot
Zelda Link to the Past
    - No known errors
Iridion II
    - Odd Graphics Errors
Iridion 3D
    - Tiling Graphics Problem
