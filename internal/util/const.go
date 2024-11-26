package util

import (
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"log"
	"os"
	"path"
	"strings"
	"text/template"
)

const (
	AppId   = "io.github.efogdev.mpris-timer"
	AppName = "Play Timer"
	BaseFPS = 30

	plasmaFPS     = 6
	width         = 128
	height        = 128
	padding       = 8
	strokeWidth   = 16
	roundedOrigin = -86 // -90 is top center. this looks better IMO
	bgStrokeColor = "#535353"
)

const svgTemplate = `
<svg width="{{.Width}}" height="{{.Height}}">
  <style>{{if .HasShadow}}#progress{filter: drop-shadow(-4px 7px 6px rgb(16 16 16 / 0.2));}{{end}}</style>
  <circle cx="{{.CenterX}}" cy="{{.CenterY}}" r="{{.Radius}}" fill="none" stroke="{{.BgStrokeColor}}" stroke-width="{{.BaseWidth}}" />
  <circle id="progress"
		cx="{{.CenterX}}" cy="{{.CenterY}}" r="{{.Radius}}" fill="none" stroke="{{.FgStrokeColor}}"
		stroke-width="{{.StrokeWidth}}" stroke-dasharray="{{.Circumference}}" stroke-dashoffset="{{.DashOffset}}"
		transform="rotate({{if .Rounded}}{{.CustomOrigin}}{{else}}-90{{end}} {{.CenterX}} {{.CenterY}})"
		{{if .Rounded}} stroke-linecap="round"{{end}}
	/>
</svg>`

var (
	IsPlasma bool
	CacheDir string
	DataDir  string
	svgTpl   *template.Template
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
	HasShadow     bool
	Rounded       bool
	CustomOrigin  int
	Progress      int
}

func init() {
	IsPlasma = os.Getenv("XDG_CURRENT_DESKTOP") == "KDE"

	var err error
	svgTpl, err = tpl.Parse(svgTemplate)
	if err != nil {
		log.Println(err)
	}

	DataDir = glib.GetUserDataDir()
	if !strings.Contains(DataDir, AppId) {
		DataDir = path.Join(DataDir, AppId)
	}

	CacheDir, _ = os.UserHomeDir()
	CacheDir = path.Join(CacheDir, ".var", "app", AppId, "cache")

	_ = os.MkdirAll(CacheDir, 0755)
	_ = os.MkdirAll(DataDir, 0755)
}

func CalculateFps() int {
	fps := BaseFPS
	if IsPlasma {
		fps = plasmaFPS
	}
	if Overrides.LowFPS {
		log.Println("lower fps requested")
		fps /= 2
	}

	return fps
}

func bool2int(b bool) int {
	if b {
		return 1
	}
	return 0
}
