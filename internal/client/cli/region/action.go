package region

import (
	"SensorContinuum/internal/client/comunication/api"
	"SensorContinuum/internal/client/environment"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/utils"
	"fmt"
	"strings"
	"time"
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

func ListRegions() {
	regions, err := api.GetRegions()
	if err != nil {
		logger.Log.Error("Errore nel recupero delle regioni: ", err)
		return
	}
	logger.Log.Debug("🔎 Trovate ", len(regions), " regioni")

	line := strings.Repeat(sepHeavy, 70)
	fmt.Printf("%s\n%s🌍 Regioni disponibili (%d)%s\n%s\n", line, cyanBold, len(regions), reset, line)
	if len(regions) == 0 {
		fmt.Println("  🚫 Nessuna regione trovata.")
	} else {
		// Intestazioni senza colori ANSI per allineamento corretto
		fmt.Printf("%-20s │ %-12s\n", "Nome", "Macrozone")
		fmt.Println(strings.Repeat(sepLight, 36))
		for _, r := range regions {
			fmt.Printf("%s%-20s%s │ %s%12d%s\n", green, r.Name, reset, yellow, r.MacrozoneCount, reset)
		}
	}
	fmt.Printf("%s\n", line)
}

func GetRegionDetailsByName() {
	fmt.Print(yellow + "Nome della regione: " + reset)
	name := utils.ReadInput()
	region, err := api.GetRegionByName(name)
	if err != nil {
		logger.Log.Error("Errore nel recupero della regione: ", err)
		return
	}
	if region == nil {
		fmt.Println("🚫 Regione non trovata.")
		return
	}

	line := strings.Repeat(sepHeavy, 70)
	fmt.Printf("%s\n%s🌍 Dettagli Regione%s\n%s\n", line, cyanBold, reset, line)
	fmt.Printf("  🆔️  Nome:           %s\n", region.Name)
	fmt.Printf("  🏢  Macrozone:      %d\n", region.MacrozoneCount)
	fmt.Printf("%s\n", line)

	// Macrozone
	fmt.Printf("%sMacrozone%s\n", cyanBold, reset)
	if len(region.Macrozones) > 0 {
		fmt.Printf("%-20s │ %-12s │ %-12s │ %-8s\n", "Nome", "Latitudine", "Longitudine", "Zone")
		fmt.Println(strings.Repeat(sepLight, 62))
		for _, m := range region.Macrozones {
			fmt.Printf("%-20s │ %12.6f │ %12.6f │ %8d\n", m.Name, m.Lat, m.Lon, m.ZoneCount)
		}
	} else {
		fmt.Println("  🚫 Nessuna macrozona associata.")
	}
	fmt.Printf("%s\n", line)

	// Hub
	fmt.Printf("%sHub di Regione%s\n", cyanBold, reset)
	if len(region.Hubs) > 0 {
		fmt.Printf("%-36s │ %-18s │ %-19s │ %-19s\n", "ID", "Servizio", "Registrato", "Ultima attività")
		fmt.Println(strings.Repeat(sepLight, 100))
		for _, h := range region.Hubs {
			color := reset
			diff := int(time.Now().Sub(h.LastSeen).Minutes())
			if diff > environment.UnhealthyTime {
				color = red
			} else {
				color = green
			}
			fmt.Printf("%s%-36s │ %-18s │ %-19s │ %-19s%s\n",
				color,
				h.Id, h.Service,
				h.RegistrationTime.Format("2006-01-02 15:04:05"),
				h.LastSeen.Format("2006-01-02 15:04:05"),
				reset,
			)
		}
	} else {
		fmt.Println("  🚫 Nessun hub associato alla regione.")
	}
	fmt.Printf("%s\n", line)
}
