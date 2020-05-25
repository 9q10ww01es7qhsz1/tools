package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/9q10ww01es7qhsz1/tools/exchangesum/parser"
)

func filterEntries(entries parser.Entries, currency string) parser.Entries {
	var filteredEntries parser.Entries

	for _, entry := range entries {
		if entry.Status != "confirm_payment" {
			continue
		}

		if entry.Currency != currency {
			continue
		}

		filteredEntries = append(filteredEntries, entry)
	}

	return filteredEntries
}

func yearMonthIndex(date time.Time) string {
	return fmt.Sprintf("%d/%d", date.Year(), date.Month())
}

func main() {
	var (
		locName  string
		cur      string
		filename string
		limit    int
	)

	flag.StringVar(&locName, "loc", "Europe/Moscow", "time location")
	flag.StringVar(&cur, "cur", "RUB", "currency")
	flag.StringVar(&filename, "f", "Exchange history.csv", "exchange history file")
	flag.IntVar(&limit, "limit", -1, "limit")
	flag.Parse()

	loc, err := time.LoadLocation(locName)
	if err != nil {
		log.Fatalln(fmt.Errorf("failed to load location: %w", err))
	}

	f, err := os.Open(filename)
	if err != nil {
		log.Fatalln(fmt.Errorf("failed to open file: %w", err))
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.TrimLeadingSpace = true

	entries, err := parser.FromCSV(r, loc)
	if err != nil {
		log.Println(fmt.Errorf("failed to parse entries: %w", err))
		return
	}

	entries = filterEntries(entries, cur)
	sort.Sort(sort.Reverse(entries))

	stats := monthStats{}
	indexes := map[string]int{}

	var total uint64

	for _, entry := range entries {
		mapIdx := yearMonthIndex(entry.Date)

		idx, ok := indexes[mapIdx]
		if !ok {
			if len(stats) == limit {
				break
			}

			stats = append(stats, monthStat{
				Year:  entry.Date.Year(),
				Month: entry.Date.Month(),
				Loc:   loc,
			})

			idx = len(stats) - 1
			indexes[mapIdx] = idx
		}

		stats[idx].Value += entry.FiatAmount
		total += entry.FiatAmount
	}

	sort.Sort(stats)

	var sb strings.Builder

	fmt.Fprintf(&sb, "Currency: %s. Total: %d\n", cur, total)

	for _, m := range stats {
		fmt.Fprintf(&sb, "%d/%s: %d\n", m.Year, m.Month, m.Value)
	}

	fmt.Printf(sb.String())
}
