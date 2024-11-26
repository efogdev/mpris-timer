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
	stopProf := profile()
	if stopProf != nil {
		defer stopProf()
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	glibDone := util.RegisterApp(ctx)
	util.LoadPrefs()
	util.LoadFlags()
	go util.InitCache()

	if util.Overrides.UseUI && util.Overrides.Duration > 0 {
		log.Fatalf("UI can't be used with -start")
	}

	// UI by default
	if !util.Overrides.UseUI && util.Overrides.Duration == 0 {
		util.Overrides.UseUI = true
	}

	if util.Overrides.UseUI {
		log.Println("UI requested")
		<-glibDone
		ui.Init()
	}

	timer, err := core.NewTimerPlayer(util.Overrides.Duration, util.Overrides.Title)
	if err != nil {
		log.Fatalf("create timer: %v", err)
	}

	log.Printf("timer requested, duration = %d sec", util.Overrides.Duration)
	if err = timer.Start(); err != nil {
		log.Fatalf("start timer: %v", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	select {
	case <-timer.Done:
		log.Println("timer done")
		wg := sync.WaitGroup{}

		if util.Overrides.Notify {
			wg.Add(1)
			log.Printf("notification requested")
			go func() {
				ui.Notify(timer.Name, util.Overrides.Text)
				wg.Done()
			}()
		}

		if util.Overrides.Sound {
			wg.Add(1)
			log.Printf("sound requested")
			go func() {
				err = util.PlaySound()
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

func profile() func() {
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
