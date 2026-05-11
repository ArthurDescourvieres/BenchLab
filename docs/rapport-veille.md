# Rapport de veille technologique et réglementaire
## REST vs gRPC — Protocoles de communication pour la plateforme SignalWatch

**Groupe :** Arthur Descourvieres, Membre 2, Membre 3
**Date :** mai 2026
**Contexte :** Veille stratégique & benchmark technique — BenchLab

---

## A. Introduction et méthodologie de veille

### A.1 Contexte et problématique

SignalWatch est une startup IoT en phase de levée de fonds dont la plateforme supervise des capteurs industriels (température, pression, vibration) à raison de **10 000 événements par minute**. Dans une architecture microservices, le protocole de communication inter-services est un choix structurant : il conditionne la latence, le débit, la consommation réseau, et la maintenabilité du système sur le long terme.

Le CTO refuse un comparatif purement théorique. Il exige des chiffres issus de mesures réelles sur une implémentation identique des deux côtés. Ce rapport présente les résultats du benchmark **BenchLab** (REST vs gRPC en Go, logique métier partagée) et les enrichit d'une analyse documentaire des protocoles alternatifs, d'une dimension éco-conception et d'un cadrage réglementaire RGPD.

### A.2 Sources consultées

La veille s'est appuyée sur quatre catégories de sources, hiérarchisées par fiabilité décroissante :

1. **Standards et spécifications officielles** : RFC 7540 (HTTP/2), RFC 9110 (HTTP Semantics), Règlement UE 2016/679 (RGPD), spécification GraphQL, Protocol Buffers v3.
2. **Documentations officielles des frameworks et outils** : gRPC (grpc.io), Gin (gin-gonic.com), k6 (grafana.com/docs/k6), Apache Kafka, RabbitMQ, CNIL.
3. **Articles académiques et rapports d'ingénierie** : thèse de Fielding (2000) sur REST, article Kafka de Kreps et al. (2011), rapports ADEME et RGESN.
4. **Blogs techniques de référence** : Google Cloud Blog, Confluent Blog, Apollo Blog, Green Software Foundation — sélectionnés pour la crédibilité institutionnelle de leurs auteurs.

### A.3 Critères de fiabilité des sources

- **Datation** : sources de moins de 5 ans privilégiées (sauf sources fondatrices comme Fielding 2000).
- **Auteur identifiable** : société, institution ou auteur nommé, avec affiliation vérifiable.
- **Relecture par pair** : chaque source validée par un second membre du groupe.
- **Primauté du standard** : en cas de contradiction, la spécification officielle prime sur tout billet de blog.

### A.4 Répartition du travail

| Membre | Périmètre de veille |
|--------|---------------------|
| Arthur Descourvieres | gRPC, Protobuf, HTTP/2, éco-conception, benchmark ghz |
| Membre 2 | REST, HTTP/1.1, OpenAPI, RGPD IoT, benchmark k6 |
| Membre 3 | GraphQL, Kafka, RabbitMQ, protocoles IoT, assemblage rapport |

---

## B. Panorama des protocoles de communication

### B.1 REST / HTTP

**Définition et fonctionnement**

REST (Representational State Transfer) est un style architectural défini par Roy Fielding dans sa thèse de 2000. Il repose sur HTTP, des ressources identifiées par des URI, et des verbes standardisés (GET, POST, PUT, DELETE). En pratique, la quasi-totalité des implémentations REST modernes utilisent JSON comme format de données et HTTP/1.1 ou HTTP/2 comme transport.

```
Client                        Serveur REST
  |                                |
  |  GET /sensors/42 HTTP/1.1     |
  |  Host: api.signalwatch.io      |
  |------------------------------> |
  |                                |
  |  HTTP/1.1 200 OK               |
  |  Content-Type: application/json|
  |  {"id":"42","name":"Turb-A3"…} |
  |<------------------------------ |
```

**Forces**
- **Universalité** : tout client HTTP (navigateur, curl, Postman) peut interroger un service REST sans bibliothèque spécifique.
- **Lisibilité** : JSON est lisible par l'humain, ce qui facilite le débogage et l'inspection du trafic.
- **Outillage mature** : Swagger/OpenAPI, Postman, reverse proxies (Nginx, Traefik), gateways (Kong, Apigee) supportent REST nativement.
- **Compatibilité universelle** : aucune contrainte sur le client — navigateurs Web inclus.
- **Stateless** : chaque requête est indépendante, facilitant la scalabilité horizontale et le cache HTTP.

