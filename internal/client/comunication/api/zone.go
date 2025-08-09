package api

import (
	"SensorContinuum/internal/client/comunication"
	"SensorContinuum/internal/client/environment"
	"SensorContinuum/pkg/types"
	"encoding/json"
	"errors"
	"strings"
)

const macrozoneNamePlaceholder = "{macrozone}"
const zoneNameSearchPlaceholder = "{name}"

// Restituisce la lista delle zone per una macrozona
func GetZones(regionName, macrozoneName string) ([]types.Zone, error) {
	url := strings.Replace(environment.ZoneListUrl, regionNamePlaceholder, regionName, 1)
	url = strings.Replace(url, macrozoneNamePlaceholder, macrozoneName, 1)
	body, err := comunication.GetApiData(url)
	if err != nil {
		return nil, err
	}
	var zones []types.Zone
	if err := json.Unmarshal([]byte(body), &zones); err != nil {
		return nil, err
	}
	return zones, nil
}

// Restituisce una zona per nome
func GetZoneByName(regionName, macrozoneName, zoneName string) (*types.Zone, error) {
	url := strings.Replace(environment.ZoneSearchNameUrl, regionNamePlaceholder, regionName, 1)
	url = strings.Replace(url, macrozoneNamePlaceholder, macrozoneName, 1)
	url = strings.Replace(url, zoneNameSearchPlaceholder, zoneName, 1)
	body, err := comunication.GetApiData(url)
	if err != nil {
		if errors.Is(err, comunication.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	var zone types.Zone
	if err := json.Unmarshal([]byte(body), &zone); err != nil {
		return nil, err
	}
	return &zone, nil
}
