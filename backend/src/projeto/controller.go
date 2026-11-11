package projeto

import (
	"net/http"
	"strconv"

	"ARIC/src/middleware"
	"ARIC/utils/apperror"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	uc *UseCases
}

func NewController(uc *UseCases) *Controller {
	return &Controller{uc: uc}
}

// ---------- helpers ----------

func usuario(c *gin.Context) (int, bool) {
	id, ok := middleware.UsuarioID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
	}
	return id, ok
}

func paramID(c *gin.Context, nome string) (int, bool) {
	id, err := strconv.Atoi(c.Param(nome))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": nome + " inválido"})
		return 0, false
	}
	return id, true
}

func queryInt(c *gin.Context, nome string) int {
	v, _ := strconv.Atoi(c.Query(nome))
	return v
}

// ctxUsuarioProjeto lê o usuário logado e o :id da rota.
func ctxUsuarioProjeto(c *gin.Context) (int, int, bool) {
	uid, ok := usuario(c)
	if !ok {
		return 0, 0, false
	}
	pid, ok := paramID(c, "id")
	return uid, pid, ok
}

func responder(c *gin.Context, status int, dados any, err error) {
	if err != nil {
		apperror.Responder(c, err)
		return
	}
	if dados == nil {
		c.Status(status)
		return
	}
	c.JSON(status, dados)
}

// ---------- catálogo ----------

func (ct *Controller) Categorias(c *gin.Context) {
	cats, err := ct.uc.Categorias(c.Request.Context())
	responder(c, http.StatusOK, cats, err)
}

func (ct *Controller) Dificuldades(c *gin.Context) {
	difs, err := ct.uc.Dificuldades(c.Request.Context())
	responder(c, http.StatusOK, difs, err)
}

// ---------- leitura ----------

func (ct *Controller) Listar(c *gin.Context) {
	uid, ok := usuario(c)
	if !ok {
		return
	}
	var f Filtros
	if err := c.ShouldBindQuery(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "filtros inválidos"})
		return
	}
	pag, err := ct.uc.Listar(c.Request.Context(), uid, f)
	responder(c, http.StatusOK, pag, err)
}

func (ct *Controller) Meus(c *gin.Context) {
	uid, ok := usuario(c)
	if !ok {
		return
	}
	pag, err := ct.uc.Meus(c.Request.Context(), uid, queryInt(c, "pagina"), queryInt(c, "limite"))
	responder(c, http.StatusOK, pag, err)
}

func (ct *Controller) Salvos(c *gin.Context) {
	uid, ok := usuario(c)
	if !ok {
		return
	}
	pag, err := ct.uc.Salvos(c.Request.Context(), uid, queryInt(c, "pagina"), queryInt(c, "limite"))
	responder(c, http.StatusOK, pag, err)
}

// Destaque devolve 204 quando não há destaque na semana.
func (ct *Controller) Destaque(c *gin.Context) {
	uid, ok := usuario(c)
	if !ok {
		return
	}
	card, err := ct.uc.Destaque(c.Request.Context(), uid)
	if err == nil && card == nil {
		c.Status(http.StatusNoContent)
		return
	}
	responder(c, http.StatusOK, card, err)
}

func (ct *Controller) Detalhe(c *gin.Context) {
	uid, pid, ok := ctxUsuarioProjeto(c)
	if !ok {
		return
	}
	d, err := ct.uc.Detalhe(c.Request.Context(), uid, pid)
	responder(c, http.StatusOK, d, err)
}

// ---------- escrita ----------

func (ct *Controller) EnviarImagem(c *gin.Context) {
	if _, ok := usuario(c); !ok {
		return
	}
	arquivo, err := c.FormFile("imagem")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "envie o arquivo no campo 'imagem'"})
		return
	}
	url, err := ct.uc.EnviarImagem(arquivo)
	responder(c, http.StatusCreated, gin.H{"url": url}, err)
}

func (ct *Controller) Criar(c *gin.Context) {
	uid, ok := usuario(c)
	if !ok {
		return
	}
	var req ProjetoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados do projeto inválidos"})
		return
	}
	d, err := ct.uc.Criar(c.Request.Context(), uid, req)
	responder(c, http.StatusCreated, d, err)
}

