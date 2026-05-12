// Scénario B — écriture : 500 POST, 5 VU
import http from "k6/http";
import { check } from "k6";
import { Trend, Rate } from "k6/metrics";

const responseSizeBytes = new Trend("response_size_bytes");
const errors = new Rate("errors");

const base = __ENV.BASE_URL || "http://localhost:8080";

export const options = {
  vus: 5,
  iterations: 500,
  summaryTrendStats: ["avg", "min", "med", "max", "p(75)", "p(90)", "p(95)", "p(99)"],
};

export default function () {
  const body = JSON.stringify({
    name: "Bench-Write-Fixed",
    type: "PRESSURE",
    location: "Line-2",
    unit: "bar",
    status: "ACTIVE",
    last_value: 1.02,
    last_reading_at: "2026-01-15T10:00:00Z",
  });
  const res = http.post(`${base}/sensors`, body, {
    headers: { "Content-Type": "application/json" },
  });
  responseSizeBytes.add(res.body.length);
  const ok = res.status === 201;
  errors.add(!ok);
  check(res, { "201": (r) => r.status === 201 });
}

export function handleSummary(data) {
  return {
    "benchmark/results/k6-rest-b-summary.json": JSON.stringify(data, null, 2),
  };
}
