# Usa immagine ufficiale di Go
FROM golang:1.24.5 AS builder

# Imposta directory di lavoro
WORKDIR /app

# Copia i sorgenti nella build context
COPY . .

# Compila binario statico
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-linkmode external -extldflags '-static'" -o ./cmd/api-backend/region/regionList ./cmd/api-backend/region/regionList.go

# Crea un'immagine minima solo per creare il file ZIP
FROM alpine:latest AS packager

# Installa zip
RUN apk --no-cache add zip

WORKDIR /lambda

# Copia binario compilato
COPY --from=builder /app/cmd/api-backend/region/regionList .

# Rendi sicuro il file
RUN chmod +x regionList

# Crea file ZIP con solo il binario
RUN zip regionList.zip regionList

# Output finale
CMD ["cat", "regionList.zip"]
