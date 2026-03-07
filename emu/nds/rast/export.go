package rast

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"time"
	"unsafe"

	"github.com/aabalke/guac/emu/nds/rast/gl"
)

const (
    FORMAT_OBJ = iota
    FORMAT_GLTF
)

type Export struct {
    Format      int
    Directory    string
    Rasterizer  *Rasterizer
    ShadowPolys bool

    obj string
    mtl string

    face      int
    mesh      int
    material  int

    usedTextures map[uintptr]int
}

func NewExport(dir string, format int, shadowPolys bool, rast *Rasterizer) *Export {
    return &Export{
        Directory:   dir,
        Format:      format,
        ShadowPolys: shadowPolys,
        Rasterizer:  rast,
    }
}

func (e *Export) Export() {
    fmt.Printf("Preparing Scene\n")

    polys := e.Rasterizer.Buffers.GetBuffer().Polys

    switch e.Format {
    case FORMAT_GLTF:
        panic("unsupported export type gltf")
    default:
        e.ExportObj(polys)
    }

    fmt.Printf("Exported Scene\n")
}

func (e *Export) ExportObj(polys []Polygon) {

    if err := os.MkdirAll(e.Directory, 0755); err != nil {
        panic(err)
    }

    e.obj = fmt.Sprintf("# Guac Emulator Export. %v\n", time.Now())
    e.obj += fmt.Sprintf("mtllib guac.mtl\n")

    e.mtl = fmt.Sprintf("# Guac Emulator Export. %v\n", time.Now())
    e.face = 1
    e.mesh = 1
    e.material = 1
    e.usedTextures = make(map[uintptr]int)

    for i := range len(polys) {
        e.exportPoly(&polys[i])
    }

    if ok := writeFile(e.Directory + "guac.obj", e.obj); !ok {
        panic("Failed to write export scene file obj")
    }

    if ok := writeFile(e.Directory + "guac.mtl", e.mtl); !ok {
        panic("Failed to write export scene file mtl")
    }
}

func (e *Export) exportPoly(p *Polygon) {

	if len(p.Vertices) == 0 {
        return
	} else if shadow := p.Mode == 3; shadow && !e.ShadowPolys {
        return
    }

    e.exportMesh()

	switch p.PrimitiveType {
	case PRIM_SEP_TRI:

		for i := 0; i < len(p.Vertices); i += 3 {
            w := float32(p.Vertices[i].NdsTexture.Width)
            h := float32(p.Vertices[i].NdsTexture.Height)
            e.exportVertex(p.Vertices[i+2], w, h)
            e.exportVertex(p.Vertices[i+1], w, h)
            e.exportVertex(p.Vertices[i+0], w, h)
            e.exportTexture(p.Vertices[i])
            e.exportFace(3)
		}

	case PRIM_SEP_QUAD:

		for i := 0; i < len(p.Vertices); i += 4 {
            w := float32(p.Vertices[i].NdsTexture.Width)
            h := float32(p.Vertices[i].NdsTexture.Height)
            e.exportVertex(p.Vertices[i+3], w, h)
            e.exportVertex(p.Vertices[i+2], w, h)
            e.exportVertex(p.Vertices[i+1], w, h)
            e.exportVertex(p.Vertices[i+0], w, h)
            e.exportTexture(p.Vertices[i])
            e.exportFace(4)
		}

	case PRIM_TRI_STRIP:

		for i := 2; i < len(p.Vertices); i++ {
            w := float32(p.Vertices[i].NdsTexture.Width)
            h := float32(p.Vertices[i].NdsTexture.Height)

			if clockwise := i&1 == 1; clockwise {
                e.exportVertex(p.Vertices[i-2], w, h)
                e.exportVertex(p.Vertices[i-1], w, h)
                e.exportVertex(p.Vertices[i-0], w, h)
                e.exportTexture(p.Vertices[i])
                e.exportFace(3)
				continue
			}

            e.exportVertex(p.Vertices[i-0], w, h)
            e.exportVertex(p.Vertices[i-1], w, h)
            e.exportVertex(p.Vertices[i-2], w, h)
            e.exportTexture(p.Vertices[i])
            e.exportFace(3)

		}

	case PRIM_QUAD_STRIP:

		for i := 2; i+1 < len(p.Vertices); i += 2 {

            w := float32(p.Vertices[i].NdsTexture.Width)
            h := float32(p.Vertices[i].NdsTexture.Height)

            e.exportVertex(p.Vertices[i+0], w, h)
            e.exportVertex(p.Vertices[i+1], w, h)
            e.exportVertex(p.Vertices[i-1], w, h)
            e.exportVertex(p.Vertices[i-2], w, h)
            e.exportTexture(p.Vertices[i])
            e.exportFace(4)
		}
	}
}

