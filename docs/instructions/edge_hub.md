# Istruzioni per l'Edge Hub

L'Edge Hub è implementato come un'applicazione distribuita Go che opera al livello più vicino ai sensori fisici.

* **Ingestione e Registrazione Dati:** L'Edge Hub agisce come consumer per i dati inviati dai Sensor Agents tramite un broker MQTT locale (broker interno all'Edge Zone o all'Edge Macrozone). Riceve, valida e registra tutti i messaggi.
* **Architettura a Microservizi Cooperanti:** L'applicazione principale può essere avviata in diverse modalità operative (*Service Modes*), permettendo a componenti specializzati di coesistere e lavorare sullo stesso set di dati (tramite Redis Cache), ad esempio:
  * **Filtro:** Applica il rilevamento degli outlier in tempo reale su ogni serie di dati ricevuta, utilizzando metodi statistici.
  * **Aggregatore:** Aggrega i dati filtrati a intervalli regolari, riducendo la granularità e il volume dei dati inviati ai livelli superiori.
  * **Cleaner:** Gestisce i timeout dei sensori, considerandoli come *unhealthy* se non inviano dati o heartbeat entro limiti predefiniti.
* **Stato Condiviso:** Utilizza un'istanza Redis locale per mantenere lo stato dei sensori (es. ultimi $N$ valori per il calcolo della deviazione standard, stato di salute, configurazione).
* **Resilienza e Coordinamento:** Grazie a Redis, supporta la logica di *Leader Election* in scenari a basse risorse computazionali, garantendo che solo un'istanza dell'Hub esegua i task critici di aggregazione e pulizia.

-----

## Variabili d'Ambiente dell'Edge Hub

Le variabili d'ambiente definiscono l'identità dell'Hub, i suoi endpoint di comunicazione e i parametri operativi di resilienza e aggregazione.

### A\. Identità e Modalità Operative

| Variabile            | Descrizione                                                                                                        | Default / Valori Ammessi                                                                                                                |
|:---------------------|:-------------------------------------------------------------------------------------------------------------------|:----------------------------------------------------------------------------------------------------------------------------------------|
| **`EDGE_MACROZONE`** | **Obbligatoria.** Identifica la macrozona logica a cui appartiene l'Hub.                                           | Nessun Default                                                                                                                          |
| **`EDGE_ZONE`**      | **Obbligatoria.** Identifica la zona logica specifica.                                                             | Nessun Default                                                                                                                          |
| **`HUB_ID`**         | Identificatore univoco dell'istanza Hub.                                                                           | **UUID Generato**                                                                                                                       |
| **`OPERATION_MODE`** | Controlla il ciclo di vita del servizio.                                                                           | `loop` (ciclo infinito) / `once`                                                                                                        |
| **`SERVICE_MODE`**   | Definisce il ruolo operativo specifico di questa istanza Edge Hub.                                                 | `edge-hub` (default, tutte le funzionalità) / `edge-hub-filter` / `edge-hub-aggregator` / `edge-hub-cleaner` / `edge-hub-configuration` |

### B\. Configurazione MQTT

Il sistema distingue tra i broker usati per la comunicazione **sensor_agent-edge_hub** e **edge_hub-proximity_hub**.

