package main

import (
	"SensorContinuum/internal/api-backend/environment"
	macrozoneAPI "SensorContinuum/internal/api-backend/macrozone"
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

	err := environment.SetupEnvironment()
	if err != nil {
		return types.CreateErrorResponse(http.StatusInternalServerError, "Errore nella configurazione dell'ambiente", err)
	}

	// Recupera la regione dai path parameters
	region := request.PathParameters["region"]
	// Recupera i giorni dai query parameters
	daysStr := request.QueryStringParameters["days"]
	if daysStr == "" {
		daysStr = "60" // default
	}
	days, err := strconv.Atoi(daysStr)
	if err != nil {
		return types.CreateErrorResponse(http.StatusBadRequest, "Parametro 'days' non valido", err)
	}

	// Recupera la data dai query parameters (opzionale)
	// Se non è specificata, usa la data di 2 giorni fa
	dateStr := request.QueryStringParameters["date"]
	var date time.Time
	if dateStr == "" {
		date = time.Now().AddDate(0, 0, -2)
	} else {
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return types.CreateErrorResponse(http.StatusBadRequest, "Parametro 'date' non valido, deve essere nel formato YYYY-MM-DD", err)
		}
		today := time.Now().Truncate(24 * time.Hour)
		maxDate := today.Add(-environment.YearlyVariationMinimum)
		// La data non può essere successiva a due giorni fa
		// (perché i dati di ieri potrebbero non essere ancora completi)
		if date.Truncate(24 * time.Hour).After(maxDate) {
			return types.CreateErrorResponse(http.StatusBadRequest, "Parametro 'date' non valido, deve essere antecedente a ieri", fmt.Errorf("data futura"))
		}
	}

	body, err, statusCode, errorMsg := computeTrends(region, days, date)
	if err != nil {
		return types.CreateErrorResponse(statusCode, errorMsg, err)
	}

	return events.APIGatewayProxyResponse{
		Body:       string(body),
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json"},
	}, nil
}

func computeTrends(region string, days int, date time.Time) ([]byte, error, int, string) {

	ctx := context.Background()

	// 1. Ottieni tutte le macrozone
	macrozones, err := macrozoneAPI.GetMacrozonesList(ctx, region)
	if err != nil {
		return nil, err, http.StatusInternalServerError, "Errore nel recupero delle macrozone"
	}
	if len(macrozones) == 0 {
		return nil, err, http.StatusNotFound, "Nessuna macrozona trovata per la regione specificata"
	}

	// 2. Calcola le tendenze per tutte le macrozone
	trends, err := macrozoneAPI.GetMacrozonesTrendsSimilarity(ctx, macrozones, days, date)
	if err != nil {
		return nil, err, http.StatusInternalServerError, "Errore nel calcolo delle tendenze"
	}

	// 3. Serializza in JSON
	body, err := json.MarshalIndent(trends, "", "  ")
	return body, err, http.StatusOK, ""
}

func main() {
	_, exists := os.LookupEnv("AWS_LAMBDA_RUNTIME_API")
	if exists {
		// La funzione è in esecuzione in AWS Lambda
		lambda.Start(handler)
	} else {
		// La funzione è in esecuzione in locale
		body, err, _, _ := computeTrends("region-001", 60, time.Now().AddDate(0, 0, -1))
		if err != nil {
			panic(err)
		}
		fmt.Println(string(body))
	}
}
