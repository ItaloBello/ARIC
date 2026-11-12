package denuncia

// Todos os conteúdos denunciáveis numa forma só: tipo, id, autor, se está no ar,
// título, trecho e o "pai" (projeto do comentário, post da resposta) para o app abrir.
const alvos = `
    SELECT 'projeto'::text AS tipo, p.id, p.autor_id, (p.status = 'publicado') AS visivel,
           p.titulo::text AS titulo, left(p.descricao, 300) AS trecho, NULL::int AS pai_id
    FROM aric.projetos p
    UNION ALL
    SELECT 'projeto_comentario', c.id, c.usuario_id, (c.deletado_em IS NULL),
           NULL, left(c.conteudo, 300), c.projeto_id
    FROM aric.projetos_comentarios c
    UNION ALL
    SELECT 'forum_post', f.id, f.autor_id, (f.deletado_em IS NULL),
           f.titulo::text, left(f.conteudo, 300), NULL
    FROM aric.forum_posts f
    UNION ALL
    SELECT 'forum_resposta', r.id, r.autor_id, (r.deletado_em IS NULL),
           NULL, left(r.conteudo, 300), r.post_id
    FROM aric.forum_respostas r`

const qAlvo = `
SELECT autor_id, visivel FROM (` + alvos + `) a
WHERE a.tipo = $1 AND a.id = $2`

const qDenunciasHoje = `
SELECT count(*) FROM aric.denuncias
WHERE denunciante_id = $1 AND criado_em > now() - interval '24 hours'`

// Sem linha de volta = a pessoa já tinha denunciado este conteúdo.
const qInserir = `
INSERT INTO aric.denuncias (denunciante_id, alvo_tipo, alvo_id, motivo, descricao)
VALUES ($1, $2, $3, $4, NULLIF($5, ''))
ON CONFLICT (denunciante_id, alvo_tipo, alvo_id) DO NOTHING
RETURNING id, criado_em`

// Fila agrupada por conteúdo: os mais denunciados primeiro, depois os mais antigos.
const qFila = `
WITH pend AS (
    SELECT alvo_tipo, alvo_id,
           count(*) AS denuncias,
           array_agg(DISTINCT motivo ORDER BY motivo) AS motivos,
           (array_remove(array_agg(descricao ORDER BY criado_em DESC), NULL))[1:3] AS descricoes,
           min(criado_em) AS primeira_em,
           max(criado_em) AS ultima_em
    FROM aric.denuncias
    WHERE status = 'pendente'
    GROUP BY alvo_tipo, alvo_id
)
SELECT pend.alvo_tipo, pend.alvo_id, pend.denuncias, pend.motivos,
       COALESCE(pend.descricoes, '{}') AS descricoes, pend.primeira_em, pend.ultima_em,
       a.titulo, a.trecho, a.pai_id, a.visivel,
       a.autor_id, u.nome AS autor_nome,
       COUNT(*) OVER () AS total
FROM pend
JOIN (` + alvos + `) a ON a.tipo = pend.alvo_tipo AND a.id = pend.alvo_id
LEFT JOIN aric.usuarios u ON u.id = a.autor_id
ORDER BY pend.denuncias DESC, pend.primeira_em
LIMIT $1 OFFSET $2`

// Trava as denúncias pendentes do conteúdo (dois moderadores não resolvem ao mesmo tempo).
const qTravarPendentes = `
SELECT id FROM aric.denuncias
WHERE alvo_tipo = $1 AND alvo_id = $2 AND status = 'pendente'
FOR UPDATE`

const qResolver = `
UPDATE aric.denuncias
SET status = $3, resolvido_por = $4, resolvido_em = now()
WHERE alvo_tipo = $1 AND alvo_id = $2 AND status = 'pendente'`

// ---------- remoção do conteúdo denunciado ----------

// Projeto vai para "oculto" com o motivo, como numa rejeição: o autor vê por quê.
const qRemoverProjeto = `
UPDATE aric.projetos
SET status = 'oculto', motivo_moderacao = $2, moderado_por = $3, moderado_em = now()
WHERE id = $1 AND status IN ('publicado', 'em_revisao')`

const qRemoverComentario = `
UPDATE aric.projetos_comentarios SET deletado_em = now() WHERE id = $1 AND deletado_em IS NULL`

const qRemoverPost = `
UPDATE aric.forum_posts SET deletado_em = now() WHERE id = $1 AND deletado_em IS NULL`

const qRemoverResposta = `
UPDATE aric.forum_respostas SET deletado_em = now() WHERE id = $1 AND deletado_em IS NULL`

// resposta aceita removida: a dúvida volta para "em aberto"
const qDesmarcarAceita = `
UPDATE aric.forum_posts SET resposta_aceita_id = NULL WHERE resposta_aceita_id = $1`
