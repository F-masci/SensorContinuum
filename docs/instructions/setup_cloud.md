# Setup Cloud AWS

Il Setup Cloud è un processo a più fasi orchestrato da script Bash, che utilizzano AWS CloudFormation per il provisioning dell'infrastruttura e vari servizi per la configurazione logica. Questo processo garantisce la creazione della rete, degli endpoint pubblici e della persistenza dati, fornendo la base operativa per il resto del sistema *Sensor Continuum*.

-----

### Configurazione DNS Pubblico ([`setup_dns.sh`](../../deploy/scripts/setup_dns.sh))

Questo è il primo e più critico step di configurazione, poiché stabilisce la risoluzione dei nomi di dominio per tutti i servizi successivi (API, Database e Sito Web).

1.  **Creazione Hosted Zone:** Lo script esegue il deployment dello stack `sc-public-dns` tramite CloudFormation, che crea la Zona Pubblica di Route 53 per *sensor-continuum.it*.
    ```bash
    aws cloudformation deploy --stack-name "sc-public-dns" --template-file "../cloudformation/public-dns.yaml"
    ```
2.  ⚠️ **Passaggio Manuale Cruciale:** Una volta completato il deploy, lo script recupera e stampa i Name Server assegnati da AWS. Questi indirizzi sono l'unica informazione che l'utente deve configurare **manualmente** presso il proprio *registrar* di dominio. Tutti i passi successivi dipendono dal successo di questa configurazione esterna.

-----

### Setup Rete VPC per Lambda e API Gateway ([`setup_lambda.sh`](../../deploy/scripts/setup_lambda.sh))

Questa fase combina il provisioning dell'infrastruttura di rete privata e la creazione dell'interfaccia pubblica tramite API Gateway, gestendo automaticamente la complessità dei certificati SSL.

#### 1\. Setup Rete VPC per Lambda

Lo script deploya lo stack `sc-lambda-network`, definito in [`lambda-network.yaml`](../../deploy/cloudformation/lambda-network.yaml) che crea la rete dedicata, pubblica e privata:

* Una VPC dedicata: `sc-lambda-vpc`.
* Due Subnet Private: `sc-subnet-lambda-private-1` e `-2`.
* Un NAT Gateway e un Security Group dedicato: `sc-lambda-natgw` e `sc-sg-lambda`.

Al termine, lo script estrae gli ID di VPC, Security Group e Subnet per la successiva iniezione di parametri.

#### 2\. Deploy API Gateway, ACM e Configurazione CNAME

Lo script continua deployando lo stack `sc-lambda-api` tramite il file [`lambda-api.yaml`](../../deploy/cloudformation/lambda-api.yaml), che crea l'API Gateway HTTP e la richiesta di Certificato ACM per il dominio personalizzato *api.sensor-continuum.it*.
Il meccanismo cruciale è un processo in background che monitora gli eventi di CloudFormation. Questo processo intercetta il CNAME di validazione richiesto da ACM e lo inserisce automaticamente in Route 53, sbloccando l'emissione del certificato.
Infine, recupera il dominio canonico dell'API Gateway e crea il record CNAME finale nella Zona Pubblica, rendendo l'URL *api.sensor-continuum.it* risolvibile e protetto.

-----

### Deploy Database Cloud dei Metadati Globali ([`setup_cloud_db.sh`](../../deploy/scripts/setup_cloud_db.sh))

Questo passaggio implementa il database centrale Aurora PostgreSQL per i metadati globali, insieme alla sua infrastruttura di rete dedicata.

1.  **Deployment Cluster Aurora e Rete Dedicata:** Lo script deploya lo stack `sc-cloud-metadata-db` utilizzando [`cloud-db.yaml`](../../deploy/cloudformation/cloud-db.yaml). A differenza dei passaggi precedenti, questo template crea una nuova VPC dedicata (`sc-cloud-meta-db-vpc`) con subnet pubbliche, un Internet Gateway e le relative tabelle di routing. Colloca il cluster Aurora PostgreSQL su queste subnet e lo rende pubblicamente accessibile (con accesso limitato alla porta `5433` tramite il Security Group).
    ```bash
    aws cloudformation deploy --stack-name "sc-cloud-metadata-db" \
      --template-file "$TEMPLATE_FILE"
    ```
    (***Nota: lo script non utilizza i parametri di rete di step precedenti***)
