# Istruzioni per l'Edge Hub

L'**Edge Hub** è implementato come un'applicazione distribuita Go (Go-Routine based) che opera al livello più vicino ai sensori fisici. Il suo ruolo principale nel **Compute Continuum** è quello di ridurre il traffico dati verso il cloud e fornire funzionalità di **elaborazione a bassa latenza**.

1.  **Ingestione e Registrazione Dati (MQTT):** L'Edge Hub agisce come consumer per i dati inviati dai **Sensor Agents** tramite un broker MQTT locale (broker interno all'Edge Zone o all'Edge Macrozone). Riceve, valida e registra tutti i messaggi.
2.  **Architettura a Microservizi Cooperanti:** Il progetto sfrutta un pattern a microservizi all'interno dell'unico container Edge Hub. L'applicazione principale può essere avviata in diverse modalità operative (Service Modes), permettendo a componenti specializzati di coesistere e lavorare sullo stesso set di dati (tramite Redis Cache), ad esempio:
    * **Filtro:** Applica il rilevamento degli **Outlier** in tempo reale su ogni serie di dati ricevuta, utilizzando metodi statistici (deviazione standard).
    * **Aggregatore:** Aggrega i dati filtrati a intervalli regolari, riducendo la granularità e il volume dei dati inviati ai livelli superiori.
    * **Cleaner:** Gestisce i timeout dei sensori, marcandoli come *unhealthy* se non inviano dati o heartbeat entro limiti predefiniti, gestendo così la **resilienza del sistema**.
3.  **Stato Condiviso (Redis):** Utilizza un'istanza Redis locale per mantenere lo stato dei sensori (es. ultimi $N$ valori per il calcolo della deviazione standard, stato di salute, configurazione).
4.  **Resilienza e Coordinamento:** Grazie a Redis, supporta la logica di **Leader Election** (`LeaderKey`) in scenari di alta disponibilità, garantendo che solo un'istanza dell'Hub esegua i task critici di aggregazione e pulizia.

---

## Variabili d'Ambiente dell'Edge Hub

Le variabili d'ambiente definiscono l'identità dell'Hub, i suoi endpoint di comunicazione e i parametri operativi di resilienza e aggregazione.

### 1. Identità e Modalità Operative

| Variabile | Descrizione                                                                                                         | Default / Valori Ammessi |
| :--- |:--------------------------------------------------------------------------------------------------------------------| :--- |
| **`EDGE_MACROZONE`** | **Obbligatoria.** Identifica la macrozona logica a cui appartiene l'Hub (es. `RomaMacro`).                          | Nessun Default |
| **`EDGE_ZONE`** | **Obbligatoria.** Identifica la zona logica specifica (es. `TorVergata`).  | Nessun Default |
| **`HUB_ID`** | Identificatore univoco dell'istanza Hub. Fondamentale per il logging e l'identificazione in ambienti multi-istanza. | **UUID Generato** |
| **`OPERATION_MODE`** | Controlla il ciclo di vita del servizio.                                                                            | `loop` (ciclo infinito) / `once` |
| **`SERVICE_MODE`** | Definisce il ruolo operativo specifico di questa istanza Edge Hub.                                                  | `edge-hub` (default, tutte le funzionalità) / `edge-hub-filter` / `edge-hub-aggregator` / `edge-hub-cleaner` / `edge-hub-configuration` |

### 2. Configurazione MQTT

Il sistema distingue tra i broker usati per la comunicazione **sensore-hub** e **hub-hub (prossimità)**.

| Variabile | Descrizione | Default |
| :--- | :--- | :--- |
| **`MQTT_BROKER_PROTOCOL`** | Protocollo comune (es. `tcp`, `ws`). Utilizzato come fallback per i parametri specifici. | `tcp` |
| **`MQTT_BROKER_ADDRESS`** | Indirizzo comune (es. `localhost` o hostname DNS). Fallback per i parametri specifici. | `localhost` |
| **`MQTT_BROKER_PORT`** | Porta comune (es. `1883`). Fallback per i parametri specifici. | `1883` |
| **`MQTT_SENSOR_BROKER_ADDRESS`** | Indirizzo specifico del broker a cui i Sensor Agents si collegano e da cui l'Hub consuma. | Valore di `MQTT_BROKER_ADDRESS` |
| **`MQTT_HUB_BROKER_ADDRESS`** | Indirizzo specifico del broker (tipicamente il Proximity Hub) a cui l'Edge Hub invia i dati aggregati. | Valore di `MQTT_BROKER_ADDRESS` |

