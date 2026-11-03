package login

import "testing"

func cadastroValido() Cadastro {
	return Cadastro{
		Nome:           "  Maria   Silva ",
		Email:          " Maria@Email.COM ",
		Senha:          "Senha@123",
		ConfirmarSenha: "Senha@123",
		Cep:            "18035-000",
		NumeroEndereco: "100",
	}
}

func TestCadastroNormalizar(t *testing.T) {
	c := cadastroValido()
	c.Normalizar()
	if c.Nome != "Maria Silva" || c.Email != "maria@email.com" || c.Cep != "18035000" {
		t.Fatalf("normalização inesperada: %+v", c)
	}
}

func TestCadastroValidar(t *testing.T) {
	casos := map[string]func(*Cadastro){
		"sem nome":          func(c *Cadastro) { c.Nome = "" },
		"email inválido":    func(c *Cadastro) { c.Email = "maria" },
		"email com nome":    func(c *Cadastro) { c.Email = "Maria <maria@email.com>" },
		"email sem domínio": func(c *Cadastro) { c.Email = "maria@localhost" },
		"cep curto":         func(c *Cadastro) { c.Cep = "1803" },
		"sem número":        func(c *Cadastro) { c.NumeroEndereco = "" },
		"senhas diferentes": func(c *Cadastro) { c.ConfirmarSenha = "Outra@123" },
		"senha curta":       func(c *Cadastro) { c.Senha, c.ConfirmarSenha = "Ab@1", "Ab@1" },
		"sem maiúscula":     func(c *Cadastro) { c.Senha, c.ConfirmarSenha = "senha@123", "senha@123" },
		"sem especial":      func(c *Cadastro) { c.Senha, c.ConfirmarSenha = "Senha1234", "Senha1234" },
		"sem dígito":        func(c *Cadastro) { c.Senha, c.ConfirmarSenha = "Senha@abc", "Senha@abc" },
	}

	ok := cadastroValido()
	ok.Normalizar()
	if err := ok.Validar(); err != nil {
		t.Fatalf("cadastro válido rejeitado: %v", err)
	}

	for nome, alterar := range casos {
		c := cadastroValido()
		alterar(&c)
		c.Normalizar()
		if err := c.Validar(); err == nil {
			t.Errorf("%s: esperava erro", nome)
		}
	}
}
