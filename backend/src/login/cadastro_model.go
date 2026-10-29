package login

import (
	"ARIC/src/usuario"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"strings"
)

var viaCepURL = "https://viacep.com.br/ws/"



type Cadastro struct {
	Nome                string `form:"nome"`
	Email               string `form:"email"`
	Senha               string `form:"senha"`
	ConfirmarSenha      string `form:"confirmarSenha"`
	Foto                *string
	Bio                 *string `form:"bio"`
	Cep                 *string `form:"cep"`
	EstadoID            int     `json:"estado_id"`
	CidadeID            int     `json:"cidade_id"`
	Bairro              *string `form:"bairro"`
	Logradouro          *string `form:"logradouro"`
	NumeroEndereco      *string `form:"numeroEndereco"`
	ComplementoEndereco *string `form:"complementoEndereco"`
}

type viaCepResponse struct {
	Cep        string `json:"cep"`
	Logradouro string `json:"logradouro"`
	Bairro     string `json:"bairro"`
	Localidade string `json:"localidade"`
	UF         string `json:"uf"`
	Erro       bool   `json:"erro"`
}

func (c *Cadastro) ToUsuario() *usuario.Usuario {
	return &usuario.Usuario{
		Nome:                sql.NullString{String: c.Nome, Valid: true},
		Email:               sql.NullString{String: c.Email, Valid: true},
		SenhaHash:           sql.NullString{String: c.Senha, Valid: true},
		FotoURL:             sql.NullString{String: *c.Foto, Valid: true},
		Bio:                 sql.NullString{String: *c.Bio, Valid: true},
		Cep:                 sql.NullString{String: *c.Cep, Valid: true},
		Logradouro:          sql.NullString{String: *c.Logradouro, Valid: true},
		NumeroEndereco:      sql.NullString{String: *c.NumeroEndereco, Valid: true},
		ComplementoEndereco: sql.NullString{String: *c.ComplementoEndereco, Valid: true},
	}
}

func (c *Cadastro) NormalizarBairro() string {
	// Normaliza o nome do bairro para facilitar buscas e comparações
	normalizado := *c.Bairro

	// Converte para minúsculas
	normalizado = strings.ToLower(normalizado)

	// Remove espaços extras
	normalizado = strings.TrimSpace(normalizado)

	//remover espaços duplicados
	normalizado = strings.Join(strings.Fields(normalizado), " ")

	// Remove acentos e caracteres especiais
	normalizado = strings.Map(func(r rune) rune {
		switch r {
		case 'á', 'à', 'ã', 'â', 'ä':
			return 'a'
		case 'é', 'è', 'ê', 'ë':
			return 'e'
		case 'í', 'ì', 'î', 'ï':
			return 'i'
		case 'ó', 'ò', 'õ', 'ô', 'ö':
			return 'o'
		case 'ú', 'ù', 'û', 'ü':
			return 'u'
		case 'ç':
			return 'c'
		default:
			return r
		}
	}, normalizado)

	return normalizado
}

func (c *Cadastro) SenhaValida() (bool, error) {
	if c.Senha != c.ConfirmarSenha {
		return false, fmt.Errorf("as senhas não coincidem")
	}
	if len(c.Senha) < 8 {
		return false, fmt.Errorf("a senha deve ter pelo menos 8 caracteres")
	}
	if !hasUppercase(c.Senha) {
		return false, fmt.Errorf("a senha deve conter pelo menos uma letra maiúscula")
	}
	if !hasLowercase(c.Senha) {
		return false, fmt.Errorf("a senha deve conter pelo menos uma letra minúscula")
	}
	if !hasDigit(c.Senha) {
		return false, fmt.Errorf("a senha deve conter pelo menos um dígito")
	}
	if !hasSpecialChar(c.Senha) {
		return false, fmt.Errorf("a senha deve conter pelo menos um caractere especial")
	}
	return true, nil
}

func (c *Cadastro) VerificarCep() error {
	// Se informou CEP, usa o CEP como fonte principal.
	if c.Cep != nil && strings.TrimSpace(*c.Cep) != "" {
		return c.buscarPorCep()
	} else {
		return errors.New("CEP não informado")
	}
}
func (c *Cadastro) buscarPorCep() error {
	cep := strings.TrimSpace(*c.Cep)
	cep = strings.ReplaceAll(cep, "-", "")
	cep = strings.ReplaceAll(cep, ".", "")

	if len(cep) != 8 {
		return errors.New("CEP inválido")
	}

	httpResponse, err := http.Get(viaCepURL + cep + "/json/")
	if err != nil {
		return fmt.Errorf("erro ao consultar o CEP: %w", err)
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode != http.StatusOK {
		return errors.New("erro ao consultar o CEP")
	}

	var resposta viaCepResponse

	if err := json.NewDecoder(httpResponse.Body).Decode(&resposta); err != nil {
		return fmt.Errorf("erro ao interpretar resposta do ViaCEP: %w", err)
	}

	if resposta.Erro {
		return errors.New("CEP inválido")
	}

	// Normaliza o CEP retornado pela API.
	cep = strings.ReplaceAll(resposta.Cep, "-", "")

	// Preenche o cadastro.
	c.Cep = &cep
	c.Bairro = &resposta.Bairro
	c.Logradouro = &resposta.Logradouro

	return nil
}


func hasUppercase(s string) bool {
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			return true
		}
	}
	return false
}
func hasLowercase(s string) bool {
	for _, r := range s {
		if r >= 'a' && r <= 'z' {
			return true
		}
	}
	return false
}
func hasDigit(s string) bool {
	for _, r := range s {
		if r >= '0' && r <= '9' {
			return true
		}
	}
	return false
}
func hasSpecialChar(s string) bool {
	specialChars := "!@#$%^&*()-_=+[]{}|;:'\",.<>?/`~"
	for _, r := range s {
		if strings.ContainsRune(specialChars, r) {
			return true
		}
	}
	return false
}
