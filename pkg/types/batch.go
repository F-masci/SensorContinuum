package types

import (
	"errors"
	"sync"
	"time"
)

// BatchEngine è un motore di batching generico che raccoglie elementi di tipo T e li elabora in batch
type BatchEngine[T any] struct {
	items    []T
	counter  int
	ticker   *time.Ticker
	saveFunc func(*BatchEngine[T]) error
	maxCount int
	timeout  time.Duration
	stopChan chan struct{}
	mu       sync.Mutex
}

// NewBatchEngine crea un batch generico
func NewBatchEngine[T any](maxCount int, timeout time.Duration, save func(*BatchEngine[T]) error) (*BatchEngine[T], error) {

	// Controlla che i parametri siano validi
	if save == nil {
		return nil, errors.New("save function cannot be nil")
	}
	if maxCount <= 0 {
		return nil, errors.New("maxCount must be greater than 0")
	}
	if timeout <= 0 {
		return nil, errors.New("timeout must be greater than 0")
	}

	// Inizializza il batch
	be := &BatchEngine[T]{
		items:    make([]T, 0),
		counter:  0,
		saveFunc: save,
		maxCount: maxCount,
		timeout:  timeout,
		stopChan: make(chan struct{}),
	}

	// Avvia il ticker per il timeout
	be.startTicker()
	return be, nil
}

// startTicker avvia un ticker che salva i dati ogni timeout se ci sono dati nel batch
func (be *BatchEngine[T]) startTicker() {

	// Se il ticker esiste già, fermalo
	if be.ticker != nil {
		be.ticker.Stop()
	}

	// Crea un nuovo ticker
	be.ticker = time.NewTicker(be.timeout)
	go func() {
		for {
			select {
			// Al timeout del ticker, salva i dati se ce ne sono
			case <-be.ticker.C:
				// Acquisisci il lock per evitare salvataggi concorrenti
				if !be.mu.TryLock() {
					// Se non riesce a prendere il lock, significa che un altro
					// processo di salvataggio è in corso, quindi salta questo ciclo
					continue
				}

				// Salva i dati se la funzione è definita
				if be.saveFunc != nil {
					_ = be.saveFunc(be)
				}

				// Dopo il salvataggio, resetta il batch e riavvia il ticker
				// rilasciando il lock
				be.Clear()
				be.mu.Unlock()
			case <-be.stopChan:
				return
			}
		}
	}()
}

// Add aggiunge un nuovo dato al batch e salva se il batch è pieno
func (be *BatchEngine[T]) Add(item T) {

	// Acquisisci il lock per evitare race condition
	be.mu.Lock()
	defer be.mu.Unlock()

	if be.counter >= be.maxCount && be.saveFunc != nil {
		// Se il batch è pieno, salva i dati
		_ = be.saveFunc(be)
		// Dopo il salvataggio, resetta il batch
		be.Clear()
	}
	be.items = append(be.items, item)
	be.counter++
}

func (be *BatchEngine[T]) Clear() {
	be.items = make([]T, 0)
	be.counter = 0
}

func (be *BatchEngine[T]) Stop() {

	if be.stopChan != nil {
		close(be.stopChan)
		be.stopChan = nil
	}

	if be.ticker != nil {
		be.ticker.Stop()
		be.ticker = nil
	}

}

func (be *BatchEngine[T]) Count() int {
	return be.counter
}

func (be *BatchEngine[T]) Items() []T {
	return be.items
}
