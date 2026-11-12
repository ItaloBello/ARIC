package middleware

import (
	"ARIC/utils/token"
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const usuarioIDKey = "usuarioID"

// SessaoValida diz se o token emitido em "emitidoEm" ainda vale para o usuário:
// a conta existe, não foi excluída e a senha não foi trocada depois da emissão.
type SessaoValida func(ctx context.Context, usuarioID int, emitidoEm time.Time) (bool, error)

// Auth exige o header "Authorization: Bearer <token>" e coloca o id do usuário no contexto.
// Além da assinatura, confere no banco se a sessão ainda vale (uma consulta por chave primária).
func Auth(tm *token.Manager, sessaoValida SessaoValida) gin.HandlerFunc {
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

		var emitidoEm time.Time
		if claims.IssuedAt != nil {
			emitidoEm = claims.IssuedAt.Time
		}
		valida, err := sessaoValida(c.Request.Context(), claims.UsuarioID, emitidoEm)
		if err != nil {
			log.Println("[auth] erro ao verificar sessão:", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "erro interno"})
			return
		}
		if !valida {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "sessão encerrada, faça login novamente"})
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
