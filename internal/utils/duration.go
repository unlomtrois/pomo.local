package utils

import (
	"strings"
	"time"
)

// ShortDuration returns the most compact string representation.
//
// Examples: 1h0m0s -> 1h, 2m0s -> 2m, 0s -> 0s
func ShortDuration(d time.Duration) string {
	s := d.String()
	// Only trim if it's longer than a simple "30s" or "5s"
	// and ends in the "0s" or "0m" pattern.
	if len(s) > 3 {
		if d%time.Minute == 0 {
			s = strings.TrimSuffix(s, "0s")
		}
		if d%time.Hour == 0 {
			s = strings.TrimSuffix(s, "0m")
		}
	}
	return s
}
