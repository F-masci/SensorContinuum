package main

import (
	"SensorContinuum/internal/client/comunication"
	"fmt"
	"log"
)

func main() {

	c := comunication.NewClient("http://localhost:4566/restapis/9h0c72wjbe/dev/_user_request_")
	data, err := c.GetData("region")
	if err != nil {
		log.Fatalf("Errore nella richiesta: %v", err)
	}

	fmt.Println("Risposta dal server:")
	fmt.Println(data)
}
