package types

type ErrorResponse struct {
	Error  string `json:"error"`
	Detail string `json:"detail"`
}
