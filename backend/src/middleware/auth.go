package middleware

import (
	"ARIC/utils/token"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const usuarioIDKey = "usuarioID"

// Auth exige o header "Authorization: Bearer <token>" e coloca o id do usuário no contexto.
func Auth(tm *token.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		tipo, tokenStr, ok := strings.Cut(header, " ")
		if !ok || !strings.EqualFold(tipo, "Bearer") || strings.TrimSpace(tokenStr) == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token não informado"})
			return
		}

		claims, err := tm.Validar(strings.TrimSpace(tokenStr))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token inválido ou expirado"})
			return
		}

		c.Set(usuarioIDKey, claims.UsuarioID)
		c.Next()
	}
}

// UsuarioID retorna o id do usuário autenticado (só funciona em rotas com o middleware Auth).
func UsuarioID(c *gin.Context) (int, bool) {
	v, ok := c.Get(usuarioIDKey)
	if !ok {
		return 0, false
	}
	id, ok := v.(int)
	return id, ok
}
