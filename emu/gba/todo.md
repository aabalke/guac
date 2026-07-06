# Graphical

0. Bitmap modes need to be treated as Background 2, for effects, object priority etc

1. Mode 4: fix objects, and blinking
- Doom II, Need for Speed Underground

2. Affine Object cut out

3. NES Game Support

4. RTC

# Audio

1. Need to replace Oto version 1.0.1 with Ebitengine built-in audio handler.
This is problematic because oto uses a writer which handles over and under runs
in a way, but I cannot get ebitengine to do the same.

2. Need to fix pitch and volume of the analog channels, particularly volume of
WAVE and pitch of NOISE.

# Things that are not planned

- Serial Comms
- Other peripherals

# todo

- rtc options
- nes game support
- gb / gbc on gba


	// if c.Reg.R[15] == 0x8022B08 {
	// if V[15] >= 235312 {
	//	B[15] = c.Reg.R[15] == 0x8022B08
	fmt.Printf("OP %08X %08X PC %08X SCH %08X\n",
		c.Op[0], c.Op[1], c.Reg.R[15], c.gba.Scheduler.Now())
	//}

	if V[15] > 300_000 {
		// if V[15] > 235315 {
		os.Exit(0)
	}

	V[15]++



gba pokemon emerald fps 
nano 310fps
