# Testing


### Atem2069/armwrestler-fixed

👍 Arm Alu
👍 Arm Ldr/Str
👍 Arm Lsm/Stm
👍 Thumb Alu
👍 Thumb Ldr/Str
👍 Thumb Lsm/Stm

👍 Arm v5TE
    👍 CLZ
    👍 LDRD
    👍 MRC
    👍 QADD
    👍 SMLABB
    👍 SMLABT
    👍 SMLATB
    👍 SMLATT

### RockPolish/rockwrestler - custom build

included:


👍 Armv4
    👍 Condition Codes

👍 Armv5
    👍 CLZ
    👍 QADD, QSUB
    👍 QDADD, QDSUB
    👍 SMULxy
    👍 SMLAxy
    👍 SMULWy
    👍 SMLAWy
    👍 SMLALxy
    👍 BLX
    👍 PC SPEC
    👍 LDM / STM (16)

👍 Ipc
    👍 Ipcsync
    👍 Ipcfifo
    👍 Ipcfifo irq

👍 Ds math
    👍 Sqrt 32
    👍 Sqrt 64
    👍 Div 32/32
    👍 Div 64/32
    👍 Div 64/64

👍 Memory
    👍 Wram Cnt
        Includes: Check to make sure mode 3 does not clear arm9, just hides
        Includes: Check to make sure mode 3 does not clear arm7, mirror accessible

    👍 Vram Cnt
    👍 Tcm

👍 Initial State
    👍 Ipc/Irq/Cpsr (Some emulators have Q FLAG, others do not)
    👍 Cp15

### Imran Nazar & LiraNuna / TinyFB
👍 TinyFB (removes all fluff)

### shonumi / hello world
👍 hello world (identical outcome to TinyFB but with additional init functionality)

### gbe-plus

These tests do not work on any emulator - as far as I can tell. Mostly due to changes in devkitPro.
(Tested on gbe+, desmume, melon, nocash)

❌ Dma
    ❌ Transfer Test
    👍 Dma Fill Test
👍 Irq
👍 Math
❌ Memory
    ❌ Reading (Just 16bit read)
    👍 Writing
    ❌ Shared Wram Check (Fails)
    ❌ Mirror Check (Just Main Memory)
    ❌ Bios Ram Usage
👍 Thumb
❌ Timer

### Devkitpro examples

Graphics

Backgrounds
    👍 16 bit color bmp
    👍 256 bit color bmp
    ❌ all in one
        ❌ Basic
            👍 1, 2, 3, 4, 5
            ❌ 6, 7, 8, 9
            ❌ 10, 11, 12, 13

    👍 double buffer
    ❌ rotation

### Personal Test to Build

IPC
- Need to test clearing bits of sync
- Need more tests of irq, seems like more are needed
- Need test to check that bools are for correct fifos, and not reversed (irqnotempty, full, empty, irq empty etc etc etc)

TCM
- Need to test disable and enabled bits
- ITCM load mode, let ITCM and DTCM overlap, let ITCM and other memory section overlap
- need to test dtcm and main memory overlap
- Out of bounds?

VRAM
- Need checking of all MST/OFS combinations per bank, overlap testing

STRD
- rockwrestler and armwrestler have no test for STRD

DMA, Timer, Cartridge Tests?

SDT PLD (Cache Prepare for Load opcode)

Need tests for 
NOT ALIGNING PC (movs 15, 14 etc)
BLX r15, r14 (MUST BE + 3, for return thumb setting .BLX, .... BX back will need thumb setting)
BLX ARM needs to occur before Cond
