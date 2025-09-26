#!/usr/bin/env bash
#
# analyze_failure.sh
#
# Analizza i log Docker (ultimo --window secondi) per calcolare failure/missing rate
# e confrontarlo con il valore atteso teorico e con il valore atteso corretto
# considerando una probabilità di "miss" (cioè la probabilità che in simulazione
# un sensore NON generi effettivamente il dato), i messaggi reali inviati dai simulatori,
# i valori mancanti generati e gli outlier simulati.
#
# Caratteristiche principali:
#  - legge i log di più container Docker passati in una singola stringa (es. "hub1 hub2")
#  - conta i messaggi basandosi su pattern robusti trovati
#  - estrae sensori unici con regex (es. sensor-agent-05)
#  - calcola expected = sensors * floor(window/interval)
#  - calcola expected_adjusted = round(expected * (1 - miss_prob))
#  - calcola messaggi reali inviati dai simulatori
#  - conta messaggi non inviati e outlier generati
#  - stima failure rate e percentuale di outlier non riconosciuti
#  - calcola falsi positivi e missing rate reale
#
# Uso:
#   ./analyze_failure.sh [--window 120] [--interval 5] [--containers "hub1 hub2"] [--type all|valid|outlier] [--miss 0.1]

# -----------------------------
# Parametri di default
# -----------------------------
WINDOW="180"
INTERVAL="5"
CONTAINERS_STR="zone-hub-filter-floor-001-01 zone-hub-filter-floor-001-02"
TYPE="all"
MISS_PROB="0.15"
ENABLE_CSV=false

NUM_SENSORS=50
SIMULATOR_CONTAINER_STR=""

for i in $(seq -w 1 $NUM_SENSORS); do
  SIMULATOR_CONTAINER_STR+="sensors-sensor-agent-${i}-1 "
done

# Rimuove spazi finali usando parameter expansion
SIMULATOR_CONTAINER_STR="${SIMULATOR_CONTAINER_STR%%[[:space:]]}"

# -----------------------------
# Funzione di help / usage
# -----------------------------
usage() {
  cat <<EOF
Uso: $0 [--window <sec>] [--interval <sec>] [--containers "<c1 c2 ...>"] [--type all|valid|outlier] [--miss <prob>] [--simulators "<s1 s2 ...>"] [--csv]

Parametri:
  --window     : durata finestra osservazione in secondi (default: 180)
  --interval   : intervallo atteso tra messaggi dallo stesso sensore (default: 5)
  --containers : lista container Docker da analizzare (default: "hub1 hub2")
  --type       : tipo di messaggi da contare (all|valid|outlier) (default: all)
  --miss       : probabilità (0..1) che in simulazione il sensore non generi il dato (default: 0.15)
  --simulators : container dei simulatori per contare messaggi reali inviati
  --csv        : abilita salvataggio dei risultati in CSV

Esempio:
  $0 --window 120 --interval 5 --containers "hub1 hub2" --type all --miss 0.1 --simulators "sim1 sim2" --csv
EOF
  exit 1
}

# -----------------------------
# Parsing argomenti
# -----------------------------
while [[ $# -gt 0 ]]; do
  case "$1" in
    --window) WINDOW="$2"; shift 2 ;;
    --interval) INTERVAL="$2"; shift 2 ;;
    --containers) CONTAINERS_STR="$2"; shift 2 ;;
    --type) TYPE="$2"; shift 2 ;;
    --miss) MISS_PROB="$2"; shift 2 ;;
    --simulators) SIMULATOR_CONTAINER_STR="$2"; shift 2 ;;
    --csv) ENABLE_CSV=true; shift ;;
    -*) echo "[ERRORE] parametro sconosciuto: $1"; usage ;;
    *) echo "[ERRORE] parametro posizionale non riconosciuto: $1"; usage ;;
  esac
done

# -----------------------------
# Validazioni di base
# -----------------------------
if [[ -z "$WINDOW" || -z "$INTERVAL" || -z "$CONTAINERS_STR" ]]; then
  echo "[ERRORE] --window, --interval e --containers sono obbligatori."
  usage
fi

if ! [[ "$WINDOW" =~ ^[0-9]+$ ]] || ! [[ "$INTERVAL" =~ ^[0-9]+$ ]]; then
  echo "[ERRORE] --window e --interval devono essere interi positivi."
  exit 1
fi
WINDOW_INT=$((WINDOW))
INTERVAL_INT=$((INTERVAL))

if (( INTERVAL_INT <= 0 )); then
  echo "[ERRORE] --interval deve essere > 0"; exit 1
fi
if (( WINDOW_INT <= 0 )); then
  echo "[ERRORE] --window deve essere > 0"; exit 1
