package login

import (
	"ARIC/src/cidade"
	"ARIC/src/usuario"
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

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

// ---------- edição de perfil ----------

// ErrUsuarioNaoEncontrado: o UPDATE não achou o usuário (inexistente ou apagado).
var ErrUsuarioNaoEncontrado = errors.New("usuário não encontrado")

const qAtualizarPerfil = `
UPDATE aric.usuarios SET nome = $2, bio = $3, atualizado_em = now()
WHERE id = $1 AND deletado_em IS NULL`

// Devolve a foto antiga para o arquivo poder ser apagado.
// A CTE lê o valor de antes do UPDATE.
const qTrocarFoto = `
WITH antiga AS (
    SELECT id, foto_url FROM aric.usuarios
    WHERE id = $1 AND deletado_em IS NULL
    FOR UPDATE
)
UPDATE aric.usuarios u
SET foto_url = $2, atualizado_em = now()
FROM antiga
WHERE u.id = antiga.id
RETURNING antiga.foto_url`

const qAtualizarEndereco = `
UPDATE aric.usuarios
SET cep = $2, bairro_id = $3, logradouro = $4, numero_endereco = $5,
    complemento_endereco = $6, atualizado_em = now()
WHERE id = $1 AND deletado_em IS NULL`

// Troca a senha e, no mesmo comando, invalida os tokens emitidos antes de $3.
const qAtualizarSenha = `
UPDATE aric.usuarios SET senha_hash = $2, sessoes_validas_desde = $3, atualizado_em = now()
WHERE id = $1 AND deletado_em IS NULL`

func (r *LoginRepository) exec1(ctx context.Context, query string, args ...any) error {
	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrUsuarioNaoEncontrado
	}
	return nil
}

func (r *LoginRepository) AtualizarPerfil(ctx context.Context, id int, nome string, bio sql.NullString) error {
	return r.exec1(ctx, qAtualizarPerfil, id, nome, bio)
}

func (r *LoginRepository) TrocarFoto(ctx context.Context, id int, nova sql.NullString) (sql.NullString, error) {
	var antiga sql.NullString
	err := r.db.GetContext(ctx, &antiga, qTrocarFoto, id, nova)
	return antiga, err
}

func (r *LoginRepository) AtualizarEndereco(ctx context.Context, id int, cep string, bairroID int, logradouro, numero string, complemento sql.NullString) error {
	return r.exec1(ctx, qAtualizarEndereco, id, cep, bairroID, logradouro, numero, complemento)
}

func (r *LoginRepository) AtualizarSenha(ctx context.Context, id int, hash string, sessoesValidasDesde time.Time) error {
	return r.exec1(ctx, qAtualizarSenha, id, hash, sessoesValidasDesde)
}
