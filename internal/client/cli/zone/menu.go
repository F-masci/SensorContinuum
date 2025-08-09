package zone

import (
	"SensorContinuum/pkg/utils"
	"fmt"
	"strings"
)

const (
	green    = "\033[32m"
	yellow   = "\033[33m"
	cyanBold = "\033[1;36m"
	reset    = "\033[0m"
	sepHeavy = "═"
	sepLight = "─"
)

// ZoneMenu mostra il menù di gestione delle zone di una macrozona
func ZoneMenu(regionName, macrozoneName string) {
	for {
		line := strings.Repeat(sepHeavy, 60)
		fmt.Printf("\n%s\n%s🟦 Gestione Zone%s\n%s\n", line, cyanBold, reset, line)
		fmt.Printf("Regione: %s%s%s | Macrozona: %s%s%s\n", green, regionName, reset, green, macrozoneName, reset)
		fmt.Println(strings.Repeat(sepLight, 40))
		fmt.Printf("%s1%s) 📋 Lista zone\n", green, reset)
		fmt.Printf("%s2%s) 🔎 Cerca zona per nome\n", green, reset)
		fmt.Printf("%s0%s) ⬅️  Torna al menu macrozona\n", yellow, reset)
		fmt.Println(strings.Repeat(sepLight, 40))
		fmt.Print(yellow + "Seleziona un'opzione: " + reset)
		choice := utils.ReadInput()
		switch choice {
		case "1":
			listZones(regionName, macrozoneName)
		case "2":
			fmt.Print(yellow + "Nome della zona: " + reset)
			zoneName := utils.ReadInput()
			getZoneByName(regionName, macrozoneName, zoneName)
		case "0":
			return
		default:
			fmt.Println("❌ Opzione non valida.")
		}
	}
}
