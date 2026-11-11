package forum

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"mime/multipart"
	"strings"

	"ARIC/src/infra/database"
	"ARIC/src/missao"
	"ARIC/src/xp"
	"ARIC/utils/apperror"
	"ARIC/utils/uploader"
	"ARIC/utils/web"

	"github.com/jmoiron/sqlx"
)

// XP que o autor de uma resposta ganha quando ela é aceita (uma vez por resposta).
const xpRespostaAceita = 50

type UseCases struct {
	repo *Repository
}

func NewUseCases(repo *Repository) *UseCases {
	return &UseCases{repo: repo}
}

func (u *UseCases) ehModerador(ctx context.Context, usuarioID int) (bool, error) {
	papel, err := u.repo.Papel(ctx, usuarioID)
	if err != nil {
		return false, err
	}
	return papel == "moderador" || papel == "admin", nil
}

// base busca o post e devolve 404 se não existir ou estiver apagado.
func (u *UseCases) base(ctx context.Context, ex executor, postID int, paraAtualizar bool) (*postBase, error) {
	b, err := u.repo.Base(ctx, ex, postID, paraAtualizar)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && b.Deletado) {
		return nil, apperror.NotFound("post não encontrado")
	}
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (u *UseCases) respostaBase(ctx context.Context, ex executor, respostaID int) (*respostaBase, error) {
	r, err := u.repo.RespostaBase(ctx, ex, respostaID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.NotFound("resposta não encontrada")
	}
	return r, err
}

func (u *UseCases) validarCategoria(ctx context.Context, id int) error {
	if id == 0 {
		return nil
	}
	existe, err := u.repo.CategoriaExiste(ctx, id)
	if err != nil {
		return interno("erro ao verificar categoria", err)
	}
	if !existe {
		return apperror.BadRequest("categoria inválida")
	}
	return nil
}

// ---------- posts ----------

func (u *UseCases) Listar(ctx context.Context, usuarioID int, f Filtros) (*web.Pagina[Post], error) {
	if err := f.normalizar(); err != nil {
		return nil, apperror.BadRequest(err.Error())
	}
	return u.listar(ctx, consulta{
		usuarioID:   usuarioID,
		busca:       f.Busca,
		categoriaID: f.CategoriaID,
		tipo:        f.Tipo,
		situacao:    f.Situacao,
		autorID:     f.AutorID,
		ordem:       f.Ordem,
		limite:      f.Limite,
		offset:      (f.Pagina - 1) * f.Limite,
	}, f.Pagina)
}

func (u *UseCases) Salvos(ctx context.Context, usuarioID, pagina, limite int) (*web.Pagina[Post], error) {
	pagina, limite = web.Paginacao(pagina, limite)
	return u.listar(ctx, consulta{
		usuarioID:     usuarioID,
		somenteSalvos: true,
		ordem:         "recentes",
		limite:        limite,
		offset:        (pagina - 1) * limite,
	}, pagina)
}

func (u *UseCases) listar(ctx context.Context, c consulta, pagina int) (*web.Pagina[Post], error) {
	posts, total, err := u.repo.Listar(ctx, c)
	if err != nil {
		return nil, interno("erro ao listar posts", err)
	}
	return &web.Pagina[Post]{Itens: posts, Total: total, Pagina: pagina, Limite: c.limite}, nil
}

func (u *UseCases) Detalhe(ctx context.Context, usuarioID, postID int) (*PostDetalhe, error) {
	p, err := u.repo.Post(ctx, usuarioID, postID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.NotFound("post não encontrado")
	}
	if err != nil {
		return nil, interno("erro ao buscar post", err)
	}
	conteudo, err := u.repo.Conteudo(ctx, postID)
	if err != nil {
		return nil, interno("erro ao buscar post", err)
	}
	mod, err := u.ehModerador(ctx, usuarioID)
	if err != nil {
		return nil, interno("erro ao verificar permissão", err)
	}

	autor := p.Autor.ID == usuarioID
	return &PostDetalhe{
		Post:        *p,
		Conteudo:    conteudo,
		PodeEditar:  autor,
		PodeApagar:  autor || mod,
		PodeAceitar: autor && p.Tipo == TipoDuvida,
	}, nil
}