| Variabile                         | Descrizione                                                                                             | Default                          |
|:----------------------------------|:--------------------------------------------------------------------------------------------------------|:---------------------------------|
| **`MQTT_BROKER_PROTOCOL`**        | Protocollo comune. Utilizzato come fallback per i parametri specifici.                                  | `tcp`                            |
| **`MQTT_BROKER_ADDRESS`**         | Indirizzo comune. Fallback per i parametri specifici.                                                   | `localhost`                      |
| **`MQTT_BROKER_PORT`**            | Porta comune. Fallback per i parametri specifici.                                                       | `1883`                           |
| **`MQTT_SENSOR_BROKER_PROTOCOL`** | Protocollo specifico del broker a cui i Sensor Agents si collegano e da cui l'Edge Hub consuma.         | Valore di `MQTT_BROKER_PROTOCOL` |
| **`MQTT_SENSOR_BROKER_ADDRESS`**  | Indirizzo specifico del broker a cui i Sensor Agents si collegano e da cui l'Edge Hub consuma.          | Valore di `MQTT_BROKER_ADDRESS`  |
| **`MQTT_SENSOR_BROKER_PORT`**     | Porta specifica del broker a cui i Sensor Agents si collegano e da cui l'Edge Hub consuma.              | Valore di `MQTT_BROKER_PORT`     |
| **`MQTT_HUB_BROKER_PROTOCOL`**    | Protocollo specifico del broker (tipicamente il Proximity Hub) a cui l'Edge Hub invia i dati aggregati. | Valore di `MQTT_BROKER_PROTOCOL` |
| **`MQTT_HUB_BROKER_ADDRESS`**     | Indirizzo specifico del broker (tipicamente il Proximity Hub) a cui l'Edge Hub invia i dati aggregati.  | Valore di `MQTT_BROKER_ADDRESS`  |
| **`MQTT_HUB_BROKER_PORT`**        | Porta specifica del broker (tipicamente il Proximity Hub) a cui l'Edge Hub invia i dati aggregati.      | Valore di `MQTT_BROKER_PORT`     |

**Nota sui Topic:** I topic MQTT sono composti dinamicamente utilizzando `$EDGE_MACROZONE` e `$EDGE_ZONE` (es. `sensor-data/RomaMacro/TorVergata`) per garantire un'organizzazione gerarchica e isolata per zona.

### C\. Parametri di Resilienza MQTT

Questi parametri controllano il comportamento del client MQTT dell'Hub in caso di disconnessioni o fallimenti di pubblicazione/sottoscrizione.

| Variabile                       | Descrizione                                                       | Default        |
|:--------------------------------|:------------------------------------------------------------------|:---------------|
| **`MAX_RECONNECTION_INTERVAL`** | Intervallo massimo tra i tentativi di riconnessione.              | $10$ secondi   |
| **`MAX_RECONNECTION_TIMEOUT`**  | Timeout massimo per un singolo tentativo di riconnessione.        | $10$ secondi   |
| **`MAX_RECONNECTION_ATTEMPTS`** | Numero massimo di tentativi di riconnessione prima di arrendersi. | $10$ tentativi |
| **`MAX_SUBSCRIPTION_TIMEOUT`**  | Timeout per l'operazione di sottoscrizione ai topic.              | $5$ secondi    |
| **`MESSAGE_PUBLISH_TIMEOUT`**   | Timeout per la pubblicazione di un singolo messaggio.             | $5$ secondi    |
| **`MESSAGE_PUBLISH_ATTEMPTS`**  | Numero di tentativi di pubblicazione di un messaggio.             | $3$ tentativi  |
| **`MESSAGE_CLEANING_TIMEOUT`**  | Timeout relativo alla pulizia dei messaggi non inviati.           | $5$ secondi    |

### D\. Configurazione Redis

| Variabile           | Descrizione                                                     | Default     |
|:--------------------|:----------------------------------------------------------------|:------------|
| **`REDIS_ADDRESS`** | Indirizzo dell'istanza Redis utilizzata per lo stato condiviso. | `localhost` |
| **`REDIS_PORT`**    | Porta dell'istanza Redis.                                       | `6379`      |

### E\. Costanti di Elaborazione

Queste impostazioni, definite come costanti nell'ambiente, governano l'algoritmo di filtraggio e la logica di gestione dei sensori.

