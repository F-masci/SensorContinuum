package building

import (
	"SensorContinuum/internal/client/comunication/api"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/utils"
	"fmt"
)

func listBuildings() {
	buildings, err := api.GetBuildings()
	if err != nil {
		logger.Log.Error("Error retrieving buildings: ", err)
		return
	}
	logger.Log.Debug("Found ", len(buildings), " buildings")
	fmt.Println("Edifici disponibili:")
	for _, b := range buildings {
		fmt.Printf("- %s (ID: %d, Regione: %d, Lat: %.6f, Lon: %.6f)\n", b.Name, b.Id, b.RegionId, b.Lat, b.Lon)
	}
}

func getBuildingByID() {
	fmt.Print("ID edificio: ")
	id := utils.ReadInput()
	building, err := api.GetBuildingByID(id)
	if err != nil {
		logger.Log.Error("Error retrieving building: ", err)
		return
	}
	fmt.Printf("Dettagli edificio:\nNome: %s\nID: %d\nRegione: %d\nLat: %.6f\nLon: %.6f\n", building.Name, building.Id, building.RegionId, building.Lat, building.Lon)
}

func getBuildingByName() {
	fmt.Print("Nome edificio: ")
	name := utils.ReadInput()
	building, err := api.GetBuildingByName(name)
	if err != nil {
		logger.Log.Error("Error retrieving building: ", err)
		return
	}
	fmt.Printf("Dettagli edificio:\nNome: %s\nID: %d\nRegione: %d\nLat: %.6f\nLon: %.6f\n", building.Name, building.Id, building.RegionId, building.Lat, building.Lon)
}

func getBuildingsByRegion() {
	fmt.Print("ID regione: ")
	regionID := utils.ReadInput()
	buildings, err := api.GetBuildingsByRegion(regionID)
	if err != nil {
		logger.Log.Error("Error retrieving buildings by region: ", err)
		return
	}
	fmt.Printf("Edifici nella regione %s:\n", regionID)
	if len(buildings) == 0 {
		fmt.Println("Nessun edificio trovato.")
		return
	}
	for _, b := range buildings {
		fmt.Printf("- %s (ID: %d, Lat: %.6f, Lon: %.6f)\n", b.Name, b.Id, b.Lat, b.Lon)
	}
}
