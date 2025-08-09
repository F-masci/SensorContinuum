package api

import (
	"SensorContinuum/internal/client/comunication"
	"SensorContinuum/internal/client/environment"
	"SensorContinuum/pkg/types"
	"encoding/json"
	"errors"
	"strings"
)

const regionNamePlaceholder = "{region}"
const macrozoneNameSearchPlaceholder = "{name}"

func GetMacrozones(regionName string) ([]types.Macrozone, error) {
	url := strings.Replace(environment.MacrozoneListUrl, regionNamePlaceholder, regionName, 1)
	body, err := comunication.GetApiData(url)
	if err != nil {
		return nil, err
	}
	var buildings []types.Macrozone
	if err := json.Unmarshal([]byte(body), &buildings); err != nil {
		return nil, err
	}
	return buildings, nil
}

func GetMacrozoneByName(regionName string, name string) (*types.Macrozone, error) {
	url := strings.Replace(environment.MacrozoneSearchNameUrl, regionNamePlaceholder, regionName, 1)
	url = strings.Replace(url, macrozoneNameSearchPlaceholder, name, 1)
	body, err := comunication.GetApiData(url)
	if err != nil {
		if errors.Is(err, comunication.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	var building types.Macrozone
	if err := json.Unmarshal([]byte(body), &building); err != nil {
		return nil, err
	}
	return &building, nil
}