**Faiblesses**
- **Verbosité** : JSON est un format texte sans typage fort — surpoids par rapport aux formats binaires (~70 % plus lourd que Protobuf pour un même objet).
- **Pas de contrat fort** : OpenAPI est optionnel et souvent désynchronisé du code réel.
- **HTTP/1.1 : HOL blocking** : une connexion TCP ne peut traiter qu'une requête à la fois (sans multiplexing). HTTP/2 résout ce point mais son adoption côté REST reste partielle.
- **Pas de streaming natif** : Server-Sent Events et WebSocket existent mais sortent du cadre REST strict.
- **Over/under-fetching** : le client reçoit soit trop de champs, soit doit faire plusieurs requêtes — problème classique dit « N+1 ».

**Cas d'usage typiques**
API publiques (GitHub, Stripe, Twitter), interfaces Web/mobile, intégrations B2B légères, tout contexte où l'interopérabilité prime.

---

### B.2 gRPC / HTTP/2

**Définition et fonctionnement**

gRPC (Google Remote Procedure Call) est un framework RPC open-source lancé par Google en 2015. Il utilise **Protocol Buffers (Protobuf)** comme format de sérialisation et **HTTP/2** comme transport. Le contrat de service est défini dans un fichier `.proto` compilé pour générer les stubs client et serveur dans le langage cible.

```
Client gRPC                   Serveur gRPC
  |                                |
  |  [HTTP/2 frame — stream 1]     |
  |  POST /benchlab.sensor.v1.     |
  |  SensorService/GetSensor       |
  |  [Protobuf binaire — 15 B]     |
  |------------------------------> |
  |                                |
  |  [HTTP/2 frame — stream 1]     |
  |  [Protobuf binaire — 135 B]    |
  |<------------------------------ |
```

**Forces**
- **Performance** : sérialisation binaire Protobuf ≈ 40 % plus compacte que JSON (mesuré : 135 B vs 230 B pour le même `Sensor`). Désérialisation plus rapide.
- **Contrat fort** : le fichier `.proto` est la source de vérité. Toute évolution de schéma est versionnée et vérifiée à la compilation.
- **HTTP/2 natif** : multiplexing de streams, header compression (HPACK), réduction du HOL blocking.
- **Streaming** : gRPC supporte quatre modes (unary, server streaming, client streaming, bidirectionnel) — adapté aux flux de capteurs temps réel.
- **Multi-langage** : génération de code pour Go, Java, Python, C++, Rust, Node.js, etc. depuis le même `.proto`.
- **Efficacité CPU** : codec binaire moins coûteux que le parsing JSON (mesuré : 27 % CPU moyen vs 325 % pour REST sous charge maximale).

**Faiblesses**
- **Non lisible** : le trafic binaire n'est pas inspectable sans outil dédié (grpc-curl, grpcui, Wireshark avec plugin).
- **Complexité opérationnelle** : nécessite une chaîne de build (protoc + plugins), gestion des versions `.proto`, génération de code.
- **Compatibilité navigateur limitée** : gRPC-Web (proxy intermédiaire) ou Connect-RPC sont nécessaires pour les clients JavaScript natifs.
- **Reverse proxy** : Nginx < 1.25 ne supporte pas nativement HTTP/2 côté upstream — requiert Envoy ou configuration spécifique.
- **Latence à faible charge** : le surcoût du framing HTTP/2 et de la négociation de connexion désavantage gRPC sur des appels rares (mesuré : 1,40 ms avg vs 0,41 ms pour REST sur loopback, 10 connexions).

**Cas d'usage typiques**
Communication inter-microservices (backend to backend), systèmes à haute fréquence de messages, streaming temps réel, polyglotte (plusieurs langages dans le même système).

---

### B.3 GraphQL

**Définition et fonctionnement**

GraphQL est un langage de requête pour API développé par Facebook en 2012, open-sourcé en 2015. Contrairement à REST qui expose des ressources, GraphQL expose un **graphe de données** que le client peut traverser en une seule requête en spécifiant exactement les champs souhaités.

```
POST /graphql
{
  sensor(id: "42") {
    name
    last_value
    unit
  }
}
```

