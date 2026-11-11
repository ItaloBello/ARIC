package projeto

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"unicode/utf8"
)

const (
	pastaImagens   = "./uploads/projetos"
	urlImagens     = "/uploads/projetos"
	maxCategorias  = 5
	maxMateriais   = 30
	maxPassos      = 30
	maxImagens     = 10
	maxComentario  = 2000
	maxMotivo      = 500
	limitePadrao   = 20
	limiteMaximo   = 50
	maxTempoMinuto = 7 * 24 * 60
)

// Normalizar tira espaços sobrando dos textos.
func (r *ProjetoRequest) Normalizar() {
	r.Titulo = strings.Join(strings.Fields(r.Titulo), " ")
	r.Subtitulo = strings.Join(strings.Fields(r.Subtitulo), " ")
	r.Descricao = strings.TrimSpace(r.Descricao)
	for i := range r.Materiais {
		m := &r.Materiais[i]
		m.Nome = strings.Join(strings.Fields(m.Nome), " ")
		m.Quantidade = strings.TrimSpace(m.Quantidade)
		m.ImagemURL = strings.TrimSpace(m.ImagemURL)
	}
	for i := range r.Passos {
		p := &r.Passos[i]
		p.Titulo = strings.Join(strings.Fields(p.Titulo), " ")
		p.Descricao = strings.TrimSpace(p.Descricao)
		p.ImagemURL = strings.TrimSpace(p.ImagemURL)
	}
	for i := range r.Imagens {
		r.Imagens[i] = strings.TrimSpace(r.Imagens[i])
	}
}

// Validar checa tudo que não depende do banco. Rascunho pode ficar incompleto
// (sem materiais, passos ou imagens), mas o que vier precisa ser válido.
func (r *ProjetoRequest) Validar() error {
	if err := tamanho("título", r.Titulo, 3, 255); err != nil {
		return err
	}
	if err := tamanho("subtítulo", r.Subtitulo, 0, 255); err != nil {
		return err
	}
	if !r.Rascunho {
		if r.Subtitulo == "" {
			return errors.New("subtítulo é obrigatório")
		}
		if r.Descricao == "" {
			return errors.New("descrição é obrigatória")
		}
	}
	if utf8.RuneCountInString(r.Descricao) > 10000 {
		return errors.New("descrição deve ter no máximo 10000 caracteres")
	}
	if r.DificuldadeID <= 0 {
		return errors.New("dificuldade é obrigatória")
	}
	if r.TempoEstimado <= 0 || r.TempoEstimado > maxTempoMinuto {
		return fmt.Errorf("tempo estimado deve estar entre 1 e %d minutos", maxTempoMinuto)
	}
	if r.ResiduosKg < 0 || r.ResiduosKg > 1000 {
		return errors.New("resíduos (kg) deve estar entre 0 e 1000")
	}

	if err := r.validarCategorias(); err != nil {
		return err
	}
	if err := r.validarMateriais(); err != nil {
		return err
	}
	if err := r.validarPassos(); err != nil {
		return err
	}
	return r.validarImagens()
}

func (r *ProjetoRequest) validarCategorias() error {
	if len(r.Categorias) > maxCategorias {
		return fmt.Errorf("no máximo %d categorias", maxCategorias)
	}
	if !r.Rascunho && len(r.Categorias) == 0 {
		return errors.New("escolha pelo menos uma categoria")
	}
	vistas := map[int]bool{}
	for _, id := range r.Categorias {
		if id <= 0 || vistas[id] {
			return errors.New("categorias inválidas ou repetidas")
		}
		vistas[id] = true
	}
	return nil
}

