# BenchLab — Cheatsheet de présentation

> Objectif : pouvoir répondre aux questions sans paniquer. Tout est expliqué simplement, du code aux benchmarks.

---

## 1. Le projet en une phrase

> "BenchLab compare deux façons de faire dialoguer des microservices — **REST** (texte JSON sur HTTP/1.1) et **gRPC** (binaire Protobuf sur HTTP/2) — sur la **même logique métier** (gestion de capteurs industriels), pour mesurer les vraies différences de perf."

**Pourquoi c'est intelligent comme protocole de test :** la logique métier est partagée (le package `internal/sensorsvc`). Donc toute différence de perf vient du **protocole**, pas de l'algo. C'est le principe d'un benchmark propre : ne faire varier qu'une seule chose.

---

## 2. Architecture du code (à savoir par cœur)

```
BenchLab/
├── store/                 ← couche données (en mémoire, pas de DB)
│   ├── store.go           ← interface Store + struct Sensor
│   └── memory.go          ← implémentation MemoryStore (map + RWMutex)
│
├── internal/sensorsvc/    ← LOGIQUE MÉTIER PARTAGÉE
│   └── service.go         ← validations (type, status, RFC3339), Create/Get/Update...
│
├── rest-service/          ← serveur REST (port 8080)
│   ├── main.go            ← démarrage Gin
│   ├── router.go          ← routes /sensors
│   └── handlers.go        ← parse JSON → appelle sensorsvc → renvoie JSON
│
├── grpc-service/          ← serveur gRPC (port 9090)
│   ├── proto/sensor.proto ← contrat (messages + RPC)
│   ├── gen/...            ← code Go généré depuis le .proto (par protoc)
│   ├── main.go            ← démarrage du serveur gRPC
│   └── server.go          ← implémente les RPC → appelle sensorsvc
│
└── benchmark/
    ├── scripts/           ← scripts k6 (.js) + ghz (.sh/.ps1)
    ├── cmd/
    │   ├── monitor/       ← outil maison qui mesure CPU/RAM d'un PID (1 Hz)
    │   ├── payloadsize/   ← compare la taille JSON vs Protobuf
    │   └── seedsensor/    ← crée un capteur de test avant un bench gRPC
    └── results/           ← sorties JSON/CSV
```

**Le point clé à retenir** : les deux serveurs (REST + gRPC) appellent le **même `sensorsvc.SensorService`**. Le diagramme mental est :

```
Client REST ─→ Gin handler ─→ sensorsvc ─→ MemoryStore (map)
Client gRPC ─→ gRPC server ─→ sensorsvc ─→ MemoryStore (map)
                              └── identique des deux côtés ──┘
```

---

## 3. Comment marche le code REST (rest-service/)

