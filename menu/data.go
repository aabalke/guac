package menu

import (
	"encoding/json"
	"io"
	"os"
    "strings"

	"github.com/hajimehoshi/ebiten/v2"
)

type GameData struct {
	RomPath string `json:"RomPath"`
	SavPath string
	ArtPath string `json:"ArtPath"`
    Image   *ebiten.Image
    Type int
}

const path = "./roms.json"

const (
    NONE = iota
    GB
    GBA
)

func LoadGameData() []GameData {

    f, err := os.Open(path)
    if err != nil {
        panic(err)
    }

    bytes, err := io.ReadAll(f)
    if err != nil {
        panic(err)
    }

    var data []GameData

    if err := json.Unmarshal(bytes, &data); err != nil {
        panic(err)
    }

    for i, v := range data {
        img, err := loadImage(v.ArtPath)
        if err != nil {
            panic(err)
        }

        data[i].Image = img

        data[i].Type = getConsole(data[i].RomPath)
    }

    return data
}

func WriteGameData(gameData *[]GameData) {

    data, err := json.MarshalIndent(gameData, "", " ")
    if err != nil {
        panic(err)
    }

    if err := os.WriteFile(path, data, 0644); err != nil {
        panic(err)
    }
}

func getConsole(path string) int {

    switch {
    case strings.HasSuffix(path, ".gb"):
        return GB
    case strings.HasSuffix(path, ".gbc"):
        return GB
    case strings.HasSuffix(path, ".gba"):
        return GBA
    default:
        panic("Flag Parsing Error. RomPath in roms.json must end with gba, gbc, gb extension")
    }
}

func ReorderGameData(gameData *[]GameData, idx int) []GameData {

    temp := []GameData{(*gameData)[idx]}

    for i, v := range (*gameData) {
        if i != idx {
            temp = append(temp, v)
        }
    }

    return temp
}
