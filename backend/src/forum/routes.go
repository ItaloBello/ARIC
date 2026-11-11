package forum

import "github.com/gin-gonic/gin"

// SetupRoutes registra as rotas do fórum. Todas exigem login (auth).
func SetupRoutes(r *gin.RouterGroup, ct *Controller, auth gin.HandlerFunc) {
	f := r.Group("/forum", auth)

	f.POST("/imagens", ct.EnviarImagem)

	p := f.Group("/posts")
	p.GET("", ct.Listar)
	p.POST("", ct.Criar)
	p.GET("/salvos", ct.Salvos)

	p.GET("/:id", ct.Detalhe)
	p.PUT("/:id", ct.Editar)
	p.DELETE("/:id", ct.Remover)

	p.PUT("/:id/curtida", ct.CurtirPost)
	p.DELETE("/:id/curtida", ct.DescurtirPost)
	p.PUT("/:id/salvo", ct.Salvar)
	p.DELETE("/:id/salvo", ct.RemoverSalvo)

	p.PUT("/:id/resposta-aceita", ct.Aceitar)
	p.DELETE("/:id/resposta-aceita", ct.Desmarcar)

	p.GET("/:id/respostas", ct.ListarRespostas)
	p.POST("/:id/respostas", ct.Responder)

	resp := f.Group("/respostas")
	resp.PUT("/:id", ct.EditarResposta)
	resp.DELETE("/:id", ct.RemoverResposta)
	resp.PUT("/:id/curtida", ct.CurtirResposta)
	resp.DELETE("/:id/curtida", ct.DescurtirResposta)
}
