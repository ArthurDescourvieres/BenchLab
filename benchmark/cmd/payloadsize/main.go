// payloadsize compare la taille d'un même capteur sérialisé en JSON (REST) et en Protobuf (gRPC).
package main

import (
	"encoding/json"
	"log"
	"os"

	sensorv1 "github.com/ArthurDescourvieres/BenchLab/grpc-service/gen/benchlab/sensor/v1"
	"github.com/ArthurDescourvieres/BenchLab/store"
	"google.golang.org/protobuf/proto"
)

type out struct {
	Description          string `json:"description"`
	JSONBytes            int    `json:"json_sensor_bytes"`
	ProtobufBytes        int    `json:"protobuf_sensor_bytes"`
	RatioJSONOverProto   float64 `json:"ratio_json_over_proto"`
	Note                 string `json:"note"`
}

func main() {
	// Valeurs alignées sur les scripts de benchmark (réponse typique GetSensor / GET).
	s := store.Sensor{
		ID:            "550e8400-e29b-41d4-a716-446655440000",
		Name:          "Bench-Setup",
		Type:          "TEMPERATURE",
		Location:      "Lab",
		Unit:          "°C",
		Status:        "ACTIVE",
		LastValue:     21.5,
		LastReadingAt: "2026-01-15T10:00:00Z",
		CreatedAt:     "2026-01-15T10:00:01Z",
	}

	jb, err := json.Marshal(s)
	if err != nil {
		log.Fatal(err)
	}

	pmsg := &sensorv1.Sensor{
		Id:            s.ID,
		Name:          s.Name,
		Type:          s.Type,
		Location:      s.Location,
		Unit:          s.Unit,
		Status:        s.Status,
		LastValue:     s.LastValue,
		LastReadingAt: s.LastReadingAt,
		CreatedAt:     s.CreatedAt,
	}
	pb, err := proto.Marshal(pmsg)
	if err != nil {
		log.Fatal(err)
	}

	jn, pn := len(jb), len(pb)
	ratio := float64(jn) / float64(pn)
	payload := out{
		Description:        "Taille du corps d'un Sensor (même contenu sémantique) : JSON HTTP vs message Protobuf seul (sans en-têtes HTTP/2 ni gRPC).",
		JSONBytes:          jn,
		ProtobufBytes:      pn,
		RatioJSONOverProto: ratio,
		Note:               "Pour le trafic réel gRPC, ajouter l'en-tête de frame (5 octets) + longueur du message par appel unary.",
	}

	const path = "benchmark/results/payload-size.json"
	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(payload); err != nil {
		log.Fatal(err)
	}
	log.Printf("écrit %s (JSON=%d o, Protobuf=%d o)", path, jn, pn)
}
