# Istruzioni per l'Intermediate Fog Hub

L'**Intermediate Fog Hub** (o **Hub di Regione**) rappresenta il livello più alto e centralizzato del piano **Fog** nel **Compute Continuum**. Operando a livello di **Regione** (l'intera area geografica gestita), funge da punto di aggregazione finale e di persistenza per tutti i dati consolidati provenienti dai livelli sottostanti.

## Architettura e Funzionalità ad Alto Livello

1.  **Ingestione Dati (Kafka Consumer):** L'Hub funge da **Consumer** principale sui topic Kafka, ricevendo tutti i flussi di dati (tempo reale, statistiche aggregate, configurazione e heartbeat) inoltrati dai Proximity Fog Hub. Il `KafkaGroupId` fisso (`intermediate-fog-hub`) assicura il bilanciamento del carico tra le istanze.
2.  **Archiviazione Primaria (Sensor DB):** Utilizza un database relazionale persistente (**PostgreSQL/TimescaleDB**), denominato **Sensor DB**, come fonte di verità per tutte le serie storiche aggregate a livello regionale.
3.  **Aggregazione Finale:** Il microservizio **Aggregator** esegue l'aggregazione statistica finale per intervalli di **30 minuti**, consolidando i dati per l'archiviazione storica a lungo termine.
4.  **Gestione Metadati (Metadata DB):** Si connette a un database di metadati dedicato (**Metadata DB**) per gestire la configurazione e lo stato globale del sistema a livello regionale.

---

## Variabili d'Ambiente dell'Intermediate Fog Hub

La configurazione del servizio è gestita interamente tramite variabili d'ambiente.

### 1. Identità e Modalità Operative

Queste variabili definiscono l'identità del servizio e il suo ruolo specifico.

| Variabile            | Descrizione                                          | Default / Valori Ammessi                                                                                                                                                                                   | Stato     |
|:---------------------|:-----------------------------------------------------|:-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:----------|
| **`HUB_ID`**         | Identificatore univoco dell'istanza Hub.             | UUID Generato                                                                                                                                                                                              | Opzionale |
| **`OPERATION_MODE`** | Controlla il ciclo di vita del servizio.             | `loop` (ciclo continuo) / `once`                                                                                                                                                                           | Opzionale |
| **`SERVICE_MODE`**   | Definisce il ruolo operativo specifico dell'istanza. | `intermediate_hub` (Tutte le funzionalità) / `intermediate_hub_realtime` / `intermediate_hub_statistics` / `intermediate_hub_configuration` / `intermediate_hub_heartbeat` / `intermediate_hub_aggregator` | Opzionale |

---

### 2. Connessione Kafka (Input: Proximity Hub $\to$ Intermediate Hub)

Queste variabili configurano l'Hub come **Consumer** sui topic popolati dal livello Proximity (Macrozona).

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

### 3. Configurazione Database Persistenti (PostgreSQL)

L'Intermediate Hub si connette a due database principali per la sua operatività: il **Sensor DB** per i dati e il **Metadata DB** per i metadati.

#### A. DB dei Dati dei Sensori (Sensor DB - Archiviazione)

| Variabile                      | Descrizione                             | Default           |
|:-------------------------------|:----------------------------------------|:------------------|
| **`POSTGRES_SENSOR_USER`**     | Nome utente per il Sensor DB.           | `admin`           |
| **`POSTGRES_SENSOR_PASSWORD`** | Password per il Sensor DB.              | `adminpass`       |
| **`POSTGRES_SENSOR_HOST`**     | Indirizzo del Sensor DB.                | `localhost`       |
| **`POSTGRES_SENSOR_PORT`**     | Porta del Sensor DB.                    | `5432`            |
| **`POSTGRES_SENSOR_DATABASE`** | Nome del database dei dati dei sensori. | `sensorcontinuum` |

#### B. DB Metadati di Regione (Metadata DB - Configurazione)

| Variabile                      | Descrizione                                | Default           |
|:-------------------------------|:-------------------------------------------|:------------------|
| **`POSTGRES_REGION_USER`**     | Nome utente per il Metadata DB (metadati). | `admin`           |
| **`POSTGRES_REGION_PASSWORD`** | Password per il Metadata DB (metadati).    | `adminpass`       |
| **`POSTGRES_REGION_HOST`**     | Indirizzo del Metadata DB.                 | `localhost`       |
| **`POSTGRES_REGION_PORT`**     | Porta del Metadata DB.                     | `5434`            |
| **`POSTGRES_REGION_DATABASE`** | Nome del database dei metadati di regione. | `sensorcontinuum` |

