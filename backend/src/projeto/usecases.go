package projeto

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"mime/multipart"
	"strings"
	"unicode/utf8"

	"ARIC/src/infra/database"
	"ARIC/src/missao"
	"ARIC/src/xp"
	"ARIC/utils/apperror"
	"ARIC/utils/uploader"
	"ARIC/utils/web"

	"github.com/jmoiron/sqlx"
)

// XP que o autor ganha quando o projeto é aprovado pela primeira vez.
const xpPublicacao = 100

var configImagem = uploader.UploadConfig{
	DestFolder:   pastaImagens,
	BaseURL:      urlImagens,
	MaxSize:      5 * 1024 * 1024, // 5MB
	AllowedTypes: []string{".jpg", ".jpeg", ".png", ".webp"},
}

type UseCases struct {
	repo *Repository
}

func NewUseCases(repo *Repository) *UseCases {
	return &UseCases{repo: repo}
}

// ehModerador consulta o papel no banco (não confia no token).
func (u *UseCases) ehModerador(ctx context.Context, usuarioID int) (bool, error) {
	papel, err := u.repo.Papel(ctx, usuarioID)
	if err != nil {
		return false, err
	}
	return papel == "moderador" || papel == "admin", nil
}

// podeVer: publicado é público; os outros status só o autor e moderadores veem.
// Removido só moderador vê.
func podeVer(b *baseRow, usuarioID int, moderador bool) bool {
	switch {
	case moderador:
		return true
	case b.Status == StatusPublicado:
		return true
	case b.Status == StatusRemovido:
		return false
	default:
		return b.ehAutor(usuarioID)
	}
}

// baseVisivel busca o projeto e devolve 404 se o usuário não pode vê-lo
// (não revela que existe um rascunho com esse id).
func (u *UseCases) baseVisivel(ctx context.Context, usuarioID, projetoID int) (*baseRow, bool, error) {
	b, err := u.repo.Base(ctx, u.repo.DB(), projetoID, false)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, apperror.NotFound("projeto não encontrado")
	}
	if err != nil {
		return nil, false, interno("erro ao buscar projeto", err)
	}
	mod, err := u.ehModerador(ctx, usuarioID)
	if err != nil {
		return nil, false, interno("erro ao verificar permissão", err)
	}
	if !podeVer(b, usuarioID, mod) {
		return nil, false, apperror.NotFound("projeto não encontrado")
	}
	return b, mod, nil
}

// basePublicado exige projeto publicado (curtir, salvar, concluir, avaliar, comentar).
func (u *UseCases) basePublicado(ctx context.Context, usuarioID, projetoID int) (*baseRow, bool, error) {
	b, mod, err := u.baseVisivel(ctx, usuarioID, projetoID)
	if err != nil {
		return nil, false, err
	}
	if b.Status != StatusPublicado {
		return nil, false, apperror.Conflict("o projeto ainda não está publicado")
	}
	return b, mod, nil
}

// ---------- leitura ----------

func (u *UseCases) Listar(ctx context.Context, usuarioID int, f Filtros) (*web.Pagina[Card], error) {
	f.normalizar()
	c := consulta{
		usuarioID:     usuarioID,
		busca:         f.Busca,
		categoriaID:   f.CategoriaID,
		dificuldadeID: f.DificuldadeID,
		autorID:       f.AutorID,
		ordem:         f.Ordem,
		limite:        f.Limite,
		offset:        (f.Pagina - 1) * f.Limite,
		status:        []string{StatusPublicado},
	}

	switch f.Escopo {
	case "":
	case "bairro", "cidade":
		bairroID, cidadeID, err := u.repo.LocalUsuario(ctx, usuarioID)
		if err != nil {
			return nil, interno("erro ao buscar bairro do usuário", err)
		}
		if f.Escopo == "bairro" {
			c.bairroID = bairroID
		} else {
			c.cidadeID = cidadeID
		}
	default:
		return nil, apperror.BadRequest("escopo deve ser 'bairro' ou 'cidade'")
	}

	return u.listar(ctx, c, f.Pagina)
}

// Meus lista os projetos do próprio usuário em qualquer status (menos removido).
func (u *UseCases) Meus(ctx context.Context, usuarioID, pagina, limite int) (*web.Pagina[Card], error) {
	pagina, limite = web.Paginacao(pagina, limite)
	return u.listar(ctx, consulta{
		usuarioID: usuarioID,
		autorID:   usuarioID,
		ordem:     "recentes",
		limite:    limite,
		offset:    (pagina - 1) * limite,
		status:    []string{StatusRascunho, StatusRevisao, StatusPublicado, StatusOculto},
	}, pagina)
}

