package main

import (
	"SensorContinuum/internal/api-backend/zone"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/types"
	"context"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	region := request.PathParameters["region"]
	macrozone := request.PathParameters["macrozone"]
	name := request.PathParameters["name"]

	ctx := context.Background()
	zoneDetail, err := zone.GetZoneByName(ctx, region, macrozone, name)
	if err != nil {
		errBody, _ := json.Marshal(types.ErrorResponse{
			Error:  "Errore nel recupero delle zona",
			Detail: err.Error(),
		})
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       string(errBody),
			Headers:    map[string]string{"Content-Type": "application/json"},
		}, nil
	}
	if zoneDetail == nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusNotFound,
			Body:       `{"error":"Zona non trovata"}`,
			Headers:    map[string]string{"Content-Type": "application/json"},
		}, nil
	}

	body, err := json.Marshal(zoneDetail)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, err
	}

	return events.APIGatewayProxyResponse{
		Body:       string(body),
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json"},
	}, nil
}

func main() {
	logger.CreateLogger(logger.GetCloudContext())
	lambda.Start(handler)
}
