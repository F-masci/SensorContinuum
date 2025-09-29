# Documentazione Tecnica del Progetto Sensor Continuum

La cartella `docs` costituisce il repository centrale per tutti i documenti di supporto, i risultati sperimentali e le istruzioni operative, ed è organizzata in sezioni tematiche.

---

## Report Ufficiale e Dataset

Il documento ufficiale che descrive in dettaglio la progettazione, l'implementazione e la valutazione sperimentale del sistema è contenuto direttamente in questa cartella. I dati grezzi utilizzati per tali analisi sono qui archiviati per garantire la riproducibilità.

* **Report Tecnico:** [`Sensor Continuum Report - Renzi Maurizio 0369222, Masci Francesco 0365258.pdf`](./Sensor%20Continuum%20Report%20-%20Renzi%20Maurizio%200369222,%20Masci%20Francesco%200365258.pdf)
* **Dataset:** I file di dati grezzi in CSV acquisiti durante le sessioni di test sono collocati in questa cartella [`Dataset Failure`](./analyze_failure.csv) [`Dataset Throughput`](./analyze_throughput.csv).

---

## Istruzioni Operative

La sottocartella [`instructions`](./instructions/) raccoglie la documentazione per la configurazione e l'esecuzione del sistema in diversi ambienti.

Qui si trovano le istruzioni operative complete per:
* Il deployment locale tramite l'ambiente containerizzato Docker Compose.
* Il provisioning e il deployment dell'infrastruttura su piattaforma Cloud AWS.
* La guida all'esecuzione degli agenti di simulazione e la configurazione iniziale.

---

## Analisi di Performance e Risultati

Le sottocartelle seguenti contengono i grafici utilizzati per illustrare l'architettura e i risultati delle analisi. Queste immagini sono state integrate sia nel Report Tecnico che nel file [`README.md`](../README.md) principale del repository.

* [`deploy`](./deploy/): Contiene i diagrammi di architettura e deployment del sistema.
* [`performance`](./performance): Raccoglie i grafici generati dai dataset, utilizzati per rappresentare le metriche di Miss Rate, Throughput, Latenza e Affidabilità.
* [`flux`](./flux): Contiene le immagini dei log che attestano la correttezza del flusso di esecuzione e della comunicazione tra i diversi livelli del sistema.
