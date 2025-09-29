# Istruzioni per il Sensor Agent

Il Sensor Agent rappresenta il livello più basso del sistema. Non è un sensore fisico, ma un simulatore che legge dati storici da un dataset CSV e li pubblica periodicamente come misurazioni in un formato standardizzato verso l'Edge Hub tramite il protocollo MQTT.

Il suo funzionamento ad alto livello include:
1.  **Caricamento Dati:** Legge un dataset predefinito.
2.  **Identificazione:** Genera un ID univoco e si registra nelle macrozone/zone di appartenenza.
3.  **Simulazione Temporale:** Utilizza un *offset* per simulare l'invio dei dati come se fossero raccolti in tempo reale, partendo da un punto indietro nel .
4.  **Trasmissione:** Pubblica le misurazioni sul *topic* MQTT dedicato, garantendo la connettività e gestendo i tentativi di riconnessione al broker.

-----

## Variabili d'Ambiente per la Configurazione

La configurazione del Sensor Agent avviene interamente tramite variabili d'ambiente, suddivise in quattro categorie principali:

### A\. Parametri di Identificazione e Ambiente

Definiscono l'appartenenza gerarchica del Sensor Agent e il suo ID.

| Variabile                 | Descrizione                                                                                                 | Valori Ammessi (Default)             |
|:--------------------------|:------------------------------------------------------------------------------------------------------------|:-------------------------------------|
| **`EDGE_MACROZONE`**      | **Obbligatoria.** Identificativo della macrozona (Edge di secondo livello) a cui il sensore appartiene.     | Stringa (es. `RegioneA`)             |
| **`EDGE_ZONE`**           | **Obbligatoria.** Identificativo della zona (Edge di primo livello) a cui il sensore appartiene.            | Stringa (es. `Zona1`)                |
| **`SENSOR_ID_GENERATOR`** | Metodo per generare l'ID univoco del sensore.                                                               | **`uuid`** (Default), **`hostname`** |
| **`SENSOR_ID`**           | ID univoco del sensore. Se non specificato, viene generato automaticamente in base a `SENSOR_ID_GENERATOR`. | Stringa (es. UUID generato)          |

### B\. Parametri di Simulazione

Queste variabili controllano l'origine e la natura dei dati simulati.

| Variabile                         | Descrizione                                                                                 | Valori Ammessi (Default)                                               |
|:----------------------------------|:--------------------------------------------------------------------------------------------|:-----------------------------------------------------------------------|
| **`SENSOR_LOCATION`**             | Definisce il contesto del sensore.                                                          | **`indoor`** (Default), **`outdoor`**                                  |
| **`SENSOR_TYPE`**                 | Definisce la grandezza fisica misurata.                                                     | **`temperature`** (Default), **`humidity`**, **`pressure`**            |
| **`SIMULATION_SENSOR_REFERENCE`** | Riferimento al modello di sensore fisico simulato, cruciale per l'interpretazione dei dati. | **`bmp280`** (Default) o altri 17 modelli (es. `DHT22`, `SPS30`, ecc.) |
| **`SIMULATION_VALUE_COLUMN`**     | Nome della colonna nel file CSV da cui leggere il valore del sensore.                       | **`[SENSOR_TYPE]`** (Default)                                          |
| **`SIMULTION_TIMESTAMP_COLUMN`**  | Nome della colonna nel file CSV contenente i timestamp.                                     | **`timestamp`** (Default)                                              |
| **`SIMULATION_SEPARATOR`**        | Carattere separatore utilizzato nel file CSV.                                               | **`;`** (Default)                                                      |
| **`SIMULATION_TIMESTAMP_FORMAT`** | Formato Go (Layout di riferimento `2006-01-02T15:04:05`) del timestamp nel CSV.             | **`2006-01-02T15:04:05`** (Default)                                    |
| **`SIMULATION_OFFSET_DAY`**       | Numero di giorni da sottrarre alla data corrente per iniziare la simulazione storica.       | Intero non negativo (**2** Default)                                    |

### C\. Parametri di Comunicazione

Controllano la connessione al broker MQTT dell'Edge Hub.

| Variabile                       | Descrizione                                                       | Valori Ammessi (Default)                |
|:--------------------------------|:------------------------------------------------------------------|:----------------------------------------|
| **`MQTT_BROKER_PROTOCOL`**      | Protocollo del broker MQTT.                                       | `tcp` (Default)                         |
| **`MQTT_BROKER_ADDRESS`**       | Indirizzo IP/Hostname del broker MQTT.                            | `mosquitto` (Default, tipico in Docker) |
| **`MQTT_BROKER_PORT`**          | Porta di connessione del broker MQTT.                             | `1883` (Default)                        |
| **`MAX_RECONNECTION_INTERVAL`** | Intervallo massimo (in secondi) tra i tentativi di riconnessione. | Intero positivo (**10** Default)        |
| **`MAX_RECONNECTION_TIMEOUT`**  | Timeout massimo (in secondi) per i tentativi di riconnessione.    | Intero positivo (**10** Default)        |
| **`MAX_RECONNECTION_ATTEMPTS`** | Numero massimo di tentativi di riconnessione.                     | Intero positivo (**10** Default)        |
| **`MESSAGE_PUBLISH_TIMEOUT`**   | Timeout (in secondi) per l'invio di un singolo messaggio MQTT.    | Intero positivo (**5** Default)         |

