package ui

import (
	"fmt"
	"fyne.io/systray"
	"log"
	"mpris-timer/internal/core"
	"os"
)

var (
	trayIcon bool
	progress *systray.MenuItem
	restart  *systray.MenuItem
	play     *systray.MenuItem
	quit     *systray.MenuItem
)

func CreateTrayIcon(timer *core.TimerPlayer) {
	log.Print("tray icon requested")

	// :)
	defer func() {
		recover()
	}()

	if trayIcon {
		log.Print("unexpected: tray icon already initialized")
		return
	}

	initCh := make(chan struct{})
	trayIcon = true
	timer.AddSubscription(func(event core.PropsChangedEvent) {
		if !timer.IsFinished {
			go updateTray(event)
		}
	})

	go systray.Run(func() {
		systray.SetIcon(iconPNG)
		systray.SetTitle(core.Overrides.Title)

		progress = systray.AddMenuItem("Not active", "Current progress")
		systray.AddSeparator()
		play = systray.AddMenuItem("Continue", "Play/pause")
		restart = systray.AddMenuItem("Restart", "Restart timer")
		quit = systray.AddMenuItem("Quit", "Stop timer and quit")
		initCh <- struct{}{}
	}, func() {
		systray.Quit()
	})

	<-initCh
	for {
		select {
		case <-quit.ClickedCh:
			os.Exit(0)
		case <-restart.ClickedCh:
			_ = timer.Previous()
		case <-play.ClickedCh:
			_ = timer.PlayPause()
		}
	}
}

func updateTray(event core.PropsChangedEvent) {
	if !trayIcon {
		return
	}

	// :)
	defer func() {
		recover()
	}()

	if event.IsPaused {
		play.SetTitle("Continue")
	} else {
		play.SetTitle("Pause")
	}

	if event.Img != "" {
		iconBytes, err := core.Pngify(event.Img)
		if err == nil {
			systray.SetIcon(iconBytes)
		}
	}

	progress.SetTitle(fmt.Sprintf("%s: %s", core.Overrides.Title, event.Text))
}
