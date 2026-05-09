// Scénario C — charge progressive sur GET (10 → 100 VU).
//
// Stages k6 : montée linéaire 10 → 30 → 60 → 100 VU (45s par palier), puis
// 30s de stabilisation à 100 VU. Total : 4 min. À comparer avec
// grpc-c-ramp.sh qui utilise des paliers discrets (limite native de ghz) :
// les bornes (10 et 100) sont identiques, l'approche diffère car les outils
// le sont. La divergence est documentée dans benchmark/BENCH-CONDITIONS.md.
import http from "k6/http";
import { check } from "k6";
import { Trend, Rate } from "k6/metrics";

const responseSizeBytes = new Trend("response_size_bytes");
const errors = new Rate("errors");

const base = __ENV.BASE_URL || "http://localhost:8080";

const createPayload = JSON.stringify({
  name: "Bench-Ramp",
  type: "VIBRATION",
  location: "Motor-1",
  unit: "mm/s",
  status: "ACTIVE",
  last_value: 0.4,
  last_reading_at: "2026-01-15T10:00:00Z",
});

export const options = {
  stages: [
    { duration: "45s", target: 10 },
    { duration: "45s", target: 30 },
    { duration: "45s", target: 60 },
    { duration: "45s", target: 100 },
    { duration: "30s", target: 100 },
  ],
  summaryTrendStats: ["avg", "min", "med", "max", "p(90)", "p(95)", "p(99)"],
};

export function setup() {
  const res = http.post(`${base}/sensors`, createPayload, {
    headers: { "Content-Type": "application/json" },
  });
  if (res.status !== 201) {
    throw new Error(`setup POST failed: ${res.status} ${res.body}`);
  }
  return { id: res.json().id };
}

export default function (data) {
  const res = http.get(`${base}/sensors/${data.id}`);
  responseSizeBytes.add(res.body.length);
  const ok = res.status === 200;
  errors.add(!ok);
  check(res, { "200": (r) => r.status === 200 });
}

export function handleSummary(data) {
  return {
    "benchmark/results/k6-rest-c-summary.json": JSON.stringify(data, null, 2),
  };
}
