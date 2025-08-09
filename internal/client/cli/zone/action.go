package zone

import (
	"SensorContinuum/internal/client/comunication/api"
	"SensorContinuum/pkg/logger"
	"fmt"
	"strings"
)

func listZones(regionName, macrozoneName string) {
	zones, err := api.GetZones(regionName, macrozoneName)
	if err != nil {
		logger.Log.Error("Ultima Errore nel recupero delle zone: ", err)
		return
	}
	line := strings.Repeat(sepHeavy, 70)
	fmt.Printf("%s\n%sZone disponibili%s\n%s\n", line, cyanBold, reset, line)
	if len(zones) == 0 {
		fmt.Println("  🚫 Nessuna zona trovata.")
	} else {
		for _, z := range zones {
			fmt.Printf("%s- 🟢 %s%s %s(Creato il: %s)%s\n", green, z.Name, reset, yellow, z.CreationTime.Format("2006-01-02 15:04:05"), reset)
		}
	}
	fmt.Printf("%s\n", line)
}

func getZoneByName(regionName, macrozoneName, zoneName string) {
	zone, err := api.GetZoneByName(regionName, macrozoneName, zoneName)
	if err != nil {
		logger.Log.Error("Errore nel recupero della zona: ", err)
		return
	}
	if zone == nil {
		fmt.Println("🚫 Zona non trovata.")
		return
	}

	line := strings.Repeat(sepHeavy, 70)
	fmt.Printf("%s\n%sDettagli Zona%s\n%s\n", line, cyanBold, reset, line)
	fmt.Printf("  🆔️  Zona:        %s\n", zone.Name)
	fmt.Printf("  🏢  Macrozona:   %s\n", zone.MacrozoneName)
	fmt.Printf("  🌍  Regione:     %s\n", zone.RegionName)
	fmt.Printf("  📅  Creata il:   %s\n", zone.CreationTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("%s\n", line)

	// Hub di zona
	fmt.Printf("%sHub di Zona%s\n", cyanBold, reset)
	if len(zone.Hubs) > 0 {
		fmt.Printf("%-36s │ %-22s │ %-19s │ %-19s\n", "ID", "Servizio", "Registrazione", "Ultima attività")
		fmt.Println(strings.Repeat("─", 105))
		for _, hub := range zone.Hubs {
			fmt.Printf("%-36s │ %-22s │ %-19s │ %-19s\n",
				hub.Id,
				hub.Service,
				hub.RegistrationTime.Format("2006-01-02 15:04:05"),
				hub.LastSeen.Format("2006-01-02 15:04:05"),
			)
		}
	} else {
		fmt.Println("  🚫 Nessun hub di zona associato.")
	}
	fmt.Printf("%s\n", line)

	// Sensori associati
	fmt.Printf("%sSensori Associati%s\n", cyanBold, reset)
	if len(zone.Sensors) > 0 {
		fmt.Printf("%-36s │ %-15s │ %-15s │ %-15s │ %-20s │ %-20s\n",
			"ID", "Macrozona", "Zona", "Tipo", "Registrazione", "Ultima attività")
		fmt.Println(strings.Repeat("─", 130))
		for _, sensor := range zone.Sensors {
			fmt.Printf("%-36s │ %-15s │ %-15s │ %-15s │ %-20s │ %-20s\n",
				sensor.Id,
				sensor.MacrozoneName,
				sensor.ZoneName,
				sensor.Type,
				sensor.RegistrationTime.Format("2006-01-02 15:04:05"),
				sensor.LastSeen.Format("2006-01-02 15:04:05"),
			)
		}
	} else {
		fmt.Println("  🛰 Nessun sensore associato.")
	}
	fmt.Printf("%s\n", line)
}
