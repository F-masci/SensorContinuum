package utils

import (
	"os"
	"os/signal"
	"syscall"
)

func WaitForTerminationSignal() {
	// creazione canale quit che attende segnali per terminare in modo controllato

	quit := make(chan os.Signal, 1)
	// il canale quit riceve i segnali SIGINT e SIGTERM
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}
