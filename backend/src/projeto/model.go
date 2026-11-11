package projeto

import (
	"database/sql"
	"encoding/json"
	"time"

	"ARIC/src/missao"
	"ARIC/src/xp"
	"ARIC/utils/web"
)

// Status possíveis (coluna projetos.status)
const (
	StatusRascunho  = "rascunho"
	StatusRevisao   = "em_revisao"
	StatusPublicado = "publicado"
	StatusOculto    = "oculto"
	StatusRemovido  = "removido"
)

// ---------- entrada ----------

// ProjetoRequest cria ou edita um projeto (JSON). As imagens são enviadas antes em
// POST /projetos/imagens, e aqui vão só as URLs devolvidas.
type ProjetoRequest struct {
	Titulo        string            `json:"titulo"`
	Subtitulo     string            `json:"subtitulo"`
	Descricao     string            `json:"descricao"`
	DificuldadeID int               `json:"dificuldadeId"`
	TempoEstimado int               `json:"tempoEstimado"` // minutos
	ResiduosKg    float64           `json:"residuosKg"`
	Categorias    []int             `json:"categorias"`
	Materiais     []MaterialRequest `json:"materiais"`
	Passos        []PassoRequest    `json:"passos"`
	Imagens       []string          `json:"imagens"` // a primeira é a capa
	// true = salva como rascunho; false = envia para revisão
	Rascunho bool `json:"rascunho"`
}

type MaterialRequest struct {
	Nome       string `json:"nome"`
	Quantidade string `json:"quantidade"`
	ImagemURL  string `json:"imagemUrl"`
}

type PassoRequest struct {
	Titulo    string `json:"titulo"`
	Descricao string `json:"descricao"`
	ImagemURL string `json:"imagemUrl"`
}

// Filtros da listagem (query string de GET /projetos)
type Filtros struct {
	Busca         string `form:"busca"`
	CategoriaID   int    `form:"categoria"`
	DificuldadeID int    `form:"dificuldade"`
	AutorID       int    `form:"autor"`
	Escopo        string `form:"escopo"` // "" | "bairro" | "cidade" (vizinhança de quem pede)
	Ordem         string `form:"ordem"`  // recentes | populares | relevancia
	Pagina        int    `form:"pagina"`
	Limite        int    `form:"limite"`
}

// ---------- saída ----------

type Dificuldade struct {
	ID        int    `json:"id"`
	Descricao string `json:"descricao"`
}

type Autor struct {
	ID      int     `json:"id"`
	Nome    string  `json:"nome"`
	FotoURL *string `json:"fotoUrl"`
	Bairro  *string `json:"bairro"`
}

// Card é o resumo usado nas listas (Home, busca, salvos, meus projetos).
type Card struct {
	ID              int             `json:"id"`
	Titulo          string          `json:"titulo"`
	Subtitulo       string          `json:"subtitulo"`
	Status          string          `json:"status"`
	MotivoModeracao *string         `json:"motivoModeracao,omitempty"`
	XP              int             `json:"xp"`
	TempoEstimado   int             `json:"tempoEstimado"`
	ResiduosKg      float64         `json:"residuosKg"`
	Dificuldade     Dificuldade     `json:"dificuldade"`
	CapaURL         *string         `json:"capaUrl"`
	Autor           *Autor          `json:"autor"` // null = projeto oficial do ARIC
	Categorias      json.RawMessage `json:"categorias"`
	Curtidas        int             `json:"curtidas"`
	Conclusoes      int             `json:"conclusoes"` // "1,2 mil pessoas já fizeram"
	NotaMedia       *float64        `json:"notaMedia"`
	Curtido         bool            `json:"curtido"`
	Salvo           bool            `json:"salvo"`
	Concluido       bool            `json:"concluido"`
	CriadoEm        time.Time       `json:"criadoEm"`
	PublicadoEm     *time.Time      `json:"publicadoEm"`
}

type Material struct {
	Nome       string  `db:"nome" json:"nome"`
	Quantidade string  `db:"quantidade" json:"quantidade"`
	ImagemURL  *string `db:"imagem_url" json:"imagemUrl"`
}

type Passo struct {
	Numero    int     `db:"numero" json:"numero"`
	Titulo    string  `db:"titulo" json:"titulo"`
	Descricao string  `db:"descricao" json:"descricao"`
	ImagemURL *string `db:"imagem_url" json:"imagemUrl"`
}

type Detalhe struct {
	Card
	Descricao    string     `json:"descricao"`
	Imagens      []string   `json:"imagens"`
	Materiais    []Material `json:"materiais"`
	Passos       []Passo    `json:"passos"`
	MinhaNota    *int       `json:"minhaNota"`
	Comentarios  int        `json:"comentarios"`
	AtualizadoEm time.Time  `json:"atualizadoEm"`
	PodeEditar   bool       `json:"podeEditar"`
}

type ConclusaoResponse struct {
	XPGanho           int                `json:"xpGanho"`
	MissoesConcluidas []missao.Concluida `json:"missoesConcluidas"`
	xp.Totais                            // xpTotal, nivelId, nivelTitulo
}

