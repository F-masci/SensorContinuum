package types

import (
	"encoding/json"
	"errors"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/segmentio/kafka-go"
)

type SensorData struct {
	EdgeMacrozone string  `json:"macrozone"`
	EdgeZone      string  `json:"zone"`
	SensorID      string  `json:"sensor_id"`
	Timestamp     int64   `json:"timestamp"`
	Type          string  `json:"type"`
	Data          float64 `json:"data"`
}

func CreateSensorDataFromMQTT(msg MQTT.Message) (SensorData, error) {
	var data SensorData
	err := json.Unmarshal(msg.Payload(), &data)
	return data, err
}

func CreateSensorDataFromKafka(msg kafka.Message) (SensorData, error) {
	var data SensorData
	err := json.Unmarshal(msg.Value, &data)
	return data, err
}

type SensorDataBatch struct {
	SensorData []SensorData `json:"sensor_data"`
	counter    int
	ticker     *time.Ticker
	SaveData   func(*SensorDataBatch) error
	maxCount   int
	timeout    time.Duration
	stopChan   chan struct{}
}

// NewSensorDataBatch crea un nuovo batch di dati sensori con salvataggio e ticker
func NewSensorDataBatch(maxCount int, timeout time.Duration, saveDataFunction func(*SensorDataBatch) error) (*SensorDataBatch, error) {

	// Controlla che i parametri siano validi
	if saveDataFunction == nil {
		return nil, errors.New("saveDataFunction cannot be nil")
	}
	if maxCount <= 0 {
		return nil, errors.New("maxCount must be greater than 0")
	}
	if timeout <= 0 {
		return nil, errors.New("timeout must be greater than 0")
	}

	// Inizializza il batch
	batch := &SensorDataBatch{
		SensorData: make([]SensorData, 0),
		counter:    0,
		SaveData:   saveDataFunction,
		maxCount:   maxCount,
		timeout:    timeout,
		stopChan:   make(chan struct{}),
	}

	// Avvia il ticker per il salvataggio periodico
	batch.startTicker()
	return batch, nil
}

// startTicker avvia un ticker che salva i dati ogni timeout se ci sono dati nel batch
func (sdb *SensorDataBatch) startTicker() {

	// Se il ticker esiste già, fermalo
	if sdb.ticker != nil {
		sdb.ticker.Stop()
	}

	// Crea un nuovo ticker
	sdb.ticker = time.NewTicker(sdb.timeout)
	go func() {
		for {
			select {
			// Al timeout del ticker, salva i dati se ce ne sono
			case <-sdb.ticker.C:
				if sdb.SaveData != nil {
					_ = sdb.SaveData(sdb)
				}
				// Dopo il salvataggio, resetta il batch e riavvia il ticker
				sdb.Clear()
				sdb.restartTicker()
			case <-sdb.stopChan:
				return
			}
		}
	}()
}

// restartTicker ferma il ticker corrente e ne crea uno nuovo
func (sdb *SensorDataBatch) restartTicker() {
	sdb.ticker.Stop()
	sdb.ticker = time.NewTicker(sdb.timeout)
}

// AddSensorData aggiunge un nuovo dato sensore al batch e salva se il batch è pieno
func (sdb *SensorDataBatch) AddSensorData(data SensorData) {
	if sdb.counter >= sdb.maxCount && sdb.SaveData != nil {
		_ = sdb.SaveData(sdb)
		sdb.Clear()
		sdb.restartTicker()
	}
	sdb.SensorData = append(sdb.SensorData, data)
	sdb.counter++
}

// Count restituisce il numero di dati sensori nel batch
func (sdb *SensorDataBatch) Count() int {
	return sdb.counter
}

// Clear resetta il batch di dati sensori
func (sdb *SensorDataBatch) Clear() {
	sdb.SensorData = make([]SensorData, 0)
	sdb.counter = 0
}

// Stop ferma il ticker e chiude il canale di stop
func (sdb *SensorDataBatch) Stop() {
	close(sdb.stopChan)
	if sdb.ticker != nil {
		sdb.ticker.Stop()
	}
}

