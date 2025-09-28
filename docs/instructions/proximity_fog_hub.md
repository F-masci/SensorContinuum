# Istruzioni per il Proximity Fog Hub

Il **Proximity Fog Hub** rappresenta il secondo livello di calcolo nel **Compute Continuum**, posizionandosi al di sopra degli Edge Hub e operando a livello di **Macrozona** logica.
Il suo ruolo principale è quello di ricevere i dati aggregati e filtrati da tutti gli Edge Hub di una data macrozona, costruire i dati statistici e inviare i risultati al livello superiore (Intermediate Fog Hub).

## Architettura e Funzionalità ad Alto Livello

1.  **Comunicazione Ingressa (MQTT):** L'Hub consuma i dati e i messaggi di controllo inviati dagli Edge Hub sulla sua macrozona utilizzando il protocollo **MQTT**. Sfrutta le *Shared Subscriptions* (`$share/proximity-fog-hub...`) per garantire che i messaggi in ingresso siano distribuiti tra le istanze attive, supportando così l'alta disponibilità.
2.  **Stato e Resilienza (PostgreSQL):** A differenza dell'Edge Hub, che usa Redis, il Proximity Fog Hub necessita di un database relazionale persistente come **PostgreSQL** per tre scopi principali:
    * **Memorizzazione Dati:** Archiviazione dei dati ricevuti e delle statistiche aggregate.
    * **Aggregazione a Lungo Termine:** Esegue l'aggregazione finale dei dati (ogni **15 minuti**) prima della trasmissione.
    * **Outbox Pattern:** Utilizza il database come **Outbox** per garantire la resilienza. I messaggi destinati al livello superiore (Kafka) vengono prima scritti nel database, marcati come *pending*, e solo dopo un invio riuscito vengono marcati come *sent* dal servizio Dispatcher.
3.  **Comunicazione in Uscita (Kafka):** I dati finali (dati in tempo reale, statistiche aggregate e heartbeat) vengono inviati al livello superiore (Intermediate Fog Hub) tramite il sistema di *message queue* **Kafka**, garantendo un throughput elevato e scalabilità.

---

## Variabili d'Ambiente del Proximity Fog Hub

La configurazione è gestita tramite variabili d'ambiente, suddivise per funzione.

### 1. Identità e Modalità Operative

Queste variabili definiscono l'identità gerarchica del servizio e il suo ruolo specifico.

| Variabile            | Descrizione                                                              | Default / Valori Ammessi                                                                                                           | Stato        |
|:---------------------|:-------------------------------------------------------------------------|:-----------------------------------------------------------------------------------------------------------------------------------|:-------------|
| **`EDGE_MACROZONE`** | **Obbligatoria**. Identifica la macrozona logica a cui appartiene l'Hub. | Nessun Default                                                                                                                     | Obbligatoria |
| **`HUB_ID`**         | Identificatore univoco dell'istanza Hub.                                 | UUID Generato                                                                                                                      | Opzionale    |
| **`OPERATION_MODE`** | Controlla il ciclo di vita del servizio.                                 | `loop` / `once`                                                                                                                    | Opzionale    |
| **`SERVICE_MODE`**   | Definisce il ruolo specifico dell'istanza.                               | `proximity_hub` (Tutte le funzionalità) / `proximity_hub_aggregator` / `proximity_hub_dispatcher` / `proximity_hub_cleaner` / ecc. | Opzionale    |

---

### 2. Connessione MQTT (Ingresso: Edge Hub $\to$ Proximity Hub)

Queste variabili configurano l'Hub come *consumer* dei dati provenienti dagli Edge Hub.

