package debug

import (
	"bufio"
	"os"
)

func readFile(path string) (buf []uint8, length int, ok bool) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, false
	}
	return buf, len(buf), true
}

func writeFile(path string, buf []uint8) bool {

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
