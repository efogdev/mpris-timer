package core

import (
	"bytes"
	"fmt"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
	"image"
	"image/png"
	"log"
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

	// unfortunately PNG is needed for tray icon
	pngCacheLoaded bool
	pngCacheMu     sync.RWMutex
	pngCache       = make(map[string]struct{})
)

func InitCache() {
	walk(CacheDir)

	cacheMu.Lock()
	cacheLoaded = true
	cacheMu.Unlock()

	pngCacheMu.Lock()
	pngCacheLoaded = true
	pngCacheMu.Unlock()
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
	if cacheLoaded {
		cache[filename] = struct{}{}
	}
	cacheMu.Unlock()

	return filename, nil
}

func Pngify(filename string) ([]byte, error) {
	pngFilename := filename + ".png"

	pngCacheMu.RLock()
	if pngCacheLoaded {
		_, exists := pngCache[pngFilename]
		if exists {
			pngCacheMu.RUnlock()
			out, err := os.ReadFile(pngFilename)
			if err != nil {
				return nil, err
			}
			return out, nil
		}
	}
	pngCacheMu.RUnlock()

	in, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() { _ = in.Close() }()

	icon, err := oksvg.ReadIconStream(in)
	if err != nil {
		return nil, err
	}

	icon.SetTarget(0, 0, float64(width), float64(width))
	fullImage := image.NewRGBA(image.Rect(0, 0, width, width))
	icon.Draw(rasterx.NewDasher(width, width, rasterx.NewScannerGV(width, width, fullImage, fullImage.Bounds())), 1)

	center := width / 2
	cropTo := 96
	halfCrop := cropTo / 2

	cropRect := image.Rect(center-halfCrop, center-halfCrop, center+halfCrop, center+halfCrop)
	croppedImage := fullImage.SubImage(cropRect).(*image.RGBA)

	// ToDo I don't even know why it's mirrored bruh
	mirroredImage := image.NewRGBA(image.Rect(0, 0, cropTo, cropTo))
	for x := 0; x < cropTo; x++ {
		for y := 0; y < cropTo; y++ {
			mirroredImage.Set(cropTo-x-1, y, croppedImage.At(croppedImage.Bounds().Min.X+x, croppedImage.Bounds().Min.Y+y))
		}
	}

	out := bytes.Buffer{}
	err = png.Encode(&out, mirroredImage)
	if err != nil {
		return nil, err
	}

	go func() {
		err := os.WriteFile(pngFilename, out.Bytes(), 0644)
		if err != nil {
			log.Printf("writing PNG cache: %v", err)
		}

		pngCacheMu.Lock()
		if pngCacheLoaded {
			pngCache[pngFilename] = struct{}{}
		}
		pngCacheMu.Unlock()
	}()

	return out.Bytes(), nil
}

func walk(filename string) {
	_ = filepath.Walk(filename, func(path string, info os.FileInfo, err error) error {
		if err != nil || filename == path {
			return err
		}

		if info.IsDir() {
			walk(path)
		} else {
			ext := filepath.Ext(info.Name())
			switch ext {
			case ".svg":
				cacheMu.Lock()
				cache[path] = struct{}{}
				cacheMu.Unlock()
			case ".png":
				pngCacheMu.Lock()
				pngCache[path] = struct{}{}
				pngCacheMu.Unlock()
			}
		}

		return nil
	})
}
