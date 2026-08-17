(function () {
	function addAttachmentRow() {
		var list = document.getElementById('attachment-list');
		if (!list) return;
		var row = document.createElement('div');
		row.className = 'flex items-center gap-2 w-full';
		row.innerHTML =
			'<input type="file" name="files" class="file-input file-input-bordered w-full" />' +
			'<button type="button" class="btn btn-sm btn-ghost btn-circle attachment-remove" title="Entfernen">✕</button>';
		list.appendChild(row);
	}

	document.addEventListener('click', function (e) {
		if (e.target.closest('#add-attachment')) {
			addAttachmentRow();
			return;
		}
		var chipRemove = e.target.closest('.chip-remove');
		if (chipRemove) {
			var chip = chipRemove.closest('.attachment-chip');
			if (chip) chip.remove();
			return;
		}
		var removeBtn = e.target.closest('.attachment-remove');
		if (removeBtn) {
			var row = removeBtn.parentElement;
			var list = row.parentElement;
			if (list.querySelectorAll('input[type=file]').length > 1) {
				row.remove();
			} else {
				var input = row.querySelector('input[type=file]');
				if (input) input.value = '';
			}
		}
	});
})();
