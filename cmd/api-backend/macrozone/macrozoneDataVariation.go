package main

import (
	"SensorContinuum/internal/api-backend/environment"
	macrozoneAPI "SensorContinuum/internal/api-backend/macrozone"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/types"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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

	// Recupera la data dai query parameters (opzionale)
	// Se non è specificata, usa la data di 2 giorni fa
	dateStr := request.QueryStringParameters["date"]
	var date time.Time
	var err error
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

	body, err, statusCode, errorMsg := computeVariations(region, date)
	if err != nil {
		return types.CreateErrorResponse(statusCode, errorMsg, err)
	}

	return events.APIGatewayProxyResponse{
		Body:       string(body),
		StatusCode: http.StatusOK,
		Headers:    map[string]string{"Content-Type": "application/json"},
	}, nil
}

func computeVariations(region string, date time.Time) ([]byte, error, int, string) {

	ctx := context.Background()

	// 1. Ottieni tutte le macrozone
	macrozones, err := macrozoneAPI.GetMacrozonesList(ctx, region)
	if err != nil {
		return nil, err, http.StatusInternalServerError, "Errore nel recupero delle macrozone"
	}
	if len(macrozones) == 0 {
		return nil, err, http.StatusNotFound, "Nessuna macrozona trovata per la regione specificata"
	}

	// 2. Calcola variazione annuale
	macrozoneVariations, err := macrozoneAPI.GetMacrozonesYearlyVariation(ctx, macrozones, date)
	if err != nil {
		return nil, err, http.StatusInternalServerError, "Errore nel calcolo delle variazioni annuali"
	}

	// 3. Serializza in JSON
	body, err := json.MarshalIndent(macrozoneVariations, "", "  ")
	return body, err, http.StatusOK, ""
}

func main() {
	logger.CreateLogger(logger.GetCloudContext())
	lambda.Start(handler)
}
