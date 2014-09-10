package main

import (
	"fmt"
	"time"
)

func formatDuration(duration time.Duration) string {
	var ret string
	var m, h, d, y time.Duration
	if duration < 0 {
		ret += "-"
		duration = -duration
	}
	m = duration / time.Minute
	if m >= 60 {
		h = m / 60
		m = m % 60
	}
	if h >= 24 {
		d = h / 24
		h = h % 24
	}
	if d > 365 {
		y = d / 365
		d = d % 365
	}
	if y > 0 {
		ret += fmt.Sprintf("%dyears.", y)
	}
	if d > 0 {
		ret += fmt.Sprintf("%ddays.", d)
	}
	if h > 0 {
		ret += fmt.Sprintf("%dhours.", h)
	}
	if m > 0 {
		ret += fmt.Sprintf("%dmins.", m)
	}
	return ret
}
