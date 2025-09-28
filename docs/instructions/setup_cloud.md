# Setup Cloud AWS

Il **Setup Cloud** è un processo a più fasi orchestrato da script Bash, che utilizzano **AWS CloudFormation** per il provisioning dell'infrastruttura e vari servizi per la configurazione logica. Questo processo garantisce la creazione della rete, degli endpoint pubblici e della persistenza dati, fornendo la base operativa per il resto del *Compute Continuum*.

-----

### A. Configurazione DNS Pubblico (`setup_dns.sh`)

Questo è il primo e più critico step di configurazione, poiché stabilisce la risoluzione dei nomi di dominio per tutti i servizi successivi (API, Database e Sito Web).

1.  **Creazione Hosted Zone:** Lo script esegue il deployment dello stack **`sc-public-dns`** tramite CloudFormation, che crea la **Zona Pubblica di Route 53** (es. `sensor-continuum.it`).
    ```bash
    aws cloudformation deploy --stack-name "sc-public-dns" --template-file "../cloudformation/public-dns.yaml"
    ```
2.  **Passaggio Manuale Cruciale:** Una volta completato il deploy, lo script recupera e stampa i **Name Server (NS)** assegnati da AWS. Questi indirizzi sono l'unica informazione che l'utente deve configurare **manualmente** presso il proprio *registrar* di dominio. Tutti i passi successivi dipendono dal successo di questa configurazione esterna.

-----

### B. Setup Rete VPC per Lambda e API Gateway (`setup_lambda.sh`)

Questa fase combina il provisioning dell'infrastruttura di rete privata e la creazione dell'interfaccia pubblica (API Gateway), gestendo automaticamente la complessità dei certificati SSL.

#### 1\. Setup Rete VPC per Lambda

Lo script deploya lo stack **`sc-lambda-network`** (`lambda-network.yaml`) che crea la rete dedicata e privata:

* Una **VPC** dedicata.
* Due **Subnet Private** (`sc-subnet-lambda-private-1` e `-2`).
* Un **NAT Gateway** e un **Security Group** dedicato (`sc-sg-lambda`).

Al termine, lo script estrae gli ID di VPC, Security Group e Subnet (es. `${VPC_ID}`, `${SUBNET1}`) per la successiva iniezione di parametri.

#### 2\. Deploy API Gateway, ACM e Configurazione CNAME

Lo script continua deployando lo stack **`sc-lambda-api`** (`lambda-api.yaml`), che crea:

* L'**API Gateway HTTP** e la richiesta di **Certificato ACM** per il dominio personalizzato (`api.sensor-continuum.it`).
* **Validazione DNS Automatica:** Il meccanismo cruciale è un **processo in background** che monitora gli eventi di CloudFormation. Questo processo intercetta il **CNAME di validazione** richiesto da ACM e lo inserisce automaticamente in Route 53, sbloccando l'emissione del certificato.
* **Configurazione Endpoint Pubblico:** Infine, recupera il dominio canonico dell'API Gateway e crea il record **CNAME** finale nella Zona Pubblica, rendendo l'URL **`api.sensor-continuum.it`** risolvibile e protetto.

-----

### C. Deploy Database Cloud dei Metadati (`setup_cloud_db.sh`)

Questo passaggio implementa il database centrale per i metadati globali, assicurando che sia raggiungibile in modo sicuro all'interno della VPC.

1.  **Deployment Cluster Aurora:** Lo script deploya lo stack **`sc-cloud-metadata-db`** (`cloud-db.yaml`), iniettando gli ID di **VPC** e **Subnet** estratti nello step precedente. Questo colloca il cluster **Aurora PostgreSQL** nella rete privata.
    ```bash
    aws cloudformation deploy --stack-name "sc-cloud-metadata-db" \
      --template-file "$TEMPLATE_FILE" \
      --parameter-overrides VpcId="$VPC_ID" Subnet1="$SUBNET1" Subnet2="$SUBNET2"
    ```
