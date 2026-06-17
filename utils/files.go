package utils

import (
	"bufio"
	"os"

	"github.com/sqweek/dialog"
)

func ReadFile(path string) (buf []uint8, length int, ok bool) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, false
	}
	return buf, len(buf), true
}

func WriteFile(path string, buf []uint8) bool {

	f, err := os.Create(path)
	if err != nil {
		return false
	}
	defer f.Close()

	writer := bufio.NewWriter(f)

	_, err = writer.Write(buf)
	if err != nil {
		return false
	}

	return true
}

func OpenFile(title, desc string, extensions ...string) string {

	if len(extensions) == 0 {
		extensions = append(extensions, "*")
	}

	file, err := dialog.File().Title(title).Filter(desc, extensions...).Load()
	if err != nil {
		return ""
	}

	return file
}

func OpenDirectory(title, defaultPath string) string {

	directory, err := dialog.Directory().Title(title).Browse()
	if err != nil {
		return defaultPath
	}

	return directory
}

func IsFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func IsDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return info.IsDir()
}
