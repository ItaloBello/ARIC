package gamificacao

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

const (
	// linhas do ranking devolvidas (a posição do usuário vem à parte se ele estiver abaixo)
	limiteRanking = 50
	// fotos aguardando análise por missão; evita mandar várias de uma vez
	maxPendentes = 3
	maxMotivo    = 300
)

var configImagem = uploader.UploadConfig{
	DestFolder:   "./uploads/missoes",
	BaseURL:      "/uploads/missoes",
	MaxSize:      5 * 1024 * 1024, // 5MB
	AllowedTypes: []string{".jpg", ".jpeg", ".png", ".webp"},
}

type UseCases struct {
	repo *Repository
}

func NewUseCases(repo *Repository) *UseCases {
	return &UseCases{repo: repo}
}

// ---------- perfil ----------

// Perfil serve tanto para o próprio usuário quanto para o perfil público de outro.
// Não tem e-mail nem endereço.
func (u *UseCases) Perfil(ctx context.Context, usuarioID int) (*Perfil, error) {
	row, err := u.repo.Perfil(ctx, usuarioID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.NotFound("usuário não encontrado")
	}
	if err != nil {
		return nil, interno("erro ao buscar perfil", err)
	}

	p := &Perfil{
		ID:            row.ID,
		Nome:          row.Nome,
		FotoURL:       web.StrPtr(row.FotoURL),
		Bio:           web.StrPtr(row.Bio),
		Bairro:        row.Bairro,
		Cidade:        row.Cidade,
		UF:            row.UF,
		MembroDesde:   row.CriadoEm,
		XPTotal:       row.XPTotal,
		SequenciaDias: row.SequenciaAtual,
		Impacto: Impacto{
			ResiduosKg:         row.ResiduosKg,
			ProjetosConcluidos: row.ProjetosConcluidos,
			ProjetosPublicados: row.ProjetosPublicados,
			VidasInspiradas:    row.VidasInspiradas,
		},
	}

	if row.NivelNumero.Valid {
		p.Nivel = &Nivel{Numero: int(row.NivelNumero.Int64), Titulo: row.NivelTitulo.String, MinimoXP: int(row.NivelMinimoXP.Int64)}
	}
	if row.ProxNumero.Valid {
		p.ProximoNivel = &Nivel{Numero: int(row.ProxNumero.Int64), Titulo: row.ProxTitulo.String, MinimoXP: int(row.ProxMinimoXP.Int64)}
	}
	p.XPParaProximo, p.ProgressoNivel = progressoNivel(row.XPTotal, p.Nivel, p.ProximoNivel)

	if row.LigaID.Valid {
		p.Liga = &LigaResumo{
			ID:    int(row.LigaID.Int64),
			Nome:  row.LigaNome.String,
			Ordem: int(row.LigaOrdem.Int64),
			Icone: web.StrPtr(row.LigaIcone),
		}
		pos, err := u.repo.PosicaoNaLiga(ctx, p.Liga.ID, usuarioID)
		if err != nil {
			return nil, interno("erro ao buscar posição na liga", err)
		}
		if pos != nil {
			p.Liga.Posicao = &pos.Posicao
			p.Liga.XPSemana = pos.XPSemana
		}
	}
	return p, nil
}

// progressoNivel calcula quanto falta e a porcentagem da barra entre o nível atual e o próximo.
func progressoNivel(xpTotal int, atual, proximo *Nivel) (falta, percentual int) {
	if proximo == nil {
		return 0, 100 // nível máximo
	}
	base := 0
	if atual != nil {
		base = atual.MinimoXP
	}
	falta = proximo.MinimoXP - xpTotal
	if falta < 0 {
		falta = 0
	}
	faixa := proximo.MinimoXP - base
	if faixa <= 0 {
		return falta, 100
	}
	percentual = (xpTotal - base) * 100 / faixa
	return falta, max(0, min(100, percentual))
}

func (u *UseCases) Niveis(ctx context.Context) ([]NivelCatalogo, error) {
	n, err := u.repo.Niveis(ctx)
	if err != nil {
		return nil, interno("erro ao listar níveis", err)
	}
	return n, nil
}

