package save

import (
	//"os"

	"github.com/aabalke33/guac/emu/gba"
)

type SaveState struct {
	Filepath string
}

func (s *SaveState) Save(gba *gba.GBA) {
}
