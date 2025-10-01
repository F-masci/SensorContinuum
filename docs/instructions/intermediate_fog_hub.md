# Istruzioni per l'Intermediate Fog Hub

L'Intermediate Fog Hub rappresenta il livello più alto del piano Fog nel sistema Sensor Continuum. Operando a livello di Regione, funge da punto di aggregazione finale e di persistenza per tutti i dati consolidati provenienti dai livelli sottostanti.

Il suo funzionamento ad alto livello include:
* **Ingestione Dati:** L'Hub funge da consumer principale sui topic Kafka, ricevendo tutti i flussi di dati (tempo reale, statistiche aggregate, configurazione e heartbeat) inoltrati dai Proximity Fog Hub. Il `KafkaGroupId` fisso a `intermediate-fog-hub` assicura il bilanciamento del carico tra le istanze.
* **Archiviazione Primaria:** Utilizza un database relazionale persistente **PostgreSQL/TimescaleDB**, denominato *Sensor DB*, come strato di persistenza per tutte le serie storiche aggregate a livello regionale.
* **Aggregazione Finale:** Il microservizio *Aggregator* esegue l'aggregazione statistica finale a intervalli regolari costanti, consolidando i dati per l'archiviazione storica a lungo termine.
* **Gestione Metadati:** Si connette a un database di metadati dedicato, il *Metadata DB*, per gestire la configurazione e lo stato globale del sistema a livello regionale.

-----

## Variabili d'Ambiente dell'Intermediate Fog Hub

La configurazione del servizio è gestita interamente tramite variabili d'ambiente.

### A\. Identità e Modalità Operative

Queste variabili definiscono l'identità del servizio e il suo ruolo specifico.

| Variabile            | Descrizione                                          | Default / Valori Ammessi                                                                                                                                                                                   | Stato     |
|:---------------------|:-----------------------------------------------------|:-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:----------|
| **`HUB_ID`**         | Identificatore univoco dell'istanza Hub.             | UUID Generato                                                                                                                                                                                              | Opzionale |
| **`OPERATION_MODE`** | Controlla il ciclo di vita del servizio.             | `loop` (ciclo continuo) / `once`                                                                                                                                                                           | Opzionale |
| **`SERVICE_MODE`**   | Definisce il ruolo operativo specifico dell'istanza. | `intermediate_hub` (Tutte le funzionalità) / `intermediate_hub_realtime` / `intermediate_hub_statistics` / `intermediate_hub_configuration` / `intermediate_hub_heartbeat` / `intermediate_hub_aggregator` | Opzionale |

---

### B\ Connessione Kafka (Input: Proximity Hub $\to$ Intermediate Hub)

Queste variabili configurano l'Hub come consumer sui topic popolati dal livello Proximity.

| Variabile                                            | Descrizione                                                              | Default                                                   |
|:-----------------------------------------------------|:-------------------------------------------------------------------------|:----------------------------------------------------------|
| **`KAFKA_BROKER_ADDRESS`**                           | Indirizzo IP/Hostname del broker Kafka.                                  | `localhost`                                               |
| **`KAFKA_BROKER_PORT`**                              | Porta del broker Kafka.                                                  | `9094`                                                    |
| **`KAFKA_COMMIT_TIMEOUT`**                           | Timeout per il commit degli offset Kafka (in secondi).                   | $5$                                                       |
| **`KAFKA_MAX_ATTEMPTS`**                             | Numero massimo di tentativi di invio di messaggi su Kafka.               | $10$                                                      |
| **`KAFKA_ATTEMPT_DELAY`**                            | Ritardo tra i tentativi di invio di messaggi su Kafka (in millisecondi). | $750$                                                     |
| **`KAFKA_PROXIMITY_FOG_HUB_REALTIME_DATA_TOPIC`**    | Topic per i dati in tempo reale.                                         | `aggregated-data-proximity-fog-hub`                       |
| **`KAFKA_PROXIMITY_FOG_HUB_AGGREGATED_STATS_TOPIC`** | Topic per le statistiche aggregate.                                      | `statistics-data-proximity-fog-hub`                       |
| **`KAFKA_PROXIMITY_FOG_HUB_CONFIGURATION_TOPIC`**    | Topic per i messaggi di configurazione.                                  | `configuration-proximity-fog-hub`                         |
| **`KAFKA_PROXIMITY_FOG_HUB_HEARTBEAT_TOPIC`**        | Topic per i messaggi di heartbeat.                                       | `heartbeats-proximity-fog-hub`                            |

