package utils

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/hajimehoshi/dialog"
)

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

func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		return fmt.Errorf("unsupported platform")
	}

	return cmd.Start()
}

func OpenLink(path string) {
	if err := openBrowser(path); err != nil {
		panic(err)
	}
}