func (u *UseCases) HistoricoXP(ctx context.Context, usuarioID, pagina, limite int) (*web.Pagina[EventoXP], error) {
	pagina, limite = web.Paginacao(pagina, limite)
	rows, err := u.repo.HistoricoXP(ctx, usuarioID, limite, (pagina-1)*limite)
	if err != nil {
		return nil, interno("erro ao buscar histórico de XP", err)
	}
	itens := make([]EventoXP, 0, len(rows))
	total := 0
	for _, r := range rows {
		ev := EventoXP{ID: r.ID, Quantidade: r.Quantidade, Origem: r.Origem, Descricao: web.StrPtr(r.Descricao), CriadoEm: r.CriadoEm}
		if r.ReferenciaID.Valid {
			ev.ReferenciaID = &r.ReferenciaID.Int64
		}
		itens = append(itens, ev)
		total = r.Total
	}
	return &web.Pagina[EventoXP]{Itens: itens, Total: total, Pagina: pagina, Limite: limite}, nil
}

// ---------- ligas ----------

func (u *UseCases) Ligas(ctx context.Context) ([]Liga, error) {
	l, err := u.repo.Ligas(ctx)
	if err != nil {
		return nil, interno("erro ao listar ligas", err)
	}
	return l, nil
}

// Ranking monta a tela Comunidade: top 50 da liga do usuário, a posição dele
// e o resultado da última semana fechada.
func (u *UseCases) Ranking(ctx context.Context, usuarioID int) (*Ranking, error) {
	liga, err := u.repo.LigaDoUsuario(ctx, usuarioID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.NotFound("usuário sem liga")
	}
	if err != nil {
		return nil, interno("erro ao buscar liga", err)
	}

	// a liga inteira vem numa consulta só: corta o top 50 e acha o usuário
	todos, err := u.repo.RankingLiga(ctx, liga.ID)
	if err != nil {
		return nil, interno("erro ao montar ranking", err)
	}

	r := &Ranking{
		Liga:        liga.Liga,
		Primeira:    liga.Primeira,
		Ultima:      liga.Ultima,
		FimSemana:   liga.FimSemana,
		TotalNaLiga: len(todos),
	}
	for i := range todos {
		if todos[i].UsuarioID == usuarioID {
			todos[i].Eu = true
			eu := todos[i]
			r.MinhaPosicao = &eu
		}
	}
	r.Itens = todos[:min(len(todos), limiteRanking)]

	if r.UltimaSemana, err = u.repo.UltimaSemana(ctx, usuarioID); err != nil {
		return nil, interno("erro ao buscar semana anterior", err)
	}
	return r, nil
}

// ---------- missões ----------

func (u *UseCases) MissoesDaSemana(ctx context.Context, usuarioID int) (*MissoesResponse, error) {
	rows, err := u.repo.MissoesDaSemana(ctx, usuarioID)
	if err != nil {
		return nil, interno("erro ao listar missões", err)
	}

	resp := &MissoesResponse{Missoes: make([]MissaoSemana, 0, len(rows))}
	for _, r := range rows {
		resp.FimSemana = r.FimSemana
		m := MissaoSemana{
			ID:             r.ID,
			Titulo:         r.Titulo,
			Descricao:      r.Descricao,
			ImagemURL:      web.StrPtr(r.ImagemURL),
			Meta:           r.Meta,
			Unidade:        r.Unidade,
			XP:             r.XP,
			Progresso:      min(r.Progresso, r.Meta),
			Concluida:      r.ConcluidaEm.Valid,
			EnviaFoto:      r.Acao == AcaoRegistroManual,
			FotosEmAnalise: r.Pendentes,
		}
		m.Percentual = m.Progresso * 100 / m.Meta
		if r.ConcluidaEm.Valid {
			m.ConcluidaEm = &r.ConcluidaEm.Time
			m.Progresso, m.Percentual = m.Meta, 100
		}
		if r.CategoriaID.Valid {
			m.Categoria = &Categoria{ID: int(r.CategoriaID.Int64), Nome: r.CategoriaNome.String, Cor: web.StrPtr(r.CategoriaCor)}
		}
		resp.Missoes = append(resp.Missoes, m)
	}
	return resp, nil
}

