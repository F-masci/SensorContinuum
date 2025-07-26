package region

import (
	"SensorContinuum/internal/client/comunication/api"
	"SensorContinuum/pkg/utils"
	"fmt"
)

func listRegions() {
	regions, err := api.GetRegions()
	if err != nil {
		fmt.Printf("Errore nel recupero delle regioni: %v\n", err)
		return
	}
	fmt.Println("Zone disponibili:")
	for _, r := range regions {
		fmt.Printf("- %s (Lat: %.2f, Lon: %.2f)\n", r.Name, r.Latitude, r.Longitude)
	}
}

func getRegionDetails() {
	fmt.Print("Nome della zona: ")
	name := utils.ReadInput()
	region, err := api.GetRegion(name)
	if err != nil {
		fmt.Printf("Errore nel recupero della zona: %v\n", err)
		return
	}
	fmt.Printf("Dettagli zona:\nNome: %s\nLatitudine: %.2f\nLongitudine: %.2f\n", region.Name, region.Latitude, region.Longitude)
}

/*func addRegion() {
	var r structure.Region
	fmt.Print("Nome nuova zona: ")
	r.Name = utils.ReadInput()
	fmt.Print("Latitudine: ")
	fmt.Scanf("%f\n", &r.Latitude)
	fmt.Print("Longitudine: ")
	fmt.Scanf("%f\n", &r.Longitude)
	if err := api.CreateRegion(r); err != nil {
		fmt.Printf("Errore nell'aggiunta: %v\n", err)
	} else {
		fmt.Println("Zona aggiunta con successo.")
	}
}

func updateRegion() {
	var r structure.Region
	fmt.Print("Nome zona da modificare: ")
	r.Name = utils.ReadInput()
	fmt.Print("Nuova latitudine: ")
	fmt.Scanf("%f\n", &r.Latitude)
	fmt.Print("Nuova longitudine: ")
	fmt.Scanf("%f\n", &r.Longitude)
	if err := api.UpdateRegion(r); err != nil {
		fmt.Printf("Errore nella modifica: %v\n", err)
	} else {
		fmt.Println("Zona modificata con successo.")
	}
}

func deleteRegion() {
	fmt.Print("Nome zona da eliminare: ")
	name := utils.ReadInput()
	if err := api.DeleteRegion(name); err != nil {
		fmt.Printf("Errore nell'eliminazione: %v\n", err)
	} else {
		fmt.Println("Zona eliminata con successo.")
	}
}
*/