fi
if (( INTERVAL_INT > WINDOW_INT )); then
  echo "[WARN] --interval ($INTERVAL_INT s) > --window ($WINDOW_INT s). Attesi 0 o 1 messaggi per sensore."
fi

if ! awk "BEGIN{ if ($MISS_PROB+0 >= 0 && $MISS_PROB+0 <= 1) exit 0; exit 1 }"; then
  echo "[ERRORE] --miss deve essere compreso tra 0 e 1."; exit 1
fi

case "$TYPE" in
  all) PATTERN="Processing data for sensor" ;;
  valid) PATTERN="Data is valid for sensor" ;;
  outlier) PATTERN="Outlier detected and discarded for sensor" ;;
  *) echo "[ERRORE] Tipo non valido: $TYPE"; exit 1 ;;
esac

read -r -a CONTAINERS_ARRAY <<< "$CONTAINERS_STR"
read -r -a SIMULATOR_ARRAY <<< "$SIMULATOR_CONTAINER_STR"

if (( ${#CONTAINERS_ARRAY[@]} == 0 )); then
  echo "[ERRORE] lista containers vuota."; exit 1
fi
if ! command -v docker >/dev/null 2>&1; then
  echo "[ERRORE] 'docker' non trovato."; exit 1
fi

# -----------------------------
# Recupero log dai container principali
# -----------------------------
TMPFILE="$(mktemp /tmp/analyze_failure.XXXXXX)"
trap 'rm -f "$TMPFILE"' EXIT

echo ">>> Analisi logs Docker principali"
echo "Finestra osservazione            : $WINDOW_INT s"
echo "Intervallo atteso per sensore    : $INTERVAL_INT s"
echo "Tipo conteggio (pattern)         : $TYPE -> '$PATTERN'"
echo "Probabilità miss (simulazione)   : $MISS_PROB"
echo "Containers                       : ${CONTAINERS_ARRAY[*]}"
echo

for c in "${CONTAINERS_ARRAY[@]}"; do
  echo ">> Recupero log container: $c"
  if ! docker logs --since "${WINDOW}s" "$c" >> "$TMPFILE" 2>/dev/null; then
    echo "   [WARN] impossibile leggere logs di '$c'"
  fi
done

# -----------------------------
# Conteggi principali
# -----------------------------
RECEIVED=$(grep -c -F "$PATTERN" "$TMPFILE" || true)
VALID=$(grep -c -F "Data is valid for sensor" "$TMPFILE" || true)
OUTLIERS=$(grep -c -F "Outlier detected and discarded for sensor" "$TMPFILE" || true)
ERRS=$(grep -c -i "\[ERROR\]" "$TMPFILE" || true)

SENSOR_ID_LIST=$(grep -oE "sensor-agent-[0-9]+" "$TMPFILE" | sort -u || true)
if [[ -z "$SENSOR_ID_LIST" ]]; then UNIQUE_SENSORS=0; else UNIQUE_SENSORS=$(printf "%s\n" "$SENSOR_ID_LIST" | wc -l); fi

SAMPLES_PER_SENSOR=$(( WINDOW_INT / INTERVAL_INT ))
EXPECTED=$(( UNIQUE_SENSORS * SAMPLES_PER_SENSOR ))
EXPECTED_ADJ=$(awk -v e="$EXPECTED" -v m="$MISS_PROB" 'BEGIN{ adj=e*(1-m); printf("%.0f", adj) }')

EXPECTED_MISSED=$(( EXPECTED - RECEIVED )); (( EXPECTED_MISSED < 0 )) && EXPECTED_MISSED=0

MISSING_RATE_G=$(awk -v e="$EXPECTED" -v r="$RECEIVED" 'BEGIN { rate=(e-r)/e*100; if(rate<0) rate=0; printf("%.2f", rate) }')
MISSING_RATE_ADJ=$(awk -v e="$EXPECTED_ADJ" -v r="$RECEIVED" 'BEGIN { rate=(e-r)/e*100; if(rate<0) rate=0; printf("%.2f", rate) }')

# -----------------------------
# Analisi simulatori sensori
# -----------------------------
SIM_SENT=0
SIM_MISSING=0
SIM_OUTLIER=0
TMP_SIM="$(mktemp /tmp/analyze_failure_sim.XXXXXX)"
trap 'rm -f "$TMP_SIM"' EXIT

if (( ${#SIMULATOR_ARRAY[@]} > 0 )); then
  echo ">>> Analisi log simulatori"
  echo "Containers simulatori            : ${SIMULATOR_ARRAY[*]}"
  for s in "${SIMULATOR_ARRAY[@]}"; do
    echo ">> Recupero log simulator: $s"
    if ! docker logs --since "${WINDOW}s" "$s" >> "$TMP_SIM" 2>/dev/null; then
      echo "   [WARN] impossibile leggere logs simulator '$s'"
    fi
  done

  SIM_SENT=$(grep -c "Sensor reading:" "$TMP_SIM" || true)
  SIM_MISSING=$(grep -c "Generating missing value" "$TMP_SIM" || true)
  SIM_OUTLIER=$(grep -c "Generating outlier" "$TMP_SIM" || true)
fi

MISSED=$(( SIM_SENT - RECEIVED )); (( MISSED < 0 )) && MISSED=0

MISSING_RATE_G_SIM=$(awk -v sent="$SIM_SENT" -v miss="$MISSED" 'BEGIN { if(sent>0) { rate=miss/sent*100; if(rate<0) rate=0; printf("%.2f", rate) } else { printf("N/A") } }')
MISSING_RATE_ADJ_SIM=$(awk -v sent="$SIM_SENT" -v miss="$MISSED" -v sim_miss="$SIM_MISSING" 'BEGIN { if(sent>0) { rate=(miss - sim_miss)/sent*100; if(rate<0) rate=0; printf("%.2f", rate) } else { printf("N/A") } }')

if (( SIM_OUTLIER > 0 )); then
  OUTLIER_ERROR_PERCENT=$(awk -v gen="$SIM_OUTLIER" -v detected="$OUTLIERS" \
    'BEGIN { printf("%.2f", ( (detected - gen >= 0 ? detected - gen : gen - detected) / gen )*100 ) }')
else
  OUTLIER_ERROR_PERCENT="N/A"
fi

# -----------------------------
# Output dettagliato
# -----------------------------
echo
echo "=================== Risultati Analisi Failure ==================="
echo "Sensori unici osservati           : $UNIQUE_SENSORS"
if (( UNIQUE_SENSORS > 0 )); then
  echo "Elenco sensori unici:"
  printf "  %s\n" $SENSOR_ID_LIST | sed 's/^/    - /'
fi
echo
echo "Messaggi attesi teorici           : $EXPECTED"
echo "Messaggi attesi con miss          : $EXPECTED_ADJ"
echo "Messaggi effettivi (pattern)      : $RECEIVED"
echo "  - validi                        : $VALID"
echo "  - outlier scartati              : $OUTLIERS"
echo "  - error log                     : $ERRS"
echo "  - miss stimati (teorici)        : $EXPECTED_MISSED"
echo
echo "Missing rate (grezzo)             : $MISSING_RATE_G %"
echo "Missing rate (aggiustato miss)    : $MISSING_RATE_ADJ %"
echo
if (( ${#SIMULATOR_ARRAY[@]} > 0 )); then
  echo ">>> Dati dai simulatori"
  echo "Messaggi reali inviati            : $SIM_SENT"
  echo "Valori mancanti generati          : $SIM_MISSING"
  echo "Valori mancanti osservati         : $MISSED"
  echo "Missing rate osservato            : $MISSING_RATE_G_SIM %"
  echo "Missing rate con miss             : $MISSING_RATE_ADJ_SIM %"
  echo
  echo "Outlier generati dai simulatori   : $SIM_OUTLIER"
  echo "Outlier rilevati dagli hub        : $OUTLIERS"
  echo "Percentuale errore outlier        : $OUTLIER_ERROR_PERCENT %"
fi
echo "================================================================="
echo

# -----------------------------
# Salvataggio dei dati in CSV
# -----------------------------
if $ENABLE_CSV; then
  CSV_FILE="analyze_failure.csv"
  if [[ ! -f "$CSV_FILE" ]]; then
    echo "timestamp,window,interval,type,miss_prob,unique_sensors,expected,expected_adj,received,valid,expected_missed,errs,sim_sent,sim_missing,sim_missed,missing_rate_g,missing_rate_adj,missing_rate_g_sim,missing_rate_adj_sim,sim_outlier,outliers,outlier_error_percent" > "$CSV_FILE"
  fi

  echo "$(date +'%Y-%m-%d %H:%M:%S'),$WINDOW_INT,$INTERVAL_INT,$TYPE,$MISS_PROB,$UNIQUE_SENSORS,$EXPECTED,$EXPECTED_ADJ,$RECEIVED,$VALID,$EXPECTED_MISSED,$ERRS,$SIM_SENT,$SIM_MISSING,$MISSED,$MISSING_RATE_G,$MISSING_RATE_ADJ,$MISSING_RATE_G_SIM,$MISSING_RATE_ADJ_SIM,$SIM_OUTLIER,$OUTLIERS,$OUTLIER_ERROR_PERCENT" >> "$CSV_FILE"

  echo ">>> Dati salvati in CSV: $CSV_FILE"
fi