func (u *UseCases) EnviarImagem(arquivo *multipart.FileHeader) (string, error) {
	res, err := uploader.SaveFile(arquivo, configImagem)
	if err != nil {
		return "", apperror.BadRequest(err.Error())
	}
	return res.URL, nil
}

// EnviarEvidencia registra a foto de uma missão manual (ex.: garrafas entregues).
// O progresso só anda quando um moderador aprovar.
func (u *UseCases) EnviarEvidencia(ctx context.Context, usuarioID, missaoID int, req EvidenciaRequest) (*Evidencia, error) {
	req.FotoURL = strings.TrimSpace(req.FotoURL)
	if req.FotoURL == "" {
		return nil, apperror.BadRequest("envie a foto da missão")
	}
	if err := uploader.URLLocal(req.FotoURL, configImagem); err != nil {
		return nil, apperror.BadRequest(err.Error())
	}

	m, err := u.repo.Missao(ctx, missaoID)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && !m.Ativa) {
		return nil, apperror.NotFound("missão não encontrada")
	}
	if err != nil {
		return nil, interno("erro ao buscar missão", err)
	}
	if m.Acao != AcaoRegistroManual {
		return nil, apperror.Conflict("esta missão avança sozinha, não precisa de foto")
	}
	if req.Quantidade == 0 {
		req.Quantidade = 1
	}
	if req.Quantidade < 1 || req.Quantidade > m.Meta {
		return nil, apperror.BadRequest("quantidade deve ser entre 1 e a meta da missão")
	}

	var ev *evidenciaCriadaRow
	err = database.EmTransacao(ctx, u.repo.DB(), func(tx *sqlx.Tx) error {
		p, err := u.repo.ProgressoDaSemana(ctx, tx, usuarioID, missaoID)
		if err != nil {
			return err
		}
		if p.ConcluidaEm.Valid {
			return apperror.Conflict("você já concluiu esta missão nesta semana")
		}
		if p.Pendentes >= maxPendentes {
			return apperror.Conflict("aguarde a análise das fotos enviadas")
		}
		ev, err = u.repo.InserirEvidencia(ctx, tx, p.ID, req.FotoURL, req.Quantidade)
		return err
	})
	if err != nil {
		return nil, repassar("erro ao enviar foto", err)
	}

	return &Evidencia{
		ID: ev.ID, MissaoID: missaoID, MissaoTitulo: m.Titulo, FotoURL: req.FotoURL,
		Quantidade: req.Quantidade, Status: ev.Status, CriadoEm: ev.CriadoEm,
	}, nil
}

func (u *UseCases) MinhasEvidencias(ctx context.Context, usuarioID, pagina, limite int) (*web.Pagina[Evidencia], error) {
	pagina, limite = web.Paginacao(pagina, limite)
	rows, err := u.repo.MinhasEvidencias(ctx, usuarioID, limite, (pagina-1)*limite)
	if err != nil {
		return nil, interno("erro ao listar fotos enviadas", err)
	}
	itens := make([]Evidencia, 0, len(rows))
	total := 0
	for _, r := range rows {
		semana := r.SemanaInicio
		e := Evidencia{
			ID: r.ID, MissaoID: r.MissaoID, MissaoTitulo: r.MissaoTitulo, FotoURL: r.FotoURL,
			Quantidade: r.Quantidade, Status: r.Status, MotivoRejeicao: web.StrPtr(r.MotivoRejeicao),
			SemanaInicio: &semana, CriadoEm: r.CriadoEm,
		}
		if r.AvaliadoEm.Valid {
			e.AvaliadoEm = &r.AvaliadoEm.Time
		}
		itens = append(itens, e)
		total = r.Total
	}
	return &web.Pagina[Evidencia]{Itens: itens, Total: total, Pagina: pagina, Limite: limite}, nil
}

// ---------- moderação de evidências ----------

