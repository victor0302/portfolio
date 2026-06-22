// runs on every page. theme is set pre-render by an inline <head> script
// in each html file, so there's no flash of the wrong theme on load.

// live clock in the topbar
function clock() {
  const el = document.getElementById('clock');
  if (!el) return;
  const d = new Date();
  const pad = n => String(n).padStart(2, '0');
  el.textContent = pad(d.getHours()) + ':' + pad(d.getMinutes()) + ':' + pad(d.getSeconds());
}
clock();
setInterval(clock, 1000);

// theme toggle button — persists choice in localStorage
const __html = document.documentElement;
const __label = document.getElementById('theme-label');
const __btn = document.getElementById('theme-toggle');
function __syncThemeLabel() {
  if (__label) __label.textContent = __html.getAttribute('data-theme') === 'dark' ? 'dark' : 'light';
}
__syncThemeLabel();
if (__btn) {
  __btn.addEventListener('click', () => {
    const next = __html.getAttribute('data-theme') === 'dark' ? 'light' : 'dark';
    __html.setAttribute('data-theme', next);
    localStorage.setItem('theme', next);
    __syncThemeLabel();
  });
}
