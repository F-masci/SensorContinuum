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
	region := request.PathParameters["region"]
	daysStr := request.QueryStringParameters["days"]

	days, err := strconv.Atoi(daysStr)
	if err != nil {
		days = 60 // default
	}

	ctx := context.Background()
	macrozones, err := macrozoneAPI.GetMacrozonesList(ctx, region)
	if err != nil {
		errBody, _ := json.Marshal(types.ErrorResponse{
			Error:  "Errore nel recupero delle macrozone",
			Detail: err.Error(),
		})
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       string(errBody),
			Headers:    map[string]string{"Content-Type": "application/json"},
		}, nil
	}

	trends, err := macrozoneAPI.GetTrendSimilarityPerMacrozones(ctx, macrozones, days)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, err
	}

	body, _ := json.Marshal(trends)
	return events.APIGatewayProxyResponse{
		Body:       string(body),
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json"},
	}, nil
}

func main() {
	lambda.Start(handler)
}
