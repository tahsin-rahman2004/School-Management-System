// admissions.go - manages class seat vacancies and online admission applications
// submitted from admission-form.html.
// Seat counts are stored in data/seats.json and applications in data/admissions.json.
package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// SeatConfig holds the available seats for each class and Class 9 group.
type SeatConfig struct {
	Classes map[string]int `json:"classes"`
	Groups  map[string]int `json:"groups"`
}

// SchoolStats holds the "School at a glance" numbers shown on home.html.
type SchoolStats struct {
	TotalStudents int `json:"totalStudents"`
	TotalTeachers int `json:"totalTeachers"`
	TotalStaff    int `json:"totalStaff"`
	Established   int `json:"established"`
}

// ResultPDFRecord is an official class result sheet PDF published by the school office.
type ResultPDFRecord struct {
	ID          string `json:"id"`
	Exam        string `json:"exam"`
	Class       string `json:"class"`
	Title       string `json:"title"`
	FilePath    string `json:"filePath"`
	PublishedAt string `json:"publishedAt"`
}

func defaultStats() SchoolStats {
	return SchoolStats{TotalStudents: 1850, TotalTeachers: 85, TotalStaff: 42, Established: 1948}
}

// One online admission application submitted from admission-form.html.
type AdmissionApplication struct {
	ID             string `json:"id"`
	StudentName    string `json:"student"`
	StudentBangla  string `json:"studentBangla"`
	Class          string `json:"class"`
	Group          string `json:"group"`
	DOB            string `json:"dob"`
	Gender         string `json:"gender"`
	GuardianName   string `json:"guardian"`
	GuardianPhone  string `json:"phone"`
	GuardianEmail  string `json:"email"`
	Address        string `json:"address"`
	PreviousSchool string `json:"previousSchool"`
	Status         string `json:"status"`
	SubmittedAt    string `json:"submittedAt"`
}

func defaultSeats() SeatConfig {
	groups := map[string]int{"science": 5, "humanities": 3, "business": 0}
	return SeatConfig{
		Classes: map[string]int{
			"2": 18, "3": 12, "4": 0, "5": 9,
			"6": 40, "7": 15, "8": 0,
			"9":  groups["science"] + groups["humanities"] + groups["business"],
			"10": 5,
		},
		Groups: groups,
	}
}

func (a *App) seatsFile() string {
	return filepath.Join(filepath.Dir(a.file), "seats.json")
}

func (a *App) admissionsFile() string {
	return filepath.Join(filepath.Dir(a.file), "admissions.json")
}

func (a *App) achievementsFile() string {
	return filepath.Join(filepath.Dir(a.file), "achievements.json")
}

func (a *App) statsFile() string {
	return filepath.Join(filepath.Dir(a.file), "stats.json")
}

func (a *App) resultPDFsFile() string {
	return filepath.Join(filepath.Dir(a.file), "result_pdfs.json")
}

func (a *App) loadAdmissions() error {
	if data, err := os.ReadFile(a.seatsFile()); err == nil {
		if err = json.Unmarshal(data, &a.seats); err != nil {
			return err
		}
	} else if os.IsNotExist(err) {
		a.seats = defaultSeats()
	} else {
		return err
	}
	if a.seats.Classes == nil || a.seats.Groups == nil {
		a.seats = defaultSeats()
	}

	if data, err := os.ReadFile(a.admissionsFile()); err == nil {
		if err = json.Unmarshal(data, &a.admissions); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	if data, err := os.ReadFile(a.achievementsFile()); err == nil {
		if err = json.Unmarshal(data, &a.achievements); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	if data, err := os.ReadFile(a.statsFile()); err == nil {
		if err = json.Unmarshal(data, &a.stats); err != nil {
			return err
		}
	} else if os.IsNotExist(err) {
		a.stats = defaultStats()
	} else {
		return err
	}
	if a.stats.Established == 0 {
		a.stats = defaultStats()
	}

	if data, err := os.ReadFile(a.resultPDFsFile()); err == nil {
		if err = json.Unmarshal(data, &a.resultPDFs); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (a *App) saveResultPDFs() error {
	return writeJSONAtomic(a.resultPDFsFile(), a.resultPDFs)
}

func (a *App) saveStats() error {
	return writeJSONAtomic(a.statsFile(), a.stats)
}

func (a *App) saveSeats() error {
	return writeJSONAtomic(a.seatsFile(), a.seats)
}

func (a *App) saveAdmissions() error {
	return writeJSONAtomic(a.admissionsFile(), a.admissions)
}

func (a *App) saveAchievements() error {
	return writeJSONAtomic(a.achievementsFile(), a.achievements)
}

// Writes any value as formatted JSON to a temporary file and renames it into place.
func writeJSONAtomic(target string, v any) error {
	dir := filepath.Dir(target)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, "tmp-*")
	if err != nil {
		return err
	}
	name := f.Name()
	defer os.Remove(name)
	if _, err = f.Write(data); err != nil {
		_ = f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(name, target)
}