2.  **Configurazione Hostname Pubblici:** Lo script recupera gli endpoint di Writer e Reader del cluster Aurora. Crea tre record CNAME nella Zona Pubblica di Route 53 di `sensor-continuum.it` per semplificare l'accesso: `write.cloud.metadata-db.sensor-continuum.it`, `read-only.cloud.metadata-db.sensor-continuum.it` e l'alias principale `cloud.metadata-db.sensor-continuum.it`che punta l'endpoint di lettura.
3.  **Inizializzazione Schema:** Utilizzando le credenziali cablate, lo script si connette tramite il client `psql` all'endpoint Writer (sulla porta `5433`) ed esegue in sequenza gli script SQL [`init-cloud-metadata-db.sql`](../../configs/postgresql/init-cloud-metadata-db.sql) e [`init-metadata.sql`](../../configs/postgresql/init-metadata.sql), popolando lo schema iniziale del database.

-----

### Setup Hosting Sito Web ([`setup_site.sh`](../../deploy/scripts/setup_site.sh))

L'ultimo step del setup Cloud prepara l'ambiente di hosting statico per l'interfaccia utente.

1.  **Creazione App Amplify:** Lo script verifica o crea l'applicazione AWS Amplify `Sensor Continuum`, che serve per l'hosting statico.
2.  **Associazione Dominio:** Viene associato il dominio personalizzato *sensor-continuum.it* e *www.sensor-continuum.it* all'app Amplify.
3.  **Setup S3:** Viene verificata e creata l'esistenza del bucket S3 di supporto.

> **⚠️ Avvertenza**
>
> Lo script [`setup_site.sh`](../../deploy/scripts/setup_site.sh) prepara l'ambiente. Il deployment effettivo del codice del sito web è gestito dal comando `npm run deploy`, che esegue i seguenti passi:
>
> * **Build del Progetto**: Esegue react-scripts build.
> 
> * **Sincronizzazione S3**: Esegue lo script [`./deploy.sh`](../../site/deploy.sh), che a sua volta effettua una sincronizzazione diretta della cartella [`build`](../../site/) verso il bucket S3 configurato: `s3://sensor-continuum-site`.
> 
> Dopo la sincronizzazione su S3, il deployment non è tecnicamente completo, poiché è necessario accedere alla dashboard di AWS Amplify per innescare la finalizzazione della distribuzione, dato che l'hosting è configurato per servire i contenuti direttamente da un ambiente collegato ad Amplify.

-----

### Deploy delle Funzioni Lambda ([`deploy_lambda.sh`](../../deploy/scripts/deploy_lambda.sh))

La fase di creazione delle funzioni Lambda è gestita interamente dallo script [`deploy_lambda.sh`](../../deploy/scripts/deploy_lambda.sh), il quale orchestra il ciclo di vita completo di ogni singola funzione, dal build del codice all'integrazione finale con l'API Gateway HTTP. Lo script sfrutta l'*AWS Serverless Application Model* (SAM) per automatizzare il deployment, assicurando che le funzioni siano isolate e configurate per accedere alle risorse private.

Il processo di deployment di ciascuna Lambda segue una rigorosa sequenza di sette passaggi chiave:

