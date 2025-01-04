package core

import (
	"fmt"
	"github.com/godbus/dbus/v5"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	baseInterval      = time.Millisecond * 5
	lowPresicionAfter = time.Second * 300
)

type PropsChangedEvent struct {
	Text     string
	Img      string
	IsPaused bool
}

type TimerPlayer struct {
	Name           string
	Done           chan struct{}
	IsFinished     bool
	tickerDone     chan struct{}
	emitter        chan PropsChangedEvent
	serviceName    string
	playbackStatus string
	progressText   string
	img            string
	progress       float64
	isPaused       bool
	fps            int
	duration       time.Duration
	startTime      time.Time
	pausedAt       time.Time
	interval       time.Duration
	pausedFor      time.Duration
	objectPath     dbus.ObjectPath
	conn           *dbus.Conn
	subscribers    []func(event PropsChangedEvent)
}

func NewTimerPlayer(seconds int, name string) (*TimerPlayer, error) {
	if seconds <= 0 {
		return nil, fmt.Errorf("duration must be positive")
	}

	fps := CalculateFps()
	interval := baseInterval
	if time.Second*time.Duration(seconds) > lowPresicionAfter {
		log.Printf("low precision requested, duration > %d", lowPresicionAfter)
		interval += interval / 2
	}

	return &TimerPlayer{
		Name:           name,
		duration:       time.Duration(seconds) * time.Second,
		objectPath:     "/org/mpris/MediaPlayer2",
		playbackStatus: "Playing",
		interval:       interval,
		fps:            fps,
		tickerDone:     make(chan struct{}),
		emitter:        make(chan PropsChangedEvent, 1),
		Done:           make(chan struct{}, 1),
	}, nil
}

func (p *TimerPlayer) AddSubscription(onProgress func(event PropsChangedEvent)) {
	p.subscribers = append(p.subscribers, onProgress)
}

func (p *TimerPlayer) Start() error {
	id := strconv.Itoa(int(time.Now().UnixMicro()))[8:]
	conn, err := dbus.SessionBus()
	if err != nil {
		return fmt.Errorf("connect to session bus: %w", err)
	}

	p.conn = conn
	p.serviceName = fmt.Sprintf("org.mpris.MediaPlayer2.%s.run-%s", AppId, id)

	reply, err := conn.RequestName(p.serviceName, dbus.NameFlagAllowReplacement)
	if err != nil || reply != dbus.RequestNameReplyPrimaryOwner {
		return fmt.Errorf("request bus: %v", err)
	}

	if err = p.exportInterfaces(); err != nil {
		return fmt.Errorf("export interfaces: %w", err)
	}

	p.startTime = time.Now()
	go p.runTicker()
	go p.emitLoop()

	return nil
}

func (p *TimerPlayer) Destroy() {
	defer func() { _ = p.conn.Close() }()
	close(p.emitter)

	if p.Done != nil {
		p.Done <- struct{}{}
	}

	close(p.Done)
}

func (p *TimerPlayer) runTicker() {
	log.Printf("player requested, fps = %d", p.fps)

	p.progress = 0.0
	mu := sync.Mutex{}
	img, _ := MakeProgressCircle(0)
	img = "file://" + img
	p.img = img

	renderTicker := time.NewTicker(time.Duration(int64(time.Second) / int64(p.fps)))
	defer renderTicker.Stop()
	go func() {
		for range renderTicker.C {
			mu.Lock()
			img, _ = MakeProgressCircle(p.progress)
			img = "file://" + img
			p.img = img
			mu.Unlock()
		}
	}()

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()
	for {
		select {
		case <-p.tickerDone:
			p.Destroy()
			return
		case <-ticker.C:
			if p.isPaused {
				continue
			}

			elapsed := time.Since(p.startTime) - p.pausedFor
			timeLeft := p.duration - elapsed

			mu.Lock()
			p.progress = math.Min(100, (float64(elapsed)/float64(p.duration))*100)
			if p.progress == 100 {
				p.IsFinished = true
				p.broadcast()
				p.Destroy()
				mu.Unlock()
				return
			}

			p.progressText = FormatDuration(timeLeft)
			p.emitter <- PropsChangedEvent{
				Text:     p.progressText,
				Img:      img,
				IsPaused: p.isPaused,
			}
			mu.Unlock()

			p.broadcast()
		}
	}
}

