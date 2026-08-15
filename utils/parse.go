package utils

import (
	"strings"

	"github.com/aabalke/guac/common/file"
)

type RomType int

const (
	NONE RomType = iota
	GB
	GBA
	NDS
	NONE_ZIP
)

func GetRomType(path string) RomType {
	switch {
	case strings.HasSuffix(path, ".gb"), strings.HasSuffix(path, ".gbc"):
		return GB
	case strings.HasSuffix(path, ".gba"):
		return GBA
	case strings.HasSuffix(path, ".nds"):
		return NDS
	case strings.HasSuffix(path, ".zip"):
		switch suffix := file.GetZipType(path); suffix {
		case ".gb", ".gbc":
			return GB
		case ".gba":
			return GBA
		case ".nds":
			return NDS
		default:
			return NONE_ZIP
		}
	}

	return NONE
}
