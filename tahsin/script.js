const menuBtn = document.getElementById("menuBtn");
const menuClose = document.getElementById("menuClose");
const menu = document.getElementById("menu");

if (menuBtn && menuClose && menu) {
  function setMenu(open) {
    menu.classList.toggle("open", open);
    document.body.classList.toggle("menu-open", open);
    menuBtn.setAttribute("aria-expanded", String(open));
  }
  menuBtn.addEventListener("click", () => setMenu(!menu.classList.contains("open")));
  menuClose.addEventListener("click", () => {
    setMenu(false);
    menuBtn.focus();
  });
  const links = menu.querySelectorAll("a");
  function updateActiveLink() {
    const currentPage = location.pathname.split("/").pop() || "t.html";
    links.forEach(link => {
      const target = new URL(link.href);
      const targetPage = target.pathname.split("/").pop();
      const active = targetPage === currentPage &&
        ((["gallery.html", "results.html", "academic-calendar.html", "about.html"].includes(currentPage)) || target.hash === (location.hash || "#home"));
      link.classList.toggle("active", active);
      if (active) link.setAttribute("aria-current", (["gallery.html", "results.html", "academic-calendar.html", "about.html"].includes(currentPage)) ? "page" : "location");
      else link.removeAttribute("aria-current");
    });
  }
  links.forEach(link => link.addEventListener("click", () => setMenu(false)));
  document.addEventListener("click", event => {
    if (menu.classList.contains("open") &&
        !menu.contains(event.target) && !menuBtn.contains(event.target)) {
      setMenu(false);
    }
  });
  document.addEventListener("keydown", event => {
    if (event.key === "Escape" && menu.classList.contains("open")) {
      setMenu(false);
      menuBtn.focus();
    }
  });
  window.addEventListener("hashchange", updateActiveLink);
  updateActiveLink();
}

// Home always returns to the top, including when #home is already in the URL.
// Browsers can otherwise retain a restored scroll position or leave a repeated
// fragment navigation where it is.
function restoreHomePosition() {
  if (document.getElementById("home") && location.hash === "#home") {
    window.scrollTo({ top: 0, left: 0, behavior: "instant" });
  }
}
document.querySelectorAll('a[href="t.html#home"]').forEach(link => {
  link.addEventListener("click", event => {
    const target = new URL(link.href);
    if (target.pathname !== location.pathname) return;
    event.preventDefault();
    if (location.hash !== "#home") history.pushState(null, "", target.href);
    if (menu) menu.classList.remove("open");
    document.body.classList.remove("menu-open");
    if (menuBtn) menuBtn.setAttribute("aria-expanded", "false");
    restoreHomePosition();
    window.dispatchEvent(new Event("hashchange"));
  });
});
window.addEventListener("pageshow", restoreHomePosition);
window.addEventListener("hashchange", restoreHomePosition);