func (u *UseCases) EvidenciasPendentes(ctx context.Context, pagina, limite int) (*web.Pagina[EvidenciaPendente], error) {
	pagina, limite = web.Paginacao(pagina, limite)
	rows, err := u.repo.EvidenciasPendentes(ctx, limite, (pagina-1)*limite)
	if err != nil {
		return nil, interno("erro ao listar fotos pendentes", err)
	}
	itens := make([]EvidenciaPendente, 0, len(rows))
	total := 0
	for _, r := range rows {
		e := EvidenciaPendente{
			ID: r.ID, FotoURL: r.FotoURL, Quantidade: r.Quantidade, CriadoEm: r.CriadoEm,
			SemanaInicio: r.SemanaInicio, Progresso: r.Progresso,
		}
		e.Missao.ID, e.Missao.Titulo, e.Missao.Meta, e.Missao.Unidade = r.MissaoID, r.MissaoTitulo, r.Meta, r.Unidade
		e.Usuario.ID, e.Usuario.Nome, e.Usuario.FotoURL = r.UsuarioID, r.UsuarioNome, web.StrPtr(r.UsuarioFotoURL)
		itens = append(itens, e)
		total = r.Total
	}
	return &web.Pagina[EvidenciaPendente]{Itens: itens, Total: total, Pagina: pagina, Limite: limite}, nil
}

// avaliar abre a transação, trava a evidência e checa se ainda está pendente.
func (u *UseCases) avaliar(ctx context.Context, moderadorID, evidenciaID int, fn func(tx *sqlx.Tx, ev *evidenciaAvaliarRow) error) error {
	err := database.EmTransacao(ctx, u.repo.DB(), func(tx *sqlx.Tx) error {
		ev, err := u.repo.EvidenciaParaAvaliar(ctx, tx, evidenciaID)
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.NotFound("foto não encontrada")
		}
		if err != nil {
			return err
		}
		if ev.Status != "pendente" {
			return apperror.Conflict("esta foto já foi analisada")
		}
		if ev.UsuarioID == moderadorID {
			return apperror.Forbidden("um moderador não analisa a própria foto")
		}
		return fn(tx, ev)
	})
	return repassar("erro ao analisar foto", err)
}

// AprovarEvidencia soma a quantidade no progresso da semana em que a foto foi enviada.
// Se bater a meta, conclui a missão e dá o XP.
func (u *UseCases) AprovarEvidencia(ctx context.Context, moderadorID, evidenciaID int) (*AprovacaoResponse, error) {
	resp := &AprovacaoResponse{}
	err := u.avaliar(ctx, moderadorID, evidenciaID, func(tx *sqlx.Tx, ev *evidenciaAvaliarRow) error {
		if err := u.repo.MarcarEvidencia(ctx, tx, ev.ID, "aprovada", nil, moderadorID); err != nil {
			return err
		}
		p, err := u.repo.SomarProgresso(ctx, tx, ev.UsuarioMissaoID, ev.Quantidade)
		if err != nil {
			return err
		}
		resp.Progresso = min(p.Progresso, p.Meta)

		if p.ConcluidaEm.Valid || p.Progresso < p.Meta {
			return nil
		}
		if err := u.repo.ConcluirProgresso(ctx, tx, p.ID); err != nil {
			return err
		}
		if _, err := xp.Dar(ctx, tx, p.UsuarioID, p.XP, xp.MissaoConcluida, p.ID, "Missão: "+p.Titulo); err != nil {
			return err
		}
		resp.MissaoConcluida = &missao.Concluida{ID: p.MissaoID, Titulo: p.Titulo, XP: p.XP}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (u *UseCases) RejeitarEvidencia(ctx context.Context, moderadorID, evidenciaID int, motivo string) error {
	motivo = strings.TrimSpace(motivo)
	if motivo == "" {
		return apperror.BadRequest("informe o motivo para o usuário")
	}
	if utf8.RuneCountInString(motivo) > maxMotivo {
		return apperror.BadRequest("motivo deve ter no máximo 300 caracteres")
	}
	return u.avaliar(ctx, moderadorID, evidenciaID, func(tx *sqlx.Tx, ev *evidenciaAvaliarRow) error {
		return u.repo.MarcarEvidencia(ctx, tx, ev.ID, "rejeitada", motivo, moderadorID)
	})
}

// ---------- erros ----------

func interno(msg string, err error) error {
	log.Printf("[gamificacao] %s: %v", msg, err)
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
