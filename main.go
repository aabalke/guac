package main

import (
	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/ui"
)

func main() {

	config.Conf.Decode()
	// parse flags

	// if no -r flag
	ui.StartEngine()
}