// AggregatedStats contiene i dati statistici calcolati ogni tot minuti dal Proximity-Fog-Hub
// e inviati tramite kafka all' intermediate-fog-hub per essere memorizzati nel database centrale
type AggregatedStats struct {
	ID            string  `json:"id,omitempty"`
	Timestamp     int64   `json:"timestamp"`
	Region        string  `json:"region,omitempty"`
	Macrozone     string  `json:"macrozone,omitempty"`
	Zone          string  `json:"zone,omitempty"`
	Type          string  `json:"type"`
	Min           float64 `json:"min"`
	Max           float64 `json:"max"`
	Avg           float64 `json:"avg"`
	Sum           float64 `json:"sum,omitempty"`
	Count         int     `json:"count,omitempty"`
	WeightedAvg   float64 `json:"weighted_avg,omitempty"`
	WeightedSum   float64 `json:"weighted_sum,omitempty"`
	WeightedCount float64 `json:"weighted_count,omitempty"`
}

// CreateAggregatedStatsFromKafka deserializza un messaggio Kafka in AggregatedStats
func CreateAggregatedStatsFromKafka(msg kafka.Message) (AggregatedStats, error) {
	var stats AggregatedStats
	err := json.Unmarshal(msg.Value, &stats)
	return stats, err
}

type AggregatedStatsBatch struct {
	Stats     []AggregatedStats `json:"aggregated_stats"`
	counter   int
	ticker    *time.Ticker
	SaveStats func(*AggregatedStatsBatch) error
	maxCount  int
	timeout   time.Duration
	stopChan  chan struct{}
}

// NewAggregatedStatsBatch crea un nuovo batch di statistiche aggregate con salvataggio e ticker
func NewAggregatedStatsBatch(maxCount int, timeout time.Duration, saveStatsFunction func(*AggregatedStatsBatch) error) (*AggregatedStatsBatch, error) {
	if saveStatsFunction == nil {
		return nil, errors.New("saveStatsFunction cannot be nil")
	}
	if maxCount <= 0 {
		return nil, errors.New("maxCount must be greater than 0")
	}
	if timeout <= 0 {
		return nil, errors.New("timeout must be greater than 0")
	}

	batch := &AggregatedStatsBatch{
		Stats:     make([]AggregatedStats, 0),
		counter:   0,
		SaveStats: saveStatsFunction,
		maxCount:  maxCount,
		timeout:   timeout,
		stopChan:  make(chan struct{}),
	}
	batch.startTicker()
	return batch, nil
}

// startTicker avvia un ticker che salva le statistiche ogni timeout se ci sono dati nel batch
func (asb *AggregatedStatsBatch) startTicker() {
	if asb.ticker != nil {
		asb.ticker.Stop()
	}
	asb.ticker = time.NewTicker(asb.timeout)
	go func() {
		for {
			select {
			case <-asb.ticker.C:
				if asb.SaveStats != nil {
					_ = asb.SaveStats(asb)
				}
				asb.Clear()
				asb.restartTicker()
			case <-asb.stopChan:
				return
			}
		}
	}()
}

// restartTicker ferma il ticker corrente e ne crea uno nuovo
func (asb *AggregatedStatsBatch) restartTicker() {
	asb.ticker.Stop()
	asb.ticker = time.NewTicker(asb.timeout)
}

// AddAggregatedStats aggiunge una nuova statistica e salva se il batch è pieno
func (asb *AggregatedStatsBatch) AddAggregatedStats(stats AggregatedStats) {
	if asb.counter >= asb.maxCount && asb.SaveStats != nil {
		_ = asb.SaveStats(asb)
		asb.Clear()
		asb.restartTicker()
	}
	asb.Stats = append(asb.Stats, stats)
	asb.counter++
}

// Count restituisce il numero di statistiche nel batch
func (asb *AggregatedStatsBatch) Count() int {
	return asb.counter
}

// Clear resetta il batch di statistiche aggregate
func (asb *AggregatedStatsBatch) Clear() {
	asb.Stats = make([]AggregatedStats, 0)
	asb.counter = 0
}

// Stop ferma il ticker e chiude il canale di stop
func (asb *AggregatedStatsBatch) Stop() {
	close(asb.stopChan)
	if asb.ticker != nil {
		asb.ticker.Stop()
	}
}
