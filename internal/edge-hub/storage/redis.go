package storage

import (
	"SensorContinuum/internal/edge-hub/environment"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/structure"
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

var RedisClient *redis.Client

func InitRedisConnection() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
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

	readings := make([]structure.SensorData, 0, len(vals))
	for _, v := range vals {
		var d structure.SensorData
		if err := json.Unmarshal([]byte(v), &d); err == nil {
			t, err := time.Parse(time.RFC3339, d.Timestamp)
			if err != nil {
				continue
			}
			logger.Log.Debug("Checking reading timestamp: ", t, " against minute: ", minute)
			tUTC := t.UTC()
			minuteUTC := minute.UTC()
			if tUTC.Year() == minuteUTC.Year() && tUTC.Month() == minuteUTC.Month() && tUTC.Day() == minuteUTC.Day() &&
				tUTC.Hour() == minuteUTC.Hour() && tUTC.Minute() == minuteUTC.Minute() {
				readings = append(readings, d)
			}
		}
	}
	return readings, nil
}
