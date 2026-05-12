// Scénario A bis — lecture unitaire REST avec compression gzip activée côté serveur.
// Mêmes paramètres que k6-rest-a-read.js (1000 itérations, 10 VU, GET /sensors/{id})
// mais le client demande explicitement gzip via Accept-Encoding.
// IMPORTANT : le service REST doit être démarré avec REST_GZIP=1 :
//   REST_GZIP=1 go run ./rest-service        (Bash / Linux / macOS)
//   $env:REST_GZIP=1; go run ./rest-service  (PowerShell Windows)
// La métrique response_size_bytes mesure la taille du corps tel que reçu par k6.
// Si le client k6 décompresse automatiquement (cas par défaut sur certains builds),
// la valeur sera celle décompressée — voir aussi data_received pour la taille filaire.
import http from "k6/http";
import { check } from "k6";
import { Trend, Rate } from "k6/metrics";

const responseSizeBytes = new Trend("response_size_bytes");
const errors = new Rate("errors");

const base = __ENV.BASE_URL || "http://localhost:8080";

const payload = JSON.stringify({
  name: "Bench-Setup-Gzip",
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
  return { id: res.json().id };
}

export default function (data) {
  const res = http.get(`${base}/sensors/${data.id}`, {
    headers: { "Accept-Encoding": "gzip" },
    compression: "gzip",
  });
  responseSizeBytes.add(res.body.length);
  const ok = res.status === 200;
  errors.add(!ok);
  check(res, { "200": (r) => r.status === 200 });
}

export function handleSummary(data) {
  return {
    "benchmark/results/k6-rest-a-gzip-summary.json": JSON.stringify(data, null, 2),
  };
}
