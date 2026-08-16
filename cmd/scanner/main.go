package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/othie12/scanner-api/config"
	database "github.com/othie12/scanner-api/internals/db"
	"github.com/othie12/scanner-api/internals/routes"
)

func main() {

	config.LoadConfig()

	if err := database.Connect(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	mainRouter := gin.Default()
	mainRouter.Use(cors.New(config.CorsConfig))

	// Serve static files
	mainRouter.Static("/static", "./static")
	mainRouter.Static("/assets", "./static/assets")
	// Serve index.html for React Router fallback
	mainRouter.NoRoute(func(c *gin.Context) {
		requestPath := c.Request.URL.Path
		staticPath := "./static" + requestPath

		if _, err := os.Stat(staticPath); err == nil {
			// File exists, serve it
			c.File(staticPath)
		} else {
			// Otherwise fallback to index.html for React Router
			c.File("./static/index.html")
		}
	})

	// Api routes
	routes.ApiRouter(mainRouter)

	// mainRouter.GET("/", func(c *gin.Context) {
	// 	c.JSON(200, gin.H{"message": "Welcome to " + config.ServerConfig.Name})
	// })

	fmt.Println("Attempting to start server on port: " + config.ServerConfig.Port)
	if err := mainRouter.Run(":" + config.ServerConfig.Port); err != nil {
		fmt.Println("Failed to start server: " + err.Error())
	}
	fmt.Printf("Server started on host: %s and port %s\n", config.ServerConfig.Name, config.ServerConfig.Port)
}
