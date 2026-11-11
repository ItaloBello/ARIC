package login

import (
	"ARIC/src/cidade"
	"ARIC/src/usuario"
	"context"
	"errors"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

var ErrEmailDuplicado = errors.New("email duplicado")

type LoginRepository struct {
	db *sqlx.DB
}

func NewLoginRepository(db *sqlx.DB) *LoginRepository {
	return &LoginRepository{db: db}
}

const colunasUsuario = `id, nome, email, senha_hash, foto_url, bio, cep, bairro_id, logradouro,
	numero_endereco, complemento_endereco, xp_total, nivel_id, sequencia_dias, papel,
	criado_em, atualizado_em, deletado_em`

// GetCidadeByNomeUF busca a cidade retornada pelo ViaCEP (localidade + uf).
func (r *LoginRepository) GetCidadeByNomeUF(ctx context.Context, nome, uf string) (*cidade.Cidade, error) {
	const query = `
		SELECT c.id, c.nome, c.estado_id
		FROM aric.cidades c
		JOIN aric.estados e ON e.id = c.estado_id
		WHERE lower(c.nome) = lower($1) AND e.sigla = $2
		LIMIT 1`

	var c cidade.Cidade
	if err := r.db.GetContext(ctx, &c, query, nome, uf); err != nil {
		return nil, err
	}
	return &c, nil
}

// UpsertBairro cria o bairro se ainda não existir e devolve o id (novo ou existente).
func (r *LoginRepository) UpsertBairro(ctx context.Context, nome, nomeNormalizado string, cidadeID int) (int, error) {
	const query = `
		INSERT INTO aric.bairros (nome, nome_normalizado, cidade_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (nome_normalizado, cidade_id)
		DO UPDATE SET nome = bairros.nome
		RETURNING id`

	var id int
	err := r.db.GetContext(ctx, &id, query, nome, nomeNormalizado, cidadeID)
	return id, err
}

func (r *LoginRepository) GetNivelInicialID(ctx context.Context) (int, error) {
	const query = `SELECT id FROM aric.niveis ORDER BY minimo_xp ASC, numero ASC LIMIT 1`
	var id int
	err := r.db.GetContext(ctx, &id, query)
	return id, err
}

func (r *LoginRepository) EmailExiste(ctx context.Context, email string) (bool, error) {
	const query = `SELECT EXISTS (SELECT 1 FROM aric.usuarios WHERE email = $1)`
	var existe bool
	err := r.db.GetContext(ctx, &existe, query, email)
	return existe, err
}

func (r *LoginRepository) CreateUsuario(ctx context.Context, u *usuario.Usuario) (*usuario.Usuario, error) {
	query := `
		INSERT INTO aric.usuarios
			(nome, email, senha_hash, foto_url, bio, cep, bairro_id, logradouro,
			 numero_endereco, complemento_endereco, nivel_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING ` + colunasUsuario

	var criado usuario.Usuario
	err := r.db.GetContext(ctx, &criado, query,
		u.Nome, u.Email, u.SenhaHash, u.FotoURL, u.Bio, u.Cep, u.BairroID, u.Logradouro,
		u.NumeroEndereco, u.ComplementoEndereco, u.NivelID)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" && strings.Contains(pqErr.Constraint, "email") {
			return nil, ErrEmailDuplicado
		}
		return nil, err
	}
	return &criado, nil
}

func (r *LoginRepository) GetUsuarioByEmail(ctx context.Context, email string) (*usuario.Usuario, error) {
	query := `SELECT ` + colunasUsuario + ` FROM aric.usuarios WHERE email = $1 AND deletado_em IS NULL`
	var u usuario.Usuario
	if err := r.db.GetContext(ctx, &u, query, email); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *LoginRepository) GetUsuarioByID(ctx context.Context, id int) (*usuario.Usuario, error) {
	query := `SELECT ` + colunasUsuario + ` FROM aric.usuarios WHERE id = $1 AND deletado_em IS NULL`
	var u usuario.Usuario
	if err := r.db.GetContext(ctx, &u, query, id); err != nil {
		return nil, err
	}
	return &u, nil
}