func (u *UseCases) EnviarImagem(arquivo *multipart.FileHeader) (string, error) {
	res, err := uploader.SaveFile(arquivo, configImagem)
	if err != nil {
		return "", apperror.BadRequest(err.Error())
	}
	return res.URL, nil
}

// Criar publica na hora (o fórum não passa por revisão; moderadores podem apagar depois).
// Dúvidas avançam missões de "perguntar_forum".
func (u *UseCases) Criar(ctx context.Context, usuarioID int, req PostRequest) (*PostCriado, error) {
	req.Normalizar()
	if err := req.Validar(false); err != nil {
		return nil, apperror.BadRequest(err.Error())
	}
	if err := u.validarCategoria(ctx, req.CategoriaID); err != nil {
		return nil, err
	}

	var id int
	missoes := []missao.Concluida{}
	err := database.EmTransacao(ctx, u.repo.DB(), func(tx *sqlx.Tx) error {
		var err error
		if id, err = u.repo.InserirPost(ctx, tx, usuarioID, req); err != nil {
			return err
		}
		if req.Tipo != TipoDuvida {
			return nil
		}
		cats := []int{}
		if req.CategoriaID > 0 {
			cats = append(cats, req.CategoriaID)
		}
		missoes, err = missao.RegistrarAcao(ctx, tx, usuarioID, missao.AcaoPerguntarForum, cats)
		return err
	})
	if err != nil {
		return nil, interno("erro ao criar post", err)
	}

	d, err := u.Detalhe(ctx, usuarioID, id)
	if err != nil {
		return nil, err
	}
	return &PostCriado{PostDetalhe: *d, MissoesConcluidas: missoes}, nil
}

// Editar muda título, texto, imagem e categoria. O tipo não muda.
func (u *UseCases) Editar(ctx context.Context, usuarioID, postID int, req PostRequest) (*PostDetalhe, error) {
	req.Normalizar()
	if err := req.Validar(true); err != nil {
		return nil, apperror.BadRequest(err.Error())
	}
	if err := u.validarCategoria(ctx, req.CategoriaID); err != nil {
		return nil, err
	}

	b, err := u.base(ctx, u.repo.DB(), postID, false)
	if err != nil {
		return nil, repassar("erro ao buscar post", err)
	}
	if b.AutorID != usuarioID {
		return nil, apperror.Forbidden("só o autor pode editar o post")
	}
	if err := u.repo.AtualizarPost(ctx, postID, req); err != nil {
		return nil, interno("erro ao editar post", err)
	}
	return u.Detalhe(ctx, usuarioID, postID)
}

func (u *UseCases) Remover(ctx context.Context, usuarioID, postID int) error {
	b, err := u.base(ctx, u.repo.DB(), postID, false)
	if err != nil {
		return repassar("erro ao buscar post", err)
	}
	if b.AutorID != usuarioID {
		mod, err := u.ehModerador(ctx, usuarioID)
		if err != nil {
			return interno("erro ao verificar permissão", err)
		}
		if !mod {
			return apperror.Forbidden("só o autor ou um moderador pode apagar o post")
		}
	}
	if err := u.repo.RemoverPost(ctx, postID); err != nil {
		return interno("erro ao apagar post", err)
	}
	return nil
}

func (u *UseCases) CurtirPost(ctx context.Context, usuarioID, postID int, curtir bool) (*CurtidaResponse, error) {
	if _, err := u.base(ctx, u.repo.DB(), postID, false); err != nil {
		return nil, repassar("erro ao buscar post", err)
	}
	total, err := u.repo.CurtirPost(ctx, postID, usuarioID, curtir)
	if err != nil {
		return nil, interno("erro ao curtir post", err)
	}
	return &CurtidaResponse{Curtido: curtir, Curtidas: total}, nil
}

func (u *UseCases) SalvarPost(ctx context.Context, usuarioID, postID int, salvar bool) error {
	// tirar dos salvos funciona mesmo se o post foi apagado
	if salvar {
		if _, err := u.base(ctx, u.repo.DB(), postID, false); err != nil {
			return repassar("erro ao buscar post", err)
		}
	}
	if err := u.repo.SalvarPost(ctx, postID, usuarioID, salvar); err != nil {
		return interno("erro ao salvar post", err)
	}
	return nil
}