**Nota sui Topic:** I topic MQTT sono composti dinamicamente utilizzando `$EDGE_MACROZONE` e `$EDGE_ZONE` (es. `sensor-data/RomaMacro/TorVergata`) per garantire un'organizzazione gerarchica e isolata per zona.

### 3. Parametri di Resilienza MQTT

Questi parametri controllano il comportamento del client MQTT dell'Hub in caso di disconnessioni o fallimenti di pubblicazione/sottoscrizione.

| Variabile | Descrizione | Default (in secondi/tentativi) |
| :--- | :--- | :--- |
| **`MAX_RECONNECTION_INTERVAL`** | Intervallo massimo tra i tentativi di riconnessione. | $10$ secondi |
| **`MAX_RECONNECTION_TIMEOUT`** | Timeout massimo per un singolo tentativo di riconnessione. | $10$ secondi |
| **`MAX_RECONNECTION_ATTEMPTS`** | Numero massimo di tentativi di riconnessione prima di arrendersi. | $10$ tentativi |
| **`MAX_SUBSCRIPTION_TIMEOUT`** | Timeout per l'operazione di sottoscrizione ai topic. | $5$ secondi |
| **`MESSAGE_PUBLISH_TIMEOUT`** | Timeout per la pubblicazione di un singolo messaggio. | $5$ secondi |
| **`MESSAGE_PUBLISH_ATTEMPTS`** | Numero di tentativi di pubblicazione di un messaggio. | $3$ tentativi |
| **`MESSAGE_CLEANING_TIMEOUT`** | Timeout relativo alla pulizia dei messaggi non inviati. | $5$ secondi |

### 4. Configurazione Redis

| Variabile | Descrizione | Default |
| :--- | :--- | :--- |
| **`REDIS_ADDRESS`** | Indirizzo dell'istanza Redis utilizzata per lo stato condiviso. | `localhost` |
| **`REDIS_PORT`** | Porta dell'istanza Redis. | `6379` |

### 5. Configurazioni di Elaborazione (Costanti)

Queste impostazioni, definite come costanti nell'ambiente, governano l'algoritmo di filtraggio e la logica di gestione dei sensori.

| Nome Costante | Valore              | Descrizione Funzionale |
| :--- |:--------------------| :--- |
| **`AggregationInterval`** | $1$ minuto          | L'intervallo di tempo per cui i dati dei sensori vengono raccolti prima di essere aggregati in un unico record. |
| **`AggregationFetchOffset`** | $-2$ minuti         | Offset di tempo negativo utilizzato per garantire che i dati con ritardi di rete vengano inclusi nell'aggregazione (es. l'aggregazione di un minuto finisce 2 minuti *dopo* l'intervallo). |
| **`LeaderTTL`** | $70$ secondi        | Time-To-Live per la chiave di *Leader Election* in Redis. Se la chiave scade, un'altra istanza può reclamare il ruolo di Leader. |
| **`HistoryWindowSize`** | $100$               | Numero massimo di campioni storici memorizzati in Redis per ogni sensore. Utilizzato come base per l'algoritmo di filtraggio. |
| **`FilteringStdDevFactor`** | $3.0$               | Il fattore di deviazione standard ($\sigma$) utilizzato per il rilevamento degli outlier. Un valore ricevuto è considerato outlier se si trova oltre $\pm 3\sigma$ dalla media della finestra storica. |
| **`UnhealthySensorTimeout`** | $5$ minuti          | Il periodo di tempo dopo il quale un sensore che non invia dati o heartbeat viene marcato come *unhealthy* dal Cleaner Service. |
| **`RegistrationSensorTimeout`** | $6$ ore             | L'intervallo dopo il quale un sensore registrato ma inattivo può essere rimosso dal sistema. |

