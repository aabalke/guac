# 🥑 Guac GBA/GBA/DMG Emulator

Guac is an Emulator written in golang for Gameboy, Gameboy Color and Gameboy
Advance handheld consoles.

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
the game metadata in the following format. At this time Art **MUST** be pngs.

```
[
 {
  "Name": "The Minish Cap",
  "RomPath": "./rom/gba/the_minish_cap.gba",
  "ArtPath": "./art/the_minish_cap.png",
  "Year": 2005,
  "Console": "Gameboy Advance"
 },
 ...]
 ```

# Installation / Building

Please use the releases for precompiled binaries.

```
go build -ldflags="-H=windowsgui"
```


# Keybindings / Controller Support

## Emulator

| Key   | DualSense | Binding           |
|-------|-----------|-------------------|
| Enter |           |                   |
| P     |           |                   |
| M     |           |                   |
| F11   |           | Toggle Fullscreen |

## Gameboy / Gameboy Color

| Key   | DualSense | Binding |
|-------|-----------|---------|
| Enter |           |         |
| P     |           |         |
| M     |           |         |

## Gameboy Advance

| Key   | DualSense | Binding |
|-------|-----------|---------|
| J     |           |         |
| K     |           |         |
| L     |           |         |
| ;     |           |         |

# Testing

Check the ./emu folder for individual consoles. These consoles will have
"testing.md" files showing currently passing tests and tested games.

# Developers

If you are interested: [aaronbalke.com](some deep dives regarding emulation).
