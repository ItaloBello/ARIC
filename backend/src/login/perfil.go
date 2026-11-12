package login

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"mime/multipart"
	"strings"
	"time"

	"ARIC/src/acesso"
	"ARIC/src/usuario"
	"ARIC/utils/apperror"
	"ARIC/utils/uploader"
	"ARIC/utils/viacep"
)

// ---------- entrada ----------

// PerfilRequest: nome e bio. Bio vazia apaga a bio.
type PerfilRequest struct {
	Nome string `json:"nome"`
	Bio  string `json:"bio"`
}

// EnderecoRequest: o bairro sai do CEP. Bairro e logradouro só são usados quando o
// ViaCEP não os informa (CEP único da cidade), igual ao cadastro.
type EnderecoRequest struct {
	Cep                 string `json:"cep"`
	Bairro              string `json:"bairro"`
	Logradouro          string `json:"logradouro"`
	NumeroEndereco      string `json:"numeroEndereco"`
	ComplementoEndereco string `json:"complementoEndereco"`
}

type SenhaRequest struct {
	SenhaAtual     string `json:"senhaAtual"`
	NovaSenha      string `json:"novaSenha"`
	ConfirmarSenha string `json:"confirmarSenha"`
}

// ---------- regras ----------

func (u *LoginUseCases) AtualizarPerfil(ctx context.Context, id int, req PerfilRequest) (*usuario.UsuarioResponse, error) {
	nome := strings.Join(strings.Fields(req.Nome), " ")
	bio := strings.TrimSpace(req.Bio)
	if err := validarNome(nome); err != nil {
		return nil, apperror.BadRequest(err.Error())
	}
	if err := validarBio(bio); err != nil {
		return nil, apperror.BadRequest(err.Error())
	}

	if err := u.repo.AtualizarPerfil(ctx, id, nome, usuario.NullString(bio)); err != nil {
		return nil, erroUsuario("erro ao atualizar perfil", err)
	}
	return u.Me(ctx, id)
}

// AtualizarFoto troca a foto e apaga o arquivo antigo.
func (u *LoginUseCases) AtualizarFoto(ctx context.Context, id int, foto *multipart.FileHeader) (*usuario.UsuarioResponse, error) {
	res, err := uploader.SaveFile(foto, configFotoUsuario)
	if err != nil {
		return nil, apperror.BadRequest(err.Error())
	}

	antiga, err := u.repo.TrocarFoto(ctx, id, usuario.NullString(res.URL))
	if err != nil {
		// não deixa a foto nova órfã se o banco falhar
		apagarFoto(res.URL)
		return nil, erroUsuario("erro ao salvar foto", err)
	}
	apagarFoto(antiga.String)
	return u.Me(ctx, id)
}

func (u *LoginUseCases) RemoverFoto(ctx context.Context, id int) (*usuario.UsuarioResponse, error) {
	antiga, err := u.repo.TrocarFoto(ctx, id, sql.NullString{})
	if err != nil {
		return nil, erroUsuario("erro ao remover foto", err)
	}
	apagarFoto(antiga.String)
	return u.Me(ctx, id)
}

// AtualizarEndereco consulta o CEP novo e recalcula o bairro (criando-o se for novo).
func (u *LoginUseCases) AtualizarEndereco(ctx context.Context, id int, req EnderecoRequest) (*usuario.UsuarioResponse, error) {
	cep := viacep.LimparCep(req.Cep)
	numero := strings.TrimSpace(req.NumeroEndereco)
	complemento := strings.TrimSpace(req.ComplementoEndereco)
	if err := validarEndereco(cep, numero, complemento); err != nil {
		return nil, apperror.BadRequest(err.Error())
	}

	bairroID, logradouro, err := u.resolverEndereco(ctx, cep,
		strings.Join(strings.Fields(req.Bairro), " "), strings.Join(strings.Fields(req.Logradouro), " "))
	if err != nil {
		return nil, err
	}

	if err := u.repo.AtualizarEndereco(ctx, id, cep, bairroID, logradouro, numero, usuario.NullString(complemento)); err != nil {
		return nil, erroUsuario("erro ao atualizar endereço", err)
	}
	return u.Me(ctx, id)
}

// TrocarSenha exige a senha atual e aplica as mesmas regras do cadastro.
// Derruba as sessões dos outros aparelhos e devolve um token novo para este.
func (u *LoginUseCases) TrocarSenha(ctx context.Context, id int, req SenhaRequest, ip string) (*AuthResponse, error) {
	if req.SenhaAtual == "" {
		return nil, apperror.BadRequest("informe a senha atual")
	}
	if err := validarSenhaForte(req.NovaSenha, req.ConfirmarSenha); err != nil {
		return nil, apperror.BadRequest(err.Error())
	}

	user, err := u.repo.GetUsuarioByID(ctx, id)
	if err != nil {
		return nil, erroUsuario("erro ao buscar usuário", err)
	}
	if err := u.confirmarSenha(ctx, user, req.SenhaAtual, ip); err != nil {
		return nil, err
	}
	if user.VerificarSenha(req.NovaSenha) {
		return nil, apperror.BadRequest("a nova senha deve ser diferente da atual")
	}

	if err := user.DefinirSenha(req.NovaSenha); err != nil {
		return nil, interno("erro ao gerar hash da senha", err)
	}
	// tokens emitidos antes deste instante param de valer; o novo usa exatamente este instante
	agora := time.Now().Truncate(time.Second)
	if err := u.repo.AtualizarSenha(ctx, id, user.SenhaHash, agora); err != nil {
		return nil, erroUsuario("erro ao trocar senha", err)
	}
	tok, expira, err := u.tokens.GerarEm(id, agora)
	if err != nil {
		return nil, interno("erro ao gerar token", err)
	}
	return &AuthResponse{Token: tok, ExpiraEm: expira, Usuario: user.ToResponse()}, nil
}

// confirmarSenha confere a senha do próprio usuário (trocar senha, excluir conta)
// respeitando o mesmo limite de tentativas do login.
// Erro devolve 400, e não 401, para o app não entender como sessão expirada.
func (u *LoginUseCases) confirmarSenha(ctx context.Context, user *usuario.Usuario, senha, ip string) error {
	if err := u.limitador.VerificarSenha(ctx, user.Email, ip); err != nil {
		return err
	}
	if !user.VerificarSenha(senha) {
		u.limitador.RegistrarSenha(ctx, acesso.Senha, user.Email, ip, false)
		return apperror.BadRequest("senha atual incorreta")
	}
	u.limitador.RegistrarSenha(ctx, acesso.Senha, user.Email, ip, true)
	return nil
}

// ---------- helpers ----------

// apagarFoto remove do disco uma foto de usuário (só dentro de /uploads/usuarios).
func apagarFoto(url string) {
	if url == "" || uploader.URLLocal(url, configFotoUsuario) != nil {
		return
	}
	if err := uploader.DeleteFile(url); err != nil {
		log.Println("[login] erro ao apagar foto antiga:", err)
	}
}

// erroUsuario: usuário sumiu (apagado) vira 404; o resto vira 500.
func erroUsuario(msg string, err error) error {
	if errors.Is(err, sql.ErrNoRows) || errors.Is(err, ErrUsuarioNaoEncontrado) {
		return apperror.NotFound("usuário não encontrado")
	}
	return interno(msg, err)
}
