package projeto

import (
	"ARIC/src/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRoutes registra as rotas de projetos. Todas exigem login (auth).
// As de moderação exigem papel moderador ou admin.
func SetupRoutes(r *gin.RouterGroup, ct *Controller, repo *Repository, auth gin.HandlerFunc) {
	v := r.Group("", auth)

	v.GET("/categorias", ct.Categorias)
	v.GET("/dificuldades", ct.Dificuldades)

	p := v.Group("/projetos")
	p.GET("", ct.Listar)
	p.POST("", ct.Criar)
	p.POST("/imagens", ct.EnviarImagem)
	p.GET("/meus", ct.Meus)
	p.GET("/salvos", ct.Salvos)
	p.GET("/destaque", ct.Destaque)

	p.GET("/:id", ct.Detalhe)
	p.PUT("/:id", ct.Editar)
	p.DELETE("/:id", ct.Remover)

	p.POST("/:id/conclusao", ct.Concluir)
	p.PUT("/:id/curtida", ct.Curtir)
	p.DELETE("/:id/curtida", ct.Descurtir)
	p.PUT("/:id/salvo", ct.Salvar)
	p.DELETE("/:id/salvo", ct.RemoverSalvo)
	p.PUT("/:id/avaliacao", ct.Avaliar)

	p.GET("/:id/comentarios", ct.ListarComentarios)
	p.POST("/:id/comentarios", ct.Comentar)
	p.DELETE("/:id/comentarios/:comentarioId", ct.RemoverComentario)

	mod := v.Group("/moderacao", middleware.ExigirPapel(repo.Papel, "moderador", "admin"))
	mod.GET("/projetos", ct.FilaModeracao)
	mod.POST("/projetos/:id/aprovar", ct.Aprovar)
	mod.POST("/projetos/:id/rejeitar", ct.Rejeitar)
	mod.PUT("/destaque", ct.DefinirDestaque)
}
