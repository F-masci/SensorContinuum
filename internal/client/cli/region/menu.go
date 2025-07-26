package region

import (
	"SensorContinuum/pkg/utils"
	"fmt"
)

// RegionMenu Sottomenù gestione zone stateless
func RegionMenu() {
	for {
		fmt.Println("\n--- Gestione Zone ---")
		fmt.Println("1) Lista zone")
		fmt.Println("2) Dettaglio zona")
		fmt.Println("3) Aggiungi zona")
		fmt.Println("4) Modifica zona")
		fmt.Println("5) Elimina zona")
		fmt.Println("0) Torna al menù principale")
		fmt.Print("Seleziona un'opzione: ")
		choice := utils.ReadInput()
		switch choice {
		case "1":
			listRegions()
		case "2":
			getRegionDetails()
		case "3":
			// addRegion()
		case "4":
			// updateRegion()
		case "5":
			// deleteRegion()
		case "0":
			return
		default:
			fmt.Println("Opzione non valida.")
		}
	}
}
