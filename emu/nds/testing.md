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

Make sure zeros pc check is off

👍 Dma
👍 Irq
👍 Math
👍 Memory
    👍 Reading
    👍 Writing
    👍 Mirror Check
    👍 Bios Ram Usage
👍 Thumb
👍 Timer (prescaler values are roughly correct)

### Devkitpro examples

Audio

Card

Debugging

Wifi

FileSystem
    ❌ libfatdir
    👍 nitrodir
    ❌ overlays (matches nocash, libfat fails)


Graphics

    3D
        👍 Display List
        👍 Simple Quad
        👍 Simple Tri
        Textured Cube - Needs Capture
        👍 Textured Quad
        Ortho (I think needs light and toon, but no visible effect)
        nehe
            👍 Lesson 01
            👍 Lesson 02
            👍 Lesson 03
            👍 Lesson 04
            👍 Lesson 05
            Lesson 06
            Lesson 07
            Lesson 08
            Lesson 09
            Lesson 10
            Lesson 10b
            Lesson 11
        👍 Mixed Text 3D
        Palette Cube
            
            ❌ A3I5 Translucent Texture (alpha not setup)
            👍 4-Color Palette Texture     
            👍 16-Color Palette Texture    
            👍 256-Color Palette Texture   
            ❌ 4x4-Texel Compressed Texture
            ❌ A5I3 Translucent Texture (alpha not setup)
            👍 Direct Texture              
        👍 BoxTest



    capture

    Effects (windows match nocash but not melon, not sure on correct)
    👍 Ext_palettes
    gl2d
    grit
    Printing

    👍 Backgrounds
        👍 16 bit color bmp
        👍 256 bit color bmp
        👍 all in one
            👍 Basic
                👍 1, 2, 3, 4, 5
                👍 6, 7, 8, 9
                👍 10, 11, 12, 13

            👍 Bitmap
                👍 1, 2, 3, 4
                👍 5, 6
                👍 7, 8, 9, 10

            👍 Scrolling
            👍 Advanced
    
        👍 double buffer
        👍 rotation

    👍 Sprites
        👍 allocation test
        👍 animate simple
        👍 bitmap sprites
        👍 fire and sprites
        👍 simple
        👍 sprite extended palettes
        👍 sprite rotate

👍 hello_world

input
    addon
    keyboard
        async
        👍 stdin
    touch_pad
        touch_look
        👍 touch_test

👍 pxi

templates

time

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

# Games (Decrypted)

### Nonpanic Crash
Housemd
Zelda Spirit Tracks
Zelda Phantom Hourglass
Sonic Chronicles The Dark Brotherhood
Mario And Luigi Bowsers Inside Story
Pokemon Heartgold
Metroid Prime Hunters

### Booting (Buggyie)
Nintendogs Best Friends
Animal Crossing Wild World
Brain Age
Sonic Colors
Big Brain Academy
Tetris Ds
Cooking Mama
Brain Age 2
Pokemon Diamond
Pokemon Ranger
Mario Party
Mario and Luigi Partners in Time
Mario Kart Ds
Super Mario 64
New Super Mario Bros

### No Known errors
Yoshis Island
Pokemon Blue Rescue Team
