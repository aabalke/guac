package profiler

import (
	"fmt"
	"os"
	"runtime/pprof"
	"time"

	"github.com/aabalke/guac/config"
)

var (
	t time.Time
	f *os.File
)

const UNLIMITED_FPS = 0x1000_0000

func Profile(frame uint64) {
	p := &config.Conf.Profile

	switch {
	case frame == p.StartTick:

		var err error
		f, err = os.Create(p.FilePath)
		if err != nil {
			panic(err)
		}

		if err := pprof.StartCPUProfile(f); err != nil {
			panic(err)
		}

		fmt.Printf("Starting Profiler...\n")
		t = time.Now()

	case frame >= p.EndTick:
		dur := time.Since(t).Seconds()
		reqDur := (float64(p.EndTick-p.StartTick) / 60.0)

		pprof.StopCPUProfile()
		if err := f.Close(); err != nil {
			panic(err)
		}

		fmt.Printf("Closing Profiler: Duration: %.2f, %.2fx faster\n", time.Since(t).Seconds(), reqDur/dur)
		os.Exit(0)
	}
}
