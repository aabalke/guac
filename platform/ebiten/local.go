package ui

import (
	"fmt"
	"io"
	"strings"

	"github.com/BurntSushi/toml"
)

type Localization map[string]string

func (l *Localization) AsSlice(k string) []string {
	return strings.Split((*l)[k], ",")
}

type LangOptions int

const (
	ENGLISH LangOptions = iota
	SPANISH
)

var langAbbreviations = map[LangOptions]string{
	ENGLISH: "en",
	SPANISH: "es",
}

func NewLocalization(lang LangOptions) Localization {
	abb, ok := langAbbreviations[lang]
	if !ok {
		abb = "en"
	}

	path := "assets/local/" + abb + ".toml"

	f, err := embeddedAssets.Open(path)
	if err != nil {
		panic(fmt.Sprintf("could not get localization file. err: %s", err))
	}

	var l map[string]any
	out := make(Localization)

	data, err := io.ReadAll(f)
	if err != nil {
		panic(fmt.Sprintf("could not read localization file. err: %s", err))
	}

	_, err = toml.Decode(string(data), &l)
	if err != nil {
		panic(fmt.Sprintf("could not decode localization file. err: %s", err))
	}

	flattenToml("", l, out)

	return out
}

func flattenToml(prefix string, in map[string]any, out map[string]string) {
	for k, v := range in {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}

		switch val := v.(type) {
		case map[string]any:
			flattenToml(key, val, out)
		case string:
			out[key] = val
		case []any:
			var vs []string
			for _, v := range val {
				vs = append(vs, v.(string))
			}
			out[key] = strings.Join(vs, ",")
		}
	}
}
