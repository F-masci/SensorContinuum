package cli

import (
	"SensorContinuum/internal/client/cli/building"
	"SensorContinuum/internal/client/cli/region"
	"SensorContinuum/pkg/utils"
	"fmt"
)

// Menù principale stateless
func MainMenu() {
	for {
		fmt.Println("\n--- Menù Principale ---")
		fmt.Println("1) Gestione zone")
		fmt.Println("2) Gestione edifici")
		fmt.Println("0) Esci")
		fmt.Print("Seleziona un'opzione: ")
		choice := utils.ReadInput()
		switch choice {
		case "1":
			region.RegionMenu()
		case "2":
			building.BuildingMenu()
		case "0":
			fmt.Println("Uscita...")
			return
		default:
			fmt.Println("Opzione non valida.")
		}
	}
}
