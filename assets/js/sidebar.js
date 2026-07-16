(function() {
  function initDrawer() {
    var desktopDrawer = document.getElementById('sidebar-drawer-desktop');
    if (desktopDrawer && window.innerWidth >= 1024) {
      desktopDrawer.checked = true;
    }
  }

  function updateSidebar() {
    var path = window.location.pathname;
    var details = document.querySelectorAll('.internal-sidebar details');
    details.forEach(function(d) {
      var summary = d.querySelector('summary');
      if (!summary) return;
      var catUrl = summary.getAttribute('hx-get') || '';
      if (path.startsWith(catUrl) && catUrl !== '/internal') {
        d.setAttribute('open', '');
      } else {
        d.removeAttribute('open');
      }
    });
    var links = document.querySelectorAll('.sidebar-link');
    links.forEach(function(link) {
      link.classList.remove('active');
      var href = link.getAttribute('href');
      if (href === path) {
        link.classList.add('active');
      }
    });
  }

  document.addEventListener('htmx:afterSwap', function(evt) {
    if (evt.detail.target && evt.detail.target.id === 'main-content') {
      updateSidebar();
    }
  });

  document.addEventListener('DOMContentLoaded', function() {
    initDrawer();
    updateSidebar();
  });
  document.addEventListener('htmx:load', updateSidebar);
})();
