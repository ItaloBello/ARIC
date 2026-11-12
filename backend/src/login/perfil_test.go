package login

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidacoesDoPerfil(t *testing.T) {
	if err := validarNome(""); err == nil {
		t.Error("nome vazio deveria falhar")
	}
	if err := validarNome(strings.Repeat("a", 256)); err == nil {
		t.Error("nome com 256 caracteres deveria falhar")
	}
	if err := validarBio(strings.Repeat("á", 500)); err != nil {
		t.Errorf("bio com 500 caracteres (acentuados) deveria passar: %v", err)
	}
	if err := validarBio(strings.Repeat("a", 501)); err == nil {
		t.Error("bio com 501 caracteres deveria falhar")
	}

	enderecos := []struct {
		cep, numero, compl string
		ok                 bool
	}{
		{"18035000", "10", "", true},
		{"1803500", "10", "", false},
		{"18035000", "", "", false},
		{"18035000", strings.Repeat("1", 21), "", false},
		{"18035000", "10", strings.Repeat("a", 256), false},
	}
	for _, e := range enderecos {
		err := validarEndereco(e.cep, e.numero, e.compl)
		if (err == nil) != e.ok {
			t.Errorf("validarEndereco(%q, %q, %d chars): ok=%v, erro=%v", e.cep, e.numero, len(e.compl), e.ok, err)
		}
	}

	if err := validarSenhaForte("Nova@1234", "Nova@1234"); err != nil {
		t.Errorf("senha forte rejeitada: %v", err)
	}
	if err := validarSenhaForte("Nova@1234", "Nova@12345"); err == nil {
		t.Error("confirmação diferente deveria falhar")
	}
	if err := validarSenhaForte("nova@1234", "nova@1234"); err == nil {
		t.Error("senha sem maiúscula deveria falhar")
	}
}

// apagarFoto só pode remover arquivos dentro de uploads/usuarios.
func TestApagarFotoSoNaPastaDeUsuarios(t *testing.T) {
	dir := t.TempDir()
	for _, p := range []string{"uploads/usuarios", "uploads/projetos"} {
		if err := os.MkdirAll(filepath.Join(dir, p), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	foto := filepath.Join(dir, "uploads/usuarios/velha.jpg")
	outro := filepath.Join(dir, "uploads/projetos/capa.jpg")
	for _, f := range []string{foto, outro} {
		if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(dir)

	apagarFoto("/uploads/projetos/capa.jpg")
	apagarFoto("/uploads/usuarios/../projetos/capa.jpg")
	apagarFoto("")
	if _, err := os.Stat(outro); err != nil {
		t.Error("apagarFoto removeu um arquivo fora de uploads/usuarios")
	}

	apagarFoto("/uploads/usuarios/velha.jpg")
	if _, err := os.Stat(foto); !os.IsNotExist(err) {
		t.Error("apagarFoto deveria ter removido a foto antiga")
	}
}
