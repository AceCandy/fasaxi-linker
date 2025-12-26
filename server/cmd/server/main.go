package main

import (
	"fmt"
	"log"

	"github.com/fasaxi-linker/servergo/internal/api"
)

func main() {
	handler, err := api.NewHandler()
	if err != nil {
		log.Fatalf("Failed to initialize handler: %v", err)
	}

	r := api.SetupRouter(handler)

	params := "0.0.0.0:9090"
	fmt.Printf("Starting server on %s\n", params)
	if err := r.Run(params); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
