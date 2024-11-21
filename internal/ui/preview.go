package ui

import (
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"log"
	"mpris-timer/internal/util"
	"time"
)

// ToDo: investigate
//
//   - memory leak
//   - occasional freezes, crashes, low (2-3) fps
func renderPreview(box *gtk.Image) {
	tickFor := time.Second * 5                 // 5 seconds timer
	ticker := time.NewTicker(time.Second / 30) // 30 base fps
	defer ticker.Stop()

	go func() {
		if prefsWin == nil {
			return
		}

		prefsWin.ConnectCloseRequest(func() bool {
			ticker.Stop()
			return false
		})
	}()

	// memory leak is terrible w/o this
	imgCache := make(map[string]*gdk.Paintable)
	timeStart := time.Now()
	for range ticker.C {
		if prefsWin == nil || box == nil || !prefsWin.IsVisible() {
			continue
		}

		timePassed := time.Since(timeStart).Microseconds()
		percent := float64(timePassed) / float64(tickFor.Microseconds()) * 100
		if percent >= 100 {
			timeStart = time.Now()
			continue
		}

		imgFilename, err := util.MakeProgressCircle(percent)
		if err != nil {
			log.Printf("render preview: %v", err)
			continue
		}

		cached := imgCache[imgFilename]
		if cached != nil {
			box.SetFromPaintable(cached)
			continue
		}

		img := gtk.NewImageFromFile(imgFilename)
		if img == nil {
			continue
		}

		paintable := img.Paintable()
		if paintable == nil {
			continue
		}

		curImg := paintable.CurrentImage()
		if curImg == nil {
			continue
		}

		imgCache[imgFilename] = curImg
		box.SetFromPaintable(curImg)
	}
}
