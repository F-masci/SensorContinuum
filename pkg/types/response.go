package types

// ErrorResponse rappresenta la struttura di una risposta di errore
type ErrorResponse struct {
	Error  string `json:"error"`
	Detail string `json:"detail"`
}
