package forum

// $1 é sempre o usuário logado (para curtido/salvo).
const selectPost = `
SELECT f.id, f.tipo, f.titulo, left(f.conteudo, 200) AS resumo, f.imagem_url,
       f.criado_em, f.atualizado_em, f.resposta_aceita_id,
       f.autor_id, u.nome AS autor_nome, u.foto_url AS autor_foto_url, b.nome AS autor_bairro,
       f.categoria_id, c.nome AS categoria_nome, c.cor AS categoria_cor,
       cont.curtidas, cont.respostas,
       EXISTS (SELECT 1 FROM aric.forum_posts_curtidas x
               WHERE x.post_id = f.id AND x.usuario_id = $1) AS curtido,
       EXISTS (SELECT 1 FROM aric.forum_posts_salvos x
               WHERE x.post_id = f.id AND x.usuario_id = $1) AS salvo,
       COUNT(*) OVER () AS total
FROM aric.forum_posts f
JOIN aric.usuarios u ON u.id = f.autor_id
JOIN aric.bairros b ON b.id = u.bairro_id
LEFT JOIN aric.categorias c ON c.id = f.categoria_id
CROSS JOIN LATERAL (
    SELECT (SELECT count(*) FROM aric.forum_posts_curtidas x WHERE x.post_id = f.id) AS curtidas,
           (SELECT count(*) FROM aric.forum_respostas r
            WHERE r.post_id = f.id AND r.deletado_em IS NULL) AS respostas
) cont
`

// qListar atende as abas do fórum, a busca, "meus posts" e "salvos".
//
//	$1 usuário logado   $2 busca      $3 categoria_id   $4 tipo ('duvida'|'dica')
//	$5 situacao ('abertas'|'resolvidas')                 $6 autor_id
//	$7 só salvos        $8 post_id    $9 ordem          $10 limite   $11 offset
//
// ordem: 'recentes' | 'votadas' | 'relevancia' (precisa de busca)
// "Dúvidas abertas" = tipo duvida sem resposta aceita.
const qListar = selectPost + `
WHERE f.deletado_em IS NULL
  AND ($2::text IS NULL OR f.busca @@ plainto_tsquery('portuguese', $2::text))
  AND ($3::int IS NULL OR f.categoria_id = $3::int)
  AND ($4::text IS NULL OR f.tipo = $4::text)
  AND ($5::text IS NULL
       OR ($5::text = 'abertas' AND f.tipo = 'duvida' AND f.resposta_aceita_id IS NULL)
       OR ($5::text = 'resolvidas' AND f.resposta_aceita_id IS NOT NULL))
  AND ($6::int IS NULL OR f.autor_id = $6::int)
  AND (NOT $7::bool OR EXISTS (SELECT 1 FROM aric.forum_posts_salvos s
                               WHERE s.post_id = f.id AND s.usuario_id = $1))
  AND ($8::int IS NULL OR f.id = $8::int)
ORDER BY
  CASE WHEN $9::text = 'votadas' THEN cont.curtidas END DESC NULLS LAST,
  CASE WHEN $9::text = 'relevancia' AND $2::text IS NOT NULL
       THEN ts_rank(f.busca, plainto_tsquery('portuguese', $2::text)) END DESC NULLS LAST,
  f.criado_em DESC,
  f.id DESC
LIMIT $10 OFFSET $11`

const qConteudo = `SELECT conteudo FROM aric.forum_posts WHERE id = $1`

const qPostBase = `
SELECT id, autor_id, tipo, categoria_id, resposta_aceita_id, (deletado_em IS NOT NULL) AS deletado
FROM aric.forum_posts WHERE id = $1`

const qPostBaseParaAtualizar = qPostBase + ` FOR UPDATE`

const qPapel = `SELECT papel FROM aric.usuarios WHERE id = $1 AND deletado_em IS NULL`

const qCategoriaExiste = `SELECT EXISTS (SELECT 1 FROM aric.categorias WHERE id = $1)`

// ---------- posts ----------

const qInserirPost = `
INSERT INTO aric.forum_posts (autor_id, tipo, titulo, conteudo, imagem_url, categoria_id)
VALUES ($1, $2, $3, $4, NULLIF($5, ''), $6)
RETURNING id`

