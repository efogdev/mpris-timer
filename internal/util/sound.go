package util

import (
	"bytes"
	_ "embed"
	"github.com/hajimehoshi/go-mp3"
	"github.com/hajimehoshi/oto/v2"
	"time"
)

//go:embed res/ding.mp3
var sound []byte

func PlaySound() error {
	dec, err := mp3.NewDecoder(bytes.NewReader(sound))
	if err != nil {
		return err
	}

	ctx, ready, err := oto.NewContext(dec.SampleRate(), 2, 2)
	if err != nil {
		return err
	}
	<-ready

	player := ctx.NewPlayer(dec)
	defer func() { _ = player.Close() }()
	player.SetVolume(Overrides.Volume)
	player.Play()

	for player.IsPlaying() {
		time.Sleep(10 * time.Millisecond)
	}

	return nil
}
