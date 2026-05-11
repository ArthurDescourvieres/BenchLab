// Scénario A — lecture unitaire : 1000 itérations, 10 VU, GET /sensors/{id}
import http from "k6/http";
import { check } from "k6";
import { Trend, Rate } from "k6/metrics";

const responseSizeBytes = new Trend("response_size_bytes");
const errors = new Rate("errors");

const base = __ENV.BASE_URL || "http://localhost:8080";

const payload = JSON.stringify({
  name: "Bench-Setup",
  type: "TEMPERATURE",
  location: "Lab",
  unit: "°C",
  status: "ACTIVE",
  last_value: 21.5,
  last_reading_at: "2026-01-15T10:00:00Z",
});

export const options = {
  vus: 10,
  iterations: 1000,
  summaryTrendStats: ["avg", "min", "med", "max", "p(75)", "p(90)", "p(95)", "p(99)"],
};

export function setup() {
  const res = http.post(`${base}/sensors`, payload, {
    headers: { "Content-Type": "application/json" },
  });
  if (res.status !== 201) {
    throw new Error(`setup POST failed: ${res.status} ${res.body}`);
  }
  const body = res.json();
  return { id: body.id };
}

export default function (data) {
  const res = http.get(`${base}/sensors/${data.id}`);
  responseSizeBytes.add(res.body.length);
  const ok = res.status === 200;
  errors.add(!ok);
  check(res, { "200": (r) => r.status === 200 });
}

function restSummary(data) {
  const sep = "────────────────────────────────────────────────────────────";
  const m = data.metrics;
  const dur = data.state.testRunDurationMs / 1000;
  const count = m.iterations.values.count;
  const rps = m.iterations.values.rate;
  const d = m.http_req_duration.values;
  const ms = (v) => `${v.toFixed(2)} ms`;
  const pad = (s, n) => s.padEnd(n);
  const failed = m.http_req_failed ? m.http_req_failed.values.passes : 0;

  return [
    sep,
    `  Requêtes  : ${count}  |  RPS : ${Math.round(rps)} req/s  |  Durée : ${dur.toFixed(2)} s`,
    `  Latence   : avg=${pad(ms(d.avg), 12)}  min=${pad(ms(d.min), 12)}  max=${ms(d.max)}`,
    `              p50    ${ms(d.med)}`,
    `              p75    ${ms(d["p(75)"])}`,
    `              p90    ${ms(d["p(90)"])}`,
    `              p95    ${ms(d["p(95)"])}`,
    `              p99    ${ms(d["p(99)"])}`,
    failed > 0 ? `  Erreurs   : ${failed} requête(s) échouée(s)` : "  Erreurs   : aucune",
    `  Statuts   : ${failed === 0 ? count + " OK" : count - failed + " OK, " + failed + " KO"}`,
    sep,
  ].join("\n");
}

export function handleSummary(data) {
  return {
    "benchmark/results/k6-rest-a-summary.json": JSON.stringify(data, null, 2),
    "-": restSummary(data),
  };
}
