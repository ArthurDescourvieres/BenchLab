# Conditions de test — BenchLab

Ce document recense les conditions matérielles et logicielles dans lesquelles les benchmarks ont été exécutés. Il est requis par la consigne du projet (reproductibilité). Les valeurs ci-dessous doivent refléter la machine ayant produit les fichiers présents dans `benchmark/results/`.

> Les champs marqués `<auto>` sont remplis automatiquement par le script `benchmark/scripts/collect-system-info.sh` (ou `.ps1`). Lance-le après chaque changement de machine ou de version d'outil.

## Machine

| Champ | Valeur |
| :--- | :--- |
| Hostname | `<auto>` |
| OS | `<auto>` |
| Architecture | `<auto>` |
| CPU (modèle) | `<auto>` |
| CPU (cœurs logiques) | `<auto>` |
| RAM totale | `<auto>` |

## Versions des outils

| Outil | Version |
| :--- | :--- |
| Go | `<auto>` |
| k6 | `<auto>` |
| ghz | `<auto>` |
| protoc | `<auto>` |

## Versions des dépendances Go (extraites de `go.mod`)

| Dépendance | Version |
| :--- | :--- |
| `github.com/gin-gonic/gin` | v1.12.0 |
| `google.golang.org/grpc` | v1.71.1 |
| `google.golang.org/protobuf` | v1.36.10 |
| `github.com/gin-contrib/gzip` | (ajoutée pour la variante gzip — voir `go.mod`) |
| `github.com/shirou/gopsutil/v3` | (ajoutée pour le monitoring CPU/RAM — voir `go.mod`) |

## Conditions réseau

- Tous les benchmarks tournent en **loopback local** (`localhost`) — pas de latence réseau réelle.
- Aucun proxy / VPN actif pendant la mesure.
- Service et outil de mesure sur la **même machine** (deux processus séparés).

## État système recommandé pendant les mesures

- Aucune autre charge CPU significative (fermer IDE, navigateurs lourds, conteneurs autres que ceux du test).
- Mode énergie « performances maximales » sur Windows / désactiver le throttling thermique.
- Service à mesurer démarré avec `gin.ReleaseMode` (par défaut dans `rest-service/main.go`).
- Ports utilisés : `:8080` pour REST, `:9090` pour gRPC (modifiables via `PORT` / `GRPC_PORT`).

## Comment regénérer ce document automatiquement

Lance le collecteur depuis la racine du dépôt :

```bash
# Bash (Linux / macOS / Git Bash sous Windows)
bash benchmark/scripts/collect-system-info.sh
```

```powershell
# PowerShell (Windows)
.\benchmark\scripts\collect-system-info.ps1
```

Sortie attendue :
- `benchmark/results/system-info.json` — version structurée, exploitable pour le rapport.
- `benchmark/results/system-info.txt` — version lisible humaine.

## Commandes de reproduction

Voir le `README.md` (sections Benchmarks REST et Benchmarks gRPC) pour les commandes exactes par scénario, ou utilise la cible globale du `Makefile` :

```bash
make bench-all
```