---

### C\. Configurazione Database Persistenti

L'Intermediate Hub si connette a due database principali per la sua operatività: il *Sensor DB* per i dati e il *Metadata DB* per i metadati.

#### DB dei Dati dei Sensori (Sensor DB - Archiviazione)

| Variabile                      | Descrizione                             | Default           |
|:-------------------------------|:----------------------------------------|:------------------|
| **`POSTGRES_SENSOR_USER`**     | Nome utente per il Sensor DB.           | `admin`           |
| **`POSTGRES_SENSOR_PASSWORD`** | Password per il Sensor DB.              | `adminpass`       |
| **`POSTGRES_SENSOR_HOST`**     | Indirizzo del Sensor DB.                | `localhost`       |
| **`POSTGRES_SENSOR_PORT`**     | Porta del Sensor DB.                    | `5432`            |
| **`POSTGRES_SENSOR_DATABASE`** | Nome del database dei dati dei sensori. | `sensorcontinuum` |

#### DB Metadati di Regione (Metadata DB - Configurazione)

| Variabile                      | Descrizione                                | Default           |
|:-------------------------------|:-------------------------------------------|:------------------|
| **`POSTGRES_REGION_USER`**     | Nome utente per il Metadata DB.            | `admin`           |
| **`POSTGRES_REGION_PASSWORD`** | Password per il Metadata DB.               | `adminpass`       |
| **`POSTGRES_REGION_HOST`**     | Indirizzo del Metadata DB.                 | `localhost`       |
| **`POSTGRES_REGION_PORT`**     | Porta del Metadata DB.                     | `5434`            |
| **`POSTGRES_REGION_DATABASE`** | Nome del database dei metadati di regione. | `sensorcontinuum` |

---

### D\. Configurazioni di Elaborazione e Batching

Queste costanti e variabili governano la logica di aggregazione e l'efficienza nell'ingestione dei dati da Kafka al database.

#### Costanti di Aggregazione

| Parametro Logico                 | Valore costante | Funzione Corrispondente                                                                      |
|:---------------------------------|:----------------|:---------------------------------------------------------------------------------------------|
| **Intervallo Aggregazione**      | $30$ minuti     | Frequenza con cui l'Aggregator esegue l'aggregazione finale dei dati.                        |
| **Offset di Recupero Dati**      | $-20$ minuti    | Offset negativo di tempo per includere dati con ritardi di ricezione nell'aggregazione.      |
| **Offset Iniziale Aggregazione** | $-24$ ore       | Offset di tempo usato al primo avvio per includere eventuali dati più vecchi.                |
| **Lock ID Aggregazione**         | $472$           | ID del lock `pg_advisory_lock` sul database per garantire che solo un Aggregator sia attivo. |

#### E\. Parametri di Batching

| Variabile                                               | Descrizione                                                          | Default            |
|:--------------------------------------------------------|:---------------------------------------------------------------------|:-------------------|
| **`SENSOR_DATA_BATCH_SIZE`** / **`_TIMEOUT`**           | Dimensione e Timeout del batch per i dati **in tempo reale**.        | $100$ msg / $15$ s |
| **`AGGREGATED_DATA_BATCH_SIZE`** / **`_TIMEOUT`**       | Dimensione e Timeout del batch per i dati **aggregati**.             | $100$ msg / $15$ s |
| **`CONFIGURATION_MESSAGE_BATCH_SIZE`** / **`_TIMEOUT`** | Dimensione e Timeout del batch per i messaggi di **configurazione**. | $50$ msg / $5$ s   |
| **`HEARTBEAT_MESSAGE_BATCH_SIZE`** / **`_TIMEOUT`**     | Dimensione e Timeout del batch per i messaggi di **heartbeat**.      | $50$ msg / $5$ s   |

---

### F\. Parametri di Logging e Health Check