### 6. Health Check Server

| Variabile | Descrizione | Default |
| :--- | :--- | :--- |
| **`HEALTHZ_SERVER`** | Flag booleano per attivare un server HTTP semplice che risponde allo stato di salute del servizio (`/healthz`). | `false` |
| **`HEALTHZ_SERVER_PORT`** | Porta su cui il server Health Check si mette in ascolto. | `8080` |

-----

## Deploy dell'Edge Hub in Locale con Docker Compose

La distribuzione locale dell'Edge Hub è gestita dal file **`deploy/compose/edge-hub.yaml`** che definisce sia l'applicazione Edge Hub principale che i suoi microservizi di supporto, in particolare la cache **Redis**.

### 1\. Struttura di Docker Compose

Il file Compose definisce l'Edge Hub non come un unico servizio, ma come una **serie di microservizi cooperanti** che condividono la stessa immagine ma eseguono ruoli diversi:

* **Servizi Principali:** `zone-hub-filter`, `zone-hub-cleaner`, `zone-hub-aggregator`, `zone-hub-configurator`. Questi servizi sono scalati orizzontalmente per garantire tolleranza ai guasti e parallelismo nell'elaborazione.
* **Dipendenza:** `zone-hub-redis-cache`, che fornisce la cache di stato condiviso necessaria a tutti i servizi Hub.
* **Rete:** Viene creata una rete *bridge* isolata (`zone-hub-${EDGE_ZONE}-bridge`) per il traffico interno.

Il file sfrutta le estensioni YAML (`x-zone-hub-base`, `x-zone-hub-env`) per ridurre la ridondanza, ereditando le configurazioni base (come `image`, `build`, `healthcheck` e `restart`) e differenziando solo le variabili d'ambiente critiche (come `HUB_ID` e `SERVICE_MODE`).

### 2\. Variabili d'Ambiente Obbligatorie

Il corretto funzionamento del file Compose dipende dalla risoluzione dinamica di quattro variabili essenziali, utilizzate per l'identificazione, l'indirizzamento e la creazione di risorse:

| Variabile | Necessità | Utilizzo nel Compose |
| :--- | :--- | :--- |
| **`EDGE_MACROZONE`** | Naming e Indirizzamento MQTT | Definisce il nome del container e l'indirizzo DNS del broker. |
| **`EDGE_ZONE`** | Naming e Naming delle Risorse | Cruciale per nominare container, volumi, reti e per l'indirizzo DNS del broker. |
| **`REGION`** | Indirizzamento MQTT | Utilizzata per completare l'hostname del broker MQTT (es. `zone.macrozone.mqtt-broker.REGION.sensor-continuum.local`). |
| **`REDIS_PORT`** | Mappatura Porta | Definisce la porta host per accedere a Redis (default `6379` se non specificato altrimenti). |

Senza queste variabili, Docker Compose non sarà in grado di popolare correttamente le estensioni YAML e i nomi delle risorse, portando al fallimento del deployment.

### 3\. Preparazione e Esecuzione del Deploy

Il deployment in locale richiede due passaggi fondamentali: la creazione del volume e l'avvio tramite Compose.

#### A. Creazione del Volume Persistente

Poiché il volume `zone-hub-cache` è definito come **`external: true`** nel file Compose, deve essere creato **manualmente** prima dell'avvio per garantirne la persistenza e il riutilizzo tra diversi cicli di deploy:

**Esempio di creazione del volume per la zona `TorVergata` (assumendo `$EDGE_ZONE=TorVergata`):**

```bash
docker volume create zone-hub-TorVergata-cache-data
```

#### B. Avvio con Docker Compose

Dopo aver creato il volume, l'avvio viene eseguito tramite il comando standard di Docker Compose:

```bash
# Esempio di avvio
docker compose -f edge-hub.yaml up -d
```

