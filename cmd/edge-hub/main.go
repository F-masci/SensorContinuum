package main

import (
	"net"
	"net/rpc"

	"SensorContinuum/internal/edge-hub"
	"SensorContinuum/pkg/logger"
)

// TODO: Configurare il contesto del logger dinamicamente
func getContext() logger.Context {
	return logger.Context{
		"service": "edge-hub",
		"hubID":   "hub-1",
	}
}

func idService(l net.Listener) {
	for {
		conn, err := l.Accept()
		if err != nil {
			logger.Log.Error("Error accepting connection: ", err)
			continue
		}
		logger.Log.Info("Accepted connection from ", conn.RemoteAddr())
		go rpc.ServeConn(conn)
	}
}

func main() {

	logger.CreateLogger(getContext())
	logger.Log.Info("Starting Edge Hub...")

	logger.Log.Info("Starting ID service...")
	rpc.Register(new(edge_hub.SensorIDService))
	l, err := net.Listen("tcp", ":1234")
	if err != nil {
		panic(err)
	}

	logger.Log.Info("ID service listening on port 1234...")
	go idService(l)
	defer l.Close()

	for {
		logger.Log.Info("Edge Hub is running...")
		select {}
	}

}
