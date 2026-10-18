package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type Student struct {
	ID       int
	Name     string
	Class    int
	Section  string
	Roll     int
	Guardian string
	Phone    string
	Present  bool
	FeePaid  bool
	Marks    map[string]int
}

type Row struct {
	S     Student
	Total int
	GPA   float64
	Grade string
}

type Store struct {
	mu       sync.Mutex
	students []*Student
	nextID   int
}

var subjects = []string{"Bangla", "English", "Math", "Science", "Social"}

const monthlyFee = 800 // ৳ per student per month

func gradeOf(m int) (string, float64) {
	switch {
	case m >= 80:
		return "A+", 5
	case m >= 70:
		return "A", 4
	case m >= 60:
		return "A-", 3.5
	case m >= 50:
		return "B", 3
	case m >= 40:
		return "C", 2
	case m >= 33:
		return "D", 1
	}
	return "F", 0
}

func (st *Store) add(n string, c int, sec string, roll int, g, p string, marks map[string]int) {
	st.nextID++
	st.students = append(st.students, &Student{st.nextID, n, c, sec, roll, g, p, false, false, marks})
}

func seed() *Store {
	st := &Store{}
	mk := func(a, b, c, d, e int) map[string]int {
		return map[string]int{"Bangla": a, "English": b, "Math": c, "Science": d, "Social": e}
	}
	st.add("Rahim Uddin", 9, "A", 1, "Abdul Karim", "01711-000001", mk(82, 75, 90, 78, 70))
	st.add("Fatima Akter", 9, "A", 2, "Nasir Ahmed", "01811-000002", mk(91, 88, 79, 85, 92))
	st.add("Tanvir Hasan", 8, "B", 5, "Mizanur Rahman", "01911-000003", mk(65, 58, 72, 61, 55))
	st.add("Nusrat Jahan", 10, "A", 3, "Shafiqul Islam", "01611-000004", mk(77, 69, 83, 74, 80))
	st.add("Sabbir Ahmed", 7, "C", 11, "Jahangir Alam", "01511-000005", mk(45, 38, 30, 42, 50))
	st.add("Mim Chowdhury", 10, "B", 8, "Kamal Hossain", "01711-000006", mk(88, 92, 95, 90, 86))
	st.students[1].Present, st.students[3].Present, st.students[0].Present = true, true, true
	st.students[1].FeePaid, st.students[5].FeePaid = true, true
	return st
}

func (st *Store) rows(class int) []Row {
	var out []Row
	for _, s := range st.students {
		if class != 0 && s.Class != class {
			continue
		}
		total, pts, fail := 0, 0.0, false
		for _, sub := range subjects {
			m := s.Marks[sub]
			total += m
			_, p := gradeOf(m)
			pts += p
			if p == 0 {
				fail = true
			}
		}
		gpa := pts / float64(len(subjects))
		if fail {
			gpa = 0
		}
		g := "F"
		switch {
		case gpa == 5:
			g = "A+"
		case gpa >= 4:
			g = "A"
		case gpa >= 3.5:
			g = "A-"
		case gpa >= 3:
			g = "B"
		case gpa >= 2:
			g = "C"
		case gpa >= 1:
			g = "D"
		}
		out = append(out, Row{*s, total, gpa, g})
	}
	return out
}

var tpl = template.Must(template.New("p").Funcs(template.FuncMap{
	"f2": func(f float64) string { return strconv.FormatFloat(f, 'f', 2, 64) },
}).Parse(page))