func (u *UseCases) Salvos(ctx context.Context, usuarioID, pagina, limite int) (*web.Pagina[Card], error) {
	pagina, limite = web.Paginacao(pagina, limite)
	return u.listar(ctx, consulta{
		usuarioID:     usuarioID,
		ordem:         "recentes",
		limite:        limite,
		offset:        (pagina - 1) * limite,
		status:        []string{StatusPublicado},
		somenteSalvos: true,
	}, pagina)
}

// FilaModeracao: projetos aguardando revisão, mais antigos primeiro.
func (u *UseCases) FilaModeracao(ctx context.Context, usuarioID, pagina, limite int) (*web.Pagina[Card], error) {
	pagina, limite = web.Paginacao(pagina, limite)
	return u.listar(ctx, consulta{
		usuarioID: usuarioID,
		ordem:     "antigos",
		limite:    limite,
		offset:    (pagina - 1) * limite,
		status:    []string{StatusRevisao},
	}, pagina)
}

func (u *UseCases) listar(ctx context.Context, c consulta, pagina int) (*web.Pagina[Card], error) {
	cards, total, err := u.repo.Listar(ctx, c)
	if err != nil {
		return nil, interno("erro ao listar projetos", err)
	}
	return &web.Pagina[Card]{Itens: cards, Total: total, Pagina: pagina, Limite: c.limite}, nil
}

func (u *UseCases) Detalhe(ctx context.Context, usuarioID, projetoID int) (*Detalhe, error) {
	b, _, err := u.baseVisivel(ctx, usuarioID, projetoID)
	if err != nil {
		return nil, err
	}

	card, atualizado, err := u.repo.Card(ctx, usuarioID, projetoID)
	if err != nil {
		return nil, interno("erro ao buscar projeto", err)
	}
	d, err := u.repo.Conteudo(ctx, usuarioID, projetoID)
	if err != nil {
		return nil, interno("erro ao buscar conteúdo do projeto", err)
	}

	d.Card = *card
	d.AtualizadoEm = atualizado
	d.PodeEditar = b.ehAutor(usuarioID) && b.Status != StatusRemovido
	return d, nil
}

// Destaque devolve o projeto destaque da semana, ou nil se não houver.
func (u *UseCases) Destaque(ctx context.Context, usuarioID int) (*Card, error) {
	id, err := u.repo.DestaqueAtual(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, interno("erro ao buscar destaque", err)
	}
	card, _, err := u.repo.Card(ctx, usuarioID, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, interno("erro ao buscar destaque", err)
	}
	// destaque tirado do ar depois de escolhido não aparece
	if card.Status != StatusPublicado {
		return nil, nil
	}
	return card, nil
}

func (u *UseCases) Categorias(ctx context.Context) ([]Categoria, error) {
	cats, err := u.repo.Categorias(ctx)
	if err != nil {
		return nil, interno("erro ao listar categorias", err)
	}
	return cats, nil
}

func (u *UseCases) Dificuldades(ctx context.Context) ([]DificuldadeCatalogo, error) {
	difs, err := u.repo.Dificuldades(ctx)
	if err != nil {
		return nil, interno("erro ao listar dificuldades", err)
	}
	return difs, nil
}

// ---------- escrita ----------

// EnviarImagem salva uma imagem e devolve a URL para usar no projeto.
func (u *UseCases) EnviarImagem(arquivo *multipart.FileHeader) (string, error) {
	res, err := uploader.SaveFile(arquivo, configImagem)
	if err != nil {
		return "", apperror.BadRequest(err.Error())
	}
	return res.URL, nil
}

func (u *UseCases) validarRequest(ctx context.Context, req *ProjetoRequest) error {
	req.Normalizar()
	if err := req.Validar(); err != nil {
		return apperror.BadRequest(err.Error())
	}
	ok, err := u.repo.CategoriasExistem(ctx, req.Categorias)
	if err != nil {
		return interno("erro ao verificar categorias", err)
	}
	if !ok {
		return apperror.BadRequest("categoria inválida")
	}
	return nil
}

func statusInicial(req ProjetoRequest) string {
	if req.Rascunho {
		return StatusRascunho
	}
	return StatusRevisao
}

