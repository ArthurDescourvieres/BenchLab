# Support de présentation — BenchLab
## REST vs gRPC pour SignalWatch
**10 slides · mai 2026**

---

## Slide 1 — Contexte SignalWatch et problématique

**Titre :** SignalWatch — Choisir le bon protocole pour 10 000 capteurs/min

**Contenu :**

> *Une startup IoT en levée de fonds. Une architecture microservices à concevoir. Un CTO qui veut des chiffres, pas des opinions.*

**SignalWatch en chiffres :**
- Supervision de capteurs industriels : température, pression, vibration
- **10 000 événements par minute** → ~167 messages/seconde
- Architecture cible : microservices distribués
- Contraintes : latence < 5 ms p99, zéro perte de données, éco-conception

**La question :** REST (JSON/HTTP/1.1) ou gRPC (Protobuf/HTTP/2) ?

**Notre approche :** 2 services identiques, 3 scénarios mesurés, 1 recommandation chiffrée.

---

## Slide 2 — Panorama des protocoles (1/2) : REST et gRPC

**Titre :** REST vs gRPC — Les deux candidats principaux

**REST / HTTP**
```
Client → GET /sensors/42 → Serveur
         ← { "id":"42", "name":"Turb-A3", "last_value":87.3 } ←
```
✅ Universel (navigateurs, curl, Postman)
✅ JSON lisible et débogable
✅ Cache HTTP natif
❌ Verbeux (+70 % vs binaire)
❌ Pas de streaming natif
❌ HOL blocking HTTP/1.1

**gRPC / HTTP/2**
```
Client → GetSensor(id=42) [Protobuf 15B] → Serveur
         ← Sensor { id, name, last_value } [135B] ←
```
✅ Binaire compact (−41 % vs JSON)
✅ Streaming bidirectionnel natif
✅ Contrat fort (.proto = source de vérité)
✅ HTTP/2 : multiplexing, header compression
❌ Non lisible sans outil
❌ Complexité build (protoc)

---

## Slide 3 — Panorama des protocoles (2/2) : Alternatives

**Titre :** GraphQL et Message Brokers — Quand les utiliser ?

**GraphQL**
- Le client choisit exactement les champs qu'il veut
- ✅ Élimine over/under-fetching
- ❌ Pas de cache HTTP · Pas de streaming natif · N+1 problem
- **Pour SignalWatch :** inadapté au flux haute fréquence inter-services

**Apache Kafka**
```
Capteur → [publish] → Kafka topic → [consume] → Microservice
                                  → [consume] → Analytics
```
- Communication **asynchrone** — découplage total
- ✅ Débit extrême · Durabilité · Replay
- ❌ Latence plus élevée (batch) · Complexité opérationnelle
- **Pour SignalWatch :** complémentaire (pipeline de traitement) — pas un remplacement

**RabbitMQ**
- AMQP · Routage flexible · Support MQTT natif (IoT edge)
- ✅ Latence faible · Routing sophistiqué
- **Pour SignalWatch :** pertinent pour la collecte edge → cloud avec MQTT

---

## Slide 4 — Résultats benchmark : Scénario A — Lecture

**Titre :** Lecture unitaire — 1 000 requêtes, 10 connexions

| Métrique | REST (k6) | gRPC (ghz) |
|----------|:---------:|:----------:|
| Latence p50 | **0,53 ms** | ~1,4 ms |
| Latence p95 | **1,26 ms** | 3,25 ms |
| Latence p99 | **2,24 ms** | 7,33 ms |
| Débit (RPS) | **14 324** | 5 241 |
| Taille réponse | 225 B | **135 B** |
| Erreurs | 0 % | 0 % |

**Graphique :** barres groupées p95/p99 REST vs gRPC (voir `benchmark/results/`)

**Key message :**
> À faible charge en loopback, REST est 2,7× plus rapide — le surcoût HTTP/2 ne s'amortit pas.
> Mais gRPC est 40 % plus compact sur le payload.

---

## Slide 5 — Résultats benchmark : Scénario B — Écriture

**Titre :** Écriture (Create) — 500 requêtes, 5 connexions

| Métrique | REST (k6) | gRPC (ghz) |
|----------|:---------:|:----------:|
| Latence p95 | **1,20 ms** | 1,53 ms |
| Latence p99 | **1,85 ms** | 2,80 ms |
| Débit (RPS) | **6 506** | 4 087 |
| Taille réponse | 232 B | **135 B** |
| Erreurs | 0 % | 0 % |

**Key message :**
> L'écart se resserre en écriture (1,6× vs 2,7× en lecture).
> La charge de sérialisation symétrique atténue l'avantage REST.

---

## Slide 6 — Résultats benchmark : Scénario C — Montée en charge

**Titre :** Comportement sous charge : 10 → 100 connexions

**REST (ramp continu k6) :**
| RPS moyen | p95 | p99 |
|:---------:|:---:|:---:|
| 37 385 | 3,46 ms | **8,02 ms** |

**gRPC (paliers ghz) :**
| Connexions | RPS | p99 |
|:----------:|:---:|:---:|
| 10 | 7 530 | 2,96 ms |
| 50 | 11 298 | 22,04 ms |
| 100 | 17 032 | **34,07 ms** |

**Graphique :** courbe RPS et p99 par palier de concurrence gRPC (voir résultats)

