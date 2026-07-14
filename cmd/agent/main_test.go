package main

import (
	"testing"
	"time"

	"github.com/boonkerz/roster/internal/shared"
)

func TestFreqDue(t *testing.T) {
	now := time.Date(2026, 6, 29, 12, 0, 0, 0, time.UTC) // Montag
	zero := time.Time{}

	cases := []struct {
		freq string
		last time.Time
		want bool
	}{
		{"", now.Add(-time.Second), true}, // jeden Checkin
		{"5m", now.Add(-3 * time.Minute), false},
		{"5m", now.Add(-6 * time.Minute), true},
		{"5m", zero, true},
		{"1h", now.Add(-30 * time.Minute), false},
		{"1h", now.Add(-90 * time.Minute), true},
		{"daily", now.Add(-2 * time.Hour), false}, // heute schon gelaufen
		{"daily", now.AddDate(0, 0, -1), true},    // gestern
		{"daily", zero, true},
		{"weekly", now.Add(-2 * time.Hour), false}, // gleiche ISO-Woche (Montag)
		{"weekly", now.AddDate(0, 0, -8), true},    // Vorwoche
		{"monthly", now.AddDate(0, 0, -5), false},  // gleicher Monat
		{"monthly", now.AddDate(0, -1, 0), true},   // Vormonat
		{"yearly", now.AddDate(0, -2, 0), false},   // gleiches Jahr
		{"yearly", now.AddDate(-1, 0, 0), true},    // Vorjahr
	}
	for _, c := range cases {
		if got := freqDue(c.freq, c.last, now, "", ""); got != c.want {
			t.Errorf("freqDue(%q, last %v) = %v, erwartet %v", c.freq, c.last, got, c.want)
		}
	}

	// Frequency hat Vorrang vor Legacy-ScheduleType.
	tk := shared.TaskSpec{Frequency: "1h", ScheduleType: "daily"}
	if !taskDue(tk, now.Add(-2*time.Hour), now) {
		t.Error("taskDue sollte bei Frequency=1h und 2h Abstand fällig sein")
	}
}
