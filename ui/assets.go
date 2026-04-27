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

var transparentNine = image.NewNineSliceColor(color.Transparent)

//go:embed assets
var embeddedAssets embed.FS

type UiResources struct {
	bg, fg, sec          *image.NineSlice
	bgClr, fgClr, secClr *color.Color
	fonts                *fonts
	checkbox             *widget.CheckboxImage
	buttonImage          *widget.ButtonImage

	icon []img.Image

	ui *ebitenui.UI
}

func NewUIResources() (*UiResources, error) {

	var (
		conf = &config.Conf.Ui
		bg   = image.NewNineSliceColor(conf.MenuBackgroundColor)
		fg   = image.NewNineSliceColor(conf.MenuForegroundColor)
		sec  = image.NewNineSliceColor(conf.MenuSecondaryColor)
	)

	f, err := embeddedAssets.Open("assets/graphics/icon.png")
	if err != nil {
		panic(err)
	}

	icon, err := png.Decode(f)
	if err != nil {
		panic(err)
	}

	fonts, err := loadFonts()
	if err != nil {
		return nil, err
	}

	checkboxImage, err := loadCheckboxImage(conf.MenuForegroundColor)
	if err != nil {
		panic(err)
	}

	buttonImage := &widget.ButtonImage{
		Idle:    transparentNine,
		Hover:   sec,
		Pressed: sec,
	}

	return &UiResources{
		icon:        []img.Image{icon},
		fonts:       fonts,
		bg:          bg,
		fg:          fg,
		sec:         sec,
		bgClr:       &conf.MenuBackgroundColor,
		fgClr:       &conf.MenuForegroundColor,
		secClr:      &conf.MenuSecondaryColor,
		buttonImage: buttonImage,
		checkbox:    checkboxImage,
	}, nil
}

func (u *UiResources) Update() {

	conf := &config.Conf.Ui
	u.bg = image.NewNineSliceColor(conf.MenuBackgroundColor)
	u.fg = image.NewNineSliceColor(conf.MenuForegroundColor)
	u.sec = image.NewNineSliceColor(conf.MenuSecondaryColor)
	u.buttonImage.Hover = u.sec
	u.buttonImage.Pressed = u.sec

	checkboxImage, err := loadCheckboxImage(conf.MenuForegroundColor)
	if err != nil {
		panic(err)
	}

	u.checkbox = checkboxImage
}

type fonts struct {
	face      *text.Face
	smallFace *text.Face
}

func loadFonts() (*fonts, error) {

	fontFaceRegular := "assets/fonts/museo_slab_500.otf"

	face, err := loadFont(fontFaceRegular, 36)
	if err != nil {
		return nil, err
	}

	smallFace, err := loadFont(fontFaceRegular, 24)
	if err != nil {
		return nil, err
	}

	return &fonts{
		face:      &face,
		smallFace: &smallFace,
	}, nil
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

func newImageFromFile(path string) (*ebiten.Image, error) {
	f, err := embeddedAssets.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	i, _, err := ebitenutil.NewImageFromReader(f)
	return i, err
}

func loadImageNineSlice(path string, centerWidth int, centerHeight int) (*image.NineSlice, error) {
	i, err := newImageFromFile(path)
	if err != nil {
		return nil, err
	}
	w := i.Bounds().Dx()
	h := i.Bounds().Dy()
	return image.NewNineSlice(i,
			[3]int{(w - centerWidth) / 2, centerWidth, w - (w-centerWidth)/2 - centerWidth},
			[3]int{(h - centerHeight) / 2, centerHeight, h - (h-centerHeight)/2 - centerHeight}),
		nil
}

func loadCheckboxImage(clr color.Color) (*widget.CheckboxImage, error) {
	f1, err := embeddedAssets.Open("assets/graphics/checkbox_idle.png")
	if err != nil {
		return nil, err
	}
	defer f1.Close()
	unchecked, _, _ := ebitenutil.NewImageFromReader(f1)

	f2, err := embeddedAssets.Open("assets/graphics/checkbox_checked.png")
	if err != nil {
		return nil, err
	}
	defer f2.Close()
	checked, _, _ := ebitenutil.NewImageFromReader(f2)

	unchecked = TintImage(unchecked, clr)
	checked = TintImage(checked, clr)

	s := [3]int{32, 0, 0}

	return &widget.CheckboxImage{
		Unchecked: image.NewNineSlice(unchecked, s, s),
		Checked:   image.NewNineSlice(checked, s, s),
	}, nil
}

func TintImage(src *ebiten.Image, c color.Color) *ebiten.Image {

	r, g, b, a := c.RGBA()
	rf := float32(r) / 0xFFFF
	gf := float32(g) / 0xFFFF
	bf := float32(b) / 0xFFFF
	af := float32(a) / 0xFFFF

	w, h := src.Bounds().Size().X, src.Bounds().Size().Y
	dst := ebiten.NewImage(w, h)

	op := &ebiten.DrawImageOptions{}
	op.ColorScale.Scale(rf, gf, bf, af)

	dst.DrawImage(src, op)
	return dst
}

type PageId int

const (
	PAGE_HOME PageId = iota
	PAGE_PAUSE
	PAGE_SETTINGS
)

type UiPage struct {
	Id PageId
	ui *ebitenui.UI
}
