// Package web reúne helpers de HTTP usados por todos os módulos.
package web

import (
	"database/sql"
	"net/http"
	"strconv"

	"ARIC/src/middleware"
	"ARIC/utils/apperror"

	"github.com/gin-gonic/gin"
)

const (
	LimitePadrao = 20
	LimiteMaximo = 50
)

// Pagina é o formato de toda lista paginada: {"itens", "total", "pagina", "limite"}.
type Pagina[T any] struct {
	Itens  []T `json:"itens"`
	Total  int `json:"total"`
	Pagina int `json:"pagina"`
	Limite int `json:"limite"`
}

// Paginacao aplica padrão (página 1, 20 itens) e limite máximo (50).
func Paginacao(pagina, limite int) (int, int) {
	if pagina < 1 {
		pagina = 1
	}
	if limite < 1 {
		limite = LimitePadrao
	}
	if limite > LimiteMaximo {
		limite = LimiteMaximo
	}
	return pagina, limite
}

// Usuario devolve o id do usuário logado ou responde 401.
func Usuario(c *gin.Context) (int, bool) {
	id, ok := middleware.UsuarioID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
	}
	return id, ok
}

// ParamID lê um id positivo da rota ou responde 400.
func ParamID(c *gin.Context, nome string) (int, bool) {
	id, err := strconv.Atoi(c.Param(nome))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": nome + " inválido"})
		return 0, false
	}
	return id, true
}

// UsuarioEID lê o usuário logado e o id da rota.
func UsuarioEID(c *gin.Context, nome string) (int, int, bool) {
	uid, ok := Usuario(c)
	if !ok {
		return 0, 0, false
	}
	id, ok := ParamID(c, nome)
	return uid, id, ok
}

func QueryInt(c *gin.Context, nome string) int {
	v, _ := strconv.Atoi(c.Query(nome))
	return v
}

// Responder envia dados com o status, ou o erro. Sem dados, só o status (ex.: 204).
func Responder(c *gin.Context, status int, dados any, err error) {
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

func StrPtr(s sql.NullString) *string {
	if !s.Valid {
		return nil
	}
	return &s.String
}
