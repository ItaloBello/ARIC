package login

import "github.com/gin-gonic/gin"

// SetupLoginRoutes registra:
//
//	POST /auth/registro  (público)
//	POST /auth/login     (público)
//	GET  /auth/me        (exige token)
func SetupLoginRoutes(r *gin.RouterGroup, ct *LoginController, auth gin.HandlerFunc) {
	g := r.Group("/auth")
	g.POST("/registro", ct.Registro)
	g.POST("/login", ct.Login)
	g.GET("/me", auth, ct.Me)
}
