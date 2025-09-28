# Guida al Deployment del Sensor Continuum

Il sistema **Sensor Continuum** è un'architettura distribuita progettata per operare in un ambiente di **Compute Continuum**, distribuendo il carico computazionale su cinque livelli (dal sensore al cloud). Questa documentazione fornisce le istruzioni per il deployment del sistema sui diversi livelli.

-----

## Prerequisiti e Dipendenze

Per eseguire la build, il test e il deployment dell'applicazione, è necessario disporre dei seguenti strumenti:

| Dipendenza                                 | Utilizzo                                                                                              |
|:-------------------------------------------|:------------------------------------------------------------------------------------------------------|
| **Go** (Golang)                            | Necessario per la compilazione dei microservizi Hub e Sensor Agent, qualora la build fosse richiesta. |
| **Docker / Docker Compose**                | Obbligatorio per il deployment locale e per l'esecuzione del sistema in containers.                   |
| **AWS CLI**                                | Necessario per interagire con i servizi AWS (S3, EC2, CloudFormation, ...) e gestire le credenziali.  |
| **AWS SAM** (Serverless Application Model) | Utilizzato per il deployment dei componenti Serverless nel Cloud.                                     |

---

## Deployment

| Modalità di Deploy          | Vantaggi                                                                                                | Svantaggi / Prerequisiti Operativi                                                                                                                                                                   |
|:----------------------------|:--------------------------------------------------------------------------------------------------------|:-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **AWS (Consigliata)**       | **Automazione Completa** (CloudFormation, Route53), provisioning dinamico della rete e degli hostnames. | Richiede l'accesso e la configurazione dell'ambiente AWS.                                                                                                                                            |
| **Locale (Docker Compose)** | Ideale per test e sviluppo rapido.                                                                      | Richiede la **configurazione manuale del sistema host** (modifica del file `/etc/hosts`) per risolvere gli hostname complessi dei broker interni (es. `kafka-broker.region.sensor-continuum.local`). |

Il deploy segue l'ordine gerarchico del Compute Continuum, partendo dal Cloud (Intermediate) verso l'Edge (Sensor Agent).

| Livello del Continuum | Componente                             | Riferimento Istruzioni                                 | 
|:----------------------|:---------------------------------------|:-------------------------------------------------------|
| CLOUD                 | API Gateway, Lambda, Site, ...         | [`setup_cloud.md`](./setup_cloud.md)                   |
| INTERMEDIATE FOG      | Intermediate Fog Hub, Kafka, Databases | [`intermediate_fog_hub.md`](./intermediate_fog_hub.md) |
| PROXIMITY FOG         | Proximity Fog Hub, MQTT Broker         | [`proximity_fog_hub.md`](./proximity_fog_hub.md)       |
| EDGE FOG              | Edge Hub                               | [`edge_hub.md`](./edge_hub.md)                         |
| EDGE DEVICE           | Sensor Agent                           | [`sensor_agent.md`](./sensor_agent.md)                 |

### Deploy automatizzato su AWS

Il deployment su AWS è gestito da un set di script Bash situati nella directory **`deploy/scripts`**. Questo approccio è **totalmente automatizzato** e **molto più semplice** rispetto al deployment in locale.

Gli script utilizzano **AWS CloudFormation** e **Docker Compose** per l'orchestrazione, seguendo l'ordine gerarchico del Continuum:

1.  **`setup_bucket.sh`**: Prepara il bucket S3 di supporto per tutti gli asset.
2.  **`setup_dns.sh`**: Configura la zona DNS pubblica globale.
3.  **`setup_lambda.sh`**: Configura la rete, l'API Gateway e gestisce i certificati SSL.
4.  **`setup_cloud_db.sh`**: Deploya il database cloud per i metadati.
5.  **`setup_site.sh`**: Configura l'hosting del sito web (Frontend).
6.  **`deploy_region.sh`**: Crea l'infrastruttura Intermediate Fog Hub.
7.  **`deploy_macrozone.sh`** e **`deploy_zone.sh`**: Creano i livelli Proximity ed Edge.

#### Fase Cloud: Componenti e Dipendenza DNS

Il livello **CLOUD** gestisce l'interfaccia pubblica, la logica serverless e i metadati. È composto dalla **Zona DNS Pubblica**, l'**Infrastruttura di Rete Lambda**, l'**API Gateway**, il **Database Cloud dei Metadati** e l'**Hosting del Sito Web**.

##### A. Setup DNS Pubblico (`setup_dns.sh`)

Questo script è il punto di partenza e stabilisce l'autorità DNS per il dominio del progetto.

