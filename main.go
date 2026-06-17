package main

import (
	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/config/file"
	"github.com/aabalke/guac/config/flags"
	"github.com/aabalke/guac/ui"
)

func main() {
	file.Decode()
	flags.Decode()

	if config.Conf.General.Headless {
		StartHeadless()
		return
	}

	ui.StartEngine()
}