| Variabile                            | Descrizione                                                | Default     |
|:-------------------------------------|:-----------------------------------------------------------|:------------|
| **`MQTT_BROKER_PROTOCOL`**           | Protocollo di connessione al broker MQTT.                  | `tcp`       |
| **`MQTT_BROKER_ADDRESS`**            | Indirizzo IP/Hostname del broker MQTT.                     | `mosquitto` |
| **`MQTT_BROKER_PORT`**               | Porta di connessione del broker MQTT.                      | `1883`      |
| **`MQTT_MAX_RECONNECTION_INTERVAL`** | Intervallo massimo tra i tentativi di riconnessione (sec). | $10$        |
| **`MQTT_MAX_RECONNECTION_ATTEMPTS`** | Numero massimo di tentativi di riconnessione.              | $10$        |
| **`MQTT_MESSAGE_PUBLISH_TIMEOUT`**   | Timeout per l'invio dei messaggi (sec).                    | $5$         |

**Topic di Sottoscrizione (Derivati):**
* **Dati Filtrati:** `$share/proximity-fog-hub_<MACROZONE>/filtered-data/<MACROZONE>` (usa Shared Subscription).
* **Configurazione:** `configuration/hub/<MACROZONE>`
* **Heartbeat:** `heartbeat/<MACROZONE>`

---

### 3. Connessione Kafka (Uscita: Proximity Hub $\to$ Intermediate Fog Hub)

Queste variabili configurano l'Hub come *producer* verso il livello superiore.

| Variabile                                            | Descrizione                                    | Default                              |
|:-----------------------------------------------------|:-----------------------------------------------|:-------------------------------------|
| **`KAFKA_BROKER_ADDRESS`**                           | Indirizzo del broker Kafka.                    | `kafka`                              |
| **`KAFKA_BROKER_PORT`**                              | Porta del broker Kafka.                        | `9092`                               |
| **`KAFKA_PUBLISH_TIMEOUT`**                          | Timeout per la pubblicazione su Kafka (sec).   | $5$                                  |
| **`KAFKA_PROXIMITY_FOG_HUB_REALTIME_DATA_TOPIC`**    | Topic per i dati in tempo reale.               | `proximity-fog-hub-realtime-data`    |
| **`KAFKA_PROXIMITY_FOG_HUB_AGGREGATED_STATS_TOPIC`** | Topic per le statistiche aggregate.            | `proximity-fog-hub-aggregated-stats` |

---

### 4. Configurazione PostgreSQL (Cache Locale)

Queste variabili sono necessarie per la connessione al database locale utilizzato per la cache persistente e l'Outbox pattern.

| Variabile               | Descrizione                      | Default           |
|:------------------------|:---------------------------------|:------------------|
| **`POSTGRES_USER`**     | Nome utente del database.        | `admin`           |
| **`POSTGRES_PASSWORD`** | Password del database.           | `adminpass`       |
| **`POSTGRES_HOST`**     | Indirizzo del server PostgreSQL. | `localhost`       |
| **`POSTGRES_PORT`**     | Porta del server PostgreSQL.     | `5432`            |
| **`POSTGRES_DATABASE`** | Nome del database.               | `sensorcontinuum` |

---

### 5. Configurazioni di Elaborazione

Sebbene siano costanti nel codice Go, governano l'intervallo di esecuzione dei microservizi critici.

| Parametro Logico            | Valore costante | Funzione Corrispondente                                                                   |
|:----------------------------|:----------------|:------------------------------------------------------------------------------------------|
| **Intervallo Aggregazione** | $15$ minuti     | Frequenza con cui l'Aggregator esegue l'aggregazione finale dei dati nel DB.              |
| **Intervallo Dispatcher**   | $2$ minuti      | Frequenza con cui il Dispatcher controlla la tabella Outbox per inviare messaggi a Kafka. |
| **Dimensione Batch Outbox** | $50$ messaggi   | Numero di messaggi tentati di inviare a Kafka in ogni ciclo del Dispatcher.               |
| **Intervallo Cleaner**      | $5$ minuti      | Frequenza con cui il Cleaner elimina dal DB i messaggi inviati con successo.              |

---

### 6. Health Check Server

