package core

import (
	"bytes"
	_ "embed"
	"github.com/hajimehoshi/go-mp3"
	"github.com/hajimehoshi/oto/v2"
	"io"
	"log"
	"os"
	"time"
)

//go:embed res/ding.mp3
var defaultSound []byte
var sound []byte

func LoadSound() error {
	if Overrides.SoundFilename == "" {
		log.Println("default sound init requested")
		sound = defaultSound
		return nil
	}

	log.Printf("custom sound requested: %s", Overrides.SoundFilename)
	file, err := os.Open(Overrides.SoundFilename)
	if err != nil {
		return err
	}

	sound, err = io.ReadAll(file)
	if err != nil {
		return err
	}

	return nil
}

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
