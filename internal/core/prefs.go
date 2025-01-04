package core

import (
	"fmt"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"math"
	"regexp"
)

type Prefs struct {
	ShowPresets     bool
	PresetsOnRight  bool
	Presets         []string
	ProgressColor   string
	EnableSound     bool
	Volume          float64
	ShouldNotify    bool
	DefaultPreset   string
	DefaultTitle    string
	DefaultText     string
	SoundFilename   string
	ActivatePreset  bool
	RememberWinSize bool
	ForceTrayIcon   bool
	Shadow          bool
	Rounded         bool
	LowFPS          bool
	ShowTitle       bool
	WindowWidth     uint
	WindowHeight    uint
}

var (
	UserPrefs Prefs
	settings  *gio.Settings
)

func LoadPrefs() {
	if settings == nil {
		settings = gio.NewSettings(AppId)
	}

	UserPrefs = Prefs{
		EnableSound:     settings.Boolean("enable-sound"),
		Volume:          settings.Double("volume"),
		ShouldNotify:    settings.Boolean("enable-notification"),
		ShowPresets:     settings.Boolean("show-presets"),
		PresetsOnRight:  settings.Boolean("presets-on-right"),
		Presets:         settings.Strv("presets"),
		ProgressColor:   settings.String("progress-color"),
		DefaultPreset:   settings.String("default-preset"),
		DefaultTitle:    settings.String("default-title"),
		DefaultText:     settings.String("default-text"),
		SoundFilename:   settings.String("sound-filename"),
		ActivatePreset:  settings.Boolean("activate-preset"),
		RememberWinSize: settings.Boolean("remember-window-size"),
		Shadow:          settings.Boolean("shadow"),
		Rounded:         settings.Boolean("rounded"),
		LowFPS:          settings.Boolean("low-fps"),
		ForceTrayIcon:   settings.Boolean("force-tray-icon"),
		ShowTitle:       settings.Boolean("show-title"),
		WindowWidth:     settings.Uint("window-width"),
		WindowHeight:    settings.Uint("window-height"),
	}
}

func HexFromRGBA(rgba *gdk.RGBA) string {
	r := int(math.Round(float64(rgba.Red()) * 255))
	g := int(math.Round(float64(rgba.Green()) * 255))
	b := int(math.Round(float64(rgba.Blue()) * 255))

	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}

// RGBAFromHex assumes the value is correct and ignores alpha
func RGBAFromHex(hex string) (*gdk.RGBA, error) {
	rgba := gdk.NewRGBA(0, 0, 0, 255)
	ok := rgba.Parse(hex)
	if !ok {
		return nil, fmt.Errorf("invalid hex string")
	}

	return &rgba, nil
}

func SetShowPresets(value bool) {
	UserPrefs.ShowPresets = value
	settings.SetBoolean("show-presets", value)
}

func SetPresetsOnRight(value bool) {
	UserPrefs.PresetsOnRight = value
	settings.SetBoolean("presets-on-right", value)
}

func SetEnableSound(value bool) {
	Overrides.Sound = true
	UserPrefs.EnableSound = value
	settings.SetBoolean("enable-sound", value)
}

func SetEnableNotification(value bool) {
	Overrides.Notify = true
	UserPrefs.ShouldNotify = value
	settings.SetBoolean("enable-notification", value)
}

func SetActivatePreset(value bool) {
	UserPrefs.ActivatePreset = value
	settings.SetBoolean("activate-preset", value)
}

func SetRememberWindowSize(value bool) {
	UserPrefs.RememberWinSize = value
	settings.SetBoolean("remember-window-size", value)
}

func SetShadow(value bool) {
	Overrides.HasShadow = value
	UserPrefs.Shadow = value
	settings.SetBoolean("shadow", value)
}

func SetRounded(value bool) {
	Overrides.Rounded = value
	UserPrefs.Rounded = value
	settings.SetBoolean("rounded", value)
}

func SetShowTitle(value bool) {
	UserPrefs.ShowTitle = value
	settings.SetBoolean("show-title", value)
}

func SetLowFPS(value bool) {
	Overrides.LowFPS = value
	UserPrefs.LowFPS = value
	settings.SetBoolean("low-fps", value)
}

func SetForceTrayIcon(value bool) {
	Overrides.ForceTrayIcon = value
	UserPrefs.ForceTrayIcon = value
	settings.SetBoolean("force-tray-icon", value)
}

func SetProgressColor(value string) {
	if !regexp.MustCompile(`^#([0-9A-Fa-f]{3}|[0-9A-Fa-f]{6})$`).MatchString(value) {
		return
	}

	Overrides.Color = value
	UserPrefs.ProgressColor = value
	settings.SetString("progress-color", value)
}

func SetPresets(value []string) {
	UserPrefs.Presets = value
	settings.SetStrv("presets", value)
}

func SetDefaultPreset(value string) {
	UserPrefs.DefaultPreset = value
	settings.SetString("default-preset", value)
}

func SetDefaultTitle(value string) {
	Overrides.Title = value
	UserPrefs.DefaultTitle = value
	settings.SetString("default-title", value)
}

func SetSoundFilename(value string) {
	Overrides.SoundFilename = value
	UserPrefs.SoundFilename = value
	settings.SetString("sound-filename", value)

	err := LoadSound()
	if err != nil {
		SetSoundFilename("")
	}
}

func SetDefaultText(value string) {
	Overrides.Text = value
	UserPrefs.DefaultText = value
	settings.SetString("default-text", value)
}

func SetVolume(value float64) {
	Overrides.Volume = value
	UserPrefs.Volume = value
	settings.SetDouble("volume", value)
}

func SetWindowSize(width uint, height uint) {
	UserPrefs.WindowWidth = width
	UserPrefs.WindowHeight = height
	settings.SetUint("window-width", width)
	settings.SetUint("window-height", height)
}
