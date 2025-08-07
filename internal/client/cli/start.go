package cli

import (
	"SensorContinuum/internal/client/cli/building"
	"SensorContinuum/internal/client/cli/region"
	"SensorContinuum/pkg/utils"
	"fmt"
)

// MainMenu Menù principale
func MainMenu() {
	for {
		fmt.Println("\n--- Menù Principale ---")
		fmt.Println("1) Lista zone disponibili")
		fmt.Println("2) Ricerca zona per ID")
		fmt.Println("3) Ricerca zona per nome")
		fmt.Println("4) Gestione Edifici")
		fmt.Println("0) Esci")
		fmt.Print("Seleziona un'opzione: ")
		choice := utils.ReadInput()
		switch choice {
		case "1":
			region.ListRegions()
		case "2":
			region.GetRegionDetailsByID()
		case "3":
			region.GetRegionDetailsByName()
		case "4":
			fmt.Print("Nome della regione: ")
			regionName := utils.ReadInput()
			// Passa il nome della regione al menù degli edifici
			building.BuildingMenu(regionName)
		case "0":
			fmt.Println("Uscita...")
			return
		default:
			fmt.Println("Opzione non valida.")
		}
	}
}
