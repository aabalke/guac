# Testing

# Blargg Tests

👍 Cpu Instructions
👍 Instruction Timings
👍 Interrupt Timings
👍 Memory Timings
👍 Memory Timings 2
❌ DMG Sound 2
    👍 00
    👍 01
    👍 02
    👍 03
    👍 04
    ❌ 05
    👍 06
    👍 07
    ❌ 08 - 21 not 22, otherwise good
    ❌ 09
    ❌ 10
    👍 11 - regs after power
        note: retrio/gb-test-roms version states NR41 shouldn't be affected, but
        aquach/gameboy says should. Bgb and sameboy both have N41 unaffected. So 
        we will use not affected version.
    ❌ 12
❌ CGB Sound 2

Oam Bug 2
    👍 Lcd Sync
    ❌ Causes
    👍 Non Causes
    ❌ Scanline Timing
    ❌ Timing Bug
    👍 Timing No Bug
    ❌ Timing Effect
    ❌ Inst Timing

# Mooneye Acceptance Test

## General

❌ add sp e timing
❌ boot div dmg0
❌ boot div dmgABCmgb
❌ boot div S
❌ boot div2 S
❌ boot hwio dmg0
❌ boot hwio dmgABCmgb
❌ boot hwio S
❌ boot regs dmg0
❌ boot regs dmgABC
❌ boot regs mgb
❌ boot regs sgb
❌ boot regs sgb2
❌ call timing
❌ call timing2
❌ call cc_timing
❌ call cc_timing2
❌ di timing GS
👍 div timing
❌ ei sequence
👍 ei timing
👍 halt ime0 ei
❌ halt ime0 nointr_timing
👍 halt ime1 timing
❌ halt ime1 timing2 GS
❌ if ie registers
👍 inst timings
❌ jp timing
❌ jp cc timing
❌ ld hl sp e timing
❌ oam dma_restart
❌ oam dma start
❌ oam dma timing
❌ pop timing
❌ push timing
👍 rapid di ei
❌ ret timing
❌ ret cc timing
❌ reti timing
👍 reti intr timing
❌ rst timing

## Bits

👍 mem oam
👍 reg f
❌ unused_hwio_GS

## Timer

👍 div write
❌ rapid toggle
❌ tim00 div trigger
👍 tim00
❌ tim01 div trigger
👍 tim01 
❌ tim10 div trigger
👍 tim10
❌ tim11 div trigger
👍 tim11
❌ tima reload
❌ tima write reloading
❌ tma write reloading

## Misc Tests

👍 Manual Sprite Priority
👍 Daa Instruction
❌ Interrupt Handling ie Push
❌ Serial boot sclk align dmgABCmgb
 
## OAM DMA
 
👍  basic
❌ reg_read
❌ sources GS
 
## PPU
 
❌ hblank ly scx timing GS
❌ intr 1 2 timing GS
❌ intr 2 0 timing
❌ intr 2 mode0 timing
❌ intr 2 mode3 timing
❌ intr 2 oam ok timing
❌ intr 2 mode0 timing sprites
❌ lcdon timing GS
❌ lcdon write timing GS
❌ stat irq blocking
❌ stat lyc onoff
❌ vblank stat intr GS

## Emulator Only

### MBC1

👍 bits bank1
👍 bits bank2
❌ bits mode
👍 bits ramg
❌ rom 512kb
❌ rom 1Mb
❌ rom 2Mb
❌ rom 4Mb
❌ rom 8Mb
❌ rom 16Mb
❌ ram 64kb
❌ ram 256kb
❌ multicart rom 8Mb

### MBC2

### MBC3

## Misc Tests
