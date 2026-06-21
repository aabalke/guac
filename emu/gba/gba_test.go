package gba

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/config/file"
)

type TestResult struct {
	Name   string
	Passed bool
}

func TestAlyosha(t *testing.T) {
	directory := `C:\dev\repos\emulators\roms\gba_test_roms\alyosha-tas-gba-tests`
	t.Logf("Alyosha Tas Gba Test Suite %s\n", time.Now().Format(time.RFC3339))

	perRomHandler := func(file string, results *[]TestResult) {
		finish := false
		passed := false

		gba := NewGBA(file, nil)

		gba.InstInjectionFunc = func(op uint32) {
			if op == 0xEAFF_FFFE {
				passed = gba.Cpu.Reg.R[12] == 0
				finish = true
			}
		}

		func() {
			defer func() {
				recover()
			}()

			for range 60 * 10 {
				gba.Update(false)

				if finish {
					break
				}
			}
		}()

		gba.Close()

		p, _ := filepath.Rel(directory, file)

		name := strings.TrimSuffix(p, ".gba")

		*results = append(*results, TestResult{
			Name:   name,
			Passed: passed,
		})

		if !passed {
			t.Errorf("Failed %s", file)
		}
	}

	testDirOfRoms("Alyosha Tas Gba Tests", "testing_alyosha.md", directory, perRomHandler, t)
}

func testDirOfRoms(title, outputFile, directory string, perRomHandler func(string, *[]TestResult), t *testing.T) {
	file.Decode()

	config.Conf.General.Headless = true

	files, err := FindFiles(directory, ".gba")
	if err != nil {
		panic(err)
	}

	var results []TestResult

	for _, file := range files {
		perRomHandler(file, &results)
	}

	err = WriteMarkdown(title, results, outputFile)
	if err != nil {
		t.Fatal(err)
	}
}

func FindFiles(root, ext string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if filepath.Ext(path) == ext {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

func WriteMarkdown(header string, results []TestResult, filename string) error {
	passed, total := 0, len(results)
	for _, r := range results {
		if r.Passed {
			passed++
		}
	}

	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n", header)
	fmt.Fprintf(&b, "Results generated %s\n\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(&b, "Passing %d/%d %02d%%\n\n", passed, total, (passed*100)/total)

	slices.SortFunc(results, func(a, b TestResult) int {
		return strings.Compare(a.Name, b.Name)
	})

	for _, r := range results {
		if r.Passed {
			fmt.Fprintf(&b, "👍 %s\n", r.Name)
		} else {
			fmt.Fprintf(&b, "❌ %s\n", r.Name)
		}
	}

	return os.WriteFile(filename, []byte(b.String()), 0o644)
}
