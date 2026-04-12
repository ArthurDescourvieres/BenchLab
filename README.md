# BenchLab

Comparaison **REST vs gRPC** pour le contexte SignalWatch (microservices Go, même stockage, benchmarks reproductibles).

Ce fichier documente **l’initialisation du projet** sur ta machine.

## Prérequis

- **Go** : version compatible avec celle indiquée dans `go.mod` (actuellement `go 1.26.2`). Vérifie avec :

  ```bash
  go version
  ```


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
- API capteurs : préfixe sous `/sensors` (code dans `rest-service/` : `main.go`, `handlers.go`, `router.go`, `service.go` ; stockage partagé dans `store/`).

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