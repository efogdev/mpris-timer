package util

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
)

var (
	tpl = template.New("svg")

	cacheLoaded bool
	cacheMu     sync.RWMutex
	cache       = make(map[string]struct{})
)

func InitCache() {
	walk(CacheDir)

	cacheMu.Lock()
	cacheLoaded = true
	cacheMu.Unlock()
}

func MakeProgressCircle(progress float64) (string, error) {
	progress = math.Max(0, math.Min(100, progress))
	dirname := path.Join(CacheDir, strings.ToUpper(strings.Replace(Overrides.Color, "#", "", 1)))
	footprint := fmt.Sprintf("sh%v.r%v.%.2f", bool2int(Overrides.HasShadow), bool2int(Overrides.Rounded), progress)
	filename := path.Join(dirname, footprint+".svg")

	cacheMu.RLock()
	if cacheLoaded {
		_, exists := cache[filename]
		if exists {
			cacheMu.RUnlock()
			return filename, nil
		}
	}
	cacheMu.RUnlock()

	if _, err := os.Stat(filename); err == nil {
		cacheMu.Lock()
		cache[filename] = struct{}{}
		cacheMu.Unlock()
		return filename, nil
	}

	centerX := width / 2
	centerY := height / 2
	radius := float64(width)/2 - float64(strokeWidth) - float64(padding)
	baseWidth := int(math.Round(strokeWidth * 0.25))
	circumference := 2 * math.Pi * radius
	dashOffset := circumference * (1 - progress/100)

	params := svgParams{
		Width:         width,
		Height:        height,
		CenterX:       centerX,
		CenterY:       centerY,
		Radius:        radius,
		BaseWidth:     baseWidth,
		StrokeWidth:   strokeWidth,
		FgStrokeColor: Overrides.Color,
		BgStrokeColor: bgStrokeColor,
		Circumference: circumference,
		DashOffset:    dashOffset,
		HasShadow:     Overrides.HasShadow,
		Rounded:       Overrides.Rounded,
		CustomOrigin:  roundedOrigin,
		Progress:      int(progress),
	}

	var buf bytes.Buffer
	err := svgTpl.Execute(&buf, params)
	if err != nil {
		return "", err
	}

	_ = os.MkdirAll(dirname, 0755)
	err = os.WriteFile(filename, buf.Bytes(), 0644)
	if err != nil {
		return "", fmt.Errorf("write SVG: %w", err)
	}

	cacheMu.Lock()
	cache[filename] = struct{}{}
	cacheMu.Unlock()

	return filename, nil
}

func walk(filename string) {
	_ = filepath.Walk(filename, func(path string, info os.FileInfo, err error) error {
		if err != nil || filename == path {
			return err
		}

		if !info.IsDir() && filepath.Ext(info.Name()) == ".svg" {
			cache[path] = struct{}{}
		}

		if info.IsDir() {
			walk(path)
		}

		return nil
	})
}
