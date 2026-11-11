package missao

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
)

// Ações que fazem missões andarem sozinhas (coluna missoes.acao).
const (
	AcaoConcluirProjeto = "concluir_projeto"
	AcaoPublicarProjeto = "publicar_projeto"
	AcaoResponderForum  = "responder_forum"
	AcaoPerguntarForum  = "perguntar_forum"
)

// Executor é satisfeito por *sqlx.DB e *sqlx.Tx.
type Executor interface {
	SelectContext(ctx context.Context, dest any, query string, args ...any) error
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type Concluida struct {
	ID     int    `db:"id" json:"id"`
	Titulo string `db:"titulo" json:"titulo"`
	XP     int    `db:"xp" json:"xp"`
}

// Soma 1 no progresso semanal das missões ativas com essa ação. Se a missão tiver
// categoria, só conta quando a ação envolve aquela categoria. Missão já concluída
// na semana não anda mais.
const qAvancarProgresso = `
INSERT INTO aric.usuarios_missoes AS um (usuario_id, missao_id, semana_inicio, progresso)
SELECT $1, m.id, aric.inicio_semana(now()), 1
FROM aric.missoes m
WHERE m.ativa
  AND m.acao = $2
  AND (m.categoria_id IS NULL OR m.categoria_id = ANY($3::int[]))
ON CONFLICT (usuario_id, missao_id, semana_inicio)
DO UPDATE SET progresso = um.progresso + 1, atualizado_em = now()
WHERE um.concluida_em IS NULL`

// Marca como concluídas as que bateram a meta e dá o XP de cada uma.
const qConcluirMissoes = `
WITH concluidas AS (
    UPDATE aric.usuarios_missoes um
    SET concluida_em = now(), atualizado_em = now()
    FROM aric.missoes m
    WHERE m.id = um.missao_id
      AND um.usuario_id = $1
      AND um.semana_inicio = aric.inicio_semana(now())
      AND um.concluida_em IS NULL
      AND um.progresso >= m.meta
    RETURNING um.id AS usuario_missao_id, m.id, m.titulo, m.xp
),
xp AS (
    INSERT INTO aric.eventos_xp (usuario_id, quantidade, origem, referencia_id, descricao)
    SELECT $1, c.xp, 'missao_concluida', c.usuario_missao_id, 'Missão: ' || c.titulo
    FROM concluidas c
    ON CONFLICT (usuario_id, origem, referencia_id) DO NOTHING
)
SELECT id, titulo, xp FROM concluidas`

// RegistrarAcao avança as missões ligadas à ação e devolve as que foram concluídas agora.
// Chamar dentro da mesma transação da ação (ex.: junto com a conclusão do projeto).
func RegistrarAcao(ctx context.Context, ex Executor, usuarioID int, acao string, categorias []int) ([]Concluida, error) {
	if categorias == nil {
		categorias = []int{}
	}
	if _, err := ex.ExecContext(ctx, qAvancarProgresso, usuarioID, acao, pq.Array(categorias)); err != nil {
		return nil, err
	}

	concluidas := []Concluida{}
	if err := ex.SelectContext(ctx, &concluidas, qConcluirMissoes, usuarioID); err != nil {
		return nil, err
	}
	return concluidas, nil
}