**Key message :**
> REST : p99 stable à 8 ms. gRPC : p99 explose à 34 ms à 100 connexions.
> Risque SLO pour gRPC en pics de trafic — à surveiller en production.

---

## Slide 7 — Analyse éco-conception

**Titre :** Éco-conception — gRPC gagne sur tous les fronts

**Payload :**
```
JSON   ████████████████████  230 B  (100 %)
Proto  ████████████          135 B  ( 59 %)   → −41 %
```

**CPU sous charge maximale (scénario C) :**
```
REST   ████████████████████  325 % moy  (12× plus)
gRPC   ██                     27 % moy
```

**RAM sous charge maximale :**
```
REST   27,4 Mo · gRPC   17,5 Mo   → −36 %
```

**Extrapolation SignalWatch — 10 000 evt/min :**

| | REST | gRPC | Économie |
|-|:----:|:----:|:--------:|
| Bande passante/jour | 3,31 GB | 1,94 GB | **−1,37 GB** |
| Bande passante/an | ~1,2 TB | ~0,7 TB | **−500 GB** |

**Key message :** gRPC = −41 % réseau · −92 % CPU · −36 % RAM. Choix éco-responsable (RGESN critère 6.1).

---

## Slide 8 — RGPD et réglementation IoT

**Titre :** Données de capteurs et RGPD — Ce qu'il faut savoir

**Quelles données sont personnelles chez SignalWatch ?**

| Champ | Qualification | Action |
|-------|--------------|--------|
| `location` "Bâtiment C - Salle 12" | ⚠️ Potentiellement personnel | Pseudonymiser si lié à un poste nominatif |
| `last_value` + `last_reading_at` | ⚠️ Personnel si corrélé à un opérateur | Traiter comme donnée de surveillance |
| `name` capteur | ✅ Non personnel (machine) | Pas d'action spécifique |

**Obligations clés :**
- **Base légale** : intérêt légitime (art. 6.1.f) — le plus courant en industrie
- **Durée de conservation** : 6 mois (temps réel) → 5 ans (incidents ICPE)
- **Droits des personnes** : accès, opposition, effacement (sous réserve obligations légales)

**Spécificités industrielles :**
- **ICPE** : obligations de conservation des mesures imposées par arrêté préfectoral
- **Directive NIS2** : notification incidents < 24h, audit cybersécurité

**Recommandation :** Privacy by Design — collecte minimale, TLS 1.3, registre des traitements (art. 30).

---

## Slide 9 — Matrice de décision et recommandation

**Titre :** Notre recommandation pour SignalWatch

**Matrice multicritères (score /5) :**

| Critère | Poids | REST | gRPC |
|---------|:-----:|:----:|:----:|
| Latence haute charge (WAN) | 15 % | ★★★ | ★★★★ |
| Bande passante / payload | 15 % | ★★★ | ★★★★★ |
| Débit (RPS) sous charge | 15 % | ★★★★ | ★★★ |
| Consommation CPU/RAM | 10 % | ★★ | ★★★★★ |
| Maintenabilité / contrat | 10 % | ★★★ | ★★★★ |
| Streaming temps réel | 10 % | ★★ | ★★★★★ |
| Interopérabilité | 10 % | ★★★★★ | ★★★ |
| Éco-conception | 5 % | ★★★ | ★★★★★ |
| Courbe d'apprentissage | 5 % | ★★★★★ | ★★★ |
| **Score pondéré** | **100 %** | **3,35** | **3,90** |

**→ Recommandation : Architecture hybride**

```
Clients Web/Mobile  ←→  REST API Gateway  ←→  [Envoy proxy]  ←→  gRPC inter-services
```

- **gRPC** : communication backend ↔ backend (haute fréquence, streaming)
- **REST** : API externe (compatibilité universelle, lisibilité)

---

## Slide 10 — Conclusion et limites

**Titre :** Ce qu'on a mesuré — Ce qu'on ne sait pas encore

**Ce que nos mesures montrent :**
- ✅ REST est plus rapide à **faible charge en loopback** (2,7× en latence p95)
- ✅ gRPC est plus efficient en **CPU** (12×), **RAM** (−36 %), **réseau** (−41 %)
- ✅ gRPC supporte le **streaming natif** — clé pour 10 000 evt/min
- ⚠️ La p99 gRPC se dégrade à haute concurrence (34 ms à 100 conn.) — à surveiller

**Ce que nos mesures ne montrent pas :**
- ❌ Latence en **WAN réel** (50 ms RTT, TLS, load balancer)
- ❌ Impact d'une **base de données** réelle (TimescaleDB, PostgreSQL)
- ❌ Performance du **gRPC streaming** vs polling REST

**Prochaines étapes recommandées :**
1. Reproduire en WAN (simuler 50 ms RTT avec `tc netem`)
2. Implémenter `StreamSensorReadings` en gRPC et comparer
3. Évaluer MQTT pour la collecte edge → cloud
4. Activer TLS et mesurer l'impact sur la latence

**Message final :**
> Les chiffres confirment gRPC pour le backend SignalWatch. La p99 sous charge appelle à la vigilance. Le monitoring continu (OpenTelemetry) est indispensable en production.

---

*Présentation BenchLab — Arthur Descourvieres, Membre 2, Membre 3 — mai 2026*
*Données : `benchmark/results/` · Code : github.com/ArthurDescourvieres/BenchLab*
