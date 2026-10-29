package login

import (
	"ARIC/src/usuario"
	"ARIC/utils/apperror"
	"database/sql"
)

type LoginUseCases struct {
	repo *LoginRepository
}

func NewLoginUseCases(repo *LoginRepository) *LoginUseCases {
	return &LoginUseCases{repo: repo}
}

func (u *LoginUseCases) Registro(c Cadastro) error {
	if !(c.Senha == c.ConfirmarSenha) {
		return apperror.BadRequest("senhas devem ser iguais")
	}
	senhaValida, err := c.SenhaValida()
	if err != nil {
		return apperror.BadRequest(err.Error())
	}
	if !senhaValida {
		return apperror.BadRequest("a senha é inválida")
	}

	if c.EstadoID <= 0 {
		return apperror.BadRequest("estado inválido")
	}
	if c.CidadeID <= 0 {
		return apperror.BadRequest("cidade inválida")
	}

	if c.Cep == nil || *c.Cep == "" {
		return apperror.BadRequest("cep não informado")
	}

	// Verifica se a cidade pertence ao estado selecionado
	cidade, err := u.repo.GetCidadeByIDEstadoID(c.CidadeID, c.EstadoID)
	if err != nil {
		if err == sql.ErrNoRows {
			return apperror.BadRequest("cidade não pertence ao estado selecionado")
		}
		return apperror.Internal("erro ao buscar cidade: " + err.Error())
	}

	err = c.VerificarCep()
	if err != nil {
		return apperror.BadRequest("erro ao verificar CEP: " + err.Error())
	}

	if c.Bairro == nil || *c.Bairro == "" {
		return apperror.BadRequest("bairro não encontrado para o CEP informado")
	}

	bairroNormalizado := c.NormalizarBairro()

	var bairroID int

	// Busca o bairro dentro da cidade
	bairro, err := u.repo.GetBairroByNomeCidade(bairroNormalizado, cidade.ID)
	if err == sql.ErrNoRows {
		// Bairro ainda não existe
		bairroID, err = u.repo.CreateBairro(*c.Bairro, bairroNormalizado, cidade.ID)
		if err != nil {
			return apperror.Internal("erro ao criar bairro: " + err.Error())
		}
	} else if err != nil {
		return apperror.Internal("erro ao buscar bairro: " + err.Error())
	} else {
		// Bairro já existe
		bairroID = bairro.ID
	}

	user := c.ToUsuario()
	user.HashSenha()
	user.BairroID = sql.NullInt64{Int64: int64(bairroID), Valid: true}

	_, err = u.repo.CreateUsuario(user)
	if err != nil {
		return apperror.Internal("erro ao criar usuario: " + err.Error())
	}

	return nil
}

func (u *LoginUseCases) Login(l Login) (*usuario.Usuario, error) {
	user := l.ToUsuario()
	usuarioDB, err := u.repo.GetUsuarioByEmail(user.Email.String)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperror.Unauthorized("email ou senha inválidos")
		}
		return nil, apperror.Internal("erro ao buscar usuário: " + err.Error())
	}
	if !user.VerificarSenha(usuarioDB.SenhaHash.String) {
		return nil, apperror.Unauthorized("email ou senha inválidos")
	}
	
	usuarioDB.SenhaHash = sql.NullString{} // Limpa a senha antes de retornar

	return usuarioDB, nil
}
