package gba

func (gba *GBA) AudioSampleEvent(overshoot int64, arg any) bool {
	gba.Apu.SoundClock()
	gba.Scheduler.schedule(EVENT_SND_SAMPLE_GEN, 1, CYCLES_PER_SND_GEN-overshoot, gba.AudioSampleEvent, nil)
	return false
}

func (gba *GBA) HblankEvent(overshoot int64, arg any) bool {
	dispstat := &gba.Mem.Dispstat
	dispstat.SetHBlank(true)
	if (*dispstat>>4)&1 != 0 {
		gba.Irq.SetIRQ(1)
	}

	if vcount := gba.Mem.IO[0x6]; vcount < SCREEN_HEIGHT {
		updateBackgrounds(gba, &gba.PPU.Dispcnt)
		gba.PPU.bgPriorities = gba.getBgPriority(uint32(vcount), gba.PPU.Dispcnt.Mode, &gba.PPU.Backgrounds)
		gba.PPU.objPriorities = gba.getObjPriority(uint32(vcount), &gba.PPU.Objects)
		gba.scanlineGraphics(uint32(vcount))
		gba.PPU.Backgrounds[2].BgAffineUpdate()
		gba.PPU.Backgrounds[3].BgAffineUpdate()
		gba.checkDmas(DMA_MODE_HBL)
	}

	return false
}

func (gba *GBA) ScanlineEndEvent(overshoot int64, arg any) bool {
	dispstat := &gba.Mem.Dispstat
	vcount := &gba.Mem.IO[0x6]

	dispstat.SetHBlank(false)

	*vcount++

	switch *vcount {
	case SCREEN_HEIGHT:
		dispstat.SetVBlank(true)
		gba.checkDmas(DMA_MODE_VBL)
	// bios/bios.gba needs irq set on screen_height, iridion 3d needs screen_height + 1
	// I believe this is cycle related
	case SCREEN_HEIGHT + 1:
		if (*dispstat>>3)&1 != 0 {
			gba.Irq.SetIRQ(0)
		}
	}

	match := dispstat.GetLYC() == *vcount
	dispstat.SetVCFlag(match)

	if vcounterIRQ := (*dispstat>>5)&1 != 0; vcounterIRQ && match {
		gba.Irq.SetIRQ(2)
	}

	gba.Scheduler.schedule(EVENT_END_SCANLINE, 1, CYCLES_SCANLINE-overshoot, gba.ScanlineEndEvent, nil)
	gba.Scheduler.schedule(EVENT_HBK, 1, CYCLES_HDRAW-overshoot, gba.HblankEvent, nil)
	return false
}

func (gba *GBA) FrameEndEvent(overshoot int64, arg any) bool {
	dispstat := &gba.Mem.Dispstat
	vcount := &gba.Mem.IO[0x6]

	gba.Apu.Play(gba.Muted, true)
	gba.Frame++
	gba.Image.WritePixels(gba.Pixels)
	*vcount = 0
	dispstat.SetVBlank(false)

	match := dispstat.GetLYC() == *vcount
	dispstat.SetVCFlag(match)

	if vcounterIRQ := (*dispstat>>5)&1 != 0; vcounterIRQ && match {
		gba.Irq.SetIRQ(2)
	}
	gba.PPU.Backgrounds[2].BgAffineReset()
	gba.PPU.Backgrounds[3].BgAffineReset()

	gba.Scheduler.schedule(EVENT_END_FRAME, 1, CYCLES_FRAME-overshoot, gba.FrameEndEvent, nil)

	return true
}
