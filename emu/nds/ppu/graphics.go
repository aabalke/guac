package ppu

import (
	"encoding/binary"
	"sync"
)

var (
	b16 = binary.LittleEndian.Uint16
	wg  = sync.WaitGroup{}
)

func (ppu *PPU) Graphics(y uint32, singleThread bool) {

	a := &ppu.EngineA
	b := &ppu.EngineB

    for i := range 4 {
        a.Backgrounds[i].SetSize()
        b.Backgrounds[i].SetSize()
    }

	a.getBgPriority(y)
	b.getBgPriority(y)
	a.getObjPriority(y)
	b.getObjPriority(y)

	ppu.buildFrame(y, singleThread)

	a.Backgrounds[2].BgAffineUpdate()
	a.Backgrounds[3].BgAffineUpdate()
	b.Backgrounds[2].BgAffineUpdate()
	b.Backgrounds[3].BgAffineUpdate()
}

func (ppu *PPU) buildFrame(y uint32, singleThread bool) {

	if singleThread {

		a := &ppu.EngineA
		capture := &ppu.Capture
		renderingB := ppu.Rasterizer.Buffers.BisRendering

		switch a.Dispcnt.DisplayMode {
		case 0:
			ppu.screenoff(y, a)

		case 1:

			ppu.standard(y, a)
			if capture.ActiveCapture {
				capture.CaptureLine(y, renderingB)
			}

		case 2:
			if capture.ActiveCapture {
				ppu.standard(y, a)
				capture.CaptureLine(y, renderingB)
			}

			ppu.vramDisplay(y, a)

		case 3:
			panic("main memory fifo display unsupported")
		}

		b := &ppu.EngineB
		switch b.Dispcnt.DisplayMode {
		case 0:
			ppu.screenoff(y, b)
		case 1:
			ppu.standard(y, b)
		}
		return
	}

	wg.Add(2)
	go func() {
		defer wg.Done()

		a := &ppu.EngineA
		capture := &ppu.Capture
		renderingB := ppu.Rasterizer.Buffers.BisRendering

		switch a.Dispcnt.DisplayMode {
		case 0:
			ppu.screenoff(y, a)

		case 1:

			ppu.standard(y, a)
			if capture.ActiveCapture {
				capture.CaptureLine(y, renderingB)
			}

		case 2:
			if capture.ActiveCapture {
				ppu.standard(y, a)
				capture.CaptureLine(y, renderingB)
			}

			ppu.vramDisplay(y, a)

		case 3:
			panic("main memory fifo display unsupported")
		}
	}()

	go func() {
		defer wg.Done()

		b := &ppu.EngineB
		switch b.Dispcnt.DisplayMode {
		case 0:
			ppu.screenoff(y, b)
		case 1:
			ppu.standard(y, b)
		}
	}()

	wg.Wait()
}