---

### 4. Configurazioni di Elaborazione e Batching

Queste costanti e variabili governano la logica di aggregazione e l'efficienza nell'ingestione dei dati da Kafka al database.

#### Costanti di Aggregazione (Dimensione/Timeout in secondi)

| Parametro Logico                 | Valore costante        | Funzione Corrispondente                                                                          |
|:---------------------------------|:-----------------------|:-------------------------------------------------------------------------------------------------|
| **Intervallo Aggregazione**      | $30$ minuti            | Frequenza con cui l'Aggregator esegue l'aggregazione finale dei dati.                            |
| **Offset di Recupero Dati**      | $-20$ minuti           | Offset negativo di tempo per includere dati con ritardi di ricezione nell'aggregazione.          |
| **Offset Iniziale Aggregazione** | $-24$ ore              | Offset di tempo usato al primo avvio per includere eventuali dati più vecchi.                    |
| **Lock ID Aggregazione**         | $472$                  | ID del lock sul database (`pg_advisory_lock`) per garantire che solo un Aggregator sia attivo.   |

#### Parametri di Batching (Dimensione/Timeout in secondi)

| Variabile                                               | Descrizione                                                          | Default            |
|:--------------------------------------------------------|:---------------------------------------------------------------------|:-------------------|
| **`SENSOR_DATA_BATCH_SIZE`** / **`_TIMEOUT`**           | Dimensione e Timeout del batch per i dati **in tempo reale**.        | $100$ msg / $15$ s |
| **`AGGREGATED_DATA_BATCH_SIZE`** / **`_TIMEOUT`**       | Dimensione e Timeout del batch per i dati **aggregati**.             | $100$ msg / $15$ s |
| **`CONFIGURATION_MESSAGE_BATCH_SIZE`** / **`_TIMEOUT`** | Dimensione e Timeout del batch per i messaggi di **configurazione**. | $50$ msg / $5$ s   |
| **`HEARTBEAT_MESSAGE_BATCH_SIZE`** / **`_TIMEOUT`**     | Dimensione e Timeout del batch per i messaggi di **heartbeat**.      | $50$ msg / $5$ s   |

---

### 5. Health Check Server

| Variabile                 | Descrizione                                                                                    | Default |
|:--------------------------|:-----------------------------------------------------------------------------------------------|:--------|
| **`HEALTHZ_SERVER`**      | Flag booleano per attivare il server HTTP per il controllo dello stato di salute (`/healthz`). | `false` |
| **`HEALTHZ_SERVER_PORT`** | Porta su cui il server Health Check si mette in ascolto.                                       | `8080`  |

---

## Deploy dell'Intermediate Hub in Locale con Docker Compose

Il deployment dell'**Intermediate Fog Hub** (IFH) in un ambiente locale richiede l'orchestrazione di tre macro-componenti interdipendenti: il Broker Kafka, i Database Regionali e la suite di microservizi.

-----

## 1\. Struttura di Docker Compose

