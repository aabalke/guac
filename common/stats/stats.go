package stats

import (
	"context"
	"math"
	"sync/atomic"
	"time"
)

type Stats struct {
	frameCnt atomic.Uint64
	fps      atomic.Uint64
}

func NewStats() *Stats {
	return &Stats{}
}

func (s *Stats) TickFrame() { s.frameCnt.Add(1) }

func (s *Stats) RunSampler(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	var lastFrames uint64
	lastTime := time.Now()

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:

			var (
				elapsed = now.Sub(lastTime).Seconds()
				frames  = s.frameCnt.Load()
				fps     = float64(frames-lastFrames) / elapsed
			)

			s.fps.Store(math.Float64bits(fps))

			lastFrames = frames
			lastTime = now
		}
	}
}

func (s *Stats) FPS() float64  { return math.Float64frombits(s.fps.Load()) }
func (s *Stats) Frame() uint64 { return s.frameCnt.Load() }
