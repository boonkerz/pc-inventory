package main

import (
	"testing"
	"time"

	"github.com/thomaspeterson/pc-inventory/internal/shared"
)

func TestParseHHMM(t *testing.T) {
	if m, ok := parseHHMM("02:30"); !ok || m != 150 {
		t.Errorf("02:30 => %d,%v erwartet 150,true", m, ok)
	}
	if _, ok := parseHHMM("25:00"); ok {
		t.Error("25:00 sollte ungültig sein")
	}
	if _, ok := parseHHMM("abc"); ok {
		t.Error("abc sollte ungültig sein")
	}
}

func TestWeekdayAllowed(t *testing.T) {
	if !weekdayAllowed("", time.Monday) {
		t.Error("leer = alle Tage erlaubt")
	}
	if !weekdayAllowed("1,3,5", time.Wednesday) { // 3 = Mittwoch
		t.Error("Mittwoch (3) sollte erlaubt sein")
	}
	if weekdayAllowed("1,3,5", time.Tuesday) { // 2 = Dienstag
		t.Error("Dienstag (2) sollte nicht erlaubt sein")
	}
}

func TestTaskDueInterval(t *testing.T) {
	now := time.Date(2026, 6, 28, 12, 0, 0, 0, time.Local)
	task := shared.TaskSpec{ScheduleType: "interval", IntervalMinutes: 60}
	if !taskDue(task, time.Time{}, now) {
		t.Error("erstmals (last=zero) sollte fällig sein")
	}
	if taskDue(task, now.Add(-30*time.Minute), now) {
		t.Error("vor 30 Min gelaufen, Intervall 60 -> nicht fällig")
	}
	if !taskDue(task, now.Add(-90*time.Minute), now) {
		t.Error("vor 90 Min gelaufen, Intervall 60 -> fällig")
	}
}

func TestTaskDueDaily(t *testing.T) {
	// Sonntag, 28.6.2026, 03:00 lokal
	now := time.Date(2026, 6, 28, 3, 0, 0, 0, time.Local)
	task := shared.TaskSpec{ScheduleType: "daily", DailyTime: "02:00"}
	if !taskDue(task, time.Time{}, now) {
		t.Error("nach 02:00 und heute noch nicht gelaufen -> fällig")
	}
	// heute schon gelaufen (um 02:30)
	already := time.Date(2026, 6, 28, 2, 30, 0, 0, time.Local)
	if taskDue(task, already, now) {
		t.Error("heute bereits gelaufen -> nicht fällig")
	}
	// vor der Uhrzeit
	before := time.Date(2026, 6, 28, 1, 0, 0, 0, time.Local)
	if taskDue(task, time.Time{}, before) {
		t.Error("vor 02:00 -> nicht fällig")
	}
	// gestern gelaufen, jetzt nach Uhrzeit -> fällig
	yesterday := time.Date(2026, 6, 27, 2, 30, 0, 0, time.Local)
	if !taskDue(task, yesterday, now) {
		t.Error("gestern gelaufen, heute nach 02:00 -> fällig")
	}
}