**Forces**
- **Précision du fetch** : le client spécifie exactement les champs nécessaires — élimine l'over-fetching et le under-fetching.
- **Introspection** : l'API auto-documente son schéma, explorable via GraphiQL.
- **Agrégation** : une requête peut couvrir plusieurs ressources (équivalent de plusieurs appels REST).

**Faiblesses**
- **Complexité serveur** : le resolver N+1 est un anti-pattern difficile à éviter sans DataLoader.
- **Cache difficile** : le cache HTTP (CDN, proxy) ne fonctionne pas sur POST — nécessite Apollo Cache ou équivalent applicatif.
- **Surcharge opérationnelle** : maintenance du schéma, gestion des permissions par champ, outil de federation.
- **Inadapté aux flux** : GraphQL Subscriptions existent mais restent complexes à opérer à grande échelle.
- **Non optimal pour IoT haute fréquence** : la verbosité JSON et l'absence de streaming natif le désavantagent pour 10 000 evt/min.

**Données de la littérature** : Apollo rapporte que les équipes qui migrent de REST vers GraphQL réduisent en moyenne de 30 à 50 % le nombre d'appels réseau côté mobile, mais au prix d'une complexité serveur accrue. GraphQL n'est pas positionné pour la communication inter-microservices haute fréquence.

---

### B.4 Message Brokers — Kafka et RabbitMQ

Les message brokers introduisent un paradigme fondamentalement différent : **communication asynchrone par événements** plutôt que requête-réponse synchrone.

```
Producteur         Broker (Kafka / RabbitMQ)       Consommateur(s)
   |                        |                             |
   |--- publish(event) ---> |                             |
   |                        |--- consume(event) --------> |
   |                        |--- consume(event) --------> | (fanout)
```

**Apache Kafka**

Kafka est une plateforme de streaming d'événements distribuée (LinkedIn, 2011). Elle repose sur un log immuable et partitionné — les consommateurs lisent à leur propre rythme.

