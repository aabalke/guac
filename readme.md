# <img width="24" height="24" alt="icon" src="https://github.com/user-attachments/assets/d86dfbbf-a12b-4cc5-843f-0efa84047eb9" /> guac: GBA, GBC, DMG Emulator

Guac is an Emulator written in golang for Gameboy, Gameboy Color and Gameboy
Advance handheld consoles.

![gb500](https://github.com/user-attachments/assets/e65c8cd3-c7c6-4ee4-9b8e-8ea3d1c5d5ea)![gba500](https://github.com/user-attachments/assets/bc770659-3f35-4c90-b295-9e0c994ad929)


# Installation / Building

See Releases for Windows and Linux precompiled binaries.

Building from source is possible with golang > 1.24.5, using:

```
go build .
```

# Getting Started

In both command line and console mode, save files are placed in the same directory
as the rom file (ex. "harvest_moon.gba", "harvest_moon.gba.save")

## Command line

Run the executable with a rom path (gb, gbc, and gba are required extensions) to
immediately enter the game.

```
.\guac -r="../rom/pokemon_emerald.gba"
```

## Console Mode

Run the executable without flags to use console mode, which initalizes a Game
Selection Screen.

```
.\guac
```

### Setting up Console Mode

At root, create a "roms.json" file. This file will hold the game metadata in the
following format. At this time Art must be 1:1 pngs or jpgs.

```
[
 {
  "RomPath": "./rom/gba/the_minish_cap.gba",
  "ArtPath": "./art/the_minish_cap.png",
 },
 ...]
 ```

# Configuration

Emulator settings can be configured using the config.toml file at root.
If you would like to return to the default config.toml file, delete any 
present config.toml file and run the emulator.

## Configurable Options

1. Keyboard / Controller Input
2. Backdrop Color
3. Original DMG Gameboy Palette
4. Menu Game Density
5. Experimental Performance Improvements (HLE Bios, Multithreading etc)

# Testing

Check the ./emu folder for individual consoles. These consoles will have
"testing.md" files showing currently passing tests and tested games.

# Contributing

Please contribute! At this time I am mostly interested in getting game errors
fixed. Cycle accuracy, and serial communication are not a priority at this
time. At this time, AI contributions will be rejected. For what needs work see
"todo.md" pages for each console.
