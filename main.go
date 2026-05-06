package main

import (
	"github.com/aabalke/guac/config/file"
	"github.com/aabalke/guac/ui"
)

func main() {

	file.Decode()
	// parse flags

	// if no -r flag
	ui.StartEngine()
}
