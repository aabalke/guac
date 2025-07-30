# 🥑 Guac GBA/GBC/DMG Emulator

Guac is an Emulator written in golang for Gameboy, Gameboy Color and Gameboy
Advance handheld consoles.

# Installation / Building

Releases precompiled binaries are available. However, you can build using:

```
go build -ldflags="-H=windowsgui"
```

# Getting Started

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

In the same directory as ./guac, create a "roms.json" file. This file will hold
the game metadata in the following format. At this time Art must be 1:1 pngs or jpgs.

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

# Testing

Check the ./emu folder for individual consoles. These consoles will have
"testing.md" files showing currently passing tests and tested games.

# Developers

If you are interested: [aaronbalke.com](some deep dives regarding emulation).