1.  **Deploy Stack CloudFormation (`public-dns.yaml`):** Crea lo stack **`sc-public-dns`**, che istanzia la **Zona Pubblica di Route 53** (es. `sensor-continuum.it`).
2.  **Configurazione Name Server (NS):** Lo script estrae i **Name Server** generati da AWS. Questi valori devono essere copiati e configurati **manualmente** presso il *registrar* del dominio per delegare la gestione del DNS ad AWS.


> ##### ⚠️ Dipendenza Critica
>
> L'unica vera dipendenza per il corretto funzionamento degli script è la creazione della **Zona Pubblica dei DNS** tramite `setup_dns.sh`. Questo script esporta l'**ID della Hosted Zone Pubblica** (`sensor-continuum-hosted-zone-id`), essenziale per configurare correttamente tutti i record DNS (API Gateway, Database, Sito Web).

##### B. Setup Lambda (`setup_lambda.sh`)

Lo script combina il setup della **Rete VPC** per le Lambda con il deploy dell'**API Gateway**.

1.  **Setup Rete Lambda (`lambda-network.yaml`)**: Crea la **VPC** dedicata, le **Subnet Private** e il **NAT Gateway**.
2.  **Deploy API Gateway (`lambda-api.yaml`)**: Crea l'**API Gateway**, il **Certificato ACM** e il **Dominio Personalizzato** (`api.sensor-continuum.it`), gestendo in automatico la **Validazione DNS** e la configurazione del CNAME finale nella Hosted Zone Pubblica.

##### C. Deploy Database Cloud dei Metadati (`setup_cloud_metadata_db.sh`)

Questo passaggio implementa il database **Aurora PostgreSQL** centrale per i metadati globali.

1.  **Deploy Stack CloudFormation (`cloud-db.yaml`):** Crea il cluster Aurora PostgreSQL, ricevendo in *parameter overrides* gli ID di **VPC** e **Subnet Private** dallo step precedente.
2.  **Configurazione Hostname Pubblici (DNS):** Lo script crea i record **CNAME** per gli Endpoint Writer e Reader di Aurora nella Zona Pubblica.
3.  **Inizializzazione Schema (SQL):** Esegue i file SQL **`init-cloud-metadata-db.sql`** e **`init-metadata.sql`** per popolare lo schema.

##### D. Setup Sito Web (Frontend) (`setup_site.sh`)

Questo script configura l'hosting per il sito web frontend che interagirà con l'API Gateway.

1.  **Configurazione Hosting Amplify:** Crea o verifica l'esistenza dell'app **AWS Amplify** per l'hosting statico.
2.  **Associazione Dominio:** Associa il dominio personalizzato (`sensor-continuum.it`) e il sottodominio `www` all'app Amplify.
3.  **Setup Bucket S3:** Crea un bucket S3 di supporto.

> **⚠️ AVVERTENZA**
> 
> Lo script `setup_site.sh` prepara l'ambiente Amplify e configura il dominio personalizzato. Tuttavia, il **deployment effettivo** del codice del sito web **non può essere automatizzato completamente** tramite questo script Bash. Per deployare il contenuto nel servizio di hosting Amplify, è necessario utilizzare l'**interfaccia web di AWS Amplify** o collegare l'app a un repository Git.

##### E. Deploy delle Funzioni Lambda

Il deployment delle funzioni **Lambda** che gestiscono la logica **Serverless** dell'API è affidato allo script **`deploy_lambda.sh`**, che orchestra l'intero processo utilizzando l'**AWS Serverless Application Model (SAM)**. Lo script è chiamato con quattro parametri essenziali: il **nome dello Stack CloudFormation**, la **cartella del codice**, il **nome della Funzione Lambda** e la **Route API Path** corrispondente. Questo processo automatizza il *build*, l'iniezione dei parametri di rete (VPC/Security Group) e l'integrazione finale con l'**API Gateway HTTP**.

Di seguito sono riportati alcuni esempi delle chiamate di deployment che coprono i diversi livelli di accesso ai dati:

* **Endpoint Lista Regioni:** Per un semplice accesso ai metadati:

  ```bash
  ./deploy_lambda.sh region-list-stack region regionList "/region/list"
  ```

* **Dati Aggregati per Macrozona:** Per recuperare i dati statistici consolidati di un livello intermedio:

  ```bash
  ./deploy_lambda.sh macrozone-data-aggregated-name-stack macrozone macrozoneDataAggregatedName "/macrozone/data/aggregated/{region}/{macrozone}"
  ```