type CurtidaResponse struct {
	Curtido  bool `json:"curtido"`
	Curtidas int  `json:"curtidas"`
}

type AvaliacaoResponse struct {
	MinhaNota  int      `json:"minhaNota"`
	NotaMedia  *float64 `json:"notaMedia"`
	Avaliacoes int      `json:"avaliacoes"`
}

type Comentario struct {
	ID         int       `json:"id"`
	Conteudo   string    `json:"conteudo"`
	CriadoEm   time.Time `json:"criadoEm"`
	Autor      Autor     `json:"autor"`
	PodeApagar bool      `json:"podeApagar"`
}

type Categoria struct {
	ID    int     `db:"id" json:"id"`
	Nome  string  `db:"nome" json:"nome"`
	Tipo  string  `db:"tipo" json:"tipo"`
	Cor   *string `db:"cor" json:"cor"`
	Icone *string `db:"icone" json:"icone"`
}

type DificuldadeCatalogo struct {
	ID        int    `db:"id" json:"id"`
	Descricao string `db:"descricao" json:"descricao"`
	XP        int    `db:"xp_padrao" json:"xp"`
}

// ---------- linhas do banco ----------

type cardRow struct {
	ID              int             `db:"id"`
	Titulo          string          `db:"titulo"`
	Subtitulo       string          `db:"subtitulo"`
	Status          string          `db:"status"`
	MotivoModeracao sql.NullString  `db:"motivo_moderacao"`
	XP              int             `db:"xp"`
	TempoEstimado   int             `db:"tempo_estimado"`
	ResiduosKg      float64         `db:"residuos_kg"`
	DificuldadeID   int             `db:"dificuldade_id"`
	Dificuldade     string          `db:"dificuldade"`
	DataCriacao     time.Time       `db:"data_criacao"`
	DataAtualizacao time.Time       `db:"data_atualizacao"`
	PublicadoEm     sql.NullTime    `db:"publicado_em"`
	CapaURL         sql.NullString  `db:"capa_url"`
	AutorID         sql.NullInt64   `db:"autor_id"`
	AutorNome       sql.NullString  `db:"autor_nome"`
	AutorFotoURL    sql.NullString  `db:"autor_foto_url"`
	AutorBairro     sql.NullString  `db:"autor_bairro"`
	Categorias      []byte          `db:"categorias"`
	Curtidas        int             `db:"curtidas"`
	Conclusoes      int             `db:"conclusoes"`
	NotaMedia       sql.NullFloat64 `db:"nota_media"`
	Curtido         bool            `db:"curtido"`
	Salvo           bool            `db:"salvo"`
	Concluido       bool            `db:"concluido"`
	Total           int             `db:"total"`
}

func (r cardRow) toCard() Card {
	c := Card{
		ID:              r.ID,
		Titulo:          r.Titulo,
		Subtitulo:       r.Subtitulo,
		Status:          r.Status,
		MotivoModeracao: web.StrPtr(r.MotivoModeracao),
		XP:              r.XP,
		TempoEstimado:   r.TempoEstimado,
		ResiduosKg:      r.ResiduosKg,
		Dificuldade:     Dificuldade{ID: r.DificuldadeID, Descricao: r.Dificuldade},
		CapaURL:         web.StrPtr(r.CapaURL),
		Categorias:      json.RawMessage(r.Categorias),
		Curtidas:        r.Curtidas,
		Conclusoes:      r.Conclusoes,
		Curtido:         r.Curtido,
		Salvo:           r.Salvo,
		Concluido:       r.Concluido,
		CriadoEm:        r.DataCriacao,
	}
	if len(c.Categorias) == 0 {
		c.Categorias = json.RawMessage("[]")
	}
	if r.NotaMedia.Valid {
		c.NotaMedia = &r.NotaMedia.Float64
	}
	if r.PublicadoEm.Valid {
		c.PublicadoEm = &r.PublicadoEm.Time
	}
	if r.AutorID.Valid {
		c.Autor = &Autor{
			ID:      int(r.AutorID.Int64),
			Nome:    r.AutorNome.String,
			FotoURL: web.StrPtr(r.AutorFotoURL),
			Bairro:  web.StrPtr(r.AutorBairro),
		}
	}
	return c
}

type baseRow struct {
	ID          int           `db:"id"`
	AutorID     sql.NullInt64 `db:"autor_id"`
	Status      string        `db:"status"`
	XP          int           `db:"xp"`
	PublicadoEm sql.NullTime  `db:"publicado_em"`
}

func (b baseRow) ehAutor(usuarioID int) bool {
	return b.AutorID.Valid && int(b.AutorID.Int64) == usuarioID
}

type comentarioRow struct {
	ID           int            `db:"id"`
	Conteudo     string         `db:"conteudo"`
	CriadoEm     time.Time      `db:"criado_em"`
	AutorID      int            `db:"autor_id"`
	AutorNome    string         `db:"autor_nome"`
	AutorFotoURL sql.NullString `db:"autor_foto_url"`
	Total        int            `db:"total"`
}
