# Testing

❌ RockPolish/rockwrestler fails Memory / Tcm 0x12 (MelonDS + no$gba fails 0x10)
👍 Atem2069/armwrestler-fixed
👍 arm7wrestler
👍 Imran Nazar & LiraNuna / TinyFB
👍 shonumi/hello world
👍 shonumi/gbe-plus-nds-tests

### Devkitpro examples

audio/maxmod/audio_modes
audio/maxmod/basic_sound
audio/maxmod/reverb
audio/maxmod/song_events_example
audio/maxmod/song_events_example2
audio/maxmod/streaming
audio/micrecord

❌ card/eeprom             - crash
❌ debugging/exceptiontest - nothing

❌ dswifi/ap_search   - not supported
❌ dswifi/autoconnect - not supported
❌ dswifi/httpget     - not supported

filesystem/libfatdir
filesystem/nitrodir
filesystem/overlays

graphics/3d/3d_both_screens
graphics/3d/boxtest
graphics/3d/display_list
graphics/3d/display_list_2
graphics/3d/env_mapping
graphics/3d/mixed_text
graphics/3d/nehe/lesson01
graphics/3d/nehe/lesson02
graphics/3d/nehe/lesson03
graphics/3d/nehe/lesson04
graphics/3d/nehe/lesson05
graphics/3d/nehe/lesson06
graphics/3d/nehe/lesson07
graphics/3d/nehe/lesson08
graphics/3d/nehe/lesson09
graphics/3d/nehe/lesson10
graphics/3d/nehe/lesson10b
graphics/3d/nehe/lesson11
graphics/3d/ortho
graphics/3d/paletted_cube
graphics/3d/picking
graphics/3d/simple_quad
graphics/3d/simple_tri
graphics/3d/textured_cube
graphics/3d/textured_quad
graphics/3d/toon_shading

👍 graphics/gl2d/2dplus3d
❌ graphics/gl2d/dual_screen - need 3d line support
❌ graphics/gl2d/fonts
    3d vector blending needs to mix all not just opposite. This is because quads
    are split into triangles and interpolated only by 3 vertices
❌ graphics/gl2d/primitives  - need 3d line support
❌ graphics/gl2d/scrolling   - with jit have problems
❌ graphics/gl2d/sprites     - spinning sprites not blended properly

👍 graphics/ext_palettes
👍 graphics/grit
👍 graphics/effects
❌ graphics/capture - direct bitmap fails

👍 graphics/backgrounds/16bitcolormap
👍 graphics/backgrounds/256bitcolormap
👍 graphics/backgrounds/double_buffer
👍 graphics/backgrounds/rotation

👍 graphics/backgrounds/all_in_one/basic
👍 graphics/backgrounds/all_in_one/bitmap
👍 graphics/backgrounds/all_in_one/scrolling
❌ graphics/backgrounds/all_in_one/advanced - x mosaic on tiled not working

👍 graphics/sprites/allocation_test
👍 graphics/sprites/animate_simple
❌ graphics/sprites/bitmap_sprites - direct bitmap fails
👍 graphics/sprites/fire_and_sprites
👍 graphics/sprites/simple
👍 graphics/sprites/sprite_extended_palettes
👍 graphics/sprites/sprite_rotate

👍 hello_world
❌ input/addon - not supported
👍 input/keyboard/async
👍 input/keyboard/stdin
👍 input/touch_pad/touch_look
👍 input/touch_pad/touch_test
👍 pxi/pxi
👍 time/realtimeclock
👍 time/stopwatch
👍 time/timercallback

# Games (Decrypted)
Animal Crossing
Chrono Trigger
Mario Kart
- Flickering Textures
- Shadows
Pokemon Diamond
- Large 3d meshes
Pokemon Mystery Dungeon Explorers of Time
Sonic Chronicles the Dark Brotherhood
- Edges on transparent (same as zelda?)
Super Mario 64
👍 Tetris DS
Yoshi's Island
- minor decompressed sound problem

# 0.0.2 Release General Goals
- Culling
- Faster Frame rate sound
- 3d "2d" texture bug - bad timing bug
- ADMCP decompress error
- Bug on startup - this goes back very far (pre 0.0.1)
- Screenshotting
- Save States

Mario and Luigi Bowsers Inside Story
- "2d" 3d textures getting corrupted