Il file Compose di riferimento, `intermediate-fog-hub.yaml`, definisce la logica applicativa Go come una **serie di microservizi cooperanti** (basati sull'immagine `fmasci/sc-intermediate-fog-hub:latest`). Tutti i servizi applicativi sono scalati a **due istanze** per l'Alta Disponibilità (HA).

| Servizio                | Istanze | Modalità                         | Funzione Principale                                                                           |
|:------------------------|:--------|:---------------------------------|:----------------------------------------------------------------------------------------------|
| **REALTIME CONSUMER**   | 2       | `intermediate_hub_realtime`      | Consumo dati **in tempo reale** da Kafka e persistenza nel Sensor DB.                         |
| **STATISTICS CONSUMER** | 2       | `intermediate_hub_statistics`    | Consumo delle **statistiche aggregate** da Kafka e persistenza nel Sensor DB.                 |
| **CONFIGURATOR**        | 2       | `intermediate_hub_configuration` | Consumo messaggi di **configurazione** da Kafka e aggiornamento del Metadata DB.              |
| **HEARTBEAT**           | 2       | `intermediate_hub_heartbeat`     | Consumo messaggi di **heartbeat** per tracciamento stato Hub.                                 |
| **AGGREGATOR**          | 2       | `intermediate_hub_aggregator`    | Esecuzione dell'aggregazione statistica finale (ogni **30 minuti**, scrittura nel Sensor DB). |

-----

## 2\. Variabili d'Ambiente Obbligatorie

Il deployment è interamente parametrico. Tutte le variabili essenziali devono essere definite in un file locale (es. **`.env`**) per risolvere nomi di rete, FQDN e mappature delle porte tra l'Hub e la sua infrastruttura.

| Variabile                           | Descrizione                                  | Utilizzo nel Compose (Hostname)                        |
|:------------------------------------|:---------------------------------------------|:-------------------------------------------------------|
| **`REGION`**                        | Identificativo univoco della regione logica. | Cruciale per tutti gli hostname e i nomi di rete.      |
| **`KAFKA_PORT`**                    | Porta host per il Broker Kafka.              | Mappatura porta.                                       |
| **`KAFKA_PUBLIC_IP`**               | Indirizzo IP pubblico/locale dell'host.      | Utilizzato internamente da Kafka per l'indirizzamento. |
| **`POSTGRES_REGION_METADATA_PORT`** | Porta host per il **Region Metadata DB**.    | Mappatura porta.                                       |
| **`POSTGRES_REGION_SENSOR_PORT`**   | Porta host per il **Region Sensor DB**.      | Mappatura porta.                                       |

-----

## 3\. Deployment dell'Infrastruttura Regionale (Prerequisito)

L'avvio dei Microservizi Hub richiede che l'infrastruttura di supporto (Kafka e Database) sia attiva, accessibile e inizializzata. L'infrastruttura può essere deployata con implementazioni custom o utilizzando i template forniti:

* **Broker Kafka:**
  * Compose: `deploy/compose/kafka-broker.yml`
  * Dockerfile: `deploy/docker/kafka-init-topics.Dockerfile`
* **Database Regionali:**
  * Compose: `deploy/compose/region-databases.yml`
  * Dockerfile:
    * `deploy/docker/region-metadata-database.Dockerfile` (PostGIS)
    * `deploy/docker/region-sensor-database.Dockerfile` (TimescaleDB)

### A. Creazione dei Volumi Persistenti

Se si utilizzano i template Compose forniti per l'infrastruttura, i volumi dati per Kafka e i Database sono definiti come **`external: true`** e devono essere creati **manualmente** prima dell'avvio dei servizi che li utilizzano.

**Esempio di creazione dei volumi (assumendo `$REGION=Lazio`):**

```bash
docker volume create kafka-data-Lazio
docker volume create region-metadata-data-Lazio
docker volume create region-sensor-data-Lazio
```

### B. Broker Kafka: Requisiti di Inizializzazione

Il Broker deve essere accessibile all'hostname **`kafka-broker.${REGION}.sensor-continuum.local`**. I servizi Hub richiedono l'esistenza dei seguenti topic, con le relative configurazioni per le partizioni e la retention policy:

| Nome Topic Richiesto                    | Configurazione                                    |
|:----------------------------------------|:--------------------------------------------------|
| **`aggregated-data-proximity-fog-hub`** | Standard                                          |
| **`configuration-proximity-fog-hub`**   | Standard                                          |
| **`statistics-data-proximity-fog-hub`** | Standard                                          |
| **`heartbeats-proximity-fog-hub`**      | `cleanup.policy=compact,delete` (Compacted Topic) |

### C. Database Regionali: Requisiti di Inizializzazione

Le due istanze database devono essere accessibili ai rispettivi hostname e devono avere gli schemi SQL inizializzati tramite gli script forniti.

#### 1\. Region Metadata Database

* **Hostname:** Deve risolvere a **`metadata-db.${REGION}.sensor-continuum.local`**.
* **Tabelle Richieste** (create dallo script **`configs/postgresql/init-region-metadata-db.sql`**):
    * **`region_hubs`**: Traccia gli Intermediate Hub (se scalati o replicati).
    * **`macrozone_hubs`**: Traccia lo stato e la registrazione di tutti i Proximity Hub.
    * **`zone_hubs`**: Traccia lo stato e la registrazione di tutti gli Edge Hub.
    * **`sensors`**: Contiene tutti i metadati (configurazione, stato, location) dei sensori.

#### 2\. Region Sensor Database

* **Hostname:** Deve risolvere a **`measurement-db.${REGION}.sensor-continuum.local`**.
* **Schema Richiesto** (inizializzato dallo script **`configs/postgresql/init-region-sensors-db.sql`**):
    * **Hypertable:** **`sensor_measurements`** (contiene tutti i dati in tempo reale; deve essere configurata come Hypertable su colonna `time`).
    * **Tabelle Aggregate:**
        * **`region_aggregated_statistics`**
        * **`macrozone_aggregated_statistics`**
        * **`zone_aggregated_statistics`**
    * **Viste di Aggregazione Continua:** Viste materializzate (es. `*_daily_agg`, `*_monthly_agg`) ottimizzate per query storiche.

-----

## 4\. Preparazione e Esecuzione del Deploy

Una volta che l'infrastruttura (Kafka e DBs) è attiva e inizializzata, si può procedere con l'avvio del layer applicativo dell'Hub.

#### A. Risoluzione del Nome Host dei Broker e dei Database

I microservizi dell'Hub devono potersi connettere ai servizi di infrastruttura utilizzando hostname complessi che richiedono una risoluzione DNS locale:

1.  **Broker Kafka (Livello Intermediate):** `kafka-broker.${REGION}.sensor-continuum.local`
2.  **Region Metadata DB:** `metadata-db.${REGION}.sensor-continuum.local`
3.  **Region Sensor DB:** `measurement-db.${REGION}.sensor-continuum.local`

In un ambiente Docker locale, si hanno le seguenti opzioni per la risoluzione:

* **Utilizzo di `extra_hosts`:** Configurare la sezione `extra_hosts` nel file Docker Compose dei microservizi Hub, mappando l'indirizzo del broker all'IP appropriato.
* **Modifica del File `/etc/hosts`:** Modificare il file **`/etc/hosts`** del sistema operativo host per puntare i record DNS utilizzati direttamente all'indirizzo IP del nodo che ospita i servizi.

#### B. Avvio con Docker Compose

Dopo aver creato i volumi, inizializzato l'infrastruttura e definito le variabili d'ambiente (nel file **`.env`**), l'avvio viene eseguito tramite il comando standard di Docker Compose.

```bash
# Avvio di tutti i microservizi dell'Intermediate Fog Hub
docker compose -f intermediate-fog-hub.yaml up -d
```

> **⚠️ NOTA OPERATIVA**
>
> **Dipendenza e Ordine di Avvio:**
> I microservizi Hub (Go) sono progettati per essere resilienti. Non usano `depends_on`, ma entrano in un ciclo di attesa, ritentando la connessione finché il **Kafka Broker** e i **Database Regionali** non sono completamente disponibili e pronti per le transazioni.
>
> **File d'Ambiente:**
> Per standardizzare il deployment, è cruciale definire le variabili d'ambiente nel file separato **`.env`**.
>
> **Esempio di Deploy Completo:**
>
> ```bash
> # Avvio del layer applicativo Hub con nome progetto 'Lazio_Region_Hub'
> docker compose -f intermediate-fog-hub.yaml --env-file .env -p Lazio_Region_Hub up -d
> ```

-----

## Deploy su AWS dell'Intermediate Fog Hub

Il deployment dell'infrastruttura regionale è gestito da **AWS CloudFormation** tramite lo script Bash orchestratore **`deploy_region.sh`**.
Lo script esegue un deployment modulare e sequenziale dell'intera infrastruttura regionale, di cui l'Hub Intermediate costituisce la fase applicativa finale (**Componente `services`**).

### 1\. Preparazione: Caricamento Asset su S3

Prima di eseguire il deployment, tutti gli asset necessari (script di installazione, file Docker Compose, file di servizio Systemd e script di analisi) devono essere caricati in un **Bucket S3** dedicato.

Questo processo è documentato in [`setup_bucket.md`](./setup_bucket.md) e gestito dallo script **`deploy/scripts/setup_bucket.sh`**.

### 2\. Orchestrazione dell'Infrastruttura Regionale

Lo script **`deploy_region.sh`** funge da orchestratore principale, eseguendo il deployment modulare dell'intera infrastruttura regionale in quattro passaggi sequenziali tramite template AWS CloudFormation. L'Intermediate Fog Hub come applicazione risiede nel componente **Services**, che dipende dalla creazione corretta delle fasi precedenti.

#### A. Componente VPC (Template `VPC.yaml`)

Questa è la fase fondamentale che crea il contesto di rete isolato e sicuro per tutti i servizi della regione.
1.  **Creazione Rete:** Vengono create la **VPC** dedicata (`$REGION-vpc`), una **Subnet** e un **Security Group** (`$REGION-sg`).
2.  **DNS Interno:** Viene istanziata una **Private Hosted Zone** Route 53 (es. `$REGION.sensor-continuum.local`), vitale per consentire ai servizi (Kafka, Databases, Services) di comunicare tra loro utilizzando hostname risolvibili internamente.
3.  **Preparazione:** Lo script recupera gli ID di VPC, Subnet e Security Group per passarli come input ai template successivi.

#### B. Componente Kafka (Template `Kafka.yaml`)

Questa fase stabilisce il sistema di messaggistica asincrona e ad alta throughput per l'ingestion dei dati aggregati dal livello Proximity.
1.  **Istanziamento Broker:** Viene creata l'istanza **EC2 Kafka Broker** (`$REGION-kafka-broker`) all'interno della VPC e Subnet appena definite.
2.  **Accesso e Connettività:** Si genera una **KeyPair** (`$REGION-kafka-key`) per l'accesso SSH. Il broker viene registrato nella Hosted Zone privata, rendendo l'hostname interno (es. `kafka-broker.$HOSTED_ZONE_NAME`) l'endpoint ufficiale per tutti i consumer.

#### C. Componente Databases (Template `Databases.yaml`)

Questa fase implementa l'infrastruttura di persistenza per l'archiviazione a lungo termine dei dati di misurazione e metadati.
1.  **Istanziamento Database:** Viene creata l'istanza **EC2 Databases** (`$REGION-databases`) che ospiterà i database (Measurement e Metadata).
2.  **DNS Pubblico/Privato:** L'istanza è configurata per avere record DNS sia privati (per l'accesso interno) sia **pubblici** (es. `region.measurement-db.sensor-continuum.it`), utilizzando l'ID della Hosted Zone pubblica.
3.  **Dipendenza:** Questa fase dipende dalla Private Hosted Zone creata dalla VPC.

