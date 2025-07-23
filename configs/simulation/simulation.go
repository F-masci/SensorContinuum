package simulation

const TIMEOUT = 5000 // Timeout in millisecondi per la simulazione

const CSV_DIR = "csv"

const MISSING_PROBABILITY = 0.15 // Probabilità di non generare un valore (simulazione di dati mancanti)

const OUTLIER_PROBABILITY = 0.10 // Probabilità di generare un outlier
const OUTLIER_MULTIPLIER = 20.0  // Moltiplicatore (sulla deviazione standard) per generare un outlier
const OUTLIER_ADDITION = 0.0     // Valore aggiunto (nella media) per generare un outlier
