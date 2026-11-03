package login

import (
	"ARIC/src/bairro"
	"ARIC/src/usuario"
	"ARIC/utils/apperror"
	hashsenha "ARIC/utils/hash_senha"
	"ARIC/utils/token"
	"ARIC/utils/uploader"
	"ARIC/utils/viacep"
	"context"
	"database/sql"
	"errors"
	"log"
	"mime/multipart"
	"strings"
)

var configFotoUsuario = uploader.UploadConfig{
	DestFolder:   "./uploads/usuarios",
	BaseURL:      "/uploads/usuarios",
	MaxSize:      5 * 1024 * 1024, // 5MB
	AllowedTypes: []string{".jpg", ".jpeg", ".png"},
}

// Hash usado quando o e-mail não existe, para o login levar o mesmo tempo
// e não revelar quais e-mails estão cadastrados.
var hashFalso, _ = hashsenha.Encriptar("senha-falsa-para-tempo-constante")

type LoginUseCases struct {
	repo   *LoginRepository
	cep    *viacep.Client
	tokens *token.Manager
}

func NewLoginUseCases(repo *LoginRepository, cep *viacep.Client, tokens *token.Manager) *LoginUseCases {
	return &LoginUseCases{repo: repo, cep: cep, tokens: tokens}
}

// Registro valida tudo, só então salva a foto (se houver) e cria o usuário.
// Devolve o usuário já autenticado.
func (u *LoginUseCases) Registro(ctx context.Context, c Cadastro, foto *multipart.FileHeader) (*AuthResponse, error) {
	c.Normalizar()
	if err := c.Validar(); err != nil {
		return nil, apperror.BadRequest(err.Error())
	}

	existe, err := u.repo.EmailExiste(ctx, c.Email)
	if err != nil {
		return nil, interno("erro ao verificar e-mail", err)
	}
	if existe {
		return nil, apperror.Conflict("e-mail já cadastrado")
	}

	// Endereço a partir do CEP
	end, err := u.cep.Buscar(ctx, c.Cep)
	switch {
	case errors.Is(err, viacep.ErrCepInvalido), errors.Is(err, viacep.ErrCepNaoEncontrado):
		return nil, apperror.BadRequest(err.Error())
	case err != nil:
		log.Println("[Registro] ViaCEP:", err)
		return nil, apperror.BadGateway("não foi possível consultar o CEP agora, tente novamente")
	}

	cid, err := u.repo.GetCidadeByNomeUF(ctx, end.Localidade, end.UF)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.BadRequest("cidade do CEP não encontrada: " + end.Localidade + "/" + end.UF)
	}
	if err != nil {
		return nil, interno("erro ao buscar cidade", err)
	}

	// CEP único de cidade pequena vem sem bairro/logradouro: usa o que o usuário digitou
	nomeBairro := primeiroNaoVazio(end.Bairro, c.Bairro)
	if nomeBairro == "" {
		return nil, apperror.BadRequest("este CEP não informa o bairro, preencha o campo bairro")
	}
	logradouro := primeiroNaoVazio(end.Logradouro, c.Logradouro)
	if logradouro == "" {
		return nil, apperror.BadRequest("este CEP não informa a rua, preencha o campo logradouro")
	}

	bairroID, err := u.repo.UpsertBairro(ctx, nomeBairro, bairro.Normalizar(nomeBairro), cid.ID)
	if err != nil {
		return nil, interno("erro ao salvar bairro", err)
	}

	nivelID, err := u.repo.GetNivelInicialID(ctx)
	if err != nil {
		return nil, interno("erro ao buscar nível inicial (a tabela aric.niveis está populada?)", err)
	}

	novo := &usuario.Usuario{
		Nome:                c.Nome,
		Email:               c.Email,
		Bio:                 usuario.NullString(c.Bio),
		Cep:                 c.Cep,
		BairroID:            bairroID,
		Logradouro:          logradouro,
		NumeroEndereco:      c.NumeroEndereco,
		ComplementoEndereco: usuario.NullString(c.ComplementoEndereco),
		NivelID:             sql.NullInt64{Int64: int64(nivelID), Valid: true},
	}
	if err := novo.DefinirSenha(c.Senha); err != nil {
		return nil, interno("erro ao gerar hash da senha", err)
	}

	// Foto por último, para não sobrar arquivo órfão quando a validação falha
	if foto != nil {
		res, err := uploader.SaveFile(foto, configFotoUsuario)
		if err != nil {
			return nil, apperror.BadRequest(err.Error())
		}
		novo.FotoURL = usuario.NullString(res.URL)
	}

	criado, err := u.repo.CreateUsuario(ctx, novo)
	if err != nil {
		if novo.FotoURL.Valid {
			if delErr := uploader.DeleteFile(novo.FotoURL.String); delErr != nil {
				log.Println("[Registro] erro ao remover foto órfã:", delErr)
			}
		}
		if errors.Is(err, ErrEmailDuplicado) {
			return nil, apperror.Conflict("e-mail já cadastrado")
		}
		return nil, interno("erro ao criar usuário", err)
	}

	return u.autenticar(criado)
}

func (u *LoginUseCases) Login(ctx context.Context, l Login) (*AuthResponse, error) {
	email := strings.ToLower(strings.TrimSpace(l.Email))
	if email == "" || l.Senha == "" {
		return nil, apperror.BadRequest("informe e-mail e senha")
	}

	user, err := u.repo.GetUsuarioByEmail(ctx, email)
	if errors.Is(err, sql.ErrNoRows) {
		_ = hashsenha.CompararSenha(hashFalso, l.Senha)
		return nil, apperror.Unauthorized("e-mail ou senha inválidos")
	}
	if err != nil {
		return nil, interno("erro ao buscar usuário", err)
	}

	if !user.VerificarSenha(l.Senha) {
		return nil, apperror.Unauthorized("e-mail ou senha inválidos")
	}

	return u.autenticar(user)
}

func (u *LoginUseCases) Me(ctx context.Context, id int) (*usuario.UsuarioResponse, error) {
	user, err := u.repo.GetUsuarioByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.NotFound("usuário não encontrado")
	}
	if err != nil {
		return nil, interno("erro ao buscar usuário", err)
	}
	resp := user.ToResponse()
	return &resp, nil
}

func (u *LoginUseCases) autenticar(user *usuario.Usuario) (*AuthResponse, error) {
	tok, expira, err := u.tokens.Gerar(user.ID)
	if err != nil {
		return nil, interno("erro ao gerar token", err)
	}
	return &AuthResponse{Token: tok, ExpiraEm: expira, Usuario: user.ToResponse()}, nil
}

// interno registra o erro real no log e devolve uma mensagem genérica ao cliente.
func interno(msg string, err error) error {
	log.Printf("[login] %s: %v", msg, err)
	return apperror.Internal(msg)
}

func primeiroNaoVazio(valores ...string) string {
	for _, v := range valores {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
