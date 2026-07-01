package gba

func (gba *GBA) AudioSampleEvent(late int64, arg any) {
	gba.Apu.SoundClock()
	gba.Scheduler.schedule(EVENT_SND_SAMPLE_GEN, 1, CYCLES_PER_SND_GEN-late, gba.AudioSampleEvent, nil)
}

func (gba *GBA) HblankEvent(late int64, arg any) {
	dispstat := &gba.Mem.Dispstat
	dispstat.SetHBlank(true)
	if (*dispstat>>4)&1 != 0 {
		gba.Irq.SetIRQ(1)
	}

	vcount := gba.Mem.IO[6]

	gba.Dma[3].videoDma(vcount)

	if vcount < SCREEN_HEIGHT {
		updateBackgrounds(gba, &gba.PPU.Dispcnt)
		gba.PPU.bgPriorities = gba.getBgPriority(uint32(vcount), gba.PPU.Dispcnt.Mode, &gba.PPU.Backgrounds)
		gba.PPU.objPriorities = gba.getObjPriority(uint32(vcount), &gba.PPU.Objects)
		gba.scanlineGraphics(uint32(vcount))
		gba.PPU.Backgrounds[2].BgAffineUpdate()
		gba.PPU.Backgrounds[3].BgAffineUpdate()
		gba.checkDmas(DMA_MODE_HBL)
	}
}

func (gba *GBA) ScanlineEndEvent(late int64, arg any) {
	dispstat := &gba.Mem.Dispstat
	vcount := &gba.Mem.IO[6]

	dispstat.SetHBlank(false)

	*vcount++

	switch *vcount {

	case SCREEN_HEIGHT:
		dispstat.SetVBlank(true)
		gba.checkDmas(DMA_MODE_VBL)
		// bios/bios.gba needs irq set on screen_height, iridion 3d needs screen_height + 1
		// I believe this is cycle related
		if (*dispstat>>3)&1 != 0 {
			gba.Irq.SetIRQ(0)
		}

	case 227:
		dispstat.SetVBlank(false)
	}

	match := dispstat.GetLYC() == *vcount
	dispstat.SetVCFlag(match)

	if vcounterIRQ := (*dispstat>>5)&1 != 0; vcounterIRQ && match {
		gba.Irq.SetIRQ(2)
	}

	gba.Scheduler.schedule(EVENT_END_SCANLINE, 1, CYCLES_SCANLINE-late, gba.ScanlineEndEvent, nil)
	gba.Scheduler.schedule(EVENT_HBK, 1, CYCLES_HDRAW-late, gba.HblankEvent, nil)
}

func (gba *GBA) FrameEndEvent(late int64, arg any) {
	dispstat := &gba.Mem.Dispstat

	gba.Apu.Play(gba.Muted, true)
	gba.Frame++
	gba.Image.WritePixels(gba.Pixels)

	gba.Mem.IO[6] = 0
	match := dispstat.GetLYC() == 0
	dispstat.SetVCFlag(match)

	if vcounterIRQ := (*dispstat>>5)&1 != 0; vcounterIRQ && match {
		gba.Irq.SetIRQ(2)
	}
	gba.PPU.Backgrounds[2].BgAffineReset()
	gba.PPU.Backgrounds[3].BgAffineReset()

	gba.Scheduler.schedule(EVENT_END_FRAME, 1, CYCLES_FRAME-late, gba.FrameEndEvent, nil)
}