// ---------- resposta aceita ----------

type AceiteResponse struct {
	RespostaAceitaID *int `json:"respostaAceitaId"`
	Resolvido        bool `json:"resolvido"`
}

// Aceitar marca a resposta que resolveu a dúvida. Só o autor da dúvida aceita,
// e não pode aceitar a própria resposta. Quem respondeu ganha XP (uma vez por resposta,
// mesmo que o autor troque a aceita depois).
func (u *UseCases) Aceitar(ctx context.Context, usuarioID, postID, respostaID int) (*AceiteResponse, error) {
	err := database.EmTransacao(ctx, u.repo.DB(), func(tx *sqlx.Tx) error {
		b, err := u.base(ctx, tx, postID, true)
		if err != nil {
			return err
		}
		if b.AutorID != usuarioID {
			return apperror.Forbidden("só o autor da dúvida pode aceitar uma resposta")
		}
		if b.Tipo != TipoDuvida {
			return apperror.Conflict("dicas não têm resposta aceita")
		}

		r, err := u.respostaBase(ctx, tx, respostaID)
		if err != nil {
			return err
		}
		if r.PostID != postID {
			return apperror.NotFound("resposta não encontrada neste post")
		}
		if r.AutorID == usuarioID {
			return apperror.Forbidden("você não pode aceitar a própria resposta")
		}

		if err := u.repo.Aceitar(ctx, tx, postID, respostaID); err != nil {
			return err
		}
		_, err = xp.Dar(ctx, tx, r.AutorID, xpRespostaAceita, xp.ForumRespostaAceita, respostaID, "Resposta aceita no fórum")
		return err
	})
	if err != nil {
		return nil, repassar("erro ao aceitar resposta", err)
	}
	return &AceiteResponse{RespostaAceitaID: &respostaID, Resolvido: true}, nil
}

// Desmarcar volta a dúvida para "em aberto". O XP já dado não é retirado.
func (u *UseCases) Desmarcar(ctx context.Context, usuarioID, postID int) (*AceiteResponse, error) {
	err := database.EmTransacao(ctx, u.repo.DB(), func(tx *sqlx.Tx) error {
		b, err := u.base(ctx, tx, postID, true)
		if err != nil {
			return err
		}
		if b.AutorID != usuarioID {
			return apperror.Forbidden("só o autor da dúvida pode desmarcar a resposta")
		}
		return u.repo.Aceitar(ctx, tx, postID, 0)
	})
	if err != nil {
		return nil, repassar("erro ao desmarcar resposta", err)
	}
	return &AceiteResponse{RespostaAceitaID: nil, Resolvido: false}, nil
}

// ---------- respostas ----------

func (u *UseCases) ListarRespostas(ctx context.Context, usuarioID, postID, pagina, limite int) (*web.Pagina[Resposta], error) {
	if _, err := u.base(ctx, u.repo.DB(), postID, false); err != nil {
		return nil, repassar("erro ao buscar post", err)
	}
	mod, err := u.ehModerador(ctx, usuarioID)
	if err != nil {
		return nil, interno("erro ao verificar permissão", err)
	}

	pagina, limite = web.Paginacao(pagina, limite)
	rows, total, err := u.repo.ListarRespostas(ctx, usuarioID, postID, limite, (pagina-1)*limite)
	if err != nil {
		return nil, interno("erro ao listar respostas", err)
	}

	itens := make([]Resposta, 0, len(rows))
	for _, r := range rows {
		itens = append(itens, toResposta(r, usuarioID, mod))
	}
	return &web.Pagina[Resposta]{Itens: itens, Total: total, Pagina: pagina, Limite: limite}, nil
}

func toResposta(r respostaRow, usuarioID int, mod bool) Resposta {
	return Resposta{
		ID:           r.ID,
		Conteudo:     r.Conteudo,
		Aceita:       r.Aceita,
		Autor:        Autor{ID: r.AutorID, Nome: r.AutorNome, FotoURL: web.StrPtr(r.AutorFotoURL), Bairro: r.AutorBairro},
		Curtidas:     r.Curtidas,
		Curtido:      r.Curtido,
		CriadoEm:     r.CriadoEm,
		AtualizadoEm: r.AtualizadoEm,
		PodeEditar:   r.AutorID == usuarioID,
		PodeApagar:   r.AutorID == usuarioID || mod,
	}
}

