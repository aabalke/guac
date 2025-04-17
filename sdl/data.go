package sdl

import (
	"encoding/json"
	"io"
	"os"
)

type GameData struct {
	Name    string `json:"Name"`
	RomPath string `json:"RomPath"`
	SavPath string `json:"SavPath"`
	ArtPath string `json:"ArtPath"`
	Year    int    `json:"Year"`
	Console string `json:"Console"`
}

const path = "./roms.json"

func LoadGameData() *[]GameData {

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

    return &data
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

func ReorderGameData(gameData *[]GameData, idx int) []GameData {

    temp := []GameData{(*gameData)[idx]}

    for i, v := range (*gameData) {
        if i != idx {
            temp = append(temp, v)
        }
    }

    return temp
}
