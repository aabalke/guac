# 3D Video Implimentation Status

The following are outstanding errors or unimplimented features

All matrix cmds, light params, and vertex cmds are implimented.

## General

Alpha errors
Capture with 3d has problems.
Line Segments not implimented

## Display Control

Disp3DCnt Partial
Viewport Partial
1DotDepth - added, makes no difference I believe
AlphaTest

## Polygon Attributes

Polygon Back Surface, Front Surface - Culling
Depth-value for Translucent Pixels
Far-plane intersecting polygons
Depth Test, Draw Pixels with Depth
Polygon ID
View Volume Clipping is not implimented.

## Shadow Polygons

Shadows are unimplimented

## Textures

Coords, params, blends, and formats are all implimented.
Texture Coordinates Transformation Mode 3 - Vertex source is untested.

## Toon, Edge, Fog, Alpha-Blending, Anti-Aliasing

Alpha-Blending incorrect
AntiAliasing not implimented
Toon, Fog, and Edge implimented

## Status

GXFifo is partially implimented.

## Tests

All tests are implimented.

## Rear-Plane

Should be all good. Cache may not update properly.

## 3D Final Output
Scrolling - need to fix region size
priority - i believe working
special effects - alpha blending incorrect
window freature - need to force alpha blending above
