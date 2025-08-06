package menu

import (
	"image/color"
	"slices"

	_ "embed"

	"github.com/aabalke/guac/config"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
)

type Menu struct {
	SelectedIdx int
	Data        []GameData

	menuPlayer *MenuPlayer

    Splash *ebiten.Image
}

//go:embed res/splash.jpg
var splashBytes []byte

func NewMenu(context *audio.Context) *Menu {
    splash, err := loadImageEmbed(splashBytes)
    if err != nil {
        panic(err)
    }
	m := &Menu{
		Data: LoadGameData(),
        Splash: splash,
	}

	p, err := NewMenuPlayer(context)
	if err != nil {
		panic(err)
	}

	m.menuPlayer = p

	return m
}

func (m *Menu) InputHandler(keys []ebiten.Key, buttons []ebiten.StandardGamepadButton, frame uint64) bool {

    if frame < 120 {
        return false
    }

	gamesPerRow := config.Conf.GamesPerRow

	m.menuPlayer.handleChannels()

	keyConfig := config.Conf.KeyboardConfig
	buttonConfig := config.Conf.ControllerConfig

	for _, key := range keys {
		keyStr := key.String()

		switch {
		case slices.Contains(keyConfig.Up, keyStr):
			if m.SelectedIdx-gamesPerRow < 0 {
				m.menuPlayer.update(1)
			} else {
				m.menuPlayer.update(0)
				m.SelectedIdx = max(0, (m.SelectedIdx)-gamesPerRow)
			}
		case slices.Contains(keyConfig.Down, keyStr):
			if m.SelectedIdx+gamesPerRow > len(m.Data)-1 {
				m.menuPlayer.update(1)
			} else {
				m.menuPlayer.update(0)
				m.SelectedIdx = min(len(m.Data)-1, (m.SelectedIdx)+gamesPerRow)
			}
		case slices.Contains(keyConfig.Right, keyStr):
			if m.SelectedIdx+1 > len(m.Data)-1 {
				m.menuPlayer.update(1)
			} else {
				m.menuPlayer.update(0)
				m.SelectedIdx = min(len(m.Data)-1, (m.SelectedIdx)+1)
			}
		case slices.Contains(keyConfig.Left, keyStr):
			if m.SelectedIdx-1 < 0 {
				m.menuPlayer.update(1)
			} else {
				m.menuPlayer.update(0)
				m.SelectedIdx = max(0, (m.SelectedIdx)-1)
			}
		case slices.Contains(keyConfig.Select, keyStr):
			m.menuPlayer.update(2)
			return true
		}
	}

	for _, button := range buttons {
		buttonStr := int(button)

		switch {
		case slices.Contains(buttonConfig.Select, buttonStr):
			m.menuPlayer.update(2)
			return true
		case slices.Contains(buttonConfig.Right, buttonStr):
			if m.SelectedIdx+1 > len(m.Data)-1 {
				m.menuPlayer.update(1)
			} else {
				m.menuPlayer.update(0)
				m.SelectedIdx = min(len(m.Data)-1, (m.SelectedIdx)+1)
			}
		case slices.Contains(buttonConfig.Left, buttonStr):
			if m.SelectedIdx-1 < 0 {
				m.menuPlayer.update(1)
			} else {
				m.menuPlayer.update(0)
				m.SelectedIdx = max(0, (m.SelectedIdx)-1)
			}
		case slices.Contains(buttonConfig.Up, buttonStr):
			if m.SelectedIdx-gamesPerRow < 0 {
				m.menuPlayer.update(1)
			} else {
				m.menuPlayer.update(0)
				m.SelectedIdx = max(0, (m.SelectedIdx)-gamesPerRow)
			}
		case slices.Contains(buttonConfig.Down, buttonStr):
			if m.SelectedIdx+gamesPerRow > len(m.Data)-1 {
				m.menuPlayer.update(1)
			} else {
				m.menuPlayer.update(0)
				m.SelectedIdx = min(len(m.Data)-1, (m.SelectedIdx)+gamesPerRow)
			}
		}
	}

	return false
}

func (m *Menu) DrawMenu(screen *ebiten.Image, frame uint64) {

    if frame < 120 {
        screen.Fill(color.White)
        m.SplashImage(screen)
        return
    }

	sw, _ := screen.Bounds().Dx(), screen.Bounds().Dy()
	elementUnit := float64(sw / config.Conf.GamesPerRow)

	row := float64(m.SelectedIdx / config.Conf.GamesPerRow)
	//maxRow := float64((len(m.Data) - 1) / config.Conf.Menus.GamesPerRow)

	//var rowOffset float64
	//switch row {
	//case 0: rowOffset = 0
	//case maxRow:

	//    // not sure how to handle currently

	//    if maxRow * elementUnit > float64(screen.Bounds().Dy()) {
	//        rowOffset = (elementUnit * row) - float64(screen.Bounds().Dy())
	//    } else {
	//        rowOffset = elementUnit * row
	//    }

	//default:
	//    rowOffset = elementUnit * row
	//}

	rowOffset := elementUnit * row

	for i := range len(m.Data) {
		x := float64(i%config.Conf.GamesPerRow) * elementUnit
		y := float64(i/config.Conf.GamesPerRow)*elementUnit - rowOffset
		m.Image(screen, x, y, elementUnit, i)
	}
}

func (m *Menu) Image(screen *ebiten.Image, x, y, elementUnit float64, i int) {

	img := (m.Data)[i].Image

	s := (elementUnit / float64(img.Bounds().Dx()))

	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Scale(s, s)
	opts.GeoM.Translate(x, y)

	if shadeUnselected := i != m.SelectedIdx; shadeUnselected {
		opts.ColorScale.ScaleAlpha(0.5)
	}

	screen.DrawImage(img, opts)
}

func (m *Menu) SplashImage(screen *ebiten.Image) {//, x, y, elementUnit float64, i int) {

	img := m.Splash

	sw, sh := float64(screen.Bounds().Dx()), float64(screen.Bounds().Dy())
	iw, ih := float64(img.Bounds().Dx()), float64(img.Bounds().Dy())

	scaleX := sw / iw / 4
	scaleY := sh / ih / 4
	scale := min(scaleX, scaleY)

	offsetX := (sw - (iw * scale)) / 2
	offsetY := (sh - (ih * scale)) / 2

	opts := &ebiten.DrawImageOptions{}
	//opts.GeoM.Scale(s, s)
	opts.GeoM.Scale(scale, scale)
	opts.GeoM.Translate(offsetX, offsetY)

	//if shadeUnselected := i != m.SelectedIdx; shadeUnselected {
	//	opts.ColorScale.ScaleAlpha(0.5)
	//}

	screen.DrawImage(img, opts)
}
