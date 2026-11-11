package projeto

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

// ---------- catálogo ----------

func (ct *Controller) Categorias(c *gin.Context) {
	cats, err := ct.uc.Categorias(c.Request.Context())
	web.Responder(c, http.StatusOK, cats, err)
}

func (ct *Controller) Dificuldades(c *gin.Context) {
	difs, err := ct.uc.Dificuldades(c.Request.Context())
	web.Responder(c, http.StatusOK, difs, err)
}

// ---------- leitura ----------

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

func (ct *Controller) Meus(c *gin.Context) {
	uid, ok := web.Usuario(c)
	if !ok {
		return
	}
	pag, err := ct.uc.Meus(c.Request.Context(), uid, web.QueryInt(c, "pagina"), web.QueryInt(c, "limite"))
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

// Destaque devolve 204 quando não há destaque na semana.
func (ct *Controller) Destaque(c *gin.Context) {
	uid, ok := web.Usuario(c)
	if !ok {
		return
	}
	card, err := ct.uc.Destaque(c.Request.Context(), uid)
	if err == nil && card == nil {
		c.Status(http.StatusNoContent)
		return
	}
	web.Responder(c, http.StatusOK, card, err)
}

func (ct *Controller) Detalhe(c *gin.Context) {
	uid, pid, ok := web.UsuarioEID(c, "id")
	if !ok {
		return
	}
	d, err := ct.uc.Detalhe(c.Request.Context(), uid, pid)
	web.Responder(c, http.StatusOK, d, err)
}

// ---------- escrita ----------

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
	var req ProjetoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados do projeto inválidos"})
		return
	}
	d, err := ct.uc.Criar(c.Request.Context(), uid, req)
	web.Responder(c, http.StatusCreated, d, err)
}

func (ct *Controller) Editar(c *gin.Context) {
	uid, pid, ok := web.UsuarioEID(c, "id")
	if !ok {
		return
	}
	var req ProjetoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados do projeto inválidos"})
		return
	}
	d, err := ct.uc.Editar(c.Request.Context(), uid, pid, req)
	web.Responder(c, http.StatusOK, d, err)
}

func (ct *Controller) Remover(c *gin.Context) {
	uid, pid, ok := web.UsuarioEID(c, "id")
	if !ok {
		return
	}
	err := ct.uc.Remover(c.Request.Context(), uid, pid)
	web.Responder(c, http.StatusNoContent, nil, err)
}

// ---------- interações ----------

func (ct *Controller) Concluir(c *gin.Context) {
	uid, pid, ok := web.UsuarioEID(c, "id")
	if !ok {
		return
	}
	var body struct {
		FotoURL string `json:"fotoUrl"`
	}
	// corpo é opcional
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "corpo inválido"})
			return
		}
	}
	resp, err := ct.uc.Concluir(c.Request.Context(), uid, pid, body.FotoURL)
	web.Responder(c, http.StatusCreated, resp, err)
}

func (ct *Controller) curtir(c *gin.Context, curtir bool) {
	uid, pid, ok := web.UsuarioEID(c, "id")
	if !ok {
		return
	}
	resp, err := ct.uc.Curtir(c.Request.Context(), uid, pid, curtir)
	web.Responder(c, http.StatusOK, resp, err)
}

func (ct *Controller) Curtir(c *gin.Context)    { ct.curtir(c, true) }
func (ct *Controller) Descurtir(c *gin.Context) { ct.curtir(c, false) }

func (ct *Controller) salvar(c *gin.Context, salvar bool) {
	uid, pid, ok := web.UsuarioEID(c, "id")
	if !ok {
		return
	}
	err := ct.uc.Salvar(c.Request.Context(), uid, pid, salvar)
	web.Responder(c, http.StatusOK, gin.H{"salvo": salvar}, err)
}

func (ct *Controller) Salvar(c *gin.Context)       { ct.salvar(c, true) }
func (ct *Controller) RemoverSalvo(c *gin.Context) { ct.salvar(c, false) }

func (ct *Controller) Avaliar(c *gin.Context) {
	uid, pid, ok := web.UsuarioEID(c, "id")
	if !ok {
		return
	}
	var body struct {
		Nota int `json:"nota"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "informe a nota"})
		return
	}
	resp, err := ct.uc.Avaliar(c.Request.Context(), uid, pid, body.Nota)
	web.Responder(c, http.StatusOK, resp, err)
}

// ---------- comentários ----------

func (ct *Controller) ListarComentarios(c *gin.Context) {
	uid, pid, ok := web.UsuarioEID(c, "id")
	if !ok {
		return
	}
	pag, err := ct.uc.ListarComentarios(c.Request.Context(), uid, pid, web.QueryInt(c, "pagina"), web.QueryInt(c, "limite"))
	web.Responder(c, http.StatusOK, pag, err)
}

func (ct *Controller) Comentar(c *gin.Context) {
	uid, pid, ok := web.UsuarioEID(c, "id")
	if !ok {
		return
	}
	var body struct {
		Conteudo string `json:"conteudo"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "informe o conteúdo"})
		return
	}
	com, err := ct.uc.Comentar(c.Request.Context(), uid, pid, body.Conteudo)
	web.Responder(c, http.StatusCreated, com, err)
}

func (ct *Controller) RemoverComentario(c *gin.Context) {
	uid, pid, ok := web.UsuarioEID(c, "id")
	if !ok {
		return
	}
	cid, ok := web.ParamID(c, "comentarioId")
	if !ok {
		return
	}
	err := ct.uc.RemoverComentario(c.Request.Context(), uid, pid, cid)
	web.Responder(c, http.StatusNoContent, nil, err)
}

// ---------- moderação ----------

func (ct *Controller) FilaModeracao(c *gin.Context) {
	uid, ok := web.Usuario(c)
	if !ok {
		return
	}
	pag, err := ct.uc.FilaModeracao(c.Request.Context(), uid, web.QueryInt(c, "pagina"), web.QueryInt(c, "limite"))
	web.Responder(c, http.StatusOK, pag, err)
}

func (ct *Controller) Aprovar(c *gin.Context) {
	uid, pid, ok := web.UsuarioEID(c, "id")
	if !ok {
		return
	}
	err := ct.uc.Aprovar(c.Request.Context(), uid, pid)
	web.Responder(c, http.StatusNoContent, nil, err)
}

func (ct *Controller) Rejeitar(c *gin.Context) {
	uid, pid, ok := web.UsuarioEID(c, "id")
	if !ok {
		return
	}
	var body struct {
		Motivo string `json:"motivo"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "informe o motivo"})
		return
	}
	err := ct.uc.Rejeitar(c.Request.Context(), uid, pid, body.Motivo)
	web.Responder(c, http.StatusNoContent, nil, err)
}

func (ct *Controller) DefinirDestaque(c *gin.Context) {
	var body struct {
		ProjetoID int `json:"projetoId"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.ProjetoID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "informe projetoId"})
		return
	}
	err := ct.uc.DefinirDestaque(c.Request.Context(), body.ProjetoID)
	web.Responder(c, http.StatusNoContent, nil, err)
}
