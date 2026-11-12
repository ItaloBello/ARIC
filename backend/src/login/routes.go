package login

import "github.com/gin-gonic/gin"

// SetupLoginRoutes registra:
//
//	POST   /auth/registro     (público)
//	POST   /auth/login        (público)
//	GET    /auth/me           dados do usuário logado
//	PUT    /auth/me           nome e bio
//	PUT    /auth/me/foto      troca a foto (multipart, campo "foto")
//	DELETE /auth/me/foto      remove a foto
//	PUT    /auth/me/endereco  CEP, número e complemento (recalcula o bairro)
//	PUT    /auth/me/senha     troca a senha (exige a atual)
func SetupLoginRoutes(r *gin.RouterGroup, ct *LoginController, auth gin.HandlerFunc) {
	g := r.Group("/auth")
	g.POST("/registro", ct.Registro)
	g.POST("/login", ct.Login)

	me := g.Group("/me", auth)
	me.GET("", ct.Me)
	me.PUT("", ct.AtualizarPerfil)
	me.PUT("/foto", ct.AtualizarFoto)
	me.DELETE("/foto", ct.RemoverFoto)
	me.PUT("/endereco", ct.AtualizarEndereco)
	me.PUT("/senha", ct.TrocarSenha)
}
