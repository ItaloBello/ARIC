package main

import (
	"ARIC/src/infra/database"
	"ARIC/src/infra/router"
	"ARIC/src/middleware"
	"ARIC/utils/token"
	"fmt"
	"log"
	"os"
	"strings"

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

	// O limite de tentativas usa o IP do cliente. Atrás do Apache, o IP real vem no
	// X-Forwarded-For, mas só podemos confiar nesse cabeçalho quando ele vem do proxy;
	// senão qualquer um troca de "IP" a cada tentativa. Padrão: proxy na mesma máquina.
	proxies := strings.Split(os.Getenv("TRUSTED_PROXIES"), ",")
	if os.Getenv("TRUSTED_PROXIES") == "" {
		proxies = []string{"127.0.0.1", "::1"}
	}
	if err := r.SetTrustedProxies(proxies); err != nil {
		log.Fatal("TRUSTED_PROXIES inválido: ", err)
	}
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
