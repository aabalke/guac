# Generator for Arm Cpus

The Gameboy Advance and Nintendo DS both use ARM7TDMI 32bit RISC CPU, 33MHz (arm7)
and the nds additionally uses a ARM946E-S 32bit RISC CPU, 66MHz (arm9).

Both the arm7 and arm9 cpu have almost identical functionality. In order to keep
differences between each cpu minimal and changes accounted for, this generator
creates the arm7 adn arm9 code based on the template.


