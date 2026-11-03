package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func request(a *App, method, path string, form url.Values, cookie *http.Cookie) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, "http://school.test"+path, strings.NewReader(form.Encode()))
	if method == "POST" {
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.Header.Set("Origin", "http://school.test")
	}
	if cookie != nil {
		r.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	a.ServeHTTP(w, r)
	return w
}
func TestStudentLoginLifecycle(t *testing.T) {
	file := filepath.Join(t.TempDir(), "accounts.json")
	a, err := newApp(file)
	if err != nil {
		t.Fatal(err)
	}
	u := User{Email: "student@example.com", Name: "Test student", Role: "student", Class: "6", StudentID: "6001", MustChange: true}
	setPassword(&u, "temporary-pass-123")
	if err = a.put(u); err != nil {
		t.Fatal(err)
	}
	if got := request(a, "GET", "/student", nil, nil); got.Header().Get("Location") != "/login" {
		t.Fatal("dashboard accessible without login")
	}
	login := request(a, "POST", "/login", url.Values{"email": {u.Email}, "password": {"temporary-pass-123"}}, nil)
	if login.Header().Get("Location") != "/change-password" {
		t.Fatal("first login must change password")
	}
	c := login.Result().Cookies()[0]
	if !c.HttpOnly || c.SameSite != http.SameSiteStrictMode {
		t.Fatal("cookie flags missing")
	}
	if got := request(a, "GET", "/student", nil, c); got.Header().Get("Location") != "/change-password" {
		t.Fatal("first login bypass")
	}
	s := a.sessions[c.Value]
	form := url.Values{"current": {"temporary-pass-123"}, "password": {"NewPass8"}, "confirm": {"NewPass8"}}
	if got := request(a, "POST", "/change-password", form, c); got.Code != 403 {
		t.Fatal("CSRF accepted")
	}
	form.Set("csrf", s.CSRF)
	changed := request(a, "POST", "/change-password", form, c)
	if changed.Header().Get("Location") != "/student" {
		t.Fatalf("change failed: %s", changed.Body.String())
	}
	fresh := changed.Result().Cookies()[0]
	if got := request(a, "GET", "/student", nil, fresh); got.Code != 200 || !strings.Contains(got.Body.String(), "Test student") {
		t.Fatal("student dashboard missing")
	}
	if got := request(a, "GET", "/admin", nil, fresh); got.Code != 403 {
		t.Fatal("student accessed admin")
	}
	if got := request(a, "GET", "/student", nil, c); got.Header().Get("Location") != "/login" {
		t.Fatal("old session remained")
	}
	reloaded, err := newApp(file)
	if err != nil {
		t.Fatal(err)
	}
	saved := reloaded.users[u.Email]
	if saved.MustChange || !matches(saved, "NewPass8") || matches(saved, "temporary-pass-123") {
		t.Fatal("password persistence failed")
	}
	for _, path := range []string{"/data/accounts.json", "/.git/config", "/main.go", "/templates/portal.html", "/tahsin/academic-calendar.html"} {
		if got := request(a, "GET", path, nil, nil); got.Code != 404 {
			t.Fatalf("private path exposed: %s", path)
		}
	}
}
func TestAdminCreationResetAndOrigins(t *testing.T) {
	a, err := newApp(filepath.Join(t.TempDir(), "accounts.json"))
	if err != nil {
		t.Fatal(err)
	}
	a.users["office@example.com"] = User{Email: "office@example.com", Role: "admin"}
	a.sessions["admin-session"] = Session{"office@example.com", "csrf", time.Now().Add(time.Hour)}
	c := &http.Cookie{Name: "school_session", Value: "admin-session"}
	form := url.Values{"csrf": {"csrf"}, "action": {"create"}, "email": {"pupil@example.com"}, "name": {"Pupil"}, "student_id": {"1001"}, "class": {"1"}}
	w := request(a, "POST", "/admin", form, c)
	if !strings.Contains(w.Body.String(), "Temporary password:") {
		t.Fatal(w.Body.String())
	}
	u, ok := a.users["pupil@example.com"]
	if !ok || !u.MustChange || u.Role != "student" {
		t.Fatal("account creation failed")
	}
	oldHash := u.Hash
	a.sessions["pupil-session"] = Session{u.Email, "pupil-csrf", time.Now().Add(time.Hour)}
	form.Set("action", "reset")
	request(a, "POST", "/admin", form, c)
	if a.users[u.Email].Hash == oldHash {
		t.Fatal("reset did not change password")
	}
	if _, ok := a.sessions["pupil-session"]; ok {
		t.Fatal("reset did not revoke sessions")
	}
	r := httptest.NewRequest("POST", "http://school.test/admin", strings.NewReader(form.Encode()))
	r.Header.Set("Origin", "http://attacker.test")
	r.AddCookie(c)
	w = httptest.NewRecorder()
	a.ServeHTTP(w, r)
	if w.Code != 403 {
		t.Fatal("cross-origin post accepted")
	}
	a.attempts["192.0.2.1"] = Attempts{10, time.Now().Add(time.Minute)}
	w = request(a, "POST", "/login", url.Values{"email": {"bad@example.com"}, "password": {"wrong"}}, nil)
	if w.Code != 429 {
		t.Fatal("login rate limit missing")
	}
}

