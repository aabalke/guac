# <img width="36" height="36" alt="icon" src="https://github.com/user-attachments/assets/d86dfbbf-a12b-4cc5-843f-0efa84047eb9" /> guac: NDS, GBA, GBC, DMG Emulator

Guac is an Emulator written in golang for Gameboy, Gameboy Color, Gameboy
Advance, and Nintendo DS handheld consoles.

[Original Breakdown](https://youtu.be/BP_sMHJ99n0)
[NDS Update](https://youtu.be/AsWBItlGmZg)

![gb500](https://github.com/user-attachments/assets/e65c8cd3-c7c6-4ee4-9b8e-8ea3d1c5d5ea)![gba500](https://github.com/user-attachments/assets/bc770659-3f35-4c90-b295-9e0c994ad929)![nds500](https://github.com/user-attachments/assets/5c4c34d7-3665-4b84-94d7-8e56ee803fec)

# Installation / Building

See Releases for Windows and Linux precompiled binaries.

Building from source is possible with golang > 1.26.0, using:

```
go build .
```

# Getting Started

In both command line and console mode, save files are placed in the same directory
as the rom file (ex. "harvest_moon.gba", "harvest_moon.gba.save")

## Command line

Run the executable with a rom path (gb, gbc, gba, nds are required extensions) to
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
Watch trailing commas in json, it got me many a times.

```
[
 {
  "RomPath": "./rom/gba/the_minish_cap.gba",
  "ArtPath": "./art/the_minish_cap.png"
 }
 ...]
 ```

# Configuration

Emulator settings can be configured using the config.toml file at root.
If you would like to return to the default config.toml file, delete any 
present config.toml file and run the emulator.

## Configurable Options

### General
1. Keyboard / Controller Input
2. Backdrop Color
3. Menu Game Density
4. FPS Control

### GB / GBC
1. DMG Gameboy Palette

### GBA
1. Optimizations (Idle looping, sound clock updated)

### NDS
1. Jit Parameters
2. Bios and Firmware Options
3. Screen Layout, Sizing and Rotation
4. Real Time Clock offset
5. 3D Scene Export options

# Testing

Check the ./emu folder for individual consoles. These consoles will have
"testing.md" files showing currently passing tests and tested games.
