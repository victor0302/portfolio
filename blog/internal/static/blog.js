// theme toggle button — persists choice in localStorage.
// initial data-theme is set pre-render by the inline <head> script in
// layout.html.tmpl, so there's no flash of the wrong theme on load.

(function () {
  var html = document.documentElement;
  var label = document.getElementById('theme-label');
  var btn = document.getElementById('theme-toggle');

  function syncLabel() {
    if (label) label.textContent = html.getAttribute('data-theme') === 'dark' ? 'dark' : 'light';
  }
  syncLabel();

  if (btn) {
    btn.addEventListener('click', function () {
      var next = html.getAttribute('data-theme') === 'dark' ? 'light' : 'dark';
      html.setAttribute('data-theme', next);
      localStorage.setItem('theme', next);
      syncLabel();
    });
  }
})();
