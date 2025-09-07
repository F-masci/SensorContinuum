FROM golang:1.24.5 AS builder

ARG LAMBDA_PATH=
WORKDIR /app
COPY . .

# Scarica tutte le dipendenze Go dichiarate in go.mod/go.sum
# RUN go mod tidy
RUN go mod download

RUN if [ -n "$LAMBDA_PATH" ]; then \
      src="$LAMBDA_PATH"; \
      case "$src" in *.go) src="${src%.go}";; esac; \
      out="cmd/api-backend/$src/main"; \
      mkdir -p "$(dirname "$out")"; \
      CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-linkmode external -extldflags '-static'" -o "$out" "cmd/api-backend/$src.go"; \
    else \
      find ./cmd/api-backend -type f -name '*.go' -exec sh -c \
        'for f; do \
            out="${f%.go}"; \
            mkdir -p "$(dirname "$out")"; \
            CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-linkmode external -extldflags '\''-static'\''" -o "$out" "$f"; \
        done' sh {} +; \
    fi

FROM alpine:3.19 AS packager

RUN apk update && apk add --no-cache zip unzip bash

WORKDIR /lambda

ARG LAMBDA_PATH=
COPY --from=builder /app/cmd/api-backend /app/cmd/api-backend

RUN if [ -n "$LAMBDA_PATH" ]; then \
      src="$LAMBDA_PATH"; \
      case "$src" in *.go) src="${src%.go}";; esac; \
      mkdir -p "/tmp/bin/$(dirname "$src")"; \
      cp -r /app/cmd/api-backend/$src /tmp/bin/$(dirname "$src")/; \
    else \
      cp -r /app/cmd/api-backend /tmp/bin/; \
    fi

RUN cd /tmp/bin && \
    find . -type f -executable ! -name '*.go' | while read bin; do \
        dir=$(dirname "$bin"); \
        name=$(basename "$bin"); \
        mkdir -p "/lambda/$dir"; \
        cp "/tmp/bin/$bin" "/lambda/$dir/$name"; \
        chmod +x "/lambda/$dir/$name"; \
        cd "/lambda/$dir"; \
        zip "${name}.zip" "$name"; \
        rm "$name"; \
        cd - > /dev/null; \
    done

CMD ["ls", "-lR", "/lambda"]