func (u *UseCases) Criar(ctx context.Context, usuarioID int, req ProjetoRequest) (*Detalhe, error) {
	if err := u.validarRequest(ctx, &req); err != nil {
		return nil, err
	}

	var id int
	err := database.EmTransacao(ctx, u.repo.DB(), func(tx *sqlx.Tx) error {
		var err error
		if id, err = u.repo.Inserir(ctx, tx, req, usuarioID, statusInicial(req)); err != nil {
			return err
		}
		return u.repo.SalvarFilhos(ctx, tx, id, req)
	})
	if errors.Is(err, ErrDificuldadeInvalida) {
		return nil, apperror.BadRequest("dificuldade inválida")
	}
	if err != nil {
		return nil, interno("erro ao criar projeto", err)
	}
	return u.Detalhe(ctx, usuarioID, id)
}

// Editar substitui o conteúdo. Só o autor edita. Projeto publicado ou oculto
// volta para revisão (ou rascunho, se pedido).
func (u *UseCases) Editar(ctx context.Context, usuarioID, projetoID int, req ProjetoRequest) (*Detalhe, error) {
	if err := u.validarRequest(ctx, &req); err != nil {
		return nil, err
	}

	err := database.EmTransacao(ctx, u.repo.DB(), func(tx *sqlx.Tx) error {
		b, err := u.repo.Base(ctx, tx, projetoID, true)
		if errors.Is(err, sql.ErrNoRows) || (err == nil && b.Status == StatusRemovido) {
			return apperror.NotFound("projeto não encontrado")
		}
		if err != nil {
			return err
		}
		if !b.ehAutor(usuarioID) {
			return apperror.Forbidden("só o autor pode editar o projeto")
		}
		if err := u.repo.Atualizar(ctx, tx, projetoID, req, statusInicial(req)); err != nil {
			return err
		}
		return u.repo.SalvarFilhos(ctx, tx, projetoID, req)
	})
	if errors.Is(err, ErrDificuldadeInvalida) {
		return nil, apperror.BadRequest("dificuldade inválida")
	}
	if err != nil {
		return nil, repassar("erro ao editar projeto", err)
	}
	return u.Detalhe(ctx, usuarioID, projetoID)
}

// Remover: autor ou moderador. Soft delete (status removido).
func (u *UseCases) Remover(ctx context.Context, usuarioID, projetoID int) error {
	b, mod, err := u.baseVisivel(ctx, usuarioID, projetoID)
	if err != nil {
		return err
	}
	if !b.ehAutor(usuarioID) && !mod {
		return apperror.Forbidden("só o autor ou um moderador pode remover o projeto")
	}
	if b.Status == StatusRemovido {
		return nil
	}
	if err := u.repo.Remover(ctx, projetoID); err != nil {
		return interno("erro ao remover projeto", err)
	}
	return nil
}

// ---------- moderação ----------

// Aprovar publica o projeto. Na primeira aprovação o autor ganha XP de publicação
// e avança missões de "publicar_projeto".
func (u *UseCases) Aprovar(ctx context.Context, moderadorID, projetoID int) error {
	err := database.EmTransacao(ctx, u.repo.DB(), func(tx *sqlx.Tx) error {
		b, err := u.repo.Base(ctx, tx, projetoID, true)
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.NotFound("projeto não encontrado")
		}
		if err != nil {
			return err
		}
		if b.Status != StatusRevisao {
			return apperror.Conflict("só projetos em revisão podem ser aprovados")
		}
		if b.ehAutor(moderadorID) {
			return apperror.Forbidden("um moderador não aprova o próprio projeto")
		}

		aprovado, err := u.repo.Aprovar(ctx, tx, projetoID, moderadorID)
		if err != nil {
			return err
		}
		if !aprovado {
			return apperror.Conflict("o projeto mudou de status, recarregue a fila")
		}

		primeiraVez := !b.PublicadoEm.Valid
		if !primeiraVez || !b.AutorID.Valid {
			return nil
		}
		autorID := int(b.AutorID.Int64)
		if _, err := xp.Dar(ctx, tx, autorID, xpPublicacao, xp.ProjetoPublicado, projetoID, "Projeto publicado"); err != nil {
			return err
		}
		cats, err := u.repo.CategoriasDoProjeto(ctx, tx, projetoID)
		if err != nil {
			return err
		}
		_, err = missao.RegistrarAcao(ctx, tx, autorID, missao.AcaoPublicarProjeto, cats)
		return err
	})
	return repassar("erro ao aprovar projeto", err)
}

