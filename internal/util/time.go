package util

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

func TimeFromPreset(preset string) time.Time {
	partsLen := len(strings.Split(preset, ":"))
	switch partsLen {
	case 2:
		result, err := time.Parse("04:05", preset)
		if err != nil {
			log.Fatalf("parse preset %s: %v", preset, err)
		}
		return result
	case 3:
		result, err := time.Parse("15:04:05", preset)
		if err != nil {
			log.Fatalf("parse preset %s: %v", preset, err)
		}
		return result
	default:
		log.Printf("parse preset %s: too many parts", preset)
		return time.Time{}
	}
}

func TimeFromParts(hours int, minutes int, seconds int) time.Time {
	result, err := time.Parse("15:04:05", fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds))
	if err != nil {
		log.Printf("parse parts %d %d %d: %v", hours, minutes, seconds, err)
		return time.Time{}
	}

	return result
}

func TimeFromStrings(hours string, minutes string, seconds string) time.Time {
	hoursInt, err := strconv.Atoi(hours)
	if err != nil {
		log.Printf("parse hours %s: %v", hours, err)
		return time.Time{}
	}

	minutesInt, err := strconv.Atoi(minutes)
	if err != nil {
		log.Printf("parse minutes %s: %v", minutes, err)
		return time.Time{}
	}

	secondsInt, err := strconv.Atoi(seconds)
	if err != nil {
		log.Printf("parse seconds %s: %v", seconds, err)
		return time.Time{}
	}

	return TimeFromParts(hoursInt, minutesInt, secondsInt)
}

func NumToLabelText(num int) string {
	if num > 59 || num < 0 {
		log.Printf("NumToLabelText: num must be between 0 and 59")
		return "00"
	}

	return fmt.Sprintf("%02d", num)
}

// FormatDuration converts a time.Duration to a string in the format "HH:MM:SS" or "MM:SS".
// Hours are only included if the duration is 1 hour or longer.
// The output is zero-padded and rounded to the nearest second.
// Examples:
//   - 1h30m45s -> "01:30:45"
//   - 45m30s   -> "45:30"
//   - 30s      -> "00:30"
//
// Heil hyperoptimization.
// Voices in my head told me to do so.
func FormatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		buf := make([]byte, 8)
		buf[0] = '0' + byte(h/10)
		buf[1] = '0' + byte(h%10)
		buf[2] = ':'
		buf[3] = '0' + byte(m/10)
		buf[4] = '0' + byte(m%10)
		buf[5] = ':'
		buf[6] = '0' + byte(s/10)
		buf[7] = '0' + byte(s%10)
		return string(buf)
	}

	buf := make([]byte, 5)
	buf[0] = '0' + byte(m/10)
	buf[1] = '0' + byte(m%10)
	buf[2] = ':'
	buf[3] = '0' + byte(s/10)
	buf[4] = '0' + byte(s%10)
	return string(buf)
}
