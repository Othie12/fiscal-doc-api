package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/othie12/scanner-api/config"
	"github.com/othie12/scanner-api/internals/api/services"
	database "github.com/othie12/scanner-api/internals/db"
	"github.com/othie12/scanner-api/internals/db/models"
)

func main() {
	config.LoadConfig()
	if err := database.Connect(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	SeedUser()
}

func SeedUser() {
	userService := services.NewUserService()
	var dto models.User
	jsonFile, err := os.Open("seed.json")

	if err != nil {
		fmt.Println("Failed to open seed.json", err)
		return
	}
	defer jsonFile.Close()

	fmt.Println("Successfully Opened seed.json")

	byteValue, _ := io.ReadAll(jsonFile)

	json.Unmarshal(byteValue, &dto)
	if dto.Username == "" {
		fmt.Println("User's object contains no content")
		return
	}

	user, _, err := userService.Create(dto)
	if err != nil {
		fmt.Println("Encountered an error while creating seed user: ", err.Error())
		return
	}

	fmt.Printf("User seeded successfully\n%v\n\n", user)
}
