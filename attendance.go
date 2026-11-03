package main

import (
	"math/rand/v2"
	"time"
)

// Fixed seed keeps the preview stable between page reloads.
func sampleAttendance(subjects []Subject) []AttendanceRecord {
	rng := rand.New(rand.NewPCG(27, 2026))
	var records []AttendanceRecord
	for _, subject := range subjects {
		date := time.Date(2026, time.October, 26, 12, 0, 0, 0, time.UTC)
		for count := 0; count < 20; date = date.AddDate(0, 0, -1) {
			if date.Weekday() == time.Friday || date.Weekday() == time.Saturday {
				continue
			}
			status := "present"
			switch roll := rng.IntN(10); {
			case roll < 2:
				status = "absent"
			case roll == 2:
				status = "late"
			}
			records = append(records, AttendanceRecord{Date: date.Format("2006-01-02"), Subject: subject.Name, Status: status})
			count++
		}
	}
	return records
}