> **⚠️ NOTA OPERATIVA**
>
> **File d'Ambiente:**
> Per standardizzare il deployment, è cruciale definire le variabili d'ambiente (almeno `EDGE_MACROZONE`, `EDGE_ZONE` e `REGION`) in un file separato (es. **`.env`**). Il file Compose le utilizza per risolvere tutti i nomi dei container, i nomi di rete e l'indirizzo del broker MQTT.
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
>
> **Gestione del Volume Persistente:**
> Il volume `zone-hub-cache` è definito come `external: true` nel file Compose, il che significa che deve essere creato **manualmente prima del deploy** per garantirne la persistenza:
>
> ```bash
> docker volume create zone-hub-TorVergata-cache-data
> ```

-----

## Deploy su AWS dell'Edge Hub

L'Edge Hub (e le sue dipendenze, inclusa la cache Redis) è distribuito sull'istanza **EC2** dedicata alla zona tramite **Docker Compose**, con l'orchestrazione gestita da **AWS CloudFormation** e lo script Bash **`deploy_zone.sh`**.

---

## ⚠️ Prerequisiti di Deployment

La fase di **Deploy su AWS**, gestita dallo script **`deploy_zone.sh`**, è l'ultima di una sequenza di provisioning che stabilisce l'infrastruttura di rete a livello regionale e di macrozona.

**Prima di eseguire lo script di deployment della Zona, è indispensabile che le seguenti risorse AWS siano già state create e configurate attraverso i rispettivi script di provisioning di Livello Superiore (`deploy_region.sh` e `deploy_macrozone.sh`):**

| Risorsa                          | Livello di Creazione | Variabile nello script | Riferimento Logico |
|:---------------------------------| :--- | :--- | :--- |
| **Virtual Private Cloud (VPC)**  | Regionale | `$VPC_ID` | Identificata da **`$REGION-vpc`** |
| **Subnet Pubblica**              | Macrozona | `$SUBNET_ID` | Identificata da **`$REGION-$MACROZONE-subnet`** |
| **Security Group**               | Macrozona | `$SECURITY_GROUP_ID` | Identificata da **`$REGION-$MACROZONE-sg`** |
| **Route Table Pubblica**         | Regionale | `$ROUTE_TABLE_ID` | Identificata da **`$REGION-vpc-public-rt`** |
| **Route 53 Hosted Zone Privata** | Regionale | `$HOSTED_ZONE_ID` | Identificata da **`$REGION.sensor-continuum.local`** |

Lo script `deploy_zone.sh` non crea le risorse di rete; piuttosto, **le cerca e le recupera** tramite le funzioni `find_vpc_id`, `find_subnet_id`, etc. Se queste risorse non esistono o non corrispondono ai tag di denominazione attesi (`$VPC_NAME`, `$SUBNET_NAME`, etc.), lo script fallirà, non potendo lanciare lo stack CloudFormation.

**Sequenza di Deployment Rigorosa:**

1.  **Deploy Livello Regionale:** Esecuzione dello script per creare la **VPC**, la **Hosted Zone** e la **Route Table Pubblica**.
2.  **Deploy Livello Macrozona:** Esecuzione dello script per creare la **Subnet** e il **Security Group** associati.
3.  **Deploy Livello Zona/Edge:** Esecuzione dello script **`deploy_zone.sh`** per installare l'istanza EC2, i container (Edge Hub, Sensor Agents) e configurare i record DNS specifici (`$SENSOR_MQTT_BROKER_HOSTNAME`, ecc.) che puntano all'IP dell'istanza.

Assicurarsi che la catena di provisioning dell'infrastruttura AWS sia stata completata correttamente è la condizione necessaria per un deploy di successo dell'Edge Zone.

---

## 1\. Preparazione: Caricamento Asset su S3

Prima di eseguire il deployment, tutti gli asset necessari (script di installazione, file Docker Compose, file di servizio Systemd e script di analisi) devono essere caricati in un **Bucket S3** dedicato.

Questo processo è documentato in [`setup_bucket.md`](./setup_bucket.md) e gestito dallo script **`deploy/scripts/setup_bucket.sh`**.

---

### 2. Deploy Edge Zone tramite CloudFormation

Lo script **`deploy_zone.sh`** orchestra il processo di deployment.

