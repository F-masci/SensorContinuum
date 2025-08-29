package types

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
)

// ErrorResponse rappresenta la struttura di una risposta di errore
type ErrorResponse struct {
	Error  string `json:"error"`
	Detail string `json:"detail"`
}

// Event rappresenta i parametri che passerai tramite EventBridge
type Event struct {
	Region string `json:"region"`
	Days   int    `json:"days"`
}

// Funzione di supporto per gestire errori
func CreateErrorResponse(status int, msg string, err error) (events.APIGatewayProxyResponse, error) {
	errResp := ErrorResponse{
		Error: msg,
	}
	if err != nil {
		errResp.Detail = err.Error()
	}
	errBody, _ := json.Marshal(errResp)

	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Body:       string(errBody),
		Headers:    map[string]string{"Content-Type": "application/json"},
	}, nil
}
