package login

import (
	"ARIC/utils/apperror"
	"ARIC/utils/uploader"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type LoginController struct {
	uc *LoginUseCases
}

func NewLoginController(uc *LoginUseCases) *LoginController {
	return &LoginController{uc: uc}
}

func (ct *LoginController) Registro(c *gin.Context) {
	var request Cadastro
	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erro ao processar os dados de cadastro"})
		return
	}

	file, err := c.FormFile("foto")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erro ao processar o arquivo de foto"})
		return
	}

	if file != nil {
		result, err := uploader.SaveFile(file, uploader.UploadConfig{
			DestFolder:   "./uploads/usuarios",
			BaseURL:      "/uploads/usuarios",
			MaxSize:      5 * 1024 * 1024, // 5MB
			AllowedTypes: []string{".jpg", ".jpeg", ".png"},
		})

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		request.Foto = &result.URL
	}

	err = ct.uc.Registro(request)
	if err != nil {
		var apperr *apperror.AppError
		if errors.As(err, &apperr) {
			c.IndentedJSON(apperr.StatusCode, gin.H{"error": apperr.Message})
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Usuário cadastrado com sucesso"})
}

func (ct *LoginController) Login(c *gin.Context) {
	var request Login
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erro ao processar os dados de login"})
		return
	}

	usuario, err := ct.uc.Login(request)
	if err != nil {
		var apperr *apperror.AppError
		if errors.As(err, &apperr) {
			c.IndentedJSON(apperr.StatusCode, gin.H{"error": apperr.Message})
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Login realizado com sucesso", "usuario": usuario})
}
