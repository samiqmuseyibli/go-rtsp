package main

import (
	"log"
	"os"

	"go-rtsp-streamer/internal/server"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8088"
	}

	srv := server.New()
	log.Printf("Starting RTSP Stream Manager on port %s", port)
	log.Fatal(srv.Start(":" + port))
}