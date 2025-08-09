package environment

import (
	"SensorContinuum/pkg/logger"
	"github.com/joho/godotenv"
	"os"
	"path/filepath"
)

const (
	RegionListUrlEnv          = "REGION_LIST_URL"
	RegionSearchNameUrlEnv    = "REGION_SEARCH_NAME_URL"
	MacrozoneListUrlEnv       = "MACROZONE_LIST_URL"
	MacrozoneSearchNameUrlEnv = "MACROZONE_SEARCH_NAME_URL"
	ZoneListUrlEnv            = "ZONE_LIST_URL"
	ZoneSearchNameUrlEnv      = "ZONE_SEARCH_NAME_URL"
)

var (
	RegionListUrl          string
	RegionSearchNameUrl    string
	MacrozoneListUrl       string
	MacrozoneSearchNameUrl string
	ZoneListUrl            string
	ZoneSearchNameUrl      string
)

// UnhealthyTime indica quanto tempo (in minuti) le risorse possono non comunicare
// prima di essere mostrate come non più attive.
var UnhealthyTime = 10

func SetupEnvironment() error {

	logger.Log.Debug("Loading environment variables from .env file")
	err := godotenv.Load(filepath.Join("internal", "client", "environment", ".env"))
	if err != nil {
		logger.Log.Warn("Error loading .env file, using environment variables instead: %v", err)
	} else {
		logger.Log.Debug("Environment variables loaded from .env file")
	}

	var exists bool

	RegionListUrl, exists = os.LookupEnv(RegionListUrlEnv)
	if !exists {
		return os.ErrNotExist
	}
	logger.Log.Debug("Region list url: ", RegionListUrl)

	RegionSearchNameUrl, exists = os.LookupEnv(RegionSearchNameUrlEnv)
	if !exists {
		return os.ErrNotExist
	}
	logger.Log.Debug("Region search name url: ", RegionSearchNameUrl)

	MacrozoneListUrl, exists = os.LookupEnv(MacrozoneListUrlEnv)
	if !exists {
		return os.ErrNotExist
	}
	logger.Log.Debug("Macrozone list url: ", MacrozoneListUrl)

	MacrozoneSearchNameUrl, exists = os.LookupEnv(MacrozoneSearchNameUrlEnv)
	if !exists {
		return os.ErrNotExist
	}
	logger.Log.Debug("Macrozone search name url: ", MacrozoneSearchNameUrl)

	ZoneListUrl, exists = os.LookupEnv(ZoneListUrlEnv)
	if !exists {
		return os.ErrNotExist
	}
	logger.Log.Debug("Zone list url: ", ZoneListUrl)

	ZoneSearchNameUrl, exists = os.LookupEnv(ZoneSearchNameUrlEnv)
	if !exists {
		return os.ErrNotExist
	}
	logger.Log.Debug("Zone search name url: ", ZoneSearchNameUrl)

	return nil
}