2.  **Configurazione Hostname Pubblici:** Recupera gli endpoint di Writer e Reader da Aurora e crea tre record **CNAME** nella Zona Pubblica per un accesso semplificato, come `write.cloud.metadata-db.sensor-continuum.it`.
3.  **Inizializzazione Schema:** Utilizzando il client `psql`, lo script si connette all'endpoint Writer per eseguire gli script SQL **`init-cloud-metadata-db.sql`** e **`init-metadata.sql`**, popolando lo schema iniziale e i metadati.

-----

### D. Setup Hosting Sito Web Metadata DB(`setup_site.sh`)

L'ultimo step del setup Cloud prepara l'ambiente di hosting statico per l'interfaccia utente.

1.  **Creazione App Amplify:** Lo script verifica o crea l'applicazione **AWS Amplify** (`Sensor Continuum`) che serve per l'hosting statico.
2.  **Associazione Dominio:** Viene associato il dominio personalizzato (`sensor-continuum.it` e `www`) all'app Amplify.
3.  **Setup S3:** Viene verificata e creata l'esistenza del bucket S3 di supporto.

> **Avvertenza:** Lo script `setup_site.sh` **prepara l'ambiente**. Il **deployment effettivo** del codice del sito web Metadata DBdeve essere completato manualmente o tramite un repository Git collegato alla console AWS Amplify.

-----

### E. Deploy delle Funzioni Lambda (`deploy_lambda.sh`)

La fase di creazione delle funzioni **Lambda** è gestita interamente dallo script `deploy_lambda.sh`, il quale orchestra il ciclo di vita completo di ogni singola funzione, dal *build* del codice all'integrazione finale con l'**API Gateway HTTP**. Lo script sfrutta l'**AWS Serverless Application Model (SAM)** per automatizzare il deployment, assicurando che le funzioni siano isolate e configurate per accedere alle risorse private.

Il processo di deployment di ciascuna Lambda segue una rigorosa sequenza di sette passaggi chiave:

1.  **Generazione Dinamica del Template:** Lo script recupera gli ID del **Security Group** (`sc-sg-lambda`) e delle **Subnet Private** (`sc-subnet-lambda-private-1`, `-2`) create durante il setup della rete VPC. Successivamente, utilizza `sed` per iniettare questi parametri, insieme al nome della funzione, in un template SAM di base. Questa iniezione è fondamentale, poiché costringe l'esecuzione della Lambda all'interno della VPC privata (sezione **`VpcConfig`**), garantendo che possa connettersi in modo sicuro al database PostgreSQL dei metadati.
2.  **Build del Codice:** Il comando **`sam build`** prepara l'artefatto di deployment (un file `.zip` contenente il codice della funzione).
3.  **Deploy SAM:** Viene verificata l'esistenza del bucket S3 dedicato (`sensor-continuum-lambda`) e quindi eseguito **`sam deploy`**. Questo comando carica l'artefatto e crea lo stack CloudFormation (es. `region-list-stack`), istanziando la funzione Lambda.
4.  **Recupero API ID:** Lo script recupera l'ID dell'API Gateway (`Sensor Continuum API`) precedentemente creato.
5.  **Configurazione Permessi di Invocation:** Tramite `aws lambda add-permission`, viene concesso un permesso esplicito (**`lambda:InvokeFunction`**) all'API Gateway per chiamare la Lambda su un ARN specifico (associato alla route, es. `GET /region/list`).
6.  **Creazione Integrazione:** Viene creata una **integrazione HTTP API** di tipo **`AWS_PROXY`** che mappa direttamente l'endpoint API alla Lambda. Questo tipo di integrazione garantisce che l'intera richiesta HTTP venga inoltrata alla funzione.
7.  **Creazione Route:** Infine, viene creata la **Route GET** nell'API Gateway (es. `GET /region/list`), che viene agganciata all'ID dell'integrazione creata nello step precedente.

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