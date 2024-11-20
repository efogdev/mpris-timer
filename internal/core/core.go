package core

import (
	"fmt"
	"github.com/godbus/dbus/v5"
	"log"
	"math"
	"mpris-timer/internal/util"
	"os"
	"strconv"
	"time"
)

const (
	baseFPS      = 30
	baseInterval = time.Second / baseFPS
)

type PropsChangedEvent struct {
	iface string
	props map[string]dbus.Variant
}

type TimerPlayer struct {
	Done           chan struct{}
	Name           string
	serviceName    string
	objectPath     dbus.ObjectPath
	conn           *dbus.Conn
	duration       time.Duration
	startTime      time.Time
	isPaused       bool
	pausedAt       time.Time
	interval       time.Duration
	pausedFor      time.Duration
	tickerDone     chan struct{}
	emitter        chan PropsChangedEvent
	playbackStatus string
}

func NewTimerPlayer(seconds int, name string) (*TimerPlayer, error) {
	if seconds <= 0 {
		return nil, fmt.Errorf("duration must be positive")
	}

	interval := baseInterval
	if seconds > 300 {
		interval = time.Second / (baseFPS / 1.5)
	}

	return &TimerPlayer{
		Name:           name,
		duration:       time.Duration(seconds) * time.Second,
		objectPath:     "/org/mpris/MediaPlayer2",
		playbackStatus: "Playing",
		interval:       interval,
		tickerDone:     make(chan struct{}),
		emitter:        make(chan PropsChangedEvent),
		Done:           make(chan struct{}, 1),
	}, nil
}

func (p *TimerPlayer) Start() error {
	id := strconv.Itoa(int(time.Now().UnixMicro()))[8:]
	conn, err := dbus.SessionBus()
	if err != nil {
		return fmt.Errorf("connect to session bus: %w", err)
	}

	p.conn = conn
	p.serviceName = fmt.Sprintf("org.mpris.MediaPlayer2.%s.run-%s", util.AppId, id)

	reply, err := conn.RequestName(p.serviceName, dbus.NameFlagAllowReplacement)
	if err != nil || reply != dbus.RequestNameReplyPrimaryOwner {
		return fmt.Errorf("request bus: %v", err)
	}

	if err = p.exportInterfaces(); err != nil {
		return fmt.Errorf("export interfaces: %w", err)
	}

	p.startTime = time.Now()
	go p.tick()
	go p.emit()

	return nil
}

func (p *TimerPlayer) Destroy() {
	_ = p.conn.Close()
	close(p.emitter)
	close(p.Done)
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

func (p *TimerPlayer) emit() {
	for payload := range p.emitter {
		p.emitPropertiesChanged(payload.iface, payload.props)
	}
}

func (p *TimerPlayer) tick() {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-p.tickerDone:
			p.Done <- struct{}{}
			p.Destroy()
			return
		case <-ticker.C:
			if p.isPaused {
				continue
			}

			elapsed := time.Since(p.startTime) - p.pausedFor
			progress := math.Min(100, (float64(elapsed)/float64(p.duration))*100)
			if progress == 100 {
				p.Done <- struct{}{}
				return
			}

			timeLeft := p.duration - elapsed
			progressImg, err := util.MakeProgressCircle(progress)
			if err != nil {
				log.Printf("create progress svg: %v", err)
				continue
			}

			p.emitter <- PropsChangedEvent{
				iface: "org.mpris.MediaPlayer2.Player",
				props: map[string]dbus.Variant{
					"PlaybackStatus": dbus.MakeVariant(p.playbackStatus),
					"Metadata": dbus.MakeVariant(map[string]dbus.Variant{
						"mpris:trackid": dbus.MakeVariant(dbus.ObjectPath("/track/1")),
						"xesam:title":   dbus.MakeVariant(p.Name),
						"xesam:artist":  dbus.MakeVariant([]string{util.FormatDuration(timeLeft)}),
						"mpris:artUrl":  dbus.MakeVariant("file://" + progressImg),
					}),
				},
			}
		}
	}
}

func (p *TimerPlayer) emitPropertiesChanged(iface string, changed map[string]dbus.Variant) {
	err := p.conn.Emit(p.objectPath, "org.freedesktop.DBus.Properties.PropertiesChanged",
		iface, changed, []string{})
	if err != nil {
		log.Printf("emit properties: %v", err)
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
			return dbus.MakeVariant(util.AppName), nil
		case "DesktopEntry":
			return dbus.MakeVariant(util.AppId), nil
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
		props["Identity"] = dbus.MakeVariant(util.AppName)
		props["DesktopEntry"] = dbus.MakeVariant(util.AppId)
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
