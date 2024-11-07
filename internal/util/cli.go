package util

import (
	"flag"
	"github.com/efogdev/gotk4-adwaita/pkg/adw"
	"log"
)

type overrides struct {
	Notify   bool
	Sound    bool
	Volume   float64
	Silence  int
	UseUI    bool
	Duration int
	Title    string
	Text     string
	Color    string
}

var Overrides = overrides{}

func LoadFlags() {
	flag.BoolVar(&Overrides.Notify, "notify", UserPrefs.EnableNotification, "Send desktop notification")
	flag.BoolVar(&Overrides.Sound, "sound", UserPrefs.EnableSound, "Play sound")
	flag.Float64Var(&Overrides.Volume, "volume", UserPrefs.Volume, "Volume [0-1]")
	flag.IntVar(&Overrides.Silence, "silence", 0, "Play this milliseconds of silence before the actual audio — might be helpful for audio devices that wake up not immediately")
	flag.BoolVar(&Overrides.UseUI, "ui", false, "Show timepicker UI (default true)")
	flag.IntVar(&Overrides.Duration, "start", 0, "Start the timer immediately")
	flag.StringVar(&Overrides.Title, "title", UserPrefs.DefaultTitle, "Name/title of the timer")
	flag.StringVar(&Overrides.Text, "text", UserPrefs.DefaultText, "Notification text")
	flag.StringVar(&Overrides.Color, "color", UserPrefs.ProgressColor, "Progress color for the player")
	flag.Parse()

	if Overrides.Color == "default" {
		Overrides.Color = HexFromRGBA(adw.StyleManagerGetDefault().AccentColorRGBA())
		log.Printf("using gtk accent color: %s", Overrides.Color)
	}
}