// O tipo não muda na edição (uma dúvida resolvida não pode virar dica).
const qAtualizarPost = `
UPDATE aric.forum_posts
SET titulo = $2, conteudo = $3, imagem_url = NULLIF($4, ''), categoria_id = $5, atualizado_em = now()
WHERE id = $1 AND deletado_em IS NULL`

const qRemoverPost = `
UPDATE aric.forum_posts SET deletado_em = now() WHERE id = $1 AND deletado_em IS NULL`

const qAceitar = `
UPDATE aric.forum_posts SET resposta_aceita_id = $2 WHERE id = $1`

const (
	qCurtirPost    = `INSERT INTO aric.forum_posts_curtidas (post_id, usuario_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	qDescurtirPost = `DELETE FROM aric.forum_posts_curtidas WHERE post_id = $1 AND usuario_id = $2`
	qCurtidasPost  = `SELECT count(*) FROM aric.forum_posts_curtidas WHERE post_id = $1`

	qSalvarPost   = `INSERT INTO aric.forum_posts_salvos (post_id, usuario_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	qRemoverSalvo = `DELETE FROM aric.forum_posts_salvos WHERE post_id = $1 AND usuario_id = $2`
)

// ---------- respostas ----------

// A aceita vem primeiro, depois as mais curtidas, depois as mais antigas.
const qListarRespostas = `
SELECT r.id, r.conteudo, r.criado_em, r.atualizado_em,
       r.autor_id, u.nome AS autor_nome, u.foto_url AS autor_foto_url, b.nome AS autor_bairro,
       (r.id IS NOT DISTINCT FROM f.resposta_aceita_id) AS aceita,
       cont.curtidas,
       EXISTS (SELECT 1 FROM aric.forum_respostas_curtidas x
               WHERE x.resposta_id = r.id AND x.usuario_id = $1) AS curtido,
       COUNT(*) OVER () AS total
FROM aric.forum_respostas r
JOIN aric.forum_posts f ON f.id = r.post_id
JOIN aric.usuarios u ON u.id = r.autor_id
JOIN aric.bairros b ON b.id = u.bairro_id
CROSS JOIN LATERAL (
    SELECT count(*) AS curtidas FROM aric.forum_respostas_curtidas x WHERE x.resposta_id = r.id
) cont
WHERE r.post_id = $2 AND r.deletado_em IS NULL
  AND ($5::int IS NULL OR r.id = $5::int)
ORDER BY aceita DESC, cont.curtidas DESC, r.criado_em ASC, r.id ASC
LIMIT $3 OFFSET $4`

const qInserirResposta = `
WITH nova AS (
    INSERT INTO aric.forum_respostas (post_id, autor_id, conteudo)
    VALUES ($1, $2, $3)
    RETURNING id, criado_em, atualizado_em, autor_id
)
SELECT n.id, n.criado_em, n.atualizado_em, u.nome AS autor_nome, u.foto_url AS autor_foto_url,
       b.nome AS autor_bairro
FROM nova n
JOIN aric.usuarios u ON u.id = n.autor_id
JOIN aric.bairros b ON b.id = u.bairro_id`

const qRespostaBase = `
SELECT id, post_id, autor_id FROM aric.forum_respostas WHERE id = $1 AND deletado_em IS NULL`

const qAtualizarResposta = `
UPDATE aric.forum_respostas SET conteudo = $2, atualizado_em = now()
WHERE id = $1 AND deletado_em IS NULL
RETURNING atualizado_em`

const qRemoverResposta = `
UPDATE aric.forum_respostas SET deletado_em = now() WHERE id = $1 AND deletado_em IS NULL`

// Resposta apagada deixa de ser a aceita (a dúvida volta para "em aberto").
const qDesmarcarAceitaApagada = `
UPDATE aric.forum_posts SET resposta_aceita_id = NULL WHERE resposta_aceita_id = $1`

const (
	qCurtirResposta    = `INSERT INTO aric.forum_respostas_curtidas (resposta_id, usuario_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	qDescurtirResposta = `DELETE FROM aric.forum_respostas_curtidas WHERE resposta_id = $1 AND usuario_id = $2`
	qCurtidasResposta  = `SELECT count(*) FROM aric.forum_respostas_curtidas WHERE resposta_id = $1`
)
