package ui

import (
	"context"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/efogdev/gotk4-adwaita/pkg/adw"
	"github.com/google/uuid"
	"log"
	"mpris-timer/internal/core"
)

func Notify(title string, text string) {
	log.Printf("notify: %s", title)

	if !core.Overrides.UseUI {
		sendNotification(core.App, title, text)
	} else {
		nApp := adw.NewApplication(core.AppId, gio.ApplicationNonUnique)
		nApp.ConnectActivate(func() {
			sendNotification(nApp, title, text)
		})

		_ = nApp.Register(context.Background())
		nApp.Run(nil)
	}
}

func sendNotification(app *adw.Application, title string, text string) {
	id, _ := uuid.NewV7()
	actionName := "app." + id.String()
	app.AddAction(gio.NewSimpleAction(actionName, nil))

	n := gio.NewNotification(title)
	n.SetBody(text)
	n.SetPriority(gio.NotificationPriorityUrgent)
	n.SetDefaultAction(actionName)
	n.SetIcon(gio.NewBytesIcon(glib.NewBytes(icon)))

	app.SendNotification(id.String(), n)
}
