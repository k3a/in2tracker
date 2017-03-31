package main

import (
	"fmt"
	"time"
)

type timeDifference struct {
	year, month, day, hour, min, sec int
}

func (td *timeDifference) String() string {
	str := ""
	if td.year > 0 {
		str += fmt.Sprintf("%dy", td.year)
	}
	if td.month > 0 {
		str += fmt.Sprintf("%dm", td.month)
	}
	if td.day > 0 {
		str += fmt.Sprintf("%dd", td.day)
	}
	if td.hour > 0 {
		if len(str) > 0 {
			str += " "
		}
		str += fmt.Sprintf("%dh", td.hour)
	}
	if td.min > 0 {
		str += fmt.Sprintf("%dm", td.min)
	}
	if td.sec > 0 {
		str += fmt.Sprintf("%ds", td.sec)
	}
	return str
}

func TimeDifference(a, b time.Time) *timeDifference {
	retDiff := &timeDifference{}

	if a.Location() != b.Location() {
		b = b.In(a.Location())
	}
	if a.After(b) {
		a, b = b, a
	}
	y1, M1, d1 := a.Date()
	y2, M2, d2 := b.Date()

	h1, m1, s1 := a.Clock()
	h2, m2, s2 := b.Clock()

	retDiff.year = y2 - y1
	retDiff.month = int(M2 - M1)
	retDiff.day = d2 - d1
	retDiff.hour = h2 - h1
	retDiff.min = m2 - m1
	retDiff.sec = s2 - s1

	// Normalize negative values
	if retDiff.sec < 0 {
		retDiff.sec += 60
		retDiff.min--
	}
	if retDiff.min < 0 {
		retDiff.min += 60
		retDiff.hour--
	}
	if retDiff.hour < 0 {
		retDiff.hour += 24
		retDiff.day--
	}
	if retDiff.day < 0 {
		// days in month:
		t := time.Date(y1, M1, 32, 0, 0, 0, 0, time.UTC)
		retDiff.day += 32 - t.Day()
		retDiff.month--
	}
	if retDiff.month < 0 {
		retDiff.month += 12
		retDiff.year--
	}

	return retDiff
}

func TimeDifferenceSince(since time.Time) *timeDifference {
	return TimeDifference(since, time.Now())
}