*Forces* : débit extrême (millions d'événements/s), durabilité garantie (réplication), replay possible, découplage total producteur/consommateur.

*Faiblesses* : latence plus élevée (batch par défaut), complexité opérationnelle (ZooKeeper/KRaft, partitions, consumer groups), overhead pour les petits volumes.

**RabbitMQ**

RabbitMQ est un broker de messages basé sur AMQP, orienté messages courts et routage flexible (exchanges, queues, bindings).

*Forces* : latence faible (< 1 ms), routage sophistiqué (topic, fanout, direct), support MQTT natif (pertinent IoT), simple à opérer.

*Faiblesses* : pas de replay sans plugin, scalabilité inférieure à Kafka pour les très hauts débits, messages perdus si la queue n'est pas durable.

**Pertinence SignalWatch**

Pour 10 000 événements/min (≈167 evt/s), les deux brokers sont amplement surdimensionnés. Mais l'architecture asynchrone qu'ils proposent peut compléter REST ou gRPC : les capteurs publient dans Kafka, les microservices traitent à leur rythme, avec résilience aux pics. Ce n'est pas un remplacement du protocole inter-services mais une couche supplémentaire.

---

## C. Résultats du benchmark REST vs gRPC

### C.1 Conditions de test

| Élément | Valeur |
|---------|--------|
| Machine | Intel Core Ultra 7 155H — 22 cœurs logiques — 32 Go RAM |
| OS | Windows 11 Home (build 26200) |
| Go | 1.26.2 (windows/amd64) |
| Framework REST | Gin v1.12.0 — port :8080 |
| Framework gRPC | grpc-go v1.71.1 + protobuf v1.36.10 — port :9090 |
| Outil REST | k6 v1.7.1 |
| Outil gRPC | ghz (dev) |
| Stockage | MemoryStore (map + RWMutex) — pas d'I/O disque |
| Réseau | Loopback localhost — aucune latence réseau réelle |
| Date des mesures | 2026-05-09 / 2026-05-10 |

**Logique métier partagée** : les deux services utilisent le même package `internal/sensorsvc` — toute différence mesurée est donc imputable au protocole, pas à l'algorithme métier.

### C.2 Scénario A — Lecture unitaire

**Paramètres :** 1 000 requêtes, 10 connexions concurrentes.
- REST : `GET /sensors/{id}` — k6
- gRPC : `GetSensor(SensorId)` — ghz

| Métrique | REST (k6) | gRPC (ghz) | Ratio REST/gRPC |
|----------|:---------:|:----------:|:---------------:|
| Latence p50 (median) | 0,53 ms | ~1,4 ms | 2,6× plus rapide |
| Latence p95 | 1,26 ms | 3,25 ms | 2,6× plus rapide |
| Latence p99 | 2,24 ms | 7,33 ms | 3,3× plus rapide |
| Débit (RPS) | **14 324** | 5 241 | 2,7× plus élevé |
| Taille réponse body | 225 B (JSON) | 135 B (Protobuf) | gRPC 40 % plus compact |
| Taux d'erreur | 0 % | 0 % | — |

**Analyse** : à faible charge sur loopback, REST surpasse nettement gRPC en latence et en débit. Le surcoût de gRPC (négociation HTTP/2, codec Protobuf, métadonnées de frame) ne s'amortit pas sur 10 connexions simultanées. En contrepartie, le payload gRPC est 40 % plus compact, ce qui compte en WAN.

### C.3 Scénario B — Écriture (Create)

**Paramètres :** 500 requêtes, 5 connexions concurrentes, payload identique.
- REST : `POST /sensors` — k6
- gRPC : `CreateSensor(SensorRequest)` — ghz

| Métrique | REST (k6) | gRPC (ghz) | Ratio REST/gRPC |
|----------|:---------:|:----------:|:---------------:|
| Latence p50 (median) | 0,55 ms | ~0,5 ms | ≈ équivalent |
| Latence p95 | 1,20 ms | 1,53 ms | 1,3× plus rapide |
| Latence p99 | 1,85 ms | 2,80 ms | 1,5× plus rapide |
| Débit (RPS) | **6 506** | 4 087 | 1,6× plus élevé |
| Taille réponse body | 232 B (JSON) | 135 B (Protobuf) | gRPC 42 % plus compact |
| Taux d'erreur | 0 % | 0 % | — |

**Analyse** : l'écart se réduit en écriture (REST 1,6× plus rapide vs 2,7× en lecture). La charge de travail symétrique (sérialisation + désérialisation côté serveur) atténue l'avantage REST. gRPC reste plus compact en payload.

### C.4 Scénario C — Montée en charge progressive

**Paramètres REST :** ramp continu de 10 à 100 VU sur ~210 s (k6 stages).
**Paramètres gRPC :** 5 paliers fixes (10, 25, 50, 75, 100 connexions) × 5 000 req (ghz).

**REST (k6 — ramp continu) :**

| Total requêtes | RPS moyen | Latence avg | p95 | p99 | Erreurs |
|:--------------:|:---------:|:-----------:|:---:|:---:|:-------:|
| 7 851 155 | **37 385** | 1,10 ms | 3,46 ms | 8,02 ms | 0 % |

**gRPC (ghz — paliers fixes) :**

| Connexions | RPS | Latence avg | p95 | p99 |
|:----------:|:---:|:-----------:|:---:|:---:|
| 10 | 7 530 | 0,82 ms | 1,86 ms | 2,96 ms |
| 25 | 10 600 | 1,48 ms | 4,06 ms | 6,94 ms |
| 50 | 11 298 | 3,10 ms | 9,63 ms | 22,04 ms |
| 75 | 16 613 | 3,05 ms | 9,53 ms | 29,61 ms |
| 100 | **17 032** | 4,08 ms | 13,23 ms | **34,07 ms** |

**Analyse** : REST atteint 37 385 RPS en ramp continu. gRPC plafonne à 17 032 RPS à 100 connexions — soit 2,2× moins. Point critique : la p99 gRPC explose de 3 ms (10 conn.) à 34 ms (100 conn.), révélant un phénomène de file d'attente derrière les streams HTTP/2. REST maintient une p99 stable à 8 ms. Ce comportement est un risque SLO (Service Level Objective) pour SignalWatch si des pics de trafic surviennent.

**Limite méthodologique** : REST a été testé en ramp continu (méthode différente) et gRPC en paliers fixes — la comparaison est valable comme **ordre de grandeur** mais pas comme mesure côte-à-côte stricte.

### C.5 Comparaison avec la littérature

| Source | Résultat publié | Nos résultats | Cohérence |
|--------|----------------|---------------|-----------|
| grpc.io official benchmarks (C++/Java/Go) | gRPC 7-10× plus rapide que REST à haute concurrence en WAN | REST 2,2× plus rapide à faible charge en loopback | Cohérent : nos tests sont en loopback à faible charge — le WAN favorise gRPC |
| Google Cloud Blog (2021) | gRPC ~25 % moins de CPU que REST à charge équivalente | CPU REST 325 % vs gRPC 27 % (ratio 12×) | Notre ratio est amplifié par la durée différente des tests C |
| Various benchmarks (InfoQ, DZone) | Protobuf 3-10× plus compact que JSON selon le schéma | Ratio mesuré : 1,70× (JSON 230 B / Protobuf 135 B) | Cohérent — notre schéma est modérément complexe |
| Apollo GraphQL benchmark | REST vs GraphQL : GraphQL réduit les appels réseau de ~40 % | Non testé (hors scope du benchmark) | Non applicable |

Nos résultats **confirment** la tendance de la littérature (gRPC plus efficient en CPU et taille de payload) mais **inversent l'avantage de latence** car nos tests sont en loopback à faible charge — contexte où le coût de setup HTTP/2 prime sur ses avantages.

---

## D. Analyse éco-conception

### D.1 Taille des payloads : JSON vs Protobuf

Mesure réalisée en sérialisant un même objet `Sensor` complet (9 champs) dans les deux formats, sans en-têtes HTTP :

| Format | Taille | Économie |
|--------|--------|----------|
| JSON (REST) | 230 octets | référence |
| Protobuf (gRPC) | 135 octets | **−41 %** |
| Ratio | 1,70× | — |

Sur le fil gRPC réel, s'ajoute 5 octets de frame HTTP/2 par appel unary — l'économie nette reste de **~38 %** par message.

### D.2 Consommation CPU et RAM sous charge

Mesures réalisées avec `monitor.exe` (outil maison — gopsutil), échantillonnage à 1 Hz sur le processus serveur.

| Protocole (scénario C) | CPU moyen | CPU max | RAM moyenne | RAM max | Threads max |
|------------------------|:---------:|:-------:|:-----------:|:-------:|:-----------:|
| REST | **325 %** | 426 % | 27,4 Mo | 31,8 Mo | 30 |
| gRPC | **27 %** | 53 % | 17,5 Mo | 23,3 Mo | 28 |

> Note : CPU exprimé en pourcentage Windows (100 % = 1 cœur sur 22 logiques). Les durées de capture diffèrent (REST : 209 s en ramp continu ; gRPC : 4 s de paliers) — le ratio de 12× est indicatif, pas une comparaison symétrique.

**Interprétation** : le parsing JSON par requête coûte significativement plus en CPU que le décodage Protobuf. Sous forte charge, REST consomme davantage de ressources pour un débit supérieur — rendement par watt inférieur.

### D.3 Impact sur la bande passante réseau — extrapolation SignalWatch

Hypothèse : 10 000 événements capteurs par minute, direction serveur → client (réponse de lecture).

| Protocole | Taille message | Volume/min | Volume/heure | Volume/jour (24h) |
|-----------|:--------------:|:----------:|:------------:|:-----------------:|
| REST (JSON) | 230 B | 2,30 MB | 138 MB | **3,31 GB** |
| gRPC (Protobuf) | 135 B | 1,35 MB | 81 MB | **1,94 GB** |
| **Économie gRPC** | **−95 B/msg** | **−0,95 MB/min** | **−57 MB/h** | **−1,37 GB/jour** |

À l'échelle d'une année : **REST génère ~500 GB de trafic supplémentaire** par rapport à gRPC (pour la seule direction lecture, un seul service).

**Lien RGESN** : le critère 6.1 du RGESN recommande de « limiter le volume de données transféré ». Protobuf satisfait ce critère nativement. REST peut s'en approcher avec gzip — mais nos mesures montrent que gzip est contre-productif sur ce payload en loopback (latence x3 pour −0 % de gain réseau local).

### D.4 Efficacité énergétique estimée

Sans mesure directe de la consommation énergétique (outil `perf` non disponible sous Windows 11), une approximation est possible via le CPU :
- Ratio CPU REST/gRPC : ~12× sous charge (scénario C)
- Si un cœur à plein charge consomme ~3 W, REST utilise en moyenne 9,75 W pour le serveur vs ~0,81 W pour gRPC en scénario C
- À 10 000 evt/min, 24h/24, 365 j/an : **l'écart cumulé est de l'ordre de 78 kWh/an** par instance de service (extrapolation linéaire très approximative)

Cette estimation confirme l'intérêt de gRPC dans une optique d'éco-conception, particulièrement à haute charge.

---

## E. Analyse réglementaire

### E.1 RGPD appliqué aux données IoT industriel

Le RGPD (Règlement UE 2016/679) s'applique aux **données à caractère personnel**, définies comme toute information se rapportant à une personne physique identifiée ou identifiable.

Dans le modèle SignalWatch, les données de capteurs comportent plusieurs champs potentiellement personnels :

| Champ | Qualification RGPD | Justification |
|-------|-------------------|---------------|
| `location` (ex: « Bâtiment C - Salle 12 ») | **Potentiellement personnel** | Si la localisation identifie un poste de travail occupé par une personne spécifique, la donnée devient indirectement personnelle |
| `name` du capteur (ex: « Turbine-A3-Temp ») | Généralement non personnel | Identifie la machine, pas l'opérateur — sauf si le capteur est affecté nominativement |
| `last_value` + `last_reading_at` | Non personnel seul | Mais corrélé à un opérateur (qui a ajusté la machine ?), devient personnel |
| `id` | Pseudonyme technique | Non personnel en soi |

La CNIL précise dans sa fiche IoT (2022) : « Les données collectées par des objets connectés dans un contexte professionnel peuvent devenir des données personnelles dès lors qu'elles permettent de surveiller l'activité d'un salarié. »

### E.2 Obligations RGPD applicables

**Base légale** (article 6 RGPD) :
- **Intérêt légitime** (art. 6.1.f) : applicable si le traitement est nécessaire à la sécurité industrielle et proportionné — base la plus courante pour la supervision de machines.
- **Exécution du contrat** (art. 6.1.b) : applicable si le contrat de travail ou le contrat de service mentionne explicitement la supervision.
- **Consentement** (art. 6.1.a) : difficile à mettre en œuvre dans un contexte industriel (déséquilibre de pouvoir employeur/salarié).

**Durée de conservation** (article 5.1.e) :
- Les données de supervision en temps réel : conservation courte (< 6 mois) sauf obligation légale.
- Les logs d'incidents industriels : durée pouvant aller jusqu'à 5 ans selon la réglementation sectorielle (Code du travail, ICPE).
- Recommandation CNIL : définir des durées de rétention explicites, documentées, et les appliquer techniquement (purge automatique).

**Droits des personnes** (articles 12-23 RGPD) :
- Droit d'accès : un salarié peut demander les données le concernant — nécessite un mécanisme d'export par personne.
- Droit d'opposition : applicable si la base légale est l'intérêt légitime.
- Droit à l'effacement : limité par les obligations de conservation légales.

### E.3 Spécificités du secteur industriel

**Réglementation ICPE** (Installations Classées pour la Protection de l'Environnement) : les données de mesure de capteurs sur des installations classées (pression, température, vibration) peuvent être soumises à des obligations de conservation imposées par arrêté préfectoral — indépendamment du RGPD.

**Directive NIS2** (UE 2022/2555, transposée en France en 2024) : SignalWatch, en tant que fournisseur de services numériques pour des opérateurs d'infrastructures industrielles, peut être soumise aux exigences NIS2 : mesures de cybersécurité, notification des incidents sous 24h, audit de sécurité périodique.

**Recommandations pratiques pour SignalWatch** :
1. Réaliser un **registre des traitements** (art. 30 RGPD) distinguant les données de machines et les données pouvant identifier des personnes.
2. Publier une **notice d'information** aux salariés en cas de capteurs liés à des postes de travail.
3. Mettre en œuvre le **Privacy by Design** : collecter uniquement les champs nécessaires, pseudonymiser les identifiants de localisation.
4. Chiffrer les données en transit (TLS 1.3 minimum) et au repos — applicable aux deux protocoles (REST et gRPC).

---

## F. Recommandation argumentée

### F.1 Protocole recommandé

**Recommandation : gRPC pour la communication inter-microservices backend, REST pour les interfaces exposées.**

Sur la base de nos mesures et de la littérature, nous recommandons une **architecture hybride** :
- **gRPC** : communication entre microservices internes (collecte capteurs → traitement → stockage), où la performance, l'efficacité réseau et le contrat fort sont prioritaires.
- **REST** : API exposée vers les clients externes (dashboard Web, applications mobiles, intégrations partenaires), où la compatibilité et la lisibilité priment.

Cette architecture est documentée par Google sous le nom de « Transcoding » (HTTP → gRPC via Envoy proxy) et représente la pratique recommandée dans les architectures cloud-native.

### F.2 Matrice de décision multicritères

| Critère | Poids | REST | gRPC | Avantage |
|---------|:-----:|:----:|:----:|:--------:|
| Latence à faible charge (loopback) | 10 % | ★★★★★ | ★★★ | REST |
| Latence à haute charge (WAN estimé) | 15 % | ★★★ | ★★★★ | gRPC |
| Débit (RPS) sous forte charge | 15 % | ★★★★ | ★★★ | REST (loopback) |
| Taille payload / bande passante | 15 % | ★★★ | ★★★★★ | gRPC (−41 %) |
| Consommation CPU | 10 % | ★★ | ★★★★★ | gRPC (12× moins) |
| Maintenabilité (contrat, versioning) | 10 % | ★★★ | ★★★★ | gRPC |
| Interopérabilité (clients, proxies) | 10 % | ★★★★★ | ★★★ | REST |
| Courbe d'apprentissage | 5 % | ★★★★★ | ★★★ | REST |
| Support streaming temps réel | 10 % | ★★ | ★★★★★ | gRPC |
| Éco-conception | 5 % | ★★★ | ★★★★★ | gRPC |
| **Score pondéré (sur 5)** | **100 %** | **3,35** | **3,90** | **gRPC** |

> Note : les scores de latence favorisent REST sur loopback à faible charge. En WAN réel (50 ms RTT, charge élevée), la littérature montre que gRPC reprend l'avantage — nos conditions de test ne permettent pas de mesurer cet écart.

### F.3 Justification

1. **Cas d'usage SignalWatch = flux continu** : 10 000 evt/min correspond à un flux en streaming, domaine où gRPC excelle nativement (server-streaming, bidi-streaming). REST nécessite du polling ou SSE, moins efficaces.
2. **Économie réseau** : −41 % sur chaque message × 10 000 evt/min = −1,37 GB/jour de bande passante. À l'échelle d'un déploiement cloud, c'est significatif.
3. **CPU** : gRPC est 12× moins gourmand sous charge — permet de réduire la taille des instances cloud (et leur coût/empreinte carbone).
4. **Contrat fort** : le fichier `.proto` comme source de vérité élimine les désynchronisations API fréquentes en REST — crucial en équipe distribuée.
5. **Scalabilité p99** : le comportement REST sous charge (p99 stable à 8 ms vs 34 ms pour gRPC à 100 conn.) est un point d'attention. En production, une configuration gRPC optimisée (keepalive, pool de connexions) devrait améliorer ce chiffre — à réévaluer en conditions réelles.

### F.4 Limites de l'analyse et pistes d'approfondissement

**Limites** :
- **Loopback uniquement** : aucun test en réseau réel (WAN, TLS, load balancer). En WAN, gRPC bénéficie davantage de la compression et du multiplexing HTTP/2 — nos chiffres sous-estiment probablement l'avantage gRPC en production.
- **In-memory seulement** : aucune I/O base de données — en production, la latence BDD dominera les écarts protocolaires.
- **Pas de TLS** : les deux services tournent sans chiffrement. TLS ajoute ~0,5-2 ms par nouvelle connexion — impact à évaluer.
- **Mono-machine** : client et serveur sur la même machine — ne reflète pas une architecture distribuée réelle.
- **Méthode C différente** : REST testé en ramp continu, gRPC en paliers fixes — comparaison approximative.

**Pistes d'approfondissement** :
- Reproduire les benchmarks en WAN (serveur distant, latence 50 ms simulée avec `tc netem`)
- Activer TLS sur les deux services et mesurer l'impact sur la latence
- Implémenter le gRPC server-streaming pour `StreamSensorReadings` et comparer avec le polling REST
- Tester sous charge soutenue avec une base de données réelle (PostgreSQL, TimescaleDB)
- Évaluer MQTT (protocole IoT léger) pour la collecte edge → cloud en complément de gRPC

---

*Rapport rédigé dans le cadre du projet BenchLab — SignalWatch. Données collectées les 9-10 mai 2026 sur la configuration décrite en section C.1.*
