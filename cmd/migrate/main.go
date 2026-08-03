package main

import (
	"fmt"

	"github.com/othie12/scanner-api/config"
	database "github.com/othie12/scanner-api/internals/db"
)

func main() {
	config.LoadConfig()
	if err := database.MysqlConnect(); err != nil {
		fmt.Printf("Failed to initialize database: %v", err)
	}

	if err := database.Migrate(); err != nil {
		fmt.Printf("Failed to run migrations: %v", err)
	}

	fmt.Println("Migrations completed successfully")
}