| Variabile                 | Descrizione                                                                                     | Default                                       |
|:--------------------------|:------------------------------------------------------------------------------------------------|:----------------------------------------------|
| **`HEALTHZ_SERVER`**      | Flag booleano per attivare il server HTTP per il controllo dello stato di salute  (`/healthz`). | `false`                                       |
| **`HEALTHZ_SERVER_PORT`** | Porta su cui il server Health Check si mette in ascolto.                                        | `8080`                                        |
| **`LOG_LEVEL`**           | Livello di dettaglio per l'output del logger.                                                   | `error` (Default), `warning`, `info`, `debug` |

-----

## Deploy in Locale dell'Intermediate Fog Hub

Il deployment dell'Intermediate Fog Hub (IFH) in un ambiente locale richiede l'orchestrazione di tre macro-componenti interdipendenti: il Broker Kafka, i Database Regionali e la suite di microservizi.

### Struttura di Docker Compose

Il file Compose di riferimento, [`intermediate-fog-hub.yaml`](../../deploy/compose/intermediate-fog-hub.yaml), definisce la logica applicativa come una serie di microservizi cooperanti (basati sull'immagine `fmasci/sc-intermediate-fog-hub:latest`). Tutti i servizi applicativi sono scalati a due istanze per l'Alta Disponibilità (HA).

| Servizio                | Istanze | Modalità                         | Funzione Principale                                                              |
|:------------------------|:--------|:---------------------------------|:---------------------------------------------------------------------------------|
| **REALTIME CONSUMER**   | 2       | `intermediate_hub_realtime`      | Consumo dati **in tempo reale** da Kafka e persistenza nel Sensor DB.            |
| **STATISTICS CONSUMER** | 2       | `intermediate_hub_statistics`    | Consumo delle **statistiche aggregate** da Kafka e persistenza nel Sensor DB.    |
| **CONFIGURATOR**        | 2       | `intermediate_hub_configuration` | Consumo messaggi di **configurazione** da Kafka e aggiornamento del Metadata DB. |
| **HEARTBEAT**           | 2       | `intermediate_hub_heartbeat`     | Consumo messaggi di **heartbeat** per tracciamento stato Hub.                    |
| **AGGREGATOR**          | 2       | `intermediate_hub_aggregator`    | Esecuzione dell'aggregazione statistica finale .                                 |

Il file sfrutta le estensioni YAML per ereditare configurazioni comuni e garantire la coerenza tra le istanze.

```yaml
# --- environment base ---
x-region-hub-env: &region-hub-env
  KAFKA_BROKER_ADDRESS: "kafka-broker.${REGION}.sensor-continuum.local"
  KAFKA_BROKER_PORT: "${KAFKA_PORT}"
  POSTGRES_CLOUD_HOST: "metadata-db.cloud.sensor-continuum.local"
  POSTGRES_CLOUD_PORT: "${POSTGRES_CLOUD_METADATA_PORT}"
  POSTGRES_REGION_HOST: "metadata-db.${REGION}.sensor-continuum.local"
  POSTGRES_REGION_PORT: "${POSTGRES_REGION_METADATA_PORT}"
  POSTGRES_SENSOR_HOST: "measurement-db.${REGION}.sensor-continuum.local"
  POSTGRES_SENSOR_PORT: "${POSTGRES_REGION_SENSOR_PORT}"
  OPERATION_MODE: "loop"
  HEALTHZ_SERVER: "true"
  HEALTHZ_SERVER_PORT: "8080"

# --- blocco base per tutti i region hub ---
x-region-hub-base: &region-hub-base
  image: fmasci/sc-intermediate-fog-hub:latest
  build:
    context: ../..
    dockerfile: deploy/docker/intermediate-fog-hub.Dockerfile
  healthcheck:
    test: [ "CMD", "curl", "-f", "http://localhost:8080/healthz" ]
    interval: 60s
    timeout: 30s
    retries: 10
  restart: unless-stopped
  networks:
    - region-hub-bridge
```

### Variabili d'Ambiente Obbligatorie

Il deployment è interamente parametrico. Tutte le variabili essenziali devono essere definite in per risolvere nomi di rete, FQDN e mappature delle porte tra l'Hub e la sua infrastruttura.

| Variabile                           | Descrizione                                  | Utilizzo nel Compose (Hostname)                        |
|:------------------------------------|:---------------------------------------------|:-------------------------------------------------------|
| **`REGION`**                        | Identificativo univoco della regione logica. | Cruciale per tutti gli hostname e i nomi di rete.      |
| **`KAFKA_PORT`**                    | Porta host per il Broker Kafka.              | Mappatura porta.                                       |
| **`KAFKA_PUBLIC_IP`**               | Indirizzo IP pubblico/locale dell'host.      | Utilizzato internamente da Kafka per l'indirizzamento. |
| **`POSTGRES_REGION_METADATA_PORT`** | Porta host per il **Region Metadata DB**.    | Mappatura porta.                                       |
| **`POSTGRES_REGION_SENSOR_PORT`**   | Porta host per il **Region Sensor DB**.      | Mappatura porta.                                       |

### Deployment dell'Infrastruttura Regionale

L'avvio dei Microservizi Hub richiede che l'infrastruttura di supporto (Kafka e Database) sia attiva, accessibile e inizializzata. L'infrastruttura può essere deployata con implementazioni custom o utilizzando i template forniti:

* **Broker Kafka:** [`deploy/compose/kafka-broker.yml`](../../deploy/compose/kafka-broker.yml)
* **Database Regionali:**
  * [`deploy/compose/region-databases.yml`](../../deploy/compose/region-databases.yml) 
  * [`deploy/docker/region-metadata-database.Dockerfile`](../../deploy/docker/region-metadata-database.Dockerfile) (PostGIS)
  * [`deploy/docker/region-sensor-database.Dockerfile`](../../deploy/docker/region-sensor-database.Dockerfile) (TimescaleDB)

#### Creazione dei Volumi Persistenti

Se si utilizzano i template Compose forniti per l'infrastruttura, i volumi dati per Kafka e i Database sono definiti come `external: true` e devono essere creati **manualmente** prima dell'avvio dei servizi che li utilizzano.

**Esempio di creazione dei volumi per la region `Lazio`:**

```bash
docker volume create kafka-data-Lazio
docker volume create region-metadata-data-Lazio
docker volume create region-sensor-data-Lazio
```

### Requisiti di Inizializzazione di Kafka

Se non si utilizza il template Compose fornito per Kafka, è necessario assicurarsi che il Broker sia correttamente configurato e che i topic richiesti siano creati.

Il Broker deve essere accessibile all'hostname `kafka-broker.${REGION}.sensor-continuum.local`*. I servizi Hub richiedono l'esistenza dei seguenti topic, con le relative configurazioni per le partizioni e la retention policy:

| Nome Topic Richiesto                    | Configurazione                                    |
|:----------------------------------------|:--------------------------------------------------|
| **`aggregated-data-proximity-fog-hub`** | Standard                                          |
| **`configuration-proximity-fog-hub`**   | Standard                                          |
| **`statistics-data-proximity-fog-hub`** | Standard                                          |
| **`heartbeats-proximity-fog-hub`**      | `cleanup.policy=compact,delete` (Compacted Topic) |

### Requisiti di Inizializzazione dei Database Regionali

Se non si utilizza il template Compose fornito per i Database, è necessario assicurarsi che entrambi i database siano correttamente configurati e inizializzati con gli schemi SQL richiesti.
Le due istanze database devono essere accessibili ai rispettivi hostname e devono avere gli schemi SQL inizializzati tramite gli script forniti.

#### A\. Region Metadata Database

* **Hostname:** Deve risolvere a `metadata-db.${REGION}.sensor-continuum.local`.
* **Tabelle Richieste** (create dallo script [`init-region-metadata-db.sql`](../../configs/postgresql/init-region-metadata-db.sql)):
    * **`region_hubs`**: Traccia gli Intermediate Hub (se scalati o replicati).
    * **`macrozone_hubs`**: Traccia lo stato e la registrazione di tutti i Proximity Hub.
    * **`zone_hubs`**: Traccia lo stato e la registrazione di tutti gli Edge Hub.
    * **`sensors`**: Contiene tutti i metadati (configurazione, stato, location) dei sensori.

#### B\. Region Sensor Database

* **Hostname:** Deve risolvere a `measurement-db.${REGION}.sensor-continuum.local`.
* **Schema Richiesto** (inizializzato dallo script [`init-region-sensors-db.sql`](../../configs/postgresql/init-region-sensors-db.sql)):
    * **Hypertable:** **`sensor_measurements`** (contiene tutti i dati in tempo reale; deve essere configurata come Hypertable su colonna `time`).
    * **Tabelle Aggregate:**
        * **`region_aggregated_statistics`**
        * **`macrozone_aggregated_statistics`**
        * **`zone_aggregated_statistics`**
    * **Viste di Aggregazione Continua:** Viste materializzate (es. `*_daily_agg`, `*_monthly_agg`) ottimizzate per query storiche.

### Preparazione ed Esecuzione del Deploy

Una volta che l'infrastruttura è attiva e inizializzata, si può procedere con l'avvio del layer applicativo dell'Hub.

#### 1\. Risoluzione del Nome Host dei Broker e dei Database

I microservizi dell'Hub devono potersi connettere ai servizi di infrastruttura utilizzando hostname complessi che richiedono una risoluzione DNS locale:

1.  **Broker Kafka:** `kafka-broker.${REGION}.sensor-continuum.local`
2.  **Region Metadata DB:** `metadata-db.${REGION}.sensor-continuum.local`
3.  **Region Sensor DB:** `measurement-db.${REGION}.sensor-continuum.local`

In un ambiente Docker locale, si hanno le seguenti opzioni per la risoluzione:

* **Utilizzo di `extra_hosts`:** Configurare la sezione `extra_hosts` nel file Docker Compose dei microservizi Hub, mappando l'indirizzo del broker all'IP appropriato.
* **Modifica del File `/etc/hosts`:** Modificare il file **`/etc/hosts`** del sistema operativo host per puntare i record DNS utilizzati direttamente all'indirizzo IP del nodo che ospita i servizi.

#### 2\. Avvio con Docker Compose

Dopo aver configurato il risolutore di hostname locale, inizializzata l'infrastruttura e definito le variabili d'ambiente, l'avvio viene eseguito tramite il comando standard di Docker Compose.

```bash
# Avvio di tutti i microservizi dell'Intermediate Fog Hub
docker compose -f intermediate-fog-hub.yaml up -d
```

> **⚠️ NOTA OPERATIVA**
>
> **Dipendenza e Ordine di Avvio:**
> I microservizi Hub sono progettati per essere resilienti. Non usano `depends_on`, ma entrano in un ciclo di attesa, ritentando la connessione finché il Kafka Broker e i Database Regionali non sono completamente disponibili e pronti per le transazioni.
>
> **File d'Ambiente:**
> Per standardizzare il deployment, è cruciale definire le variabili d'ambiente nel file separato **`Lazio.env`**.
>
> **Esempio di Deploy Completo:**
>
> ```bash
> # Avvio del layer applicativo Hub con nome progetto 'Lazio_Region_Hub'
> docker compose -f intermediate-fog-hub.yaml --env-file Lazio.env -p Lazio_Region_Hub up -d
> ```

-----

## Deploy su AWS dell'Intermediate Fog Hub

Il deployment dell'infrastruttura regionale è gestito da AWS CloudFormation tramite lo script Bash orchestratore [`deploy_region.sh`](../../deploy/scripts/deploy_region.sh).
Lo script esegue un deployment modulare e sequenziale dell'intera infrastruttura regionale, di cui l'Hub Intermediate costituisce la fase finale.

### 1\. Caricamento Asset su S3

Prima di eseguire il deployment, tutti gli asset necessari (script di installazione, file Docker Compose, file di servizio Systemd e script di analisi) devono essere caricati in un Bucket S3 dedicato.

Questo processo è documentato in [`setup_bucket.md`](./setup_bucket.md) e gestito dallo script [`setup_bucket.sh`](../../deploy/scripts/setup_bucket.sh).

### 2\. Orchestrazione dell'Infrastruttura Regionale

Lo script [`deploy_region.sh`](../../deploy/scripts/deploy_region.sh) funge da orchestratore principale, eseguendo il deployment modulare dell'intera infrastruttura regionale in quattro passaggi sequenziali tramite template AWS CloudFormation. L'Intermediate Fog Hub come applicazione risiede nel componente Services, che dipende dalla creazione corretta delle fasi precedenti.

#### A\. Componente VPC (Template [`VPC.yaml`](../../deploy/cloudformation/region/VPC.yaml))

Questa è la fase fondamentale che crea il contesto di rete isolato e sicuro per tutti i servizi della regione.
1.  **Creazione Rete:** Vengono create la VPC dedicata `$REGION-vpc`, una Subnet e un Security Group `$REGION-sg`.
2.  **DNS Interno:** Viene istanziata una Private Hosted Zone Route 53 `$REGION.sensor-continuum.local`, vitale per consentire ai servizi (Kafka, Databases, Services) di comunicare tra loro utilizzando hostname risolvibili internamente.
3.  **Preparazione:** Lo script recupera gli ID di VPC, Subnet e Security Group per passarli come input ai template successivi.

#### B\. Componente Kafka (Template [`Kafka.yaml`](../../deploy/cloudformation/region/Kafka.yaml))

Questa fase stabilisce il sistema di messaggistica asincrona e ad alta throughput per l'ingestion dei dati aggregati dal livello Proximity.
1.  **Istanziamento Broker:** Viene creata l'istanza EC2 Kafka Broker all'interno della VPC e Subnet appena definite.
2.  **Accesso e Connettività:** Si genera una KeyPair `$REGION-kafka-key` per l'accesso SSH. Il broker viene registrato nella Hosted Zone privata, rendendo l'hostname interno `kafka-broker.$HOSTED_ZONE_NAME` l'endpoint ufficiale per tutti i consumer.

#### C\. Componente Databases (Template [`Databases.yaml`](../../deploy/cloudformation/region/Databases.yaml))

Questa fase implementa l'infrastruttura di persistenza per l'archiviazione a lungo termine dei dati di misurazione e metadati.
1.  **Istanziamento Database:** Viene creata l'istanza EC2 Databases che ospiterà i databases.
2.  **DNS Pubblico/Privato:** L'istanza è configurata per avere record DNS sia privati (per l'accesso interno) sia pubblici, utilizzando l'ID della Hosted Zone pubblica.

