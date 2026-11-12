package login

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"path"
	"strings"
	"time"

	"ARIC/src/infra/database"
	"ARIC/utils/apperror"
	"ARIC/utils/uploader"

	"github.com/jmoiron/sqlx"
)

// ---------- sessão ----------

// Vale se a conta existe, não foi excluída e o token foi emitido depois da última troca de senha.
const qSessaoValida = `
SELECT deletado_em IS NULL AND (sessoes_validas_desde IS NULL OR sessoes_validas_desde <= $2)
FROM aric.usuarios WHERE id = $1`

func (r *LoginRepository) SessaoValida(ctx context.Context, usuarioID int, emitidoEm time.Time) (bool, error) {
	var valida bool
	err := r.db.GetContext(ctx, &valida, qSessaoValida, usuarioID, emitidoEm)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return valida, err
}

// ---------- exclusão de conta ----------

// ExcluirContaRequest: a senha confirma que é o próprio dono.
// ApagarConteudo escolhe o destino do que a pessoa publicou:
//
//	false: projetos publicados, posts, respostas e comentários continuam, assinados como "Usuário removido"
//	true:  tudo isso é apagado também (e as imagens saem do servidor)
type ExcluirContaRequest struct {
	Senha          string `json:"senha"`
	ApagarConteudo bool   `json:"apagarConteudo"`
}

// Arquivos que saem do servidor: foto de perfil, fotos de missão e de "eu fiz" sempre;
// imagens de projetos não publicados sempre; imagens de posts e de projetos publicados
// só se o usuário pediu para apagar o conteúdo.
const qArquivosDoUsuario = `
SELECT url FROM (
    SELECT foto_url AS url FROM aric.usuarios WHERE id = $1
    UNION ALL
    SELECT ev.foto_url FROM aric.missoes_evidencias ev
    JOIN aric.usuarios_missoes um ON um.id = ev.usuario_missao_id
    WHERE um.usuario_id = $1
    UNION ALL
    SELECT foto_url FROM aric.projetos_conclusoes WHERE usuario_id = $1
    UNION ALL
    SELECT imagem_url FROM aric.forum_posts WHERE autor_id = $1 AND $2::bool
    UNION ALL
    SELECT m.midia_url FROM aric.projetos_midia m JOIN aric.projetos p ON p.id = m.projeto_id
    WHERE p.autor_id = $1 AND ($2::bool OR p.status <> 'publicado')
    UNION ALL
    SELECT mt.imagem_url FROM aric.projetos_materiais mt JOIN aric.projetos p ON p.id = mt.projeto_id
    WHERE p.autor_id = $1 AND ($2::bool OR p.status <> 'publicado')
    UNION ALL
    SELECT ps.imagem_url FROM aric.projetos_passos ps JOIN aric.projetos p ON p.id = ps.projeto_id
    WHERE p.autor_id = $1 AND ($2::bool OR p.status <> 'publicado')
) arquivos
WHERE url IS NOT NULL AND url <> ''`

// Só quando o usuário pede para apagar o conteúdo.
var qsApagarConteudo = []string{
	// resposta aceita apagada: a dúvida volta para "em aberto"
	`UPDATE aric.forum_posts SET resposta_aceita_id = NULL
	 WHERE resposta_aceita_id IN (SELECT id FROM aric.forum_respostas WHERE autor_id = $1)`,
	`UPDATE aric.forum_respostas SET deletado_em = now() WHERE autor_id = $1 AND deletado_em IS NULL`,
	`UPDATE aric.forum_posts SET deletado_em = now() WHERE autor_id = $1 AND deletado_em IS NULL`,
	`UPDATE aric.projetos_comentarios SET deletado_em = now() WHERE usuario_id = $1 AND deletado_em IS NULL`,
	`UPDATE aric.projetos SET status = 'removido', data_atualizacao = now()
	 WHERE autor_id = $1 AND status = 'publicado'`,
}

