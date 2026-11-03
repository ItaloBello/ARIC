package usuario

import (
	hashsenha "ARIC/utils/hash_senha"
	"database/sql"
	"time"
)

type Usuario struct {
	ID                  int            `db:"id"`
	Nome                string         `db:"nome"`
	Email               string         `db:"email"`
	SenhaHash           string         `db:"senha_hash"`
	FotoURL             sql.NullString `db:"foto_url"`
	Bio                 sql.NullString `db:"bio"`
	Cep                 string         `db:"cep"`
	BairroID            int            `db:"bairro_id"`
	Logradouro          string         `db:"logradouro"`
	NumeroEndereco      string         `db:"numero_endereco"`
	ComplementoEndereco sql.NullString `db:"complemento_endereco"`
	XpTotal             int            `db:"xp_total"`
	NivelID             sql.NullInt64  `db:"nivel_id"`
	SequenciaDias       int            `db:"sequencia_dias"`
	CriadoEm            time.Time      `db:"criado_em"`
	AtualizadoEm        time.Time      `db:"atualizado_em"`
	DeletadoEm          sql.NullTime   `db:"deletado_em"`
}

// UsuarioResponse é o formato enviado ao front (sem senha e sem sql.Null*).
type UsuarioResponse struct {
	ID                  int       `json:"id"`
	Nome                string    `json:"nome"`
	Email               string    `json:"email"`
	FotoURL             *string   `json:"fotoUrl"`
	Bio                 *string   `json:"bio"`
	Cep                 string    `json:"cep"`
	BairroID            int       `json:"bairroId"`
	Logradouro          string    `json:"logradouro"`
	NumeroEndereco      string    `json:"numeroEndereco"`
	ComplementoEndereco *string   `json:"complementoEndereco"`
	XpTotal             int       `json:"xpTotal"`
	NivelID             *int64    `json:"nivelId"`
	SequenciaDias       int       `json:"sequenciaDias"`
	CriadoEm            time.Time `json:"criadoEm"`
}

func (u *Usuario) DefinirSenha(senha string) error {
	hash, err := hashsenha.Encriptar(senha)
	if err != nil {
		return err
	}
	u.SenhaHash = hash
	return nil
}

func (u *Usuario) VerificarSenha(senha string) bool {
	return hashsenha.CompararSenha(u.SenhaHash, senha) == nil
}

func (u *Usuario) ToResponse() UsuarioResponse {
	return UsuarioResponse{
		ID:                  u.ID,
		Nome:                u.Nome,
		Email:               u.Email,
		FotoURL:             strPtr(u.FotoURL),
		Bio:                 strPtr(u.Bio),
		Cep:                 u.Cep,
		BairroID:            u.BairroID,
		Logradouro:          u.Logradouro,
		NumeroEndereco:      u.NumeroEndereco,
		ComplementoEndereco: strPtr(u.ComplementoEndereco),
		XpTotal:             u.XpTotal,
		NivelID:             intPtr(u.NivelID),
		SequenciaDias:       u.SequenciaDias,
		CriadoEm:            u.CriadoEm,
	}
}

func strPtr(s sql.NullString) *string {
	if !s.Valid {
		return nil
	}
	return &s.String
}

func intPtr(i sql.NullInt64) *int64 {
	if !i.Valid {
		return nil
	}
	return &i.Int64
}

// NullString converte string vazia em NULL.
func NullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
