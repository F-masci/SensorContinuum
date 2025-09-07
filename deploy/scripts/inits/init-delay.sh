#!/bin/bash
#
# init-delay.sh
# Aggiunge/rimuove latenza e jitter a livello EC2 (e opzionalmente Docker)
#

DELAY="200ms"
JITTER="50ms"
IFACE="eth0"   # interfaccia primaria EC2
DOCKER_IFACE="docker0" # interfaccia bridge Docker

usage() {
  echo "Uso: $0 <apply|clear> [--delay Xms] [--jitter Yms] [--include-docker]"
  exit 1
}

ACTION=$1
shift || true

if [[ -z "$ACTION" ]]; then
  ACTION="apply"
fi


# Parse args
while [[ $# -gt 0 ]]; do
  case "$1" in
    --delay) DELAY="$2"; shift 2 ;;
    --jitter) JITTER="$2"; shift 2 ;;
    --include-docker) INCLUDE_DOCKER=1; shift ;;
    *) usage ;;
  esac
done

# Installo tc se non c'è
if ! command -v tc &> /dev/null; then
  echo "[INFO] Installo iproute (tc)..."
  sudo yum install -y iproute-tc
else
  echo "[INFO] iproute (tc) già installato."
fi

if [[ "$ACTION" == "apply" ]]; then
  echo "[INFO] Applico latenza $DELAY ±$JITTER su $IFACE (escludo SSH)"
  echo "[DEBUG] Rimuovo eventuali regole esistenti su $IFACE"
  sudo tc qdisc del dev $IFACE root 2>/dev/null

  echo "[DEBUG] Aggiungo root qdisc prio su $IFACE"
  sudo tc qdisc add dev $IFACE root handle 1: prio

  echo "[DEBUG] Aggiungo netem delay su $IFACE (parent 1:3)"
  sudo tc qdisc add dev $IFACE parent 1:3 handle 30: netem delay $DELAY $JITTER distribution normal

  echo "[DEBUG] Filtro: escludo traffico SSH (porta 22) dalla latenza"
  sudo tc filter add dev $IFACE protocol ip parent 1:0 prio 1 u32 match ip dport 22 0xffff flowid 1:1
  sudo tc filter add dev $IFACE protocol ip parent 1:0 prio 1 u32 match ip sport 22 0xffff flowid 1:1

  echo "[DEBUG] Filtro: tutto il resto del traffico riceve delay"
  sudo tc filter add dev $IFACE protocol ip parent 1:0 prio 2 u32 match ip dst 0.0.0.0/0 flowid 1:3
  sudo tc filter add dev $IFACE protocol ip parent 1:0 prio 2 u32 match ip src 0.0.0.0/0 flowid 1:3

  if [[ "$INCLUDE_DOCKER" == "1" ]]; then
    echo "[INFO] Applico anche al traffico tra container (bridge $DOCKER_IFACE)"
    echo "[DEBUG] Rimuovo eventuali regole esistenti su $DOCKER_IFACE"
    sudo tc qdisc del dev $DOCKER_IFACE root 2>/dev/null
    echo "[DEBUG] Aggiungo netem delay su $DOCKER_IFACE"
    sudo tc qdisc add dev $DOCKER_IFACE root netem delay $DELAY $JITTER distribution normal
  fi

  echo "[INFO] Latenza applicata."

elif [[ "$ACTION" == "clear" ]]; then
  echo "[INFO] Rimuovo regole su $IFACE e $DOCKER_IFACE"
  sudo tc qdisc del dev $IFACE root 2>/dev/null
  sudo tc qdisc del dev $DOCKER_IFACE root 2>/dev/null
  echo "[INFO] Ripristinato."
else
  usage
fi
