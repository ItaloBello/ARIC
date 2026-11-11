package projeto

// As queries ficam aqui, separadas do repo, para ficar fácil de ler e de testar no psql.

// Colunas e joins do card de projeto (lista, destaque e cabeçalho do detalhe).
// $1 é sempre o usuário logado (para curtido/salvo/concluído).
const selectCard = `
SELECT p.id, p.titulo, p.subtitulo, p.status, p.motivo_moderacao, p.xp, p.tempo_estimado,
       p.residuos_kg, p.dificuldade AS dificuldade_id, d.descricao AS dificuldade,
       p.data_criacao, p.data_atualizacao, p.publicado_em,
       capa.midia_url AS capa_url,
       p.autor_id, au.nome AS autor_nome, au.foto_url AS autor_foto_url, ab.nome AS autor_bairro,
       COALESCE(cat.categorias, '[]') AS categorias,
       cont.curtidas, cont.conclusoes, cont.nota_media,
       EXISTS (SELECT 1 FROM aric.projetos_curtidas x
               WHERE x.projeto_id = p.id AND x.usuario_id = $1) AS curtido,
       EXISTS (SELECT 1 FROM aric.projetos_salvos x
               WHERE x.projeto_id = p.id AND x.usuario_id = $1) AS salvo,
       EXISTS (SELECT 1 FROM aric.projetos_conclusoes x
               WHERE x.projeto_id = p.id AND x.usuario_id = $1) AS concluido,
       COUNT(*) OVER () AS total
FROM aric.projetos p
JOIN aric.dificuldades_projetos d ON d.id = p.dificuldade
LEFT JOIN aric.usuarios au ON au.id = p.autor_id
LEFT JOIN aric.bairros ab ON ab.id = au.bairro_id
LEFT JOIN LATERAL (
    SELECT m.midia_url FROM aric.projetos_midia m
    WHERE m.projeto_id = p.id ORDER BY m.ordem LIMIT 1
) capa ON TRUE
LEFT JOIN LATERAL (
    SELECT json_agg(json_build_object('id', c.id, 'nome', c.nome, 'cor', c.cor) ORDER BY c.nome) AS categorias
    FROM aric.projetos_categorias pc
    JOIN aric.categorias c ON c.id = pc.categoria_id
    WHERE pc.projeto_id = p.id
) cat ON TRUE
CROSS JOIN LATERAL (
    SELECT (SELECT count(*) FROM aric.projetos_curtidas x WHERE x.projeto_id = p.id) AS curtidas,
           (SELECT count(*) FROM aric.projetos_conclusoes x WHERE x.projeto_id = p.id) AS conclusoes,
           (SELECT round(avg(a.nota), 1)::float8 FROM aric.projetos_avaliacoes a
            WHERE a.projeto_id = p.id) AS nota_media
) cont
`

// qListar atende a Home, a busca, "meus projetos", "salvos" e a fila de moderação.
// Filtros opcionais são NULL quando não usados.
//
//	$1  usuário logado         $2  busca (texto)          $3  categoria_id
//	$4  dificuldade_id         $5  autor_id               $6  bairro_id do autor
//	$7  cidade_id do autor     $8  ordem                  $9  limite
//	$10 offset                 $11 status permitidos      $12 só salvos pelo usuário
//	$13 projeto_id
//
// ordem: 'recentes' | 'populares' | 'relevancia' (precisa de busca) | 'antigos' (fila de moderação)
const qListar = selectCard + `
WHERE p.status = ANY($11::text[])
  AND ($2::text IS NULL OR p.busca @@ plainto_tsquery('portuguese', $2::text))
  AND ($3::int IS NULL OR EXISTS (SELECT 1 FROM aric.projetos_categorias pc
                                  WHERE pc.projeto_id = p.id AND pc.categoria_id = $3::int))
  AND ($4::int IS NULL OR p.dificuldade = $4::int)
  AND ($5::int IS NULL OR p.autor_id = $5::int)
  AND ($6::int IS NULL OR au.bairro_id = $6::int)
  AND ($7::int IS NULL OR ab.cidade_id = $7::int)
  AND (NOT $12::bool OR EXISTS (SELECT 1 FROM aric.projetos_salvos s
                                WHERE s.projeto_id = p.id AND s.usuario_id = $1))
  AND ($13::int IS NULL OR p.id = $13::int)
ORDER BY
  CASE WHEN $8::text = 'populares' THEN cont.curtidas END DESC NULLS LAST,
  CASE WHEN $8::text = 'relevancia' AND $2::text IS NOT NULL
       THEN ts_rank(p.busca, plainto_tsquery('portuguese', $2::text)) END DESC NULLS LAST,
  CASE WHEN $8::text = 'antigos' THEN p.data_atualizacao END ASC NULLS LAST,
  COALESCE(p.publicado_em, p.data_criacao) DESC,
  p.id DESC
LIMIT $9 OFFSET $10`

