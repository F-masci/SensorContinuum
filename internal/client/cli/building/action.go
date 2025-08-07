package building

import (
	"SensorContinuum/internal/client/comunication/api"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/utils"
	"fmt"
)

func listBuildings(regionName string) {
	buildings, err := api.GetBuildings(regionName)
	if err != nil {
		logger.Log.Error("Error retrieving buildings: ", err)
		return
	}
	logger.Log.Debug("Found ", len(buildings), " buildings")
	fmt.Println("Edifici disponibili:")
	for _, b := range buildings {
		fmt.Printf("- %s (ID: %d, Lat: %.6f, Lon: %.6f)\n", b.Name, b.Id, b.Lat, b.Lon)
	}
}

func getBuildingByID(regionName string) {
	fmt.Print("ID edificio: ")
	id := utils.ReadInput()
	building, err := api.GetBuildingByID(regionName, id)
	if err != nil {
		logger.Log.Error("Error retrieving building: ", err)
		return
	}
	fmt.Printf(
		"Dettagli edificio:\nNome: %s\nID: %d\nLat: %.6f\nLon: %.6f\nRegistrato il: %s\nUltima comunicazione: %s\n",
		building.Name,
		building.Id,
		building.Lat,
		building.Lon,
		building.RegistrationTime.Format("2006-01-02 15:04:05"),
		building.LastComunication.Format("2006-01-02 15:04:05"),
	)
}

func getBuildingByName(regionName string) {
	fmt.Print("Nome edificio: ")
	name := utils.ReadInput()
	building, err := api.GetBuildingByName(regionName, name)
	if err != nil {
		logger.Log.Error("Error retrieving building: ", err)
		return
	}
	fmt.Printf(
		"Dettagli edificio:\nNome: %s\nID: %d\nLat: %.6f\nLon: %.6f\nRegistrato il: %s\nUltima comunicazione: %s\n",
		building.Name,
		building.Id,
		building.Lat,
		building.Lon,
		building.RegistrationTime.Format("2006-01-02 15:04:05"),
		building.LastComunication.Format("2006-01-02 15:04:05"),
	)
}
