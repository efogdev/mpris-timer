package core

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/efogdev/gotk4-adwaita/pkg/adw"
)

var App *adw.Application

// RegisterApp must be called before UI init
func RegisterApp(ctx context.Context) chan struct{} {
	adw.Init()

	done := make(chan struct{})
	App = adw.NewApplication(AppId, gio.ApplicationNonUnique)
	App.ConnectStartup(func() {
		if Overrides.Color == "default" {
			Overrides.Color = HexFromRGBA(adw.StyleManagerGetDefault().AccentColorRGBA())
			log.Printf("using gtk accent color: %s", Overrides.Color)
		}

		IsPlasma = strings.ToUpper(os.Getenv("XDG_CURRENT_DESKTOP")) == "KDE"
		IsGnome = strings.ToUpper(os.Getenv("XDG_CURRENT_DESKTOP")) == "GNOME"

		ignoreKdeTheme := strings.ToUpper(os.Getenv("PLAY_TIMER_IGNORE_KDE_THEME")) != ""
		if IsPlasma && !ignoreKdeTheme {
			BreezeTheme = true

			if adw.StyleManagerGetDefault().Dark() {
				_ = os.Setenv("GTK_THEME", "Breeze:dark")
			} else {
				_ = os.Setenv("GTK_THEME", "Breeze:light")
			}
		}
	})

	go func() {
		_ = App.Register(ctx)
		done <- struct{}{}
	}()

	return done
}
