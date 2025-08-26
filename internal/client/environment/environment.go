package environment

import (
	"SensorContinuum/pkg/logger"
	"github.com/joho/godotenv"
	"os"
	"path/filepath"
)

const (
	RegionListUrlEnv              = "REGION_LIST_URL"
	RegionSearchNameUrlEnv        = "REGION_SEARCH_NAME_URL"
	RegionAggregatedDataUrlEnv    = "REGION_DATA_AGGREGATED_URL"
	MacrozoneListUrlEnv           = "MACROZONE_LIST_URL"
	MacrozoneSearchNameUrlEnv     = "MACROZONE_SEARCH_NAME_URL"
	MacrozoneAggregatedDataUrlEnv = "MACROZONE_DATA_AGGREGATED_URL"
	ZoneListUrlEnv                = "ZONE_LIST_URL"
	ZoneSearchNameUrlEnv          = "ZONE_SEARCH_NAME_URL"
	ZoneRawSensorDataUrlEnv       = "ZONE_SENSOR_DATA_RAW_URL"
	ZoneAggregatedDataUrlEnv      = "ZONE_DATA_AGGREGATED_URL"
)

var (
	RegionListUrl              string
	RegionSearchNameUrl        string
	RegionAggregatedDataUrl    string
	MacrozoneListUrl           string
	MacrozoneSearchNameUrl     string
	MacrozoneAggregatedDataUrl string
	ZoneListUrl                string
	ZoneSearchNameUrl          string
	ZoneRawSensorDataUrl       string
	ZoneAggregatedDataUrl      string
)

// UnhealthyTime indica quanto tempo (in minuti) le risorse possono non comunicare
// prima di essere mostrate come non più attive.
var UnhealthyTime = 10

func SetupEnvironment() error {

	logger.Log.Debug("Loading environment variables from .env.development file")
	err := godotenv.Load(filepath.Join("internal", "client", "environment", ".env"))
	if err != nil {
		logger.Log.Warn("Error loading .env.development file, using environment variables instead: %v", err)
	} else {
		logger.Log.Debug("Environment variables loaded from .env.development file")
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

	RegionAggregatedDataUrl, exists = os.LookupEnv(RegionAggregatedDataUrlEnv)
	if !exists {
		return os.ErrNotExist
	}
	logger.Log.Debug("Region aggregated data url: ", RegionAggregatedDataUrl)

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

	MacrozoneAggregatedDataUrl, exists = os.LookupEnv(MacrozoneAggregatedDataUrlEnv)
	if !exists {
		return os.ErrNotExist
	}
	logger.Log.Debug("Macrozone aggregated data url: ", MacrozoneAggregatedDataUrl)

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

	ZoneRawSensorDataUrl, exists = os.LookupEnv(ZoneRawSensorDataUrlEnv)
	if !exists {
		return os.ErrNotExist
	}
	logger.Log.Debug("Zone raw sensor data url: ", ZoneRawSensorDataUrl)

	ZoneAggregatedDataUrl, exists = os.LookupEnv(ZoneAggregatedDataUrlEnv)
	if !exists {
		return os.ErrNotExist
	}
	logger.Log.Debug("Zone aggregated data url: ", ZoneAggregatedDataUrl)

	return nil
}