// Rejeitar tira da fila (ou do ar, se já publicado) com um motivo que o autor vê.
func (u *UseCases) Rejeitar(ctx context.Context, moderadorID, projetoID int, motivo string) error {
	motivo = strings.TrimSpace(motivo)
	if motivo == "" {
		return apperror.BadRequest("informe o motivo para o autor")
	}
	if utf8.RuneCountInString(motivo) > maxMotivo {
		return apperror.BadRequest("motivo deve ter no máximo 500 caracteres")
	}

	b, err := u.repo.Base(ctx, u.repo.DB(), projetoID, false)
	if errors.Is(err, sql.ErrNoRows) {
		return apperror.NotFound("projeto não encontrado")
	}
	if err != nil {
		return interno("erro ao buscar projeto", err)
	}
	if b.ehAutor(moderadorID) {
		return apperror.Forbidden("um moderador não modera o próprio projeto")
	}

	ok, err := u.repo.Rejeitar(ctx, projetoID, moderadorID, motivo)
	if err != nil {
		return interno("erro ao rejeitar projeto", err)
	}
	if !ok {
		return apperror.Conflict("só projetos em revisão ou publicados podem ser rejeitados")
	}
	return nil
}

func (u *UseCases) DefinirDestaque(ctx context.Context, projetoID int) error {
	b, err := u.repo.Base(ctx, u.repo.DB(), projetoID, false)
	if errors.Is(err, sql.ErrNoRows) {
		return apperror.NotFound("projeto não encontrado")
	}
	if err != nil {
		return interno("erro ao buscar projeto", err)
	}
	if b.Status != StatusPublicado {
		return apperror.Conflict("só projetos publicados podem ser destaque")
	}
	if err := u.repo.DefinirDestaque(ctx, projetoID); err != nil {
		return interno("erro ao definir destaque", err)
	}
	return nil
}

// ---------- interações ----------

// Concluir registra que o usuário fez o projeto: dá o XP do projeto e avança
// missões de "concluir_projeto". O autor não ganha XP pelo próprio projeto.
func (u *UseCases) Concluir(ctx context.Context, usuarioID, projetoID int, fotoURL string) (*ConclusaoResponse, error) {
	fotoURL = strings.TrimSpace(fotoURL)
	if fotoURL != "" {
		if err := validarURLImagem(fotoURL); err != nil {
			return nil, apperror.BadRequest(err.Error())
		}
	}

	b, _, err := u.basePublicado(ctx, usuarioID, projetoID)
	if err != nil {
		return nil, err
	}
	if b.ehAutor(usuarioID) {
		return nil, apperror.Forbidden("você não pode concluir o próprio projeto")
	}

	resp := &ConclusaoResponse{MissoesConcluidas: []missao.Concluida{}}
	err = database.EmTransacao(ctx, u.repo.DB(), func(tx *sqlx.Tx) error {
		nova, err := u.repo.InserirConclusao(ctx, tx, projetoID, usuarioID, fotoURL)
		if err != nil {
			return err
		}
		if !nova {
			return apperror.Conflict("você já concluiu este projeto")
		}

		if resp.XPGanho, err = xp.Dar(ctx, tx, usuarioID, b.XP, xp.ProjetoConcluido, projetoID, "Projeto concluído"); err != nil {
			return err
		}

		cats, err := u.repo.CategoriasDoProjeto(ctx, tx, projetoID)
		if err != nil {
			return err
		}
		if resp.MissoesConcluidas, err = missao.RegistrarAcao(ctx, tx, usuarioID, missao.AcaoConcluirProjeto, cats); err != nil {
			return err
		}
		for _, m := range resp.MissoesConcluidas {
			resp.XPGanho += m.XP
		}
		return nil
	})
	if err != nil {
		return nil, repassar("erro ao concluir projeto", err)
	}

	totais, err := xp.DoUsuario(ctx, u.repo.DB(), usuarioID)
	if err != nil {
		return nil, interno("erro ao buscar XP", err)
	}
	resp.Totais = *totais
	return resp, nil
}

func (u *UseCases) Curtir(ctx context.Context, usuarioID, projetoID int, curtir bool) (*CurtidaResponse, error) {
	if _, _, err := u.basePublicado(ctx, usuarioID, projetoID); err != nil {
		return nil, err
	}
	total, err := u.repo.Curtir(ctx, projetoID, usuarioID, curtir)
	if err != nil {
		return nil, interno("erro ao curtir projeto", err)
	}
	return &CurtidaResponse{Curtido: curtir, Curtidas: total}, nil
}

