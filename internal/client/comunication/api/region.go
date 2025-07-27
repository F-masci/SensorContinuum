package api

import (
	"SensorContinuum/internal/client/comunication"
	"SensorContinuum/internal/client/environment"
	"SensorContinuum/pkg/structure"
	"encoding/json"
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

func GetRegionById(id string) (*structure.Region, error) {
	url := strings.Replace(environment.RegionSearchIdUrl, "{id}", id, 1)
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

func GetRegionByName(name string) (*structure.Region, error) {
	url := strings.Replace(environment.RegionSearchNameUrl, "{name}", name, 1)
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
