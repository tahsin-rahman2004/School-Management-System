# Bangladesh School Management System (Demo)

A small demo built with Go's standard library only. Data is stored in memory (resets on restart).

## Run
    go run .
Open http://localhost:8080

## Features
- Dashboard: totals, today's attendance, fees collected/due (in ৳)
- Students: list, add, filter by class
- Attendance: mark present/absent for today
- Results: marks per subject with the Bangladesh GPA (out of 5.00) and letter grades
- Fees: monthly tuition status, mark as paid
- JSON API: GET /api/students

## Ideas for next steps
Replace the in-memory store with SQLite/PostgreSQL, add login (teacher/admin), and add SMS notices for guardians.