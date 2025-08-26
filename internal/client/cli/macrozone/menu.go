package macrozone

import (
	"SensorContinuum/internal/client/cli/zone"
	"SensorContinuum/pkg/utils"
	"fmt"
	"strings"
)

const (
	red      = "\033[31m"
	green    = "\033[32m"
	yellow   = "\033[33m"
	cyanBold = "\033[1;36m"
	reset    = "\033[0m"
	sepHeavy = "═"
	sepLight = "─"
)

func MacrozoneMenu(regionName string) {
	for {
		line := strings.Repeat(sepHeavy, 60)
		fmt.Printf("\n%s\n%s🏢 Gestione Regione (%s)%s\n%s\n", line, cyanBold, regionName, reset, line)
		fmt.Printf("%s1%s) 📋 Lista macrozone\n", green, reset)
		fmt.Printf("%s2%s) 🔎 Cerca macrozona per nome\n", green, reset)
		fmt.Printf("%s3%s) 🏢 Gestione macrozona\n", green, reset)
		fmt.Printf("%s0%s) ⬅️  Torna al menu principale\n", yellow, reset)
		fmt.Println(strings.Repeat(sepLight, 40))
		fmt.Print(yellow + "Seleziona un'opzione: " + reset)
		choice := utils.ReadInput()
		switch choice {
		case "1":
			listMacrozones(regionName)
		case "2":
			getMacrozoneByName(regionName)
		case "3":
			fmt.Print(yellow + "Nome della macrozona: " + reset)
			macrozoneName := utils.ReadInput()
			zone.ZoneMenu(regionName, macrozoneName)
		case "0":
			return
		default:
			fmt.Println("❌ Opzione non valida.")
		}
	}
}
