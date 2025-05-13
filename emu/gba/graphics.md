# Graphics

Regular Sprites

- Similar to bitmaps

1. Load Graphics and Palette
2. Set attr, size in OAM
3. Set Obj and mapping mode in DISPCNT

2 Types of 8x8 mini bitmaps:
4bpp 32 bytes long
8bpp 64 bytes long

                      lower        higher
OVRAM = Object VRAM = 0x0601_0000, 0x0605_0000

char(tile)block 0x4000 length (512 tiles)

in lower: always 32 byte offsets, if 8bpp use 0, 2, 4, 6 etc

Sprite Palettes = 0x500_0200

1D vs 2D Mapping:
2D          1D
1234----    12345678
5678----    90122345
9012----    --------
2345----    --------


