package storage

import (
	"SensorContinuum/internal/edge-hub/environment"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/types"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/redis/go-redis/v9"
)

const sensorMetadataKey = "sensor:%s:metadata"
const sensorHistoryKey = "sensor:%s:history"

const leaderKey = "edge-hub-leader"
const leaderTTL = 70 * time.Second // Attende che il leader rinnovi il suo status

var RedisClient *redis.Client

func InitRedisConnection() {
	logger.Log.Debug("Initializing Redis connection with address: ", environment.RedisAddress, " and port: ", environment.RedisPort)
	RedisClient = redis.NewClient(&redis.Options{
		Addr: environment.RedisAddress + ":" + environment.RedisPort,
	})
	logger.Log.Debug("Redis client initialized")
	if err := RedisClient.Ping(context.Background()).Err(); err != nil {
		logger.Log.Error("Failed to connect to Redis: ", err)
		panic(fmt.Sprintf("Failed to connect to Redis at %s:%s", environment.RedisAddress, environment.RedisPort))
	}
}

// TryOrRenewLeader prova ad acquisire il lock di leader election
func TryOrRenewLeader(ctx context.Context, instanceID string) (bool, error) {
	// Prova ad acquisire il lock
	ok, err := RedisClient.SetNX(ctx, leaderKey, instanceID, leaderTTL).Result()
	if err != nil {
		return false, err
	}
	if ok {
		// Sei diventato leader ora
		return true, nil
	}
	// Controlla chi è il leader attuale
	val, err := RedisClient.Get(ctx, leaderKey).Result()
	if err != nil {
		return false, err
	}
	if val == instanceID {
		// Sei già leader, rinnova il TTL
		_, err = RedisClient.Expire(ctx, leaderKey, leaderTTL).Result()
		return true, err
	}
	// Non sei il leader
	return false, nil
}

// AddSensor Aggiunge un nuovo sensore alla cache Redis.
// Controlla se il sensore esiste già prima di aggiungerlo
func AddSensor(ctx context.Context, sensor types.Sensor) (bool, error) {
	key := fmt.Sprintf(sensorMetadataKey, sensor.Id)
	b, err := json.Marshal(sensor)
	if err != nil {
		return false, err
	}
	// SETNX: aggiunge solo se la chiave non esiste
	added, err := RedisClient.SetNX(ctx, key, b, 0).Result()
	if err != nil {
		return false, err
	}
	if !added {
		return true, nil // già esiste
	}
	logger.Log.Info("Sensor added to Redis: ", sensor.Id)
	return false, nil
}

// AddSensorHistory Salva un nuovo dato SensorData nella lista Redis del sensore.
func AddSensorHistory(ctx context.Context, data types.SensorData) error {
	key := fmt.Sprintf(sensorHistoryKey, data.SensorID)
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	pipe := RedisClient.Pipeline()
	pipe.LPush(ctx, key, b)
	pipe.LTrim(ctx, key, 0, int64(environment.HistoryWindowSize-1))
	_, err = pipe.Exec(ctx)
	return err
}

// GetSensorHistory Recupera le ultime n letture per un dato sensore.
func GetSensorHistory(ctx context.Context, sensorID string, n int) ([]types.SensorData, error) {
	key := fmt.Sprintf(sensorHistoryKey, sensorID)
	vals, err := RedisClient.LRange(ctx, key, 0, int64(n-1)).Result()
	if err != nil {
		return nil, err
	}
	readings := make([]types.SensorData, 0, len(vals))
	for _, v := range vals {
		var d types.SensorData
		if err := json.Unmarshal([]byte(v), &d); err == nil {
			readings = append(readings, d)
		}
	}
	return readings, nil
}

// GetSensorHistoryByMinute recupera le letture per un dato sensore che corrispondono al minuto specificato.
func GetSensorHistoryByMinute(ctx context.Context, sensorID string, minute time.Time) ([]types.SensorData, error) {
	key := fmt.Sprintf(sensorHistoryKey, sensorID)
	vals, err := RedisClient.LRange(ctx, key, 0, int64(environment.HistoryWindowSize-1)).Result()
	if err != nil {
		return nil, err
	}

	logger.Log.Debug("Retrieved ", len(vals), " readings for sensor ", sensorID, " at minute ", minute)

	readings := make([]types.SensorData, 0, len(vals))
	for _, v := range vals {
		var d types.SensorData
		if err := json.Unmarshal([]byte(v), &d); err == nil {
			t := time.Unix(d.Timestamp, 0).UTC()
			logger.Log.Debug("Checking reading timestamp: ", t, " against minute: ", minute)
			if t.Year() == minute.Year() && t.Month() == minute.Month() && t.Day() == minute.Day() &&
				t.Hour() == minute.Hour() && t.Minute() == minute.Minute() {
				readings = append(readings, d)
			}
		} else {
			logger.Log.Error("Error unmarshalling sensor data: ", err)
			continue
		}
	}
	return readings, nil
}

func GetAllSensorIDs(ctx context.Context) ([]string, error) {
	set := mapset.NewSet[string]()
	keysMeta, err := RedisClient.Keys(ctx, "sensor:*:metadata").Result()
	if err != nil {
		return nil, err
	}
	set.Append(keysMeta...)
	keysHist, err := RedisClient.Keys(ctx, "sensor:*:history").Result()
	if err != nil {
		return nil, err
	}
	set.Append(keysHist...)

	keys := set.ToSlice()
	sensorIDs := make([]string, 0, len(keys))
	for _, key := range keys {
		sensorID := strings.TrimPrefix(key, "sensor:")
		sensorID = strings.TrimSuffix(sensorID, ":metadata")
		logger.Log.Debug("Retrieving sensor ID from redis: ", sensorID)
		sensorIDs = append(sensorIDs, sensorID)
	}
	return sensorIDs, nil
}

func RemoveSensorHistory(ctx context.Context, sensorID string) error {
	key := fmt.Sprintf(sensorHistoryKey, sensorID)
	_, err := RedisClient.Del(ctx, key).Result()
	if err != nil {
		logger.Log.Error("Error removing sensor history for sensor ", sensorID, ": ", err)
		return err
	}
	logger.Log.Info("Removed sensor history for sensor ", sensorID)
	return nil
}

func RemoveSensor(ctx context.Context, sensorID string) error {
	key := fmt.Sprintf(sensorMetadataKey, sensorID)
	_, err := RedisClient.Del(ctx, key).Result()
	if err != nil {
		logger.Log.Error("Error removing sensor ", sensorID, ": ", err)
		return err
	}
	logger.Log.Info("Removed sensor ", sensorID)
	return nil
}
