package main

import (
	"time"
)

type monthStat struct {
	Year  int
	Month time.Month
	Loc   *time.Location
	Value uint64
}

type monthStats []monthStat

func (s monthStats) Len() int {
	return len(s)
}

func (s monthStats) Less(i, j int) bool {
	x := time.Date(s[i].Year, s[i].Month, 0, 0, 0, 0, 0, s[i].Loc)
	y := time.Date(s[j].Year, s[j].Month, 0, 0, 0, 0, 0, s[j].Loc)
	return x.Before(y)
}

func (s monthStats) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
