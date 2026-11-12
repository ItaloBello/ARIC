package login

import (
	"ARIC/utils/viacep"
	"errors"
	"net/mail"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Cadastro aceita multipart/form-data (com foto) ou JSON (sem foto).
// Estado e cidade não são enviados: vêm do CEP.
// Bairro e logradouro só são usados quando o ViaCEP não os retorna (CEP único da cidade).
type Cadastro struct {
	Nome                string `form:"nome" json:"nome"`
	Email               string `form:"email" json:"email"`
	Senha               string `form:"senha" json:"senha"`
	ConfirmarSenha      string `form:"confirmarSenha" json:"confirmarSenha"`
	Bio                 string `form:"bio" json:"bio"`
	Cep                 string `form:"cep" json:"cep"`
	Bairro              string `form:"bairro" json:"bairro"`
	Logradouro          string `form:"logradouro" json:"logradouro"`
	NumeroEndereco      string `form:"numeroEndereco" json:"numeroEndereco"`
	ComplementoEndereco string `form:"complementoEndereco" json:"complementoEndereco"`
}

// Normalizar limpa espaços, deixa o e-mail em minúsculas e o CEP só com dígitos.
// A senha não é alterada.
func (c *Cadastro) Normalizar() {
	c.Nome = strings.Join(strings.Fields(c.Nome), " ")
	c.Email = strings.ToLower(strings.TrimSpace(c.Email))
	c.Bio = strings.TrimSpace(c.Bio)
	c.Cep = viacep.LimparCep(c.Cep)
	c.Bairro = strings.Join(strings.Fields(c.Bairro), " ")
	c.Logradouro = strings.Join(strings.Fields(c.Logradouro), " ")
	c.NumeroEndereco = strings.TrimSpace(c.NumeroEndereco)
	c.ComplementoEndereco = strings.TrimSpace(c.ComplementoEndereco)
}

// Validar checa os campos que não dependem do banco nem do ViaCEP.
func (c *Cadastro) Validar() error {
	if err := validarNome(c.Nome); err != nil {
		return err
	}
	if c.Email == "" {
		return errors.New("e-mail é obrigatório")
	}
	if !emailValido(c.Email) {
		return errors.New("e-mail inválido")
	}
	if err := validarEndereco(c.Cep, c.NumeroEndereco, c.ComplementoEndereco); err != nil {
		return err
	}
	if err := validarBio(c.Bio); err != nil {
		return err
	}
	return validarSenhaForte(c.Senha, c.ConfirmarSenha)
}

func validarNome(nome string) error {
	if nome == "" {
		return errors.New("nome é obrigatório")
	}
	if utf8.RuneCountInString(nome) > 255 {
		return errors.New("nome deve ter no máximo 255 caracteres")
	}
	return nil
}

func validarBio(bio string) error {
	if utf8.RuneCountInString(bio) > maxBio {
		return errors.New("a bio deve ter no máximo 500 caracteres")
	}
	return nil
}

func validarEndereco(cep, numero, complemento string) error {
	if len(cep) != 8 {
		return errors.New("CEP deve ter 8 dígitos")
	}
	if numero == "" {
		return errors.New("número do endereço é obrigatório")
	}
	if len(numero) > 20 {
		return errors.New("número do endereço deve ter no máximo 20 caracteres")
	}
	if utf8.RuneCountInString(complemento) > 255 {
		return errors.New("complemento deve ter no máximo 255 caracteres")
	}
	return nil
}

const maxBio = 500

// validarSenhaForte: 8 a 72 caracteres, com maiúscula, minúscula, número e caractere especial.
func validarSenhaForte(senha, confirmar string) error {
	if senha != confirmar {
		return errors.New("as senhas não coincidem")
	}
	if utf8.RuneCountInString(senha) < 8 {
		return errors.New("a senha deve ter pelo menos 8 caracteres")
	}
	// bcrypt ignora tudo depois de 72 bytes
	if len(senha) > 72 {
		return errors.New("a senha deve ter no máximo 72 caracteres")
	}

	var maiuscula, minuscula, digito, especial bool
	for _, r := range senha {
		switch {
		case unicode.IsUpper(r):
			maiuscula = true
		case unicode.IsLower(r):
			minuscula = true
		case unicode.IsDigit(r):
			digito = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			especial = true
		}
	}

	switch {
	case !maiuscula:
		return errors.New("a senha deve conter pelo menos uma letra maiúscula")
	case !minuscula:
		return errors.New("a senha deve conter pelo menos uma letra minúscula")
	case !digito:
		return errors.New("a senha deve conter pelo menos um dígito")
	case !especial:
		return errors.New("a senha deve conter pelo menos um caractere especial")
	}
	return nil
}

func emailValido(email string) bool {
	if len(email) > 255 {
		return false
	}
	addr, err := mail.ParseAddress(email)
	// ParseAddress aceita "Nome <email>", então exigimos que seja só o endereço
	return err == nil && addr.Address == email && strings.Contains(email[strings.LastIndex(email, "@"):], ".")
}
