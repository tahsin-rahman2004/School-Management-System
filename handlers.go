package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type View struct {
	Attendance                        []AttendanceRecord
	DemoAttendance                    bool
	Subjects                          []Subject
	LoginPath                         string
	AdminLogin                        bool
	Title, Mode, Error, Success, CSRF string
	User                              User
	Students                          []User
}

func (a *App) render(w http.ResponseWriter, v View) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := a.pages.Execute(w, v); err != nil {
		log.Print(err)
	}
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "same-origin")
	w.Header().Set("X-Frame-Options", "DENY")
	if !strings.HasPrefix(r.URL.Path, "/api") && (strings.HasSuffix(r.URL.Path, ".css") || strings.HasSuffix(r.URL.Path, ".js") || strings.HasPrefix(r.URL.Path, "/images/") || strings.HasPrefix(r.URL.Path, "/downloads/")) {
		a.static(w, r)
		return
	}
	switch r.URL.Path {
	case "/login", "/admin/login", "/change-password", "/student", "/student/attendance", "/admin", "/logout", "/academics.html":
	default:
		a.static(w, r)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	if r.Method != "GET" && r.Method != "POST" {
		w.WriteHeader(405)
		return
	}
	if r.Method == "POST" {
		scheme := "http"
		if r.TLS != nil || a.secure {
			scheme = "https"
		}
		if r.Header.Get("Origin") != scheme+"://"+r.Host {
			http.Error(w, "Request origin rejected", 403)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 16384)
		if r.ParseForm() != nil {
			http.Error(w, "Invalid form", 400)
			return
		}
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	now := time.Now()
	for id, s := range a.sessions {
		if now.After(s.Expires) {
			delete(a.sessions, id)
		}
	}
	for ip, v := range a.attempts {
		if now.After(v.Until) {
			delete(a.attempts, ip)
		}
	}
	id, s, u, logged := a.session(r)
	if r.URL.Path == "/academics.html" {
		http.Redirect(w, r, "/login", 303)
		return
	}
	if r.URL.Path == "/login" || r.URL.Path == "/admin/login" {
		role := "student"
		title := "Student login"
		if r.URL.Path == "/admin/login" {
			role = "admin"
			title = "Admin login"
		}
		if logged && u.Role == role {
			http.Redirect(w, r, destination(u), 303)
			return
		}
		v := View{Mode: "login", Title: title, LoginPath: r.URL.Path, AdminLogin: role == "admin"}
		if r.Method == "POST" {
			ip, _, _ := net.SplitHostPort(r.RemoteAddr)
			limit := a.attempts[ip]
			if limit.Count >= 10 {
				w.WriteHeader(429)
				v.Error = "Too many attempts. Try again in 15 minutes."
				a.render(w, v)
				return
			}
			limit.Count++
			if limit.Until.IsZero() {
				limit.Until = now.Add(15 * time.Minute)
			}
			a.attempts[ip] = limit
			email := strings.ToLower(strings.TrimSpace(r.FormValue("email")))
			candidate, exists := a.users[email]
			if !exists {
				candidate.Salt = "unknown-account"
				candidate.Hash = strings.Repeat("0", 64)
			}
			if matches(candidate, r.FormValue("password")) && exists && candidate.Role == role {
				delete(a.attempts, ip)
				delete(a.sessions, id)
				sid := token()
				a.sessions[sid] = Session{email, token(), now.Add(8 * time.Hour)}
				a.cookie(w, sid, 28800)
				http.Redirect(w, r, destination(candidate), 303)
				return
			}
			v.Error = "Email or password is incorrect."
		}
		a.render(w, v)
		return
	}
	if !logged {
		loginPath := "/login"
		if r.URL.Path == "/admin" {
			loginPath = "/admin/login"
		}
		http.Redirect(w, r, loginPath, 303)
		return
	}
	if r.Method == "POST" && r.FormValue("csrf") != s.CSRF {
		http.Error(w, "Session verification failed. Reload the page.", 403)
		return
	}
	if r.URL.Path == "/logout" {
		if r.Method != "POST" {
			w.WriteHeader(405)
			return
		}
		delete(a.sessions, id)
		a.cookie(w, "", -1)
		loginPath := "/login"
		if u.Role == "admin" {
			loginPath = "/admin/login"
		}
		http.Redirect(w, r, loginPath, 303)
		return
	}
	if u.MustChange && r.URL.Path != "/change-password" {
		http.Redirect(w, r, "/change-password", 303)
		return
	}
	v := View{User: u, CSRF: s.CSRF, Subjects: subjectsFor(u), Attendance: u.Attendance}
	if r.URL.Path == "/student/attendance" && u.Role == "student" && len(u.Attendance) == 0 {
		v.Attendance = sampleAttendance(v.Subjects)
		v.DemoAttendance = true
	}
	switch r.URL.Path {
	case "/change-password":
		v.Mode = "password"
		v.Title = "Choose your own password"
		if r.Method == "POST" {
			p := r.FormValue("password")
			if len(p) < 8 || len(p) > 128 {
				v.Error = "Use a password between 8 and 128 characters."
			} else if p != r.FormValue("confirm") {
				v.Error = "The passwords do not match."
			} else if !matches(u, r.FormValue("current")) {
				v.Error = "Your current password is incorrect."
			} else if matches(u, p) {
				v.Error = "Choose a different password from your current password."
			} else {
				setPassword(&u, p)
				u.MustChange = false
				if err := a.put(u); err != nil {
					v.Error = "Unable to save the password. Please try again."
				} else {
					for key, session := range a.sessions {
						if session.Email == u.Email {
							delete(a.sessions, key)
						}
					}
					sid := token()
					a.sessions[sid] = Session{u.Email, token(), now.Add(8 * time.Hour)}
					a.cookie(w, sid, 28800)
					http.Redirect(w, r, destination(u), 303)
					return
				}
			}
		}
	case "/admin":
		if u.Role != "admin" {
			http.Error(w, "School administrator access required", 403)
			return
		}
		v.Mode = "admin"
		v.Title = "School office"
		if r.Method == "POST" {
			email := strings.ToLower(strings.TrimSpace(r.FormValue("email")))
			student, exists := a.users[email]
			action := r.FormValue("action")
			if action == "group" {
				group := r.FormValue("group")
				if !exists || student.Role != "student" || (student.Class != "9" && student.Class != "10") || !validGroup(group) {
					v.Error = "Select a valid group for a Class 9–10 student."
				} else {
					student.Group = group
					if err := a.put(student); err != nil {
						v.Error = "Unable to save group."
					} else {
						v.Success = "Student group updated."
					}
				}
			} else if action == "reset" {
				if !exists || student.Role != "student" {
					v.Error = "Student account not found."
				}
			} else if action == "create" {
				class := r.FormValue("class")
				name := strings.TrimSpace(r.FormValue("name"))
				studentID := strings.TrimSpace(r.FormValue("student_id"))
				group := ""
				if class == "9" || class == "10" {
					group = r.FormValue("group")
					if !validGroup(group) {
						v.Error = "Select Science, Humanities or Business Studies for Class 9–10."
					}
				}
				classOK := false
				for _, c := range []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"} {
					if class == c {
						classOK = true
					}
				}
				if exists || !validEmail(email) || name == "" || len(name) > 100 || studentID == "" || len(studentID) > 50 || !classOK {
					v.Error = "Enter a unique email, student ID, name and class (1–10)."
				}
				for _, other := range a.users {
					if other.StudentID == studentID && studentID != "" {
						v.Error = "This student ID already has an account."
					}
				}
				student = User{Email: email, Name: name, StudentID: studentID, Class: class, Group: group, Role: "student"}
			} else {
				v.Error = "Unknown action."
			}
			if v.Error == "" && action != "group" {
				temporary := token()[:20]
				setPassword(&student, temporary)
				student.MustChange = true
				if err := a.put(student); err != nil {
					v.Error = "Unable to save the account."
				} else {
					for key, session := range a.sessions {
						if session.Email == email {
							delete(a.sessions, key)
						}
					}
					v.Success = "Account: " + email + " · Temporary password: " + temporary + " · Give these details privately to the student."
				}
			}
		}
		for _, student := range a.users {
			if student.Role == "student" {
				v.Students = append(v.Students, student)
			}
		}
	case "/student", "/student/attendance":
		if u.Role != "student" {
			http.Redirect(w, r, "/admin", 303)
			return
		}
		v.Mode = "student"
		v.Title = "Your student dashboard"
		if r.URL.Path == "/student/attendance" {
			v.Mode = "attendance"
			v.Title = "Your attendance"
		}
	}
	a.render(w, v)
}

func (a *App) static(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "HEAD" {
		w.WriteHeader(405)
		return
	}
	p := strings.TrimPrefix(r.URL.Path, "/")
	if p == "" {
		p = "home.html"
	}
	allowed := false
	if !strings.Contains(p, "/") {
		ext := filepath.Ext(p)
		allowed = ext == ".html" || ext == ".css" || ext == ".js"
		if p == "academics.js" {
			allowed = false
		}
	}
	if strings.HasPrefix(p, "images/") || strings.HasPrefix(p, "downloads/") {
		ext := strings.ToLower(filepath.Ext(p))
		allowed = ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".webp" || ext == ".svg" || ext == ".pdf" || ext == ".html"
	}
	clean := filepath.Clean(p)
	if !allowed || clean != p || strings.HasPrefix(clean, ".") {
		http.NotFound(w, r)
		return
	}
	resolved, err := filepath.EvalSymlinks(p)
	if err != nil || resolved != p {
		http.NotFound(w, r)
		return
	}
	info, err := os.Stat(p)
	if err != nil || info.IsDir() {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Cache-Control", "no-cache")
	http.ServeFile(w, r, p)
}
