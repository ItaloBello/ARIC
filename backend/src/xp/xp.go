// Package xp registra ganhos de XP. O trigger do banco atualiza xp_total,
// nível e sequência de dias a cada evento inserido.
package xp

import (
	"context"
	"database/sql"
	"errors"
)

// Origens aceitas pela coluna eventos_xp.origem
const (
	ProjetoConcluido    = "projeto_concluido"
	ProjetoPublicado    = "projeto_publicado"
	MissaoConcluida     = "missao_concluida"
	ForumRespostaAceita = "forum_resposta_aceita"
)

// Executor é satisfeito por *sqlx.DB e *sqlx.Tx.
type Executor interface {
	GetContext(ctx context.Context, dest any, query string, args ...any) error
}

// Não dá XP duas vezes pela mesma coisa: se já existe, não insere e não retorna linha.
const qInserirEvento = `
INSERT INTO aric.eventos_xp (usuario_id, quantidade, origem, referencia_id, descricao)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (usuario_id, origem, referencia_id) DO NOTHING
RETURNING quantidade`

const qTotais = `
SELECT u.xp_total, u.nivel_id, n.titulo AS nivel_titulo
FROM aric.usuarios u LEFT JOIN aric.niveis n ON n.id = u.nivel_id
WHERE u.id = $1`

// Dar registra o evento e devolve quanto foi dado (0 se a recompensa já tinha sido dada).
// Chamar na mesma transação da ação que gerou o XP.
func Dar(ctx context.Context, ex Executor, usuarioID, quantidade int, origem string, referenciaID int, descricao string) (int, error) {
	if quantidade == 0 {
		return 0, nil
	}
	var dado int
	err := ex.GetContext(ctx, &dado, qInserirEvento, usuarioID, quantidade, origem, referenciaID, descricao)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return dado, err
}

type Totais struct {
	XPTotal     int     `json:"xpTotal"`
	NivelID     *int64  `json:"nivelId"`
	NivelTitulo *string `json:"nivelTitulo"`
}

// DoUsuario devolve o XP total e o nível atual (para mostrar depois de ganhar XP).
func DoUsuario(ctx context.Context, ex Executor, usuarioID int) (*Totais, error) {
	var row struct {
		XPTotal     int            `db:"xp_total"`
		NivelID     sql.NullInt64  `db:"nivel_id"`
		NivelTitulo sql.NullString `db:"nivel_titulo"`
	}
	if err := ex.GetContext(ctx, &row, qTotais, usuarioID); err != nil {
		return nil, err
	}
	t := &Totais{XPTotal: row.XPTotal}
	if row.NivelID.Valid {
		t.NivelID = &row.NivelID.Int64
	}
	if row.NivelTitulo.Valid {
		t.NivelTitulo = &row.NivelTitulo.String
	}
	return t, nil
}
