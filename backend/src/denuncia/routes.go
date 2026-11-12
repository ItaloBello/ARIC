package denuncia

import (
	"net/http"

	"ARIC/src/middleware"
	"ARIC/utils/web"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	uc *UseCases
}

func NewController(uc *UseCases) *Controller {
	return &Controller{uc: uc}
}

// SetupRoutes:
//
//	POST /denuncias                     qualquer usuário logado
//	GET  /moderacao/denuncias           fila agrupada por conteúdo (moderador/admin)
//	POST /moderacao/denuncias/resolver  remover ou manter (moderador/admin)
func SetupRoutes(r *gin.RouterGroup, ct *Controller, auth gin.HandlerFunc, buscarPapel middleware.BuscarPapel) {
	v := r.Group("", auth)
	v.POST("/denuncias", ct.Denunciar)

	mod := v.Group("/moderacao", middleware.ExigirPapel(buscarPapel, "moderador", "admin"))
	mod.GET("/denuncias", ct.Fila)
	mod.POST("/denuncias/resolver", ct.Resolver)
}

func (ct *Controller) Denunciar(c *gin.Context) {
	uid, ok := web.Usuario(c)
	if !ok {
		return
	}
	var req DenunciaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados da denúncia inválidos"})
		return
	}
	d, err := ct.uc.Denunciar(c.Request.Context(), uid, req)
	web.Responder(c, http.StatusCreated, d, err)
}

func (ct *Controller) Fila(c *gin.Context) {
	pag, err := ct.uc.Fila(c.Request.Context(), web.QueryInt(c, "pagina"), web.QueryInt(c, "limite"))
	web.Responder(c, http.StatusOK, pag, err)
}

func (ct *Controller) Resolver(c *gin.Context) {
	uid, ok := web.Usuario(c)
	if !ok {
		return
	}
	var req ResolverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados inválidos"})
		return
	}
	err := ct.uc.Resolver(c.Request.Context(), uid, req)
	web.Responder(c, http.StatusNoContent, nil, err)
}