| Nome Costante                   | Valore             | Descrizione Funzionale                                                                                                                                                                               |
|:--------------------------------|:-------------------|:-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **`AggregationInterval`**       | $1$ minuto         | L'intervallo di tempo per cui i dati dei sensori vengono raccolti prima di essere aggregati in un unico record.                                                                                      |
| **`AggregationFetchOffset`**    | $-2$ minuti        | Offset di tempo negativo utilizzato per garantire che i dati con ritardi di rete vengano inclusi nell'aggregazione (es. l'aggregazione di un minuto finisce 2 minuti *dopo* l'intervallo).           |
| **`LeaderTTL`**                 | $70$ secondi       | Time-To-Live per la chiave di *Leader Election* in Redis. Se la chiave scade, un'altra istanza può reclamare il ruolo di Leader.                                                                     |
| **`HistoryWindowSize`**         | $100$              | Numero massimo di campioni storici memorizzati in Redis per ogni sensore. Utilizzato come base per l'algoritmo di filtraggio.                                                                        |
| **`FilteringStdDevFactor`**     | $3.0$              | Il fattore di deviazione standard $\sigma$ utilizzato per il rilevamento degli outlier. Un valore ricevuto è considerato outlier se si trova oltre $\pm 3\sigma$ dalla media della finestra storica. |
| **`UnhealthySensorTimeout`**    | $5$ minuti         | Il periodo di tempo dopo il quale un sensore che non invia dati o heartbeat viene marcato come *unhealthy* dal Cleaner Service.                                                                      |
| **`RegistrationSensorTimeout`** | $6$ ore            | L'intervallo dopo il quale un sensore registrato ma inattivo può essere rimosso dal sistema.                                                                                                         |

### F\. Parametri di Logging e Health Check

| Variabile                 | Descrizione                                                                                                     | Default                                       |
|:--------------------------|:----------------------------------------------------------------------------------------------------------------|:----------------------------------------------|
| **`HEALTHZ_SERVER`**      | Flag booleano per attivare un server HTTP semplice che risponde allo stato di salute del servizio (`/healthz`). | `false`                                       |
| **`HEALTHZ_SERVER_PORT`** | Porta su cui il server Health Check si mette in ascolto.                                                        | `8080`                                        |
| **`LOG_LEVEL`**           | Livello di dettaglio per l'output del logger.                                                                   | `error` (Default), `warning`, `info`, `debug` |

-----

## Deploy in Locale dell'Edge Hub

La distribuzione locale dell'Edge Hub è gestita dal file [`edge-hub.yaml`](../../deploy/compose/edge-hub.yaml) che definisce sia l'applicazione Edge Hub principale che i suoi microservizi di supporto, in particolare la cache Redis.

### Struttura di Docker Compose

Il file Compose definisce l'Edge Hub non come un unico servizio, ma come una serie di microservizi cooperanti che condividono la stessa immagine ma eseguono ruoli diversi:

* **Servizi Principali:** `zone-hub-filter`, `zone-hub-cleaner`, `zone-hub-aggregator`, `zone-hub-configurator`. Questi servizi sono scalati orizzontalmente per garantire tolleranza ai guasti e parallelismo nell'elaborazione.
* **Dipendenza:** `zone-hub-redis-cache`, che fornisce la cache di stato condiviso necessaria a tutti i servizi Hub.
* **Rete:** Viene creata una rete *bridge* isolata per il traffico con la cache: `zone-hub-${EDGE_ZONE}-bridge`.

Il file sfrutta le estensioni YAML per ridurre la ridondanza, ereditando le configurazioni base e differenziando solo le variabili d'ambiente critiche:

```yaml
# --- environment base per tutti i zone hub ---
x-zone-hub-env: &zone-hub-env
  EDGE_MACROZONE: "${EDGE_MACROZONE}"
  EDGE_ZONE: "${EDGE_ZONE}"
  MQTT_BROKER_ADDRESS: "${EDGE_ZONE}.${EDGE_MACROZONE}.mqtt-broker.${REGION}.sensor-continuum.local"
  REDIS_ADDRESS: "zone-hub-${EDGE_ZONE}-cache"
  OPERATION_MODE: "loop"
  HEALTHZ_SERVER: "true"
  HEALTHZ_SERVER_PORT: "8080"

# --- blocco base per tutti i zone hub ---
x-zone-hub-base: &zone-hub-base
  image: fmasci/sc-edge-hub:latest
  build:
    context: ../..
    dockerfile: deploy/docker/edge-hub.Dockerfile
  healthcheck:
    test: [ "CMD", "curl", "-f", "http://localhost:8080/healthz" ]
    interval: 60s
    timeout: 30s
    retries: 10
  restart: unless-stopped
  networks:
    - zone-hub-bridge
  depends_on:
    zone-hub-redis-cache:
      condition: service_healthy
```

