(function () {
	function sync() {
		var list = document.getElementById('fehlende-nummern-list');
		var hidden = document.getElementById('fehlende-nummern-input');
		if (!list || !hidden) return;
		var nums = [];
		list.querySelectorAll('.sn-chip').forEach(function (chip) {
			var n = chip.getAttribute('data-number');
			if (n) nums.push(n);
		});
		hidden.value = nums.join(',');
	}

	function addNumber() {
		var input = document.getElementById('fehlende-nummer-input');
		var list = document.getElementById('fehlende-nummern-list');
		if (!input || !list) return;
		var val = input.value.trim();
		if (!val || !/^\d+$/.test(val)) return;
		if (list.querySelector('.sn-chip[data-number="' + val + '"]')) return;

		var chip = document.createElement('div');
		chip.className = 'sn-chip badge badge-lg gap-1';
		chip.setAttribute('data-number', val);
		chip.innerHTML =
			'<span>' + val + '</span>' +
			'<button type="button" class="sn-chip-remove btn btn-xs btn-ghost btn-circle">✕</button>';
		list.appendChild(chip);
		input.value = '';
		sync();
	}

	document.addEventListener('click', function (e) {
		if (e.target.closest('#add-fehlende-nummer')) {
			addNumber();
			return;
		}
		var remove = e.target.closest('.sn-chip-remove');
		if (remove) {
			var chip = remove.closest('.sn-chip');
			if (chip) chip.remove();
			sync();
		}
	});

	document.addEventListener('keydown', function (e) {
		if (e.target && e.target.id === 'fehlende-nummer-input' && e.key === 'Enter') {
			e.preventDefault();
			addNumber();
		}
	});
})();