#### D\. Componente Services (Template [`services.yaml`](../../deploy/cloudformation/region/services.yaml))

Questa è la fase finale del deployment infrastrutturale, creando l'ambiente IaaS che ospiterà l'Intermediate Fog Hub (IFH).
1.  **Creazione Istanza IFH:** Viene istanziata l'istanza EC2 Services, utilizzando il tipo di istanza specificato.
2.  **Installazione via `UserData`**: Il blocco `UserData` installa Docker/Docker Compose ed esegue lo script [`deploy_intermediate_services.sh`](../../deploy/scripts/deploy/deploy_intermediate_services.sh).

##### Dettagli Operativi del Deployment Script in EC2

Lo script **`deploy_intermediate_services.sh`** è responsabile dell'avvio sicuro dei microservizi IFH e della configurazione della resilienza.

###### Sequenza di Avvio dell'Hub

* **Preparazione:** Scarica il file [`intermediate-fog-hub.yaml`](../../deploy/compose/intermediate-fog-hub.yaml) da S3, carica le variabili d'ambiente e interrompe eventuali container preesistenti.
* **Avvio Servizi:** Esegue una chiamata `docker-compose up -d` per lanciare tutti i microservizi dell'Hub tramite il template scaricato con le relative variabili d'ambiente.
* **Download Strumenti:** Scarica lo script di latenza [`init-delay.sh`](../../deploy/scripts/inits/init-delay.sh) e lo script di analisi [`analyze_throughput.sh`](../../deploy/scripts/performance/analyze_throughput.sh).

###### Gestione Operativa e Resilienza

Lo script [`deploy_intermediate_services.sh`](../../deploy/scripts/deploy/deploy_intermediate_services.sh) integra diverse funzionalità per la gestione e la simulazione di un ambiente Fog robusto:

* **Configurazione Latenza:** Esecuzione dello script [`init-delay.sh`](../../deploy/scripts/inits/init-delay.sh) per simulare latenza, jitter e packet loss sull'interfaccia di rete dell'istanza EC2, utilizzando le variabili d'ambiente (es. `${NETWORK_DELAY:-20ms}`).
* **Servizio Systemd:** Configurazione del servizio [`sc-deploy-services.service`](../../deploy/scripts/services/sc-deploy.service.template) per rieseguire automaticamente lo script di deployment all'avvio del sistema, garantendo la resilienza dell'Hub.