### Variabili d'Ambiente Obbligatorie

Il corretto funzionamento del file Compose dipende dalla risoluzione dinamica di quattro variabili essenziali, utilizzate per l'identificazione, l'indirizzamento e la creazione di risorse:

| Variabile            | Necessità                     | Utilizzo nel Compose                                                                                      |
|:---------------------|:------------------------------|:----------------------------------------------------------------------------------------------------------|
| **`EDGE_MACROZONE`** | Naming e Indirizzamento MQTT  | Definisce il nome del container e l'indirizzo DNS del broker.                                             |
| **`EDGE_ZONE`**      | Naming e Naming delle Risorse | Cruciale per nominare container, volumi, reti e per l'indirizzo DNS del broker.                           |
| **`REGION`**         | Indirizzamento MQTT           | Utilizzata per completare l'hostname del broker MQTT.                                                     |
| **`REDIS_PORT`**     | Mappatura Porta               | Definisce la porta host per accedere a Redis dall'esterno (default `6379` se non specificato altrimenti). |

Senza queste variabili, Docker Compose non sarà in grado di popolare correttamente le estensioni YAML e i nomi delle risorse, portando al fallimento del deployment.

### Preparazione e Esecuzione del Deploy

Il deployment in locale richiede due passaggi fondamentali: la creazione del volume e l'avvio tramite Compose.

#### 1\. Creazione del Volume Persistente

Poiché il volume `zone-hub-cache` è definito come `external: true` nel file Compose, deve essere creato **manualmente** prima dell'avvio per garantirne la persistenza e il riutilizzo tra diversi cicli di deploy:

**Esempio di creazione del volume per la zona `TorVergata`:**

```bash
docker volume create zone-hub-TorVergata-cache-data
```

#### 2\. Avvio con Docker Compose

Dopo aver creato il volume, l'avvio viene eseguito tramite il comando standard di Docker Compose:

```bash
# Esempio di avvio
docker compose -f edge-hub.yaml up -d
```

#### 3\. Risoluzione del Nome Host del Broker

Il template utilizza un indirizzo complesso:

```yaml
# --- environment base per tutti i zone hub ---
x-zone-hub-env: &zone-hub-env
    # ... altre variabili ...
    MQTT_BROKER_ADDRESS: "${EDGE_ZONE}.${EDGE_MACROZONE}.mqtt-broker.${REGION}.sensor-continuum.local"
    # ... altre variabili ...
```

Per risolvere questo nome di dominio e dirigerlo correttamente verso l'host del broker MQTT, si hanno due opzioni in un ambiente locale:

1.  **Utilizzo di `extra_hosts`:** Configurare la sezione `extra_hosts` nel file Docker Compose, mappando l'indirizzo del broker all'IP appropriato.
2.  **Modifica del File `/etc/hosts`:** Se necessario, è possibile modificare il file `/etc/hosts` del sistema operativo host per puntare il record DNS utilizzato direttamente all'indirizzo IP del container o del servizio che ospita il broker.


> **⚠️ NOTA OPERATIVA**
>
> **File d'Ambiente:**
> Per standardizzare il deployment, è cruciale definire le variabili d'ambiente (almeno `EDGE_MACROZONE`, `EDGE_ZONE` e `REGION`) in un file separato (es. `TorVergata.env`). Il file Compose le utilizza per risolvere tutti i nomi dei container, i nomi di rete e l'indirizzo del broker MQTT.
>
> **Esempio di caricamento con un file .env:**
>
> ```bash
> docker compose -f edge-hub.yaml --env-file TorVergata.env up -d
> ```
>
> **Nomi di Progetto per la Gestione Layer:**
> Poiché al livello Edge coesistono diversi servizi (Edge Hub e Sensor Agents), è cruciale nominare esplicitamente il progetto Docker Compose. Questo facilita l'identificazione, l'avvio e l'arresto di tutti i container che fanno parte di quella specifica Edge Zone.
>
> Si raccomanda di utilizzare il flag **`-p`** o **`--project-name`** con un nome che rifletta la regione, la macrozona e la zona (es. `Lazio_RomaMacro_TorVergata`).
>
> **Esempio di Deploy Completo:**
>
> ```bash
> docker compose -f edge-hub.yaml --env-file .env -p Lazio_RomaMacro_TorVergata up -d
> ```

