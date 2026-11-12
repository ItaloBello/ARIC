// Package denuncia: usuários denunciam conteúdo; moderadores removem ou mantêm.
package denuncia

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strings"
	"time"
	"unicode/utf8"

	"ARIC/src/infra/database"
	"ARIC/utils/apperror"
	"ARIC/utils/web"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// Tipos de conteúdo denunciável
const (
	AlvoProjeto           = "projeto"
	AlvoProjetoComentario = "projeto_comentario"
	AlvoForumPost         = "forum_post"
	AlvoForumResposta     = "forum_resposta"
)

var alvosValidos = map[string]bool{
	AlvoProjeto: true, AlvoProjetoComentario: true, AlvoForumPost: true, AlvoForumResposta: true,
}

var motivosValidos = map[string]bool{
	"spam": true, "ofensivo": true, "perigoso": true, "falso": true, "outro": true,
}

const (
	maxDescricao       = 500
	maxDenunciasDia    = 30 // por pessoa, para ninguém inundar a fila
	motivoPadraoOculto = "Removido pela moderação após denúncias"
	maxMotivoRemocao   = 500
)

// ---------- entrada e saída ----------

type DenunciaRequest struct {
	AlvoTipo  string `json:"alvoTipo"` // projeto | projeto_comentario | forum_post | forum_resposta
	AlvoID    int    `json:"alvoId"`
	Motivo    string `json:"motivo"` // spam | ofensivo | perigoso | falso | outro
	Descricao string `json:"descricao"`
}

type DenunciaCriada struct {
	ID       int       `json:"id"`
	CriadoEm time.Time `json:"criadoEm"`
}

type ResolverRequest struct {
	AlvoTipo string `json:"alvoTipo"`
	AlvoID   int    `json:"alvoId"`
	Acao     string `json:"acao"`   // remover | manter
	Motivo   string `json:"motivo"` // para projetos removidos: o autor vê (opcional)
}

// ItemFila é um conteúdo com denúncias pendentes (várias denúncias viram um item só).
type ItemFila struct {
	AlvoTipo string `json:"alvoTipo"`
	AlvoID   int    `json:"alvoId"`
	// PaiID: projeto do comentário ou post da resposta (para o app abrir o conteúdo)
	PaiID      *int      `json:"paiId"`
	Titulo     *string   `json:"titulo"`
	Trecho     string    `json:"trecho"`
	NoAr       bool      `json:"noAr"` // false se o autor já apagou
	Autor      *Autor    `json:"autor"`
	Denuncias  int       `json:"denuncias"`
	Motivos    []string  `json:"motivos"`
	Descricoes []string  `json:"descricoes"` // até 3, as mais recentes
	PrimeiraEm time.Time `json:"primeiraEm"`
	UltimaEm   time.Time `json:"ultimaEm"`
}

type Autor struct {
	ID   int    `json:"id"`
	Nome string `json:"nome"`
}

type filaRow struct {
	AlvoTipo   string         `db:"alvo_tipo"`
	AlvoID     int            `db:"alvo_id"`
	Denuncias  int            `db:"denuncias"`
	Motivos    pq.StringArray `db:"motivos"`
	Descricoes pq.StringArray `db:"descricoes"`
	PrimeiraEm time.Time      `db:"primeira_em"`
	UltimaEm   time.Time      `db:"ultima_em"`
	Titulo     sql.NullString `db:"titulo"`
	Trecho     sql.NullString `db:"trecho"`
	PaiID      sql.NullInt64  `db:"pai_id"`
	Visivel    bool           `db:"visivel"`
	AutorID    sql.NullInt64  `db:"autor_id"`
	AutorNome  sql.NullString `db:"autor_nome"`
	Total      int            `db:"total"`
}

type alvoRow struct {
	AutorID sql.NullInt64 `db:"autor_id"`
	Visivel bool          `db:"visivel"`
}

// ---------- regras ----------

type UseCases struct {
	db *sqlx.DB
}

func NewUseCases(db *sqlx.DB) *UseCases {
	return &UseCases{db: db}
}

