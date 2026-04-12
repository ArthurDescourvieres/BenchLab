// Scénario A — lecture unitaire : 1000 itérations, 10 VU, GET /sensors/{id}
import http from "k6/http";
import { check } from "k6";

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
  check(res, { "200": (r) => r.status === 200 });
}

export function handleSummary(data) {
  return {
    "benchmark/results/k6-rest-a-summary.json": JSON.stringify(data, null, 2),
  };
}
