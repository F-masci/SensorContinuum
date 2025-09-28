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

Il deployment su AWS è gestito da un set di script Bash situati nella directory **`deploy/scripts`**.
Gli script utilizzano **AWS CloudFormation** e **Docker Compose** per l'orchestrazione, seguendo l'ordine gerarchico del Continuum:

1.  **`setup_bucket.sh`**: Prepara il bucket S3 di supporto per tutti gli asset e gli script necessari al deployment.
2.  **`deploy_region.sh`**: Crea l'infrastruttura centrale del livello Intermediate Fog Hub (VPC, Kafka, Databases, Services).
3.  **`deploy_macrozone.sh`**: Crea il Proximity Fog Hub a livello di macrozona, includendo il Broker MQTT e il database locale.
4.  **`deploy_zone.sh`**: Crea l'Edge Hub e i Sensor Agents a livello di zona.

### Esempio di Deploy Completo (Singola Zona)

Il seguente esempio mostra la sequenza di comandi necessari per deployare l'intero sistema Compute Continuum (esclusa la parte Cloud) in una singola regione (`us-east-1`), con una macrozona e una zone attive, garantendo la completa operatività della pipeline dati dall'Edge all'Intermediate Fog Hub.

Si noti come gli script utilizzino nomi logici e prevedibili (`region-001`, `build-0001`, `floor-001`) per identificare univocamente le risorse:

```bash
#!/bin/bash

# 1. SETUP INIZIALE: Preparazione degli asset su AWS S3
# Questo bucket è il punto di partenza per il UserData di tutte le istanze EC2.
./setup_bucket.sh
if [ $? -ne 0 ]; then
  echo "Errore nella creazione del bucket."
  exit 1
fi

# 2. DEPLOY LIVELLO REGIONALE (INTERMEDIATE FOG HUB)
# Crea VPC, Kafka Broker, Databases e l'istanza EC2 dei Servizi nella regione US-EAST-1.
./deploy_region.sh region-001 --aws-region us-east-1
if [ $? -ne 0 ]; then
  echo "Errore nella creazione della regione region-001."
  exit 1
fi

# 3. DEPLOY LIVELLO MACROZONA E ZONE (PROXIMITY FOG E EDGE FOG)
# Inizializza la Macrozona 'build-0001'.
./deploy_macrozone.sh region-001 build-0001 --aws-region us-east-1
if [ $? -ne 0 ]; then
  echo "Errore nella creazione della macrozona build-0001."
  exit 1
fi

# Inizializza le zona 'floor-001' nella macrozona 'build-0001'.
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