package util

import (
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"os"
	"path"
	"strings"
)

const (
	AppId   = "io.github.efogdev.mpris-timer"
	AppName = "MPRIS Timer"

	width         = 256
	height        = 256
	padding       = 16
	strokeWidth   = 32
	bgStrokeColor = "#535353"
)

const svgTemplate = `
<svg width="{{.Width}}" height="{{.Height}}">
    <defs>
        <filter id="outer-shadow" x="-50%" y="-50%" width="200%" height="200%">
            <feGaussianBlur in="SourceAlpha" stdDeviation="4" />
            <feOffset dx="1" dy="1" result="offsetblur" />
            <feFlood flood-color="rgba(65, 65, 65, 0.55)" />
            <feComposite in2="offsetblur" operator="in" />
            <feMerge>
                <feMergeNode />
                <feMergeNode in="SourceGraphic" />
            </feMerge>
        </filter>

        <filter id="inner-shadow" x="-50%" y="-50%" width="160%" height="160%">
            <feGaussianBlur in="SourceAlpha" stdDeviation="5" />
            <feOffset dx="0" dy="0" result="offsetblur" />
            <feFlood flood-color="rgba(110, 110, 110, 0.45)" />
            <feComposite in2="offsetblur" operator="in" />
            <feMerge>
                <feMergeNode />
                <feMergeNode in="SourceGraphic" />
            </feMerge>
        </filter>
    </defs>

    <circle cx="{{.CenterX}}" cy="{{.CenterY}}" r="{{.Radius}}" fill="none" stroke="{{.BgStrokeColor}}" stroke-width="{{.BaseWidth}}" filter="url(#outer-shadow)" />
    <circle cx="{{.CenterX}}" cy="{{.CenterY}}" r="{{.Radius}}" fill="none" stroke="{{.FgStrokeColor}}" stroke-width="{{.StrokeWidth}}" stroke-dasharray="{{.Circumference}}" stroke-dashoffset="{{.DashOffset}}" transform="rotate(-90 {{.CenterX}} {{.CenterY}})" filter="url(#inner-shadow)" />
</svg>`

var (
	CacheDir string
	DataDir  string
)

type svgParams struct {
	Width         int
	Height        int
	CenterX       int
	CenterY       int
	Radius        float64
	FgStrokeColor string
	BgStrokeColor string
	BaseWidth     int
	StrokeWidth   int
	Circumference float64
	DashOffset    float64
}

func init() {
	DataDir = glib.GetUserDataDir()
	if !strings.Contains(DataDir, AppId) {
		DataDir = path.Join(DataDir, AppId)
	}

	CacheDir, _ = os.UserHomeDir()
	CacheDir = path.Join(CacheDir, ".var", "app", AppId, "cache")

	_ = os.MkdirAll(CacheDir, 0755)
	_ = os.MkdirAll(DataDir, 0755)
}
