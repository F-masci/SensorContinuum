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
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o sensor-agent ./cmd/sensor-agent

# Fase 2: runtime (immagine minimale)
FROM gcr.io/distroless/static:nonroot

WORKDIR /app

# Copia il binario compilato dalla fase builder
COPY --from=builder /app/sensor-agent .

# Esegui il binario
ENTRYPOINT ["/app/sensor-agent"]