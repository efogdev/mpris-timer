package main

import (
	"context"
	"log"
	"mpris-timer/internal/core"
	"mpris-timer/internal/ui"
	"mpris-timer/internal/util"
	"os"
	"os/signal"
	"runtime/pprof"
	"slices"
	"sync"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	util.RegisterApp(ctx)
	util.LoadPrefs()
	util.LoadFlags()
	util.InitCache()

	// no flag so it doesn't appear in the help
	if slices.Contains(os.Args, "profile") {
		f, err := os.Create("default.pgo")
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close()
		if err = pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	if util.Overrides.UseUI && util.Overrides.Duration > 0 {
		log.Fatalf("UI can't be used with -start")
	}

	// UI by default
	if !util.Overrides.UseUI && util.Overrides.Duration == 0 {
		util.Overrides.UseUI = true
	}

	if util.Overrides.UseUI {
		log.Println("UI requested")
		ui.Init()
	}

	timer, err := core.NewTimerPlayer(util.Overrides.Duration, util.Overrides.Title)
	if err != nil {
		log.Fatalf("failed to create timer: %v", err)
	}

	log.Printf("timer requested: %d sec", util.Overrides.Duration)
	if err = timer.Start(); err != nil {
		log.Fatalf("failed to start timer: %v", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	select {
	case <-timer.Done:
		log.Println("timer done")
		wg := sync.WaitGroup{}

		if util.Overrides.Notify {
			wg.Add(1)
			log.Printf("desktop notification requested")
			go func() {
				ui.Notify(timer.Name, util.Overrides.Text)
				wg.Done()
			}()
		}

		if util.Overrides.Sound {
			wg.Add(1)
			log.Printf("sound requested")
			go func() {
				err = util.PlaySound(false)
				if err != nil {
					log.Printf("error playing sound file: %v", err)
				}
				wg.Done()
			}()
		}

		wg.Wait()
		cancel()
	case <-sigChan:
		cancel()
	case <-ctx.Done():
		timer.Destroy()
		return
	}
}
