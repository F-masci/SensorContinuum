package sensoragent_test

import (
	"testing"
	"time"

	"SensorContinuum/internal/sensor-agent"
)

// Test 1: resChan == nil → deve loggare, ma non scrivere
func TestRun_NilChannel(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Run panicked with nil channel: %v", r)
		}
	}()
	sensor_agent.Run(2, nil) // Deve semplicemente ignorare il canale
}

// Test 2: Channel buffered con spazio sufficiente
func TestRun_BufferedChannel(t *testing.T) {
	n := 3
	ch := make(chan float64, n)

	go sensor_agent.Run(n, ch)

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
func TestRun_FullChannel(t *testing.T) {
	ch := make(chan float64, 1)
	ch <- 42.0 // Riempie il buffer

	done := make(chan struct{})
	go func() {
		sensor_agent.Run(1, ch)
		close(done)
	}()

	select {
	case <-done:
		// Success: funzione non si è bloccata, ha loggato l'errore e continuato
	case <-time.After(3 * time.Second):
		t.Fatal("Run should not block even when channel is full")
	}
}

// Test 4: Unbuffered channel (funziona solo se ricevi subito)
func TestRun_UnbufferedChannel(t *testing.T) {
	ch := make(chan float64) // unbuffered
	n := 1

	go func() {
		sensor_agent.Run(n, ch)
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
