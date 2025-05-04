package gba

const (
	SYS_SoftReset                      = 0x00
	SYS_RegisterRamReset               = 0x01
	SYS_Halt                           = 0x02
	SYS_StopSleep                      = 0x03
	SYS_IntrWait                       = 0x04
	SYS_VBlankIntrWait                 = 0x05
	SYS_Div                            = 0x06
	SYS_DivArm                         = 0x07
	SYS_Sqrt                           = 0x08
	SYS_ArcTan                         = 0x09
	SYS_ArcTan2                        = 0x0A
	SYS_CpuSet                         = 0x0B
	SYS_CpuFastSet                     = 0x0C
	SYS_GetBiosChecksum                = 0x0D
	SYS_BgAffineSet                    = 0x0E
	SYS_ObjAffineSet                   = 0x0F
	SYS_BitUnPack                      = 0x10
	SYS_LZ77UnCompReadNormalWrite8bit  = 0x11
	SYS_LZ77UnCompReadNormalWrite16bit = 0x12
	SYS_HuffUnCompReadNormal           = 0x13
	SYS_RLUnCompReadNormalWrite8bit    = 0x14
	SYS_RLUnCompReadNormalWrite16bit   = 0x15
	SYS_Diff8bitUnFilterWrite8bit      = 0x16
	SYS_Diff8bitUnFilterWrite16bit     = 0x17
	SYS_Diff16bitUnFilter              = 0x18
	SYS_SoundBias                      = 0x19
	SYS_SoundDriverInit                = 0x1A
	SYS_SoundDriverMode                = 0x1B
	SYS_SoundDriverMain                = 0x1C
	SYS_SoundDriverVSync               = 0x1D
	SYS_SoundChannelClear              = 0x1E
	SYS_MidiKey2Freq                   = 0x1F
	SYS_SoundWhatever0                 = 0x20
	SYS_SoundWhatever1                 = 0x21
	SYS_SoundWhatever2                 = 0x22
	SYS_SoundWhatever3                 = 0x23
	SYS_SoundWhatever4                 = 0x24
	SYS_MultiBoot                      = 0x25
	SYS_HardReset                      = 0x26
	SYS_CustomHalt                     = 0x27
	SYS_SoundDriverVSyncOff            = 0x28
	SYS_SoundDriverVSyncOn             = 0x29
	SYS_SoundGetJumpList               = 0x2A
)