func (ct *Controller) Editar(c *gin.Context) {
	uid, pid, ok := ctxUsuarioProjeto(c)
	if !ok {
		return
	}
	var req ProjetoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados do projeto inválidos"})
		return
	}
	d, err := ct.uc.Editar(c.Request.Context(), uid, pid, req)
	responder(c, http.StatusOK, d, err)
}

func (ct *Controller) Remover(c *gin.Context) {
	uid, pid, ok := ctxUsuarioProjeto(c)
	if !ok {
		return
	}
	err := ct.uc.Remover(c.Request.Context(), uid, pid)
	responder(c, http.StatusNoContent, nil, err)
}

// ---------- interações ----------

func (ct *Controller) Concluir(c *gin.Context) {
	uid, pid, ok := ctxUsuarioProjeto(c)
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
	responder(c, http.StatusCreated, resp, err)
}

func (ct *Controller) curtir(c *gin.Context, curtir bool) {
	uid, pid, ok := ctxUsuarioProjeto(c)
	if !ok {
		return
	}
	resp, err := ct.uc.Curtir(c.Request.Context(), uid, pid, curtir)
	responder(c, http.StatusOK, resp, err)
}

func (ct *Controller) Curtir(c *gin.Context)    { ct.curtir(c, true) }
func (ct *Controller) Descurtir(c *gin.Context) { ct.curtir(c, false) }

func (ct *Controller) salvar(c *gin.Context, salvar bool) {
	uid, pid, ok := ctxUsuarioProjeto(c)
	if !ok {
		return
	}
	err := ct.uc.Salvar(c.Request.Context(), uid, pid, salvar)
	responder(c, http.StatusOK, gin.H{"salvo": salvar}, err)
}

func (ct *Controller) Salvar(c *gin.Context)       { ct.salvar(c, true) }
func (ct *Controller) RemoverSalvo(c *gin.Context) { ct.salvar(c, false) }

func (ct *Controller) Avaliar(c *gin.Context) {
	uid, pid, ok := ctxUsuarioProjeto(c)
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
	responder(c, http.StatusOK, resp, err)
}

// ---------- comentários ----------

func (ct *Controller) ListarComentarios(c *gin.Context) {
	uid, pid, ok := ctxUsuarioProjeto(c)
	if !ok {
		return
	}
	pag, err := ct.uc.ListarComentarios(c.Request.Context(), uid, pid, queryInt(c, "pagina"), queryInt(c, "limite"))
	responder(c, http.StatusOK, pag, err)
}

func (ct *Controller) Comentar(c *gin.Context) {
	uid, pid, ok := ctxUsuarioProjeto(c)
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
	responder(c, http.StatusCreated, com, err)
}

func (ct *Controller) RemoverComentario(c *gin.Context) {
	uid, pid, ok := ctxUsuarioProjeto(c)
	if !ok {
		return
	}
	cid, ok := paramID(c, "comentarioId")
	if !ok {
		return
	}
	err := ct.uc.RemoverComentario(c.Request.Context(), uid, pid, cid)
	responder(c, http.StatusNoContent, nil, err)
}

// ---------- moderação ----------

func (ct *Controller) FilaModeracao(c *gin.Context) {
	uid, ok := usuario(c)
	if !ok {
		return
	}
	pag, err := ct.uc.FilaModeracao(c.Request.Context(), uid, queryInt(c, "pagina"), queryInt(c, "limite"))
	responder(c, http.StatusOK, pag, err)
}

func (ct *Controller) Aprovar(c *gin.Context) {
	uid, pid, ok := ctxUsuarioProjeto(c)
	if !ok {
		return
	}
	err := ct.uc.Aprovar(c.Request.Context(), uid, pid)
	responder(c, http.StatusNoContent, nil, err)
}

func (ct *Controller) Rejeitar(c *gin.Context) {
	uid, pid, ok := ctxUsuarioProjeto(c)
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
	responder(c, http.StatusNoContent, nil, err)
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
	responder(c, http.StatusNoContent, nil, err)
}
