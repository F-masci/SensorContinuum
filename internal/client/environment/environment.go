package environment

import (
	"SensorContinuum/pkg/logger"
	"github.com/joho/godotenv"
	"os"
	"path/filepath"
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

	logger.Log.Debug("Loading environment variables from .env file")
	err := godotenv.Load(filepath.Join("internal", "client", "environment", ".env"))
	if err != nil {
		logger.Log.Warn("Error loading .env file, using environment variables instead: %v", err)
	} else {
		logger.Log.Debug("Environment variables loaded from .env file")
	}

	var exists bool

	BuildingSearchIdUrl, exists = os.LookupEnv(BuildingSearchIdUrlEnv)
	if !exists {
		return os.ErrNotExist
	}
	logger.Log.Debug("Building search id url: ", BuildingSearchIdUrl)

	BuildingSearchNameUrl, exists = os.LookupEnv(BuildingSearchNameUrlEnv)
	if !exists {
		return os.ErrNotExist
	}
	logger.Log.Debug("Building search name url: ", BuildingSearchNameUrl)

	BuildingListUrl, exists = os.LookupEnv(BuildingListUrlEnv)
	if !exists {
		return os.ErrNotExist
	}
	logger.Log.Debug("Building list url: ", BuildingListUrl)

	RegionListUrl, exists = os.LookupEnv(RegionListUrlEnv)
	if !exists {
		return os.ErrNotExist
	}
	logger.Log.Debug("Region list url: ", RegionListUrl)

	RegionSearchIdUrl, exists = os.LookupEnv(RegionSearchIdUrlEnv)
	if !exists {
		return os.ErrNotExist
	}
	logger.Log.Debug("Region search id url: ", RegionSearchIdUrl)

	RegionSearchNameUrl, exists = os.LookupEnv(RegionSearchNameUrlEnv)
	if !exists {
		return os.ErrNotExist
	}
	logger.Log.Debug("Region search name url: ", RegionSearchNameUrl)

	return nil
}
