package sensor_agent

var sensorID string

func SetSensorID(id string) {
	sensorID = id
}

func GetSensorID() string {
	return sensorID
}