func (u *UseCases) Salvar(ctx context.Context, usuarioID, projetoID int, salvar bool) error {
	// tirar dos salvos funciona mesmo se o projeto saiu do ar
	if salvar {
		if _, _, err := u.basePublicado(ctx, usuarioID, projetoID); err != nil {
			return err
		}
	}
	if err := u.repo.Salvar(ctx, projetoID, usuarioID, salvar); err != nil {
		return interno("erro ao salvar projeto", err)
	}
	return nil
}

func (u *UseCases) Avaliar(ctx context.Context, usuarioID, projetoID, nota int) (*AvaliacaoResponse, error) {
	if nota < 1 || nota > 5 {
		return nil, apperror.BadRequest("a nota deve ser de 1 a 5")
	}
	if _, _, err := u.basePublicado(ctx, usuarioID, projetoID); err != nil {
		return nil, err
	}
	media, total, err := u.repo.Avaliar(ctx, projetoID, usuarioID, nota)
	if errors.Is(err, ErrNaoConcluiu) {
		return nil, apperror.Forbidden("conclua o projeto antes de avaliar")
	}
	if err != nil {
		return nil, interno("erro ao avaliar projeto", err)
	}
	return &AvaliacaoResponse{MinhaNota: nota, NotaMedia: media, Avaliacoes: total}, nil
}

// ---------- comentários ----------

func (u *UseCases) ListarComentarios(ctx context.Context, usuarioID, projetoID, pagina, limite int) (*web.Pagina[Comentario], error) {
	if _, _, err := u.baseVisivel(ctx, usuarioID, projetoID); err != nil {
		return nil, err
	}
	mod, err := u.ehModerador(ctx, usuarioID)
	if err != nil {
		return nil, interno("erro ao verificar permissão", err)
	}

	pagina, limite = web.Paginacao(pagina, limite)
	rows, total, err := u.repo.ListarComentarios(ctx, projetoID, limite, (pagina-1)*limite)
	if err != nil {
		return nil, interno("erro ao listar comentários", err)
	}

	itens := make([]Comentario, 0, len(rows))
	for _, r := range rows {
		itens = append(itens, Comentario{
			ID:         r.ID,
			Conteudo:   r.Conteudo,
			CriadoEm:   r.CriadoEm,
			Autor:      Autor{ID: r.AutorID, Nome: r.AutorNome, FotoURL: web.StrPtr(r.AutorFotoURL)},
			PodeApagar: mod || r.AutorID == usuarioID,
		})
	}
	return &web.Pagina[Comentario]{Itens: itens, Total: total, Pagina: pagina, Limite: limite}, nil
}

func (u *UseCases) Comentar(ctx context.Context, usuarioID, projetoID int, conteudo string) (*Comentario, error) {
	conteudo = strings.TrimSpace(conteudo)
	if conteudo == "" {
		return nil, apperror.BadRequest("o comentário está vazio")
	}
	if utf8.RuneCountInString(conteudo) > maxComentario {
		return nil, apperror.BadRequest("o comentário deve ter no máximo 2000 caracteres")
	}
	if _, _, err := u.basePublicado(ctx, usuarioID, projetoID); err != nil {
		return nil, err
	}

	r, err := u.repo.InserirComentario(ctx, projetoID, usuarioID, conteudo)
	if err != nil {
		return nil, interno("erro ao comentar", err)
	}
	return &Comentario{ID: r.ID, Conteudo: r.Conteudo, CriadoEm: r.CriadoEm,
		Autor:      Autor{ID: usuarioID, Nome: r.AutorNome, FotoURL: web.StrPtr(r.AutorFotoURL)},
		PodeApagar: true}, nil
}

func (u *UseCases) RemoverComentario(ctx context.Context, usuarioID, projetoID, comentarioID int) error {
	c, err := u.repo.Comentario(ctx, comentarioID)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && c.ProjetoID != projetoID) {
		return apperror.NotFound("comentário não encontrado")
	}
	if err != nil {
		return interno("erro ao buscar comentário", err)
	}
	if c.UsuarioID != usuarioID {
		mod, err := u.ehModerador(ctx, usuarioID)
		if err != nil {
			return interno("erro ao verificar permissão", err)
		}
		if !mod {
			return apperror.Forbidden("só o autor do comentário ou um moderador pode apagá-lo")
		}
	}
	if err := u.repo.RemoverComentario(ctx, comentarioID); err != nil {
		return interno("erro ao apagar comentário", err)
	}
	return nil
}

// ---------- erros ----------

func interno(msg string, err error) error {
	log.Printf("[projeto] %s: %v", msg, err)
	return apperror.Internal(msg)
}

// repassar mantém AppError (403, 404, 409...) e transforma o resto em 500.
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
