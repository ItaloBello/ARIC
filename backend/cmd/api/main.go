package main

import (
	"ARIC/src/infra/database"
	"ARIC/src/infra/router"
	"ARIC/src/middleware"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		panic(err)
	}

	db, err := database.ConnectDB()
	if err != nil {
		fmt.Println("Error connecting to the database")
		panic(err)
	}

	//gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(middleware.CleanJSON(), middleware.Cors())
	r.Static("/uploads", "./uploads")

	router.SetupRoutes(r, db)

	port := fmt.Sprintf(":%s", os.Getenv("API_PORT"))
	fmt.Println("Server running on port", port)
	if err := r.Run(port); err != nil {
		fmt.Println("Error running server:", err)
		panic(err)
	}
}
