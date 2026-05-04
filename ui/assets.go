package ui

import (
	"embed"
	img "image"
	"image/color"
	"image/png"
	"log"

	"github.com/aabalke/guac/config"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

var dmg_palettes = map[string][4]string{
	"gray": {"FFFFFF", "B0B0B0", "686868", "000000"},
	"bgb":  {"E8FCCC", "ACD490", "548C70", "142C38"},
	"same": {"C6DE8C", "84A563", "396139", "081810"},
	"sky":  {"818F38", "647D43", "566D3F", "314A2D"},
}

var theme_palettes = map[string][4]string{
	"grey":   {"0A0A0A", "FFFFFF", "000000", "CCCCCC"},
	"cherry": {"0A0A0A", "DDAAAA", "FFFFFF", "BB8888"},
	"dark":   {"222222", "222222", "CCCCCC", "447777"},
	"bgb":    {"142C38", "142C38", "ACD490", "548C70"},
}

var paddingInset = widget.Insets{
	Left:  4,
	Right: 4,
}

var buttonInset = widget.Insets{
	Left:   32,
	Right:  32,
	Top:    4,
	Bottom: 4,
}

var transparentNine = image.NewNineSliceColor(color.Transparent)

//go:embed assets
var embeddedAssets embed.FS

type Resources struct {
	bg, fg, sec          *image.NineSlice
	bgClr, fgClr, secClr *color.Color
	fonts                *fonts
	checkbox             *widget.CheckboxImage
	buttonImage          *widget.ButtonImage
	buttonBorderedImage  *widget.ButtonImage

	icon []img.Image

	localization *Localization

	ui *ebitenui.UI
}

func NewUIResources() (*Resources, error) {

	var (
		conf = &config.Conf.Ui
		bg   = image.NewNineSliceColor(conf.MenuBackgroundColor)
		fg   = image.NewNineSliceColor(conf.MenuForegroundColor)
		sec  = image.NewNineSliceColor(conf.MenuSecondaryColor)
		secb = image.NewBorderedNineSliceColor(conf.MenuSecondaryColor, conf.MenuForegroundColor, 2)
	)

	f, err := embeddedAssets.Open("assets/graphics/icon.png")
	if err != nil {
		panic(err)
	}

	icon, err := png.Decode(f)
	if err != nil {
		panic(err)
	}

	buttonImage := &widget.ButtonImage{
		Idle:         transparentNine,
		Hover:        secb,
		Pressed:      sec,
		PressedHover: secb,
	}

	buttonBorderedImage := &widget.ButtonImage{
		Idle:         image.NewBorderedNineSliceColor(color.Transparent, conf.MenuForegroundColor, 2),
		Hover:        secb,
		Pressed:      secb,
		PressedHover: secb,
	}

	r := &Resources{
		icon:                []img.Image{icon},
		fonts:               loadFonts(),
		bg:                  bg,
		fg:                  fg,
		sec:                 sec,
		bgClr:               &conf.MenuBackgroundColor,
		fgClr:               &conf.MenuForegroundColor,
		secClr:              &conf.MenuSecondaryColor,
		buttonImage:         buttonImage,
		buttonBorderedImage: buttonBorderedImage,
		checkbox:            loadCheckboxImage(conf.MenuForegroundColor),
		localization:        NewLocalization(config.Conf.Ui.Language),
	}

	return r, nil
}

func (u *Resources) Update() {

	conf := &config.Conf.Ui
	u.bg = image.NewNineSliceColor(conf.MenuBackgroundColor)
	u.fg = image.NewNineSliceColor(conf.MenuForegroundColor)
	u.sec = image.NewNineSliceColor(conf.MenuSecondaryColor)
	secb := image.NewBorderedNineSliceColor(conf.MenuSecondaryColor, conf.MenuForegroundColor, 2)

	u.buttonImage.Hover = secb
	u.buttonImage.Pressed = u.sec
	u.buttonImage.PressedHover = secb

	u.buttonBorderedImage = &widget.ButtonImage{
		Idle:         image.NewBorderedNineSliceColor(color.Transparent, conf.MenuForegroundColor, 2),
		Hover:        secb,
		Pressed:      secb,
		PressedHover: secb,
	}

	u.checkbox = loadCheckboxImage(conf.MenuForegroundColor)

}

type fonts struct {
	face      *text.Face
	smallFace *text.Face
}

func loadFonts() *fonts {

	fontFaceRegular := "assets/fonts/museo_slab_500.otf"
	//fontFaceRegular := "assets/fonts/noto_sana_jp_reg.ttf"

	face, err := loadFont(fontFaceRegular, 36)
	if err != nil {
		panic("could not load embedded font")
	}

	smallFace, err := loadFont(fontFaceRegular, 24)
	if err != nil {
		panic("could not load embedded font")
	}

	return &fonts{
		face:      &face,
		smallFace: &smallFace,
	}
}

func loadFont(path string, size float64) (text.Face, error) {
	fontFile, err := embeddedAssets.Open(path)
	if err != nil {
		return nil, err
	}

	s, err := text.NewGoTextFaceSource(fontFile)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return &text.GoTextFace{
		Source: s,
		Size:   size,
	}, nil
}

func loadCheckboxImage(clr color.Color) *widget.CheckboxImage {

	imgs := make([]*ebiten.Image, 4)

	for i, v := range []string{
		"assets/graphics/checkbox_idle.png",
		"assets/graphics/checkbox_checked.png",
		"assets/graphics/checkbox_idle_hover.png",
		"assets/graphics/checkbox_checked_hover.png",
	} {

		f, err := embeddedAssets.Open(v)
		if err != nil {
			panic("could not load checkbox from embedded assets")
		}

		defer f.Close()
		imgs[i], _, _ = ebitenutil.NewImageFromReader(f)
		imgs[i] = TintImage(imgs[i], clr)
	}

	s := [3]int{32, 0, 0}

	return &widget.CheckboxImage{
		Unchecked:        image.NewNineSlice(imgs[0], s, s),
		Checked:          image.NewNineSlice(imgs[1], s, s),
		UncheckedHovered: image.NewNineSlice(imgs[2], s, s),
		CheckedHovered:   image.NewNineSlice(imgs[3], s, s),
	}
}

func TintImage(src *ebiten.Image, c color.Color) *ebiten.Image {

	r, g, b, _ := c.RGBA()
	rf := float32(r) / 0xFFFF
	gf := float32(g) / 0xFFFF
	bf := float32(b) / 0xFFFF
	//af := float32(a) / 0xFFFF

	w, h := src.Bounds().Size().X, src.Bounds().Size().Y
	dst := ebiten.NewImage(w, h)

	op := &ebiten.DrawImageOptions{}
	op.ColorScale.Scale(rf, gf, bf, 1)

	dst.DrawImage(src, op)
	return dst
}

func NewTheme(r *Resources) *widget.Theme {
	return &widget.Theme{
		DefaultFace:      r.fonts.face,
		DefaultTextColor: *r.fgClr,
		ButtonTheme: &widget.ButtonParams{
			TextColor: &widget.ButtonTextColor{
				Idle:    *r.fgClr,
				Hover:   *r.fgClr,
				Pressed: *r.fgClr,
			},
			TextFace:    r.fonts.face,
			Image:       r.buttonImage,
			TextPadding: &buttonInset,
			TextPosition: &widget.TextPositioning{
				VTextPosition: widget.TextPositionCenter,
				HTextPosition: widget.TextPositionCenter,
			},
		},
		TextTheme: &widget.TextParams{
			Face:  r.fonts.smallFace,
			Color: *r.fgClr,
		},

		TextInputTheme: &widget.TextInputParams{
			Face: r.fonts.smallFace,
			Image: &widget.TextInputImage{
				Idle:  image.NewBorderedNineSliceColor(color.Transparent, *r.fgClr, 2),
				Hover: image.NewBorderedNineSliceColor(*r.secClr, *r.fgClr, 2),
			},
			Color: &widget.TextInputColor{
				Idle:  *r.fgClr,
				Caret: *r.fgClr,
			},
			Padding: &paddingInset,
		},
		SliderTheme: &widget.SliderParams{
			TrackImage: &widget.SliderTrackImage{
				Idle:  transparentNine,
				Hover: transparentNine,
			},
			HandleImage: &widget.ButtonImage{
				Idle:    r.sec,
				Hover:   r.fg,
				Pressed: r.fg,
			},
		},
		CheckboxTheme: &widget.CheckboxParams{
			Image: r.checkbox,
		},
	}
}
