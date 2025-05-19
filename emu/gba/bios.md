BIOS Functions

GBA Basic Functions
👍 0x00 SoftReset
👍 0x01 RegisterRamReset
👍 0x02 Halt
❌ 0x03 Stop/Sleep
➗ 0x04 IntrWait
➗ 0x05 VBlankIntrWait
👍 0x06 Div
👍 0x07 DivArm
👍 0x08 Sqrt
👍 0x09 ArcTan
👍 0x0A ArcTan2
👍 0x0B CpuSet
👍 0x0C CpuFastSet
👍 0x0D GetBiosChecksum
👍 0x0E BgAffineSet
👍 0x0F ObjAffineSet

GBA Decompression Functions
➗ 0x10 BitUnPack
👍 0x11 LZ77UnCompReadNormalWrite8bit   ;"Wram"
👍 0x12 LZ77UnCompReadNormalWrite16bit  ;"Vram"
❌ 0x13 HuffUnCompReadNormal
👍 0x14 RLUnCompReadNormalWrite8bit     ;"Wram"
👍 0x15 RLUnCompReadNormalWrite16bit    ;"Vram"
❌ 0x16 Diff8bitUnFilterWrite8bit       ;"Wram"
❌ 0x17 Diff8bitUnFilterWrite16bit      ;"Vram"
❌ 0x18 Diff16bitUnFilter
 
GBA Sound (and Multiboot/HardReset/CustomHalt)
❌ 0x19 SoundBias
❌ 0x1A SoundDriverInit
❌ 0x1B SoundDriverMode
❌ 0x1C SoundDriverMain
❌ 0x1D SoundDriverVSync
❌ 0x1E SoundChannelClear
❌ 0x1F MidiKey2Freq
❌ 0x20 SoundWhatever0
❌ 0x21 SoundWhatever1
❌ 0x22 SoundWhatever2
❌ 0x23 SoundWhatever3
❌ 0x24 SoundWhatever4
❌ 0x25 MultiBoot
❌ 0x26 HardReset
❌ 0x27 CustomHalt
❌ 0x28 SoundDriverVSyncOff
❌ 0x29 SoundDriverVSyncOn
❌ 0x2A SoundGetJumpList
