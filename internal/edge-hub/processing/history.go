package processing

import (
	"SensorContinuum/pkg/structure"
	"sync"
)

// History mantiene una finestra mobile degli ultimi dati ricevuti da un sensore.
// È thread-safe grazie all'uso di un RWMutex.
type History struct {
	Readings   []structure.SensorData
	WindowSize int
	mux        sync.RWMutex
}

// NewHistory crea una nuova istanza di History.
func NewHistory(windowSize int) *History {
	return &History{
		Readings:   make([]structure.SensorData, 0, windowSize),
		WindowSize: windowSize,
	}
}

// Add aggiunge un nuovo dato alla storia, mantenendo la dimensione della finestra.
func (h *History) Add(data structure.SensorData) {
	h.mux.Lock()
	defer h.mux.Unlock()

	// Aggiunge il nuovo dato
	h.Readings = append(h.Readings, data)

	// Se la finestra è piena, scarta il dato più vecchio
	if len(h.Readings) > h.WindowSize {
		h.Readings = h.Readings[1:]
	}
}

// GetReadings ritorna una copia sicura dei dati presenti nella storia.
func (h *History) GetReadings() []structure.SensorData {
	h.mux.RLock()
	defer h.mux.RUnlock()

	// Ritorna una copia per evitare race condition sull'uso dei dati
	readingsCopy := make([]structure.SensorData, len(h.Readings))
	copy(readingsCopy, h.Readings)
	return readingsCopy
}
