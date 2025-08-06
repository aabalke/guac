package menu

import (
	"bytes"
	_ "embed"
	"io"
	"log"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
)

const (
	MENU_AUDIO_GOOD     = 0
	MENU_AUDIO_BAD      = 1
	MENU_AUDIO_REALGOOD = 2
)

//go:embed res/sfx_good.mp3
var sfxGood []byte

//go:embed res/sfx_bad.mp3
var sfxBad []byte

//go:embed res/sfx_real_good.mp3
var sfxRealGood []byte

type MenuPlayer struct {
	audioContext                       *audio.Context
	bytesGood, bytesBad, bytesRealGood []byte
	chGood, chBad, chRealGood          chan []byte
}

func NewMenuPlayer(audioContext *audio.Context) (*MenuPlayer, error) {

	player := &MenuPlayer{
		audioContext: audioContext,
		chGood:       make(chan []byte),
		chBad:        make(chan []byte),
		chRealGood:   make(chan []byte),
	}

	go LoadSfx(player.chGood, sfxGood)
	go LoadSfx(player.chBad, sfxBad)
	go LoadSfx(player.chRealGood, sfxRealGood)

	return player, nil
}

func LoadSfx(ch chan []byte, sound []byte) {
	s, err := mp3.DecodeF32(bytes.NewReader(sound))
	if err != nil {
		log.Fatal(err)
		return
	}
	b, err := io.ReadAll(s)
	if err != nil {
		log.Fatal(err)
		return
	}

	ch <- b
}

func (p *MenuPlayer) handleChannels() {
	select {
	case p.bytesGood = <-p.chGood:
		close(p.chGood)
		p.chGood = nil
	case p.bytesBad = <-p.chBad:
		close(p.chBad)
		p.chBad = nil
	case p.bytesRealGood = <-p.chRealGood:
		close(p.chRealGood)
		p.chRealGood = nil
	default:
	}
}

func (p *MenuPlayer) update(idx int) error {

	switch idx {
	case 0:
		p.Play(p.bytesGood)
	case 1:
		p.Play(p.bytesBad)
	case 2:
		p.Play(p.bytesRealGood)
	}

	return nil
}

func (p *MenuPlayer) Play(bytes []byte) {

	if bytes == nil {
		return
	}

	p.audioContext.NewPlayerF32FromBytes(bytes).Play()
}
