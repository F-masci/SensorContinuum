package comunication

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func GetApiData(endpoint string) (string, error) {
	if endpoint == "" {
		return "", fmt.Errorf("endpoint vuoto: controlla la configurazione")
	}
	resp, err := http.Get(endpoint)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("errore HTTP: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
