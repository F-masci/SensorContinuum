#!/usr/bin/env bash
#
# analyze_throughput_hub.sh
#
# Analizza i log Docker dei container hub per calcolare:
#  - numero totale di messaggi ricevuti
#  - throughput totale (msg/min)
#  - latenza end-to-end tra generazione e processamento dei messaggi (s e min)

#
# Lo script assume che nei log dei container hub sia presente un pattern
# di tipo "Real-time sensor data received" e, opzionalmente, un timestamp di
# generazione del messaggio (campo 'generated_at=YYYY-MM-DDTHH:MM:SS.sss')

# -----------------------------
# Parametri di default
# -----------------------------
WINDOW_MIN="5"
CONTAINERS_STR="region-hub-realtime-region-001-01 region-hub-realtime-region-001-02"
PATTERN_MAIN="Real-time sensor data received"

# -----------------------------
# Funzione di help / usage
# -----------------------------
usage() {
  cat <<EOF
Uso: $0 [--window-min <min>] [--containers "<c1 c2 ...>"]

Parametri:
  --window-min  : durata finestra osservazione in minuti (default: 5)
  --containers  : lista container Docker da analizzare (default: "hub1 hub2")

Esempio:
  $0 --window-min 5 --containers "hub1 hub2"
EOF
  exit 1
}

# -----------------------------
# Parsing argomenti
# -----------------------------
while [[ $# -gt 0 ]]; do
  case "$1" in
    --window-min) WINDOW_MIN="$2"; shift 2 ;;
    --containers) CONTAINERS_STR="$2"; shift 2 ;;
    -*) echo "[ERRORE] parametro sconosciuto: $1"; usage ;;
    *) echo "[ERRORE] parametro posizionale non riconosciuto: $1"; usage ;;
  esac
done

# -----------------------------
# Validazioni di base
# -----------------------------
if ! [[ "$WINDOW_MIN" =~ ^[0-9]+$ ]]; then
  echo "[ERRORE] --window-min deve essere un intero positivo."; exit 1
fi
WINDOW_INT_MIN=$WINDOW_MIN
WINDOW_SEC=$((WINDOW_INT_MIN*60))  # Docker usa secondi

