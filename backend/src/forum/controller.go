package forum

import (
	"net/http"

	"ARIC/utils/web"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	uc *UseCases
}

func NewController(uc *UseCases) *Controller {
	return &Controller{uc: uc}
}

// ---------- posts ----------

func (ct *Controller) Listar(c *gin.Context) {
	uid, ok := web.Usuario(c)
	if !ok {
		return
	}
	var f Filtros
	if err := c.ShouldBindQuery(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "filtros inválidos"})
		return
	}
	pag, err := ct.uc.Listar(c.Request.Context(), uid, f)
	web.Responder(c, http.StatusOK, pag, err)
}

func (ct *Controller) Salvos(c *gin.Context) {
	uid, ok := web.Usuario(c)
	if !ok {
		return
	}
	pag, err := ct.uc.Salvos(c.Request.Context(), uid, web.QueryInt(c, "pagina"), web.QueryInt(c, "limite"))
	web.Responder(c, http.StatusOK, pag, err)
}

func (ct *Controller) Detalhe(c *gin.Context) {
	uid, id, ok := web.UsuarioEID(c, "id")
	if !ok {
		return
	}
	d, err := ct.uc.Detalhe(c.Request.Context(), uid, id)
	web.Responder(c, http.StatusOK, d, err)
}

func (ct *Controller) EnviarImagem(c *gin.Context) {
	if _, ok := web.Usuario(c); !ok {
		return
	}
	arquivo, err := c.FormFile("imagem")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "envie o arquivo no campo 'imagem'"})
		return
	}
	url, err := ct.uc.EnviarImagem(arquivo)
	web.Responder(c, http.StatusCreated, gin.H{"url": url}, err)
}

func (ct *Controller) Criar(c *gin.Context) {
	uid, ok := web.Usuario(c)
	if !ok {
		return
	}
	var req PostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados do post inválidos"})
		return
	}
	p, err := ct.uc.Criar(c.Request.Context(), uid, req)
	web.Responder(c, http.StatusCreated, p, err)
}

func (ct *Controller) Editar(c *gin.Context) {
	uid, id, ok := web.UsuarioEID(c, "id")
	if !ok {
		return
	}
	var req PostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados do post inválidos"})
		return
	}
	p, err := ct.uc.Editar(c.Request.Context(), uid, id, req)
	web.Responder(c, http.StatusOK, p, err)
}

func (ct *Controller) Remover(c *gin.Context) {
	uid, id, ok := web.UsuarioEID(c, "id")
	if !ok {
		return
	}
	err := ct.uc.Remover(c.Request.Context(), uid, id)
	web.Responder(c, http.StatusNoContent, nil, err)
}

func (ct *Controller) curtirPost(c *gin.Context, curtir bool) {
	uid, id, ok := web.UsuarioEID(c, "id")
	if !ok {
		return
	}
	resp, err := ct.uc.CurtirPost(c.Request.Context(), uid, id, curtir)
	web.Responder(c, http.StatusOK, resp, err)
}

func (ct *Controller) CurtirPost(c *gin.Context)    { ct.curtirPost(c, true) }
func (ct *Controller) DescurtirPost(c *gin.Context) { ct.curtirPost(c, false) }

func (ct *Controller) salvar(c *gin.Context, salvar bool) {
	uid, id, ok := web.UsuarioEID(c, "id")
	if !ok {
		return
	}
	err := ct.uc.SalvarPost(c.Request.Context(), uid, id, salvar)
	web.Responder(c, http.StatusOK, gin.H{"salvo": salvar}, err)
}

func (ct *Controller) Salvar(c *gin.Context)       { ct.salvar(c, true) }
func (ct *Controller) RemoverSalvo(c *gin.Context) { ct.salvar(c, false) }

func (ct *Controller) Aceitar(c *gin.Context) {
	uid, id, ok := web.UsuarioEID(c, "id")
	if !ok {
		return
	}
	var body struct {
		RespostaID int `json:"respostaId"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.RespostaID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "informe respostaId"})
		return
	}
	resp, err := ct.uc.Aceitar(c.Request.Context(), uid, id, body.RespostaID)
	web.Responder(c, http.StatusOK, resp, err)
}

func (ct *Controller) Desmarcar(c *gin.Context) {
	uid, id, ok := web.UsuarioEID(c, "id")
	if !ok {
		return
	}
	resp, err := ct.uc.Desmarcar(c.Request.Context(), uid, id)
	web.Responder(c, http.StatusOK, resp, err)
}

// ---------- respostas ----------

func (ct *Controller) ListarRespostas(c *gin.Context) {
	uid, id, ok := web.UsuarioEID(c, "id")
	if !ok {
		return
	}
	pag, err := ct.uc.ListarRespostas(c.Request.Context(), uid, id, web.QueryInt(c, "pagina"), web.QueryInt(c, "limite"))
	web.Responder(c, http.StatusOK, pag, err)
}

type conteudoBody struct {
	Conteudo string `json:"conteudo"`
}

func (ct *Controller) Responder(c *gin.Context) {
	uid, id, ok := web.UsuarioEID(c, "id")
	if !ok {
		return
	}
	var body conteudoBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "informe o conteúdo"})
		return
	}
	resp, err := ct.uc.Responder(c.Request.Context(), uid, id, body.Conteudo)
	web.Responder(c, http.StatusCreated, resp, err)
}

func (ct *Controller) EditarResposta(c *gin.Context) {
	uid, id, ok := web.UsuarioEID(c, "id")
	if !ok {
		return
	}
	var body conteudoBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "informe o conteúdo"})
		return
	}
	resp, err := ct.uc.EditarResposta(c.Request.Context(), uid, id, body.Conteudo)
	web.Responder(c, http.StatusOK, resp, err)
}

func (ct *Controller) RemoverResposta(c *gin.Context) {
	uid, id, ok := web.UsuarioEID(c, "id")
	if !ok {
		return
	}
	err := ct.uc.RemoverResposta(c.Request.Context(), uid, id)
	web.Responder(c, http.StatusNoContent, nil, err)
}

func (ct *Controller) curtirResposta(c *gin.Context, curtir bool) {
	uid, id, ok := web.UsuarioEID(c, "id")
	if !ok {
		return
	}
	resp, err := ct.uc.CurtirResposta(c.Request.Context(), uid, id, curtir)
	web.Responder(c, http.StatusOK, resp, err)
}

func (ct *Controller) CurtirResposta(c *gin.Context)    { ct.curtirResposta(c, true) }
func (ct *Controller) DescurtirResposta(c *gin.Context) { ct.curtirResposta(c, false) }
