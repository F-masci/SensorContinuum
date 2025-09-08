package main

import (
	"SensorContinuum/internal/api-backend/region"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/types"
	"context"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	name := request.PathParameters["name"]

	ctx := context.Background()
	regionDetail, err := region.GetRegionByName(ctx, name)
	if err != nil {
		errBody, _ := json.Marshal(types.ErrorResponse{
			Error:  "Errore nel recupero della regione",
			Detail: err.Error(),
		})
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       string(errBody),
			Headers:    map[string]string{"Content-Type": "application/json"},
		}, nil
	}
	if regionDetail == nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusNotFound,
			Body:       `{"error":"Regione non trovata"}`,
			Headers:    map[string]string{"Content-Type": "application/json"},
		}, nil
	}

	body, err := json.Marshal(regionDetail)
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
