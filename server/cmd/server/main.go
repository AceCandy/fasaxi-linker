package main

import (
	"fmt"
	"log"

	"github.com/fasaxi-linker/servergo/internal/api"
	"github.com/fasaxi-linker/servergo/internal/db"
)

func main() {
	fmt.Println("🚀 Starting Fasaxi Linker Server...")

	dbConfig, err := db.LoadConfigFromEnv()
	if err != nil {
		log.Fatalf("❌ Database configuration error: %v\n\nPlease ensure the following environment variables are set:\n  - POSTGRES_HOST\n  - POSTGRES_PORT\n  - POSTGRES_USER\n  - POSTGRES_PASSWORD\n  - POSTGRES_DB\n", err)
	}

	fmt.Println("📊 Initializing database connection...")
	if err := db.InitDB(dbConfig); err != nil {
		log.Fatalf("❌ Failed to initialize database: %v\n", err)
	}
	defer db.Close()

	fmt.Println("🔧 Initializing API handler...")
	handler, err := api.NewHandler()
	if err != nil {
		log.Fatalf("❌ Failed to initialize handler: %v", err)
	}

	r := api.SetupRouter(handler)

	params := "0.0.0.0:9090"
	fmt.Printf("✅ Server ready! Listening on %s\n", params)
	if err := r.Run(params); err != nil {
		log.Fatalf("❌ Server failed: %v", err)
	}
}
