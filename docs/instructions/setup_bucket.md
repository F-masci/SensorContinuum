# Script per il caricamento degli Asset su S3

Questo script (`deploy/scripts/setup_bucket.sh`) è essenziale per la fase di **preparazione del deployment**. Si occupa di creare l'infrastruttura di storage necessaria su AWS S3 e di popolare il bucket con tutti i file che saranno scaricati dalle istanze EC2 al momento dell'avvio.

### Esempi di Utilizzo

```bash
# Esegue l'upload di tutti gli asset predefiniti nel bucket S3 specificato
./setup_bucket.sh

# Se si desidera specificare solo alcuni file e saltare la creazione del bucket:
./setup_bucket.sh --no-create docker-install.sh analyze-failure.sh
```

### Logica Operativa Dettagliata

1.  **Verifica e Creazione Bucket**: Controlla l'esistenza del bucket S3 (nome di default: `sensor-continuum-scripts`). Se non esiste e non è specificato il flag `--no-create`, il bucket viene creato nella regione AWS definita (`$AWS_REGION`).

    ```bash
    echo "Verifica bucket $BUCKET_NAME..."
    if [[ $NO_CREATE -eq 0 ]]; then
      # ... (codice per verificare l'esistenza)
      if [[ "$EXISTS" == "no" ]]; then
        echo "Bucket non trovato. Creazione bucket $BUCKET_NAME in $AWS_REGION..."
        aws s3api $ENDPOINT_URL create-bucket \
            --bucket "$BUCKET_NAME" \
            --region "$AWS_REGION" \
            $( [[ "$AWS_REGION" != "us-east-1" ]] && echo "--create-bucket-configuration LocationConstraint=$AWS_REGION" )
        echo "Bucket creato."
      fi
    fi
    ```

2.  **Mappatura e Caricamento File**: Una mappa associativa (`FILE_MAP`) definisce le associazioni tra i pattern dei file locali e le cartelle di destinazione all'interno del bucket S3, garantendo l'organizzazione degli asset:

    | File Locale (Pattern) | Cartella di Destinazione S3 | Contenuto Tipico |
        | :--- | :--- | :--- |
    | `inits/*install*.sh` | `init/` | Script di installazione (es. `docker-install.sh`, `init-delay.sh`) |
    | `deploy/*deploy*.sh` | `deploy/` | Script di avvio (es. `deploy_edge_services.sh`) |
    | `../compose/*.y*ml` | `compose/` | File Docker Compose e template YAML |
    | `performance/analyze_*.sh` | `performance/` | Script di analisi delle performance |

    Lo script itera su questa mappa e utilizza il comando `aws s3 cp` per copiare i file:

    ```bash
    for pattern in "${!FILE_MAP[@]}"; do
      dest_path="${FILE_MAP[$pattern]}"
      # ... (logica per selezionare i file e copiare)
      aws s3 cp $ENDPOINT_URL "$f" "s3://$BUCKET_NAME/$dest_path"
      echo "Copiato $f → s3://$BUCKET_NAME/$dest_path"
    done
    ```