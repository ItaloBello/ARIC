package login

import (
	"ARIC/src/usuario"
	"time"
)

type Login struct {
	Email string `json:"email"`
	Senha string `json:"senha"`
}

// AuthResponse é devolvido no login e no registro (o usuário já sai logado).
type AuthResponse struct {
	Token    string                  `json:"token"`
	ExpiraEm time.Time               `json:"expiraEm"`
	Usuario  usuario.UsuarioResponse `json:"usuario"`
}
