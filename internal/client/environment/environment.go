package environment

import (
	"SensorContinuum/pkg/logger"
	"github.com/joho/godotenv"
	"os"
)

const (
	BuildingSearchIdUrlEnv     = "BUILDING_SEARCH_ID_URL"
	BuildingSearchNameUrlEnv   = "BUILDING_SEARCH_NAME_URL"
	BuildingSearchRegionUrlEnv = "BUILDING_SEARCH_REGION_URL"
	BuildingListUrlEnv         = "BUILDING_LIST_URL"
	RegionListUrlEnv           = "REGION_LIST_URL"
	RegionSearchIdUrlEnv       = "REGION_SEARCH_ID_URL"
	RegionSearchNameUrlEnv     = "REGION_SEARCH_NAME_URL"
)

var (
	BuildingSearchIdUrl     string
	BuildingSearchNameUrl   string
	BuildingSearchRegionUrl string
	BuildingListUrl         string
	RegionListUrl           string
	RegionSearchIdUrl       string
	RegionSearchNameUrl     string
)

func SetupEnvironment() error {

	err := godotenv.Load("internal/client/environment/.env")
	if err != nil {
		logger.Log.Warn("Error loading .env file, using environment variables instead: %v", err)
	} else {
		logger.Log.Info("Environment variables loaded from .env file")
	}

	var exists bool

	BuildingSearchIdUrl, exists = os.LookupEnv(BuildingSearchIdUrlEnv)
	if !exists {
		return os.ErrNotExist
	}
	logger.Log.Debug("Building search id url: %s", BuildingSearchIdUrl)

	BuildingSearchNameUrl, exists = os.LookupEnv(BuildingSearchNameUrlEnv)
	if !exists {
		return os.ErrNotExist
	}
	logger.Log.Debug("Building search name url: %s", BuildingSearchNameUrl)

	BuildingSearchRegionUrl, exists = os.LookupEnv(BuildingSearchRegionUrlEnv)
	if !exists {
		return os.ErrNotExist
	}
	logger.Log.Debug("Building search region url: %s", BuildingSearchRegionUrl)

	BuildingListUrl, exists = os.LookupEnv(BuildingListUrlEnv)
	if !exists {
		return os.ErrNotExist
	}
	logger.Log.Debug("Building list url: %s", BuildingListUrl)

	RegionListUrl, exists = os.LookupEnv(RegionListUrlEnv)
	if !exists {
		return os.ErrNotExist
	}
	logger.Log.Debug("Region list url: %s", RegionListUrl)

	RegionSearchIdUrl, exists = os.LookupEnv(RegionSearchIdUrlEnv)
	if !exists {
		return os.ErrNotExist
	}
	logger.Log.Debug("Region search id url: %s", RegionSearchIdUrl)

	RegionSearchNameUrl, exists = os.LookupEnv(RegionSearchNameUrlEnv)
	if !exists {
		return os.ErrNotExist
	}
	logger.Log.Debug("Region search name url: %s", RegionSearchNameUrl)

	return nil
}
