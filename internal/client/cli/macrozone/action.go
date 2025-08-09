package macrozone

import (
	"SensorContinuum/internal/client/comunication/api"
	"SensorContinuum/internal/client/environment"
	"SensorContinuum/pkg/logger"
	"SensorContinuum/pkg/utils"
	"fmt"
	"strings"
	"time"
)

func listMacrozones(regionName string) {
	macrozones, err := api.GetMacrozones(regionName)
	if err != nil {
		logger.Log.Error("Errore nel recupero delle macrozone: ", err)
		return
	}
	logger.Log.Debug("🔎 Trovate ", len(macrozones), " macrozone")

	line := strings.Repeat(sepHeavy, 70)
	fmt.Printf("%s\n%s🏢 Macrozone disponibili (%d)%s\n%s\n", line, cyanBold, len(macrozones), reset, line)
	if len(macrozones) == 0 {
		fmt.Println("  🚫 Nessuna macrozona trovata.")
	} else {
		fmt.Printf("%-20s │ %-12s │ %-12s │ %-8s\n", "Nome", "Latitudine", "Longitudine", "Zone")
		fmt.Println(strings.Repeat(sepLight, 62))
		for _, m := range macrozones {
			fmt.Printf("%s%-20s%s │ %12.6f │ %12.6f │ %8d\n", green, m.Name, reset, m.Lat, m.Lon, m.ZoneCount)
		}
	}
	fmt.Printf("%s\n", line)
}

func getMacrozoneByName(regionName string) {
	fmt.Print(yellow + "Nome macrozona: " + reset)
	name := utils.ReadInput()
	macrozone, err := api.GetMacrozoneByName(regionName, name)
	if err != nil {
		logger.Log.Error("Errore nel recupero della macrozona: ", err)
		return
	}
	if macrozone == nil {
		fmt.Println("🚫 Macrozona non trovata.")
		return
	}

	line := strings.Repeat(sepHeavy, 70)
	fmt.Printf("%s\n%s🏢 Dettagli Macrozona%s\n%s\n", line, cyanBold, reset, line)
	fmt.Printf("  🆔️  Nome:         %s\n", macrozone.Name)
	fmt.Printf("  🌍  Regione:      %s\n", regionName)
	fmt.Printf("  📍  Latitudine:   %.6f\n", macrozone.Lat)
	fmt.Printf("  📍  Longitudine:  %.6f\n", macrozone.Lon)
	fmt.Printf("  📅  Registrata il:%s\n", macrozone.CreationTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("  🏢  Numero zone:  %d\n", macrozone.ZoneCount)
	fmt.Printf("%s\n", line)

	// Zone
	fmt.Printf("%sZone%s\n", cyanBold, reset)
	if len(macrozone.Zones) > 0 {
		fmt.Printf("%-20s │ %-19s\n", "Nome", "Registrata il")
		fmt.Println(strings.Repeat(sepLight, 42))
		for _, z := range macrozone.Zones {
			fmt.Printf("%-20s │ %-19s\n", z.Name, z.CreationTime.Format("2006-01-02 15:04:05"))
		}
	} else {
		fmt.Println("  🚫 Nessuna zona associata.")
	}
	fmt.Printf("%s\n", line)

	// Hubs di macrozona
	fmt.Printf("%sHub di Macrozona%s\n", cyanBold, reset)
	if len(macrozone.Hubs) > 0 {
		fmt.Printf("%-36s │ %-18s │ %-19s │ %-19s\n", "ID", "Servizio", "Registrato", "Ultima attività")
		fmt.Println(strings.Repeat(sepLight, 100))
		for _, h := range macrozone.Hubs {
			color := reset
			if int(time.Now().Sub(h.LastSeen).Minutes()) > environment.UnhealthyTime {
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
		fmt.Println("  🚫 Nessun hub associato alla macrozona.")
	}
	fmt.Printf("%s\n", line)

	// Hubs di zona
	fmt.Printf("%sHub di Zona%s\n", cyanBold, reset)
	if len(macrozone.ZoneHubs) > 0 {
		fmt.Printf("%-38s │ %-22s │ %-22s │ %-20s │ %-19s │ %-19s\n", "ID", "Macrozona", "Zona", "Servizio", "Registrato", "Ultima attività")
		fmt.Println(strings.Repeat(sepLight, 160))
		for _, zh := range macrozone.ZoneHubs {
			color := reset
			if int(time.Now().Sub(zh.LastSeen).Minutes()) > environment.UnhealthyTime {
				color = red
			} else {
				color = green
			}
			fmt.Printf("%s%-38.38s │ %-22.22s │ %-22.22s │ %-20.20s │ %-19s │ %-19s%s\n",
				color,
				zh.Id, zh.MacrozoneName, zh.ZoneName, zh.Service,
				zh.RegistrationTime.Format("2006-01-02 15:04:05"),
				zh.LastSeen.Format("2006-01-02 15:04:05"),
				reset,
			)
		}
	} else {
		fmt.Println("  🚫 Nessun hub di zona associato alla macrozona.")
	}
	fmt.Printf("%s\n", line)

	// Sensori
	fmt.Printf("%sSensori Associati%s\n", cyanBold, reset)
	if len(macrozone.Sensors) > 0 {
		fmt.Printf("%-36s │ %-20s │ %-20s │ %-12s │ %-18s │ %-19s │ %-19s\n", "ID", "Macrozona", "Zona", "Tipo", "Riferimento", "Registrato", "Ultima attività")
		fmt.Println(strings.Repeat(sepLight, 160))
		for _, s := range macrozone.Sensors {
			color := reset
			if int(time.Now().Sub(s.LastSeen).Minutes()) > environment.UnhealthyTime {
				color = red
			} else {
				color = green
			}
			fmt.Printf("%s%-36s │ %-20s │ %-20s │ %-12s │ %-18s │ %-19s │ %-19s%s\n",
				color,
				s.Id, s.MacrozoneName, s.ZoneName, s.Type, s.Reference,
				s.RegistrationTime.Format("2006-01-02 15:04:05"),
				s.LastSeen.Format("2006-01-02 15:04:05"),
				reset,
			)
		}
	} else {
		fmt.Println("  🛰 Nessun sensore associato alla macrozona.")
	}
	fmt.Printf("%s\n", line)
}
