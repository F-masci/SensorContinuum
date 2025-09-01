package main

import (
	macrozoneAPI "SensorContinuum/internal/api-backend/macrozone"
	"SensorContinuum/internal/api-backend/storage"
	"SensorContinuum/pkg/types"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Recupera la regione dai path parameters
	region := request.PathParameters["region"]
	if region == "" {
		return types.CreateErrorResponse(http.StatusBadRequest, "Parametro 'region' mancante", nil)
	}

	// Recupera il raggio dai query parameters
	radiusStr := request.QueryStringParameters["radius"]
	radius, err := strconv.ParseFloat(radiusStr, 64)
	if err != nil {
		return types.CreateErrorResponse(http.StatusBadRequest, "Parametro 'radius' non valido", err)
	}

	// Recupera la data dai query parameters (opzionale)
	// Se non è specificata, usa la data di ieri
	dateStr := request.QueryStringParameters["date"]
	var date time.Time
	if dateStr == "" {
		date = time.Now().AddDate(0, 0, -1)
	} else {
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return types.CreateErrorResponse(http.StatusBadRequest, "Parametro 'date' non valido, deve essere nel formato YYYY-MM-DD", err)
		}
	}

	body, err, statusCode, errorMsg := computeVariationCorrelations(region, radius, date)
	if err != nil {
		return types.CreateErrorResponse(statusCode, errorMsg, err)
	}

	return events.APIGatewayProxyResponse{
		Body:       string(body),
		StatusCode: http.StatusOK,
		Headers:    map[string]string{"Content-Type": "application/json"},
	}, nil
}

func computeVariationCorrelations(region string, radius float64, date time.Time) ([]byte, error, int, string) {

	ctx := context.Background()
	_, err := storage.GetSensorPostgresDB(ctx, region)
	if err != nil {
		return nil, err, http.StatusInternalServerError, "Errore di connessione al database dei sensori"
	}

	// 1. Ottieni tutte le macrozone
	macrozones, err := macrozoneAPI.GetMacrozonesList(ctx, region)
	if err != nil {
		return nil, err, http.StatusInternalServerError, "Errore nel recupero delle macrozone"
	}
	if len(macrozones) == 0 {
		return nil, err, http.StatusNotFound, "Nessuna macrozona trovata per la regione specificata"
	}

	// 2. Calcola anomalie
	macrozoneAnomalies, err := macrozoneAPI.GetMacrozonesAnomalies(ctx, macrozones, radius, date)
	if err != nil {
		return nil, err, http.StatusInternalServerError, "Errore nel calcolo anomalie"
	}

	// 3. Serializza in JSON
	body, err := json.MarshalIndent(macrozoneAnomalies, "", "  ")
	return body, err, http.StatusOK, ""
}

func main() {
	_, exists := os.LookupEnv("AWS_LAMBDA_RUNTIME_API")
	if exists {
		// La funzione è in esecuzione in AWS Lambda
		lambda.Start(handler)
	} else {
		// La funzione è in esecuzione in locale
		body, err, _, _ := computeVariationCorrelations("region-001", 15_000.0, time.Now().AddDate(0, 0, -1))
		if err != nil {
			panic(err)
		}
		fmt.Println(string(body))
	}
}
