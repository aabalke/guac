# Generator for Arm Cpus

The Gameboy and Gameboy Color consoles both have a large overlap in graphics
processing. 2 versions of the graphcis pipeline is used so the DMG Gameboy does
not have to complete GBC conditionals; however, since they are so similar a
generator is used to create the functions, base on the template.
