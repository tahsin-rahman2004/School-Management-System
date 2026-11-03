package main

import (
	"encoding/json"
	"fmt"
	"net/mail"
	"os"
	"path/filepath"
	"strings"
)

type AttendanceRecord struct {
	Date    string
	Subject string
	Status  string
}

type User struct {
	Attendance []AttendanceRecord
	Group      string
	Email      string
	Name       string
	StudentID  string
	Class      string
	Role       string
	Salt       string
	Hash       string
	MustChange bool
}

func (a *App) save() error {
	if err := os.MkdirAll(filepath.Dir(a.file), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(a.users, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(a.file), "accounts-*")
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
	return os.Rename(name, a.file)
}

func (a *App) put(u User) error {
	old, exists := a.users[u.Email]
	a.users[u.Email] = u
	if err := a.save(); err != nil {
		if exists {
			a.users[u.Email] = old
		} else {
			delete(a.users, u.Email)
		}
		return err
	}
	return nil
}

func validEmail(s string) bool {
	m, e := mail.ParseAddress(s)
	return e == nil && m.Address == s && len(s) <= 254
}

func (a *App) load() error {
	data, err := os.ReadFile(a.file)
	if err == nil {
		if err = json.Unmarshal(data, &a.users); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (a *App) initializeAdmin(email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	if !validEmail(email) {
		return fmt.Errorf("enter a valid email")
	}
	for _, u := range a.users {
		if u.Role == "admin" {
			return fmt.Errorf("an administrator already exists")
		}
	}
	password := token()[:20]
	u := User{Email: email, Name: "School office", Role: "admin", MustChange: true}
	setPassword(&u, password)
	if err := a.put(u); err != nil {
		return err
	}
	fmt.Printf("Email: %s\nTemporary password: %s\n", email, password)
	return nil
}

func (a *App) resetAdmin(email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	u, exists := a.users[email]
	if !exists || u.Role != "admin" {
		return fmt.Errorf("administrator account not found")
	}
	password := token()[:20]
	setPassword(&u, password)
	u.MustChange = true
	if err := a.put(u); err != nil {
		return err
	}
	for id, session := range a.sessions {
		if session.Email == email {
			delete(a.sessions, id)
		}
	}
	fmt.Printf("Email: %s\nTemporary password: %s\n", email, password)
	return nil
}