**Stack** : [Gin](https://gin-gonic.com/) v1.12 — framework HTTP Go, très rapide.

### Démarrage (`main.go`)
1. Crée un routeur Gin en mode `Release` (pas de logs verbeux).
2. Si `REST_GZIP=1`, ajoute le middleware gzip (compression des réponses).
3. Crée un `MemoryStore` (map vide) → un `SensorService`.
4. Enregistre les routes via `RegisterRoutes`.
5. Écoute sur `:8080`.

### Routes (`router.go`)
```
POST   /sensors      → Create
GET    /sensors      → List
GET    /sensors/:id  → Get
PUT    /sensors/:id  → Update
DELETE /sensors/:id  → Delete
GET    /health       → 200 OK (pour vérifier que le service tourne)
```

### Handler typique (`handlers.go`, méthode `Create`)
1. Désérialise le JSON du body en struct Go (`sensorInput`).
2. Construit un `SensorPayload` (DTO du package métier).
3. Appelle `svc.Create(ctx, payload)` → la logique métier valide et stocke.
4. Renvoie le `Sensor` créé en JSON avec `201 Created`.
5. En cas d'erreur, mappe vers le bon code HTTP (`writeSensorError`) :
   - `ErrNotFound` → 404
   - erreurs de validation → 400
   - sinon → 500

---

## 4. Comment marche le code gRPC (grpc-service/)

**Stack** : grpc-go v1.71 + protobuf v1.36.

### Le fichier `.proto` (contrat)
C'est le **contrat partagé** entre client et serveur. On y définit :
- des **messages** (équivalent struct) avec des champs numérotés (les "tags")
- un **service** avec ses RPC

```protobuf
message Sensor {
  string id = 1;
  string name = 2;
  ...
}
service SensorService {
  rpc CreateSensor(CreateSensorRequest) returns (Sensor);
  rpc GetSensor(SensorId) returns (Sensor);
  ...
}
```

> **Les numéros (`= 1`, `= 2`)** servent à identifier les champs dans le format binaire. Ils ne doivent jamais changer une fois en prod.

### La génération de code
La commande `protoc` (compilateur Protobuf) lit `sensor.proto` et génère du code Go dans `grpc-service/gen/` :
- structs Go pour chaque message
- une **interface serveur** (`SensorServiceServer`) avec une méthode par RPC
- un **client** (utilisé par ghz pour taper le service)

### Démarrage (`main.go`)
1. Crée un listener TCP sur `:9090`.
2. Crée un `grpc.Server` avec keepalive (ping HTTP/2 toutes les 20 s).
3. Crée le store + service métier (comme côté REST).
4. **Enregistre l'implémentation** : `sensorv1.RegisterSensorServiceServer(s, server)`.
5. Active la **reflection gRPC** (permet à des outils comme `grpcurl` ou `ghz` de découvrir le service).

### Implémentation (`server.go`)
Pour chaque RPC, une méthode :
1. Vérifie que la requête n'est pas nulle.
2. Convertit le message proto → `SensorPayload` (DTO métier).
3. Appelle `svc.Create/Get/...` (le **même** service que REST).
4. Convertit le résultat → message proto.
5. Mappe les erreurs métier vers les **codes gRPC standard** : `NotFound`, `InvalidArgument`, `Internal`.

---

## 5. La logique métier partagée (`sensorsvc/service.go`)

C'est le **cœur** du projet — à comprendre absolument car c'est ce qui rend le bench valide.

### Validations
- `name` non vide
- `type` ∈ {`TEMPERATURE`, `PRESSURE`, `VIBRATION`}
- `status` ∈ {`ACTIVE`, `INACTIVE`, `MAINTENANCE`}
- `last_reading_at` est un timestamp **RFC3339** valide

### Stockage
Le `MemoryStore` dans `store/memory.go` :
- une `map[string]Sensor` protégée par un `sync.RWMutex` (lecture/écriture concurrente sûre)
- ID généré aléatoirement : 16 octets `crypto/rand` → encodés en hex (32 caractères)
- `CreatedAt` ajouté automatiquement au format RFC3339

> **Pourquoi pas de DB ?** Pour **isoler le coût du protocole**. Si on ajoutait PostgreSQL, on mesurerait surtout la latence DB (plusieurs ms par requête), et la différence REST/gRPC deviendrait invisible.

---

## 6. Comment marchent les benchmarks

### Les 3 scénarios

| Scénario | Charge | But |
|---|---|---|
| **A — Read** | 1 000 requêtes, 10 connexions | Mesurer le coût d'une lecture unitaire à charge faible |
| **B — Write** | 500 requêtes, 5 connexions | Idem en écriture (Create) |
| **C — Ramp** | montée 10 → 100 connexions | Voir comment ça scale |

### Les outils

#### **k6** (utilisé pour REST)
Outil JS de load testing. Tu écris un script qui définit :
- `options.vus` (Virtual Users = utilisateurs simulés)
- `options.iterations` ou `options.stages` (nombre fixe ou montée progressive)
- une fonction `default` qui fait la requête

Exemple (scénario A) : 10 VU partagent 1000 itérations → chacun en fait 100. k6 mesure pour chaque requête le temps de réponse, puis calcule `avg`, `p95`, `p99`, `max`, `RPS`.

#### **ghz** (utilisé pour gRPC)
Équivalent de k6 mais spécialisé gRPC. Il prend :
- le `.proto` (pour savoir comment encoder)
- l'adresse du serveur
- le RPC à appeler (`benchlab.sensor.v1.SensorService/GetSensor`)
- un payload JSON (qu'il convertit en Protobuf)
- `-n` (nombre total de requêtes), `-c` (concurrence), `--connections` (nb de connexions TCP)

#### **monitor.exe** (outil maison)
Programme Go que tu as écrit. Il prend un PID, attache `gopsutil` dessus, échantillonne **toutes les secondes** :
- `cpu_percent` (% CPU, où 100 % = 1 cœur logique)
- `mem_rss_mb` (RAM résidente)
- `num_threads`

Sortie : un fichier CSV. Tu l'utilises pour voir combien de ressources le serveur consomme **pendant** que k6/ghz tape dessus.

#### **payloadsize** (outil maison)
Petit programme Go qui prend un `Sensor` et le sérialise en JSON puis en Protobuf, et écrit les deux tailles dans `benchmark/results/payload-size.json`. C'est ce qui te donne le **ratio 1,70× (230 vs 135 octets)**.

---

## 7. Les chiffres à connaître par cœur

### Taille de payload
- JSON : **230 octets** | Protobuf : **135 octets** | **Ratio ~1,70×**
- Sur le fil gRPC : +5 octets de frame HTTP/2 par appel unary → gain net ~40 %.

### Scénario A (Read, faible charge)
| | RPS | avg | p95 | p99 |
|---|---|---|---|---|
| **REST** | **14 324** | 0,41 ms | 1,26 ms | 2,24 ms |
| gRPC | 5 241 | 1,40 ms | 3,25 ms | 7,33 ms |
| REST+gzip | 6 074 | 1,19 ms | 3,07 ms | 6,59 ms |

→ **REST gagne** à charge faible en loopback.

### Scénario B (Write, faible charge)
| | RPS | avg |
|---|---|---|
| **REST** | **6 506** | 0,44 ms |
| gRPC | 4 087 | 0,81 ms |

→ **REST gagne aussi**, écart plus serré.

### Scénario C (montée en charge)
- REST (k6, ramp continu 10→100 VU) : **37 385 RPS moyen**, p99 = 8 ms.
- gRPC (ghz, 5 paliers) : monte à **17 032 RPS @ 100 conn**, mais **p99 explose à 34 ms**.

### Ressources
- REST C : CPU avg **325 %** (≈ 3 cœurs), RAM ~28 Mo.
- gRPC C : CPU avg **27 %**, RAM ~17 Mo.
- → **gRPC est beaucoup plus économe en CPU/RAM**.

---

## 8. Conclusions à défendre

1. **REST est plus rapide ici, mais c'est trompeur** — c'est en loopback, sans latence réseau. Le coût d'établissement de gRPC (HTTP/2 framing, codec) ne s'amortit pas sur de tout petits appels rapides.
2. **gRPC est plus efficient côté ressources** (CPU, RAM, taille des messages).
3. **gzip est contre-productif sur petits payloads en loopback** : la compression coûte plus de CPU qu'elle n'économise de bande passante (qui est infinie en local).
4. **gRPC scale plus régulièrement en RPS** mais sa **p99 se dégrade vite** au-delà de 50 connexions concurrentes (queueing derrière le multiplexing HTTP/2).
5. **Limite majeure** : aucun chiffre ne reflète une vraie latence WAN. Sur un vrai réseau (50 ms RTT, paquets perdus), gRPC reprendrait l'avantage grâce à ses messages plus petits et à la persistance des connexions.

---

## 9. Questions piège — réponses préparées

### "Pourquoi REST gagne, alors que tout le monde dit que gRPC est plus rapide ?"
> "Parce qu'on est en loopback (`localhost`). À cette échelle, le réseau est gratuit, donc le seul coût mesuré est CPU/sérialisation. Le **framing HTTP/2** et le coût de la machine d'état gRPC ne se rentabilisent que quand on a un vrai RTT réseau ou de grosses payloads. C'est documenté comme limite dans le briefing."

### "Pourquoi pas de base de données ?"
> "Pour isoler le coût du protocole. Avec une DB, la latence I/O dominerait (plusieurs ms par requête) et la différence REST/gRPC serait masquée. Le `MemoryStore` est une `map` Go protégée par un RWMutex."

### "Pourquoi pas la même méthodologie de ramp pour les deux ?"
> "Limitation de **ghz** : il ne supporte pas un ramp continu, contrairement à k6 qui a des `stages`. On a émulé la montée par 5 runs successifs à concurrence fixe (10, 25, 50, 75, 100). C'est un compromis explicite, et la comparaison directe est imparfaite — c'est un ordre de grandeur."

### "Comment vous garantissez que vous comparez bien la même chose ?"
> "La logique métier est dans `internal/sensorsvc/service.go` et est appelée à l'identique par les deux serveurs. Validations, store, format des timestamps — tout est partagé. Seules les couches de transport diffèrent."

### "Comment vous mesurez le CPU ?"
> "Un outil maison (`benchmark/cmd/monitor/main.go`) attache un PID via la lib `gopsutil/v3` et échantillonne CPU%, RSS et nombre de threads à 1 Hz dans un CSV. Sur Windows, 100 % = 1 cœur logique sur 22."

### "Pourquoi 230 vs 135 octets ?"
> "JSON encode les noms de champs en clair (`"name":"Bench-Setup"`). Protobuf encode juste un tag numérique + la valeur en binaire. Pas de quotes, pas de virgules, pas de noms répétés. Sur ce `Sensor` on gagne ~40 %."

### "C'est quoi un VU dans k6 ?"
> "Virtual User — un utilisateur simulé. Concrètement, une **goroutine** (k6 est écrit en Go) qui boucle sur la fonction `default()`. 10 VU = 10 boucles parallèles."

### "C'est quoi un `c` / `--connections` dans ghz ?"
> "`-c` = concurrence (nb de workers parallèles). `--connections` = nb de connexions TCP/HTTP/2 ouvertes. On les met égaux pour qu'il n'y ait pas plusieurs workers qui partagent la même connexion (ce qui multiplexerait via HTTP/2 et fausserait la mesure)."

### "Pourquoi gzip dégrade les perfs ?"
> "Compresser coûte du CPU. En loopback, on n'économise rien en transport (la bande passante est virtuellement infinie). Donc le coût > le gain. En WAN sur de gros payloads, le verdict s'inverserait."

### "Pourquoi p99 explose en gRPC à 100 connexions mais pas en RPS ?"
> "Le RPS moyen monte parce que la machine encaisse globalement. Mais HTTP/2 multiplexe les requêtes sur quelques connexions ; quand la concurrence monte, des requêtes attendent dans la file → la **queue de tail** s'allonge → p99 grimpe. C'est typique des protocoles multiplexés sous charge."

### "Pourquoi RFC3339 pour les dates ?"
> "Format ISO 8601 standard pour les timestamps API (`2026-01-15T10:00:00Z`). Lisible humain, parseable trivialement, fuseau horaire explicite."

### "Reflection gRPC, c'est quoi ?"
> "Une API qui permet à un client de découvrir les services exposés sans avoir le `.proto` au préalable. Utile pour `ghz`, `grpcurl`, Postman… On l'a activée pour le confort de test."

### "Keepalive gRPC, à quoi ça sert ?"
> "Les serveurs envoient des pings HTTP/2 réguliers (ici toutes les 20 s) pour détecter les connexions mortes et garder les NAT/firewalls ouverts. Sans ça, des connexions inactives peuvent être coupées silencieusement."

### "Le RWMutex sur la map, ça change quoi en perf ?"
> "C'est un mutex à lecture/écriture : plusieurs lectures peuvent se faire en parallèle, mais une écriture bloque tout. Pour notre charge dominée par `Get`, c'est plus efficace qu'un mutex simple. Et c'est nécessaire car les goroutines Go peuvent appeler la map en parallèle (sinon : data race)."

---

## 10. À NE PAS dire

- ❌ "gRPC est plus rapide" — **faux dans nos chiffres**, à nuancer.
- ❌ "Le bench reflète la prod" — **non**, c'est du loopback synthétique.
- ❌ "On a benchmarké le réseau" — **non**, on a benchmarké CPU + sérialisation.
- ❌ "REST est mieux que gRPC" — **trop fort**, ça dépend du contexte (taille payload, RTT, throughput cible, écosystème).

---

## 11. Si on te demande comment relancer

```bash
make bench-all          # tout lancer
make bench-rest-a       # juste un scénario
```

Sortie dans `benchmark/results/` (JSON pour k6/ghz, CSV pour monitor).

---

**Mantra de la présentation** : *"On a comparé deux protocoles sur la même logique métier, en loopback. REST gagne en latence brute parce que gRPC ne s'amortit pas à cette échelle ; gRPC gagne en taille de message et en CPU. Les conclusions sont valides pour ce contexte précis et appellent une étude WAN pour conclure en prod."*
