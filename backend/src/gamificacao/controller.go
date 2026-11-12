package gamificacao

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

// ---------- perfil ----------

func (ct *Controller) MeuPerfil(c *gin.Context) {
	uid, ok := web.Usuario(c)
	if !ok {
		return
	}
	p, err := ct.uc.Perfil(c.Request.Context(), uid)
	web.Responder(c, http.StatusOK, p, err)
}

// PerfilPublico mostra o perfil de outro usuário (sem e-mail e endereço).
func (ct *Controller) PerfilPublico(c *gin.Context) {
	_, id, ok := web.UsuarioEID(c, "id")
	if !ok {
		return
	}
	p, err := ct.uc.Perfil(c.Request.Context(), id)
	web.Responder(c, http.StatusOK, p, err)
}

func (ct *Controller) Niveis(c *gin.Context) {
	n, err := ct.uc.Niveis(c.Request.Context())
	web.Responder(c, http.StatusOK, n, err)
}

func (ct *Controller) HistoricoXP(c *gin.Context) {
	uid, ok := web.Usuario(c)
	if !ok {
		return
	}
	pag, err := ct.uc.HistoricoXP(c.Request.Context(), uid, web.QueryInt(c, "pagina"), web.QueryInt(c, "limite"))
	web.Responder(c, http.StatusOK, pag, err)
}

// ---------- ligas ----------

func (ct *Controller) Ligas(c *gin.Context) {
	l, err := ct.uc.Ligas(c.Request.Context())
	web.Responder(c, http.StatusOK, l, err)
}

func (ct *Controller) Ranking(c *gin.Context) {
	uid, ok := web.Usuario(c)
	if !ok {
		return
	}
	r, err := ct.uc.Ranking(c.Request.Context(), uid)
	web.Responder(c, http.StatusOK, r, err)
}

// ---------- missões ----------

func (ct *Controller) Missoes(c *gin.Context) {
	uid, ok := web.Usuario(c)
	if !ok {
		return
	}
	m, err := ct.uc.MissoesDaSemana(c.Request.Context(), uid)
	web.Responder(c, http.StatusOK, m, err)
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

func (ct *Controller) EnviarEvidencia(c *gin.Context) {
	uid, missaoID, ok := web.UsuarioEID(c, "id")
	if !ok {
		return
	}
	var req EvidenciaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "informe fotoUrl e quantidade"})
		return
	}
	ev, err := ct.uc.EnviarEvidencia(c.Request.Context(), uid, missaoID, req)
	web.Responder(c, http.StatusCreated, ev, err)
}

func (ct *Controller) MinhasEvidencias(c *gin.Context) {
	uid, ok := web.Usuario(c)
	if !ok {
		return
	}
	pag, err := ct.uc.MinhasEvidencias(c.Request.Context(), uid, web.QueryInt(c, "pagina"), web.QueryInt(c, "limite"))
	web.Responder(c, http.StatusOK, pag, err)
}

// ---------- moderação ----------

func (ct *Controller) EvidenciasPendentes(c *gin.Context) {
	pag, err := ct.uc.EvidenciasPendentes(c.Request.Context(), web.QueryInt(c, "pagina"), web.QueryInt(c, "limite"))
	web.Responder(c, http.StatusOK, pag, err)
}

func (ct *Controller) AprovarEvidencia(c *gin.Context) {
	uid, id, ok := web.UsuarioEID(c, "id")
	if !ok {
		return
	}
	resp, err := ct.uc.AprovarEvidencia(c.Request.Context(), uid, id)
	web.Responder(c, http.StatusOK, resp, err)
}

func (ct *Controller) RejeitarEvidencia(c *gin.Context) {
	uid, id, ok := web.UsuarioEID(c, "id")
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
	err := ct.uc.RejeitarEvidencia(c.Request.Context(), uid, id, body.Motivo)
	web.Responder(c, http.StatusNoContent, nil, err)
}
