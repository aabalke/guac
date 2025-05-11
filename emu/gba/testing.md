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
❌ shades
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
 
❌ deadbody Cpu Test
