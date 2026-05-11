// BenchLab Dashboard — app.js

const REST_COLOR = '#4f83cc';
const GRPC_COLOR = '#e05c3a';

let chart = null;
let currentMetric = 'p95';
let restData = {};
let grpcData = {};

// ── Init ──────────────────────────────────────────────────────────────────────

document.addEventListener('DOMContentLoaded', async () => {
  await Promise.allSettled([loadSysInfo(), loadPayload(), loadAllResults()]);

  document.querySelectorAll('.tab').forEach(tab => {
    tab.addEventListener('click', () => {
      document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
      tab.classList.add('active');
      currentMetric = tab.dataset.metric;
      updateChart();
    });
  });

  document.getElementById('scenario-select').addEventListener('change', updateDetails);

  document.querySelectorAll('[data-target]').forEach(btn => {
    btn.addEventListener('click', () => runBenchmark(btn.dataset.target));
  });
});

// ── Data loading ──────────────────────────────────────────────────────────────

async function loadSysInfo() {
  try {
    const data = await fetchJSON('/api/sysinfo');
    const container = document.getElementById('sysinfo');
    const ramGb = (parseInt(data.ram_mb) / 1024).toFixed(0);
    const goVer = (data.tools?.go ?? '').split(' ')[2] ?? 'Go';
    container.innerHTML = [data.cpu_model, `${data.cpu_logical_cores} vCPU`, `${ramGb} GB RAM`, goVer]
      .map(t => `<span class="badge">${t}</span>`)
      .join('');
  } catch {
    document.getElementById('sysinfo').innerHTML = '<span class="badge">system-info non trouvé</span>';
  }
}

async function loadPayload() {
  try {
    const data = await fetchJSON('/api/payload');
    const container = document.getElementById('payload-card');
    const jsonB = data.json_sensor_bytes;
    const protoB = data.protobuf_sensor_bytes;
    const ratio = data.ratio_json_over_proto.toFixed(2);
    const pct = ((protoB / jsonB) * 100).toFixed(1);

    container.innerHTML = `
      <div class="payload-row">
        <span class="payload-label">JSON (REST)</span>
        <span class="payload-value payload-rest">${jsonB} octets</span>
      </div>
      <div class="payload-row">
        <span class="payload-label">Protobuf (gRPC)</span>
        <span class="payload-value payload-grpc">${protoB} octets</span>
      </div>
      <div class="payload-row">
        <span class="payload-label">Ratio JSON / Proto</span>
        <span class="payload-value payload-ratio">${ratio}×</span>
      </div>
      <div class="payload-bars">
        <div class="payload-bar-row">
          <span style="width:4rem;color:var(--rest)">REST</span>
          <div class="payload-bar-track">
            <div class="payload-bar-fill" style="width:100%;background:${REST_COLOR}"></div>
          </div>
          <span>${jsonB}B</span>
        </div>
        <div class="payload-bar-row">
          <span style="width:4rem;color:var(--grpc)">gRPC</span>
          <div class="payload-bar-track">
            <div class="payload-bar-fill" style="width:${pct}%;background:${GRPC_COLOR}"></div>
          </div>
          <span>${protoB}B</span>
        </div>
      </div>`;
  } catch {
    document.getElementById('payload-card').innerHTML =
      '<div style="color:var(--muted);font-size:.8rem">payload-size.json non trouvé</div>';
  }
}

async function loadAllResults() {
  await Promise.allSettled(
    ['a', 'b', 'c'].flatMap(s => [
      fetchJSON(`/api/results/k6-rest-${s}-summary.json`)
        .then(d => { restData[s] = parseK6(d); })
        .catch(() => {}),
      // Try canonical ghz file, fallback to -10 variant for scenario c
      fetchJSON(`/api/results/ghz-grpc-${s}.json`)
        .then(d => { grpcData[s] = parseGhz(d); })
        .catch(() =>
          fetchJSON(`/api/results/ghz-grpc-${s}-10.json`)
            .then(d => { grpcData[s] = parseGhz(d); })
            .catch(() => {})
        ),
    ])
  );
  buildChart();
  updateDetails();
}

// ── Parsers ───────────────────────────────────────────────────────────────────

function parseK6(data) {
  const m = data.metrics ?? {};
  const dur = m['http_req_duration{expected_response:true}'] ?? m['http_req_duration'];
  const v = dur?.values ?? {};
  return {
    avg: v.avg ?? 0,
    p50: v.med ?? 0,
    p90: v['p(90)'] ?? 0,
    p95: v['p(95)'] ?? 0,
    p99: v['p(99)'] ?? 0,
    rps: m['http_reqs']?.values?.rate ?? 0,
  };
}

function parseGhz(data) {
  const dist = data.latencyDistribution ?? [];
  const find = pct => {
    const e = dist.find(x => x.percentage === pct);
    return e ? e.latency / 1e6 : 0; // ns → ms
  };
  return {
    avg: (data.average ?? 0) / 1e6,
    p50: find(50),
    p90: find(90),
    p95: find(95),
    p99: find(99),
    rps: data.rps ?? 0,
  };
}

