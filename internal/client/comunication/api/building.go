package api

import (
	"SensorContinuum/internal/client/comunication"
	"SensorContinuum/internal/client/environment"
	"SensorContinuum/pkg/structure"
	"encoding/json"
	"strings"
)

const regionNamePlaceholder = "{region}"

func GetBuildings(regionName string) ([]structure.Building, error) {
	url := strings.Replace(environment.BuildingListUrl, regionNamePlaceholder, regionName, 1)
	body, err := comunication.GetApiData(url)
	if err != nil {
		return nil, err
	}
	var buildings []structure.Building
	if err := json.Unmarshal([]byte(body), &buildings); err != nil {
		return nil, err
	}
	return buildings, nil
}

func GetBuildingByID(regionName string, id string) (*structure.Building, error) {
	url := strings.Replace(environment.BuildingSearchIdUrl, regionNamePlaceholder, regionName, 1)
	url = strings.Replace(url, "{id}", id, 1)
	body, err := comunication.GetApiData(url)
	if err != nil {
		return nil, err
	}
	var building structure.Building
	if err := json.Unmarshal([]byte(body), &building); err != nil {
		return nil, err
	}
	return &building, nil
}

func GetBuildingByName(regionName string, name string) (*structure.Building, error) {
	url := strings.Replace(environment.BuildingSearchNameUrl, regionNamePlaceholder, regionName, 1)
	url = strings.Replace(url, "{name}", name, 1)
	body, err := comunication.GetApiData(url)
	if err != nil {
		return nil, err
	}
	var building structure.Building
	if err := json.Unmarshal([]byte(body), &building); err != nil {
		return nil, err
	}
	return &building, nil
}
