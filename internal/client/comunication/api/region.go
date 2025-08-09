package api

import (
	"SensorContinuum/internal/client/comunication"
	"SensorContinuum/internal/client/environment"
	"SensorContinuum/pkg/types"
	"encoding/json"
	"errors"
	"strings"
)

const regionNameSearchPlaceholder = "{name}"

func GetRegions() ([]types.Region, error) {
	body, err := comunication.GetApiData(environment.RegionListUrl)
	if err != nil {
		return nil, err
	}
	var regions []types.Region
	if err := json.Unmarshal([]byte(body), &regions); err != nil {
		return nil, err
	}
	return regions, nil
}

func GetRegionByName(name string) (*types.Region, error) {
	url := strings.Replace(environment.RegionSearchNameUrl, regionNameSearchPlaceholder, name, 1)
	body, err := comunication.GetApiData(url)
	if err != nil {
		if errors.Is(err, comunication.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	var region types.Region
	if err := json.Unmarshal([]byte(body), &region); err != nil {
		return nil, err
	}
	return &region, nil
}