func TestSeparateLoginRoles(t *testing.T) {
	a, err := newApp(filepath.Join(t.TempDir(), "accounts.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, role := range []string{"admin", "student"} {
		u := User{Email: role + "@example.com", Role: role}
		setPassword(&u, "password-for-test")
		a.users[u.Email] = u
		correct := "/login"
		wrong := "/admin/login"
		target := "/student"
		if role == "admin" {
			correct, wrong, target = wrong, correct, "/admin"
		}
		form := url.Values{"email": {u.Email}, "password": {"password-for-test"}}
		rejected := request(a, "POST", wrong, form, nil)
		if rejected.Code != 200 || len(rejected.Result().Cookies()) != 0 || !strings.Contains(rejected.Body.String(), "incorrect") {
			t.Fatal("wrong login accepted", role)
		}
		accepted := request(a, "POST", correct, form, nil)
		if accepted.Header().Get("Location") != target {
			t.Fatal("correct login rejected", role)
		}
		cookie := accepted.Result().Cookies()[0]
		page := request(a, "GET", wrong, nil, cookie)
		if page.Code != 200 || page.Header().Get("Location") != "" {
			t.Fatal("wrong-role session redirected to dashboard")
		}
		logout := request(a, "POST", "/logout", url.Values{"csrf": {a.sessions[cookie.Value].CSRF}}, cookie)
		if logout.Header().Get("Location") != correct {
			t.Fatal("wrong logout destination")
		}
	}
	if w := request(a, "GET", "/admin", nil, nil); w.Header().Get("Location") != "/admin/login" {
		t.Fatal("admin login redirect missing")
	}
}

func TestProtectedActionsAndSessionExpiry(t *testing.T) {
	a, err := newApp(filepath.Join(t.TempDir(), "accounts.json"))
	if err != nil {
		t.Fatal(err)
	}
	a.users["office@example.com"] = User{Email: "office@example.com", Role: "admin"}
	a.sessions["admin"] = Session{"office@example.com", "secret", time.Now().Add(time.Hour)}
	c := &http.Cookie{Name: "school_session", Value: "admin"}
	for _, path := range []string{"/admin", "/change-password", "/logout"} {
		for _, csrf := range []string{"", "wrong"} {
			w := request(a, "POST", path, url.Values{"csrf": {csrf}}, c)
			if w.Code != 403 {
				t.Fatalf("%s accepted invalid CSRF", path)
			}
		}
	}
	if w := request(a, "GET", "/logout", nil, c); w.Code != 405 {
		t.Fatal("GET logged user out")
	}
	if _, ok := a.sessions[c.Value]; !ok {
		t.Fatal("rejected request removed session")
	}
	a.users["pupil@example.com"] = User{Email: "pupil@example.com", Role: "student"}
	a.sessions["pupil"] = Session{"pupil@example.com", "student-token", time.Now().Add(time.Hour)}
	w := request(a, "POST", "/admin", url.Values{"csrf": {"student-token"}, "action": {"reset"}, "email": {"office@example.com"}}, &http.Cookie{Name: "school_session", Value: "pupil"})
	if w.Code != 403 {
		t.Fatal("student could perform admin action")
	}
	a.sessions["admin"] = Session{"office@example.com", "secret", time.Now().Add(-time.Second)}
	if w := request(a, "GET", "/admin", nil, c); w.Header().Get("Location") != "/admin/login" {
		t.Fatal("expired session accepted")
	}
	a.secure = true
	recorder := httptest.NewRecorder()
	a.cookie(recorder, "new-session", 28800)
	cookie := recorder.Result().Cookies()[0]
	if !cookie.Secure || !cookie.HttpOnly || cookie.SameSite != http.SameSiteStrictMode {
		t.Fatal("secure cookie flags missing")
	}
}

func TestLoginReplacesExistingSession(t *testing.T) {
	a, err := newApp(filepath.Join(t.TempDir(), "accounts.json"))
	if err != nil {
		t.Fatal(err)
	}
	a.users["office@example.com"] = User{Email: "office@example.com", Role: "admin"}
	a.sessions["old-session"] = Session{"office@example.com", "old-token", time.Now().Add(time.Hour)}
	u := User{Email: "pupil@example.com", Role: "student"}
	setPassword(&u, "password123")
	a.users[u.Email] = u
	w := request(a, "POST", "/login", url.Values{"email": {u.Email}, "password": {"password123"}}, &http.Cookie{Name: "school_session", Value: "old-session"})
	if w.Code != 303 {
		t.Fatal("login failed")
	}
	fresh := w.Result().Cookies()[0]
	if fresh.Value == "old-session" || a.sessions[fresh.Value].CSRF == "old-token" {
		t.Fatal("session was not rotated")
	}
	if _, ok := a.sessions["old-session"]; ok {
		t.Fatal("previous session still valid")
	}
	csrf := a.sessions[fresh.Value].CSRF
	request(a, "POST", "/logout", url.Values{"csrf": {csrf}}, fresh)
	if w := request(a, "GET", "/student", nil, fresh); w.Header().Get("Location") != "/login" {
		t.Fatal("logged-out session usable")
	}
}

func TestStudentProfileCannotBeSelectedByURL(t *testing.T) {
	a, err := newApp(filepath.Join(t.TempDir(), "accounts.json"))
	if err != nil {
		t.Fatal(err)
	}
	a.users["alice@example.com"] = User{Email: "alice@example.com", Name: "Alice Student", StudentID: "1001", Role: "student", Class: "7"}
	a.users["bob@example.com"] = User{Email: "bob@example.com", Name: "Bob Student", StudentID: "2002", Role: "student", Class: "9"}
	a.sessions["alice-session"] = Session{"alice@example.com", "alice-csrf", time.Now().Add(time.Hour)}
	cookie := &http.Cookie{Name: "school_session", Value: "alice-session"}
	for _, path := range []string{"/student", "/student?id=2002", "/student?student_id=2002&email=bob@example.com"} {
		w := request(a, "GET", path, nil, cookie)
		body := w.Body.String()
		if w.Code != 200 || !strings.Contains(body, "Alice Student") || !strings.Contains(body, "1001") {
			t.Fatalf("own profile missing: %s", path)
		}
		if strings.Contains(body, "Bob Student") || strings.Contains(body, "bob@example.com") || strings.Contains(body, "2002") {
			t.Fatalf("another student's profile leaked: %s", path)
		}
		if w.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("private page can be cached")
		}
		anonymous := request(a, "GET", path, nil, nil)
		if anonymous.Header().Get("Location") != "/login" {
			t.Fatal("anonymous profile access")
		}
	}
	if w := request(a, "GET", "/student/2002", nil, cookie); w.Code != 404 {
		t.Fatal("unexpected ID-based profile route")
	}
}

