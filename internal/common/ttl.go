package common

import (
	"time"
)

// TimeToLive return duration from mode.
func TimeToLive(mode string) time.Duration {
	if len(mode) == 0 {
		mode = "default"
	}

	switch mode {
	case "short":
		return time.Minute

	case "long":
		return 5 * time.Minute

	case "normal":
		fallthrough
	case "default":
		fallthrough
	default:
		return 3 * time.Minute
	}
}
