# Graphical

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
