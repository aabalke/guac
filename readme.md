# <img width="36" height="36" alt="icon" src="https://github.com/user-attachments/assets/d86dfbbf-a12b-4cc5-843f-0efa84047eb9" /> guac: NDS, GBA, GBC, DMG Emulator

Guac is an Emulator written in golang for Gameboy, Gameboy Color, Gameboy Advance, and Nintendo DS handheld consoles.

![gb500](https://github.com/user-attachments/assets/e65c8cd3-c7c6-4ee4-9b8e-8ea3d1c5d5ea)![gba500](https://github.com/user-attachments/assets/bc770659-3f35-4c90-b295-9e0c994ad929)![nds500](https://github.com/user-attachments/assets/5c4c34d7-3665-4b84-94d7-8e56ee803fec)

# Update Videos
[UI Update (v0.0.3)](https://www.youtube.com/watch?v=dVdIM_bPQrY)
[NDS Update (v0.0.2)](https://youtu.be/AsWBItlGmZg)
[Original (v0.0.1)](https://youtu.be/BP_sMHJ99n0)

# Features

## Emulation out of the box
guac does not require any bios or firmware files (but you can provide!), just the roms you are interested in using. It also has full controller support even in menus, with on screen keyboards.

## Performance
Nintendo DS emulation on guac comes with a one-of-a-kind jit compiler. This allows cpu emulation to be 2-5x faster depending on workloads. This is the first jit compiler that can interop with (more like confuse) the golang runtime to allow calls to go functions from jit code. Additionally, this is the first arm64 jit compiler written in golang. For more information please see [gojit](https://github.com/aabalke/gojit/).

## Customizable
guac comes with a TON of customizable options. The ui has customizable colors, hotkeys, and localization in english or spanish. All consoles have customizable inputs. Nintendo DS has customizable screen layouts, Real-Time Clock, and Firmware.

## 3D Export
guac comes with a 3D Scene Export, allowing the current 3d scene to be exported as a .obj file. This allow the 3D scene to be imported into 3D software such as Blender, Maya etc for debugging, artistic expression and more.

# Installation / Building

See Releases for Windows and Linux precompiled binaries.

Building from source is possible with golang > 1.26.2, using:

```
go build .
```

For specific system build quirks please see [guacemulator.com](https://guacemulator.com/).


# Getting Started

Run the executable without any flags or configuration.

```
.\guac
```

If you are interested in specific flags and configuration options please see [guacemulator.com](https://guacemulator.com/).