read -r -a CONTAINERS_ARRAY <<< "$CONTAINERS_STR"
if (( ${#CONTAINERS_ARRAY[@]} == 0 )); then
  echo "[ERRORE] lista containers vuota."; exit 1
fi

if ! command -v docker >/dev/null 2>&1; then
  echo "[ERRORE] 'docker' non trovato."; exit 1
fi

# -----------------------------
# Recupero log dai container hub
# -----------------------------
TMP_MAIN="$(mktemp /tmp/analyze_throughput_hub.XXXXXX)"
trap 'rm -f "$TMP_MAIN"' EXIT

echo ">>> Analisi log Docker dei container hub"
echo "Finestra di osservazione      : $WINDOW_INT_MIN minuti"
echo "Containers analizzati         : ${CONTAINERS_ARRAY[*]}"
echo "Pattern di interesse          : '$PATTERN_MAIN'"
echo

for c in "${CONTAINERS_ARRAY[@]}"; do
  echo ">> Recupero log container: $c"
  if ! docker logs --since "${WINDOW_SEC}s" "$c" >> "$TMP_MAIN" 2>/dev/null; then
    echo "   [WARN] impossibile leggere logs di '$c'"
  fi
done

# -----------------------------
# Conteggio messaggi ricevuti
# -----------------------------
RECEIVED=$(grep -c -F "$PATTERN_MAIN" "$TMP_MAIN" || true)
echo ">>> Messaggi totali ricevuti: $RECEIVED"

# -----------------------------
# Calcolo throughput (msg/min)
# -----------------------------
THROUGHPUT=$(awk -v rec="$RECEIVED" -v win="$WINDOW_INT_MIN" 'BEGIN{ printf("%.2f", rec/win) }')
echo ">>> Throughput totale stimato : $THROUGHPUT msg/min"

# -----------------------------
# Estrazione latenza end-to-end
# -----------------------------
echo ">>> Analisi latenza end-to-end dai log"

LATENCIES_SEC=()
while read -r line; do
  # timestamp log (YYYY/MM/DD HH:MM:SS)
  PROC_TS=$(echo "$line" | awk '{print $4 " " $5}' | sed 's/\///g')

  # timestamp generazione: primo numero di 9-10 cifre dopo qualsiasi parola (tipicamente ID sensore)
  GEN_TS=$(echo "$line" | grep -oP '\b[0-9]{9,10}\b' | head -n1)

  if [[ -n "$PROC_TS" && -n "$GEN_TS" ]]; then
    PROC_EPOCH=$(date -d "$PROC_TS" +%s 2>/dev/null || echo 0)
    GEN_EPOCH=$GEN_TS
    if (( PROC_EPOCH > 0 && GEN_EPOCH > 0 )); then
      LAT=$(awk -v p="$PROC_EPOCH" -v g="$GEN_EPOCH" 'BEGIN{ printf("%.3f", p-g) }')
      LATENCIES_SEC+=("$LAT")
    fi
  fi
done < <(grep "$PATTERN_MAIN" "$TMP_MAIN")

LAT_COUNT=${#LATENCIES_SEC[@]}
if (( LAT_COUNT > 0 )); then
  LAT_SUM_SEC=$(printf "%s\n" "${LATENCIES_SEC[@]}" | awk '{s+=$1} END{printf "%.3f", s}')
  LAT_AVG_SEC=$(awk -v sum="$LAT_SUM_SEC" -v n="$LAT_COUNT" 'BEGIN{ printf "%.3f", sum/n }')
  LAT_MAX_SEC=$(printf "%s\n" "${LATENCIES_SEC[@]}" | sort -nr | head -n1)
  LAT_AVG_MIN=$(awk -v s="$LAT_AVG_SEC" 'BEGIN{ printf "%.3f", s/60 }')
  LAT_MAX_MIN=$(awk -v s="$LAT_MAX_SEC" 'BEGIN{ printf "%.3f", s/60 }')
else
  LAT_AVG_SEC="N/A"; LAT_MAX_SEC="N/A"
  LAT_AVG_MIN="N/A"; LAT_MAX_MIN="N/A"
fi

# -----------------------------
# Output dettagliato
# -----------------------------
echo
echo "=================== Risultati Analisi Hub ==================="
echo "Finestra osservata          : $WINDOW_INT_MIN min"
echo "Containers analizzati       : ${CONTAINERS_ARRAY[*]}"
echo "Messaggi ricevuti           : $RECEIVED"
echo "Throughput totale           : $THROUGHPUT msg/min"
echo "Messaggi con latenza nota   : $LAT_COUNT"
echo "Latenza media end-to-end    : $LAT_AVG_SEC s ($LAT_AVG_MIN min)"
echo "Latenza massima end-to-end  : $LAT_MAX_SEC s ($LAT_MAX_MIN min)"
echo "============================================================"

# -----------------------------
# Salvataggio dati in CSV
# -----------------------------
CSV_FILE="analyze_throughput.csv"
if [[ ! -f "$CSV_FILE" ]]; then
  echo "timestamp,window_min,received,throughput_msg_per_min,lat_count,lat_avg_s,lat_max_s,lat_avg_min,lat_max_min" > "$CSV_FILE"
fi
echo "$(date +'%Y-%m-%d %H:%M:%S'),$WINDOW_INT_MIN,$RECEIVED,$THROUGHPUT,$LAT_COUNT,$LAT_AVG_SEC,$LAT_MAX_SEC,$LAT_AVG_MIN,$LAT_MAX_MIN" >> "$CSV_FILE"
echo ">>> Dati salvati in CSV: $CSV_FILE"
