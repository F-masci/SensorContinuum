package storage

import (
	"SensorContinuum/internal/edge-hub/environment"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/structure"
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

var RedisClient *redis.Client

func InitRedisConnection() {
	logger.Log.Debug("Initializing Redis connection with address: ", environment.RedisAddress, " and port: ", environment.RedisPort)
	RedisClient = redis.NewClient(&redis.Options{
		Addr: environment.RedisAddress + ":" + environment.RedisPort,
	})
}

// AddSensorHistory Salva un nuovo dato SensorData nella lista Redis del sensore.
func AddSensorHistory(ctx context.Context, data structure.SensorData) error {
	key := fmt.Sprintf("sensor:%s:history", data.SensorID)
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
func GetSensorHistory(ctx context.Context, sensorID string, n int) ([]structure.SensorData, error) {
	key := fmt.Sprintf("sensor:%s:history", sensorID)
	vals, err := RedisClient.LRange(ctx, key, 0, int64(n-1)).Result()
	if err != nil {
		return nil, err
	}
	readings := make([]structure.SensorData, 0, len(vals))
	for _, v := range vals {
		var d structure.SensorData
		if err := json.Unmarshal([]byte(v), &d); err == nil {
			readings = append(readings, d)
		}
	}
	return readings, nil
}

// GetSensorHistoryByMinute recupera le letture per un dato sensore che corrispondono al minuto specificato.
func GetSensorHistoryByMinute(ctx context.Context, sensorID string, minute time.Time) ([]structure.SensorData, error) {
	key := fmt.Sprintf("sensor:%s:history", sensorID)
	vals, err := RedisClient.LRange(ctx, key, 0, int64(environment.HistoryWindowSize-1)).Result()
	if err != nil {
		return nil, err
	}

	logger.Log.Debug("Retrieved ", len(vals), " readings for sensor ", sensorID, " at minute ", minute)

	readings := make([]structure.SensorData, 0, len(vals))
	for _, v := range vals {
		var d structure.SensorData
		if err := json.Unmarshal([]byte(v), &d); err == nil {
			t, err := time.Parse(time.RFC3339, d.Timestamp)
			if err != nil {
				continue
			}
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
	keys, err := RedisClient.Keys(ctx, "sensor:*:history").Result()
	if err != nil {
		return nil, err
	}
	sensorIDs := make([]string, 0, len(keys))
	for _, key := range keys {
		sensorID := strings.TrimPrefix(key, "sensor:")
		sensorID = strings.TrimSuffix(sensorID, ":history")
		logger.Log.Debug("Retrieving sensor ID from redis: ", sensorID)
		sensorIDs = append(sensorIDs, sensorID)
	}
	return sensorIDs, nil
}

func RemoveSensorHistory(ctx context.Context, sensorID string) error {
	key := fmt.Sprintf("sensor:%s:history", sensorID)
	_, err := RedisClient.Del(ctx, key).Result()
	if err != nil {
		logger.Log.Error("Error removing sensor history for sensor ", sensorID, ": ", err)
		return err
	}
	logger.Log.Info("Removed sensor history for sensor ", sensorID)
	return nil
}
