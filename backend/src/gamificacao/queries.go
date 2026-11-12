package gamificacao

// ---------- perfil ----------

// Tudo que as telas Perfil e Missões mostram de um usuário.
// sequencia_atual já vem zerada se o usuário pulou um dia (o trigger só zera na próxima ação).
const qPerfil = `
SELECT u.id, u.nome, u.foto_url, u.bio, b.nome AS bairro, c.nome AS cidade, e.sigla AS uf,
       u.xp_total, u.criado_em,
       CASE WHEN u.ultima_atividade_em >= (now() AT TIME ZONE 'America/Sao_Paulo')::date - 1
            THEN u.sequencia_dias ELSE 0 END AS sequencia_atual,
       n.numero AS nivel_numero, n.titulo AS nivel_titulo, n.minimo_xp AS nivel_minimo_xp,
       prox.numero AS prox_numero, prox.titulo AS prox_titulo, prox.minimo_xp AS prox_minimo_xp,
       l.id AS liga_id, l.nome AS liga_nome, l.ordem AS liga_ordem, l.icone AS liga_icone,
       imp.residuos_kg, imp.projetos_concluidos, pub.projetos_publicados, insp.vidas_inspiradas
FROM aric.usuarios u
JOIN aric.bairros b ON b.id = u.bairro_id
JOIN aric.cidades c ON c.id = b.cidade_id
JOIN aric.estados e ON e.id = c.estado_id
LEFT JOIN aric.niveis n ON n.id = u.nivel_id
LEFT JOIN LATERAL (
    SELECT x.numero, x.titulo, x.minimo_xp FROM aric.niveis x
    WHERE x.minimo_xp > u.xp_total ORDER BY x.minimo_xp LIMIT 1
) prox ON TRUE
LEFT JOIN aric.ligas l ON l.id = u.liga_id
CROSS JOIN LATERAL (
    SELECT COALESCE(SUM(p.residuos_kg), 0)::float8 AS residuos_kg, count(*) AS projetos_concluidos
    FROM aric.projetos_conclusoes pc JOIN aric.projetos p ON p.id = pc.projeto_id
    WHERE pc.usuario_id = u.id
) imp
CROSS JOIN LATERAL (
    SELECT count(*) AS projetos_publicados FROM aric.projetos p
    WHERE p.autor_id = u.id AND p.status = 'publicado'
) pub
CROSS JOIN LATERAL (
    -- pessoas diferentes que fizeram projetos deste usuário
    SELECT count(DISTINCT pc.usuario_id) AS vidas_inspiradas
    FROM aric.projetos_conclusoes pc JOIN aric.projetos p ON p.id = pc.projeto_id
    WHERE p.autor_id = u.id
) insp
WHERE u.id = $1 AND u.deletado_em IS NULL`

const qPosicaoNaLiga = `
SELECT posicao, xp_semana, tendencia, zona
FROM aric.ranking_liga($1)
WHERE usuario_id = $2`

const qNiveis = `
SELECT id, numero, titulo, minimo_xp FROM aric.niveis ORDER BY minimo_xp`

const qHistoricoXP = `
SELECT id, quantidade, origem, referencia_id, descricao, criado_em, COUNT(*) OVER () AS total
FROM aric.eventos_xp
WHERE usuario_id = $1
ORDER BY criado_em DESC, id DESC
LIMIT $2 OFFSET $3`

// ---------- ligas ----------

const qLigas = `
SELECT id, ordem, nome, icone, qtd_sobe, qtd_desce FROM aric.ligas ORDER BY ordem`

const qLigaDoUsuario = `
SELECT l.id, l.ordem, l.nome, l.icone, l.qtd_sobe, l.qtd_desce,
       NOT EXISTS (SELECT 1 FROM aric.ligas x WHERE x.ordem > l.ordem) AS ultima,
       NOT EXISTS (SELECT 1 FROM aric.ligas x WHERE x.ordem < l.ordem) AS primeira,
       ((aric.inicio_semana(now()) + 7)::timestamp AT TIME ZONE 'America/Sao_Paulo') AS fim_semana
FROM aric.usuarios u JOIN aric.ligas l ON l.id = u.liga_id
WHERE u.id = $1`

