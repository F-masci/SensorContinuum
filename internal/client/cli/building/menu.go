package building

import (
	"SensorContinuum/pkg/utils"
	"fmt"
)

// BuildingMenu mostra il menù di gestione degli edifici e gestisce le interazioni dell'utente
func BuildingMenu() {
	for {
		fmt.Println("\n--- Gestione Edifici ---")
		fmt.Println("1) Lista edifici")
		fmt.Println("2) Cerca edificio per ID")
		fmt.Println("3) Cerca edificio per nome")
		fmt.Println("4) Lista edifici per regione")
		fmt.Println("0) Torna al menu principale")
		fmt.Print("Seleziona un'opzione: ")
		choice := utils.ReadInput()
		switch choice {
		case "1":
			listBuildings()
		case "2":
			getBuildingByID()
		case "3":
			getBuildingByName()
		case "4":
			getBuildingsByRegion()
		case "0":
			return
		default:
			fmt.Println("Opzione non valida.")
		}
	}
}
