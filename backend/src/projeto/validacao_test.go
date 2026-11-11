package projeto

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// prepararImagens cria uploads/projetos com dois arquivos num diretório temporário
// e entra nele (validarURLImagem usa caminho relativo, como o servidor).
func prepararImagens(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "uploads", "projetos"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, nome := range []string{"capa.jpg", "passo.png"} {
		if err := os.WriteFile(filepath.Join(dir, "uploads", "projetos", nome), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(dir)
}

func requestValido() ProjetoRequest {
	return ProjetoRequest{
		Titulo:        "  Horta   Vertical  ",
		Subtitulo:     "Com garrafas PET",
		Descricao:     "Transforme garrafas em horta",
		DificuldadeID: 1,
		TempoEstimado: 25,
		ResiduosKg:    1.8,
		Categorias:    []int{1, 13},
		Materiais:     []MaterialRequest{{Nome: "Garrafa PET", Quantidade: "3"}},
		Passos:        []PassoRequest{{Titulo: "Corte", Descricao: "Corte as garrafas", ImagemURL: "/uploads/projetos/passo.png"}},
		Imagens:       []string{"/uploads/projetos/capa.jpg"},
	}
}

func TestValidarProjeto(t *testing.T) {
	prepararImagens(t)

	ok := requestValido()
	ok.Normalizar()
	if ok.Titulo != "Horta Vertical" {
		t.Errorf("título não normalizado: %q", ok.Titulo)
	}
	if err := ok.Validar(); err != nil {
		t.Fatalf("request válido rejeitado: %v", err)
	}

	casos := map[string]func(*ProjetoRequest){
		"título curto":       func(r *ProjetoRequest) { r.Titulo = "ab" },
		"sem subtítulo":      func(r *ProjetoRequest) { r.Subtitulo = "" },
		"sem descrição":      func(r *ProjetoRequest) { r.Descricao = " " },
		"sem dificuldade":    func(r *ProjetoRequest) { r.DificuldadeID = 0 },
		"tempo zero":         func(r *ProjetoRequest) { r.TempoEstimado = 0 },
		"resíduo negativo":   func(r *ProjetoRequest) { r.ResiduosKg = -1 },
		"sem categoria":      func(r *ProjetoRequest) { r.Categorias = nil },
		"categoria repetida": func(r *ProjetoRequest) { r.Categorias = []int{1, 1} },
		"6 categorias":       func(r *ProjetoRequest) { r.Categorias = []int{1, 2, 3, 4, 5, 6} },
		"sem material":       func(r *ProjetoRequest) { r.Materiais = nil },
		"material sem qtd":   func(r *ProjetoRequest) { r.Materiais[0].Quantidade = "" },
		"material repetido": func(r *ProjetoRequest) {
			r.Materiais = append(r.Materiais, MaterialRequest{Nome: "garrafa pet", Quantidade: "1"})
		},
		"sem passo":            func(r *ProjetoRequest) { r.Passos = nil },
		"sem imagem":           func(r *ProjetoRequest) { r.Imagens = nil },
		"imagem repetida":      func(r *ProjetoRequest) { r.Imagens = append(r.Imagens, r.Imagens[0]) },
		"imagem de fora":       func(r *ProjetoRequest) { r.Imagens = []string{"https://site.com/a.jpg"} },
		"imagem inexistente":   func(r *ProjetoRequest) { r.Imagens = []string{"/uploads/projetos/nao-existe.jpg"} },
		"path traversal":       func(r *ProjetoRequest) { r.Imagens = []string{"/uploads/projetos/../usuarios/x.jpg"} },
		"arquivo oculto":       func(r *ProjetoRequest) { r.Imagens = []string{"/uploads/projetos/.env"} },
		"imagem do passo ruim": func(r *ProjetoRequest) { r.Passos[0].ImagemURL = "/etc/passwd" },
	}
	for nome, alterar := range casos {
		r := requestValido()
		r.Materiais = append([]MaterialRequest(nil), r.Materiais...)
		r.Passos = append([]PassoRequest(nil), r.Passos...)
		alterar(&r)
		r.Normalizar()
		if err := r.Validar(); err == nil {
			t.Errorf("%s: esperava erro", nome)
		}
	}
}

func TestRascunhoPodeSerIncompleto(t *testing.T) {
	prepararImagens(t)

	r := ProjetoRequest{Titulo: "Ideia nova", DificuldadeID: 1, TempoEstimado: 10, Rascunho: true}
	r.Normalizar()
	if err := r.Validar(); err != nil {
		t.Fatalf("rascunho incompleto deveria passar: %v", err)
	}

	// mas o que vier ainda precisa ser válido
	r.Imagens = []string{"https://site.com/a.jpg"}
	if err := r.Validar(); err == nil || !strings.Contains(err.Error(), "imagem inválida") {
		t.Errorf("rascunho com imagem externa deveria falhar, veio %v", err)
	}
}

func TestFiltrosNormalizar(t *testing.T) {
	f := Filtros{Limite: 500, Ordem: "relevancia"}
	f.normalizar()
	if f.Pagina != 1 || f.Limite != limiteMaximo || f.Ordem != "recentes" {
		t.Errorf("normalização inesperada: %+v", f)
	}

	f = Filtros{Busca: " garrafa ", Ordem: "relevancia"}
	f.normalizar()
	if f.Busca != "garrafa" || f.Ordem != "relevancia" || f.Limite != limitePadrao {
		t.Errorf("normalização inesperada: %+v", f)
	}

	f = Filtros{Ordem: "qualquer"}
	f.normalizar()
	if f.Ordem != "recentes" {
		t.Errorf("ordem desconhecida deveria virar recentes: %q", f.Ordem)
	}
}

func TestPodeVer(t *testing.T) {
	autor := baseRow{Status: StatusRevisao}
	autor.AutorID.Valid, autor.AutorID.Int64 = true, 7

	if !podeVer(&autor, 7, false) {
		t.Error("autor deveria ver o próprio projeto em revisão")
	}
	if podeVer(&autor, 8, false) {
		t.Error("outro usuário não deveria ver projeto em revisão")
	}
	if !podeVer(&autor, 8, true) {
		t.Error("moderador deveria ver projeto em revisão")
	}

	autor.Status = StatusPublicado
	if !podeVer(&autor, 8, false) {
		t.Error("publicado é público")
	}

	autor.Status = StatusRemovido
	if podeVer(&autor, 7, false) {
		t.Error("removido não aparece nem para o autor")
	}
}