const calendarGrid = document.getElementById("calendarGrid");
if (calendarGrid) {
  const today = new Date();
  let month = new Date(today.getFullYear(), today.getMonth(), 1);
  // Illustrative dates only; replace with the school's approved schedule.
  const sampleEvents = [
    { month: 0, day: 5, title: "New academic year begins", type: "academic" },
    { month: 1, day: 12, title: "Annual sports day", type: "activity" },
    { month: 2, day: 15, endDay: 22, title: "First term assessment", type: "exam" },
    { month: 3, day: 8, title: "Parent–teacher meeting", type: "academic" },
    { month: 4, day: 18, title: "Science fair", type: "activity" },
    { month: 5, day: 14, endDay: 21, title: "Half-yearly examinations", type: "exam" },
    { month: 6, day: 12, title: "Second term begins", type: "academic" },
    { month: 7, day: 20, title: "Debate competition", type: "activity" },
    { month: 8, day: 17, endDay: 24, title: "Class assessment", type: "exam" },
    { month: 9, day: 8, title: "Parent–teacher meeting", type: "academic" },
    { month: 9, day: 22, title: "Science & innovation fair", type: "activity" },
    { month: 9, day: 28, title: "Annual exam preparation", type: "academic" },
    { month: 10, day: 15, endDay: 22, title: "Annual examinations", type: "exam" },
    { month: 11, day: 20, title: "Year-end results", type: "academic" }
  ];
  const filter = document.getElementById("calendarFilter");
  function renderCalendar() {
    const year = month.getFullYear();
    const monthIndex = month.getMonth();
    document.getElementById("calendarMonth").textContent = month.toLocaleDateString("en", { month: "long", year: "numeric" });
    const days = new Date(year, monthIndex + 1, 0).getDate();
    // Sample weekly closure: Friday and Saturday; confirm with the school.
    const holidays = Array.from({ length: days }, (_, index) => index + 1)
      .filter(day => [5, 6].includes(new Date(year, monthIndex, day).getDay()))
      .map(day => ({ day, title: "Weekly holiday", type: "holiday" }));
    const closedDays = new Set(holidays.map(event => event.day));
    // day/endDay are inclusive calendar dates. Exams skip closures without
    // extending the end date. Update these sample ranges with approved dates.
    const scheduledEvents = sampleEvents.filter(event => event.month === monthIndex)
      .flatMap(event => {
        const occurrences = [];
        for (let day = event.day; day <= Math.min(event.endDay || event.day, days); day++) {
          if (event.type === "exam" && closedDays.has(day)) continue;
          occurrences.push({ ...event, day });
        }
        return occurrences;
      });
    const monthEvents = [...scheduledEvents, ...holidays];
    const events = monthEvents.filter(event => filter.value === "all" || event.type === filter.value)
      .sort((a, b) => a.day - b.day);
    calendarGrid.replaceChildren();
    for (let i = 0; i < month.getDay(); i++) {
      const empty = document.createElement("div");
      empty.className = "calendar-day calendar-empty";
      empty.setAttribute("aria-hidden", "true");
      calendarGrid.append(empty);
    }
    for (let day = 1; day <= days; day++) {
      const cell = document.createElement("div");
      cell.className = "calendar-day";
      const dayEvents = events.filter(event => event.day === day);
      // Holiday fill takes priority if an illustrative event falls on a closure.
      const category = closedDays.has(day) ? "holiday" : ["exam", "activity", "academic"]
        .find(type => dayEvents.some(event => event.type === type));
      if (category) cell.classList.add(`day-${category}`);
      const date = document.createElement("time");
      date.dateTime = `${year}-${String(monthIndex + 1).padStart(2, "0")}-${String(day).padStart(2, "0")}`;
      date.textContent = day;
      date.setAttribute("aria-label", new Date(year, monthIndex, day).toLocaleDateString("en", { dateStyle: "full" }));
      if (year === today.getFullYear() && monthIndex === today.getMonth() && day === today.getDate()) {
        cell.classList.add("is-today");
        date.setAttribute("aria-current", "date");
      }
      cell.append(date);
      dayEvents.forEach(event => {
        const label = document.createElement("span");
        label.className = `calendar-event ${event.type}`;
        label.textContent = event.title;
        cell.append(label);
      });
      calendarGrid.append(cell);
    }
    const agenda = document.getElementById("calendarAgenda");
    agenda.replaceChildren();
    events.forEach(event => {
      const item = document.createElement("div");
      item.className = `agenda-item ${event.type}`;
      const date = document.createElement("span");
      date.className = "agenda-date";
      date.textContent = `${month.toLocaleDateString("en", { month: "short" })} ${event.day}`;
      const title = document.createElement("h4");
      title.textContent = event.title;
      item.append(date, title);
      agenda.append(item);
    });
    if (!events.length) agenda.textContent = "No sample events in this category this month.";
  }
  document.getElementById("calendarPrev").addEventListener("click", () => { month.setMonth(month.getMonth() - 1); renderCalendar(); });
  document.getElementById("calendarNext").addEventListener("click", () => { month.setMonth(month.getMonth() + 1); renderCalendar(); });
  document.getElementById("calendarToday").addEventListener("click", () => { month = new Date(today.getFullYear(), today.getMonth(), 1); renderCalendar(); });
  filter.addEventListener("change", renderCalendar);
  renderCalendar();
}
