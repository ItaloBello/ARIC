package forum

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidarPost(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "uploads", "forum"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "uploads", "forum", "foto.jpg"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)

	valido := func() PostRequest {
		return PostRequest{Tipo: " Duvida ", Titulo: "  Posso   reciclar isopor? ", Conteudo: "Onde descarto?", ImagemURL: "/uploads/forum/foto.jpg"}
	}

	r := valido()
	r.Normalizar()
	if r.Tipo != "duvida" || r.Titulo != "Posso reciclar isopor?" {
		t.Fatalf("normalização inesperada: %+v", r)
	}
	if err := r.Validar(false); err != nil {
		t.Fatalf("post válido rejeitado: %v", err)
	}

	casos := map[string]func(*PostRequest){
		"tipo inválido":         func(r *PostRequest) { r.Tipo = "aviso" },
		"título curto":          func(r *PostRequest) { r.Titulo = "Oi" },
		"título longo":          func(r *PostRequest) { r.Titulo = strings.Repeat("a", 201) },
		"sem conteúdo":          func(r *PostRequest) { r.Conteudo = "   " },
		"conteúdo longo":        func(r *PostRequest) { r.Conteudo = strings.Repeat("a", 5001) },
		"imagem externa":        func(r *PostRequest) { r.ImagemURL = "https://x.com/a.jpg" },
		"imagem de outra pasta": func(r *PostRequest) { r.ImagemURL = "/uploads/projetos/foto.jpg" },
		"categoria negativa":    func(r *PostRequest) { r.CategoriaID = -1 },
	}
	for nome, alterar := range casos {
		r := valido()
		alterar(&r)
		r.Normalizar()
		if err := r.Validar(false); err == nil {
			t.Errorf("%s: esperava erro", nome)
		}
	}

	// na edição o tipo é ignorado
	r = valido()
	r.Tipo = ""
	r.Normalizar()
	if err := r.Validar(true); err != nil {
		t.Errorf("edição sem tipo deveria passar: %v", err)
	}
}

func TestFiltros(t *testing.T) {
	f := Filtros{Ordem: "relevancia"}
	if err := f.normalizar(); err != nil || f.Ordem != "recentes" || f.Pagina != 1 || f.Limite != 20 {
		t.Errorf("padrões inesperados: %+v (%v)", f, err)
	}
	for _, f := range []Filtros{{Tipo: "aviso"}, {Situacao: "todas"}} {
		if err := f.normalizar(); err == nil {
			t.Errorf("esperava erro para %+v", f)
		}
	}
	f = Filtros{Tipo: "dica", Situacao: "abertas", Ordem: "votadas", Limite: 999}
	if err := f.normalizar(); err != nil || f.Limite != 50 || f.Ordem != "votadas" {
		t.Errorf("normalização inesperada: %+v (%v)", f, err)
	}
}

func TestCategoriasDoPost(t *testing.T) {
	var b postBase
	if len(b.categorias()) != 0 {
		t.Error("post sem categoria deveria dar lista vazia")
	}
	b.CategoriaID.Valid, b.CategoriaID.Int64 = true, 2
	if c := b.categorias(); len(c) != 1 || c[0] != 2 {
		t.Errorf("categorias inesperadas: %v", c)
	}
}
