package gba

func (gba *GBA) AudioSampleEvent(late int64, arg any) {
	gba.Apu.SoundClock(false)
	gba.Scheduler.schedule(EVENT_SND_SAMPLE_GEN, 1, CYCLES_PER_SND_GEN-late, gba.AudioSampleEvent, nil)
}

func (gba *GBA) HblankVDrawEvent(late int64, arg any) {
	gba.Mem.Dispstat |= DISP_HBL
	if (gba.Mem.Dispstat>>4)&1 != 0 {
		gba.Irq.SetIRQ(1)
	}

	vcount := gba.Mem.IO[6]
	gba.Dma.videoDma(vcount, late)

	updateBackgrounds(gba, &gba.PPU.Dispcnt)
	gba.PPU.bgPriorities = gba.getBgPriority(uint32(vcount), gba.PPU.Dispcnt.Mode, &gba.PPU.Backgrounds)
	gba.PPU.objPriorities = gba.getObjPriority(uint32(vcount), &gba.PPU.Objects)
	gba.scanlineGraphics(uint32(vcount))
	gba.PPU.Backgrounds[2].BgAffineUpdate()
	gba.PPU.Backgrounds[3].BgAffineUpdate()
	gba.Dma.raise(DMA_MODE_HBL, late)
}

func (gba *GBA) HblankVBlankEvent(late int64, arg any) {
	gba.Mem.Dispstat |= DISP_HBL
	if (gba.Mem.Dispstat>>4)&1 != 0 {
		gba.Irq.SetIRQ(1)
	}

	vcount := gba.Mem.IO[6]
	gba.Dma.videoDma(vcount, late)
}

func (gba *GBA) ScanlineEndEvent(late int64, arg any) {
	vcount := &gba.Mem.IO[6]

	gba.Mem.Dispstat &^= DISP_HBL

	*vcount++

	switch *vcount {
	case SCREEN_HEIGHT:
		gba.Mem.Dispstat |= DISP_VBL
		gba.Dma.raise(DMA_MODE_VBL, late)
		if (gba.Mem.Dispstat>>3)&1 != 0 {
			gba.Irq.SetIRQ(0)
		}

	case 227:
		gba.Mem.Dispstat &^= DISP_VBL
	case 228:
		*vcount = 0

		gba.Apu.Play(gba.Muted, gba.StdFps)
		gba.Frame++
		gba.Image.WritePixels(gba.Pixels)
		gba.PPU.Backgrounds[2].BgAffineReset()
		gba.PPU.Backgrounds[3].BgAffineReset()
	}

	gba.Mem.Dispstat &^= DISP_VCF

	if match := gba.Mem.Dispstat.GetLYC() == *vcount; match {
		gba.Mem.Dispstat |= DISP_VCF

		if vcounterIRQ := (gba.Mem.Dispstat>>5)&1 != 0; vcounterIRQ {
			gba.Irq.SetIRQ(2)
		}
	}

	gba.Scheduler.schedule(EVENT_END_SCANLINE, 1, CYCLES_SCANLINE-late, gba.ScanlineEndEvent, nil)

	if *vcount < SCREEN_HEIGHT {
		gba.Scheduler.schedule(EVENT_HBK, 1, CYCLES_HDRAW-late, gba.HblankVDrawEvent, nil)
	} else {
		gba.Scheduler.schedule(EVENT_HBK, 1, CYCLES_HDRAW-late, gba.HblankVBlankEvent, nil)
	}
}
