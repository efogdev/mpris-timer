package util

import (
	"fmt"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"math"
	"regexp"
)

type Prefs struct {
	ShowPresets        bool
	PresetsOnRight     bool
	Presets            []string
	ProgressColor      string
	EnableSound        bool
	Volume             float64
	EnableNotification bool
	DefaultPreset      string
	DefaultTitle       string
	DefaultText        string
	ActivatePreset     bool
	RememberWindowSize bool
	WindowWidth        uint
	WindowHeight       uint

	// ToDo
	CachePrefix string
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
		EnableSound:        settings.Boolean("enable-sound"),
		Volume:             settings.Double("volume"),
		EnableNotification: settings.Boolean("enable-notification"),
		ShowPresets:        settings.Boolean("show-presets"),
		PresetsOnRight:     settings.Boolean("presets-on-right"),
		Presets:            settings.Strv("presets"),
		ProgressColor:      settings.String("progress-color"),
		DefaultPreset:      settings.String("default-preset"),
		DefaultTitle:       settings.String("default-title"),
		DefaultText:        settings.String("default-text"),
		CachePrefix:        settings.String("cache-prefix"),
		ActivatePreset:     settings.Boolean("activate-preset"),
		RememberWindowSize: settings.Boolean("remember-window-size"),
		WindowWidth:        settings.Uint("window-width"),
		WindowHeight:       settings.Uint("window-height"),
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
	UserPrefs.EnableNotification = value
	settings.SetBoolean("enable-notification", value)
}

func SetActivatePreset(value bool) {
	UserPrefs.ActivatePreset = value
	settings.SetBoolean("activate-preset", value)
}

func SetRememberWindowSize(value bool) {
	UserPrefs.RememberWindowSize = value
	settings.SetBoolean("remember-window-size", value)
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
