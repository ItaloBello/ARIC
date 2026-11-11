package middleware

import (
	"context"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
)

// BuscarPapel devolve o papel do usuário ('usuario', 'moderador' ou 'admin').
type BuscarPapel func(ctx context.Context, usuarioID int) (string, error)

// ExigirPapel deixa passar só usuários com um dos papéis informados.
// O papel é lido do banco a cada requisição (não vai no token), então
// rebaixar alguém vale na hora. Usar depois do middleware Auth.
func ExigirPapel(buscar BuscarPapel, papeis ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := UsuarioID(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
			return
		}
		papel, err := buscar(c.Request.Context(), id)
		if err != nil || !slices.Contains(papeis, papel) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "sem permissão"})
			return
		}
		c.Next()
	}
}
