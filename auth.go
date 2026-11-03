package main

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"time"
)

type Session struct {
	Email, CSRF string
	Expires     time.Time
}

type Attempts struct {
	Count int
	Until time.Time
}

func token() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

func passwordHash(password, salt string) string {
	key, err := pbkdf2.Key(sha256.New, password, []byte(salt), 600000, 32)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(key)
}

func setPassword(u *User, password string) {
	u.Salt = token()
	u.Hash = passwordHash(password, u.Salt)
}

func matches(u User, password string) bool {
	return subtle.ConstantTimeCompare([]byte(passwordHash(password, u.Salt)), []byte(u.Hash)) == 1
}

func (a *App) session(r *http.Request) (string, Session, User, bool) {
	c, err := r.Cookie("school_session")
	if err != nil {
		return "", Session{}, User{}, false
	}
	s, ok := a.sessions[c.Value]
	u, exists := a.users[s.Email]
	return c.Value, s, u, ok && exists && time.Now().Before(s.Expires)
}

func (a *App) cookie(w http.ResponseWriter, id string, age int) {
	http.SetCookie(w, &http.Cookie{Name: "school_session", Value: id, Path: "/", HttpOnly: true, Secure: a.secure, SameSite: http.SameSiteStrictMode, MaxAge: age})
}

func destination(u User) string {
	if u.MustChange {
		return "/change-password"
	}
	if u.Role == "admin" {
		return "/admin"
	}
	return "/student"
}