### D\. Parametri di Logging e Health Check

| Variabile                 | Descrizione                                                                   | Valori Ammessi (Default)                      |
|:--------------------------|:------------------------------------------------------------------------------|:----------------------------------------------|
| **`HEALTHZ_SERVER`**      | Abilita un server HTTP per il controllo dello stato di salute (Health Check). | **`false`** (Default), **`true`**             |
| **`HEALTHZ_SERVER_PORT`** | Porta su cui il server Health Check si mette in ascolto.                      | **`8080`** (Default)                          |
| **`LOG_LEVEL`**           | Livello di dettaglio per l'output del logger.                                 | `error` (Default), `warning`, `info`, `debug` |

-----

## Deploy in Locale del Sensor Agent

Il Sensor Agent è distribuito come immagine Docker (`fmasci/sc-sensor-agent:latest`) ed è configurato interamente tramite variabili d'ambiente per connettersi al broker MQTT del livello Edge Hub.

### Avvio di un Singolo Sensor Agent

È possibile eseguire un singolo agente utilizzando l'immagine Docker pre-buildata o costruendola localmente. Il template base di Docker è disponibile nel file [`sensor-agent.Dockerfile`](../../deploy/docker/sensor-agent.Dockerfile).

#### Esempio di Docker Compose per Singolo Sensore

Il seguente blocco mostra il template utilizzato in un file `docker-compose.yml`. Per l'esecuzione, è necessario sostituire i placeholder o definire le variabili d'ambiente.

```yaml
services:
  sensor-agent-01:
    image: fmasci/sc-sensor-agent:latest
    # Se necessario, la build utilizza il Dockerfile:
    # build:
    #   context: ../..
    #   dockerfile: deploy/docker/sensor-agent.Dockerfile
    environment:
      - EDGE_MACROZONE=RomaMacro
      - EDGE_ZONE=TorVergata
      - SENSOR_ID=sensor-agent-01
      - SENSOR_TYPE=temperature
      - SENSOR_LOCATION=outdoor
      - SIMULATION_SENSOR_REFERENCE=ds18b20
      - MQTT_BROKER_ADDRESS=mosquitto
      - HEALTHZ_SERVER=true
      - HEALTHZ_SERVER_PORT=8080
    healthcheck:
      test: [ "CMD", "curl", "-f", "http://localhost:8080/healthz" ]
      interval: 60s
      timeout: 30s
      retries: 10
    restart: unless-stopped
    #extra_hosts:
    #  - "floor-001.sensor.mqtt-broker.${REGION}.sensor-continuum.local:192.168.0.10"
```

**Comando di Avvio:**

```bash
docker compose up -d sensor-agent-01
```

**Risoluzione del Nome Host del Broker:**

Il template utilizza un indirizzo parametrico per connettersi al broker. Per risolvere questo nome di dominio e dirigerlo correttamente verso l'host desiderato, si hanno due opzioni in un ambiente Docker locale:

* **Utilizzo di `extra_hosts`:** Decommentare e configurare la sezione `extra_hosts` nel file Docker Compose, mappando l'indirizzo del broker all'IP appropriato.
* **Modifica del File `/etc/hosts`:** Se necessario, è possibile modificare il file `/etc/hosts` del sistema operativo host per puntare il record DNS utilizzato direttamente all'indirizzo IP del container o del servizio che ospita il broker.

### Deployment di una Fleet di Sensori

Per simulare l'ambiente distribuito con una molteplicità di sensori, si ricorre al template [`sensor-agent.template.yml`](../../deploy/compose/sensor-agent.template.yml) e allo script Go [`generate-sensor-agent.go`](../../deploy/compose/generate-sensor-agents.go), entrambi presenti nella cartella [`deploy/compose`](../../deploy/compose/). Lo script genera automaticamente un file Docker Compose con il numero di agenti richiesto e parametri di simulazione casuali.

#### Procedura

1.  **Esecuzione dello Script di Generazione:**

    Eseguire lo script Go specificando il numero di sensori (es. 2) che si desidera creare:

    ```bash
    go run generate-sensor-agents.go 2
    ```

    **Output Ricevuto:**

    ```
    File generato: ./sensor-agent.generated_2.yml
    ```

    Lo script crea un nuovo file (es. `sensor-agent.generated_2.yml`) con la struttura YAML necessaria, randomizzando i parametri di simulazione (`SENSOR_TYPE`, `SENSOR_LOCATION`, `SIMULATION_SENSOR_REFERENCE`) per ciascun agente.

