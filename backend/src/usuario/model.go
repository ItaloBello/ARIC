package usuario

import (
	hashsenha "ARIC/utils/hash_senha"
	"database/sql"
	"fmt"
	"os"
)

var viaCepURL = os.Getenv("VIACEP_URL")

type Usuario struct {
	ID                  int            `json:"id" db:"id"`
	Nome                sql.NullString `json:"nome" db:"nome"`
	Email               sql.NullString `json:"email" db:"email"`
	SenhaHash           sql.NullString `db:"senha_hash"`
	FotoURL             sql.NullString `json:"fotoUrl" db:"foto_url"`
	Bio                 sql.NullString `json:"bio" db:"bio"`
	Cep                 sql.NullString `json:"cep" db:"cep"`
	BairroID            sql.NullInt64  `json:"bairroId" db:"bairro_id"`
	Logradouro          sql.NullString `json:"logradouro" db:"logradouro"`
	NumeroEndereco      sql.NullString `json:"numeroEndereco" db:"numero_endereco"`
	ComplementoEndereco sql.NullString `json:"complementoEndereco" db:"complemento_endereco"`
	XpTotal             sql.NullInt64  `json:"xpTotal" db:"xp_total"`
	NivelId             sql.NullInt64  `json:"nivelId" db:"nivel_id"`
	SequenciaDias       int            `json:"sequenciaDias" db:"sequencia_dias"`
	CriadoEm            sql.NullTime   `json:"criadoEm" db:"criado_em"`
	AtualizadoEm        sql.NullTime   `json:"atualizadoEm" db:"atualizado_em"`
	DeletadoEm          sql.NullTime   `json:"deletadoEm" db:"deletado_em"`
}

func (u *Usuario) TableName() string {
	return "usuario"
}

func (u *Usuario) HashSenha(senha string) {
	hash, err := hashsenha.Encriptar(senha)
	if err != nil {
		fmt.Println("[Usuario - HashSenha] Erro ao encriptar a senha:", err)
	}
	u.SenhaHash = sql.NullString{String: hash, Valid: true}
}