#### Funzionamento del Template CloudFormation

Il template CloudFormation esegue le seguenti operazioni cruciali per installare l'ambiente Edge Hub:

* **Creazione EC2 Instance (`ServicesInstance`)**: Avvia l'istanza EC2 che ospiterà il container Edge Hub e Redis.
* **Configurazione DNS (Route53)**: Crea un set di **Record CNAME** all'interno della Hosted Zone privata. Questi record sono essenziali per risolvere correttamente i nomi di dominio complessi dei broker MQTT (es. `$ZONE.$MACROZONE.hub.mqtt-broker.$HOSTED_ZONE_NAME`) verso l'IP privato dell'istanza EC2, permettendo la comunicazione inter-livello (Edge Hub $\leftrightarrow$ Proximity Hub).
* **Installazione Edge Hub via `UserData`**: Il blocco `UserData` (eseguito al primo boot) scarica il file **`.env`** specifico da S3 ed esegue lo script di deployment **`deploy_edge_services.sh`**, che gestirà l'avvio di tutti i container.

#### Utilizzo dello Script di Deployment

Lo script accetta parametri posizionali obbligatori per definire la gerarchia logica e opzioni facoltative per l'ambiente AWS:

| Opzione | Parametro | Descrizione                                                                                                                                                       | Valore di Default |
| :--- | :--- |:------------------------------------------------------------------------------------------------------------------------------------------------------------------| :--- |
| **Modalità di Deploy** | `--deploy=localstack` | Forza il deployment su **LocalStack**. Se non specificata, il deployment avviene sul cloud AWS reale.           | `aws` |
| **Regione AWS** | `--aws-region REGION` | Specifica la regione geografica AWS dove verranno create le risorse (e dove verranno cercate le risorse preesistenti come VPC e Subnet).                          | `us-east-1` |
| **Tipo di Istanza EC2** | `--instance-type TYPE` | Definisce il tipo di istanza EC2 su cui verranno eseguiti i container Docker (Edge Hub e Sensor Agents). Questo parametro è cruciale per dimensionare le risorse. | `t3.small` |

---

### 3. Dettagli Operativi del Deployment Script in EC2

Lo script Bash **`deploy_edge_services.sh`**, eseguito all'interno dell'istanza EC2, è interamente responsabile dell'installazione di Docker, Docker Compose e dell'avvio dell'Edge Hub.

#### Sequenza di Avvio dell'Edge Hub

1.  **Caricamento Variabili**: Carica le variabili d'ambiente dal file **`.env`** precedentemente scaricato da S3.
2.  **Download Docker Compose**: Scarica da S3 il file di configurazione **`edge-hub.yaml`**.
3.  **Avvio Edge Hub (Infrastruttura di Base)**: Utilizza `docker compose` per avviare i servizi definiti:
    * **Cache Redis**: Viene avviata la cache di stato, essenziale per la logica di filtering, aggregation e leader election.
    * **Microservizi Hub**: Vengono avviate le istanze scalate dei microservizi Hub (`filter`, `aggregator`, `cleaner`, `configurator`) utilizzando un nome di progetto Docker Compose dedicato (es. `-p edge-hub`).

#### Gestione Operativa e Resilienza

Lo script integra funzionalità di gestione essenziali per un nodo Edge robusto:

| Funzionalità | Dettaglio Operativo |
| :--- | :--- |
| **Servizio Systemd** | Configura il servizio **`sc-deploy.service`** che esegue automaticamente lo script di deployment all'avvio del sistema. Ciò garantisce che l'Edge Hub e Redis vengano ripristinati correttamente dopo un riavvio dell'istanza EC2. |
| **Simulatore di Ritardi (Latenza)** | Esegue lo script **`init-delay.sh`** per simulare condizioni di rete reali, applicando una latenza configurabile sull'interfaccia di rete dell'istanza EC2. |
| **Script di Analisi** | Scarica lo script **`analyze_failure.sh`**. Questo strumento è progettato per analizzare a posteriori le performance dell'Edge Hub (es. Missing Rate, errore nel rilevamento Outlier) utilizzando i log dei container. |