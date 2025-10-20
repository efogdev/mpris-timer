package core

import (
	"flag"
)

var Overrides = struct {
	Notify        bool
	Sound         bool
	Volume        float64
	UseUI         bool
	Duration      int
	Title         string
	Text          string
	Color         string
	HasShadow     bool
	Rounded       bool
	LowFPS        bool
	ForceTrayIcon bool
	SoundFilename string
}{}

func LoadFlags() {
	flag.BoolVar(&Overrides.Notify, "notify", UserPrefs.ShouldNotify, "Send desktop notification")
	flag.BoolVar(&Overrides.Sound, "sound", UserPrefs.EnableSound, "Play sound")
	flag.StringVar(&Overrides.SoundFilename, "soundfile", UserPrefs.SoundFilename, "Filename of the custom sound (must be .mp3)")
	flag.Float64Var(&Overrides.Volume, "volume", UserPrefs.Volume, "Volume [0-1]")
	flag.BoolVar(&Overrides.UseUI, "ui", false, "Show timepicker UI (default true)")
	flag.BoolVar(&Overrides.HasShadow, "shadow", UserPrefs.Shadow, "Shadow for progress image")
	flag.BoolVar(&Overrides.Rounded, "rounded", UserPrefs.Rounded, "Rounded corners")
	flag.BoolVar(&Overrides.LowFPS, "lowfps", UserPrefs.LowFPS, "1 fps mode (energy saver, GNOME only)")
	flag.IntVar(&Overrides.Duration, "start", 0, "Start the timer immediately, don't show UI (value in seconds)")
	flag.StringVar(&Overrides.Title, "title", UserPrefs.DefaultTitle, "Name/title of the timer")
	flag.StringVar(&Overrides.Text, "text", UserPrefs.DefaultText, "Notification text")
	flag.StringVar(&Overrides.Color, "color", UserPrefs.ProgressColor, "Progress color (#HEX) for the player, use \"default\" for the GTK accent color")
	flag.BoolVar(&Overrides.ForceTrayIcon, "tray", UserPrefs.ForceTrayIcon, "Force tray icon presence")
	flag.Parse()
}
