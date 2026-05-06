function applyTheme(isDark) {
  const el = document.getElementById('theme-toggle');
  if (el) el.checked = isDark;
  document.documentElement.setAttribute('data-theme', isDark ? 'dark' : 'light');
}

const isDark = localStorage.theme === "dark" || (!("theme" in localStorage) && window.matchMedia("(prefers-color-scheme: dark)").matches);
applyTheme(isDark);

document.addEventListener("htmx:afterSwap", function() {
  const isDark = localStorage.theme === "dark" || (!("theme" in localStorage) && window.matchMedia("(prefers-color-scheme: dark)").matches);
  applyTheme(isDark);
});
