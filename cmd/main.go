package main

import (
	"context"
	"log"
	"mpris-timer/internal/core"
	"mpris-timer/internal/ui"
	"os"
	"os/signal"
	"runtime/pprof"
	"slices"
	"sync"
)

func main() {
	stopProf := profile()
	if stopProf != nil {
		defer stopProf()
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	glibDone := core.RegisterApp(ctx)
	core.LoadPrefs()
	core.LoadFlags()
	go core.InitCache()

	if core.Overrides.Sound {
		go func() { _ = core.LoadSound() }()
	}

	if core.Overrides.UseUI && core.Overrides.Duration > 0 {
		log.Fatalf("UI can't be used with -start")
	}

	// UI by default
	if !core.Overrides.UseUI && core.Overrides.Duration == 0 {
		core.Overrides.UseUI = true
	}

	if core.Overrides.UseUI {
		log.Println("UI requested")
		<-glibDone
		ui.Init()
	}

	timer, err := core.NewTimerPlayer(core.Overrides.Duration, core.Overrides.Title)
	if err != nil {
		log.Fatalf("create timer: %v", err)
	}

	log.Printf("timer requested, duration = %d sec", core.Overrides.Duration)
	if err = timer.Start(); err != nil {
		log.Fatalf("start timer: %v", err)
	}

	if (!core.IsGnome && !core.IsPlasma) || core.Overrides.ForceTrayIcon {
		go ui.CreateTrayIcon(timer)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	select {
	case <-timer.Done:
		log.Println("timer done")
		wg := sync.WaitGroup{}

		if core.Overrides.Notify {
			wg.Add(1)
			log.Printf("notification requested")
			go func() {
				ui.Notify(timer.Name, core.Overrides.Text)
				wg.Done()
			}()
		}

		if core.Overrides.Sound {
			wg.Add(1)
			log.Printf("sound requested")
			go func() {
				err = core.PlaySound()
				if err != nil {
					log.Printf("playing sound: %v", err)
				}
				wg.Done()
			}()
		}

		wg.Wait()
	case <-sigChan:
		timer.Destroy()
	}
}

func profile() (cancel func()) {
	if !slices.Contains(os.Args, "pprof") {
		return nil
	}

	f, err := os.Create("default.pgo")
	if err != nil {
		log.Fatal("create CPU profile: ", err)
	}

	if err = pprof.StartCPUProfile(f); err != nil {
		log.Fatal("start CPU profile: ", err)
	}

	return func() {
		pprof.StopCPUProfile()
		err = f.Close()
		if err != nil {
			log.Fatal("close CPU profile file: ", err)
		}
	}
}
