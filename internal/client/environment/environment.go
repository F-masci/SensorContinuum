package environment

import (
	"SensorContinuum/pkg/logger"
	"github.com/joho/godotenv"
	"os"
)

var RegionListUrl string
var RegionDetailUrl string
var RegionUrl string

func SetupEnvironment() error {

	err := godotenv.Load("internal/client/environment/.env")
	if err != nil {
		logger.Log.Warn("Error loading .env file, using environment variables instead: %v", err)
	} else {
		logger.Log.Info("Environment variables loaded from .env file")
	}

	var exists bool

	RegionListUrl, exists = os.LookupEnv("REGION_LIST_URL")
	if !exists {
		return os.ErrNotExist
	}
	logger.Log.Debug("Region list url: %s", RegionListUrl)

	RegionDetailUrl, exists = os.LookupEnv("REGION_DETAIL_URL")
	if !exists {
		return os.ErrNotExist
	}
	logger.Log.Debug("Region detail url: %s", RegionDetailUrl)

	return nil

}
