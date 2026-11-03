package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type App struct {
	mu       sync.Mutex
	users    map[string]User
	sessions map[string]Session
	attempts map[string]Attempts
	file     string
	pages    *template.Template
	secure   bool
}

func newApp(file string) (*App, error) {
	a := &App{users: map[string]User{}, sessions: map[string]Session{}, attempts: map[string]Attempts{}, file: file, secure: os.Getenv("COOKIE_SECURE") == "true"}
	if err := a.load(); err != nil {
		return nil, err
	}
	var err error
	a.pages, err = template.ParseFiles("templates/portal.html")
	return a, err
}

func main() {
	dataFile := os.Getenv("DATA_FILE")
	if dataFile == "" {
		dataFile = "data/accounts.json"
	}
	a, err := newApp(dataFile)
	if err != nil {
		log.Fatal(err)
	}
	if len(os.Args) > 1 {
		if len(os.Args) != 3 || (os.Args[1] != "init-admin" && os.Args[1] != "reset-admin") {
			log.Fatal("Usage: go run . [init-admin|reset-admin] office@example.com")
		}
		var err error
		if os.Args[1] == "reset-admin" {
			err = a.resetAdmin(os.Args[2])
		} else {
			err = a.initializeAdmin(os.Args[2])
		}
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = "127.0.0.1:8080"
	}
	server := &http.Server{Addr: addr, Handler: a, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	log.Printf("School website: http://%s", addr)
	log.Fatal(server.ListenAndServe())
}
