# BenchLab

Comparaison **REST vs gRPC** pour le contexte SignalWatch (microservices Go, même stockage, benchmarks reproductibles).

Ce fichier documente **l’initialisation du projet** sur ta machine.

## Prérequis

- **Go** : version compatible avec celle indiquée dans `go.mod` (actuellement `go 1.26.2`). Vérifie avec :

  ```bash
  go version
  ```

- **Protobuf** (uniquement si tu régénères le client/serveur Go depuis le `.proto`) : `protoc` + plugins `protoc-gen-go` et `protoc-gen-go-grpc` (installés automatiquement par `scripts/generate-proto.sh` via `go install`).

- **Benchmarks** : [k6](https://k6.io/) pour le REST ; [ghz](https://ghz.sh/) pour le gRPC (lignes de commande documentées ci‑dessous).

> 📋 Conditions de test (machine, versions exactes des outils) : voir [`benchmark/BENCH-CONDITIONS.md`](benchmark/BENCH-CONDITIONS.md). Les valeurs sont régénérables automatiquement via `bash benchmark/scripts/collect-system-info.sh` (ou `.\benchmark\scripts\collect-system-info.ps1` sous PowerShell).


## Installer les dépendances Go

Les modules sont décrits dans `go.mod` ; les empreintes de sécurité dans `go.sum`.

**Toutes plateformes** (à lancer depuis la racine `BenchLab/`) :

```bash
go mod download
```

En cas d’ajout de dépendances plus tard : `go get chemin/du/module@version` puis commit de `go.mod` et `go.sum`.

## Compiler le service REST

**Bash** (macOS, Linux, **Git Bash / MSYS2 / terminal bash sous Windows** — depuis la racine du repo) :

```bash
go build -o "bin/rest$(go env GOEXE)" ./rest-service
```

`GOEXE` vaut `.exe` sous Windows et une chaîne vide ailleurs, donc le binaire est toujours au bon nom (`bin/rest` ou `bin/rest.exe`).

## Lancer le service REST

Par défaut le serveur écoute sur **`:8080`** (aucun fichier `.env` n’est lu automatiquement par l’application — voir la section suivante).

**Bash** (y compris Git Bash / terminal bash sous Windows) :

```bash
./bin/rest$(go env GOEXE)
```

### Port personnalisé (`PORT`)

**Bash** (macOS, Linux, Git Bash / bash sous Windows) :

```bash
export PORT=3000
./bin/rest$(go env GOEXE)
```

### Vérifier que ça tourne

- Santé : `GET http://localhost:8080/health` → réponse **200**.
- API capteurs : préfixe sous `/sensors` (code dans `rest-service/` : `main.go`, `handlers.go`, `router.go` ; logique métier partagée dans `internal/sensorsvc/` ; stockage dans `store/`).

Exemple rapide en ligne de commande :

**Bash** (avec `curl` installé — inclus dans Git Bash) :

```bash
curl -i http://localhost:8080/health
```

## Développement sans binaire intermédiaire

**Toutes plateformes** :

```bash
go run ./rest-service
```

## Service gRPC

Contrat : [`grpc-service/proto/sensor.proto`](grpc-service/proto/sensor.proto). Le code Go généré est versionné sous [`grpc-service/gen/benchlab/sensor/v1/`](grpc-service/gen/benchlab/sensor/v1/).

Pour **régénérer** après modification du proto (depuis la racine du dépôt, avec Bash) :

```bash
bash scripts/generate-proto.sh
```

### Compiler le service gRPC

```bash
go build -o "bin/grpc$(go env GOEXE)" ./grpc-service
```

### Lancer le service gRPC

Par défaut le serveur écoute sur **`:9090`**.

```bash
./bin/grpc$(go env GOEXE)
```

### Port personnalisé (`GRPC_PORT`)

```bash
export GRPC_PORT=50051
./bin/grpc$(go env GOEXE)
```

## Benchmarks (REST vs gRPC)

Chaque processus a **son propre stockage mémoire** : les scripts créent un capteur de référence (setup) avant la charge, pour des mesures comparables.

Depuis la racine `BenchLab/`, lance **deux terminaux** : un pour le service à mesurer, un pour l’outil de benchmark.

### REST (k6)

Terminal service :

```bash
go run ./rest-service
```

Terminal mesures (scénario **A** — 1000 lectures, 10 VU) :

```bash
k6 run benchmark/scripts/k6-rest-a-read.js
```

Scénario **B** — 500 écritures, 5 VU :

```bash
k6 run benchmark/scripts/k6-rest-b-write.js
```

Scénario **C** — montée progressive 10 → 100 VU sur `GET` :

```bash
k6 run benchmark/scripts/k6-rest-c-ramp.js
```

Les résumés k6 sont écrits dans `benchmark/results/` (`k6-rest-*-summary.json`). Ils incluent **`med` (p50)**, **`p(95)`** et **`p(99)`** sur les tendances de temps, une métrique **`response_size_bytes`** (taille du corps de réponse), et une **`Rate` `errors`** explicite (% de requêtes non‑2xx, complète `http_req_failed` collecté nativement par k6).

#### Variante avec compression gzip (REST + gzip)

La consigne « pour aller plus loin » demande de comparer JSON+gzip à Protobuf. Le service REST embarque un middleware gzip **désactivé par défaut**, à activer via la variable d'environnement `REST_GZIP=1`.

```bash
# Bash : activer gzip puis lancer la variante k6 dédiée
REST_GZIP=1 ./bin/rest$(go env GOEXE)
# Dans un autre terminal :
k6 run benchmark/scripts/k6-rest-a-read-gzip.js
```

```powershell
# PowerShell
$env:REST_GZIP=1; .\bin\rest.exe
# Dans un autre terminal :
k6 run benchmark\scripts\k6-rest-a-read-gzip.js
```

Le résumé est écrit dans `benchmark/results/k6-rest-a-gzip-summary.json` ; comparer la `response_size_bytes.avg` et la `data_received` à la version sans gzip.

**Taille JSON vs Protobuf** (un même capteur, hors en-têtes réseau) :

```bash
go run ./benchmark/cmd/payloadsize
```

Le fichier `benchmark/results/payload-size.json` est régénéré à chaque exécution.

### gRPC (ghz)

Terminal service :

```bash
go run ./grpc-service
```

Variables optionnelles : `GRPC_ADDR` (défaut `localhost:9090`).

Terminal mesures — **le serveur gRPC doit déjà tourner**, sinon tu verras une erreur du type *connection refused*.

**Git Bash / WSL / macOS / Linux** (avec `ghz` dans le `PATH`) :

```bash
bash benchmark/scripts/grpc-a-read.sh
bash benchmark/scripts/grpc-b-write.sh
bash benchmark/scripts/grpc-c-ramp.sh
```

**Windows (PowerShell)** — même dossier de travail : la racine du dépôt `BenchLab/` :

```powershell
Set-ExecutionPolicy -Scope CurrentUser RemoteSigned
cd <chemin-vers-votre-dossier-BenchLab>
.\benchmark\scripts\grpc-a-read.ps1
.\benchmark\scripts\grpc-b-write.ps1
.\benchmark\scripts\grpc-c-ramp.ps1
```

Les sorties ghz sont des fichiers JSON dans `benchmark/results/` (`ghz-grpc-*.json`). Pour le scénario C, ghz ne supporte pas nativement le ramp continu (cf. k6 stages) : on émule la montée en charge par 5 runs indépendants à concurrence fixe (10, 25, 50, 75, 100), un fichier de résultats par palier. Les bornes 10 et 100 restent identiques au scénario k6 REST C, ce qui permet de comparer les extrêmes.

#### Si les scripts gRPC échouent (Windows)

1. **Deux terminaux** : dans le premier, laisse tourner `go run ./grpc-service` ; dans le second, lance le script (sans fermer le premier).
2. **`ghz` introuvable** : installe-le et ajoute `%USERPROFILE%\go\bin` au PATH si tu utilises `go install`, puis rouvre le terminal.
3. **`connection refused` / `Unavailable`** : le port **9090** n’a pas de serveur → démarre le service gRPC ou définis `GRPC_ADDR` / `GRPC_PORT` de la même façon pour le client et le serveur.
4. **PowerShell bloque les `.ps1`** : exécute une fois `Set-ExecutionPolicy -Scope CurrentUser RemoteSigned` (voir ci‑dessus).
5. **Git Bash et erreur `$'\r': command not found`** : les scripts `.sh` doivent être en fin de ligne LF ; le fichier [`.gitattributes`](.gitattributes) du dépôt force ce réglage pour `benchmark/scripts/*.sh`.

### Préparer un ID gRPC sans ghz (optionnel)

```bash
go run ./benchmark/cmd/seedsensor localhost:9090
```

La commande affiche uniquement l’identifiant du capteur créé sur le serveur ciblé.

## Orchestration en une commande

Le script `benchmark/scripts/bench-all.sh` enchaîne **tous** les scénarios (REST A/B/C, gRPC A/B/C, variante gzip), démarre chaque service en arrière-plan, attache un monitor CPU/RAM, lance l'outil de mesure, puis nettoie. Un Makefile équivalent est fourni pour les utilisateurs ayant `make`.

```bash
# Tout en un (sysinfo + payload + 7 scénarios)
bash benchmark/scripts/bench-all.sh

# Un scénario seul
bash benchmark/scripts/bench-all.sh rest-a
bash benchmark/scripts/bench-all.sh grpc-c
bash benchmark/scripts/bench-all.sh rest-gzip

# Ou via Make (équivalent)
make bench-all
make bench-rest-a
make bench-grpc-c
make bench-rest-gzip
make run-dashboard
```

## Monitoring CPU/RAM pendant un benchmark (éco-conception)

Un petit programme Go (`benchmark/cmd/monitor`) sonde un PID toutes les secondes (CPU%, RAM RSS, threads) et écrit un CSV. Il est lancé automatiquement par `bench-all.sh` / le Makefile pour chaque scénario, mais peut aussi être appelé seul :

```bash
# 1) Démarrer le service à mesurer dans un terminal :
go run ./rest-service

# 2) Dans un second terminal, repérer le PID puis :
go run ./benchmark/cmd/monitor \
  -pid <PID-du-service> \
  -duration 90s \
  -interval 1s \
  -out benchmark/results/monitor-rest-a.csv

# 3) Dans un troisième terminal, lancer la charge (k6 / ghz).
```

Sortie CSV : `timestamp_unix_ms,elapsed_s,cpu_percent,mem_rss_mb,num_threads`. À exploiter dans le rapport (section éco-conception).