const qRankingLiga = `
SELECT posicao, usuario_id, nome, foto_url, bairro, xp_semana, tendencia, zona
FROM aric.ranking_liga($1)`

// Resultado da última semana fechada do usuário (para "Você subiu para a Liga Muda!").
const qUltimaSemana = `
SELECT ls.semana_inicio, ls.xp_semana, ls.posicao, ls.resultado,
       la.nome AS liga_anterior, ln.nome AS liga_nova
FROM aric.ligas_semanas ls
JOIN aric.ligas la ON la.id = ls.liga_id
JOIN aric.ligas ln ON ln.id = ls.liga_seguinte_id
WHERE ls.usuario_id = $1
ORDER BY ls.semana_inicio DESC
LIMIT 1`

// ---------- missões ----------

// Missões ativas com o progresso do usuário na semana atual. Concluídas vão para o fim.
const qMissoesDaSemana = `
SELECT m.id, m.titulo, m.descricao, m.imagem_url, m.acao, m.meta, m.unidade, m.xp,
       c.id AS categoria_id, c.nome AS categoria_nome, c.cor AS categoria_cor,
       COALESCE(um.progresso, 0) AS progresso, um.concluida_em,
       (SELECT count(*) FROM aric.missoes_evidencias ev
        WHERE ev.usuario_missao_id = um.id AND ev.status = 'pendente') AS pendentes,
       ((aric.inicio_semana(now()) + 7)::timestamp AT TIME ZONE 'America/Sao_Paulo') AS fim_semana
FROM aric.missoes m
LEFT JOIN aric.categorias c ON c.id = m.categoria_id
LEFT JOIN aric.usuarios_missoes um
       ON um.missao_id = m.id AND um.usuario_id = $1
      AND um.semana_inicio = aric.inicio_semana(now())
WHERE m.ativa
ORDER BY (um.concluida_em IS NOT NULL), m.id`

const qMissao = `
SELECT id, titulo, acao, meta, xp, ativa FROM aric.missoes WHERE id = $1`

// Garante a linha de progresso da semana atual e devolve o estado dela.
const qProgressoDaSemana = `
INSERT INTO aric.usuarios_missoes AS um (usuario_id, missao_id, semana_inicio)
VALUES ($1, $2, aric.inicio_semana(now()))
ON CONFLICT (usuario_id, missao_id, semana_inicio) DO UPDATE SET atualizado_em = um.atualizado_em
RETURNING id, progresso, concluida_em,
          (SELECT count(*) FROM aric.missoes_evidencias ev
           WHERE ev.usuario_missao_id = um.id AND ev.status = 'pendente') AS pendentes`

const qInserirEvidencia = `
INSERT INTO aric.missoes_evidencias (usuario_missao_id, foto_url, quantidade)
VALUES ($1, $2, $3)
RETURNING id, status, criado_em`

const qMinhasEvidencias = `
SELECT ev.id, ev.foto_url, ev.quantidade, ev.status, ev.motivo_rejeicao, ev.criado_em, ev.avaliado_em,
       m.id AS missao_id, m.titulo AS missao_titulo, um.semana_inicio,
       COUNT(*) OVER () AS total
FROM aric.missoes_evidencias ev
JOIN aric.usuarios_missoes um ON um.id = ev.usuario_missao_id
JOIN aric.missoes m ON m.id = um.missao_id
WHERE um.usuario_id = $1
ORDER BY ev.criado_em DESC, ev.id DESC
LIMIT $2 OFFSET $3`

// ---------- moderação de evidências ----------

