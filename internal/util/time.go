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
		log.Fatalf("parse preset %s: too many parts", preset)
		return time.Time{}
	}
}

func TimeFromParts(hours int, minutes int, seconds int) time.Time {
	result, err := time.Parse("15:04:05", fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds))
	if err != nil {
		log.Fatalf("parse parts %d %d %d: %v", hours, minutes, seconds, err)
	}

	return result
}

func TimeFromStrings(hours string, minutes string, seconds string) time.Time {
	hoursInt, err := strconv.Atoi(hours)
	if err != nil {
		log.Fatalf("parse hours %s: %v", hours, err)
	}

	minutesInt, err := strconv.Atoi(minutes)
	if err != nil {
		log.Fatalf("parse minutes %s: %v", minutes, err)
	}

	secondsInt, err := strconv.Atoi(seconds)
	if err != nil {
		log.Fatalf("parse seconds %s: %v", seconds, err)
	}

	return TimeFromParts(hoursInt, minutesInt, secondsInt)
}

func FormatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	}

	return fmt.Sprintf("%02d:%02d", m, s)
}

func NumToLabelText(num int) string {
	if num > 59 || num < 0 {
		log.Fatalf("NumToLabelText: num must be between 0 and 59")
	}

	return fmt.Sprintf("%02d", num)
}
