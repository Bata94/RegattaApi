'use strict';

window.ZeitnahmeShared = (function () {
  var airHorn = new Audio('/assets/sound/airHorn.mp3');
  airHorn.preload = 'auto';

  var wasmLoading = false;

  function playAirHorn() {
    airHorn.currentTime = 0;
    airHorn.play().catch(function () {
      var flash = document.createElement('div');
      flash.style.cssText = 'position:fixed;inset:0;background:#ff3b30;opacity:0.4;z-index:9999;pointer-events:none;transition:opacity 0.1s';
      document.body.appendChild(flash);
      setTimeout(function () { flash.style.opacity = '0'; }, 100);
      setTimeout(function () { flash.remove(); }, 300);
    });
  }

  function getWSBadgeHTML() {
    if (window.__wasm_getWSConnected && window.__wasm_getWSConnected()) {
      var latMs = window.__wasm_getLatencyMs ? window.__wasm_getLatencyMs() : -1;
      var latText = '';
      if (latMs > 0) {
        latText = ' (' + latMs + 'ms)';
      } else if (latMs === 0) {
        latText = ' (<1ms)';
      }
      return '<span id="ws-status" class="badge badge-success">Verbunden' + latText + '</span>';
    }
    var reconnecting = window.__wasm_isReconnecting ? window.__wasm_isReconnecting() : false;
    if (reconnecting) {
      return '<span id="ws-status" class="badge badge-error gap-2">Getrennt \u2014 Verbindung wird hergestellt...<span class="loading loading-spinner loading-xs"></span></span>';
    }
    return '<span id="ws-status" class="badge badge-error cursor-pointer">Getrennt \u2014 klicken zum Verbinden</span>';
  }

  function updateWSBadge(containerId) {
    var container = document.getElementById(containerId || 'ws-badge-container');
    if (!container) return;
    container.innerHTML = getWSBadgeHTML();
    var badge = document.getElementById('ws-status');
    if (badge && (!window.__wasm_getWSConnected || !window.__wasm_getWSConnected())) {
      badge.addEventListener('click', function () {
        if (window.__wasm_forceReconnect) window.__wasm_forceReconnect();
      });
    }
  }

  function formatTime(iso) {
    return new Date(iso).toLocaleTimeString('de-DE', {
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit'
    });
  }

  function showToast(msg, cls) {
    var container = document.getElementById('toast-container');
    if (!container) return;
    var t = document.createElement('div');
    t.className = 'alert ' + (cls || 'alert-info') + ' shadow-lg';
    t.textContent = msg;
    container.appendChild(t);
    setTimeout(function () {
      t.remove();
    }, 4000);
  }

  function failApp(appId, msg) {
    var app = document.getElementById(appId);
    if (app) app.innerHTML = '<div class="alert alert-error">' + msg + '</div>';
  }

  function bootApp(onReady, appId) {
    if (window.__wasm_initialized) {
      onReady();
    } else {
      loadWasm(onReady, appId);
    }
  }

  function loadWasm(onReady, appId) {
    if (wasmLoading) return;
    wasmLoading = true;

    if (!WebAssembly.instantiateStreaming) {
      failApp(appId, 'Dein Browser unterst\u00fctzt kein WebAssembly.');
      return;
    }

    var go = new Go();
    WebAssembly.instantiateStreaming(fetch('/public/wasm/zeitnahme.wasm'), go.importObject)
      .then(function (result) {
        go.run(result.instance);
        window.__wasm_initialized = true;
        onReady();
      })
      .catch(function (err) {
        console.error('WASM load failed', err);
        wasmLoading = false;
        failApp(appId, 'WASM konnte nicht geladen werden: ' + err.message);
      });
  }

  return {
    playAirHorn: playAirHorn,
    getWSBadgeHTML: getWSBadgeHTML,
    updateWSBadge: updateWSBadge,
    formatTime: formatTime,
    showToast: showToast,
    failApp: failApp,
    bootApp: bootApp,
    loadWasm: loadWasm
  };
})();

(function drainZeitnahmeBoots() {
  var queue = window.__zeitnahmeBootQueue || [];
  window.__zeitnahmeBootQueue = [];
  for (var i = 0; i < queue.length; i++) {
    ZeitnahmeShared.bootApp(queue[i].fn, queue[i].appId);
  }
})();
