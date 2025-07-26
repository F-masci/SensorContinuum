package api

import (
	"SensorContinuum/internal/client/comunication"
	"SensorContinuum/internal/client/environment"
	"SensorContinuum/pkg/structure"
	// "bytes"
	"encoding/json"
	// "fmt"
	// "io/ioutil"
	// "net/http"
	"strings"
)

func GetRegions() ([]structure.Region, error) {
	body, err := comunication.GetApiData(environment.RegionListUrl)
	if err != nil {
		return nil, err
	}
	var regions []structure.Region
	if err := json.Unmarshal([]byte(body), &regions); err != nil {
		return nil, err
	}
	return regions, nil
}

func GetRegion(name string) (*structure.Region, error) {
	url := strings.Replace(environment.RegionDetailUrl, "{id}", name, 1)
	body, err := comunication.GetApiData(url)
	if err != nil {
		return nil, err
	}
	var region structure.Region
	if err := json.Unmarshal([]byte(body), &region); err != nil {
		return nil, err
	}
	return &region, nil
}

/*func CreateRegion(region structure.Region) error {
	url := strings.Replace(environment.RegionUrl, "{id}", name, 1)
	body, err := comunication.GetApiData(url)
	data, _ := json.Marshal(region)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("errore HTTP: %s, %s", resp.Status, string(body))
	}
	return nil
}

func UpdateRegion(region structure.Region) error {
	url := strings.Replace(environment.RegionUrl, "{id}", name, 1)
	body, err := comunication.GetApiData(url)
	data, _ := json.Marshal(region)
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("errore HTTP: %s, %s", resp.Status, string(body))
	}
	return nil
}

func DeleteRegion(name string) error {
	url := strings.Replace(environment.RegionUrl, "{id}", name, 1)
	body, err := comunication.GetApiData(url)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("errore HTTP: %s, %s", resp.Status, string(body))
	}
	return nil
}
*/
