package api

import (
	"testing"
	"time"
)

func TestMonthlyCalendarLimits(t *testing.T) {
	now := time.Date(2096, time.February, 29, 12, 0, 0, 0, time.UTC)
	got, err := NextDate(now, "20960229", "m 29 2")
	if err != nil || got != "21040229" {
		t.Fatalf("следующий високосный день: дата %q, ошибка %v", got, err)
	}
	if _, err := NextDate(now, "20960229", "m 31 2,4"); err == nil {
		t.Fatal("для правила с несуществующими датами ожидалась ошибка")
	}
}

func TestNextDateUsesLocalCalendarDay(t *testing.T) {
	zone := time.FixedZone("UTC+3", 3*60*60)
	now := time.Date(2024, time.January, 27, 0, 10, 0, 0, zone)
	got, err := NextDate(now, "20240120", "d 7")
	if err != nil || got != "20240203" {
		t.Fatalf("повторение после полуночи: дата %q, ошибка %v", got, err)
	}
}
