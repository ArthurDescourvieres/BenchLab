# Plan de veille — BenchLab
## REST vs gRPC pour l'IoT industriel — SignalWatch
**Semaine du 05 au 11 mai 2026**

---

## 1. Contexte et objectif

SignalWatch collecte des données de capteurs industriels (température, pression, vibration) à raison de **10 000 événements par minute**. Le CTO doit choisir le protocole de communication inter-services le plus adapté. Ce plan de veille structure la recherche documentaire et réglementaire nécessaire pour étayer une recommandation basée sur des données mesurées.

**Livrable final :**
- Deux micro-services (REST + gRPC) en Go, benchmarkés sur 3 scénarios
- Rapport de veille 8-10 pages
- Support de présentation 10 slides max

---

## 2. Répartition thématique

| Membre | Thème principal | Thème secondaire |
|--------|----------------|-----------------|
| Arthur Descourvieres | gRPC / Protobuf / HTTP/2 · Implémentation grpc-service · Benchmark ghz | Éco-conception (RGESN, empreinte réseau) |
| Membre 2 | REST / HTTP/1.1 / OpenAPI · Implémentation rest-service · Benchmark k6 | RGPD appliqué à l'IoT industriel |
| Membre 3 | GraphQL · Message brokers (Kafka, RabbitMQ) · Protocoles IoT émergents | Assemblage rapport · Génération graphiques · Coordination rendu |

---

## 3. Sources par thème

### 3.1 — gRPC / Protobuf / HTTP/2 (Arthur)

| Source | Type | Référence |
|--------|------|-----------|
| Documentation officielle gRPC | Documentation officielle | grpc.io/docs |
| Protocol Buffers Language Guide v3 | Documentation officielle | protobuf.dev/programming-guides/proto3 |
| RFC 7540 — Hypertext Transfer Protocol Version 2 (HTTP/2) | Standard IETF | tools.ietf.org/html/rfc7540 |
| grpc-go — Performance benchmark notes | Source primaire (GitHub) | github.com/grpc/grpc-go |
| gRPC official benchmarks (Java, C++, Go) | Benchmark publié | grpc.io/docs/guides/benchmarking |
| ghz — gRPC load testing tool | Outil | ghz.sh |

**Critère de validation :** documentation officielle + dépôts maintenus activement, pas de blog anonyme.

### 3.2 — REST / HTTP/1.1 / OpenAPI (Membre 2)

| Source | Type | Référence |
|--------|------|-----------|
| Fielding, R. (2000). Architectural Styles and the Design of Network-based Software Architectures | Thèse fondatrice | ics.uci.edu/~fielding/pubs/dissertation |
| OpenAPI Specification 3.1.0 | Standard | spec.openapis.org/oas/v3.1.0 |
| RFC 9110 — HTTP Semantics | Standard IETF | tools.ietf.org/html/rfc9110 |
| Grafana k6 Documentation | Documentation outil | grafana.com/docs/k6 |
| Gin Web Framework — Documentation | Documentation framework | gin-gonic.com/docs |
| "gRPC vs REST: Understanding gRPC, OpenAPI and REST" — Google Cloud Blog | Blog de référence | cloud.google.com/blog |

### 3.3 — GraphQL (Membre 3)

| Source | Type | Référence |
|--------|------|-----------|
| GraphQL Specification (juin 2018) | Spécification officielle | spec.graphql.org |
| "GraphQL vs REST" — Apollo Blog | Blog technique | apollographql.com/blog |
| "Principled GraphQL" | Guide Apollo | principledgraphql.com |

### 3.4 — Message Brokers : Kafka, RabbitMQ (Membre 3)

| Source | Type | Référence |
|--------|------|-----------|
| Apache Kafka Documentation | Documentation officielle | kafka.apache.org/documentation |
| RabbitMQ Documentation | Documentation officielle | rabbitmq.com/docs |
| Kreps, J. et al. (2011). Kafka: A distributed messaging system for log processing | Article d'ingénierie (LinkedIn) | — |
| "Kafka vs RabbitMQ — Understanding the Differences" — Confluent Blog | Blog technique | confluent.io/blog |

### 3.5 — Éco-conception (Arthur)

| Source | Type | Référence |
|--------|------|-----------|
| RGESN — Référentiel Général d'Écoconception des Services Numériques v2 (2024) | Référentiel officiel | ecoresponsable.numerique.gouv.fr |
| Green IT — iNum Study (2023) | Rapport observatoire | greenit.fr |
| ADEME — Rapport « Empreinte environnementale du numérique mondial » (2022) | Rapport | ademe.fr |
| Green Software Foundation — Software Carbon Intensity Specification | Spécification | greensoftware.foundation |

### 3.6 — RGPD / Réglementation IoT (Membre 2)

