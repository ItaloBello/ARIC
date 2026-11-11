package router

import (
	"ARIC/src/forum"
	"ARIC/src/login"
	"ARIC/src/middleware"
	"ARIC/src/projeto"
	"ARIC/utils/token"
	"ARIC/utils/viacep"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func SetupRoutes(r *gin.Engine, db *sqlx.DB, tokens *token.Manager) {
	api := r.Group("/api")
	v1 := api.Group("/v1")

	v1.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	auth := middleware.Auth(tokens)
	cep := viacep.NewClient(os.Getenv("VIACEP_URL"))

	loginRepo := login.NewLoginRepository(db)
	loginUC := login.NewLoginUseCases(loginRepo, cep, tokens)
	login.SetupLoginRoutes(v1, login.NewLoginController(loginUC), auth)

	projetoRepo := projeto.NewRepository(db)
	projetoUC := projeto.NewUseCases(projetoRepo)
	projeto.SetupRoutes(v1, projeto.NewController(projetoUC), projetoRepo, auth)

	forumUC := forum.NewUseCases(forum.NewRepository(db))
	forum.SetupRoutes(v1, forum.NewController(forumUC), auth)
}
