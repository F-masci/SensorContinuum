package region

import (
	"SensorContinuum/pkg/utils"
	"fmt"
)

// RegionMenu mostra il menù di gestione delle zone e gestisce le interazioni dell'utente in modo stateless
func RegionMenu() {
	for {
		fmt.Println("\n--- Gestione Zone ---")
		fmt.Println("1) Lista zone")
		fmt.Println("2) Ricerca zona per ID")
		fmt.Println("3) Ricerca zona per nome")
		fmt.Println("0) Torna al menù principale")
		fmt.Print("Seleziona un'opzione: ")
		choice := utils.ReadInput()
		switch choice {
		case "1":
			listRegions()
		case "2":
			getRegionDetailsByID()
		case "3":
			getRegionDetailsByName()
		case "0":
			return
		default:
			fmt.Println("Opzione non valida.")
		}
	}
}