func TestAttendancePageAccess(t *testing.T) {
	a, err := newApp(filepath.Join(t.TempDir(), "accounts.json"))
	if err != nil {
		t.Fatal(err)
	}
	if w := request(a, "GET", "/student/attendance", nil, nil); w.Header().Get("Location") != "/login" {
		t.Fatal("attendance accessible anonymously")
	}
	a.users["pupil@example.com"] = User{Email: "pupil@example.com", Name: "Own Student", Class: "2", Role: "student", MustChange: true}
	a.sessions["pupil"] = Session{"pupil@example.com", "csrf", time.Now().Add(time.Hour)}
	c := &http.Cookie{Name: "school_session", Value: "pupil"}
	if w := request(a, "GET", "/student/attendance", nil, c); w.Header().Get("Location") != "/change-password" {
		t.Fatal("password change bypass")
	}
	u := a.users["pupil@example.com"]
	u.MustChange = false
	a.users[u.Email] = u
	w := request(a, "GET", "/student/attendance?id=someone-else", nil, c)
	if w.Code != 200 || !strings.Contains(w.Body.String(), "Own Student") || !strings.Contains(w.Body.String(), "attendance-rows") {
		t.Fatal("attendance page failed")
	}
	w = request(a, "GET", "/student", nil, c)
	if strings.Contains(w.Body.String(), "attendance-rows") || !strings.Contains(w.Body.String(), "/student/attendance") {
		t.Fatal("dashboard separation failed")
	}
}
