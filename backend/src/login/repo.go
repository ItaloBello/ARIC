package login

import (
	"ARIC/src/bairro"
	"ARIC/src/cidade"
	"ARIC/src/usuario"

	"github.com/jmoiron/sqlx"
)

type LoginRepository struct {
	db *sqlx.DB
}

func NewLoginRepository(db *sqlx.DB) *LoginRepository {
	return &LoginRepository{db: db}
}

func (r *LoginRepository) GetBairroByNomeCidade(nome string, cidadeID int) (*bairro.Bairro, error) {
	const query = `SELECT * FROM bairros WHERE nome_normalizado = $1 AND cidade_id = $2`

	var bairro bairro.Bairro
	err := r.db.Get(&bairro, query, nome, cidadeID)
	return &bairro, err
}

func (r *LoginRepository) GetCidadeByNome(nome string) (*cidade.Cidade, error) {
	const query = `SELECT * FROM ciades WHERE nome = $1`

	var cidade cidade.Cidade
	err := r.db.Get(&cidade, query, nome)
	return &cidade, err
}

func (r *LoginRepository) GetCidadeByIDEstadoID(cidadeID, estadoID int) (*cidade.Cidade, error) {
	const query = `SELECT * FROM cidades WHERE id = $1 AND estado_id = $2`

	var cidade cidade.Cidade
	err := r.db.Get(&cidade, query, cidadeID, estadoID)
	return &cidade, err
}

func (r *LoginRepository) CreateBairro(nome, nomeNormalizado string, cidadeID int) (int, error) {
	const query = `INSERT INTO bairros (nome, nome_normalizado, cidade_id) VALUES ($1, $2, $3) 
					ON CONFLICT (cidade_id, nome_normalizado)
					DO UPDATE SET nome = bairros.nome
					RETURNING id`

	var id int
	err := r.db.Get(&id, query, nome, nomeNormalizado, cidadeID)
	return id, err
}

func (r *LoginRepository) CreateUsuario(usuario *usuario.Usuario) (int, error) {
	const query = `INSERT INTO usuario (nome, email, senha_hash, foto_url, bio, cep, logradouro, numero_endereco, complemento_endereco) 
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`
	var id int
	err := r.db.Get(&id, query, usuario.Nome, usuario.Email, usuario.SenhaHash, usuario.FotoURL, usuario.Bio, usuario.Cep, usuario.Logradouro, usuario.NumeroEndereco, usuario.ComplementoEndereco)
	return id, err
}

func (r *LoginRepository) GetUsuarioByEmail(email string) (*usuario.Usuario, error) {
	const query = `SELECT * FROM usuario WHERE email = $1`
	var usuario usuario.Usuario
	err := r.db.Get(&usuario, query, email)
	return &usuario, err
}