2.  **Avvio della Fleet di Sensori:**

    Avviare tutti i sensori definiti nel file generato:

    ```bash
    docker compose -f sensor-agent.generated_2.yml up -d
    ```

In caso di malfunzionamenti, la direttiva **`restart: unless-stopped`** assicura che Docker Compose tenti di riavviare automaticamente i container dei sensori.

> **⚠️ NOTA OPERATIVA**
>
> **File d'Ambiente:**
> Per standardizzare il deployment, si consiglia di definire le variabili d'ambiente comuni in un file separato (es. **`.env`**).
>
> Per caricare un file d'ambiente specifico (es. `simulazione.env`), utilizzare l'opzione `--env-file`:
>
> ```bash
> docker compose -f sensor-agent.generated_2.yml --env-file simulazione.env up -d
> ```
>
> **Nomi di Progetto:**
> Poiché al livello Edge appartengono diversi servizi (Edge Hub e Sensor Agents), è cruciale nominare esplicitamente il progetto Docker Compose. Questo facilita l'identificazione, l'avvio e l'arresto dei container che fanno parte di quella specifica zona.
>
> Si raccomanda di utilizzare il flag **`-p`** o **`--project-name`** con un nome che rifletta la regione, la macrozona e la zona (es. `Lazio_RomaMacro_TorVergata`).
>
> **Esempio di Deploy Completo:**
>
> ```bash
> docker compose -f sensor-agent.generated_2.yml --env-file simulazione.env -p Lazio_RomaMacro_TorVergata up -d
> ```

-----

## Deploy su AWS del Sensor Agent

La fase di deploy su AWS per l'ambiente Edge Hub e i Sensor Agents viene gestita tramite un approccio infrastrutturale automatico basato su AWS CloudFormation e l'utilizzo di script ausiliari e asset ospitati su Amazon S3.

### ⚠️ Prerequisiti di Deployment

La fase di *Deploy su AWS, gestita dallo script [`deploy_zone.sh`](../../deploy/scripts/deploy_zone.sh), è l'ultima di una sequenza di provisioning che stabilisce l'infrastruttura di rete a livello regionale e di macrozona.

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

Il deployment dell'intera zona Edge (che comprende l'istanza EC2, la configurazione di Docker e i Sensor Agents) è orchestrato dallo script [`deploy_zone.sh`](../../deploy/scripts/deploy_zone.sh) che utilizza il template [`zone/services.yaml`](../../deploy/cloudformation/zone/services.yaml).

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

Lo script Bash [`deploy_edge_services.sh`](../../deploy/scripts/deploy/deploy_edge_services.sh) viene eseguito all'interno dell'istanza EC2 tramite `UserData` ed è responsabile dell'avvio dei servizi di Sensor Agent.

##### Sequenza di Avvio dei Servizi

1.  **Caricamento Variabili**: Carica le variabili d'ambiente dal file **`.env`** precedentemente scaricato da S3.
2.  **Download Docker Compose**: Scarica da S3 la configurazione generata per i sensori (`sensor-agent.generated_N.yml`, con N=50 di default).
3.  **Avvio Sensor Agents**: Utilizza `docker-compose -p sensors` per avviare i simulatori di sensori, isolandoli in un progetto separato per facilitare la gestione.

##### Gestione Operativa e Resilienza

Lo script [`deploy_edge_services.sh`](../../deploy/scripts/deploy/deploy_edge_services.sh) integra diverse funzionalità per la gestione e la simulazione di un ambiente Edge robusto:

* **Simulatore di Ritardi**: Scarica ed esegue lo script [`init-delay.sh`](../../deploy/scripts/inits/init-delay.sh) per simulare condizioni di rete reali. Applica una latenza di rete configurabile (di default impostata a `${NETWORK_DELAY:-200ms}`) sull'interfaccia di rete dell'istanza EC2.
* **Cron Job per Resilienza**: Configura un Cron Job che riavvia periodicamente i container dei sensori alle 3:00.
* **Servizio Systemd**: Configura il servizio [`sc-deploy.service`](../../deploy/scripts/services/sc-deploy.service.template) per eseguire automaticamente lo script di deployment all'avvio del sistema. Ciò garantisce che i servizi vengano ripristinati correttamente dopo un riavvio dell'istanza EC2.
* **Script di Analisi**: Scarica lo script [`analyze_failure.sh`](../../deploy/scripts/performance/analyze_failure.sh). Questo strumento è progettato per analizzare i log dei container, confrontando i messaggi ricevuti con i messaggi attesi e calcolando parametri di performance chiave come il *Missing Rate* e l'*errore nel rilevamento degli Outlier*.
