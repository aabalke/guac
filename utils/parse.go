package utils

import "strings"

type RomType int

const (
	NONE RomType = iota
	GB
	GBA
	NDS
)

func GetRomType(path string) RomType {

	switch {
	case strings.HasSuffix(path, ".gb"):
		return GB
	case strings.HasSuffix(path, ".gbc"):
		return GB
	case strings.HasSuffix(path, ".gba"):
		return GBA
	case strings.HasSuffix(path, ".nds"):
		return NDS
	default:
		return NONE
	}
}
