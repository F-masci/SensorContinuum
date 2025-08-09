package comunication

import (
	"errors"
	"fmt"
	"io"
	"net/http"
)

var ErrNotFound = errors.New("resource not found")

func GetApiData(endpoint string) (string, error) {
	if endpoint == "" {
		return "", fmt.Errorf("empty endpoint: please check your configuration")
	}
	resp, err := http.Get(endpoint)
	if err != nil {
		return "", fmt.Errorf("failed to perform HTTP GET request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("error closing response body: %v\n", err)
		}
	}(resp.Body)

	if resp.StatusCode == http.StatusNotFound {
		return "", ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected HTTP status code: got %s, expected 200 OK", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}
