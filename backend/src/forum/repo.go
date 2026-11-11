package forum

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
)

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

type consulta struct {
	usuarioID     int
	busca         string
	categoriaID   int
	tipo          string
	situacao      string
	autorID       int
	somenteSalvos bool
	postID        int
	ordem         string
	limite        int
	offset        int
}

func (r *Repository) Listar(ctx context.Context, c consulta) ([]Post, int, error) {
	var rows []postRow
	err := r.db.SelectContext(ctx, &rows, qListar,
		c.usuarioID, nullStr(c.busca), nullInt(c.categoriaID), nullStr(c.tipo), nullStr(c.situacao),
		nullInt(c.autorID), c.somenteSalvos, nullInt(c.postID), c.ordem, c.limite, c.offset)
	if err != nil {
		return nil, 0, err
	}
	posts := make([]Post, 0, len(rows))
	total := 0
	for _, row := range rows {
		posts = append(posts, row.toPost())
		total = row.Total
	}
	return posts, total, nil
}

// Post devolve um post não apagado (sql.ErrNoRows se não existir).
func (r *Repository) Post(ctx context.Context, usuarioID, postID int) (*Post, error) {
	var row postRow
	if err := r.db.GetContext(ctx, &row, qListar,
		usuarioID, nil, nil, nil, nil, nil, false, postID, "recentes", 1, 0); err != nil {
		return nil, err
	}
	p := row.toPost()
	return &p, nil
}

func (r *Repository) Conteudo(ctx context.Context, postID int) (string, error) {
	var c string
	err := r.db.GetContext(ctx, &c, qConteudo, postID)
	return c, err
}

func (r *Repository) Base(ctx context.Context, ex executor, postID int, paraAtualizar bool) (*postBase, error) {
	q := qPostBase
	if paraAtualizar {
		q = qPostBaseParaAtualizar
	}
	var b postBase
	if err := ex.GetContext(ctx, &b, q, postID); err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *Repository) Papel(ctx context.Context, usuarioID int) (string, error) {
	var papel string
	err := r.db.GetContext(ctx, &papel, qPapel, usuarioID)
	return papel, err
}

func (r *Repository) CategoriaExiste(ctx context.Context, id int) (bool, error) {
	var existe bool
	err := r.db.GetContext(ctx, &existe, qCategoriaExiste, id)
	return existe, err
}

// ---------- posts ----------

func (r *Repository) InserirPost(ctx context.Context, tx *sqlx.Tx, autorID int, req PostRequest) (int, error) {
	var id int
	err := tx.GetContext(ctx, &id, qInserirPost,
		autorID, req.Tipo, req.Titulo, req.Conteudo, req.ImagemURL, nullInt(req.CategoriaID))
	return id, err
}

func (r *Repository) AtualizarPost(ctx context.Context, postID int, req PostRequest) error {
	_, err := r.db.ExecContext(ctx, qAtualizarPost,
		postID, req.Titulo, req.Conteudo, req.ImagemURL, nullInt(req.CategoriaID))
	return err
}

func (r *Repository) RemoverPost(ctx context.Context, postID int) error {
	_, err := r.db.ExecContext(ctx, qRemoverPost, postID)
	return err
}

// Aceitar marca (ou, com respostaID 0, desmarca) a resposta aceita.
func (r *Repository) Aceitar(ctx context.Context, tx *sqlx.Tx, postID, respostaID int) error {
	_, err := tx.ExecContext(ctx, qAceitar, postID, nullInt(respostaID))
	return err
}

func (r *Repository) CurtirPost(ctx context.Context, postID, usuarioID int, curtir bool) (int, error) {
	return r.curtir(ctx, postID, usuarioID, curtir, qCurtirPost, qDescurtirPost, qCurtidasPost)
}

func (r *Repository) SalvarPost(ctx context.Context, postID, usuarioID int, salvar bool) error {
	q := qRemoverSalvo
	if salvar {
		q = qSalvarPost
	}
	_, err := r.db.ExecContext(ctx, q, postID, usuarioID)
	return err
}

// ---------- respostas ----------

func (r *Repository) ListarRespostas(ctx context.Context, usuarioID, postID, limite, offset int) ([]respostaRow, int, error) {
	return r.listarRespostas(ctx, usuarioID, postID, limite, offset, 0)
}

// Resposta devolve uma resposta no formato da lista (sql.ErrNoRows se não existir).
func (r *Repository) Resposta(ctx context.Context, usuarioID, postID, respostaID int) (*respostaRow, error) {
	rows, _, err := r.listarRespostas(ctx, usuarioID, postID, 1, 0, respostaID)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, sql.ErrNoRows
	}
	return &rows[0], nil
}

func (r *Repository) listarRespostas(ctx context.Context, usuarioID, postID, limite, offset, respostaID int) ([]respostaRow, int, error) {
	rows := []respostaRow{}
	if err := r.db.SelectContext(ctx, &rows, qListarRespostas, usuarioID, postID, limite, offset, nullInt(respostaID)); err != nil {
		return nil, 0, err
	}
	total := 0
	if len(rows) > 0 {
		total = rows[0].Total
	}
	return rows, total, nil
}

func (r *Repository) InserirResposta(ctx context.Context, tx *sqlx.Tx, postID, autorID int, conteudo string) (*respostaRow, error) {
	var row respostaRow
	if err := tx.GetContext(ctx, &row, qInserirResposta, postID, autorID, conteudo); err != nil {
		return nil, err
	}
	row.AutorID = autorID
	row.Conteudo = conteudo
	return &row, nil
}

func (r *Repository) RespostaBase(ctx context.Context, ex executor, respostaID int) (*respostaBase, error) {
	var b respostaBase
	if err := ex.GetContext(ctx, &b, qRespostaBase, respostaID); err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *Repository) AtualizarResposta(ctx context.Context, respostaID int, conteudo string) (time.Time, error) {
	var quando time.Time
	err := r.db.GetContext(ctx, &quando, qAtualizarResposta, respostaID, conteudo)
	return quando, err
}

func (r *Repository) RemoverResposta(ctx context.Context, tx *sqlx.Tx, respostaID int) error {
	if _, err := tx.ExecContext(ctx, qRemoverResposta, respostaID); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, qDesmarcarAceitaApagada, respostaID)
	return err
}

func (r *Repository) CurtirResposta(ctx context.Context, respostaID, usuarioID int, curtir bool) (int, error) {
	return r.curtir(ctx, respostaID, usuarioID, curtir, qCurtirResposta, qDescurtirResposta, qCurtidasResposta)
}

func (r *Repository) curtir(ctx context.Context, alvoID, usuarioID int, curtir bool, qSim, qNao, qTotal string) (int, error) {
	q := qNao
	if curtir {
		q = qSim
	}
	if _, err := r.db.ExecContext(ctx, q, alvoID, usuarioID); err != nil {
		return 0, err
	}
	var total int
	err := r.db.GetContext(ctx, &total, qTotal, alvoID)
	return total, err
}

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
