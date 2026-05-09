package utils

import (
	"math"

	"github.com/aabalke/guac/config"
)

func ScaleImage(sw, sh, cw, ch float64) float64 {
	switch {
	case !config.Conf.General.IntegerScaling:
		return min(sw/cw, sh/ch)
	case config.Conf.General.IntegerScalingRatio == config.DYNAMIC_INT_SCALING:
		return math.Floor(min(sw/cw, sh/ch))
	default:
		return float64(config.Conf.General.IntegerScalingRatio)
	}
}
