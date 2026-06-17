# Testing

These tests require manual entry and testing.
This means they may be out of date from the most recent commit.

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

# alloncm / MagenTests v5
❌ ColorBgOamPriority
❌ ColorOamInternalPriority
👍 Vram DMA HBlank mode
👍 KEY0 (CPU mode register) Lock After Boot
👍 STAT register PPU mode upon PPU disabled
👍 MBC out of bounds RAM access
    👍 MBC1 
    👍 MBC3 
    👍 MBC5 

# Other tests

👍 EricKirschenmann/MBC3-Tester-gb

# mattcurrie
👍 dmg-acid2
👍 gbc-acid2
❌ gbc-acid-hell

# aaaaaa123456789/rtc3test
👍 basic
❌ range
❌ sub second

# Ashiepaws/scribbltests
❌ fairylake
👍 lycscx
👍 lycscy
👍 palettely
👍 scxly
❌ statcount // need lcd enabled timing accurate
❌ winpos

# LIJI32/SameSuite

## apu
## dma
👍 gbc_dma_cont
👍 gdma_addr_mask
👍 hdma_lcd_off
👍 hdma_mode0

## interrupt
## ppu
## sgb
