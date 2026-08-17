window._toastSeenIds = window._toastSeenIds || new Set();

function dismissToast(el) {
	if (el.classList.contains('alert-error')) return;
	var id = el._dismissId || (el._dismissId = Math.random().toString(36).slice(2));
	if (window._toastSeenIds.has(id)) return;
	window._toastSeenIds.add(id);
	setTimeout(function() {
		if (el.parentElement) el.remove();
		window._toastSeenIds.delete(id);
	}, 5000);
}

function scanToasts(root) {
	var container = (root && root.id === 'toast-container') ? root
		: (root && root.getElementById) ? root.getElementById('toast-container')
		: document.getElementById('toast-container');
	if (!container) return;
	for (var i = 0; i < container.children.length; i++) {
		var el = container.children[i];
		if (el.tagName === 'DIV' && el.classList.contains('alert') && !el._dismissId) {
			dismissToast(el);
		}
	}
}

document.addEventListener('htmx:afterSwap', function(evt) {
	if (evt.detail.target && evt.detail.target.id === 'toast-container') {
		scanToasts(evt.detail.target);
	}
});

document.addEventListener('htmx:afterSettle', function(evt) {
	scanToasts();
});

scanToasts();