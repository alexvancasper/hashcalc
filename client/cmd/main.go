package main

import (
	"hash-service-client/internal/routes"
	"log"
	"syscall"
)

const DefaultPort = "8070"

func main() {
	port := DefaultPort
	if value, ok := syscall.Getenv("CS_PORT"); ok {
		port = value
	}
	log.Printf("Starting WebUI on port %s", port)
	routes.Start(port)
}