const qEvidenciasPendentes = `
SELECT ev.id, ev.foto_url, ev.quantidade, ev.criado_em,
       um.semana_inicio, um.progresso,
       m.id AS missao_id, m.titulo AS missao_titulo, m.meta, m.unidade,
       u.id AS usuario_id, u.nome AS usuario_nome, u.foto_url AS usuario_foto_url,
       COUNT(*) OVER () AS total
FROM aric.missoes_evidencias ev
JOIN aric.usuarios_missoes um ON um.id = ev.usuario_missao_id
JOIN aric.missoes m ON m.id = um.missao_id
JOIN aric.usuarios u ON u.id = um.usuario_id
WHERE ev.status = 'pendente'
ORDER BY ev.criado_em, ev.id
LIMIT $1 OFFSET $2`

const qEvidenciaParaAvaliar = `
SELECT ev.id, ev.status, ev.quantidade, ev.usuario_missao_id, um.usuario_id
FROM aric.missoes_evidencias ev
JOIN aric.usuarios_missoes um ON um.id = ev.usuario_missao_id
WHERE ev.id = $1
FOR UPDATE OF ev`

const qMarcarEvidencia = `
UPDATE aric.missoes_evidencias
SET status = $2, motivo_rejeicao = $3, avaliado_por = $4, avaliado_em = now()
WHERE id = $1 AND status = 'pendente'`

// Soma a quantidade aprovada no progresso da semana em que a foto foi enviada
// (mesmo que a aprovação aconteça na semana seguinte).
const qSomarProgresso = `
UPDATE aric.usuarios_missoes um
SET progresso = um.progresso + $2, atualizado_em = now()
FROM aric.missoes m
WHERE m.id = um.missao_id AND um.id = $1
RETURNING um.id, um.usuario_id, um.progresso, um.concluida_em, m.meta, m.xp, m.titulo, m.id AS missao_id`

const qConcluirProgresso = `
UPDATE aric.usuarios_missoes SET concluida_em = now() WHERE id = $1 AND concluida_em IS NULL`

// ---------- fechamento semanal ----------

const qSemanaParaFechar = `
SELECT aric.inicio_semana(now()) - 7 AS semana,
       EXISTS (SELECT 1 FROM aric.ligas_semanas
               WHERE semana_inicio = aric.inicio_semana(now()) - 7) AS fechada`

const qFecharSemana = `SELECT aric.fechar_semana_ligas($1::date)`

// ---------- impacto do bairro (Home) ----------

// Resíduos reaproveitados pelos moradores do bairro do usuário: hoje, nesta semana e no total.
// "Hoje" e "semana" seguem o horário de São Paulo.
const qImpactoBairro = `
WITH limites AS (
    SELECT (now() AT TIME ZONE 'America/Sao_Paulo')::date::timestamp
               AT TIME ZONE 'America/Sao_Paulo' AS inicio_hoje,
           aric.inicio_semana(now())::timestamp AT TIME ZONE 'America/Sao_Paulo' AS inicio_semana
)
SELECT b.id AS bairro_id, b.nome AS bairro,
       COALESCE(SUM(p.residuos_kg) FILTER (WHERE pc.data_conclusao >= l.inicio_hoje), 0)::float8 AS kg_hoje,
       count(pc.usuario_id) FILTER (WHERE pc.data_conclusao >= l.inicio_hoje) AS projetos_hoje,
       count(DISTINCT pc.usuario_id) FILTER (WHERE pc.data_conclusao >= l.inicio_hoje) AS pessoas_hoje,
       COALESCE(SUM(p.residuos_kg) FILTER (WHERE pc.data_conclusao >= l.inicio_semana), 0)::float8 AS kg_semana,
       COALESCE(SUM(p.residuos_kg), 0)::float8 AS kg_total
FROM aric.usuarios eu
JOIN aric.bairros b ON b.id = eu.bairro_id
CROSS JOIN limites l
LEFT JOIN aric.usuarios u ON u.bairro_id = b.id AND u.deletado_em IS NULL
LEFT JOIN aric.projetos_conclusoes pc ON pc.usuario_id = u.id
LEFT JOIN aric.projetos p ON p.id = pc.projeto_id
WHERE eu.id = $1
GROUP BY b.id, b.nome`
