# Documentazione Tecnica del Progetto Sensor Continuum

Il presente documento funge da indice e guida per la consultazione degli artefatti e della documentazione approfondita relativa al progetto **Sensor Continuum: Applicazione Distribuita nel Compute Continuum**.

Questa cartella (`docs`) costituisce il repository centrale per tutti i documenti di supporto, i risultati sperimentali e le istruzioni operative, ed è organizzata in sezioni tematiche.

---

## Report Ufficiale e Dataset

Il documento ufficiale che descrive in dettaglio la progettazione, l'implementazione e la valutazione sperimentale del sistema è contenuto direttamente in questa cartella. I dati grezzi utilizzati per tali analisi sono qui archiviati per garantire la riproducibilità.

* **Report Tecnico:** `Report_SDCC.pdf`
* **Dataset:** I file di dati grezzi in CSV acquisiti durante le sessioni di test sono collocati in questa cartella.

---

## Istruzioni Operative

La sottocartella **`instructions`** raccoglie la documentazione per la configurazione e l'esecuzione del sistema in diversi ambienti.

Qui si trovano le istruzioni operative complete per:
* Il deployment locale tramite l'ambiente containerizzato Docker Compose.
* Il provisioning e il deployment dell'infrastruttura su piattaforma Cloud AWS.
* La guida all'esecuzione degli agenti di simulazione e la configurazione iniziale.

---

## Analisi di Performance e Risorse Visive

Le sottocartelle seguenti contengono le risorse visive utilizzate per illustrare l'architettura e i risultati delle analisi. Queste immagini sono state integrate sia nel Report Tecnico che nel file `README.md` principale del repository.

* **`deploy`:** Contiene i diagrammi di architettura e deployment del sistema.
* **`performance`:** Raccoglie le visualizzazioni grafiche (boxplot, trend) generate dai dataset, utilizzate per rappresentare le metriche di Throughput, Latenza e Affidabilità.
* **`flux`:** Contiene le immagini dei log che attestano la correttezza del flusso di esecuzione e della comunicazione tra i diversi livelli del Compute Continuum.
