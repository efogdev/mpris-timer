package util

import (
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"strings"
)

type Key uint

const (
	KeyEnter Key = iota
	KeySpace
	KeyLeft
	KeyRight
	KeyUp
	KeyDown
	KeyEsc
)

func (k Key) GdkKeyvals() []uint {
	switch k {
	case KeyEnter:
		return []uint{gdk.KEY_Return, gdk.KEY_KP_Enter, gdk.KEY_ISO_Enter, gdk.KEY_3270_Enter, gdk.KEY_RockerEnter}
	case KeySpace:
		return []uint{gdk.KEY_KP_Space, gdk.KEY_space}
	case KeyLeft:
		return []uint{gdk.KEY_KP_Left, gdk.KEY_Left}
	case KeyRight:
		return []uint{gdk.KEY_KP_Right, gdk.KEY_Right}
	case KeyUp:
		return []uint{gdk.KEY_KP_Up, gdk.KEY_Up}
	case KeyDown:
		return []uint{gdk.KEY_KP_Down, gdk.KEY_Down}
	case KeyEsc:
		return []uint{gdk.KEY_Escape}
	default:
		panic("unknown key")
	}
}

func ParseKeyval(keyval uint) string {
	return strings.ReplaceAll(gdk.KeyvalName(keyval), "KP_", "")
}

func IsGdkKeyvalNumber(keyval uint) bool {
	return (keyval >= gdk.KEY_0 && keyval <= gdk.KEY_9) || (keyval >= gdk.KEY_KP_0 && keyval <= gdk.KEY_KP_9)
}