* **Dati Sensori Raw (Livello Zona):** Per la query più granulare, che include tutti i parametri gerarchici:

  ```bash
  ./deploy_lambda.sh zone-sensor-data-raw-stack zone zoneSensorDataRaw "/zone/sensor/data/raw/{region}/{macrozone}/{zone}/{sensor}"
  ```

Questa metodologia garantisce che ogni Lambda sia correttamente creata e mappata al suo specifico endpoint REST, completando l'architettura del livello Cloud.


-----

### Fase Fog ed Edge: Deployment dell'Infrastruttura Distribuita

Questa fase si occupa del deployment dei nodi distribuiti **Intermediate**, **Proximity** ed **Edge** per l'elaborazione a bassa latenza, costituendo la rete *Compute Continuum*. Il provisioning dell'infrastruttura (macchine **EC2**, reti, DNS) è gestito da **AWS CloudFormation**, mentre l'esecuzione dei microservizi applicativi su ogni istanza avviene tramite **Docker Compose**. Crucialmente, un servizio **Systemd** è configurato su ogni nodo per garantire l'avvio e il ripristino automatico dei servizi, assicurando la resilienza.

#### A. Livello Regionale (Intermediate Fog Hub - IFH)

Il nodo IFH è avviato dallo script **`deploy_region.sh`**.

1.  **Infrastruttura:** Crea la **VPC** dedicata e la **Private Hosted Zone** DNS regionale.
2.  **Servizi:** Istanziato un **Kafka Broker** e due **Databases** separati (**Measurement** e **Metadata**) contestualmente all'hub.
3.  **Gestione:** Utilizza **`init-delay.sh`** per simulare la latenza di rete e Systemd per il ripristino automatico.

#### B. Livello Macrozona (Proximity Fog Hub - PFH)

Il PFH viene avviato dallo script **`deploy_macrozone.sh`**.

1.  **Infrastruttura:** Crea le risorse di rete specifiche per la Macrozona.
2.  **Servizi:** Avvia l'istanza EC2 che ospita anche un **Broker MQTT** locale e un **PostgreSQL locale** come cache persistente (**Outbox**).
3.  **Gestione:** Crea un **Record A** in Route 53 per rendere l'endpoint MQTT risolvibile e utilizza Systemd per il riavvio automatico.

#### C. Livello Zona (Edge Fog Hub + Sensor Agents)

Questo è il livello più vicino alla fonte dei dati, avviato dallo script **`deploy_zone.sh`**.

1.  **Infrastruttura:** Avvia l'istanza **Edge Hub (EFH)** e configura i **Record CNAME** DNS per l'instradamento dei sensori.
2.  **Servizi:** L'EFH avvia un'istanza **Redis locale** come cache di stato condivisa per l'elaborazione. Avvia anche i **Sensor Agents** (simulatori di sensori).
3.  **Test e Resilienza:** Configura un **Cron Job** per simulare lo *churn* dei dispositivi e scarica lo script **`analyze_failure.sh`** per l'analisi post-mortem delle performance.

-----

#### Esempio di Deploy Completo (Singola Zona)

Il seguente esempio mostra la sequenza di comandi necessaria per deployare l'intero sistema Compute Continuum (Fase Cloud + Infrastruttura Fog) per una singola zona, garantendo l'operatività dall'Edge al Cloud.

```bash
#!/bin/bash

# ==================================
# FASE CLOUD
# ==================================

# 1. SETUP INIZIALE: Preparazione degli asset su AWS S3
./setup_bucket.sh
if [ $? -ne 0 ]; then
  echo "Errore nella creazione del bucket S3."
  exit 1
fi

# 2. CONFIGURAZIONE DNS PUBBLICO (CLOUD)
./setup_dns.sh
if [ $? -ne 0 ]; then
  echo "Errore nella configurazione DNS pubblica."
  exit 1
fi

# 3. SETUP LAMBDA (CLOUD)
# Rete VPC per Lambda, API Gateway, Certificati SSL.
./setup_lambda.sh
if [ $? -ne 0 ]; then
  echo "Errore nel deploy API e Lambda."
  exit 1
fi

# 4. DEPLOY DATABASE CLOUD (CLOUD)
# Cluster Aurora PostgreSQL per i metadati globali.
./setup_cloud_metadata_db.sh
if [ $? -ne 0 ]; then
  echo "Errore nel deploy del Database Cloud dei Metadati."
  exit 1
fi

# 5. SETUP SITO WEB (CLOUD)
# Configura l'hosting del frontend. (Il deploy del codice va fatto da console web).
./setup_site.sh
if [ $? -ne 0 ]; then
  echo "Errore nel setup del sito web."
  exit 1
fi

# ==================================
# FASE FOG ED EDGE
# ==================================

# 6. DEPLOY LIVELLO REGIONALE (INTERMEDIATE FOG HUB)
# Crea il nodo principale: VPC, MSK (Kafka) e Database Regionale.
./deploy_region.sh region-001 --aws-region us-east-1
if [ $? -ne 0 ]; then
  echo "Errore nella creazione della regione region-001."
  exit 1
fi

# 7. DEPLOY LIVELLO MACROZONA (PROXIMITY FOG HUB)
# Crea l'Hub di Macrozona, VPC Peering con la Regione, Broker MQTT.
./deploy_macrozone.sh region-001 build-0001 --aws-region us-east-1
if [ $? -ne 0 ]; then
  echo "Errore nella creazione della macrozona build-0001."
  exit 1
fi

# 8. DEPLOY LIVELLO ZONA (EDGE FOG HUB + SENSOR AGENT)
# Crea l'Hub Edge, l'Agente Sensore e VPC Peering con la Macrozona.
./deploy_zone.sh region-001 build-0001 floor-001 --aws-region us-east-1
if [ $? -ne 0 ]; then
  echo "Errore nella creazione della zona floor-001."
  exit 1
fi
```

