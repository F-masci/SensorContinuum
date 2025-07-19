package sensoragent_test

import (
	"SensorContinuum/internal/sensor-agent/simulation"
	"os"
	"testing"
	"time"

	"SensorContinuum/pkg/logger"
)

func initTestEnvironment() {
	// Inizializza il logger
	logger.CreateLogger(logger.Context{
		"service": "sensor-agent-test",
	})
}

func TestMain(m *testing.M) {
	initTestEnvironment()
	os.Exit(m.Run())
}

// Test 1: resChan == nil → deve loggare, ma non scrivere
func TestSimulate_NilChannel(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Simulate panicked with nil channel: %v", r)
		}
	}()
	simulation.Simulate(2, nil) // Deve semplicemente ignorare il canale
}

// Test 2: Channel buffered con spazio sufficiente
func TestSimulate_BufferedChannel(t *testing.T) {
	n := 3
	ch := make(chan float64, n)

	go simulation.Simulate(n, ch)

	timeout := time.After(time.Second * 5)
	count := 0

	for count < n {
		select {
		case val := <-ch:
			if val < 0 || val > 100 {
				t.Errorf("value out of range: %f", val)
			}
			count++
		case <-timeout:
			t.Fatalf("timeout waiting for values, got %d/%d", count, n)
		}
	}
}

// Test 3: Channel pieno → deve saltare invio (niente panic, niente blocco)
func TestSimulate_FullChannel(t *testing.T) {
	ch := make(chan float64, 1)
	ch <- 42.0 // Riempie il buffer

	done := make(chan struct{})
	go func() {
		simulation.Simulate(1, ch)
		close(done)
	}()

	select {
	case <-done:
		// Success: funzione non si è bloccata, ha loggato l'errore e continuato
	case <-time.After(3 * time.Second):
		t.Fatal("Simulate should not block even when channel is full")
	}
}

// Test 4: Unbuffered channel (funziona solo se ricevi subito)
func TestSimulate_UnbufferedChannel(t *testing.T) {
	ch := make(chan float64) // unbuffered
	n := 1

	go func() {
		simulation.Simulate(n, ch)
	}()

	select {
	case val := <-ch:
		if val < 0 || val > 100 {
			t.Errorf("value out of range: %f", val)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for value from unbuffered channel")
	}
}
