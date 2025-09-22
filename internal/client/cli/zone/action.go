package zone

import (
	"SensorContinuum/internal/client/comunication/api"
	"SensorContinuum/internal/client/environment"
	"SensorContinuum/pkg/logger"
	"fmt"
	"strings"
	"time"
)

const timeFormat = "2006-01-02 15:04:05"

func listZones(regionName, macrozoneName string) {
	zones, err := api.GetZones(regionName, macrozoneName)
	if err != nil {
		logger.Log.Error("Errore nel recupero delle zone: ", err)
		return
	}
	line := strings.Repeat(sepHeavy, 70)
	fmt.Printf("%s\n%sZone disponibili%s\n%s\n", line, cyanBold, reset, line)
	if len(zones) == 0 {
		fmt.Println("  🚫 Nessuna zona trovata.")
	} else {
		for _, z := range zones {
			fmt.Printf("%s- 🟢 %s%s %s(Creato il: %s)%s\n", green, z.Name, reset, yellow, z.CreationTime.Format(timeFormat), reset)
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
	fmt.Printf("  📅  Creata il:   %s\n", zone.CreationTime.Format(timeFormat))
	fmt.Printf("%s\n", line)

	// Hub di zona
	fmt.Printf("%sHub di Zona%s\n", cyanBold, reset)
	if len(zone.Hubs) > 0 {
		fmt.Printf("%-36s │ %-22s │ %-19s │ %-19s\n", "ID", "Servizio", "Registrazione", "Ultima attività")
		fmt.Println(strings.Repeat("─", 105))
		for _, hub := range zone.Hubs {
			diff := int(time.Now().Sub(hub.LastSeen).Minutes())
			color := reset
			if diff > environment.UnhealthyTime {
				color = red
			} else {
				color = green
			}
			fmt.Printf("%s%-36s │ %-22s │ %-19s │ %-19s%s\n",
				color,
				hub.Id,
				hub.Service,
				hub.RegistrationTime.Local().Format(timeFormat),
				hub.LastSeen.Local().Format(timeFormat),
				reset,
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
			diff := int(time.Now().Sub(sensor.LastSeen).Minutes())
			color := reset
			if diff > environment.UnhealthyTime {
				color = red
			} else {
				color = green
			}
			fmt.Printf("%s%-36s │ %-15s │ %-15s │ %-15s │ %-20s │ %-20s%s\n",
				color,
				sensor.Id,
				sensor.MacrozoneName,
				sensor.ZoneName,
				sensor.Type,
				sensor.RegistrationTime.Local().Format(timeFormat),
				sensor.LastSeen.Local().Format(timeFormat),
				reset,
			)
		}
	} else {
		fmt.Println("  🛰 Nessun sensore associato.")
	}
	fmt.Printf("%s\n", line)
}

func getRawSensorData(regionName, macrozoneName, zoneName, sensorID string) {
	data, err := api.GetRawSensorData(regionName, macrozoneName, zoneName, sensorID)
	if err != nil {
		logger.Log.Error("Errore nel recupero dei dati del sensore: ", err)
		return
	}
	line := strings.Repeat(sepHeavy, 70)
	fmt.Printf("%s\n%s🔬 Dati Grezzi Sensore%s\n%s\n", line, cyanBold, reset, line)
	fmt.Printf("  🆔️  Sensore:     %s\n", sensorID)
	fmt.Printf("  📍  Zona:        %s\n", zoneName)
	fmt.Printf("  🏢  Macrozona:   %s\n", macrozoneName)
	fmt.Printf("  🌍  Regione:     %s\n", regionName)
	fmt.Printf("%s\n", line)

	if len(data) == 0 {
		fmt.Println("  🚫 Nessun dato trovato per il sensore specificato.")
	} else {
		fmt.Printf("%-22s │ %-14s │ %-10s\n", "Timestamp", "Tipo", "Valore")
		fmt.Println(strings.Repeat("─", 52))
		for _, d := range data {
			t := time.Unix(d.Timestamp, 0).Format(timeFormat)
			fmt.Printf("%s%-22s │ %-14s │ %-10.2f%s\n", green, t, d.Type, d.Data, reset)
		}
	}
	fmt.Printf("%s\n", line)
}
