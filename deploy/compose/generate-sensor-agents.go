package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: generate-sensor-agents <numero agenti>")
		os.Exit(1)
	}
	numAgents, err := strconv.Atoi(os.Args[1])
	if err != nil || numAgents < 1 {
		fmt.Println("Numero agenti non valido.")
		os.Exit(1)
	}

	types := []string{"temperature", "humidity", "pressure"}
	locations := []string{"indoor", "outdoor"}
	references := []string{
		"bmp280", "dht22", "ds18b20",
	}

	templatePath := "./sensor-agent.template.yml"
	targetPath := "./sensor-agent.generated_" + strconv.Itoa(numAgents) + ".yml"

	data, err := os.ReadFile(templatePath)
	if err != nil {
		panic(err)
	}
	template := string(data)
	template = strings.ReplaceAll(template, "services:\n", "")
	template = strings.ReplaceAll(template, "services:\r\n", "")

	rand.Seed(time.Now().UnixNano())

	var result strings.Builder
	result.WriteString("services:\n\n")

	for i := 1; i <= numAgents; i++ {
		id := fmt.Sprintf("%02d", i)
		sensorReference := references[rand.Intn(len(references))]

		var sensorType string
		switch sensorReference {
		case "bmp280":
			sensorType = types[rand.Intn(len(types))]
		case "dht22":
			sensorType = types[rand.Intn(2)] // temperature o humidity
		case "ds18b20":
			sensorType = "temperature"
		}

		sensorLocation := locations[rand.Intn(len(locations))]

		agent := strings.ReplaceAll(template, "$#", id)
		agent = strings.ReplaceAll(agent, "$TYPE", sensorType)
		agent = strings.ReplaceAll(agent, "$LOCATION", sensorLocation)
		agent = strings.ReplaceAll(agent, "$REFERENCE", sensorReference)
		result.WriteString(agent)
		result.WriteString("\n\n")
	}

	err = os.WriteFile(targetPath, []byte(result.String()), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Println("File generato:", targetPath)
}
