package forum

import (
	"database/sql"
	"time"

	"ARIC/src/missao"
	"ARIC/utils/web"
)

const (
	TipoDuvida = "duvida"
	TipoDica   = "dica"
)

// ---------- entrada ----------

type PostRequest struct {
	Tipo        string `json:"tipo"` // duvida | dica (ignorado na edição)
	Titulo      string `json:"titulo"`
	Conteudo    string `json:"conteudo"`
	ImagemURL   string `json:"imagemUrl"`
	CategoriaID int    `json:"categoriaId"` // 0 = sem categoria
}

type Filtros struct {
	Busca       string `form:"busca"`
	CategoriaID int    `form:"categoria"`
	Tipo        string `form:"tipo"`     // duvida | dica
	Situacao    string `form:"situacao"` // abertas | resolvidas
	AutorID     int    `form:"autor"`
	Ordem       string `form:"ordem"` // recentes | votadas | relevancia
	Pagina      int    `form:"pagina"`
	Limite      int    `form:"limite"`
}

// ---------- saída ----------

type Autor struct {
	ID      int     `json:"id"`
	Nome    string  `json:"nome"`
	FotoURL *string `json:"fotoUrl"`
	Bairro  string  `json:"bairro"`
}

type Categoria struct {
	ID   int     `json:"id"`
	Nome string  `json:"nome"`
	Cor  *string `json:"cor"`
}

// Post é o card da lista. "resolvido" = dúvida com resposta aceita.
type Post struct {
	ID               int        `json:"id"`
	Tipo             string     `json:"tipo"`
	Titulo           string     `json:"titulo"`
	Resumo           string     `json:"resumo"` // primeiros 200 caracteres
	ImagemURL        *string    `json:"imagemUrl"`
	Resolvido        bool       `json:"resolvido"`
	RespostaAceitaID *int       `json:"respostaAceitaId"`
	Autor            Autor      `json:"autor"`
	Categoria        *Categoria `json:"categoria"`
	Curtidas         int        `json:"curtidas"`
	Respostas        int        `json:"respostas"`
	Curtido          bool       `json:"curtido"`
	Salvo            bool       `json:"salvo"`
	CriadoEm         time.Time  `json:"criadoEm"`
	AtualizadoEm     time.Time  `json:"atualizadoEm"`
}

type PostDetalhe struct {
	Post
	Conteudo    string `json:"conteudo"`
	PodeEditar  bool   `json:"podeEditar"`
	PodeApagar  bool   `json:"podeApagar"`
	PodeAceitar bool   `json:"podeAceitar"` // autor de uma dúvida
}

type Resposta struct {
	ID           int       `json:"id"`
	Conteudo     string    `json:"conteudo"`
	Aceita       bool      `json:"aceita"`
	Autor        Autor     `json:"autor"`
	Curtidas     int       `json:"curtidas"`
	Curtido      bool      `json:"curtido"`
	CriadoEm     time.Time `json:"criadoEm"`
	AtualizadoEm time.Time `json:"atualizadoEm"`
	PodeEditar   bool      `json:"podeEditar"`
	PodeApagar   bool      `json:"podeApagar"`
}

// RespostaCriada inclui as missões que a resposta concluiu (ex.: "responda 3 dúvidas").
type RespostaCriada struct {
	Resposta
	MissoesConcluidas []missao.Concluida `json:"missoesConcluidas"`
}

type PostCriado struct {
	PostDetalhe
	MissoesConcluidas []missao.Concluida `json:"missoesConcluidas"`
}

type CurtidaResponse struct {
	Curtido  bool `json:"curtido"`
	Curtidas int  `json:"curtidas"`
}

// ---------- linhas do banco ----------

type postRow struct {
	ID               int            `db:"id"`
	Tipo             string         `db:"tipo"`
	Titulo           string         `db:"titulo"`
	Resumo           string         `db:"resumo"`
	ImagemURL        sql.NullString `db:"imagem_url"`
	CriadoEm         time.Time      `db:"criado_em"`
	AtualizadoEm     time.Time      `db:"atualizado_em"`
	RespostaAceitaID sql.NullInt64  `db:"resposta_aceita_id"`
	AutorID          int            `db:"autor_id"`
	AutorNome        string         `db:"autor_nome"`
	AutorFotoURL     sql.NullString `db:"autor_foto_url"`
	AutorBairro      string         `db:"autor_bairro"`
	CategoriaID      sql.NullInt64  `db:"categoria_id"`
	CategoriaNome    sql.NullString `db:"categoria_nome"`
	CategoriaCor     sql.NullString `db:"categoria_cor"`
	Curtidas         int            `db:"curtidas"`
	Respostas        int            `db:"respostas"`
	Curtido          bool           `db:"curtido"`
	Salvo            bool           `db:"salvo"`
	Total            int            `db:"total"`
}

func (r postRow) toPost() Post {
	p := Post{
		ID:           r.ID,
		Tipo:         r.Tipo,
		Titulo:       r.Titulo,
		Resumo:       r.Resumo,
		ImagemURL:    web.StrPtr(r.ImagemURL),
		Resolvido:    r.RespostaAceitaID.Valid,
		Autor:        Autor{ID: r.AutorID, Nome: r.AutorNome, FotoURL: web.StrPtr(r.AutorFotoURL), Bairro: r.AutorBairro},
		Curtidas:     r.Curtidas,
		Respostas:    r.Respostas,
		Curtido:      r.Curtido,
		Salvo:        r.Salvo,
		CriadoEm:     r.CriadoEm,
		AtualizadoEm: r.AtualizadoEm,
	}
	if r.RespostaAceitaID.Valid {
		id := int(r.RespostaAceitaID.Int64)
		p.RespostaAceitaID = &id
	}
	if r.CategoriaID.Valid {
		p.Categoria = &Categoria{ID: int(r.CategoriaID.Int64), Nome: r.CategoriaNome.String, Cor: web.StrPtr(r.CategoriaCor)}
	}
	return p
}

type postBase struct {
	ID               int           `db:"id"`
	AutorID          int           `db:"autor_id"`
	Tipo             string        `db:"tipo"`
	CategoriaID      sql.NullInt64 `db:"categoria_id"`
	RespostaAceitaID sql.NullInt64 `db:"resposta_aceita_id"`
	Deletado         bool          `db:"deletado"`
}

func (b postBase) categorias() []int {
	if !b.CategoriaID.Valid {
		return []int{}
	}
	return []int{int(b.CategoriaID.Int64)}
}

type respostaRow struct {
	ID           int            `db:"id"`
	Conteudo     string         `db:"conteudo"`
	CriadoEm     time.Time      `db:"criado_em"`
	AtualizadoEm time.Time      `db:"atualizado_em"`
	AutorID      int            `db:"autor_id"`
	AutorNome    string         `db:"autor_nome"`
	AutorFotoURL sql.NullString `db:"autor_foto_url"`
	AutorBairro  string         `db:"autor_bairro"`
	Aceita       bool           `db:"aceita"`
	Curtidas     int            `db:"curtidas"`
	Curtido      bool           `db:"curtido"`
	Total        int            `db:"total"`
}

type respostaBase struct {
	ID      int `db:"id"`
	PostID  int `db:"post_id"`
	AutorID int `db:"autor_id"`
}
