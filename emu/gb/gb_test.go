package gb

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

func TestMooneye(t *testing.T) {
	// https://github.com/Gekkio/mooneye-test-suite
	directory := `C:\dev\repos\emulators\roms\gb_test_roms\game-boy-test-roms-v7.0\mooneye-test-suite`
	t.Logf("Mooneye Test Suite %s\n", time.Now().Format(time.RFC3339))

	perRomHandler := func(file string, results *[]TestResult) {
		finish := false
		passed := false

		gb := NewGameBoy(file, nil)

		gb.InstInjectionFunc = func(gb *GameBoy, op uint8) {
			if op == 0x40 {
				passed =
					gb.Cpu.b == 3 &&
						gb.Cpu.c == 5 &&
						gb.Cpu.d == 8 &&
						gb.Cpu.e == 13 &&
						gb.Cpu.h == 21 &&
						gb.Cpu.l == 34

				finish = true
			}
		}

		for range 60 * 10 {
			gb.Update(false)

			if finish {
				break
			}
		}

		gb.Close()

		p, _ := filepath.Rel(directory, file)

		name := strings.TrimSuffix(p, ".gb")

		*results = append(*results, TestResult{
			Name:   name,
			Passed: passed,
		})

		if !passed {
			t.Errorf("Failed %s", file)
		}
	}

	testDirOfRoms("Mooneye Acceptance Tests", "testing_mooneye.md", directory, perRomHandler, t)
}

func TestGbMicroTest(t *testing.T) {
	// https://github.com/aappleby/GBMicrotest
	directory := `C:\dev\repos\emulators\roms\gb_test_roms\game-boy-test-roms-v7.0\gbmicrotest`
	t.Logf("GBMicrotest Test Suite %s\n", time.Now().Format(time.RFC3339))

	perRomHandler := func(file string, results *[]TestResult) {
		gb := NewGameBoy(file, nil)

		for range 60 * 2 {
			gb.Update(false)
		}

		passed := gb.Read(0xFF82) == 0x1

		gb.Close()

		p, _ := filepath.Rel(directory, file)

		name := strings.TrimSuffix(p, ".gb")

		*results = append(*results, TestResult{
			Name:   name,
			Passed: passed,
		})

		if !passed {
			t.Errorf("Failed %s", file)
		}
	}

	testDirOfRoms("GBMicrotest Tests", "testing_gbmicrotest.md", directory, perRomHandler, t)
}

func testDirOfRoms(title, outputFile, directory string, perRomHandler func(string, *[]TestResult), t *testing.T) {
	file.Decode()

	config.Conf.General.Headless = true

	files, err := FindFiles(directory, ".gb")
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