func (e *Export) exportVertex(vertex gl.Vertex, w, h float32) {

    e.obj += "v "
    e.obj += fmt.Sprintf("%f ", vertex.WorldPosition.X)
    e.obj += fmt.Sprintf("%f ", vertex.WorldPosition.Y)
    e.obj += fmt.Sprintf("%f ", vertex.WorldPosition.Z)
    e.obj += fmt.Sprintf("%f ", vertex.Color.R)
    e.obj += fmt.Sprintf("%f ", vertex.Color.G)
    e.obj += fmt.Sprintf("%f ", vertex.Color.B)

    if vertex.NdsTexture == nil {
        e.obj += "vt 0 0\n"
        return
    }

    e.obj += "\nvt "
    e.obj += fmt.Sprintf("%f ", vertex.S / w)
    e.obj += fmt.Sprintf("%f ", vertex.T / h)
    e.obj += "\n"
}

func (e *Export) exportTexture(vertex gl.Vertex) {

    texture := vertex.NdsTexture

    if texture == nil {
        return
    }

    key := uintptr(unsafe.Pointer(texture.CachedTexture))

    if material, ok := e.usedTextures[key]; ok {
        e.obj += fmt.Sprintf("usemtl Material%d\n", material)
        return
    }

    e.obj += fmt.Sprintf("usemtl Material%d\n", e.material)

    e.mtl += fmt.Sprintf("newmtl Material%d\n", e.material)
    e.mtl += fmt.Sprintf("Kd 1 1 1\n")
    e.mtl += fmt.Sprintf("map_Kd texture%05d.png\n", e.material)

    SavePNG(
        fmt.Sprintf(e.Directory + "texture%05d.png", e.material),
        texture.Width,
        texture.Height,
        texture.CachedTexture,
    )

    e.usedTextures[key] = e.material
    e.material++
}

func (e *Export) exportFace(cnt int) {
    e.obj += "f "

    for range cnt {
        e.obj += fmt.Sprintf("%d/%d ", e.face, e.face)
        e.face++
    }

    e.obj += "\n"
}

func (e *Export) exportMesh() {
    e.obj += fmt.Sprintf("o mesh_%d\n", e.mesh)
    e.mesh++
}

func writeFile(path, s string) bool {

	f, err := os.Create(path)
	if err != nil {
		return false
	}
	defer f.Close()

	writer := bufio.NewWriter(f)
    _, err = writer.Write([]byte(s))
	if err != nil {
		return false
	}

    writer.Flush()

	return true
}

func floatToUint8(v float32) uint8 {
	v = max(0, min(1, v))
	return uint8(math.Round(float64(v * 255)))
}

func SavePNG(filename string, width, height int, pixels *[]gl.Color) error {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := range height {
		for x := range width {
			i := y*width + x
			c := (*pixels)[i]

			img.SetRGBA(x, y, color.RGBA{
				R: floatToUint8(c.R),
				G: floatToUint8(c.G),
				B: floatToUint8(c.B),
				A: floatToUint8(c.A),
			})
		}
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return png.Encode(f, img)
}
