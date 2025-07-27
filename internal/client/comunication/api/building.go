package api

import (
	"SensorContinuum/internal/client/comunication"
	"SensorContinuum/internal/client/environment"
	"SensorContinuum/pkg/structure"
	"encoding/json"
	"strings"
)

func GetBuildings() ([]structure.Building, error) {
	body, err := comunication.GetApiData(environment.BuildingListUrl)
	if err != nil {
		return nil, err
	}
	var buildings []structure.Building
	if err := json.Unmarshal([]byte(body), &buildings); err != nil {
		return nil, err
	}
	return buildings, nil
}

func GetBuildingByID(id string) (*structure.Building, error) {
	url := strings.Replace(environment.BuildingSearchIdUrl, "{id}", id, 1)
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

func GetBuildingByName(name string) (*structure.Building, error) {
	url := strings.Replace(environment.BuildingSearchNameUrl, "{name}", name, 1)
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

func GetBuildingsByRegion(regionID string) ([]structure.Building, error) {
	url := strings.Replace(environment.BuildingSearchRegionUrl, "{region_id}", regionID, 1)
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