// Sempre: some o que é privado ou pessoal.
var qsLimparDadosPessoais = []string{
	// rascunhos, projetos em revisão e rejeitados só o autor via
	`UPDATE aric.projetos SET status = 'removido', data_atualizacao = now()
	 WHERE autor_id = $1 AND status IN ('rascunho', 'em_revisao', 'oculto')`,
	`DELETE FROM aric.projetos_curtidas WHERE usuario_id = $1`,
	`DELETE FROM aric.projetos_salvos WHERE usuario_id = $1`,
	`DELETE FROM aric.projetos_avaliacoes WHERE usuario_id = $1`,
	`DELETE FROM aric.forum_posts_curtidas WHERE usuario_id = $1`,
	`DELETE FROM aric.forum_respostas_curtidas WHERE usuario_id = $1`,
	`DELETE FROM aric.forum_posts_salvos WHERE usuario_id = $1`,
	// fotos de missão aguardando análise não precisam mais ser analisadas
	`DELETE FROM aric.missoes_evidencias ev USING aric.usuarios_missoes um
	 WHERE um.id = ev.usuario_missao_id AND um.usuario_id = $1 AND ev.status = 'pendente'`,
	`UPDATE aric.missoes_evidencias ev SET foto_url = '' FROM aric.usuarios_missoes um
	 WHERE um.id = ev.usuario_missao_id AND um.usuario_id = $1`,
	`UPDATE aric.projetos_conclusoes SET foto_url = NULL WHERE usuario_id = $1`,
}

// Anonimiza a linha do usuário. A linha fica para não quebrar o histórico
// (XP, ligas, conclusões que contam no "pessoas que fizeram" e no impacto do bairro),
// mas sem nenhum dado que identifique a pessoa. O e-mail fica livre para um novo cadastro.
const qAnonimizarUsuario = `
UPDATE aric.usuarios
SET nome = 'Usuário removido',
    email = 'removido-' || id || '@aric.invalid',
    senha_hash = '!',
    foto_url = NULL, bio = NULL,
    cep = '00000000', logradouro = '-', numero_endereco = '-', complemento_endereco = NULL,
    papel = 'usuario',
    deletado_em = now(), atualizado_em = now(), sessoes_validas_desde = now()
WHERE id = $1 AND deletado_em IS NULL`

// ExcluirConta confirma a senha, anonimiza a conta e (se pedido) apaga o conteúdo publicado.
// Depois disso todos os tokens do usuário param de valer na hora.
func (u *LoginUseCases) ExcluirConta(ctx context.Context, id int, req ExcluirContaRequest, ip string) error {
	if req.Senha == "" {
		return apperror.BadRequest("informe sua senha para confirmar")
	}
	user, err := u.repo.GetUsuarioByID(ctx, id)
	if err != nil {
		return erroUsuario("erro ao buscar usuário", err)
	}
	if err := u.confirmarSenha(ctx, user, req.Senha, ip); err != nil {
		var appErr *apperror.AppError
		if errors.As(err, &appErr) && appErr.StatusCode == 400 {
			return apperror.BadRequest("senha incorreta")
		}
		return err
	}

	var arquivos []string
	err = database.EmTransacao(ctx, u.repo.db, func(tx *sqlx.Tx) error {
		// a lista sai antes das mudanças, que limpam as URLs
		if err := tx.SelectContext(ctx, &arquivos, qArquivosDoUsuario, id, req.ApagarConteudo); err != nil {
			return err
		}

		queries := qsLimparDadosPessoais
		if req.ApagarConteudo {
			queries = append(append([]string{}, qsApagarConteudo...), qsLimparDadosPessoais...)
		}
		for _, q := range queries {
			if _, err := tx.ExecContext(ctx, q, id); err != nil {
				return err
			}
		}

		res, err := tx.ExecContext(ctx, qAnonimizarUsuario, id)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return ErrUsuarioNaoEncontrado
		}
		return nil
	})
	if err != nil {
		return erroUsuario("erro ao excluir conta", err)
	}

	// arquivos só depois do commit: se a transação falhasse, as fotos continuariam no lugar
	for _, a := range arquivos {
		apagarUpload(a)
	}
	log.Printf("[login] conta %d excluída (apagarConteudo=%v, %d arquivos)", id, req.ApagarConteudo, len(arquivos))
	return nil
}

// apagarUpload remove um arquivo enviado pelo app (qualquer pasta dentro de /uploads),
// recusando caminhos com "../" ou que apontem para fora de /uploads.
func apagarUpload(url string) {
	limpo := path.Clean(url)
	if !strings.HasPrefix(limpo, "/uploads/") || limpo != url || strings.Contains(url, "..") {
		log.Println("[login] caminho de upload ignorado:", url)
		return
	}
	if err := uploader.DeleteFile(limpo); err != nil {
		log.Println("[login] erro ao apagar arquivo:", err)
	}
}
