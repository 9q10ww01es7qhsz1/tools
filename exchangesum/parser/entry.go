package parser

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"time"
)

const (
	TimeLayout = "02-01-2006 15:04:05"
)

type Entry struct {
	ID             uint64
	Date           time.Time
	CryptoAmount   float64
	FiatAmount     uint64
	Fee            float64
	ExchangeRate   float64
	Currency       string
	Cryptocurrency string
	PayMethod      string
	Buyer          string
	Seller         string
	Type           string
	Status         string
}

type Entries []Entry

func (e Entries) Len() int {
	return len(e)
}

func (e Entries) Less(i, j int) bool {
	return e[i].Date.Before(e[j].Date)
}

func (e Entries) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

func FromCSV(r *csv.Reader, loc *time.Location) (Entries, error) {
	var (
		entries []Entry
		header  = true
	)

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read record: %w", err)
		}

		if header {
			header = false
			continue
		}

		e, err := entryFromRecord(record, loc)
		if err != nil {
			return nil, fmt.Errorf("failed to convert record to Entry: %w", err)
		}

		entries = append(entries, e)
	}

	return entries, nil
}

func entryFromRecord(record []string, loc *time.Location) (Entry, error) {
	var (
		e   Entry
		err error
	)

	if len(record) != 13 {
		return e, fmt.Errorf("unexpected record length: %d", len(record))
	}

	e.ID, err = strconv.ParseUint(record[1], 10, 64)
	if err != nil {
		return e, fmt.Errorf("failed to parse entry ID: %w", err)
	}

	e.Date, err = time.ParseInLocation(TimeLayout, record[0], loc)
	if err != nil {
		return e, fmt.Errorf("failed to parse date: %w", err)
	}

	e.CryptoAmount, err = strconv.ParseFloat(record[2], 64)
	if err != nil {
		return e, fmt.Errorf("failed to parse crypto amount: %w", err)
	}

	e.FiatAmount, err = strconv.ParseUint(record[3], 10, 64)
	if err != nil {
		return e, fmt.Errorf("failed to parse fiat amount: %w", err)
	}

	e.Fee, err = strconv.ParseFloat(record[4], 64)
	if err != nil {
		return e, fmt.Errorf("failed to parse fee: %w", err)
	}

	e.ExchangeRate, err = strconv.ParseFloat(record[5], 64)
	if err != nil {
		return e, fmt.Errorf("failed to parse exchange rate: %w", err)
	}

	e.Currency = record[6]
	e.Cryptocurrency = record[7]
	e.PayMethod = record[8]
	e.Buyer = record[9]
	e.Seller = record[10]
	e.Type = record[11]
	e.Status = record[12]

	return e, nil
}
