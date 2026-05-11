# BenchLab — Briefing pour plan de veille (REST vs gRPC)

> Document autoporteur destiné à un LLM partenaire chargé de produire un plan de veille technologique.
> Toutes les informations utiles (contexte projet, méthodologie, résultats chiffrés) sont rassemblées ici — aucun fichier externe n'est requis.
> Date de collecte des mesures : **2026-05-09 / 2026-05-10**.

---

## 1. Contexte du projet

**BenchLab** est un laboratoire de benchmarking comparant deux protocoles de communication microservices sur une logique métier identique :

- **REST** — JSON sur HTTP/1.1
- **gRPC** — Protobuf sur HTTP/2

**Cas d'usage** : SignalWatch — supervision de capteurs industriels.
Modèle `Sensor` partagé : `id`, `name`, `type` (TEMPERATURE/PRESSURE/VIBRATION), `location`, `unit`, `status` (ACTIVE/INACTIVE/MAINTENANCE), `last_value`, `last_reading_at`, `created_at`.

**Pourquoi cela intéresse une veille techno**
Le sujet « protocoles inter-services » est en mouvement constant : montée de gRPC dans le cloud-native, alternatives Connect-RPC / gRPC-Web, retour de HTTP/3, nouveaux formats binaires (Cap'n Proto, FlatBuffers), pression observabilité (OpenTelemetry) et sécurité (mTLS, JWT, OAuth2). Disposer d'une référence chiffrée sur le même socle métier permet d'évaluer ces évolutions sans recommencer de zéro.

---

## 2. Architecture testée

| Composant | Détail |
|-----------|--------|
| `rest-service/` | Gin v1.12.0 — port `:8080` — gzip optionnel via `REST_GZIP=1` — endpoint health `GET /health` — CRUD `/sensors` |
| `grpc-service/` | grpc-go v1.71.1 + protobuf v1.36.10 — port `:9090` — proto versionné `benchlab.sensor.v1` — RPCs : `CreateSensor`, `GetSensor`, `ListSensors`, `UpdateSensor`, `DeleteSensor` |
| `internal/sensorsvc/` | **Logique métier partagée** : validations types, statuts, timestamps RFC3339. Garantit que la différence mesurée vient bien du protocole, pas de l'algo. |
| `store/` | `MemoryStore` (map + RWMutex). Pas d'I/O disque → isole le coût protocolaire de la persistance. |

Chaque service tourne dans un processus indépendant. Logique métier strictement identique des deux côtés.

---

## 3. Conditions de test

| Élément | Valeur |
|---------|--------|
| Machine | Intel Core Ultra 7 155H — 22 cœurs logiques — 32 Go RAM |
| OS | Windows 11 Home (build 26200) |
| Go | 1.26.2 (windows/amd64) |
| k6 | v1.7.1 |
| ghz | dev |
| protoc | libprotoc 34.1 |
| Réseau | **localhost loopback** — aucune latence réseau réelle |
| Topologie | Service et injecteur sur la même machine, processus séparés |
| Bruit | Aucun proxy/VPN, processus concurrents minimisés |

> ⚠ Limite majeure à garder en tête pour la veille : tous les chiffres sont en **boucle locale**. Ils mesurent le coût CPU/sérialisation, **pas** le coût réseau (latence WAN, MTU, TLS handshake, perte de paquets).

---

## 4. Méthodologie

3 scénarios appliqués aux 2 protocoles, plus une variante REST+gzip.

| Scénario | Charge | Outil REST | Outil gRPC |
|----------|--------|------------|------------|
| **A — Read** | 1 000 itérations, 10 VU/connexions | k6 (`GET /sensors/{id}`) | ghz (`GetSensor`) |
| **B — Write** | 500 itérations, 5 VU/connexions | k6 (`POST /sensors`) | ghz (`CreateSensor`) |
| **C — Ramp** | montée 10 → 100 concurrents | k6 stages continues sur ~210 s | ghz : 5 paliers fixes (10, 25, 50, 75, 100) × 5 000 req |

**Outillage transverse**
- `monitor.exe` (utilitaire maison) attache un PID, échantillonne CPU%, RAM RSS et nombre de threads à 1 Hz, sortie CSV.
- Comparaison de payload via `benchmark/cmd/payloadsize` : sérialisation d'un même `Sensor` en JSON et en Protobuf, sans en-têtes HTTP.

**Note méthodologique sur le scénario C**
- k6 (REST) utilise des `stages` → montée *continue*, donc tous les paliers sont mélangés dans une seule sortie agrégée.
- ghz (gRPC) ne sait pas faire de ramp natif → simulé par 5 runs successifs à concurrence fixe. Permet de tracer une courbe latence/RPS par palier mais pas de transition continue.

---

## 5. Résultat — taille de payload

| Format | Taille du Sensor |
|--------|-------------------|
| JSON (corps HTTP) | **230 octets** |
| Protobuf (message seul) | **135 octets** |
| Ratio JSON / Protobuf | **1,70×** |

Sur le fil gRPC réel, ajouter +5 octets de frame HTTP/2 par appel unary. Le gain net reste de **~40 %** sur la taille utile.

---

## 6. Résultats — Scénario A (Read, 1 000 reqs, 10 concurrents)

| Variante | avg | p95 | p99 | max | RPS | Réponse | Erreurs |
|----------|-----|-----|-----|-----|-----|---------|---------|
| **REST**       | 0,41 ms | 1,26 ms | 2,24 ms | 4,61 ms | **14 324** | 225 B | 0 |
| **gRPC**       | 1,40 ms | 3,25 ms | 7,33 ms | 35,1 ms | 5 241 | — | 0 |
| **REST+gzip**  | 1,19 ms | 3,07 ms | 6,59 ms | 12,97 ms | 6 074 | 230 B | 0 |

**Lecture rapide**
- À très faible charge sur loopback, REST écrase gRPC en latence et en RPS — le coût d'établissement gRPC (HTTP/2 framing, codec) ne s'amortit pas.
- gzip dégrade les performances sur ce petit payload (CPU compression > gain réseau qui n'existe pas en loopback).

---

## 7. Résultats — Scénario B (Write, 500 reqs, 5 concurrents)

| Variante | avg | p95 | p99 | max | RPS | Erreurs |
|----------|-----|-----|-----|-----|-----|---------|
| **REST** | 0,44 ms | 1,20 ms | 1,85 ms | 3,47 ms | 6 506 | 0 |
| **gRPC** | 0,81 ms | 1,53 ms | 2,80 ms | 20,67 ms | 4 087 | 0 |

REST conserve l'avantage en écriture sur charge faible. Écart plus serré qu'en lecture.

---

## 8. Résultats — Scénario C (montée en charge)

### REST (k6, ramp 10 → 100 VU sur ~210 s)

| Métrique | Valeur |
|----------|--------|
| Requêtes totales | 7 851 155 |
| RPS moyen | **37 385** |
| avg | 1,10 ms |
| p95 | 3,46 ms |
| p99 | 8,02 ms |
| max | 65,98 ms |
| Erreurs | 0 |
| VU max | 100 |

### gRPC (ghz, paliers fixes)

| Concurrence | Reqs | RPS | avg | p95 | p99 |
|------------:|-----:|----:|----:|----:|----:|
| 10  | 5 000 | **7 530**  | 0,82 ms | 1,86 ms | 2,96 ms |
| 25  | 5 000 | 10 600 | 1,48 ms | 4,06 ms | 6,94 ms |
| 50  | 5 000 | 11 298 | 3,10 ms | 9,63 ms | 22,04 ms |
| 75  | 5 000 | 16 613 | 3,05 ms | 9,53 ms | 29,61 ms |
| 100 | 5 000 | **17 032** | 4,08 ms | 13,23 ms | 34,07 ms |

**Lecture rapide**
- gRPC a une scalabilité plus régulière en RPS jusqu'à 100 connexions, mais la **p99 explose** (3 → 34 ms) — file d'attente derrière les connexions HTTP/2 multiplexées.
- REST en ramp continue passe à 37 k RPS — limite plutôt côté CPU/k6 que côté serveur.
- Comparaison directe imparfaite (méthode de ramp différente) — utilisable comme **ordre de grandeur**, pas comme benchmark côte-à-côte.

---

## 9. Résultats — ressources CPU/RAM (monitor maison)

| Test | Durée capturée | CPU avg | CPU max | RAM avg | RAM max | Threads max |
|------|---------------:|--------:|--------:|--------:|--------:|------------:|
| REST scénario C | 209 s | 325 % | 426 % | 27,4 Mo | 31,8 Mo | 30 |
| gRPC scénario C | 4 s (paliers) | 27 % | 53 % | 17,5 Mo | 23,3 Mo | 28 |
| REST A/B/A-gzip, gRPC A/B | 1 sample | 14–16 % / 6 % | — | 12–13 Mo | — | 8 |

> CPU exprimé en pourcentage Windows : 100 % = 1 cœur logique sur 22.
> Les tests A et B sont trop courts (< 1 s) pour produire plus d'un échantillon — leur monitoring est indicatif, pas représentatif.
> Mémoire toujours sous 35 Mo : `MemoryStore` n'alloue que des entrées de map.

**Lecture rapide**
- REST tape plus fort sur le CPU sous forte charge (parsing/encodage JSON par requête).
- gRPC reste plus économe — codec binaire moins coûteux, multiplexing HTTP/2.

---

## 10. Synthèse pour le plan de veille

### Conclusions des mesures

1. **REST gagne en latence absolue à faible charge sur loopback.** Le surcoût gRPC (HTTP/2, codec, métadata) ne s'amortit pas sur des appels rares.
2. **gRPC est plus efficient côté CPU/RAM** — sérialisation binaire, taille de message ~40 % plus petite, scaling plus régulier.
3. **gzip est contre-productif** sur petits payloads en loopback — coût compression > gain transport. À reconsidérer en WAN.
4. **p99 gRPC dégrade vite** au-delà de 50 connexions concurrentes — point d'attention pour la robustesse SLO.
5. **Limite majeure** : aucun chiffre ne reflète une vraie latence réseau. En WAN, la donne change : gRPC bénéficie davantage de la compression, REST souffre du keep-alive limité.

### Axes de veille technologique suggérés

| Axe | Pourquoi maintenant |
|-----|---------------------|
| **HTTP/3 (QUIC) côté REST et gRPC** | Maturation 2024-2026, adoption CDN/cloud, moins de head-of-line blocking |
| **Connect-RPC / gRPC-Web** | Alternative gRPC compatible navigateur et reverse proxies HTTP/1.1 |
| **Formats binaires alternatifs** | Cap'n Proto (zéro-copy), FlatBuffers (Google), MessagePack — niches mais pertinentes |
| **Compression moderne** | Brotli, Zstd vs gzip — bénéfices différents selon taille payload et CPU |
| **Observabilité unifiée** | OpenTelemetry traces + metrics + logs cross-protocole — clé pour comparer en prod |
| **Sécurité** | mTLS systématique en gRPC vs JWT/OAuth2 côté REST — coûts d'opération et perf TLS |
| **API Gateway / mesh** | Envoy, Linkerd, Istio — translation REST↔gRPC, impact perf/latence |
| **Streaming** | Server-sent events (REST) vs streaming gRPC bidi — cas d'usage capteurs temps réel |
| **Schémas et contrats** | OpenAPI 3.1, Protobuf evolution rules, Buf, gRPC reflection |
| **Edge computing / IoT** | Contraintes mémoire/CPU sur passerelles SignalWatch — gRPC trop lourd ? alternatives ? |

### Questions ouvertes pour le plan

- Quel coût supplémentaire WAN apporte chaque protocole (50 ms RTT, 1 % loss) ?
- À partir de quel volume de payload gzip/Brotli devient rentable côté REST ?
- gRPC streaming vs polling REST : seuil de bascule en débit messages/s ?
- Quel impact de mTLS systématique sur le RPS gRPC mesuré ici ?
- Migration progressive : REST → gRPC service par service, ou via gateway de traduction ?

---

## 11. Reproductibilité (pour information)

- Lancement complet : `make bench-all` — collecte sysinfo + payload + A/B/C + gzip.
- Scripts par scénario : `make bench-rest-a`, `make bench-grpc-c`, etc.
- Dépendances : Go 1.26+, k6, ghz, protoc + plugins go/grpc.
- Sortie : `benchmark/results/` — JSON k6/ghz, CSV monitor, system-info.

---

*Fin du briefing. Le LLM partenaire dispose ici de tout ce qu'il faut pour bâtir un plan de veille structuré sans accès au dossier `benchmark/`.*
