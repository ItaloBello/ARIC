package router

import (
	"ARIC/src/forum"
	"ARIC/src/gamificacao"
	"ARIC/src/login"
	"ARIC/src/middleware"
	"ARIC/src/projeto"
	"ARIC/utils/token"
	"ARIC/utils/viacep"
	"context"
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

	loginRepo := login.NewLoginRepository(db)
	auth := middleware.Auth(tokens, loginRepo.SessaoValida)
	cep := viacep.NewClient(os.Getenv("VIACEP_URL"))

	loginUC := login.NewLoginUseCases(loginRepo, cep, tokens)
	login.SetupLoginRoutes(v1, login.NewLoginController(loginUC), auth)

	projetoRepo := projeto.NewRepository(db)
	projetoUC := projeto.NewUseCases(projetoRepo)
	projeto.SetupRoutes(v1, projeto.NewController(projetoUC), projetoRepo, auth)

	forumUC := forum.NewUseCases(forum.NewRepository(db))
	forum.SetupRoutes(v1, forum.NewController(forumUC), auth)

	gamRepo := gamificacao.NewRepository(db)
	gamificacao.SetupRoutes(v1, gamificacao.NewController(gamificacao.NewUseCases(gamRepo)), auth, projetoRepo.Papel)

	// Fechamento semanal das ligas. Use LIGAS_JOB=off se preferir rodar pelo cron do servidor.
	if os.Getenv("LIGAS_JOB") != "off" {
		gamificacao.IniciarFechamentoLigas(context.Background(), gamRepo)
	}
}
