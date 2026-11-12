package login

import (
	"ARIC/src/middleware"
	"ARIC/utils/apperror"
	"ARIC/utils/web"
	"errors"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type LoginController struct {
	uc *LoginUseCases
}

func NewLoginController(uc *LoginUseCases) *LoginController {
	return &LoginController{uc: uc}
}

// Registro aceita multipart/form-data (com campo opcional "foto") ou JSON.
func (ct *LoginController) Registro(c *gin.Context) {
	var request Cadastro
	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "erro ao processar os dados de cadastro"})
		return
	}

	var foto *multipart.FileHeader
	if strings.HasPrefix(c.ContentType(), "multipart/form-data") {
		f, err := c.FormFile("foto")
		switch {
		case errors.Is(err, http.ErrMissingFile):
			// foto é opcional
		case err != nil:
			c.JSON(http.StatusBadRequest, gin.H{"error": "erro ao processar o arquivo de foto"})
			return
		default:
			foto = f
		}
	}

	resp, err := ct.uc.Registro(c.Request.Context(), request, foto)
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (ct *LoginController) Login(c *gin.Context) {
	var request Login
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "erro ao processar os dados de login"})
		return
	}

	resp, err := ct.uc.Login(c.Request.Context(), request)
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// Me devolve o usuário dono do token.
func (ct *LoginController) Me(c *gin.Context) {
	id, ok := middleware.UsuarioID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return
	}

	resp, err := ct.uc.Me(c.Request.Context(), id)
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func responderErro(c *gin.Context, err error) {
	apperror.Responder(c, err)
}

// ---------- edição de perfil ----------

func (ct *LoginController) AtualizarPerfil(c *gin.Context) {
	id, ok := web.Usuario(c)
	if !ok {
		return
	}
	var req PerfilRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados do perfil inválidos"})
		return
	}
	resp, err := ct.uc.AtualizarPerfil(c.Request.Context(), id, req)
	web.Responder(c, http.StatusOK, resp, err)
}

// AtualizarFoto recebe multipart/form-data com o campo "foto".
func (ct *LoginController) AtualizarFoto(c *gin.Context) {
	id, ok := web.Usuario(c)
	if !ok {
		return
	}
	foto, err := c.FormFile("foto")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "envie o arquivo no campo 'foto'"})
		return
	}
	resp, err := ct.uc.AtualizarFoto(c.Request.Context(), id, foto)
	web.Responder(c, http.StatusOK, resp, err)
}

func (ct *LoginController) RemoverFoto(c *gin.Context) {
	id, ok := web.Usuario(c)
	if !ok {
		return
	}
	resp, err := ct.uc.RemoverFoto(c.Request.Context(), id)
	web.Responder(c, http.StatusOK, resp, err)
}

func (ct *LoginController) AtualizarEndereco(c *gin.Context) {
	id, ok := web.Usuario(c)
	if !ok {
		return
	}
	var req EnderecoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados do endereço inválidos"})
		return
	}
	resp, err := ct.uc.AtualizarEndereco(c.Request.Context(), id, req)
	web.Responder(c, http.StatusOK, resp, err)
}

func (ct *LoginController) TrocarSenha(c *gin.Context) {
	id, ok := web.Usuario(c)
	if !ok {
		return
	}
	var req SenhaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados inválidos"})
		return
	}
	err := ct.uc.TrocarSenha(c.Request.Context(), id, req)
	web.Responder(c, http.StatusNoContent, nil, err)
}
