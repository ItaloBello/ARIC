package projeto

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// executor é satisfeito por *sqlx.DB e *sqlx.Tx, para o mesmo método servir
// dentro e fora de transação.
type executor interface {
	GetContext(ctx context.Context, dest any, query string, args ...any) error
	SelectContext(ctx context.Context, dest any, query string, args ...any) error
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) DB() *sqlx.DB { return r.db }

// consulta reúne os parâmetros de qListar.
type consulta struct {
	usuarioID     int
	busca         string
	categoriaID   int
	dificuldadeID int
	autorID       int
	bairroID      int
	cidadeID      int
	ordem         string
	limite        int
	offset        int
	status        []string
	somenteSalvos bool
	projetoID     int
}

func (r *Repository) Listar(ctx context.Context, c consulta) ([]Card, int, error) {
	var rows []cardRow
	err := r.db.SelectContext(ctx, &rows, qListar,
		c.usuarioID, nullStr(c.busca), nullInt(c.categoriaID), nullInt(c.dificuldadeID),
		nullInt(c.autorID), nullInt(c.bairroID), nullInt(c.cidadeID), c.ordem,
		c.limite, c.offset, pq.Array(c.status), c.somenteSalvos, nullInt(c.projetoID))
	if err != nil {
		return nil, 0, err
	}

	cards := make([]Card, 0, len(rows))
	total := 0
	for _, row := range rows {
		cards = append(cards, row.toCard())
		total = row.Total
	}
	return cards, total, nil
}

// Card busca um projeto em qualquer status (a permissão é checada no usecase).
func (r *Repository) Card(ctx context.Context, usuarioID, projetoID int) (*Card, time.Time, error) {
	var row cardRow
	err := r.db.GetContext(ctx, &row, qListar,
		usuarioID, nil, nil, nil, nil, nil, nil, "recentes", 1, 0,
		pq.Array(todosStatus), false, projetoID)
	if err != nil {
		return nil, time.Time{}, err
	}
	card := row.toCard()
	return &card, row.DataAtualizacao, nil
}

var todosStatus = []string{StatusRascunho, StatusRevisao, StatusPublicado, StatusOculto, StatusRemovido}

func (r *Repository) Conteudo(ctx context.Context, usuarioID, projetoID int) (*Detalhe, error) {
	d := &Detalhe{Imagens: []string{}, Materiais: []Material{}, Passos: []Passo{}}

	if err := r.db.GetContext(ctx, &d.Descricao, `SELECT descricao FROM aric.projetos WHERE id = $1`, projetoID); err != nil {
		return nil, err
	}
	if err := r.db.SelectContext(ctx, &d.Imagens, qImagens, projetoID); err != nil {
		return nil, err
	}
	if err := r.db.SelectContext(ctx, &d.Materiais, qMateriais, projetoID); err != nil {
		return nil, err
	}
	if err := r.db.SelectContext(ctx, &d.Passos, qPassos, projetoID); err != nil {
		return nil, err
	}

	var nota int
	err := r.db.GetContext(ctx, &nota, qMinhaNota, projetoID, usuarioID)
	switch {
	case err == nil:
		d.MinhaNota = &nota
	case !errors.Is(err, sql.ErrNoRows):
		return nil, err
	}

	if err := r.db.GetContext(ctx, &d.Comentarios, qContarComentarios, projetoID); err != nil {
		return nil, err
	}
	return d, nil
}

func (r *Repository) Base(ctx context.Context, ex executor, projetoID int, paraAtualizar bool) (*baseRow, error) {
	q := qBase
	if paraAtualizar {
		q = qBaseParaAtualizar
	}
	var b baseRow
	if err := ex.GetContext(ctx, &b, q, projetoID); err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *Repository) Papel(ctx context.Context, usuarioID int) (string, error) {
	var papel string
	err := r.db.GetContext(ctx, &papel, qPapel, usuarioID)
	return papel, err
}

func (r *Repository) LocalUsuario(ctx context.Context, usuarioID int) (bairroID, cidadeID int, err error) {
	var row struct {
		BairroID int `db:"bairro_id"`
		CidadeID int `db:"cidade_id"`
	}
	err = r.db.GetContext(ctx, &row, qLocalUsuario, usuarioID)
	return row.BairroID, row.CidadeID, err
}

