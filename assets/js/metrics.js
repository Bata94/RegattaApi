const MAX_POINTS = 60;
let chartsInitialized = false;
let charts = {};
let metricsInterval = null;
let networkPrevSent = null;
let networkPrevRecv = null;

function initCharts() {
	const canvas = document.getElementById('cpu-chart');
	if (!canvas) return false;
	if (chartsInitialized) return true;
	chartsInitialized = true;

	const createChart = (ctxId, label, color) => {
		const ctx = document.getElementById(ctxId).getContext('2d');
		return new Chart(ctx, {
			type: 'line',
			data: {
				labels: Array(MAX_POINTS).fill(''),
				datasets: [{
					label: label,
					data: Array(MAX_POINTS).fill(0),
					borderColor: color,
					backgroundColor: color + '33',
					fill: true,
					tension: 0.3,
					pointRadius: 0
				}]
			},
			options: {
				responsive: true,
				maintainAspectRatio: true,
				animation: false,
				scales: { x: { display: false }, y: { beginAtZero: true } },
				plugins: { legend: { display: false } }
			}
		});
	};

	const createNetworkChart = () => {
		const ctx = document.getElementById('network-chart').getContext('2d');
		return new Chart(ctx, {
			type: 'line',
			data: {
				labels: Array(MAX_POINTS).fill(''),
				datasets: [
					{
						label: 'Sent (MB/s)',
						data: Array(MAX_POINTS).fill(0),
						borderColor: '#22c55e',
						backgroundColor: '#22c55e33',
						fill: true,
						tension: 0.3,
						pointRadius: 0
					},
					{
						label: 'Recv (MB/s)',
						data: Array(MAX_POINTS).fill(0),
						borderColor: '#ef4444',
						backgroundColor: '#ef444433',
						fill: true,
						tension: 0.3,
						pointRadius: 0
					}
				]
			},
			options: {
				responsive: true,
				maintainAspectRatio: true,
				animation: false,
				scales: { x: { display: false }, y: { beginAtZero: true } },
				plugins: { legend: { display: true, position: 'bottom' } }
			}
		});
	};

	charts.cpu = createChart('cpu-chart', 'CPU %', '#22c55e');
	charts.ram = createChart('ram-chart', 'RAM %', '#3b82f6');
	charts.connections = createChart('connections-chart', 'Verbindungen', '#f59e0b');
	charts.load = createChart('load-chart', 'Load', '#ef4444');
	charts.latency = createChart('latency-chart', 'Latenz (ms)', '#a855f7');
	charts.goroutines = createChart('goroutines-chart', 'Goroutines', '#f97316');
	charts.heap = createChart('heap-chart', 'Heap (MB)', '#8b5cf6');
	charts.network = createNetworkChart();
	return true;
}

function cleanupMetrics() {
	if (metricsInterval) {
		clearInterval(metricsInterval);
		metricsInterval = null;
	}
	chartsInitialized = false;
	charts = {};
	networkPrevSent = null;
	networkPrevRecv = null;
}

function formatBitsPerSec(mbValue) {
	const bits = mbValue * 8 * 1024 * 1024;
	if (bits >= 1024 * 1024) {
		return (bits / (1024 * 1024)).toFixed(2) + ' MBits/s';
	} else if (bits >= 1024) {
		return (bits / 1024).toFixed(2) + ' KBits/s';
	} else {
		return bits.toFixed(0) + ' Bits/s';
	}
}

async function fetchMetrics() {
	if (!charts.cpu) return;
	try {
		const res = await fetch('/metricsApi');
		const data = await res.json();

		charts.cpu.data.datasets[0].data.shift();
		charts.cpu.data.datasets[0].data.push(data.cpu_percent);
		charts.cpu.update('none');
		document.getElementById('cpu-value').textContent = data.cpu_percent.toFixed(1) + '%';

		charts.ram.data.datasets[0].data.shift();
		charts.ram.data.datasets[0].data.push(data.ram_percent);
		charts.ram.update('none');
		const usedGB = (data.ram_used / 1024 / 1024 / 1024).toFixed(1);
		const totalGB = (data.ram_total / 1024 / 1024 / 1024).toFixed(1);
		document.getElementById('ram-value').textContent = usedGB + ' / ' + totalGB + ' GB (' + data.ram_percent.toFixed(0) + '%)';

		charts.connections.data.datasets[0].data.shift();
		charts.connections.data.datasets[0].data.push(data.connections);
		charts.connections.update('none');
		document.getElementById('connections-value').textContent = data.connections;

		charts.load.data.datasets[0].data.shift();
		charts.load.data.datasets[0].data.push(data.load_1min);
		charts.load.update('none');
		document.getElementById('load-value').textContent = data.load_1min.toFixed(2);

		charts.latency.data.datasets[0].data.shift();
		charts.latency.data.datasets[0].data.push(data.latency_ms);
		charts.latency.update('none');
		document.getElementById('latency-value').textContent = data.latency_ms.toFixed(1) + ' ms';

		charts.goroutines.data.datasets[0].data.shift();
		charts.goroutines.data.datasets[0].data.push(data.goroutines);
		charts.goroutines.update('none');
		document.getElementById('goroutines-value').textContent = data.goroutines;

		charts.heap.data.datasets[0].data.shift();
		const heapAllocMB = (data.heap_alloc / 1024 / 1024).toFixed(0);
		const heapSysMB = (data.heap_sys / 1024 / 1024).toFixed(0);
		charts.heap.data.datasets[0].data.push(data.heap_alloc / 1024 / 1024);
		charts.heap.update('none');
		document.getElementById('heap-value').textContent = heapAllocMB + ' / ' + heapSysMB + ' MB';

		if (networkPrevSent !== null && networkPrevRecv !== null) {
			const sentDelta = (data.network_bytes_sent - networkPrevSent) / 1024 / 1024;
			const recvDelta = (data.network_bytes_recv - networkPrevRecv) / 1024 / 1024;
			charts.network.data.datasets[0].data.shift();
			charts.network.data.datasets[0].data.push(sentDelta);
			charts.network.data.datasets[1].data.shift();
			charts.network.data.datasets[1].data.push(recvDelta);
			charts.network.update('none');
			document.getElementById('network-value').textContent = 'Sent: ' + formatBitsPerSec(sentDelta) + ' | Recv: ' + formatBitsPerSec(recvDelta);
		}
		networkPrevSent = data.network_bytes_sent;
		networkPrevRecv = data.network_bytes_recv;
	} catch (e) {
		console.error('Failed to fetch metrics:', e);
	}
}

function startPolling() {
	if (metricsInterval) return;
	fetchMetrics();
	metricsInterval = setInterval(fetchMetrics, 1000);
}

function waitForCanvasAndInit() {
	let attempts = 0;
	const maxAttempts = 50;
	function tryInit() {
		if (initCharts()) {
			startPolling();
			return;
		}
		attempts++;
		if (attempts < maxAttempts) {
			setTimeout(tryInit, 100);
		}
	}
	tryInit();
}

if (document.readyState === 'loading') {
	document.addEventListener('DOMContentLoaded', waitForCanvasAndInit);
} else {
	waitForCanvasAndInit();
}

document.body.addEventListener('htmx:afterSwap', function(evt) {
	const target = evt.detail.target;
	if (target.querySelector && target.querySelector('#cpu-chart')) {
		chartsInitialized = false;
		charts = {};
		waitForCanvasAndInit();
	}
});

document.body.addEventListener('htmx:beforeSwap', function(evt) {
	if (evt.detail.target.querySelector && evt.detail.target.querySelector('#cpu-chart')) {
		cleanupMetrics();
	}
});