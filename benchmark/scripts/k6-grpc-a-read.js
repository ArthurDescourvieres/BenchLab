// Scénario A — lecture unitaire : 1000 itérations, 10 VU, GetSensor (gRPC)
// Miroir exact de k6-rest-a-read.js — même outil, même modèle VU, résultats comparables.
import grpc from "k6/net/grpc";
import { check } from "k6";
import { Trend, Rate } from "k6/metrics";

const responseSizeBytes = new Trend("response_size_bytes");
const errors = new Rate("errors");

const addr = __ENV.GRPC_ADDR || "localhost:9090";

const client = new grpc.Client();
client.load(["../.."], "grpc-service/proto/sensor.proto");

export const options = {
  vus: 10,
  iterations: 1000,
  summaryTrendStats: ["avg", "min", "med", "max", "p(90)", "p(95)", "p(99)"],
};

// Connexion persistante par VU : établie une seule fois et réutilisée sur toutes les itérations.
let connected = false;

export function setup() {
  client.connect(addr, { plaintext: true });
  const res = client.invoke("benchlab.sensor.v1.SensorService/CreateSensor", {
    sensor: {
      name: "Bench-Setup",
      type: "TEMPERATURE",
      location: "Lab",
      unit: "°C",
      status: "ACTIVE",
      last_value: 21.5,
      last_reading_at: "2026-01-15T10:00:00Z",
    },
  });
  if (!res || res.status !== grpc.StatusOK) {
    throw new Error(`setup CreateSensor failed: ${JSON.stringify(res)}`);
  }
  client.close();
  return { id: res.message.id };
}

export default function (data) {
  if (!connected) {
    client.connect(addr, { plaintext: true });
    connected = true;
  }

  const res = client.invoke("benchlab.sensor.v1.SensorService/GetSensor", {
    id: data.id,
  });

  responseSizeBytes.add(JSON.stringify(res.message).length);
  const ok = !!res && res.status === grpc.StatusOK;
  errors.add(!ok);
  check(res, { "StatusOK": (r) => r && r.status === grpc.StatusOK });
}

export function handleSummary(data) {
  return {
    "benchmark/results/k6-grpc-a-summary.json": JSON.stringify(data, null, 2),
  };
}
