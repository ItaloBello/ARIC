package router

import (
	"ARIC/src/login"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func SetupRoutes(r *gin.Engine, db *sqlx.DB) {
	api := r.Group("/api")
	v1 := api.Group("/v1")

	v1.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	
	login.SetupLoginRoutes(v1, login.NewLoginController(login.NewLoginUseCases(login.NewLoginRepository(db))))
}
