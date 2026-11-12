package gamificacao

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) DB() *sqlx.DB { return r.db }

// ---------- perfil ----------

func (r *Repository) Perfil(ctx context.Context, usuarioID int) (*perfilRow, error) {
	var row perfilRow
	if err := r.db.GetContext(ctx, &row, qPerfil, usuarioID); err != nil {
		return nil, err
	}
	return &row, nil
}

// PosicaoNaLiga devolve nil se o usuário não estiver na liga (ex.: liga nula).
func (r *Repository) PosicaoNaLiga(ctx context.Context, ligaID, usuarioID int) (*posicaoRow, error) {
	var row posicaoRow
	err := r.db.GetContext(ctx, &row, qPosicaoNaLiga, ligaID, usuarioID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *Repository) Niveis(ctx context.Context) ([]NivelCatalogo, error) {
	niveis := []NivelCatalogo{}
	err := r.db.SelectContext(ctx, &niveis, qNiveis)
	return niveis, err
}

func (r *Repository) HistoricoXP(ctx context.Context, usuarioID, limite, offset int) ([]eventoRow, error) {
	rows := []eventoRow{}
	err := r.db.SelectContext(ctx, &rows, qHistoricoXP, usuarioID, limite, offset)
	return rows, err
}

// ---------- ligas ----------

func (r *Repository) Ligas(ctx context.Context) ([]Liga, error) {
	ligas := []Liga{}
	err := r.db.SelectContext(ctx, &ligas, qLigas)
	return ligas, err
}

func (r *Repository) LigaDoUsuario(ctx context.Context, usuarioID int) (*ligaUsuarioRow, error) {
	var row ligaUsuarioRow
	if err := r.db.GetContext(ctx, &row, qLigaDoUsuario, usuarioID); err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *Repository) RankingLiga(ctx context.Context, ligaID int) ([]LinhaRanking, error) {
	linhas := []LinhaRanking{}
	err := r.db.SelectContext(ctx, &linhas, qRankingLiga, ligaID)
	return linhas, err
}

func (r *Repository) UltimaSemana(ctx context.Context, usuarioID int) (*ResultadoSemana, error) {
	var row ResultadoSemana
	err := r.db.GetContext(ctx, &row, qUltimaSemana, usuarioID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// ---------- missões ----------

func (r *Repository) MissoesDaSemana(ctx context.Context, usuarioID int) ([]missaoSemanaRow, error) {
	rows := []missaoSemanaRow{}
	err := r.db.SelectContext(ctx, &rows, qMissoesDaSemana, usuarioID)
	return rows, err
}

func (r *Repository) Missao(ctx context.Context, missaoID int) (*missaoRow, error) {
	var m missaoRow
	if err := r.db.GetContext(ctx, &m, qMissao, missaoID); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repository) ProgressoDaSemana(ctx context.Context, tx *sqlx.Tx, usuarioID, missaoID int) (*progressoRow, error) {
	var p progressoRow
	if err := tx.GetContext(ctx, &p, qProgressoDaSemana, usuarioID, missaoID); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) InserirEvidencia(ctx context.Context, tx *sqlx.Tx, usuarioMissaoID int, fotoURL string, quantidade int) (*evidenciaCriadaRow, error) {
	var row evidenciaCriadaRow
	if err := tx.GetContext(ctx, &row, qInserirEvidencia, usuarioMissaoID, fotoURL, quantidade); err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *Repository) MinhasEvidencias(ctx context.Context, usuarioID, limite, offset int) ([]minhaEvidenciaRow, error) {
	rows := []minhaEvidenciaRow{}
	err := r.db.SelectContext(ctx, &rows, qMinhasEvidencias, usuarioID, limite, offset)
	return rows, err
}

// ---------- moderação ----------

func (r *Repository) EvidenciasPendentes(ctx context.Context, limite, offset int) ([]pendenteRow, error) {
	rows := []pendenteRow{}
	err := r.db.SelectContext(ctx, &rows, qEvidenciasPendentes, limite, offset)
	return rows, err
}

func (r *Repository) EvidenciaParaAvaliar(ctx context.Context, tx *sqlx.Tx, id int) (*evidenciaAvaliarRow, error) {
	var row evidenciaAvaliarRow
	if err := tx.GetContext(ctx, &row, qEvidenciaParaAvaliar, id); err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *Repository) MarcarEvidencia(ctx context.Context, tx *sqlx.Tx, id int, status string, motivo any, moderadorID int) error {
	_, err := tx.ExecContext(ctx, qMarcarEvidencia, id, status, motivo, moderadorID)
	return err
}

func (r *Repository) SomarProgresso(ctx context.Context, tx *sqlx.Tx, usuarioMissaoID, quantidade int) (*somaProgressoRow, error) {
	var row somaProgressoRow
	if err := tx.GetContext(ctx, &row, qSomarProgresso, usuarioMissaoID, quantidade); err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *Repository) ConcluirProgresso(ctx context.Context, tx *sqlx.Tx, usuarioMissaoID int) error {
	_, err := tx.ExecContext(ctx, qConcluirProgresso, usuarioMissaoID)
	return err
}

// ---------- fechamento semanal ----------

func (r *Repository) SemanaParaFechar(ctx context.Context) (time.Time, bool, error) {
	var row struct {
		Semana  time.Time `db:"semana"`
		Fechada bool      `db:"fechada"`
	}
	err := r.db.GetContext(ctx, &row, qSemanaParaFechar)
	return row.Semana, row.Fechada, err
}

func (r *Repository) FecharSemana(ctx context.Context, semana time.Time) (int, error) {
	var n int
	err := r.db.GetContext(ctx, &n, qFecharSemana, semana.Format("2006-01-02"))
	return n, err
}
