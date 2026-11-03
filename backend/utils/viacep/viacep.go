package viacep

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const urlPadrao = "https://viacep.com.br/ws/"

var (
	ErrCepInvalido      = errors.New("CEP inválido")
	ErrCepNaoEncontrado = errors.New("CEP não encontrado")
)

type Endereco struct {
	Cep        string `json:"cep"`
	Logradouro string `json:"logradouro"`
	Bairro     string `json:"bairro"`
	Localidade string `json:"localidade"`
	UF         string `json:"uf"`
	// O ViaCEP já retornou "erro": true e "erro": "true", então aceitamos os dois.
	Erro any `json:"erro"`
}

type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string) *Client {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		baseURL = urlPadrao
	}
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 5 * time.Second},
	}
}

// LimparCep mantém apenas os dígitos do CEP.
func LimparCep(cep string) string {
	var b strings.Builder
	for _, r := range cep {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func (c *Client) Buscar(ctx context.Context, cep string) (*Endereco, error) {
	cep = LimparCep(cep)
	if len(cep) != 8 {
		return nil, ErrCepInvalido
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+cep+"/json/", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao consultar o ViaCEP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest {
		return nil, ErrCepInvalido
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ViaCEP respondeu com status %d", resp.StatusCode)
	}

	var end Endereco
	if err := json.NewDecoder(resp.Body).Decode(&end); err != nil {
		return nil, fmt.Errorf("erro ao interpretar resposta do ViaCEP: %w", err)
	}

	switch v := end.Erro.(type) {
	case bool:
		if v {
			return nil, ErrCepNaoEncontrado
		}
	case string:
		if v == "true" {
			return nil, ErrCepNaoEncontrado
		}
	}

	end.Cep = LimparCep(end.Cep)
	end.Logradouro = strings.TrimSpace(end.Logradouro)
	end.Bairro = strings.TrimSpace(end.Bairro)
	end.Localidade = strings.TrimSpace(end.Localidade)
	end.UF = strings.ToUpper(strings.TrimSpace(end.UF))
	return &end, nil
}
