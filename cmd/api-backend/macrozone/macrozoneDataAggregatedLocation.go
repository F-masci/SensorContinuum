package main

import (
	"SensorContinuum/internal/api-backend/macrozone"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/types"
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Recupera parametri dalla query string
	latStr := request.QueryStringParameters["lat"]
	lonStr := request.QueryStringParameters["lon"]
	radiusStr := request.QueryStringParameters["radius"]

	// Converte i parametri in float/int
	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		return types.CreateErrorResponse(http.StatusBadRequest, "Parametro 'lat' non valido", err)
	}
	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		return types.CreateErrorResponse(http.StatusBadRequest, "Parametro 'lon' non valido", err)
	}
	radius, err := strconv.ParseFloat(radiusStr, 64)
	if err != nil {
		return types.CreateErrorResponse(http.StatusBadRequest, "Parametro 'radius' non valido", err)
	}

	body, err, statusCode, errorMsg := computeAggregatedDataByLocation(lat, lon, radius)
	if err != nil {
		return types.CreateErrorResponse(statusCode, errorMsg, err)
	}

	return events.APIGatewayProxyResponse{
		Body:       string(body),
		StatusCode: http.StatusOK,
		Headers:    map[string]string{"Content-Type": "application/json"},
	}, nil
}

func computeAggregatedDataByLocation(lat, lon, radius float64) ([]byte, error, int, string) {
	ctx := context.Background()
	aggregatedStats, err := macrozone.GetAggregatedDataByLocation(ctx, lat, lon, radius)
	if err != nil {
		return nil, err, http.StatusInternalServerError, "Errore nel recupero dei dati aggregati"
	}

	body, err := json.MarshalIndent(aggregatedStats, "", "  ")
	if err != nil {
		return nil, err, http.StatusInternalServerError, "Errore nella serializzazione dei dati"
	}

	return body, nil, http.StatusOK, ""
}

func main() {
	logger.CreateLogger(logger.GetCloudContext())
	lambda.Start(handler)
}
