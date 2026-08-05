package main

import (
	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/config/file"
	"github.com/aabalke/guac/config/flags"
	ebiten "github.com/aabalke/guac/platform/ebiten"
	"github.com/aabalke/guac/platform/headless"
)

func main() {
	file.Decode()
	flags.Decode()

	if config.Conf.General.Headless {
		headless.StartHeadless()
		return
	}

	ebiten.StartEngine()
}
