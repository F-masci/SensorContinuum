package edge_hub

import (
	"math/rand"
)

// Struttura che espone il metodo
type SensorIDService struct{}

// Argomenti e risposta (puoi personalizzare)
type EmptyArgs struct{}
type IDReply struct {
	ID string
}

// Metodo esposto via RPC
func (s *SensorIDService) GetNextID(args *EmptyArgs, reply *IDReply) error {
	reply.ID = "sensor-" + generateR(8)
	return nil
}

// Funzione per generare ID casuale
func generateR(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
