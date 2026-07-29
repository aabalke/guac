package gba

import (
	"github.com/aabalke/guac/config"
	"github.com/aabalke/guac/emu/gba/gpio"
)

var rtcCards = map[string]bool{
	"AXVE": true, /* Pokemon - Ruby Version (USA, Europe) */
	"AXPE": true, /* Pokemon - Sapphire Version (USA, Europe) */
	"BPEE": true, /* Pokemon - Emerald Version (USA, Europe) */
	"AGFE": true, /* Golden Sun - The Lost Age (USA) */
	"AGSE": true, /* Golden Sun (USA) */
	"AXPJ": true, /* Pocket Monsters - Sapphire (Japan) */
	"AXVJ": true, /* Pocket Monsters - Ruby (Japan) */
	"BKAJ": true, /* Sennen Kazoku (Japan) */
	"BPEJ": true, /* Pocket Monsters - Emerald (Japan) */
	"BR4J": true, /* Rockman EXE 4.5 - Real Operation (Japan) */
	"AXPF": true, /* Pokemon - Version Saphir (France) */
	"AXVF": true, /* Pokemon - Version Rubis (France) */
	"BPEF": true, /* Pokemon - Version Emeraude (France) */
	"AXPI": true, /* Pokemon - Versione Zaffiro (Italy) */
	"AXVI": true, /* Pokemon - Versione Rubino (Italy) */
	"BPEI": true, /* Pokemon - Versione Smeraldo (Italy) */
	"AXPD": true, /* Pokemon - Saphir-Edition (Germany) */
	"AXVD": true, /* Pokemon - Rubin-Edition (Germany) */
	"BPED": true, /* Pokemon - Smaragd-Edition (Germany) */
	"AXPS": true, /* Pokemon - Edicion Zafiro (Spain) */
	"AXVS": true, /* Pokemon - Edicion Rubi (Spain) */
	"BPES": true, /* Pokemon - Edicion Esmeralda (Spain) */
	"U3IP": true, /* Boktai - The Sun Is in Your Hand (Europe)(En,Fr,De,Es,It) */
	"U32P": true, /* Boktai 2 - Solar Boy Django (Europe)(En,Fr,De,Es,It) */
	"U3IE": true, /* Boktai - The Sun Is in Your Hand (USA) */
	"U32E": true, /* Boktai 2 - Solar Boy Django (USA) */
	"U3IJ": true, /* Bokura no Taiyou - Taiyou Action RPG (Japan) */
	"U32J": true, /* Zoku Bokura no Taiyou - Taiyou Shounen Django (Japan) */
	"U33J": true, /* Shin Bokura no Taiyou - Gyakushuu no Sabata (Japan) */
}

var solarCards = map[string]bool{
	"U3IP": true, /* Boktai - The Sun Is in Your Hand (Europe)(En,Fr,De,Es,It) */
	"U32P": true, /* Boktai 2 - Solar Boy Django (Europe)(En,Fr,De,Es,It) */
	"U3IE": true, /* Boktai - The Sun Is in Your Hand (USA) */
	"U32E": true, /* Boktai 2 - Solar Boy Django (USA) */
	"U3IJ": true, /* Bokura no Taiyou - Taiyou Action RPG (Japan) */
	"U32J": true, /* Zoku Bokura no Taiyou - Taiyou Shounen Django (Japan) */
	"U33J": true, /* Shin Bokura no Taiyou - Gyakushuu no Sabata (Japan) */
}

func (gba *GBA) AddGpios() {
	if solarCards[gba.Cartridge.Code] || config.Conf.Gba.SpecialHardware.ForceSolarSensor {
		if gba.Mem.Gpio == nil {
			gba.Mem.Gpio = gpio.NewGpio()
		}

		level := uint8(config.Conf.Gba.SpecialHardware.SolarSensorLevel)
		gba.Mem.Gpio.Devices = append(gba.Mem.Gpio.Devices, gpio.NewSolar(level))
	}

	if rtcCards[gba.Cartridge.Code] || config.Conf.Gba.SpecialHardware.ForceRtc {
		if gba.Mem.Gpio == nil {
			gba.Mem.Gpio = gpio.NewGpio()
		}

		gba.Mem.Gpio.Devices = append(gba.Mem.Gpio.Devices, gpio.NewRtc(gba.Irq))
	}
}
