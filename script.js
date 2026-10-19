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
        (currentPage === "gallery.html" || target.hash === (location.hash || "#home"));
      link.classList.toggle("active", active);
      if (active) link.setAttribute("aria-current", currentPage === "gallery.html" ? "page" : "location");
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
