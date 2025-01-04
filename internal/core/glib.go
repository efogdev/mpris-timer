package core

import (
	"context"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/efogdev/gotk4-adwaita/pkg/adw"
)

var App *adw.Application

// RegisterApp must be called before UI init
func RegisterApp(ctx context.Context) chan struct{} {
	done := make(chan struct{})
	App = adw.NewApplication(AppId, gio.ApplicationNonUnique)

	go func() {
		_ = App.Register(ctx)
		done <- struct{}{}
	}()

	return done
}
