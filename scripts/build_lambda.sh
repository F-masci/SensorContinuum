#!/usr/bin/env bash
set -euo pipefail

LAMBDA_NAME=${1:-}
BUILD_DIR="lambda_build"
OUTPUT_DIR="lambda"

# Pulizia cartelle
rm -rf "$BUILD_DIR" "$OUTPUT_DIR"
mkdir -p "$BUILD_DIR" "$OUTPUT_DIR"

# Funzione per buildare un singolo file Go
build_lambda() {
    local src="$1"
    local rel_path
    rel_path=$(realpath --relative-to="./cmd/api-backend" "$src")
    local build_out="$BUILD_DIR/$rel_path"
    mkdir -p "$(dirname "$build_out")"
    echo "Building $src -> $build_out"

    # Compila per Linux AMD64, CGO disabilitato
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o "$build_out" "$src"

    # Crea lo zip mantenendo la struttura delle cartelle
    local zip_out="$OUTPUT_DIR/${rel_path%.go}.zip"
    mkdir -p "$(dirname "$zip_out")"

    # Copia temporanea e rinomina il binario in 'bootstrap' (per Lambda custom runtime)
    tmp_file="$(dirname "$build_out")/bootstrap"
    cp "$build_out" "$tmp_file"
    zip -j "$zip_out" "$tmp_file"
    rm "$tmp_file"
}

# Compilazione
if [ -n "$LAMBDA_NAME" ]; then
    # singola lambda specifica
    if [[ "$LAMBDA_NAME" != *.go ]]; then
        LAMBDA_NAME="$LAMBDA_NAME.go"
    fi
    build_lambda "cmd/api-backend/$LAMBDA_NAME"
else
    # tutte le lambda
    find ./cmd/api-backend -type f -name '*.go' | while read -r file; do
        build_lambda "$file"
    done
fi

# Pulizia
rm -rf "$BUILD_DIR"

echo "Lambda build completata. File .zip compatibili con runtime provided.al2 / provided.al2023 disponibili in ./$OUTPUT_DIR/"
