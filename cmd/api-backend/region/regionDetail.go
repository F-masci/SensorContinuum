package main

import (
	"SensorContinuum/pkg/structure"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	id := request.PathParameters["id"]

	// Esempio: cerca la regione con quell'id
	regions := []structure.Region{
		{Name: "region-001", Longitude: 9.19, Latitude: 45.46},
		{Name: "region-002", Longitude: 13.36, Latitude: 38.11},
		{Name: "region-003", Longitude: 12.50, Latitude: 41.89},
	}

	var found *structure.Region
	for _, r := range regions {
		if r.Name == id {
			found = &r
			break
		}
	}

	if found == nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusNotFound,
			Body:       `{"error":"Regione non trovata"}`,
			Headers:    map[string]string{"Content-Type": "application/json"},
		}, nil
	}

	body, err := json.Marshal(found)
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
	lambda.Start(handler)
}