func (u *UseCases) alvo(ctx context.Context, ex sqlx.QueryerContext, tipo string, id int) (*alvoRow, error) {
	var a alvoRow
	err := sqlx.GetContext(ctx, ex, &a, qAlvo, tipo, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.NotFound("conteúdo não encontrado")
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// Denunciar registra a denúncia. Não dá para denunciar o próprio conteúdo,
// conteúdo que já saiu do ar, nem denunciar a mesma coisa duas vezes.
func (u *UseCases) Denunciar(ctx context.Context, usuarioID int, req DenunciaRequest) (*DenunciaCriada, error) {
	req.AlvoTipo = strings.TrimSpace(req.AlvoTipo)
	req.Motivo = strings.TrimSpace(req.Motivo)
	req.Descricao = strings.TrimSpace(req.Descricao)
	switch {
	case !alvosValidos[req.AlvoTipo]:
		return nil, apperror.BadRequest("alvoTipo deve ser projeto, projeto_comentario, forum_post ou forum_resposta")
	case req.AlvoID <= 0:
		return nil, apperror.BadRequest("alvoId inválido")
	case !motivosValidos[req.Motivo]:
		return nil, apperror.BadRequest("motivo deve ser spam, ofensivo, perigoso, falso ou outro")
	case req.Motivo == "outro" && req.Descricao == "":
		return nil, apperror.BadRequest("conte o que aconteceu quando o motivo for 'outro'")
	case utf8.RuneCountInString(req.Descricao) > maxDescricao:
		return nil, apperror.BadRequest("a descrição deve ter no máximo 500 caracteres")
	}

	a, err := u.alvo(ctx, u.db, req.AlvoTipo, req.AlvoID)
	if err != nil {
		return nil, repassar("erro ao buscar conteúdo", err)
	}
	if !a.Visivel {
		return nil, apperror.NotFound("conteúdo não encontrado")
	}
	if a.AutorID.Valid && int(a.AutorID.Int64) == usuarioID {
		return nil, apperror.BadRequest("você não pode denunciar o próprio conteúdo")
	}

	var hoje int
	if err := u.db.GetContext(ctx, &hoje, qDenunciasHoje, usuarioID); err != nil {
		return nil, interno("erro ao verificar denúncias", err)
	}
	if hoje >= maxDenunciasDia {
		return nil, apperror.TooManyRequests("você atingiu o limite de denúncias por hoje", 0)
	}

	var criada DenunciaCriada
	err = u.db.QueryRowxContext(ctx, qInserir, usuarioID, req.AlvoTipo, req.AlvoID, req.Motivo, req.Descricao).
		Scan(&criada.ID, &criada.CriadoEm)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.Conflict("você já denunciou este conteúdo")
	}
	if err != nil {
		return nil, interno("erro ao registrar denúncia", err)
	}
	return &criada, nil
}

func (u *UseCases) Fila(ctx context.Context, pagina, limite int) (*web.Pagina[ItemFila], error) {
	pagina, limite = web.Paginacao(pagina, limite)
	rows := []filaRow{}
	if err := u.db.SelectContext(ctx, &rows, qFila, limite, (pagina-1)*limite); err != nil {
		return nil, interno("erro ao listar denúncias", err)
	}

	itens := make([]ItemFila, 0, len(rows))
	total := 0
	for _, r := range rows {
		item := ItemFila{
			AlvoTipo:   r.AlvoTipo,
			AlvoID:     r.AlvoID,
			Titulo:     web.StrPtr(r.Titulo),
			Trecho:     r.Trecho.String,
			NoAr:       r.Visivel,
			Denuncias:  r.Denuncias,
			Motivos:    []string(r.Motivos),
			Descricoes: []string(r.Descricoes),
			PrimeiraEm: r.PrimeiraEm,
			UltimaEm:   r.UltimaEm,
		}
		if item.Descricoes == nil {
			item.Descricoes = []string{}
		}
		if r.PaiID.Valid {
			id := int(r.PaiID.Int64)
			item.PaiID = &id
		}
		if r.AutorID.Valid {
			item.Autor = &Autor{ID: int(r.AutorID.Int64), Nome: r.AutorNome.String}
		}
		itens = append(itens, item)
		total = r.Total
	}
	return &web.Pagina[ItemFila]{Itens: itens, Total: total, Pagina: pagina, Limite: limite}, nil
}

// Resolver decide todas as denúncias pendentes de um conteúdo de uma vez.
// "remover" tira o conteúdo do ar e marca as denúncias como procedentes;
// "manter" só marca como improcedentes.
func (u *UseCases) Resolver(ctx context.Context, moderadorID int, req ResolverRequest) error {
	req.Motivo = strings.TrimSpace(req.Motivo)
	switch {
	case !alvosValidos[req.AlvoTipo] || req.AlvoID <= 0:
		return apperror.BadRequest("informe alvoTipo e alvoId")
	case req.Acao != "remover" && req.Acao != "manter":
		return apperror.BadRequest("acao deve ser 'remover' ou 'manter'")
	case utf8.RuneCountInString(req.Motivo) > maxMotivoRemocao:
		return apperror.BadRequest("motivo deve ter no máximo 500 caracteres")
	}

	err := database.EmTransacao(ctx, u.db, func(tx *sqlx.Tx) error {
		var pendentes []int
		if err := tx.SelectContext(ctx, &pendentes, qTravarPendentes, req.AlvoTipo, req.AlvoID); err != nil {
			return err
		}
		if len(pendentes) == 0 {
			return apperror.Conflict("não há denúncias pendentes para este conteúdo")
		}

		a, err := u.alvo(ctx, tx, req.AlvoTipo, req.AlvoID)
		if err != nil {
			return err
		}
		if a.AutorID.Valid && int(a.AutorID.Int64) == moderadorID {
			return apperror.Forbidden("um moderador não resolve denúncias sobre o próprio conteúdo")
		}

		status := "improcedente"
		if req.Acao == "remover" {
			status = "procedente"
			if err := removerConteudo(ctx, tx, req, moderadorID); err != nil {
				return err
			}
		}
		_, err = tx.ExecContext(ctx, qResolver, req.AlvoTipo, req.AlvoID, status, moderadorID)
		return err
	})
	return repassar("erro ao resolver denúncia", err)
}

func removerConteudo(ctx context.Context, tx *sqlx.Tx, req ResolverRequest, moderadorID int) error {
	switch req.AlvoTipo {
	case AlvoProjeto:
		motivo := req.Motivo
		if motivo == "" {
			motivo = motivoPadraoOculto
		}
		_, err := tx.ExecContext(ctx, qRemoverProjeto, req.AlvoID, motivo, moderadorID)
		return err
	case AlvoProjetoComentario:
		_, err := tx.ExecContext(ctx, qRemoverComentario, req.AlvoID)
		return err
	case AlvoForumPost:
		_, err := tx.ExecContext(ctx, qRemoverPost, req.AlvoID)
		return err
	case AlvoForumResposta:
		if _, err := tx.ExecContext(ctx, qRemoverResposta, req.AlvoID); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, qDesmarcarAceita, req.AlvoID)
		return err
	}
	return nil
}

// ---------- erros ----------

func interno(msg string, err error) error {
	log.Printf("[denuncia] %s: %v", msg, err)
	return apperror.Internal(msg)
}

func repassar(msg string, err error) error {
	if err == nil {
		return nil
	}
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return interno(msg, err)
}
