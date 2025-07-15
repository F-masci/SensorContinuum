package main

import (
	"SensorContinuum/internal/sensor-agent"
	"SensorContinuum/internal/sensor-agent/comunication"
	"SensorContinuum/pkg/logger"
	"net/rpc"
	"os"
)

// TODO: Configurare il contesto del logger dinamicamente
func getContext(sensorID string) logger.Context {
	return logger.Context{
		"service":  "sensor-agent",
		"sensorID": sensorID,
	}
}

type EmptyArgs struct{}
type IDReply struct {
	ID string
}

func main() {

	address, exists := os.LookupEnv("EDGE_HUB")
	if !exists {
		address = "localhost"
	}

	// Ottiene il proprio ID
	client, err := rpc.Dial("tcp", address+":1234")
	if err != nil {
		panic(err)
	}
	defer client.Close()

	var reply IDReply
	err = client.Call("SensorIDService.GetNextID", &EmptyArgs{}, &reply)
	if err != nil {
		panic(err)
	}
	sensor_agent.SetSensorID(reply.ID)

	// Inizializza il logger con il contesto
	logger.CreateLogger(getContext(reply.ID))
	logger.Log.Info("Starting Sensor Agent...")

	sensorChannel := make(chan float64, 100)
	go sensor_agent.SimulateForever(sensorChannel)

	for data := range sensorChannel {
		comunication.Publish(data)
	}

}