func (p *TimerPlayer) emitLoop() {
	var prev *PropsChangedEvent

	for e := range p.emitter {
		if prev != nil && prev.Img == e.Img && prev.Text == e.Text {
			continue
		}

		p.emitPropertiesChanged("org.mpris.MediaPlayer2.Player", map[string]dbus.Variant{
			"PlaybackStatus": dbus.MakeVariant(p.playbackStatus),
			"Metadata": dbus.MakeVariant(map[string]dbus.Variant{
				"mpris:trackid": dbus.MakeVariant(dbus.ObjectPath("/track/1")),
				"xesam:title":   dbus.MakeVariant(p.Name),
				"xesam:artist":  dbus.MakeVariant([]string{e.Text}),
				"mpris:artUrl":  dbus.MakeVariant(e.Img),
			}),
		})

		prev = &e
	}
}

func (p *TimerPlayer) exportInterfaces() error {
	if err := p.conn.Export(p, p.objectPath, "org.mpris.MediaPlayer2"); err != nil {
		return err
	}

	if err := p.conn.Export(p, p.objectPath, "org.mpris.MediaPlayer2.Player"); err != nil {
		return err
	}

	if err := p.conn.Export(p, p.objectPath, "org.freedesktop.DBus.Properties"); err != nil {
		return err
	}

	return nil
}

func (p *TimerPlayer) emitPropertiesChanged(iface string, changed map[string]dbus.Variant) {
	err := p.conn.Emit(p.objectPath, "org.freedesktop.DBus.Properties.PropertiesChanged",
		iface, changed, []string{})
	if err != nil {
		log.Printf("emitLoop properties: %v", err)
	}
}

func (p *TimerPlayer) Raise() *dbus.Error { return nil }
func (p *TimerPlayer) Quit() *dbus.Error  { os.Exit(0); return nil }

func (p *TimerPlayer) PlayPause() *dbus.Error {
	if p.isPaused {
		p.pausedFor += time.Since(p.pausedAt)
	} else {
		p.pausedAt = time.Now()
	}

	p.isPaused = !p.isPaused
	p.playbackStatus = map[bool]string{true: "Paused", false: "Playing"}[p.isPaused]

	p.emitPropertiesChanged("org.mpris.MediaPlayer2.Player", map[string]dbus.Variant{
		"PlaybackStatus": dbus.MakeVariant(p.playbackStatus),
	})

	return nil
}

func (p *TimerPlayer) Previous() *dbus.Error {
	p.startTime = time.Now()
	p.pausedFor = 0
	p.isPaused = false
	p.playbackStatus = "Playing"
	return nil
}

func (p *TimerPlayer) Next() *dbus.Error { os.Exit(1); return nil }
func (p *TimerPlayer) Stop() *dbus.Error { os.Exit(1); return nil }

func (p *TimerPlayer) Get(iface, prop string) (dbus.Variant, *dbus.Error) {
	switch iface {
	case "org.mpris.MediaPlayer2":
		switch prop {
		case "Identity":
			return dbus.MakeVariant(AppName), nil
		case "DesktopEntry":
			return dbus.MakeVariant(AppId), nil
		}
	case "org.mpris.MediaPlayer2.Player":
		switch prop {
		case "PlaybackStatus":
			return dbus.MakeVariant(p.playbackStatus), nil
		case "CanGoNext":
			return dbus.MakeVariant(true), nil
		case "CanGoPrevious":
			return dbus.MakeVariant(true), nil
		case "CanPlay":
			return dbus.MakeVariant(true), nil
		case "CanPause":
			return dbus.MakeVariant(true), nil
		case "CanSeek":
			return dbus.MakeVariant(false), nil
		case "CanControl":
			return dbus.MakeVariant(true), nil
		}
	}
	return dbus.Variant{}, nil
}

func (p *TimerPlayer) GetAll(iface string) (map[string]dbus.Variant, *dbus.Error) {
	props := make(map[string]dbus.Variant)
	switch iface {
	case "org.mpris.MediaPlayer2":
		props["Identity"] = dbus.MakeVariant(AppName)
		props["DesktopEntry"] = dbus.MakeVariant(AppId)
	case "org.mpris.MediaPlayer2.Player":
		props["PlaybackStatus"] = dbus.MakeVariant(p.playbackStatus)
		props["CanGoNext"] = dbus.MakeVariant(true)
		props["CanGoPrevious"] = dbus.MakeVariant(true)
		props["CanPlay"] = dbus.MakeVariant(true)
		props["CanPause"] = dbus.MakeVariant(true)
		props["CanSeek"] = dbus.MakeVariant(false)
		props["CanControl"] = dbus.MakeVariant(true)
	}
	return props, nil
}

func (p *TimerPlayer) Set(_, _ string, _ dbus.Variant) *dbus.Error {
	return nil
}

func (p *TimerPlayer) broadcast() {
	if p.progress >= 100 {
		return
	}

	for _, cb := range p.subscribers {
		if cb == nil {
			continue
		}

		cb(PropsChangedEvent{
			Img:      strings.TrimPrefix(p.img, "file://"),
			Text:     p.progressText,
			IsPaused: p.isPaused,
		})
	}
}
