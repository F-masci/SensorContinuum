package main

import (
	zoneAPI "SensorContinuum/internal/api-backend/zone"
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
	region := request.PathParameters["region"]
	macrozone := request.PathParameters["macrozone"]
	zone := request.PathParameters["zone"]
	sensor := request.PathParameters["sensor"]

	var limit int
	limitStr := request.QueryStringParameters["limit"]
	if limitStr == "" {
		limit = 50
	} else {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit < 0 {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusBadRequest,
				Body:       `{"error":"Parametro 'limit' non valido"}`,
				Headers:    map[string]string{"Content-Type": "application/json"},
			}, nil
		}
	}

	ctx := context.Background()
	sensorData, err := zoneAPI.GetRawSensorData(ctx, region, macrozone, zone, sensor, limit)
	if err != nil {
		errBody, _ := json.Marshal(types.ErrorResponse{
			Error:  "Errore nel recupero dei dati del sensore",
			Detail: err.Error(),
		})
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       string(errBody),
			Headers:    map[string]string{"Content-Type": "application/json"},
		}, nil
	}
	if sensorData == nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusNotFound,
			Body:       `{"error":"Dati non trovati"}`,
			Headers:    map[string]string{"Content-Type": "application/json"},
		}, nil
	}

	body, err := json.Marshal(sensorData)
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