#### D. Componente Services (Template `services.yaml`)

Questa è la fase finale del deployment infrastrutturale, creando l'ambiente IaaS che ospiterà l'Intermediate Fog Hub (IFH).
1.  **Creazione Istanza IFH:** Viene istanziata l'istanza **EC2 Services** (`$REGION-services`), utilizzando il tipo di istanza specificato (`--instance-type`).
2.  **Innesco Applicativo:** Il template CloudFormation provvede solo all'infrastruttura EC2. L'installazione e l'avvio del software applicativo IFH sono gestiti dal Blocco **UserData**, che viene eseguito automaticamente al boot dell'istanza.

---

### 3\. Deploy Applicativo Interno

Il Blocco **UserData**, contenuto nel template `services.yaml`, è uno script Bash eseguito automaticamente all'avvio dell'istanza EC2 Services. Esso gestisce il setup del sistema operativo e l'avvio dei microservizi.

#### Funzionamento del Blocco UserData

Il `UserData` ha il compito di configurare l'ambiente operativo prima di eseguire lo script di deployment vero e proprio:
* **Installazione Pre-requisiti:** Controlla e installa **AWS CLI v2**. Successivamente, esegue gli script di inizializzazione (inclusa l'installazione di **Docker** e **Docker Compose**) scaricati da S3.
* **Download Asset:** Scarica il file di configurazione **`.env`** e lo script di deployment **`deploy_intermediate_services.sh`** dal bucket S3 nella directory `/home/ec2-user`.
* **Esecuzione Deployment:** Avvia lo script applicativo scaricato.

#### Avvio dei Microservizi

Lo script **`deploy_intermediate_services.sh`** è responsabile dell'avvio sicuro dei microservizi IFH e della configurazione della resilienza.

* **Preparazione:** Scarica il file **`intermediate-fog-hub.yaml`** da S3, carica le variabili d'ambiente e interrompe eventuali container preesistenti.
* **Avvio Servizi:** Esegue una chiamata **`docker-compose up -d`** per lanciare tutti i microservizi dell'Hub tramite il template scaricato con le relative variabili d'ambiente.
* **Download Strumenti:** Scarica lo script di **latenza** (`init-delay.sh`) e lo script di **analisi** (`analyze_throughput.sh`).

#### Gestione Operativa e Resilienza

Le seguenti funzionalità cruciali sono implementate tramite lo script applicativo per garantire un funzionamento robusto e stabile dell'IFH:

| Funzionalità               | Dettaglio Operativo                                                                                                                                                                                   |
|:---------------------------|:------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Configurazione Latenza** | Esecuzione dello script **`init-delay.sh`** per simulare latenza, jitter e packet loss sull'interfaccia di rete dell'istanza EC2, utilizzando le variabili d'ambiente (es. `${NETWORK_DELAY:-20ms}`). |
| **Servizio Systemd**       | Configurazione del servizio **`sc-deploy-services.service`** per rieseguire automaticamente lo script di deployment all'avvio del sistema, garantendo la resilienza dell'Hub.                         |