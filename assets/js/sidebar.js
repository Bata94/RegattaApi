(function() {
  var sidebarWidthMin = 12;
  var sidebarWidthMax = 25;
  var resizing = false;
  var startX = 0;
  var startWidth = 0;
  var sidebarWrap = null;
  var resizeHandle = null;

  function applySidebarWidth(w) {
    if (!sidebarWrap) return;
    var wRem = w.toFixed(2) + 'rem';
    sidebarWrap.style.width = wRem;
    document.documentElement.style.setProperty('--sidebar-width', wRem);
  }

  function initDrawer() {
    sidebarWrap = document.getElementById('desktop-sidebar-wrap');
    var checkbox = document.getElementById('sidebar-drawer-desktop');

    if (checkbox && sidebarWrap && window.innerWidth >= 1024) {
      checkbox.checked = true;
    }

    var savedWidth = localStorage.getItem('sidebarWidth');
    if (savedWidth) {
      var w = parseFloat(savedWidth);
      if (w >= sidebarWidthMin && w <= sidebarWidthMax) {
        applySidebarWidth(w);
        return;
      }
    }
    applySidebarWidth(18);
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

  function updatePageTitle() {
    var h = document.querySelector('#main-content h1') || document.querySelector('#main-content h2');
    if (h && h.textContent) {
      document.title = h.textContent.trim() + ' | MRG Regatta';
    }
  }

  function updateTopBarTitle() {
    var h = document.querySelector('#main-content h1') || document.querySelector('#main-content h2');
    var topBarH2 = document.querySelector('.topbar-title');
    if (h && h.textContent && topBarH2) {
      topBarH2.textContent = h.textContent.trim();
    }
  }

  function startResize(e) {
    if (window.innerWidth < 1024) return;
    if (!sidebarWrap) return;
    e.preventDefault();
    resizing = true;
    startX = e.clientX || (e.touches && e.touches[0].clientX);
    var currentWidth = parseFloat(sidebarWrap.style.width) || 18;
    startWidth = currentWidth;
    document.body.style.cursor = 'col-resize';
    document.body.style.userSelect = 'none';
  }

  function doResize(e) {
    if (!resizing) return;
    var clientX = e.clientX || (e.touches && e.touches[0].clientX);
    if (clientX === undefined) return;
    var delta = clientX - startX;
    var deltaRem = delta / 16;
    var newWidth = startWidth + deltaRem;
    newWidth = Math.max(sidebarWidthMin, Math.min(sidebarWidthMax, newWidth));
    applySidebarWidth(newWidth);
  }

  function stopResize() {
    if (!resizing) return;
    resizing = false;
    document.body.style.cursor = '';
    document.body.style.userSelect = '';
    var w = parseFloat(sidebarWrap.style.width) || 18;
    localStorage.setItem('sidebarWidth', w.toFixed(2));
  }

  function initResize() {
    resizeHandle = document.getElementById('resize-handle');
    if (!resizeHandle) return;
    resizeHandle.addEventListener('mousedown', startResize);
    resizeHandle.addEventListener('touchstart', startResize, { passive: true });
  }

  function initDesktopSidebarToggle() {
    var checkbox = document.getElementById('sidebar-drawer-desktop');
    if (!checkbox || !sidebarWrap) return;
    checkbox.addEventListener('change', function() {
      if (checkbox.checked) {
        var savedWidth = localStorage.getItem('sidebarWidth');
        var w = savedWidth ? parseFloat(savedWidth) : 18;
        applySidebarWidth(w);
      } else {
        sidebarWrap.style.width = '0rem';
      }
    });
  }

  document.addEventListener('mousemove', doResize);
  document.addEventListener('mouseup', stopResize);
  document.addEventListener('touchmove', doResize, { passive: true });
  document.addEventListener('touchend', stopResize);

  document.addEventListener('htmx:afterSwap', function(evt) {
    if (evt.detail.target && evt.detail.target.id === 'main-content') {
      updatePageTitle();
      updateTopBarTitle();
      updateSidebar();
    }
  });

  document.addEventListener('htmx:load', function() {
    updatePageTitle();
    updateTopBarTitle();
    sidebarWrap = document.getElementById('desktop-sidebar-wrap');
    initResize();
    initDesktopSidebarToggle();
    updateSidebar();
  });

  document.addEventListener('DOMContentLoaded', function() {
    initDrawer();
    initResize();
    initDesktopSidebarToggle();
    updateSidebar();
    updatePageTitle();
  });

  document.addEventListener('htmx:load', function() {
    sidebarWrap = document.getElementById('desktop-sidebar-wrap');
    initResize();
    initDesktopSidebarToggle();
    updateSidebar();
    updatePageTitle();
  });
})();
