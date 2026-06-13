package flags

import (
	"flag"
	"os"

	"github.com/aabalke/guac/config"
)

func Decode() {
	var (
		romPath  = flag.String("r", "", "rom path")
		profile  = flag.Bool("p", false, "use profiler")
		fps      = flag.Int("fps", 60, "target fps")
		mute     = flag.Bool("m", false, "mute")
		logger   = flag.Bool("l", false, "logger")
		showfps  = flag.Bool("show-fps", false, "show fps")
		headless = flag.Bool("headless", false, "headless")
	)

	if len(os.Args) >= 2 {
		config.Conf.General.RomPath = os.Args[1]
	}

	flag.Parse()

	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "headless":
			config.Conf.General.Headless = *headless
		case "r":
			config.Conf.General.RomPath = *romPath
		case "p":
			config.Conf.Profile.Enabled = *profile
		case "fps":
			config.Conf.General.TargetFps = *fps
		case "m":
			config.Conf.General.Muted = *mute
		case "l":
			config.Conf.General.Logger = *logger
		case "show-fps":
			config.Conf.General.ShowFps = *showfps
		}
	})
}
