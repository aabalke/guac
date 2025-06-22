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
❌ sram (matches mgba)

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

❌ Memory tests [1542/1552] (1552 with interrupts timed properly)
❌ I/O read tests [129/130] (Final on is related to channel bits not being properly set off and on)
❌ Timing tests [228/2020]
❌ Timer count-up tests [264/936]
❌ Timer IRQ tests [0/90]
👍 Shifter tests [140/140]
👍 Carry tests [93/93]
👍 Multiply long tests [52/72] (matches mgba)
👍 BIOS math tests [615/615]
❌ DMA tests [1204/1256]
❌ SIO register R/W tests [25/90]
❌ SIO timing tests [0/8]
❌ Misc. edge case tests [3/10]
❌ Video tests
    ❌ Basic Mode 3
    ❌ Basic Mode 4
    👍 Degenerate OBJ transforms
    ❌ Layer toggle
    ❌ Layer toggle 2
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

❌ irq_demo (blinking text hblank irq)
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
Boots Harvest Moon Friends of Mineral Town
Boots Hello Kitty Happy Party Pals
Boots Kirby Nightmare in Dream Land
Boots Lord of The Rings Fellowship
Boots Lord of The Rings Two Towers
Boots Mario Kart Super Circuit
Boots Mega Man Zero
Boots Metroid Fusion
Boots Mother 12
Boots Mother 3
Boots Pokémon Mystery Dungeon Red Rescue Team
Boots Pokémon Firered
Boots Pokémon LeafGreen
Boots Pokémon Emerald
Boots Pokémon Ruby
Boots Pokémon Sapphire
Boots Sonic Advance
Boots Spyro Season of Ice
Boots Superstar Saga
Boots Super Dodge Ball Advance
Boots Super Mario Advance
      Tetris Worlds (Huff)
Boots The Minish Cap
Boots Ultimate Puzzle Games
Boots Warioware Twisted
Boots Wolfenstein 3D
Boots Doom
      Doom II
Boots Zelda Link to the Past
Boots Iridion II
Boots Iridion 3D
