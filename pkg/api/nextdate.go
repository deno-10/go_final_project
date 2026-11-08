package api

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const dateLayout = "20060102"

// NextDate returns the first recurrence strictly after both the start date and today.
func NextDate(now time.Time, dstart, repeat string) (string, error) {
	start, err := time.Parse(dateLayout, dstart)
	if err != nil {
		return "", fmt.Errorf("некорректная дата задачи: %w", err)
	}
	year, month, day := now.Date()
	today := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	pivot := start
	if today.After(pivot) {
		pivot = today
	}

	if repeat == "" {
		return "", errors.New("не указано правило повторения")
	}
	fields := strings.Split(repeat, " ")
	var next time.Time
	switch fields[0] {
	case "d":
		if len(fields) != 2 {
			return "", errors.New("укажите интервал в днях")
		}
		interval, err := ruleNumber(fields[1], 1, 400)
		if err != nil {
			return "", err
		}
		// Both dates are UTC midnights. Unix seconds also work for spans over 290 years.
		days := (pivot.Unix() - start.Unix()) / (24 * 60 * 60)
		next = pivot.AddDate(0, 0, interval-int(days%int64(interval)))
	case "y":
		if len(fields) != 1 {
			return "", errors.New("у годового правила нет параметров")
		}
		next = start.AddDate(1, 0, 0)
		for !next.After(pivot) {
			next = next.AddDate(1, 0, 0)
		}
	case "w":
		if len(fields) != 2 {
			return "", errors.New("укажите дни недели")
		}
		weekdays, err := ruleSet(fields[1], 1, 7)
		if err != nil {
			return "", err
		}
		next = pivot.AddDate(0, 0, 1)
		for !weekdays[(int(next.Weekday())+6)%7+1] {
			next = next.AddDate(0, 0, 1)
		}
	case "m":
		next, err = nextMonthly(pivot, fields[1:])
		if err != nil {
			return "", err
		}
	default:
		return "", errors.New("неизвестное правило повторения")
	}
	if next.Year() > 9999 {
		return "", errors.New("следующая дата выходит за пределы формата YYYYMMDD")
	}
	return next.Format(dateLayout), nil
}

func ruleNumber(value string, min, max int) (int, error) {
	n, err := strconv.Atoi(value)
	if err != nil || n < min || n > max || n == 0 || strings.HasPrefix(value, "+") {
		return 0, fmt.Errorf("недопустимое значение в правиле повторения: %q", value)
	}
	return n, nil
}

func ruleSet(value string, min, max int) (map[int]bool, error) {
	set := make(map[int]bool)
	for _, item := range strings.Split(value, ",") {
		n, err := ruleNumber(item, min, max)
		if err != nil {
			return nil, err
		}
		set[n] = true
	}
	return set, nil
}

func nextMonthly(pivot time.Time, args []string) (time.Time, error) {
	if len(args) < 1 || len(args) > 2 {
		return time.Time{}, errors.New("укажите дни месяца и при необходимости месяцы")
	}
	days, err := ruleSet(args[0], -2, 31)
	if err != nil {
		return time.Time{}, err
	}
	var months map[int]bool
	if len(args) == 2 {
		months, err = ruleSet(args[1], 1, 12)
		if err != nil {
			return time.Time{}, err
		}
	}

	// Eight years cover the longest gap between leap days, e.g. 2096–2104.
	// This bound also lets us reject impossible rules such as "m 31 4".
	limit := pivot.AddDate(8, 0, 0)
	for date := pivot.AddDate(0, 0, 1); !date.After(limit); date = date.AddDate(0, 0, 1) {
		if months != nil && !months[int(date.Month())] {
			continue
		}
		lastDay := time.Date(date.Year(), date.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
		if days[date.Day()] || days[date.Day()-lastDay-1] {
			return date, nil
		}
	}
	return time.Time{}, errors.New("для заданных дней и месяцев нет подходящей даты")
}
