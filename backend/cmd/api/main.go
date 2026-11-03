package main

import (
	"ARIC/src/infra/database"
	"ARIC/src/infra/router"
	"ARIC/src/middleware"
	"ARIC/utils/token"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Em produção as variáveis podem vir do ambiente, então o .env é opcional
	if err := godotenv.Load(); err != nil {
		log.Println("arquivo .env não encontrado, usando variáveis de ambiente")
	}

	db, err := database.ConnectDB()
	if err != nil {
		log.Fatal("erro ao conectar no banco: ", err)
	}
	defer db.Close()

	tokens, err := token.NewManagerFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	//gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.MaxMultipartMemory = 8 << 20 // 8MB
	r.Use(middleware.CleanJSON(), middleware.Cors())
	r.Static("/uploads", "./uploads")

	router.SetupRoutes(r, db, tokens)

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("Server running on port", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("erro ao subir o servidor: ", err)
	}
}
