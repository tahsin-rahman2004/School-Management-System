// teachers.go - the teachers shown on the public Teachers page.
// Admins manage them in /admin and teachers.js reads them from /api/teachers.
package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// One teacher card. IsHead marks the Head Teacher shown at the top of the page.
type Teacher struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Role          string `json:"role"`
	Subject       string `json:"subject"`
	Qualification string `json:"qualification"`
	Email         string `json:"email"`
	Image         string `json:"image"`
	IsHead        bool   `json:"isHead"`
}

func defaultTeachers() []Teacher {
	return []Teacher{
		{
			ID:            "t-1",
			Name:          "Teacher name",
			Role:          "Head Teacher",
			Subject:       "Administration",
			Qualification: "M.A., M.Ed.",
			Email:         "headteacher@example.com",
			Image:         "images/teachers/sample-teacher-1.png",
			IsHead:        true,
		},
		{
			ID:            "t-2",
			Name:          "Teacher name",
			Role:          "Assistant teacher",
			Subject:       "Bangla",
			Qualification: "M.A., B.Ed.",
			Email:         "bangla.teacher@example.com",
			Image:         "images/teachers/sample-teacher-2-hijab.png",
		},
		{
			ID:            "t-3",
			Name:          "Teacher name",
			Role:          "Assistant teacher",
			Subject:       "English",
			Qualification: "M.A., B.Ed.",
			Email:         "english.teacher@example.com",
			Image:         "images/teachers/sample-teacher-3.png",
		},
		{
			ID:            "t-4",
			Name:          "Teacher name",
			Role:          "Assistant teacher",
			Subject:       "Mathematics",
			Qualification: "M.Sc., B.Ed.",
			Email:         "math.teacher@example.com",
			Image:         "images/teachers/sample-teacher-4.png",
		},
		{
			ID:            "t-5",
			Name:          "Teacher name",
			Role:          "Assistant teacher",
			Subject:       "Science",
			Qualification: "M.Sc., B.Ed.",
			Email:         "science.teacher@example.com",
			Image:         "images/teachers/sample-teacher-5.png",
		},
		{
			ID:            "t-6",
			Name:          "Teacher name",
			Role:          "Assistant teacher",
			Subject:       "ICT",
			Qualification: "B.Sc. (CSE)",
			Email:         "ict.teacher@example.com",
			Image:         "images/teachers/sample-teacher-6-hijab.png",
		},
		{
			ID:            "t-7",
			Name:          "Teacher name",
			Role:          "Assistant teacher",
			Subject:       "Social Studies",
			Qualification: "M.S.S., B.Ed.",
			Email:         "socialstudies.teacher@example.com",
			Image:         "images/teachers/sample-teacher-7.png",
		},
	}
}

func allowedTeacherImages() []string {
	return []string{
		"images/teachers/sample-teacher-1.png",
		"images/teachers/sample-teacher-2-hijab.png",
		"images/teachers/sample-teacher-2.png",
		"images/teachers/sample-teacher-3.png",
		"images/teachers/sample-teacher-4.png",
		"images/teachers/sample-teacher-5.png",
		"images/teachers/sample-teacher-6-hijab.png",
		"images/teachers/sample-teacher-6.png",
		"images/teachers/sample-teacher-7.png",
	}
}

// Only the portraits we ship with the site are allowed, so no one can point
// the page at an outside image.
func validTeacherImage(img string) bool {
	for _, allowed := range allowedTeacherImages() {
		if img == allowed {
			return true
		}
	}
	return false
}

func (a *App) teachersFile() string {
	return filepath.Join(filepath.Dir(a.file), "teachers.json")
}

// Reads teachers.json at start-up (or uses the default teachers).
func (a *App) loadTeachers() error {
	data, err := os.ReadFile(a.teachersFile())
	if err == nil {
		if err = json.Unmarshal(data, &a.teachers); err != nil {
			return err
		}
		return nil
	}
	if !os.IsNotExist(err) {
		return err
	}
	a.teachers = defaultTeachers()
	return nil
}

// Saves teachers safely (temporary file, then rename).
func (a *App) saveTeachers() error {
	dir := filepath.Dir(a.teachersFile())
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(a.teachers, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, "teachers-*")
	if err != nil {
		return err
	}
	name := f.Name()
	defer os.Remove(name)
	if _, err = f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	if err := os.Rename(name, a.teachersFile()); err != nil {
		return err
	}
	a.writeSQLSnapshot()
	return nil
}

// True when the role is "Head Teacher" (ignoring capital letters and spaces).
func isHeadRole(role string) bool {
	return strings.EqualFold(strings.TrimSpace(role), "Head Teacher")
}