-----

## Gestione dei Servizi Infrastrutturali

L'architettura del Sensor Continuum fa ampio uso di componenti infrastrutturali esterni: **Broker di Messaggi** e **Database Relazionali**. Il sistema è stato progettato per essere agnostico rispetto al deployment di questi componenti.

L'utente può optare per qualsiasi implementazione compatibile, a condizione che sia correttamente configurata e raggiungibile sulla rete interna.

### Broker di Messaggi

I broker di messaggi (Kafka per la comunicazione Fog-to-Cloud e MQTT per la comunicazione Edge-to-Fog) sono **esterni** ai componenti Hub e Sensor.

#### A. Kafka

Qualsiasi istanza Kafka può essere utilizzata, ma è indispensabile che al suo interno siano creati i topic corretti per il flusso dati e di controllo.

**Script di Configurazione Topic:**

Lo script `configs/kafka/init-topics.sh` definisce i topic necessari per il corretto funzionamento del Proximity Hub e dei livelli superiori:

```bash
#!/bin/bash
set -e

echo "Creating Kafka topics if not exists..."

# aggregated-data-proximity-fog-hub
kafka-topics.sh --create --if-not-exists --topic aggregated-data-proximity-fog-hub \
  --bootstrap-server kafka-01:9092 --partitions 5 --replication-factor 1

# configuration-proximity-fog-hub
kafka-topics.sh --create --if-not-exists --topic configuration-proximity-fog-hub \
  --bootstrap-server kafka-01:9092 --partitions 3 --replication-factor 1

# statistics-data-proximity-fog-hub
kafka-topics.sh --create --if-not-exists --topic statistics-data-proximity-fog-hub \
  --bootstrap-server kafka-01:9092 --partitions 5 --replication-factor 1

# heartbeats-proximity-fog-hub (compacted)
kafka-topics.sh --create --if-not-exists --topic heartbeats-proximity-fog-hub \
  --bootstrap-server kafka-01:9092 --partitions 5 --replication-factor 1 \
  --config cleanup.policy=compact,delete
```

Nel sistema è disponibile un template Docker Compose custom per Kafka che esegue automaticamente la configurazione dei topic. Il loro deployment è illustrato in dettaglio nelle istruzioni per l'[Hub di Regione](./intermediate_fog_hub.md).

#### B. MQTT

Anche per il broker MQTT, possono essere utilizzate istanze standard. Vengono fornite immagini custom (es. `fmasci/sc-mqtt-broker:latest`) il cui deploy è discusso nelle sezioni relative all'[Hub di Macrozona](./proximity_fog_hub.md).

### Database Persistenti

I database sono essenziali per le operazioni di caching e l'implementazione dell'Outbox Pattern.

1.  **Compatibilità Richiesta:** Il database deve supportare le estensioni **PostGIS** e **TimescaleDB**.
2.  **Configurazione Schema:** Gli script per la creazione delle tabelle necessarie al corretto funzionamento del sistema si trovano nella cartella `configs/postgresql`.
3.  **Deployment:**
    * **Cloud (Uso Reale):** Si consiglia l'uso di servizi cloud gestiti (es. **AWS RDS per PostgreSQL**), come avviene per il database dei metadati del sistema.
    * **Locale (Simulazione):** Per il deployment locale o su infrastrutture proprie, si possono utilizzare le immagini e i template Docker Compose sviluppati internamente, illustrati contestualmente alle istruzioni per l'[Hub di Regione](./intermediate_fog_hub.md).