1.  **Generazione Dinamica del Template:** Lo script recupera gli ID del Security Group `sc-sg-lambda` e delle Subnet Private `sc-subnet-lambda-private-1` e`-2` create durante il setup della rete VPC. Successivamente, utilizza `sed` per iniettare questi parametri, insieme al nome della funzione, in un template SAM di base [`lambda.template.yaml`](../../deploy/cloudformation/lambda.template.yaml). Questa iniezione è fondamentale, poiché costringe l'esecuzione della Lambda all'interno della VPC privata, garantendo che possa connettersi in modo sicuro al database PostgreSQL dei metadati.
2.  **Build del Codice:** Il comando `sam build` prepara l'artefatto di deployment (un file `.zip` contenente il codice della funzione).
3.  **Deploy SAM:** Viene verificata l'esistenza del bucket S3 dedicato `sensor-continuum-lambda` e quindi eseguito `sam deploy`. Questo comando carica l'artefatto e crea lo stack CloudFormation, istanziando la funzione Lambda.
4.  **Recupero API ID:** Lo script recupera l'ID dell'API Gateway `Sensor Continuum API` precedentemente creato.
5.  **Configurazione Permessi di Invocation:** Tramite `aws lambda add-permission`, viene concesso un permesso esplicito `lambda:InvokeFunction` all'API Gateway per chiamare la Lambda su un ARN specifico associato alla route.
6.  **Creazione Integrazione:** Viene creata una integrazione HTTP API di tipo `AWS_PROXY` che mappa direttamente l'endpoint API alla Lambda. Questo tipo di integrazione garantisce che l'intera richiesta HTTP venga inoltrata alla funzione.
7.  **Creazione Route:** Infine, viene creata la Route GET nell'API Gateway, che viene agganciata all'ID dell'integrazione creata nello step precedente.

#### Chiamate di Deployment Esemplari

Le chiamate di deployment seguono una logica gerarchica, definendo l'endpoint API e la funzione associata:

##### Endpoints di Livello Regionale (Region)

Questi endpoint gestiscono i metadati e i dati consolidati a livello regionale:

* Per la **Lista delle Regioni**:
  ```bash
  ./deploy_lambda.sh region-list-stack region regionList "/region/list"
  ```
* Per la **Ricerca di una Regione per Nome**:
  ```bash
  ./deploy_lambda.sh region-search-name-stack region regionSearchName "/region/search/name/{name}"
  ```
* Per i **Dati Aggregati per Regione**:
  ```bash
  ./deploy_lambda.sh region-data-aggregated-stack region regionDataAggregated "/region/data/aggregated/{region}"
  ```

##### Endpoints di Livello Macrozona (Macrozona)

Questi endpoint forniscono l'accesso a liste, dati aggregati e statistiche complesse a livello di Macrozona:

* Per la **Lista delle Macrozona** (filtrata per regione):
  ```bash
  ./deploy_lambda.sh macrozone-list-stack macrozone macrozoneList "/macrozone/list/{region}"
  ```
* Per i **Dati Aggregati per Macrozona** (dal nome):
  ```bash
  ./deploy_lambda.sh macrozone-data-aggregated-name-stack macrozone macrozoneDataAggregatedName "/macrozone/data/aggregated/{region}/{macrozone}"
  ```
* Per il **Trend Statistico della Macrozona**:
  ```bash
  ./deploy_lambda.sh macrozone-data-trend-stack macrozone macrozoneDataTrend "/macrozone/data/trend/{region}"
  ```
* Per la **Correlazione di Variazione della Macrozona**:
  ```bash
  ./deploy_lambda.sh macrozone-data-variation-correlation-stack macrozone macrozoneDataVariationCorrelation "/macrozone/data/variation/correlation/{region}"
  ```

##### Endpoints di Livello Zona (Zone)

Questi endpoint gestiscono l'accesso più granulare, inclusi i dati grezzi dei sensori:

* Per la **Lista delle Zone** (filtrata per regione e macrozona):
  ```bash
  ./deploy_lambda.sh zone-list-stack zone zoneList "/zone/list/{region}/{macrozone}"
  ```
* Per la **Ricerca di una Zona per Nome**:
  ```bash
  ./deploy_lambda.sh zone-search-name-stack zone zoneSearchName "/zone/search/name/{region}/{macrozone}/{name}"
  ```
* Per i **Dati Aggregati per Zona**:
  ```bash
  ./deploy_lambda.sh zone-data-aggregated-stack zone zoneDataAggregated "/zone/data/aggregated/{region}/{macrozone}/{zone}"
  ```
* Per i **Dati Sensori Raw (grezzi)**: Questo è l'endpoint più dettagliato, che richiede tutti e quattro i parametri gerarchici:
  ```bash
  ./deploy_lambda.sh zone-sensor-data-raw-stack zone zoneSensorDataRaw "/zone/sensor/data/raw/{region}/{macrozone}/{zone}/{sensor}"
  ```