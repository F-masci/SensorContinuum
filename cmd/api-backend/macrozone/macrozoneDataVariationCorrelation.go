package main

import (
	macrozoneAPI "SensorContinuum/internal/api-backend/macrozone"
	"SensorContinuum/pkg/types"
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Recupera la regione dai path parameters
	region := request.PathParameters["region"]
	if region == "" {
		return types.CreateErrorResponse(http.StatusBadRequest, "Parametro 'region' mancante", nil)
	}

	radiusStr := request.QueryStringParameters["radius"]
	radius, err := strconv.ParseFloat(radiusStr, 64)
	if err != nil {
		return types.CreateErrorResponse(http.StatusBadRequest, "Parametro 'radius' non valido", err)
	}

	ctx := context.Background()

	// 1. Ottieni tutte le macrozone
	macrozones, err := macrozoneAPI.GetMacrozonesList(ctx, region)
	if err != nil {
		return types.CreateErrorResponse(http.StatusInternalServerError, "Errore nel recupero delle macrozone", err)
	}
	if len(macrozones) == 0 {
		return types.CreateErrorResponse(http.StatusNotFound, "Nessuna macrozona trovata per la regione specificata", nil)
	}

	// 2. Calcola anomalie
	macrozoneAnomalies, err := macrozoneAPI.CalculateAnomaliesPerMacrozones(ctx, macrozones, radius)
	if err != nil {
		return types.CreateErrorResponse(http.StatusInternalServerError, "Errore nel calcolo anomalie", err)
	}

	// 3. Serializza in JSON
	body, _ := json.MarshalIndent(macrozoneAnomalies, "", "  ")

	return events.APIGatewayProxyResponse{
		Body:       string(body),
		StatusCode: http.StatusOK,
		Headers:    map[string]string{"Content-Type": "application/json"},
	}, nil
}

func main() {
	lambda.Start(handler)
}