// ── Chart ─────────────────────────────────────────────────────────────────────

function buildChart() {
  const ctx = document.getElementById('comparison-chart').getContext('2d');
  if (chart) chart.destroy();
  chart = new Chart(ctx, {
    type: 'bar',
    data: getChartData(),
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: {
          labels: { color: '#8891b4', font: { size: 12 } },
        },
        tooltip: {
          callbacks: {
            label: ctx => {
              const v = ctx.parsed.y;
              return currentMetric === 'rps'
                ? `${ctx.dataset.label}: ${v.toFixed(0)} req/s`
                : `${ctx.dataset.label}: ${v.toFixed(2)} ms`;
            },
          },
        },
      },
      scales: {
        x: {
          ticks: { color: '#8891b4' },
          grid: { color: '#2e3147' },
        },
        y: {
          ticks: { color: '#8891b4' },
          grid: { color: '#2e3147' },
          title: {
            display: true,
            color: '#8891b4',
            text: currentMetric === 'rps' ? 'req/s' : 'ms',
          },
        },
      },
    },
  });
}

function getChartData() {
  return {
    labels: ['Scénario A', 'Scénario B', 'Scénario C'],
    datasets: [
      {
        label: 'REST',
        data: ['a', 'b', 'c'].map(s => restData[s]?.[currentMetric] ?? 0),
        backgroundColor: REST_COLOR + 'bb',
        borderColor: REST_COLOR,
        borderWidth: 1,
      },
      {
        label: 'gRPC',
        data: ['a', 'b', 'c'].map(s => grpcData[s]?.[currentMetric] ?? 0),
        backgroundColor: GRPC_COLOR + 'bb',
        borderColor: GRPC_COLOR,
        borderWidth: 1,
      },
    ],
  };
}

function updateChart() {
  if (!chart) return;
  chart.data = getChartData();
  chart.options.scales.y.title.text = currentMetric === 'rps' ? 'req/s' : 'ms';
  chart.update();
}

// ── Details table ─────────────────────────────────────────────────────────────

function updateDetails() {
  const s = document.getElementById('scenario-select').value;
  const rest = restData[s];
  const grpc = grpcData[s];

  const rows = [
    { label: 'Débit (RPS)',   key: 'rps', fmt: v => v.toFixed(0), higherWins: true },
    { label: 'Latence moy.', key: 'avg', fmt: v => v.toFixed(2) + ' ms' },
    { label: 'P50',           key: 'p50', fmt: v => v.toFixed(2) + ' ms' },
    { label: 'P90',           key: 'p90', fmt: v => v.toFixed(2) + ' ms' },
    { label: 'P95',           key: 'p95', fmt: v => v.toFixed(2) + ' ms' },
    { label: 'P99',           key: 'p99', fmt: v => v.toFixed(2) + ' ms' },
  ];

  const tbody = document.querySelector('#details-table tbody');
  if (!rest && !grpc) {
    tbody.innerHTML = '<tr><td colspan="3" style="color:var(--muted);font-size:.8rem;padding:.5rem">Pas de données pour ce scénario</td></tr>';
    return;
  }

  tbody.innerHTML = rows.map(({ label, key, fmt, higherWins }) => {
    const rv = rest?.[key] ?? 0;
    const gv = grpc?.[key] ?? 0;
    let restWin = '', grpcWin = '';
    if (rv > 0 && gv > 0) {
      const restBetter = higherWins ? rv > gv : rv < gv;
      if (restBetter) restWin = '<span class="winner">✓</span>';
      else grpcWin = '<span class="winner">✓</span>';
    }
    return `<tr>
      <td>${label}</td>
      <td class="val-rest">${fmt(rv)}${restWin}</td>
      <td class="val-grpc">${fmt(gv)}${grpcWin}</td>
    </tr>`;
  }).join('');
}

// ── Run benchmark ─────────────────────────────────────────────────────────────

async function runBenchmark(target) {
  const buttons = document.querySelectorAll('[data-target]');
  const status = document.getElementById('run-status');
  const label = document.getElementById('run-label');
  const output = document.getElementById('run-output');

  buttons.forEach(b => (b.disabled = true));
  output.classList.add('hidden');
  status.classList.remove('hidden');
  label.textContent = `make ${target} en cours… (peut prendre plusieurs minutes)`;

  try {
    const res = await fetch(`/api/run/${target}`, { method: 'POST' });
    const data = await res.json();

    output.textContent = data.output ?? '';
    output.classList.remove('hidden');
    label.textContent = data.ok ? 'Terminé — résultats mis à jour' : 'Erreur lors de l\'exécution';

    if (data.ok) {
      await loadAllResults();
    }
  } catch {
    label.textContent = 'Erreur réseau';
  } finally {
    buttons.forEach(b => (b.disabled = false));
  }
}

// ── Utils ─────────────────────────────────────────────────────────────────────

async function fetchJSON(url) {
  const res = await fetch(url);
  if (!res.ok) throw new Error(res.status);
  return res.json();
}
