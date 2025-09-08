package environment

import (
	"SensorContinuum/pkg/logger"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
)

var CloudDatabaseUser string
var CloudDatabasePassword string
var CloudDatabaseHost string
var CloudDatabasePort string
var CloudDatabaseName string

var RegionMetadataDatabaseUser string
var RegionMetadataDatabasePassword string
var RegionMetadataDatabaseHostTemplate string
var RegionMetadataDatabasePort string
var RegionMetadataDatabaseName string

var RegionMeasurementDatabaseUser string
var RegionMeasurementDatabasePassword string
var RegionMeasurementDatabaseHostTemplate string
var RegionMeasurementDatabasePort string
var RegionMeasurementDatabaseName string

const (
	DefCloudDatabaseUser     = "sc_master"
	DefCloudDatabasePassword = "adminpass"
	DefCloudDatabaseHost     = "cloud.metadata-db.sensor-continuum.it"
	DefCloudDatabasePort     = "5433"
	DefCloudDatabaseName     = "sensorcontinuum"

	DefRegionMetadataDatabaseUser         = "admin"
	DefRegionMetadataDatabasePassword     = "adminpass"
	DefRegionMetadataDatabaseHostTemplate = "%s.metadata-db.sensor-continuum.it"
	DefRegionMetadataDatabasePort         = "5434"
	DefRegionMetadataDatabaseName         = "sensorcontinuum"

	DefRegionMeasurementDatabaseUser         = "admin"
	DefRegionMeasurementDatabasePassword     = "adminpass"
	DefRegionMeasurementDatabaseHostTemplate = "%s.measurement-db.sensor-continuum.it"
	DefRegionMeasurementDatabasePort         = "5432"
	DefRegionMeasurementDatabaseName         = "sensorcontinuum"
)

const (
	// AggregatedDataCutOff definisce il tempo massimo di validità dei dati aggregati.
	// Se i dati sono più vecchi di questo valore, non vengono considerati validi
	AggregatedDataCutOff time.Duration = 2 * time.Hour

	// YearlyVariationMinimum definisce l'offset temporale minimo per il calcolo della variazione annuale.
	// Se la data è più recente di questo valore rispetto alla data attuale, non viene considerata valida
	YearlyVariationMinimum time.Duration = 48 * time.Hour
)

func SetupEnvironment() error {

	err := godotenv.Load(filepath.Join("internal", "client", "environment", ".env"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return errors.New("Error loading .env file: " + err.Error())
	}

	var exists bool

	/* --- Cloud Database --- */

	CloudDatabaseUser, exists = os.LookupEnv("CLOUD_DATABASE_USER")
	if !exists {
		CloudDatabaseUser = DefCloudDatabaseUser
	}
	logger.Log.Debug("Cloud database user: ", CloudDatabaseUser)

	CloudDatabasePassword, exists = os.LookupEnv("CLOUD_DATABASE_PASSWORD")
	if !exists {
		CloudDatabasePassword = DefCloudDatabasePassword
	}

	CloudDatabaseHost, exists = os.LookupEnv("CLOUD_DATABASE_HOST")
	if !exists {
		CloudDatabaseHost = DefCloudDatabaseHost
	}

	CloudDatabasePort, exists = os.LookupEnv("CLOUD_DATABASE_PORT")
	if !exists {
		CloudDatabasePort = DefCloudDatabasePort
	}

	CloudDatabaseName, exists = os.LookupEnv("CLOUD_DATABASE_NAME")
	if !exists {
		CloudDatabaseName = DefCloudDatabaseName
	}

	/* --- Region Metadata Database --- */

	RegionMetadataDatabaseUser, exists = os.LookupEnv("REGION_METADATA_DATABASE_USER")
	if !exists {
		RegionMetadataDatabaseUser = DefRegionMetadataDatabaseUser
	}

	RegionMetadataDatabasePassword, exists = os.LookupEnv("REGION_METADATA_DATABASE_PASSWORD")
	if !exists {
		RegionMetadataDatabasePassword = DefRegionMetadataDatabasePassword
	}

	RegionMetadataDatabaseHostTemplate, exists = os.LookupEnv("REGION_METADATA_DATABASE_HOST_TEMPLATE")
	if !exists {
		RegionMetadataDatabaseHostTemplate = DefRegionMetadataDatabaseHostTemplate
	}

	RegionMetadataDatabasePort, exists = os.LookupEnv("REGION_METADATA_DATABASE_PORT")
	if !exists {
		RegionMetadataDatabasePort = DefRegionMetadataDatabasePort
	}

	RegionMetadataDatabaseName, exists = os.LookupEnv("REGION_METADATA_DATABASE_NAME")
	if !exists {
		RegionMetadataDatabaseName = DefRegionMetadataDatabaseName
	}

	/* --- Region Measurement Database --- */

	RegionMeasurementDatabaseUser, exists = os.LookupEnv("REGION_MEASUREMENT_DATABASE_USER")
	if !exists {
		RegionMeasurementDatabaseUser = DefRegionMeasurementDatabaseUser
	}

	RegionMeasurementDatabasePassword, exists = os.LookupEnv("REGION_MEASUREMENT_DATABASE_PASSWORD")
	if !exists {
		RegionMeasurementDatabasePassword = DefRegionMeasurementDatabasePassword
	}

	RegionMeasurementDatabaseHostTemplate, exists = os.LookupEnv("REGION_MEASUREMENT_DATABASE_HOST_TEMPLATE")
	if !exists {
		RegionMeasurementDatabaseHostTemplate = DefRegionMeasurementDatabaseHostTemplate
	}

	RegionMeasurementDatabasePort, exists = os.LookupEnv("REGION_MEASUREMENT_DATABASE_PORT")
	if !exists {
		RegionMeasurementDatabasePort = DefRegionMeasurementDatabasePort
	}

	RegionMeasurementDatabaseName, exists = os.LookupEnv("REGION_MEASUREMENT_DATABASE_NAME")
	if !exists {
		RegionMeasurementDatabaseName = DefRegionMeasurementDatabaseName
	}

	return nil

}
