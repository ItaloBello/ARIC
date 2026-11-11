package forum

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"ARIC/utils/uploader"
	"ARIC/utils/web"
)

const (
	minTitulo   = 5
	maxTitulo   = 200
	maxConteudo = 5000
)

var configImagem = uploader.UploadConfig{
	DestFolder:   "./uploads/forum",
	BaseURL:      "/uploads/forum",
	MaxSize:      5 * 1024 * 1024, // 5MB
	AllowedTypes: []string{".jpg", ".jpeg", ".png", ".webp"},
}

func (r *PostRequest) Normalizar() {
	r.Tipo = strings.ToLower(strings.TrimSpace(r.Tipo))
	r.Titulo = strings.Join(strings.Fields(r.Titulo), " ")
	r.Conteudo = strings.TrimSpace(r.Conteudo)
	r.ImagemURL = strings.TrimSpace(r.ImagemURL)
}

// Validar checa o post. Na edição o tipo não é verificado (não pode ser trocado).
func (r *PostRequest) Validar(edicao bool) error {
	if !edicao && r.Tipo != TipoDuvida && r.Tipo != TipoDica {
		return errors.New("tipo deve ser 'duvida' ou 'dica'")
	}
	n := utf8.RuneCountInString(r.Titulo)
	if n < minTitulo || n > maxTitulo {
		return fmt.Errorf("o título deve ter entre %d e %d caracteres", minTitulo, maxTitulo)
	}
	if err := ValidarConteudo(r.Conteudo); err != nil {
		return err
	}
	if r.CategoriaID < 0 {
		return errors.New("categoria inválida")
	}
	if r.ImagemURL != "" {
		if err := uploader.URLLocal(r.ImagemURL, configImagem); err != nil {
			return err
		}
	}
	return nil
}

func ValidarConteudo(conteudo string) error {
	if strings.TrimSpace(conteudo) == "" {
		return errors.New("o texto está vazio")
	}
	if utf8.RuneCountInString(conteudo) > maxConteudo {
		return fmt.Errorf("o texto deve ter no máximo %d caracteres", maxConteudo)
	}
	return nil
}

func (f *Filtros) normalizar() error {
	f.Busca = strings.TrimSpace(f.Busca)
	f.Pagina, f.Limite = web.Paginacao(f.Pagina, f.Limite)

	switch f.Tipo {
	case "", TipoDuvida, TipoDica:
	default:
		return errors.New("tipo deve ser 'duvida' ou 'dica'")
	}
	switch f.Situacao {
	case "", "abertas", "resolvidas":
	default:
		return errors.New("situacao deve ser 'abertas' ou 'resolvidas'")
	}
	switch f.Ordem {
	case "votadas":
	case "relevancia":
		if f.Busca == "" {
			f.Ordem = "recentes"
		}
	default:
		f.Ordem = "recentes"
	}
	return nil
}
