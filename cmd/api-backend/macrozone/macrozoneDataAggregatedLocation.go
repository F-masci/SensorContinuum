package main

import (
	"SensorContinuum/internal/api-backend/macrozone"
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

	// Esegue la logica applicativa
	ctx := context.Background()
	aggregatedStats, err := macrozone.GetAggregatedDataByLocation(ctx, lat, lon, radius)
	if err != nil {
		return types.CreateErrorResponse(http.StatusInternalServerError, "Errore nel recupero dei dati aggregati", err)
	}

	// Serializza in JSON
	body, _ := json.MarshalIndent(aggregatedStats, "", "  ")

	return events.APIGatewayProxyResponse{
		Body:       string(body),
		StatusCode: http.StatusOK,
		Headers:    map[string]string{"Content-Type": "application/json"},
	}, nil
}

func main() {
	lambda.Start(handler)
}