-----

## Deploy su AWS dell'Edge Hub

L'Edge Hub (e le sue dipendenze, inclusa la cache Redis) è distribuito sull'istanza EC2 dedicata alla zona tramite Docker Compose, con l'orchestrazione gestita da AWS CloudFormation e lo script Bash [`deploy_zone.sh`](../../deploy/scripts/deploy_zone.sh).

### ⚠️ Prerequisiti di Deployment

La fase di Deploy su AWS, gestita dallo script [`deploy_zone.sh`](../../deploy/scripts/deploy_zone.sh), è l'ultima di una sequenza di provisioning che stabilisce l'infrastruttura di rete a livello regionale e di macrozona.

**Prima di eseguire lo script di deployment della Zona, è indispensabile che le seguenti risorse AWS siano già state create e configurate attraverso i rispettivi script di provisioning di Livello Superiore ([`deploy_region.sh`](../../deploy/scripts/deploy_region.sh) e [`deploy_macrozone.sh`](../../deploy/scripts/deploy_macrozone.sh)) o manualmente:**

| Risorsa                          | Livello di Creazione | Variabile nello script | Riferimento Logico                                   |
|:---------------------------------|:---------------------|:-----------------------|:-----------------------------------------------------|
| **Virtual Private Cloud (VPC)**  | Regionale            | `$VPC_ID`              | Identificata da **`$REGION-vpc`**                    |
| **Subnet Pubblica**              | Macrozona            | `$SUBNET_ID`           | Identificata da **`$REGION-$MACROZONE-subnet`**      |
| **Security Group**               | Macrozona            | `$SECURITY_GROUP_ID`   | Identificata da **`$REGION-$MACROZONE-sg`**          |
| **Route Table Pubblica**         | Regionale            | `$ROUTE_TABLE_ID`      | Identificata da **`$REGION-vpc-public-rt`**          |
| **Route 53 Hosted Zone Privata** | Regionale            | `$HOSTED_ZONE_ID`      | Identificata da **`$REGION.sensor-continuum.local`** |

Lo script [`deploy_zone.sh`](../../deploy/scripts/deploy_zone.sh) non crea le risorse di rete; piuttosto, le cerca e le recupera tramite le funzioni presenti nel file [`utils.sh`](../../deploy/scripts/utils.sh): `find_vpc_id`, `find_subnet_id`, etc. Se queste risorse non esistono o non corrispondono ai tag di denominazione attesi, lo script fallirà, non potendo lanciare lo stack CloudFormation.

#### Sequenza di Deployment Rigorosa

1.  **Deploy Livello Regionale:** Esecuzione dello script per creare la VPC, la Hosted Zone e la Route Table Pubblica.
2.  **Deploy Livello Macrozona:** Esecuzione dello script per creare la Subnet e il Security Group associati.
3.  **Deploy Livello Zona/Edge:** Esecuzione dello script [`deploy_zone.sh`](../../deploy/scripts/deploy_zone.sh) per installare l'istanza EC2, i container e configurare i record DNS specifici che puntano dell'istanza contenente il broker MQTT associato alla zona.

Assicurarsi che la catena di provisioning dell'infrastruttura AWS sia stata completata correttamente è la condizione necessaria per un deploy di successo dell'Edge Zone.

### 1\. Caricamento Asset su S3

Prima di eseguire il deployment, tutti gli asset necessari (script di installazione, file Docker Compose, file di servizio Systemd e script di analisi) devono essere caricati in un Bucket S3 dedicato.

Questo processo è documentato in [`setup_bucket.md`](./setup_bucket.md) e gestito dallo script [`setup_bucket.sh`](../../deploy/scripts/setup_bucket.sh).

