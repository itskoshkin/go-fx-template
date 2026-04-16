package datetime

import (
	"fmt"
	"strings"
	"time"
)

func GetExpirationTime(expiration time.Duration) string {
	if expiration <= 0 {
		return "0 seconds"
	}

	type unit struct {
		duration time.Duration
		singular string
		plural   string
	}

	units := []unit{
		{duration: 7 * 24 * time.Hour, singular: "week", plural: "weeks"},
		{duration: 24 * time.Hour, singular: "day", plural: "days"},
		{duration: time.Hour, singular: "hour", plural: "hours"},
		{duration: time.Minute, singular: "minute", plural: "minutes"},
		{duration: time.Second, singular: "second", plural: "seconds"},
	}

	remaining := expiration.Round(time.Second)
	if remaining < time.Second {
		return "1 second"
	}

	parts := make([]string, 0, 2)
	for _, u := range units {
		if remaining < u.duration {
			continue
		}

		count := remaining / u.duration
		remaining -= count * u.duration

		label := u.plural
		if count == 1 {
			label = u.singular
		}

		parts = append(parts, fmt.Sprintf("%d %s", count, label))
		if len(parts) == 2 {
			break
		}
	}

	if len(parts) == 0 {
		return "1 second"
	}

	return strings.Join(parts, " ")
}
