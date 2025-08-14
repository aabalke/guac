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

### RockPolish/rockwrestler

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
    👍 Vram Cnt
    👍 Tcm

👍 Initial State
    👍 Ipc/Irq/Cpsr (Some emulators have Q FLAG, others do not)
    👍 Cp15

### Personal Test to Build

IPC
- Need to test clearing bits of fifo
- Need more tests of irq, seems like more are needed

TCM
- Need to test disable and enabled bits
- ITCM load mode, let ITCM and DTCM overlap, let ITCM and other memory section overlap
- Out of bounds?

WRAM
- Does mode 3 clear?

VRAM
- Need checking of all MST/OFS combinations per bank, overlap testing

STRD
- rockwrestler and armwrestler have no test for STRD

DMA, Timer, Cartridge Tests?
