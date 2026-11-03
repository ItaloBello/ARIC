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
	if c.Nome == "" {
		return errors.New("nome é obrigatório")
	}
	if utf8.RuneCountInString(c.Nome) > 255 {
		return errors.New("nome deve ter no máximo 255 caracteres")
	}
	if c.Email == "" {
		return errors.New("e-mail é obrigatório")
	}
	if !emailValido(c.Email) {
		return errors.New("e-mail inválido")
	}
	if len(c.Cep) != 8 {
		return errors.New("CEP deve ter 8 dígitos")
	}
	if c.NumeroEndereco == "" {
		return errors.New("número do endereço é obrigatório")
	}
	if len(c.NumeroEndereco) > 20 {
		return errors.New("número do endereço deve ter no máximo 20 caracteres")
	}
	return c.validarSenha()
}

func (c *Cadastro) validarSenha() error {
	if c.Senha != c.ConfirmarSenha {
		return errors.New("as senhas não coincidem")
	}
	if utf8.RuneCountInString(c.Senha) < 8 {
		return errors.New("a senha deve ter pelo menos 8 caracteres")
	}
	// bcrypt ignora tudo depois de 72 bytes
	if len(c.Senha) > 72 {
		return errors.New("a senha deve ter no máximo 72 caracteres")
	}

	var maiuscula, minuscula, digito, especial bool
	for _, r := range c.Senha {
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
