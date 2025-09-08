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
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o intermediate-fog-hub ./cmd/intermediate-fog-hub

# Fase 2: runtime minimale con curl
FROM alpine:3.22

WORKDIR /app

# Installa curl per le richieste HTTP
RUN apk add --no-cache curl

# Copia il binario compilato dalla fase builder
COPY --from=builder /app/intermediate-fog-hub .

# Esegui il binario
ENTRYPOINT ["/app/intermediate-fog-hub"]