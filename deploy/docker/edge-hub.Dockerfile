# Fase 1: build (usiamo un'immagine con toolchain Go)
FROM golang:1.24.5 AS builder

# Imposta la directory di lavoro
WORKDIR /app

# Copia go.mod per installare le dipendenze
COPY go.mod ./
RUN go mod tidy
RUN go mod download

# Copia il codice sorgente nel container
COPY . .

# Compila il binario per linux
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o edge-hub ./cmd/edge-hub

# Fase 2: runtime (immagine minimale)
FROM alpine:latest

# Crea un utente non-root per una maggiore sicurezza
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copia il binario compilato dalla fase builder
COPY --from=builder /app/edge-hub .

# Imposta i permessi corretti per l'utente non-root
RUN chown appuser:appgroup /app/edge-hub

# Passa all'utente non-root
USER appuser

# Esegui il binario
ENTRYPOINT ["./edge-hub"]