package login

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

var viaCepURL = "https://viacep.com.br/ws/"

type Login struct {
	Email string `json:"email"`
	Senha string `json:"senha"`
}

type Cadastro struct {
	Nome                string `form:"nome"`
	Email               string `form:"email"`
	Senha               string `form:"senha"`
	ConfirmarSenha      string `form:"confirmarSenha"`
	Foto                *string
	Bio                 *string `form:"bio"`
	Cep                 *string `form:"cep"`
	Estado              *string `form:"estado"`
	Cidade              *string `form:"cidade"`
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
	}

	// Sem CEP, tenta encontrar pelo endereço.
	return c.buscarPorEndereco()
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
	c.Estado = &resposta.UF
	c.Cidade = &resposta.Localidade
	c.Bairro = &resposta.Bairro
	c.Logradouro = &resposta.Logradouro

	return nil
}

func (c *Cadastro) buscarPorEndereco() error {
	if c.Estado == nil || strings.TrimSpace(*c.Estado) == "" {
		return errors.New("estado não informado")
	}

	if c.Cidade == nil || strings.TrimSpace(*c.Cidade) == "" {
		return errors.New("cidade não informada")
	}

	if c.Logradouro == nil || strings.TrimSpace(*c.Logradouro) == "" {
		return errors.New("logradouro não informado")
	}

	uf := strings.TrimSpace(*c.Estado)
	cidade := strings.TrimSpace(*c.Cidade)
	logradouro := strings.TrimSpace(*c.Logradouro)

	// URL-encoda cada parte do endereço.
	urlConsulta := viaCepURL +
		url.PathEscape(uf) + "/" +
		url.PathEscape(cidade) + "/" +
		url.PathEscape(logradouro) +
		"/json/"

	httpResponse, err := http.Get(urlConsulta)
	if err != nil {
		return fmt.Errorf("erro ao consultar o endereço: %w", err)
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode != http.StatusOK {
		return errors.New("erro ao consultar o endereço")
	}

	var respostas []viaCepResponse

	if err := json.NewDecoder(httpResponse.Body).Decode(&respostas); err != nil {
		return fmt.Errorf("erro ao interpretar resposta do ViaCEP: %w", err)
	}

	if len(respostas) == 0 {
		return errors.New("endereço não localizado")
	}

	// Usa o primeiro resultado.
	resposta := respostas[0]

	cep := strings.ReplaceAll(resposta.Cep, "-", "")

	c.Cep = &cep
	c.Estado = &resposta.UF
	c.Cidade = &resposta.Localidade
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
