package nds

import "github.com/hajimehoshi/ebiten/v2"

// For emulated nds, see ppu. Screen is just how it is displayed with the emulator

//layout = "vertical" # valid "vertical", "horizontal", "Hybrid"
//rotation = 0 # valid 0, 90, 180, 270
//gap = 0 # valid 0,
//sizing = "even" # valid "even", "emphasize top", "emphasize bottom", "only top", "only bottom"

const (
    LAYOUT_VERTICAL = iota
    LAYOUT_HORZONTAL
    LAYOUT_HYBRID
)

const (
    SIZING_EVEN = iota
    SIZING_EMP_TOP
    SIZING_EMP_BOT
    SIZING_ONLY_TOP
    SIZING_ONLY_BOTTOM
)

type Screen struct {
    Layout   int
    Sizing   int
    Gay      int
    Rotation int

	Top, Bottom *ebiten.Image
    BtmAbs BtmAbs
}

type BtmAbs struct {
    T, B, L, R, W, H int
}

func NewScreen() *Screen {

    s := &Screen{
        Top: ebiten.NewImage(SCREEN_WIDTH, SCREEN_HEIGHT),
        Bottom: ebiten.NewImage(SCREEN_WIDTH, SCREEN_HEIGHT),
    }

    return s
}

func (s *Screen) FillScreen(screen *ebiten.Image) {

    setHorizontal := false

    var (
        sw = float64(screen.Bounds().Dx())
        sh = float64(screen.Bounds().Dy())
        iw = float64(s.Top.Bounds().Dx())
        ih = float64(s.Top.Bounds().Dy())

        scaleX, scaleY float64
        offsetX, offsetY float64
    )
    if setHorizontal {
        scaleX = (sw / 2) / iw
        scaleY = sh / ih
    } else {
        scaleX = sw / iw
        scaleY = (sh / 2) / ih
    }

    scale := min(scaleX, scaleY)
    op := &ebiten.DrawImageOptions{}

    op.GeoM.Scale(scale, scale)

	if setHorizontal {
		offsetX = (sw - (iw * scale))
		offsetY = (sh - (ih * scale)) / 2
		op.GeoM.Translate(0, offsetY)
    } else {
        offsetX = (sw - (iw * scale)) / 2
        offsetY = (sh - (ih * scale))
        op.GeoM.Translate(offsetX, 0)
    }

    screen.DrawImage(s.Top, op)

	op = &ebiten.DrawImageOptions{}

	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(offsetX, offsetY)
	screen.DrawImage(s.Bottom, op)

    s.BtmAbs = BtmAbs{
		T: int(offsetY),
		B: int(offsetY + (scale * ih)),
		L: int(offsetX),
		R: int(offsetX + (scale * iw)),
		W: int(scale * iw),
		H: int(scale * ih),
	}
}
