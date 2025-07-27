package region

import (
	"SensorContinuum/internal/client/comunication/api"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/utils"
	"fmt"
)

func listRegions() {
	regions, err := api.GetRegions()
	if err != nil {
		logger.Log.Error("Error retrieving regions: ", err)
		return
	}
	logger.Log.Debug("Founded ", len(regions), " regions")
	fmt.Println("Zone disponibili:")
	for _, r := range regions {
		fmt.Printf("- %s (ID: %d, Edifici: %d)\n", r.Name, r.Id, r.BuildingCount)
	}
}

func getRegionDetailsByID() {
	fmt.Print("ID della zona: ")
	id := utils.ReadInput()
	region, err := api.GetRegionById(id)
	if err != nil {
		logger.Log.Error("Error retrieving region: ", err)
		return
	}
	logger.Log.Debug("Found region detail: ", region)
	fmt.Printf("Dettagli zona:\nNome: %s\nID: %d\nNumero edifici: %d\n", region.Name, region.Id, region.BuildingCount)
	if len(region.Buildings) > 0 {
		fmt.Println("Edifici:")
		for _, b := range region.Buildings {
			fmt.Printf("- %s (ID: %d, Lat: %.6f, Lon: %.6f)\n", b.Name, b.Id, b.Lat, b.Lon)
		}
	} else {
		fmt.Println("Nessun edificio associato.")
	}
}

func getRegionDetailsByName() {
	fmt.Print("Nome della zona: ")
	name := utils.ReadInput()
	region, err := api.GetRegionByName(name)
	if err != nil {
		logger.Log.Error("Error retrieving region: ", err)
		return
	}
	logger.Log.Debug("Found region detail: ", region)
	fmt.Printf("Dettagli zona:\nNome: %s\nID: %d\nNumero edifici: %d\n", region.Name, region.Id, region.BuildingCount)
	if len(region.Buildings) > 0 {
		fmt.Println("Edifici:")
		for _, b := range region.Buildings {
			fmt.Printf("- %s (ID: %d, Lat: %.6f, Lon: %.6f)\n", b.Name, b.Id, b.Lat, b.Lon)
		}
	} else {
		fmt.Println("Nessun edificio associato.")
	}
}
