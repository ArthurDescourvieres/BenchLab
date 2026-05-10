// Scénario B — écriture : 500 POST, 5 VU, CreateSensor (gRPC)
// Miroir exact de k6-rest-b-write.js — même outil, même modèle VU, résultats comparables.
import grpc from "k6/net/grpc";
import { check } from "k6";
import { Trend, Rate } from "k6/metrics";

const responseSizeBytes = new Trend("response_size_bytes");
const errors = new Rate("errors");

const addr = __ENV.GRPC_ADDR || "localhost:9090";

const client = new grpc.Client();
client.load(["../.."], "grpc-service/proto/sensor.proto");

export const options = {
  vus: 5,
  iterations: 500,
  summaryTrendStats: ["avg", "min", "med", "max", "p(90)", "p(95)", "p(99)"],
};

// Connexion persistante par VU.
let connected = false;

export default function () {
  if (!connected) {
    client.connect(addr, { plaintext: true });
    connected = true;
  }

  const res = client.invoke("benchlab.sensor.v1.SensorService/CreateSensor", {
    sensor: {
      name: "Bench-Write-Fixed",
      type: "PRESSURE",
      location: "Line-2",
      unit: "bar",
      status: "ACTIVE",
      last_value: 1.02,
      last_reading_at: "2026-01-15T10:00:00Z",
    },
  });

  responseSizeBytes.add(JSON.stringify(res.message).length);
  const ok = !!res && res.status === grpc.StatusOK;
  errors.add(!ok);
  check(res, { "StatusOK": (r) => r && r.status === grpc.StatusOK });
}

export function handleSummary(data) {
  return {
    "benchmark/results/k6-grpc-b-summary.json": JSON.stringify(data, null, 2),
  };
}