// Responder cria a resposta. Responder o post de outra pessoa avança missões de
// "responder_forum" (responder o próprio post não conta).
func (u *UseCases) Responder(ctx context.Context, usuarioID, postID int, conteudo string) (*RespostaCriada, error) {
	conteudo = strings.TrimSpace(conteudo)
	if err := ValidarConteudo(conteudo); err != nil {
		return nil, apperror.BadRequest(err.Error())
	}

	resp := &RespostaCriada{MissoesConcluidas: []missao.Concluida{}}
	err := database.EmTransacao(ctx, u.repo.DB(), func(tx *sqlx.Tx) error {
		b, err := u.base(ctx, tx, postID, false)
		if err != nil {
			return err
		}
		row, err := u.repo.InserirResposta(ctx, tx, postID, usuarioID, conteudo)
		if err != nil {
			return err
		}
		resp.Resposta = toResposta(*row, usuarioID, false)

		if b.AutorID == usuarioID {
			return nil
		}
		resp.MissoesConcluidas, err = missao.RegistrarAcao(ctx, tx, usuarioID, missao.AcaoResponderForum, b.categorias())
		return err
	})
	if err != nil {
		return nil, repassar("erro ao responder", err)
	}
	resp.PodeApagar = true
	return resp, nil
}

func (u *UseCases) EditarResposta(ctx context.Context, usuarioID, respostaID int, conteudo string) (*Resposta, error) {
	conteudo = strings.TrimSpace(conteudo)
	if err := ValidarConteudo(conteudo); err != nil {
		return nil, apperror.BadRequest(err.Error())
	}
	r, err := u.respostaBase(ctx, u.repo.DB(), respostaID)
	if err != nil {
		return nil, repassar("erro ao buscar resposta", err)
	}
	if r.AutorID != usuarioID {
		return nil, apperror.Forbidden("só o autor pode editar a resposta")
	}
	if _, err := u.repo.AtualizarResposta(ctx, respostaID, conteudo); err != nil {
		return nil, interno("erro ao editar resposta", err)
	}

	// devolve a resposta como aparece na lista (com curtidas e "aceita")
	row, err := u.repo.Resposta(ctx, usuarioID, r.PostID, respostaID)
	if err != nil {
		return nil, interno("erro ao buscar resposta", err)
	}
	res := toResposta(*row, usuarioID, false)
	return &res, nil
}

// RemoverResposta: autor da resposta ou moderador. Se era a aceita, a dúvida volta a ficar aberta.
func (u *UseCases) RemoverResposta(ctx context.Context, usuarioID, respostaID int) error {
	err := database.EmTransacao(ctx, u.repo.DB(), func(tx *sqlx.Tx) error {
		r, err := u.respostaBase(ctx, tx, respostaID)
		if err != nil {
			return err
		}
		if r.AutorID != usuarioID {
			mod, err := u.ehModerador(ctx, usuarioID)
			if err != nil {
				return err
			}
			if !mod {
				return apperror.Forbidden("só o autor da resposta ou um moderador pode apagá-la")
			}
		}
		return u.repo.RemoverResposta(ctx, tx, respostaID)
	})
	return repassar("erro ao apagar resposta", err)
}

func (u *UseCases) CurtirResposta(ctx context.Context, usuarioID, respostaID int, curtir bool) (*CurtidaResponse, error) {
	if _, err := u.respostaBase(ctx, u.repo.DB(), respostaID); err != nil {
		return nil, repassar("erro ao buscar resposta", err)
	}
	total, err := u.repo.CurtirResposta(ctx, respostaID, usuarioID, curtir)
	if err != nil {
		return nil, interno("erro ao curtir resposta", err)
	}
	return &CurtidaResponse{Curtido: curtir, Curtidas: total}, nil
}

// ---------- erros ----------

func interno(msg string, err error) error {
	log.Printf("[forum] %s: %v", msg, err)
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
