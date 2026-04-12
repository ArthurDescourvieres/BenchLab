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

**macOS / Linux** :

```bash
go build -o bin/rest ./cmd/rest
```

**Windows (PowerShell ou cmd, depuis la racine du repo)** :

```powershell
go build -o bin\rest.exe .\cmd\rest
```

Tu obtiens un exécutable `bin/rest` (Unix) ou `bin\rest.exe` (Windows).

## Lancer le service REST

Par défaut le serveur écoute sur **`:8080`** (aucun fichier `.env` n’est lu automatiquement par l’application — voir la section suivante).

**macOS / Linux** :

```bash
./bin/rest
```

**Windows (PowerShell)** :

```powershell
.\bin\rest.exe
```

**Windows (cmd)** :

```cmd
bin\rest.exe
```

### Port personnalisé (`PORT`)

**macOS / Linux (bash ou zsh)** :

```bash
export PORT=3000
./bin/rest
```

**Windows (PowerShell)** :

```powershell
$env:PORT = "3000"
.\bin\rest.exe
```

**Windows (cmd)** :

```cmd
set PORT=3000
bin\rest.exe
```

### Vérifier que ça tourne

- Santé : `GET http://localhost:8080/health` → réponse **200**.
- API capteurs : préfixe sous `/sensors` (détails dans `internal/rest/`).

Exemple rapide en ligne de commande :

**macOS / Linux** :

```bash
curl -i http://localhost:8080/health
```

**Windows (PowerShell)** :

```powershell
curl.exe -i http://localhost:8080/health
```

## Développement sans binaire intermédiaire

**Toutes plateformes** :

```bash
go run ./cmd/rest
```