| Variabile                 | Descrizione                                                                           | Default |
|:--------------------------|:--------------------------------------------------------------------------------------|:--------|
| **`HEALTHZ_SERVER`**      | Flag per attivare il server HTTP per il controllo dello stato di salute (`/healthz`). | `false` |
| **`HEALTHZ_SERVER_PORT`** | Porta su cui il server Health Check si mette in ascolto.                              | `8080`  |

---

## Deploy del Proximity Hub in Locale con Docker Compose

Il **Proximity Hub** è distribuito come una suite di microservizi in Go che garantisce. L'intera architettura è gestita tramite Docker Compose e si appoggia a una cache persistente **PostgreSQL/TimescaleDB**.

-----

### 1\. Struttura di Docker Compose

Il file Compose `proximity-fog-hub.yaml` definisce l'Hub non come un'unica entità, ma come una **serie di microservizi cooperanti** (basati sull'immagine `fmasci/sc-proximity-fog-hub:latest`) che eseguono ruoli specifici. Tutti i servizi dell'Hub sono scalati a due istanze per l'alta disponibilità.

| Servizio           | Istanze | Modalità                      | Funzione Principale                                                    |
|:-------------------|:--------|:------------------------------|:-----------------------------------------------------------------------|
| **POSTGRES CACHE** | 1       | (N/A)                         | Fornisce la cache persistente (TimescaleDB).                           |
| **LOCAL CACHE**    | 2       | `proximity_hub_local_cache`   | Ingestione da MQTT e persistenza idempotente nel DB.                   |
| **CONFIGURATOR**   | 2       | `proximity_hub_configuration` | Proxy per i messaggi di registrazione/configurazione (MQTT -\> Kafka). |
| **HEARTBEAT**      | 2       | `proximity_hub_heartbeat`     | Proxy per i messaggi di heartbeat (MQTT -\> Kafka).                    |
| **AGGREGATOR**     | 2       | `proximity_hub_aggregator`    | Calcolo delle statistiche (max, min, avg) su intervalli di 15 minuti.  |
| **DISPATCHER**     | 2       | `proximity_hub_dispatcher`    | Implementa l'**Outbox Pattern**; inoltro affidabile dal DB a Kafka.    |
| **CLEANER**        | 2       | `proximity_hub_cleaner`       | Manutenzione del DB; elimina i record "sent" più vecchi.               |

Il file sfrutta le estensioni YAML (`x-macrozone-hub-base`, `x-macrozone-hub-env`) per ereditare configurazioni comuni e garantire la coerenza tra le istanze.

-----

### 2\. Variabili d'Ambiente Obbligatorie

Il corretto funzionamento del file Compose dipende dalla risoluzione dinamica di tre variabili essenziali, utilizzate per l'identificazione e l'indirizzamento tra i livelli:

| Variabile            | Necessità                            | Utilizzo nel Compose                                                        |
|:---------------------|:-------------------------------------|:----------------------------------------------------------------------------|
| **`EDGE_MACROZONE`** | Naming, Indirizzamento MQTT/POSTGRES | Definisce il nome dei container e l'hostname della cache (`POSTGRES_HOST`). |
| **`REGION`**         | Indirizzamento Broker                | Cruciale per la risoluzione DNS degli indirizzi **Kafka** e **MQTT**.       |
| **`POSTGRES_PORT`**  | Mappatura Porta                      | Definisce la porta host per l'accesso a PostgreSQL (default `5432`).        |

```yaml
# --- environment base per tutti i macrozone hub ---
x-macrozone-hub-env: &macrozone-hub-env
    EDGE_MACROZONE: "${EDGE_MACROZONE}"
    KAFKA_BROKER_ADDRESS: "kafka-broker.${REGION}.sensor-continuum.local"
    KAFKA_BROKER_PORT: "9094"
    MQTT_BROKER_ADDRESS: "${EDGE_MACROZONE}.mqtt-broker.${REGION}.sensor-continuum.local"
    POSTGRES_HOST: "macrozone-hub-${EDGE_MACROZONE}-cache"
    POSTGRES_PORT: "5432"
    OPERATION_MODE: "loop"
    HEALTHZ_SERVER: "true"
    HEALTHZ_SERVER_PORT: "8080"
```

---

### 3\. Preparazione e Esecuzione del Deploy

Il deployment in locale richiede due passaggi fondamentali: la creazione del volume e l'avvio tramite Compose.

#### A. Creazione del Volume Persistente

Poiché il volume `macrozone-cache-data` è definito come **`external: true`** nel file Compose, deve essere creato **manualmente** prima dell'avvio per garantire la persistenza dei dati PostgreSQL:

**Esempio di creazione del volume (assumendo `$EDGE_MACROZONE=RomaMacro`):**

```bash
docker volume create macrozone-hub-RomaMacro-cache-data
```

```bash
# Avvio di tutti i microservizi del Proximity Hub
docker compose -f proximity-fog-hub.yaml up -d
```

#### B. Risoluzione del Nome Host dei Broker

Il Proximity Hub deve connettersi a due broker esterni. Per indirizzarli correttamente, vengono utilizzati hostname complessi che richiedono una risoluzione DNS locale:

1.  **Broker MQTT (Livello Edge):** `${EDGE_MACROZONE}.mqtt-broker.${REGION}.sensor-continuum.local`
2.  **Broker Kafka (Livello Intermediate):** `kafka-broker.${REGION}.sensor-continuum.local`

In un ambiente Docker locale, si hanno le seguenti opzioni per la risoluzione:

* **Utilizzo di `extra_hosts`:** Configurare la sezione `extra_hosts` nel file Docker Compose, mappando l'indirizzo del broker all'IP appropriato.
* **Modifica del File `/etc/hosts`:** Modificare il file **`/etc/hosts`** del sistema operativo host per puntare i record DNS utilizzati direttamente all'indirizzo IP del nodo che ospita il broker.

#### C. Avvio con Docker Compose

Dopo aver creato il volume e definito le variabili d'ambiente (nel file `.env`), l'avvio viene eseguito tramite il comando standard di Docker Compose.

> **⚠️ NOTA OPERATIVA**
>
> **Dipendenza e Ordine di Avvio:**
> Tutti i microservizi dell'Hub dipendono dallo stato **`service_healthy`** del container **`macrozone-hub-postgres-cache`**. I servizi Go attendono che il database sia completamente avviato, inizializzato con lo schema (`init-proximity-cache.sql`) e pronto per le connessioni prima di tentare di avviarsi.
>
> **File d'Ambiente:**
> Per standardizzare il deployment, è cruciale definire le variabili d'ambiente (almeno `EDGE_MACROZONE`, `EDGE_ZONE` e `REGION`) in un file separato (es. **`.env`**). Il file Compose le utilizza per risolvere tutti i nomi dei container, i nomi di rete e l'indirizzo del broker MQTT.
>
> **Esempio di caricamento con un file .env:**
>
> ```bash
> docker compose -f proximity-fog-hub.yaml --env-file RomaMacro.env up -d
> ```
> 
> **Nomi di Progetto per la Gestione Layer:**
> Si raccomanda di utilizzare il flag **`-p`** o **`--project-name`** per identificare chiaramente i servizi del Proximity Hub (es. `Lazio_RomaMacro_Hub`), facilitando l'arresto e la gestione dell'intero layer.
>
> **Esempio di Deploy Completo:**
>
> ```bash
> docker compose -f proximity-fog-hub.yaml --env-file .env -p Lazio_RomaMacro_Hub up -d
> ```

---

### 4\. Deployment del Broker MQTT della Macrozona

Contestualmente al deployment del Proximity Hub, è necessario installare un broker MQTT per ricevere i dati dai nodi Edge Hub (livello inferiore).

> **⚠️ NOTA ARCHITETTURALE**
> 
> È possibile prevedere un broker MQTT per ogni singola zona (comunicazione Sensor -\> Hub) o un **singolo broker per l'intera macrozona** (comunicazione Hub -\> Hub). Per semplicità del deployment locale, la documentazione si concentra sulla seconda casistica, con un unico broker che gestisce l'intera macrozona.

#### A. Template Docker Compose per il Broker MQTT

Il broker può essere deployato utilizzando un'immagine Mosquitto standard o, come in questo caso, un'immagine custom (`fmasci/sc-mqtt-broker:latest`) basata su `eclipse-mosquitto:latest` con l'aggiunta di file di configurazione specifici.

Di seguito è riportato il template Compose per il broker:

```yaml
services:
  mqtt-broker:
    image: fmasci/sc-mqtt-broker:latest
    build:
      context: ../..
      dockerfile: deploy/docker/mosquitto.Dockerfile
    container_name: mqtt-broker-${EDGE_MACROZONE}-01
    hostname: mqtt-broker-${EDGE_MACROZONE}-01
    ports:
      - "1883:1883"
    healthcheck:
      test: [ "CMD-SHELL", "mosquitto_sub -h localhost -t '$$SYS/broker/version' -C 1 -W 2 || exit 1" ]
      interval: 10s
      timeout: 5s
      retries: 3
    restart: unless-stopped

networks:
  mqtt-broker-bridge:
    name: mqtt-broker-${EDGE_MACROZONE}-bridge
    driver: bridge
```

#### B. Avvio del Broker

Se il broker non è già attivo, è necessario avviare il file Docker Compose specifico (`deploy/compose/mqtt-broker.yml`) in aggiunta ai servizi Hub.

```bash
# Assumendo di utilizzare il file mqtt-broker.yml
docker compose -f deploy/compose/mqtt-broker.yml --env-file .env -p Lazio_RomaMacro_Broker up -d
```

Una volta avviato, il container sarà accessibile pubblicamente sulla porta `1883`, consentendo ai microservizi Proximity Hub di sottoscrivere i dati dal livello Edge.

---

## Deploy su AWS del Proximity Hub

Il **Proximity Hub**viene distribuito su un'istanza **EC2** dedicata alla macrozona tramite **Docker Compose**, con l'orchestrazione gestita da **AWS CloudFormation** attraverso lo script Bash **`deploy_macrozone.sh`**.

-----

### ⚠️ Prerequisiti di Deployment (Infrastruttura di Rete Regionale)

Lo script **`deploy_macrozone.sh`** ha un doppio ruolo: prima **crea** le risorse di rete specifiche per la Macrozona e poi esegue il deployment dei servizi.

**Risorse Prerequisito (Create a Livello Regionale e Ricercate):**

Lo script recupera (tramite `find_vpc_id`, `find_hosted_zone_id`, etc.) le risorse create precedentemente dal deploy regionale:

| Risorsa AWS                      | Livello di Creazione | Riferimento Logico (Tag Name)    |
|:---------------------------------|:---------------------|:---------------------------------|
| **Virtual Private Cloud (VPC)**  | Regionale            | `$REGION-vpc`                    |
| **Route 53 Hosted Zone Privata** | Regionale            | `$REGION.sensor-continuum.local` |
| **Route Table Pubblica**         | Regionale            | `$REGION-vpc-public-rt`          |

**Creazione Contesto Macrozona (Eseguita da `deploy_macrozone.sh`):**

Prima di procedere con l'istanza EC2, lo script lancia il deployment del template **`Subnet.yaml`** per creare:

* **Subnet Pubblica** (`$REGION-$MACROZONE-subnet`): Viene calcolato e assegnato un CIDR disponibile all'interno della VPC.
* **Security Group** (`$REGION-$MACROZONE-sg`): Configurato per permettere tutto il traffico in ingresso e in uscita.

----

### 1\. Preparazione: Caricamento Asset su S3

Prima di eseguire il deployment, tutti gli asset necessari (script di installazione, file Docker Compose, file di servizio Systemd e script di analisi) devono essere caricati in un **Bucket S3** dedicato.

Questo processo è documentato in [`setup_bucket.md`](./setup_bucket.md) e gestito dallo script **`deploy/scripts/setup_bucket.sh`**.

-----

### 2\. Deploy Macrozona Services tramite CloudFormation

Dopo aver creato o verificato le risorse di rete, lo script `deploy_macrozone.sh` lancia lo stack dei servizi tramite il template **`services.yaml`**.

#### Funzionamento del Template CloudFormation

Il template `services.yaml` esegue le seguenti operazioni:

* **Creazione EC2 Instance (`ServicesInstance`)**: Avvia l'istanza EC2 che ospiterà l'intera suite **Proximity Hub** e il **Broker MQTT** locale.
* **Configurazione DNS (Route53 - Record A)**: Crea un **Record A** nella Hosted Zone privata, **`$MACROZONE.mqtt-broker.$REGION.sensor-continuum.local`**, che punta all'IP privato dell'istanza EC2. Questo hostname viene utilizzato da tutti gli Edge Hub per inviare i dati a questo nodo Macrozona.
* **Installazione via `UserData`**: Il blocco `UserData` installa Docker/Docker Compose ed esegue lo script **`deploy_proximity_services.sh`**.

-----

### 3\. Dettagli Operativi del Deployment Script in EC2

Lo script Bash **`deploy_proximity_services.sh`**, eseguito all'interno dell'istanza EC2, è responsabile dell'installazione e dell'avvio di **due componenti chiave**: la suite di microservizi **Proximity Hub** e il **Broker MQTT** locale.

#### A. Deployment del Broker MQTT Locale

Il Proximity Hub opera come un punto di raccolta dati per i livelli Edge. Per ricevere i dati, deploya un broker MQTT sulla stessa istanza EC2:

* **File Compose:** Il broker è definito nel file **`mqtt-broker.yaml`**.
* **Immagine:** Viene utilizzata l'immagine `fmasci/sc-mqtt-broker:latest`, basata su Mosquitto con configurazioni aggiuntive.
* **Nome Host:** Il container è accessibile internamente come **`mqtt-broker-${EDGE_MACROZONE}-01`**.
* **Funzione:** Questo broker riceve il traffico dai nodi Edge Hub e inoltra i dati ai microservizi **Local Cache** del Proximity Hub.

#### B. Sequenza di Avvio dell'Hub (Microservizi e Cache)

1.  **Cache Persistente:** Lo script verifica e crea il volume Docker **`macrozone-hub-${EDGE_MACROZONE}-cache-data`**.
2.  **Configurazione PostgreSQL:** Viene avviato il container **`macrozone-hub-postgres-cache`** isolatamente, e lo script attende che sia pronto per eseguire un'ottimizzazione critica: l'aumento del limite di connessioni per supportare tutti i microservizi Hub e i processi di aggregazione.
3.  **Avvio Completo:** Un secondo `docker compose up -d` avvia i microservizi **Local Cache**, **Dispatcher**, **Aggregator**, **Cleaner**, **Configurator**, e **Heartbeat**, che ora possono connettersi in modo affidabile alla cache PostgreSQL configurata.

#### Gestione Operativa e Resilienza

| Funzionalità               | Dettaglio Operativo                                                                                                                                                                                   |
|:---------------------------|:------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Configurazione Latenza** | Esecuzione dello script **`init-delay.sh`** per simulare latenza, jitter e packet loss sull'interfaccia di rete dell'istanza EC2, utilizzando le variabili d'ambiente (es. `${NETWORK_DELAY:-30ms}`). |
| **Servizio Systemd**       | Configurazione del servizio **`sc-deploy-services.service`** per rieseguire automaticamente lo script di deployment all'avvio del sistema, garantendo la resilienza dell'Hub e del Broker MQTT.       |