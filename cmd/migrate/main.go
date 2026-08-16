package main

import (
	"fmt"
	"log"

	"github.com/othie12/scanner-api/config"
	database "github.com/othie12/scanner-api/internals/db"
)

func main() {
	config.LoadConfig()
	if err := database.Connect(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	if err := database.Migrate(); err != nil {
		fmt.Printf("Failed to run migrations: %v", err)
	}

	fmt.Println("Migrations completed successfully")
}