func (r *ProjetoRequest) validarMateriais() error {
	if len(r.Materiais) > maxMateriais {
		return fmt.Errorf("no máximo %d materiais", maxMateriais)
	}
	if !r.Rascunho && len(r.Materiais) == 0 {
		return errors.New("informe pelo menos um material")
	}
	vistos := map[string]bool{}
	for i, m := range r.Materiais {
		if err := tamanho(fmt.Sprintf("nome do material %d", i+1), m.Nome, 1, 255); err != nil {
			return err
		}
		if err := tamanho(fmt.Sprintf("quantidade do material %d", i+1), m.Quantidade, 1, 255); err != nil {
			return err
		}
		chave := strings.ToLower(m.Nome)
		if vistos[chave] {
			return fmt.Errorf("material repetido: %s", m.Nome)
		}
		vistos[chave] = true
		if m.ImagemURL != "" {
			if err := validarURLImagem(m.ImagemURL); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *ProjetoRequest) validarPassos() error {
	if len(r.Passos) > maxPassos {
		return fmt.Errorf("no máximo %d passos", maxPassos)
	}
	if !r.Rascunho && len(r.Passos) == 0 {
		return errors.New("informe pelo menos um passo")
	}
	for i, p := range r.Passos {
		if err := tamanho(fmt.Sprintf("título do passo %d", i+1), p.Titulo, 1, 255); err != nil {
			return err
		}
		if err := tamanho(fmt.Sprintf("descrição do passo %d", i+1), p.Descricao, 1, 5000); err != nil {
			return err
		}
		if p.ImagemURL != "" {
			if err := validarURLImagem(p.ImagemURL); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *ProjetoRequest) validarImagens() error {
	if len(r.Imagens) > maxImagens {
		return fmt.Errorf("no máximo %d imagens", maxImagens)
	}
	if !r.Rascunho && len(r.Imagens) == 0 {
		return errors.New("envie pelo menos uma imagem")
	}
	vistas := map[string]bool{}
	for _, u := range r.Imagens {
		if vistas[u] {
			return errors.New("imagem repetida")
		}
		vistas[u] = true
		if err := validarURLImagem(u); err != nil {
			return err
		}
	}
	return nil
}

// validarURLImagem só aceita imagens enviadas por POST /projetos/imagens
// (impede apontar para qualquer arquivo ou site).
func validarURLImagem(u string) error {
	nome := strings.TrimPrefix(u, urlImagens+"/")
	if nome == u || nome == "" || nome != path.Base(nome) || strings.HasPrefix(nome, ".") {
		return fmt.Errorf("imagem inválida: %s (envie antes em /projetos/imagens)", u)
	}
	if _, err := os.Stat(path.Join(pastaImagens, nome)); err != nil {
		return fmt.Errorf("imagem não encontrada: %s", u)
	}
	return nil
}

func tamanho(campo, valor string, min, max int) error {
	n := utf8.RuneCountInString(valor)
	if n < min {
		if min == 1 {
			return fmt.Errorf("%s é obrigatório", campo)
		}
		return fmt.Errorf("%s deve ter pelo menos %d caracteres", campo, min)
	}
	if n > max {
		return fmt.Errorf("%s deve ter no máximo %d caracteres", campo, max)
	}
	return nil
}

// normalizar aplica padrões e limites da paginação e da ordenação.
func (f *Filtros) normalizar() {
	f.Busca = strings.TrimSpace(f.Busca)
	if f.Pagina < 1 {
		f.Pagina = 1
	}
	if f.Limite < 1 {
		f.Limite = limitePadrao
	}
	if f.Limite > limiteMaximo {
		f.Limite = limiteMaximo
	}
	switch f.Ordem {
	case "populares":
	case "relevancia":
		if f.Busca == "" {
			f.Ordem = "recentes"
		}
	default:
		f.Ordem = "recentes"
	}
}

func paginacao(pagina, limite int) (int, int) {
	if pagina < 1 {
		pagina = 1
	}
	if limite < 1 {
		limite = limitePadrao
	}
	if limite > limiteMaximo {
		limite = limiteMaximo
	}
	return pagina, limite
}