func (r *Repository) CategoriasExistem(ctx context.Context, ids []int) (bool, error) {
	if len(ids) == 0 {
		return true, nil
	}
	var n int
	err := r.db.GetContext(ctx, &n, qContarCategorias, pq.Array(ids))
	return n == len(ids), err
}

// ErrDificuldadeInvalida: a dificuldade informada não existe.
var ErrDificuldadeInvalida = errors.New("dificuldade inválida")

func (r *Repository) Inserir(ctx context.Context, tx *sqlx.Tx, req ProjetoRequest, autorID int, status string) (int, error) {
	var id int
	err := tx.GetContext(ctx, &id, qInserir,
		req.Titulo, req.Subtitulo, req.Descricao, req.DificuldadeID,
		req.TempoEstimado, req.ResiduosKg, autorID, status)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrDificuldadeInvalida
	}
	return id, err
}

func (r *Repository) Atualizar(ctx context.Context, tx *sqlx.Tx, projetoID int, req ProjetoRequest, status string) error {
	res, err := tx.ExecContext(ctx, qAtualizar, projetoID,
		req.Titulo, req.Subtitulo, req.Descricao, req.DificuldadeID,
		req.TempoEstimado, req.ResiduosKg, status)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrDificuldadeInvalida
	}
	return nil
}