| Source | Type | Référence |
|--------|------|-----------|
| Règlement (UE) 2016/679 — RGPD | Texte réglementaire | eur-lex.europa.eu |
| CNIL — Fiche pratique IoT (2022) | Guide officiel | cnil.fr |
| CNIL — Référentiel sur la durée de conservation des données | Guide officiel | cnil.fr |
| ENISA — Guidelines for Securing the IoT (2020) | Guide sécurité | enisa.europa.eu |
| Directive NIS2 (UE) 2022/2555 | Texte réglementaire | eur-lex.europa.eu |

---

## 4. Planning de la semaine (jour par jour)

| Jour | Arthur (gRPC + éco) | Membre 2 (REST + RGPD) | Membre 3 (Brokers + rapport) |
|------|---------------------|------------------------|------------------------------|
| **Lun 05/05** | Lecture docs gRPC, Protobuf, RFC HTTP/2. Prise de notes. | Lecture Fielding, OpenAPI, RFC HTTP. Prise de notes. | Lecture Kafka/RabbitMQ, GraphQL. Mise en place repo GitHub. |
| **Mar 06/05** | Implémentation `grpc-service/` (Go + grpc-go). Définition `sensor.proto`. | Implémentation `rest-service/` (Gin). Routes CRUD. | Implémentation `internal/sensorsvc/` partagé, `store/memory.go`, Makefile. |
| **Mer 07/05** | Scripts benchmark ghz (scénarios A/B/C). Script `monitor.exe`. | Scripts benchmark k6 (scénarios A/B/C + gzip). | Script `payloadsize`, `collect-system-info`. Vérification comparabilité des deux services. |
| **Jeu 08/05** | Exécution `make bench-grpc-a/b/c`. Collecte JSON/CSV. Analyse préliminaire gRPC. | Exécution `make bench-rest-a/b/c`. Collecte JSON/CSV. Analyse préliminaire REST. | Agrégation résultats, tracé graphiques (latence, RPS, charge). |
| **Ven 09/05** | Rédaction section B (gRPC/HTTP2) + section D (éco-conception, chiffres mesurés). | Rédaction section B (REST/GraphQL/Brokers) + section E (RGPD IoT). | Rédaction section C (résultats benchmark, tableaux, comparaison littérature). |
| **Sam 10/05** | Rédaction section F (recommandation, matrice). Relecture sections D+F. | Finalisation section E. Rédaction section A (intro + méthodologie). | Assemblage complet du rapport, mise en forme. Révision globale. |
| **Dim 11/05** | Slides 4-6 (résultats benchmark, graphiques). | Slides 1-3 (contexte, protocoles) + slide 8 (RGPD). | Slides 7 (éco) + slides 9-10 (reco, conclusion). Rendu GitHub. |

**Point quotidien :** 09h00, 15 min maximum, en présentiel ou appel.

---

## 5. Format de restitution interne

| Élément | Détail |
|---------|--------|
| **Point quotidien** | 09h00 — blocages, avancement, arbitrages |
| **Document partagé** | HackMD ou Google Doc — rédaction collaborative du rapport |
| **Gestion de version** | Git (ce dépôt) — une branche par livrable, PR quotidienne |
| **Critères de validation source** | Date < 5 ans (sauf fondateurs), auteur identifiable, relu par un pair |
| **Fichier de suivi** | Tableau sources validées mis à jour en continu dans ce document |

---

## 6. RACI simplifié

| Livrable | Arthur | Membre 2 | Membre 3 |
|----------|:------:|:--------:|:--------:|
| `rest-service/` (code) | C | **R/A** | I |
| `grpc-service/` + `.proto` (code) | **R/A** | C | I |
| `internal/sensorsvc/` + `store/` | **R** | **R** | **A** |
| Scripts benchmark k6 | I | **R/A** | C |
| Scripts benchmark ghz | **R/A** | I | C |
| Scripts monitor + payloadsize | **R/A** | C | C |
| Makefile orchestration | C | C | **R/A** |
| `benchmark/results/` (données brutes) | **R** | **R** | **A** |
| Rapport — Section A (Introduction + méthodologie) | C | C | **R/A** |
| Rapport — Section B (REST) | I | **R** | **A** |
| Rapport — Section B (gRPC) | **R** | I | **A** |
| Rapport — Section B (GraphQL + Brokers) | I | **R** | **A** |
| Rapport — Section C (Résultats benchmark) | C | C | **R/A** |
| Rapport — Section D (Éco-conception) | **R/A** | C | I |
| Rapport — Section E (RGPD) | I | **R/A** | C |
| Rapport — Section F (Recommandation) | C | C | **R/A** |
| Support de présentation | **R** (slides 4-6) | **R** (slides 1-3, 8) | **R/A** (slides 7, 9-10) |
| Rendu GitHub | **R/A** | C | C |

**Légende :** **R** = Responsible · **A** = Accountable · **C** = Consulted · **I** = Informed

---

*Document vivant — mis à jour au fil de la semaine.*
