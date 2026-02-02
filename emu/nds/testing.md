# Testing

❌ RockPolish/rockwrestler fails Memory / Tcm 0x12 (MelonDS + no$gba fails 0x10)
👍 Atem2069/armwrestler-fixed
👍 arm7wrestler
👍 Imran Nazar & LiraNuna / TinyFB
👍 shonumi/hello world
👍 shonumi/gbe-plus-nds-tests

### Devkitpro examples

Audio

Card

Debugging

Wifi

FileSystem
    ❌ libfatdir
    👍 nitrodir
    ❌ overlays (matches nocash, libfat fails)

Graphics

    3D
        👍 Both 3D
        👍 Display List
        👍 Simple Quad
        👍 Simple Tri
        Textured Cube - Needs Capture
        👍 Textured Quad
        Ortho (I think needs light and toon, but no visible effect)
        nehe
            👍 Lesson 01
            👍 Lesson 02
            👍 Lesson 03
            👍 Lesson 04
            👍 Lesson 05
            Lesson 06
            Lesson 07
            Lesson 08
            Lesson 09
            Lesson 10
            Lesson 10b
            Lesson 11
        👍 Mixed Text 3D
        👍 Palette Cube
       
       👍 BoxTest
    capture

    Effects (windows match nocash but not melon, not sure on correct)
    👍 Ext_palettes
    gl2d
    👍 grit

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
 graphics/sprites/fire_and_sprites
 graphics/sprites/simple
 graphics/sprites/sprite_extended_palettes
 graphics/sprites/sprite_rotate

👍 hello_world
❌ input/addon
👍 input/keyboard/async
👍 input/keyboard/stdin
👍 input/touch_pad/touch_look
👍 input/touch_pad/touch_test
👍 pxi/pxi
👍 time/realtimeclock
👍 time/stopwatch
❌ time/timercallback

# Games (Decrypted)

### 3D Texture Flashing
Mario And Luigi Bowsers Inside Story
Pokemon Ranger

### 3D Alpha Blend
New Super Mario Bros (also error with Affine16, pa,pc bgs, and same z error)

### 3D Same Z Error
Cooking Mama
Pokemon Diamond (Save Error)

### Misc 3D
Pokemon Heartgold
Metroid Prime Hunters
Nintendogs Best Friends
Zelda Spirit Tracks
Zelda Phantom Hourglass

### No known errors
Animal Crossing Wild World
Brain Age
Brain Age 2
Big Brain Academy
Tetris Ds
Yoshis Island
Pokemon Blue Rescue Team
Mario Kart Ds
Sonic Chronicles The Dark Brotherhood

### Other
Clubhouse Games
Housemd
Lego Star Wars the Complete Saga
Mario Party
Mario and Luigi Partners in Time
Sonic Colors
Super Mario 64

Fog needs work - see lesson 10 (seems good for mario kart, probably DepthW vs DepthZ related)
