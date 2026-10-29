package login

import (
	"ARIC/src/usuario"
	"database/sql"
)

type Login struct {
	Email string `json:"email"`
	Senha string `json:"senha"`
}

func (l *Login) ToUsuario() *usuario.Usuario {
	return &usuario.Usuario{
		Email:     sql.NullString{String: l.Email, Valid: true},
		SenhaHash: sql.NullString{String: l.Senha, Valid: true},
	}
}