func main() {
	st := seed()
	render := func(w http.ResponseWriter, name string, data map[string]any) {
		data["Page"] = name
		data["Subjects"] = subjects
		tpl.Execute(w, data)
	}
	lock := func(h func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) { st.mu.Lock(); defer st.mu.Unlock(); h(w, r) }
	}

	http.HandleFunc("/", lock(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		present, paid := 0, 0
		for _, s := range st.students {
			if s.Present {
				present++
			}
			if s.FeePaid {
				paid++
			}
		}
		n := len(st.students)
		render(w, "home", map[string]any{"Total": n, "Present": present, "Absent": n - present,
			"Collected": paid * monthlyFee, "Due": (n - paid) * monthlyFee})
	}))

	http.HandleFunc("/students", lock(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			r.ParseForm()
			c, _ := strconv.Atoi(r.FormValue("class"))
			roll, _ := strconv.Atoi(r.FormValue("roll"))
			if r.FormValue("name") != "" {
				st.add(r.FormValue("name"), c, r.FormValue("section"), roll, r.FormValue("guardian"), r.FormValue("phone"), map[string]int{})
			}
			http.Redirect(w, r, "/students", http.StatusSeeOther)
			return
		}
		c, _ := strconv.Atoi(r.URL.Query().Get("class"))
		render(w, "students", map[string]any{"Rows": st.rows(c), "Sel": c, "Classes": []int{6, 7, 8, 9, 10}})
	}))

	toggle := func(field string) http.HandlerFunc {
		return lock(func(w http.ResponseWriter, r *http.Request) {
			id, _ := strconv.Atoi(r.FormValue("id"))
			for _, s := range st.students {
				if s.ID == id {
					if field == "att" {
						s.Present = !s.Present
					} else {
						s.FeePaid = true
					}
				}
			}
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
		})
	}
	http.HandleFunc("/attendance/toggle", toggle("att"))
	http.HandleFunc("/fees/pay", toggle("fee"))

	for _, p := range []string{"attendance", "results", "fees"} {
		p := p
		http.HandleFunc("/"+p, lock(func(w http.ResponseWriter, r *http.Request) {
			render(w, p, map[string]any{"Rows": st.rows(0), "Fee": monthlyFee})
		}))
	}

	http.HandleFunc("/api/students", lock(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(st.students)
	}))

	log.Println("School system running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

const page = `<!doctype html><html lang="en"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1"><title>School Management System – Bangladesh</title>
<style>
:root{--g:#006a4e;--r:#f42a41;--bg:#f4f7f5;--c:#fff;--t:#1d2b25}
@media(prefers-color-scheme:dark){:root{--bg:#101815;--c:#18231e;--t:#e6efe9}}
*{box-sizing:border-box}body{margin:0;font-family:system-ui,"Noto Sans Bengali",sans-serif;background:var(--bg);color:var(--t)}
header{background:var(--g);color:#fff;padding:14px 20px}header h1{margin:0;font-size:1.2rem}header small{opacity:.8}
nav{display:flex;gap:6px;flex-wrap:wrap;padding:10px 20px;background:var(--c);border-bottom:3px solid var(--r)}
nav a{padding:6px 14px;border-radius:20px;text-decoration:none;color:var(--t)}nav a.on{background:var(--g);color:#fff}
main{max-width:980px;margin:20px auto;padding:0 16px}
.cards{display:grid;grid-template-columns:repeat(auto-fit,minmax(180px,1fr));gap:12px}
.card{background:var(--c);border-radius:12px;padding:16px;box-shadow:0 1px 4px #0002}.card b{display:block;font-size:1.8rem;color:var(--g)}
table{width:100%;border-collapse:collapse;background:var(--c);border-radius:12px;overflow:hidden}
th,td{padding:10px;text-align:left;border-bottom:1px solid #8883}th{background:var(--g);color:#fff}
.wrap{overflow-x:auto}.ok{color:var(--g);font-weight:600}.bad{color:var(--r);font-weight:600}
button,input,select{padding:8px;border-radius:8px;border:1px solid #8886;font:inherit}
button{background:var(--g);color:#fff;border:0;cursor:pointer}form.add{display:flex;gap:8px;flex-wrap:wrap;margin-top:16px}
</style></head><body>
<header><h1>🇧🇩 Ekota High School – Management System</h1><small>Demo · Dhaka, Bangladesh · Go backend</small></header>
<nav>
<a href="/" {{if eq .Page "home"}}class="on"{{end}}>Dashboard</a>
<a href="/students" {{if eq .Page "students"}}class="on"{{end}}>Students</a>
<a href="/attendance" {{if eq .Page "attendance"}}class="on"{{end}}>Attendance</a>
<a href="/results" {{if eq .Page "results"}}class="on"{{end}}>Results</a>
<a href="/fees" {{if eq .Page "fees"}}class="on"{{end}}>Fees</a></nav><main>
{{if eq .Page "home"}}<div class="cards">
<div class="card">Total students<b>{{.Total}}</b></div>
<div class="card">Present today<b>{{.Present}}</b></div>
<div class="card">Absent today<b>{{.Absent}}</b></div>
<div class="card">Fees collected<b>৳{{.Collected}}</b></div>
<div class="card">Fees due<b class="bad">৳{{.Due}}</b></div></div>
{{else if eq .Page "students"}}
<p>Filter: <a href="/students">All</a>{{range .Classes}} · <a href="/students?class={{.}}">Class {{.}}</a>{{end}}</p>
<div class="wrap"><table><tr><th>Roll</th><th>Name</th><th>Class</th><th>Guardian</th><th>Phone</th></tr>
{{range .Rows}}<tr><td>{{.S.Roll}}</td><td>{{.S.Name}}</td><td>{{.S.Class}}-{{.S.Section}}</td><td>{{.S.Guardian}}</td><td>{{.S.Phone}}</td></tr>{{end}}</table></div>
<form class="add" method="post"><input name="name" placeholder="Student name" required>
<input name="class" type="number" min="6" max="10" placeholder="Class" required style="width:90px">
<input name="section" placeholder="Sec" style="width:70px"><input name="roll" type="number" placeholder="Roll" style="width:80px">
<input name="guardian" placeholder="Guardian"><input name="phone" placeholder="01XXXXXXXXX"><button>Add student</button></form>
{{else if eq .Page "attendance"}}<div class="wrap"><table><tr><th>Name</th><th>Class</th><th>Today</th><th></th></tr>
{{range .Rows}}<tr><td>{{.S.Name}}</td><td>{{.S.Class}}-{{.S.Section}}</td>
<td>{{if .S.Present}}<span class="ok">Present</span>{{else}}<span class="bad">Absent</span>{{end}}</td>
<td><form method="post" action="/attendance/toggle"><input type="hidden" name="id" value="{{.S.ID}}"><button>Toggle</button></form></td></tr>{{end}}</table></div>
{{else if eq .Page "results"}}<p>GPA scale: 80+ = A+ (5.00), 70+ = A (4.00), 60+ = A- (3.50), 50+ = B (3.00), 40+ = C (2.00), 33+ = D (1.00). Failing any subject gives GPA 0.00.</p>
<div class="wrap"><table><tr><th>Name</th><th>Class</th>{{range .Subjects}}<th>{{.}}</th>{{end}}<th>Total</th><th>GPA</th><th>Grade</th></tr>
{{range .Rows}}{{$m := .S.Marks}}<tr><td>{{.S.Name}}</td><td>{{.S.Class}}</td>{{range $.Subjects}}<td>{{index $m .}}</td>{{end}}
<td>{{.Total}}</td><td>{{f2 .GPA}}</td><td class="{{if eq .Grade "F"}}bad{{else}}ok{{end}}">{{.Grade}}</td></tr>{{end}}</table></div>
{{else if eq .Page "fees"}}<p>Monthly tuition: ৳{{.Fee}} per student</p>
<div class="wrap"><table><tr><th>Name</th><th>Class</th><th>Amount</th><th>Status</th><th></th></tr>
{{range .Rows}}<tr><td>{{.S.Name}}</td><td>{{.S.Class}}-{{.S.Section}}</td><td>৳{{$.Fee}}</td>
<td>{{if .S.FeePaid}}<span class="ok">Paid</span>{{else}}<span class="bad">Due</span>{{end}}</td>
<td>{{if not .S.FeePaid}}<form method="post" action="/fees/pay"><input type="hidden" name="id" value="{{.S.ID}}"><button>Mark paid</button></form>{{end}}</td></tr>{{end}}</table></div>
{{end}}</main></body></html>`