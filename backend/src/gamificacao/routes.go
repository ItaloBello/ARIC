package gamificacao

import (
	"ARIC/src/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRoutes registra as rotas de gamificação. Todas exigem login (auth).
// buscarPapel é usado para proteger a moderação de fotos de missão.
func SetupRoutes(r *gin.RouterGroup, ct *Controller, auth gin.HandlerFunc, buscarPapel middleware.BuscarPapel) {
	v := r.Group("", auth)

	v.GET("/usuarios/:id/perfil", ct.PerfilPublico)

	g := v.Group("/gamificacao")
	g.GET("/perfil", ct.MeuPerfil)
	g.GET("/niveis", ct.Niveis)
	g.GET("/xp", ct.HistoricoXP)
	g.GET("/impacto-bairro", ct.ImpactoBairro)

	g.GET("/ligas", ct.Ligas)
	g.GET("/ranking", ct.Ranking)

	g.GET("/missoes", ct.Missoes)
	g.POST("/missoes/imagens", ct.EnviarImagem)
	g.POST("/missoes/:id/evidencias", ct.EnviarEvidencia)
	g.GET("/evidencias", ct.MinhasEvidencias)

	mod := v.Group("/moderacao", middleware.ExigirPapel(buscarPapel, "moderador", "admin"))
	mod.GET("/evidencias", ct.EvidenciasPendentes)
	mod.POST("/evidencias/:id/aprovar", ct.AprovarEvidencia)
	mod.POST("/evidencias/:id/rejeitar", ct.RejeitarEvidencia)
}