// SalvarFilhos substitui categorias, materiais, passos e imagens do projeto.
func (r *Repository) SalvarFilhos(ctx context.Context, tx *sqlx.Tx, projetoID int, req ProjetoRequest) error {
	for _, q := range []string{qLimparCategorias, qLimparMateriais, qLimparPassos, qLimparImagens} {
		if _, err := tx.ExecContext(ctx, q, projetoID); err != nil {
			return err
		}
	}

	if len(req.Categorias) > 0 {
		if _, err := tx.ExecContext(ctx, qInserirCategorias, projetoID, pq.Array(req.Categorias)); err != nil {
			return err
		}
	}

	if len(req.Materiais) > 0 {
		nomes, qtds, imgs := make([]string, 0), make([]string, 0), make([]string, 0)
		for _, m := range req.Materiais {
			nomes, qtds, imgs = append(nomes, m.Nome), append(qtds, m.Quantidade), append(imgs, m.ImagemURL)
		}
		if _, err := tx.ExecContext(ctx, qInserirMateriais, projetoID,
			pq.Array(nomes), pq.Array(qtds), pq.Array(imgs)); err != nil {
			return err
		}
	}

	if len(req.Passos) > 0 {
		titulos, descs, imgs := make([]string, 0), make([]string, 0), make([]string, 0)
		for _, p := range req.Passos {
			titulos, descs, imgs = append(titulos, p.Titulo), append(descs, p.Descricao), append(imgs, p.ImagemURL)
		}
		if _, err := tx.ExecContext(ctx, qInserirPassos, projetoID,
			pq.Array(titulos), pq.Array(descs), pq.Array(imgs)); err != nil {
			return err
		}
	}

	if len(req.Imagens) > 0 {
		if _, err := tx.ExecContext(ctx, qInserirImagens, projetoID, pq.Array(req.Imagens)); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) Remover(ctx context.Context, projetoID int) error {
	_, err := r.db.ExecContext(ctx, qRemover, projetoID)
	return err
}

// ---------- moderação ----------

func (r *Repository) Aprovar(ctx context.Context, tx *sqlx.Tx, projetoID, moderadorID int) (bool, error) {
	res, err := tx.ExecContext(ctx, qAprovar, projetoID, moderadorID)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

func (r *Repository) Rejeitar(ctx context.Context, projetoID, moderadorID int, motivo string) (bool, error) {
	res, err := r.db.ExecContext(ctx, qRejeitar, projetoID, moderadorID, motivo)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

func (r *Repository) CategoriasDoProjeto(ctx context.Context, ex executor, projetoID int) ([]int, error) {
	ids := []int{}
	err := ex.SelectContext(ctx, &ids, qCategoriasDoProjeto, projetoID)
	return ids, err
}

func (r *Repository) DefinirDestaque(ctx context.Context, projetoID int) error {
	_, err := r.db.ExecContext(ctx, qDefinirDestaque, projetoID)
	return err
}

func (r *Repository) DestaqueAtual(ctx context.Context) (int, error) {
	var id int
	err := r.db.GetContext(ctx, &id, qDestaqueAtual)
	return id, err
}

// ---------- interações ----------

// InserirConclusao devolve false se o usuário já tinha concluído.
func (r *Repository) InserirConclusao(ctx context.Context, tx *sqlx.Tx, projetoID, usuarioID int, fotoURL string) (bool, error) {
	var quando time.Time
	err := tx.GetContext(ctx, &quando, qInserirConclusao, projetoID, usuarioID, fotoURL)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

func (r *Repository) Curtir(ctx context.Context, projetoID, usuarioID int, curtir bool) (int, error) {
	q := qDescurtir
	if curtir {
		q = qCurtir
	}
	if _, err := r.db.ExecContext(ctx, q, projetoID, usuarioID); err != nil {
		return 0, err
	}
	var total int
	err := r.db.GetContext(ctx, &total, qCurtidas, projetoID)
	return total, err
}

func (r *Repository) Salvar(ctx context.Context, projetoID, usuarioID int, salvar bool) error {
	q := qRemoverSalvo
	if salvar {
		q = qSalvar
	}
	_, err := r.db.ExecContext(ctx, q, projetoID, usuarioID)
	return err
}

// ErrNaoConcluiu: só quem fez o projeto pode avaliar.
var ErrNaoConcluiu = errors.New("não concluiu")

func (r *Repository) Avaliar(ctx context.Context, projetoID, usuarioID, nota int) (*float64, int, error) {
	if _, err := r.db.ExecContext(ctx, qAvaliar, projetoID, usuarioID, nota); err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23503" {
			return nil, 0, ErrNaoConcluiu
		}
		return nil, 0, err
	}
	var row struct {
		Media sql.NullFloat64 `db:"media"`
		Total int             `db:"total"`
	}
	if err := r.db.GetContext(ctx, &row, qMediaNotas, projetoID); err != nil {
		return nil, 0, err
	}
	if !row.Media.Valid {
		return nil, row.Total, nil
	}
	return &row.Media.Float64, row.Total, nil
}

// ---------- comentários ----------

func (r *Repository) ListarComentarios(ctx context.Context, projetoID, limite, offset int) ([]comentarioRow, int, error) {
	rows := []comentarioRow{}
	if err := r.db.SelectContext(ctx, &rows, qListarComentarios, projetoID, limite, offset); err != nil {
		return nil, 0, err
	}
	total := 0
	if len(rows) > 0 {
		total = rows[0].Total
	}
	return rows, total, nil
}

func (r *Repository) InserirComentario(ctx context.Context, projetoID, usuarioID int, conteudo string) (*comentarioRow, error) {
	var row comentarioRow
	if err := r.db.GetContext(ctx, &row, qInserirComentario, projetoID, usuarioID, conteudo); err != nil {
		return nil, err
	}
	row.AutorID = usuarioID
	row.Conteudo = conteudo
	return &row, nil
}

type comentarioBase struct {
	ID        int `db:"id"`
	ProjetoID int `db:"projeto_id"`
	UsuarioID int `db:"usuario_id"`
}

func (r *Repository) Comentario(ctx context.Context, id int) (*comentarioBase, error) {
	var c comentarioBase
	if err := r.db.GetContext(ctx, &c, qComentario, id); err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *Repository) RemoverComentario(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, qRemoverComentario, id)
	return err
}

// ---------- catálogo ----------

func (r *Repository) Categorias(ctx context.Context) ([]Categoria, error) {
	cats := []Categoria{}
	err := r.db.SelectContext(ctx, &cats, qCategorias)
	return cats, err
}

func (r *Repository) Dificuldades(ctx context.Context) ([]DificuldadeCatalogo, error) {
	difs := []DificuldadeCatalogo{}
	err := r.db.SelectContext(ctx, &difs, qDificuldades)
	return difs, err
}

// ---------- helpers ----------

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nullInt(i int) any {
	if i == 0 {
		return nil
	}
	return i
}
