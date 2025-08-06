package menu

import (
	"bytes"
	"os"

	"image"
	_ "image/jpeg"
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
)

func loadImage(path string) (*ebiten.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}

	return ebiten.NewImageFromImage(img), nil
}

func loadImageEmbed(b []byte) (*ebiten.Image, error) {

	img, _, err := image.Decode(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	return ebiten.NewImageFromImage(img), nil
}