const qMateriais = `
SELECT material AS nome, quantidade, imagem_url
FROM aric.projetos_materiais
WHERE projeto_id = $1
ORDER BY ordem, material`

const qPassos = `
SELECT passo_numero AS numero, titulo, descricao, imagem_url
FROM aric.projetos_passos
WHERE projeto_id = $1
ORDER BY passo_numero`

const qImagens = `
SELECT midia_url
FROM aric.projetos_midia
WHERE projeto_id = $1
ORDER BY ordem`

const qMinhaNota = `
SELECT nota FROM aric.projetos_avaliacoes WHERE projeto_id = $1 AND usuario_id = $2`

const qContarComentarios = `
SELECT count(*) FROM aric.projetos_comentarios WHERE projeto_id = $1 AND deletado_em IS NULL`

// Dados mínimos para checar permissão. FOR UPDATE quando usado dentro de transação
// que vai mudar o status (evita dois moderadores aprovando ao mesmo tempo).
const qBase = `
SELECT id, autor_id, status, xp, publicado_em
FROM aric.projetos WHERE id = $1`

const qBaseParaAtualizar = qBase + ` FOR UPDATE`

const qPapel = `
SELECT papel FROM aric.usuarios WHERE id = $1 AND deletado_em IS NULL`

const qLocalUsuario = `
SELECT u.bairro_id, b.cidade_id
FROM aric.usuarios u JOIN aric.bairros b ON b.id = u.bairro_id
WHERE u.id = $1`

// ---------- escrita do projeto ----------

const qContarCategorias = `
SELECT count(*) FROM aric.categorias WHERE id = ANY($1::int[])`

// O XP vem da dificuldade. Se a dificuldade não existir, não insere nada (sql.ErrNoRows).
const qInserir = `
INSERT INTO aric.projetos
    (titulo, subtitulo, descricao, dificuldade, xp, tempo_estimado, residuos_kg, autor_id, status)
SELECT $1, $2, $3, d.id, d.xp_padrao, $5, $6, $7, $8
FROM aric.dificuldades_projetos d
WHERE d.id = $4
RETURNING id`

const qAtualizar = `
UPDATE aric.projetos p
SET titulo = $2, subtitulo = $3, descricao = $4, dificuldade = d.id, xp = d.xp_padrao,
    tempo_estimado = $6, residuos_kg = $7, status = $8,
    motivo_moderacao = NULL, data_atualizacao = now()
FROM aric.dificuldades_projetos d
WHERE d.id = $5 AND p.id = $1`

const (
	qLimparCategorias = `DELETE FROM aric.projetos_categorias WHERE projeto_id = $1`
	qLimparMateriais  = `DELETE FROM aric.projetos_materiais WHERE projeto_id = $1`
	qLimparPassos     = `DELETE FROM aric.projetos_passos WHERE projeto_id = $1`
	qLimparImagens    = `DELETE FROM aric.projetos_midia WHERE projeto_id = $1`
)

const qInserirCategorias = `
INSERT INTO aric.projetos_categorias (projeto_id, categoria_id)
SELECT $1, c FROM unnest($2::int[]) AS c`

const qInserirMateriais = `
INSERT INTO aric.projetos_materiais (projeto_id, material, quantidade, imagem_url, ordem)
SELECT $1, t.nome, t.qtd, NULLIF(t.img, ''), t.ordem
FROM unnest($2::text[], $3::text[], $4::text[]) WITH ORDINALITY AS t(nome, qtd, img, ordem)`

const qInserirPassos = `
INSERT INTO aric.projetos_passos (projeto_id, passo_numero, titulo, descricao, imagem_url)
SELECT $1, t.ordem, t.titulo, t.descricao, NULLIF(t.img, '')
FROM unnest($2::text[], $3::text[], $4::text[]) WITH ORDINALITY AS t(titulo, descricao, img, ordem)`

const qInserirImagens = `
INSERT INTO aric.projetos_midia (projeto_id, midia_url, tipo_midia, ordem)
SELECT $1, t.url, 'imagem', t.ordem
FROM unnest($2::text[]) WITH ORDINALITY AS t(url, ordem)`

const qRemover = `
UPDATE aric.projetos SET status = 'removido', data_atualizacao = now() WHERE id = $1`

// ---------- moderação ----------

