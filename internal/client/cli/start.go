package cli

import (
	"SensorContinuum/internal/client/cli/macrozone"
	"SensorContinuum/internal/client/cli/region"
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

// MainMenu Menù principale
func MainMenu() {
	for {
		line := strings.Repeat(sepHeavy, 60)
		fmt.Printf("\n%s\n%s🏠 Menù Principale%s\n%s\n", line, cyanBold, reset, line)
		fmt.Printf("%s1%s) 🌍 Lista regioni disponibili\n", green, reset)
		fmt.Printf("%s2%s) 🔎 Ricerca regione per nome\n", green, reset)
		fmt.Printf("%s3%s) 🏢 Gestione regione\n", green, reset)
		fmt.Printf("%s0%s) 🚪 Esci\n", yellow, reset)
		fmt.Println(strings.Repeat(sepLight, 40))
		fmt.Print(yellow + "Seleziona un'opzione: " + reset)
		choice := utils.ReadInput()
		switch choice {
		case "1":
			region.ListRegions()
		case "2":
			region.GetRegionDetailsByName()
		case "3":
			fmt.Print(yellow + "Nome della regione: " + reset)
			regionName := utils.ReadInput()
			macrozone.MacrozoneMenu(regionName)
		case "0":
			fmt.Println(green + "Uscita..." + reset)
			return
		default:
			fmt.Println("❌ Opzione non valida.")
		}
	}
}