### 2\. Deploy Edge Zone

Il deployment dell'intera zona Edge (che comprende l'istanza EC2, la configurazione di Docker e l'Edge Hub) è orchestrato dallo script [`deploy_zone.sh`](../../deploy/scripts/deploy_zone.sh) che utilizza il template [`zone/services.yaml`](../../deploy/cloudformation/zone/services.yaml).

#### Funzionamento del Template CloudFormation

Il template CloudFormation esegue le seguenti operazioni cruciali per il layer Edge:

* **Creazione EC2 Instance**: Avvia un'istanza EC2 (di default Amazon Linux 2) che ospiterà i container Docker.
    * Viene associato un IAM Role che fornisce i permessi necessari per accedere a S3 e Route53.
* **Configurazione con `UserData`**: Il blocco `UserData` (eseguito al primo boot) installa l'AWS CLI, scarica gli *init scripts* (come `docker-install.sh`) per installare Docker e Docker Compose, scarica il file `.env` specifico per la zona da S3 e infine esegue lo script di deployment [`deploy_edge_services.sh`](../../deploy/scripts/deploy/deploy_edge_services.sh).
* **Configurazione DNS privato**: Crea un set di Record CNAME all'interno della Hosted Zone privata relativa alla regione. Questi record sono essenziali per risolvere correttamente i nomi di dominio complessi dei broker MQTT verso l'IP privato dell'istanza EC2.

#### Utilizzo dello Script di Deployment

Lo script di avvio CloudFormation accetta parametri posizionali obbligatori per definire la regione logica del sistema, la macrozona e la zona.

##### Sintassi ed Esempi

**Sintassi Completa:**

```bash
./deploy_zone.sh region-name macrozone-name zone-name [opzioni]
```

**Esempio di Deploy Standard su AWS:**

```bash
./deploy_zone.sh Lazio RomaMacro TorVergata --aws-region us-east-1 --instance-type t3.small
```

#### Dettagli Operativi del Deployment Script in EC2

Lo script Bash [`deploy_edge_services.sh`](../../deploy/scripts/deploy/deploy_edge_services.sh) viene eseguito all'interno dell'istanza EC2 tramite `UserData` ed è responsabile dell'avvio dei servizi che compongono l'Edge Hub.

##### Sequenza di Avvio dei Servizi

1.  **Caricamento Variabili**: Carica le variabili d'ambiente dal file **`.env`** precedentemente scaricato da S3.
2.  **Download Docker Compose**: Scarica da S3 la configurazione dell'hub [`edge-hub.yaml`](../../deploy/compose/edge-hub.yaml).
3.  **Avvio Sensor Agents**: Utilizza `docker-compose -p edge-hub` per avviare il'hub, isolando i microservizi che lo compongono in un progetto separato per facilitare la gestione.

##### Gestione Operativa e Resilienza

Lo script [`deploy_edge_services.sh`](../../deploy/scripts/deploy/deploy_edge_services.sh) integra diverse funzionalità per la gestione e la simulazione di un ambiente Edge robusto:

* **Simulatore di Ritardi**: Scarica ed esegue lo script [`init-delay.sh`](../../deploy/scripts/inits/init-delay.sh) per simulare condizioni di rete reali. Applica una latenza di rete configurabile (di default impostata a `${NETWORK_DELAY:-200ms}`) sull'interfaccia di rete dell'istanza EC2.
* **Servizio Systemd**: Configura il servizio [`sc-deploy.service`](../../deploy/scripts/services/sc-deploy.service.template) per eseguire automaticamente lo script di deployment all'avvio del sistema. Ciò garantisce che i servizi vengano ripristinati correttamente dopo un riavvio dell'istanza EC2.
* **Script di Analisi**: Scarica lo script [`analyze_failure.sh`](../../deploy/scripts/performance/analyze_failure.sh). Questo strumento è progettato per analizzare i log dei container, confrontando i messaggi ricevuti con i messaggi attesi e calcolando parametri di performance chiave come il *Missing Rate* e l'*errore nel rilevamento degli Outlier*.