const qAprovar = `
UPDATE aric.projetos
SET status = 'publicado', motivo_moderacao = NULL, moderado_por = $2, moderado_em = now(),
    publicado_em = COALESCE(publicado_em, now())
WHERE id = $1 AND status = 'em_revisao'`

const qRejeitar = `
UPDATE aric.projetos
SET status = 'oculto', motivo_moderacao = $3, moderado_por = $2, moderado_em = now()
WHERE id = $1 AND status IN ('em_revisao', 'publicado')`

const qCategoriasDoProjeto = `
SELECT categoria_id FROM aric.projetos_categorias WHERE projeto_id = $1`

const qDefinirDestaque = `
INSERT INTO aric.projetos_destaques (semana_inicio, projeto_id)
VALUES (aric.inicio_semana(now()), $1)
ON CONFLICT (semana_inicio) DO UPDATE SET projeto_id = EXCLUDED.projeto_id, criado_em = now()`

const qDestaqueAtual = `
SELECT projeto_id FROM aric.projetos_destaques WHERE semana_inicio = aric.inicio_semana(now())`

// ---------- interações ----------

const qInserirConclusao = `
INSERT INTO aric.projetos_conclusoes (projeto_id, usuario_id, foto_url)
VALUES ($1, $2, NULLIF($3, ''))
ON CONFLICT (projeto_id, usuario_id) DO NOTHING
RETURNING data_conclusao`

// Não dá XP duas vezes pela mesma coisa: se já existe, não insere e não retorna linha.
const qInserirEventoXP = `
INSERT INTO aric.eventos_xp (usuario_id, quantidade, origem, referencia_id, descricao)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (usuario_id, origem, referencia_id) DO NOTHING
RETURNING quantidade`

const qXPUsuario = `
SELECT u.xp_total, u.nivel_id, n.titulo AS nivel_titulo
FROM aric.usuarios u LEFT JOIN aric.niveis n ON n.id = u.nivel_id
WHERE u.id = $1`

const (
	qCurtir    = `INSERT INTO aric.projetos_curtidas (projeto_id, usuario_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	qDescurtir = `DELETE FROM aric.projetos_curtidas WHERE projeto_id = $1 AND usuario_id = $2`
	qCurtidas  = `SELECT count(*) FROM aric.projetos_curtidas WHERE projeto_id = $1`

	qSalvar       = `INSERT INTO aric.projetos_salvos (projeto_id, usuario_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	qRemoverSalvo = `DELETE FROM aric.projetos_salvos WHERE projeto_id = $1 AND usuario_id = $2`
)

// A FK para projetos_conclusoes faz falhar (23503) se o usuário não concluiu o projeto.
const qAvaliar = `
INSERT INTO aric.projetos_avaliacoes (projeto_id, usuario_id, nota)
VALUES ($1, $2, $3)
ON CONFLICT (projeto_id, usuario_id) DO UPDATE SET nota = EXCLUDED.nota, atualizado_em = now()`

const qMediaNotas = `
SELECT round(avg(nota), 1)::float8 AS media, count(*) AS total
FROM aric.projetos_avaliacoes WHERE projeto_id = $1`

// ---------- comentários ----------

const qListarComentarios = `
SELECT c.id, c.conteudo, c.criado_em,
       u.id AS autor_id, u.nome AS autor_nome, u.foto_url AS autor_foto_url,
       COUNT(*) OVER () AS total
FROM aric.projetos_comentarios c
JOIN aric.usuarios u ON u.id = c.usuario_id
WHERE c.projeto_id = $1 AND c.deletado_em IS NULL
ORDER BY c.criado_em, c.id
LIMIT $2 OFFSET $3`

const qInserirComentario = `
WITH novo AS (
    INSERT INTO aric.projetos_comentarios (projeto_id, usuario_id, conteudo)
    VALUES ($1, $2, $3)
    RETURNING id, criado_em, usuario_id
)
SELECT n.id, n.criado_em, u.nome AS autor_nome, u.foto_url AS autor_foto_url
FROM novo n JOIN aric.usuarios u ON u.id = n.usuario_id`

const qComentario = `
SELECT id, projeto_id, usuario_id FROM aric.projetos_comentarios
WHERE id = $1 AND deletado_em IS NULL`

const qRemoverComentario = `
UPDATE aric.projetos_comentarios SET deletado_em = now() WHERE id = $1`

// ---------- catálogo ----------

const qCategorias = `
SELECT id, nome, tipo, cor, icone FROM aric.categorias ORDER BY tipo, nome`

const qDificuldades = `
SELECT id, descricao, xp_padrao FROM aric.dificuldades_projetos ORDER BY id`
