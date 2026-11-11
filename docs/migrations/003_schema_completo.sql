-- =====================================================================
-- 003 - Schema completo a partir das telas do design
--   Projetos (autor, curtidas, avaliações, comentários, salvos, destaque)
--   Fórum (perguntas/dicas, respostas, resposta aceita, curtidas, salvos)
--   Gamificação (log de XP, ligas semanais, missões)
--
-- Pré-requisitos: script.sql e os inserts (estados, cidades, níveis,
-- dificuldades, categorias). Banco antigo: rodar 002 antes.
-- O script é idempotente: pode rodar mais de uma vez sem erro.
--
-- Convenção de semana: começa na segunda-feira, horário de São Paulo.
-- Colunas "semana_inicio" guardam sempre a data dessa segunda.
-- =====================================================================

BEGIN;

-- ---------------------------------------------------------------------
-- Função utilitária: segunda-feira da semana de um instante (fuso SP)
-- Uso: aric.inicio_semana(now())
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION aric.inicio_semana(ts TIMESTAMPTZ)
RETURNS DATE
LANGUAGE sql STABLE
AS $$
    SELECT date_trunc('week', ts AT TIME ZONE 'America/Sao_Paulo')::date
$$;


-- =====================================================================
-- USUÁRIOS (complementos)
-- =====================================================================

-- papel: quem pode moderar (ocultar projeto, aprovar evidência de missão, escolher destaque)
ALTER TABLE aric.usuarios ADD COLUMN IF NOT EXISTS papel VARCHAR(20) NOT NULL DEFAULT 'usuario'
    CHECK (papel IN ('usuario', 'moderador', 'admin'));

-- dia da última ação que conta para a sequência (tela Perfil: "Sequência ativa 18 dias")
ALTER TABLE aric.usuarios ADD COLUMN IF NOT EXISTS ultima_atividade_em DATE;

CREATE INDEX IF NOT EXISTS idx_usuarios_bairro ON aric.usuarios (bairro_id);


-- =====================================================================
-- CATEGORIAS
-- tipo 'material' = do que é feito (Plástico, Vidro...)
-- tipo 'tema'     = assunto (Horta, Decoração, Upcycling...)
-- Os chips da Home misturam os dois, então ficam na mesma tabela.
-- =====================================================================

ALTER TABLE aric.categorias ADD COLUMN IF NOT EXISTS tipo VARCHAR(20) NOT NULL DEFAULT 'material'
    CHECK (tipo IN ('material', 'tema'));
ALTER TABLE aric.categorias ADD COLUMN IF NOT EXISTS cor VARCHAR(7);      -- ex.: #F2B705 (bolinha do chip)
ALTER TABLE aric.categorias ADD COLUMN IF NOT EXISTS icone VARCHAR(50);   -- nome do ícone no front

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'categorias_nome_key') THEN
        ALTER TABLE aric.categorias ADD CONSTRAINT categorias_nome_key UNIQUE (nome);
    END IF;
END $$;

INSERT INTO aric.categorias (id, nome, tipo) VALUES
    (16, 'Upcycling', 'tema'),
    (17, 'Decoração', 'tema'),
    (18, 'Compostagem', 'tema')
ON CONFLICT (id) DO NOTHING;

UPDATE aric.categorias SET tipo = 'tema' WHERE nome IN ('Energia Renovável', 'Horta', 'Alimentos', 'Água', 'Upcycling', 'Decoração', 'Compostagem');

UPDATE aric.categorias SET cor = v.cor
FROM (VALUES
    ('Plástico', '#F2B705'), ('Papelão', '#B07A4A'), ('Vidro', '#1A9E9E'),
    ('Metal', '#8A8F98'),   ('Orgânico', '#6B8E23'), ('Eletrônico', '#3B5BDB'),
    ('Têxtil', '#C2185B'),  ('Madeira', '#A0522D'),  ('Borracha', '#333333'),
    ('Papel', '#4C9AFF'),   ('Outros', '#9E9E9E')
) AS v(nome, cor)
WHERE aric.categorias.nome = v.nome AND aric.categorias.cor IS NULL;

SELECT setval(pg_get_serial_sequence('aric.categorias', 'id'), (SELECT MAX(id) FROM aric.categorias));


-- =====================================================================
-- DIFICULDADES
-- O XP do projeto vem da dificuldade, e não do autor: se o autor
-- escolhesse o XP, bastaria publicar um projeto valendo 10.000.
-- O backend copia xp_padrao para projetos.xp ao publicar.
-- =====================================================================

ALTER TABLE aric.dificuldades_projetos ADD COLUMN IF NOT EXISTS xp_padrao INT NOT NULL DEFAULT 50
    CHECK (xp_padrao >= 0);

UPDATE aric.dificuldades_projetos SET xp_padrao = 80  WHERE id = 1;  -- Iniciante (tela: +80 pts)
UPDATE aric.dificuldades_projetos SET xp_padrao = 150 WHERE id = 2;  -- Intermediário
UPDATE aric.dificuldades_projetos SET xp_padrao = 250 WHERE id = 3;  -- Avançado


-- =====================================================================
-- PROJETOS (complementos)
-- =====================================================================

-- autor NULL = projeto oficial do ARIC (cadastrado pela equipe)
ALTER TABLE aric.projetos ADD COLUMN IF NOT EXISTS autor_id INT REFERENCES aric.usuarios(id);

ALTER TABLE aric.projetos ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'publicado'
    CHECK (status IN ('rascunho', 'publicado', 'oculto', 'removido'));

-- estimativa de resíduo reaproveitado por execução (Home: "-1,8 kg"; base do "Impacto do bairro")
ALTER TABLE aric.projetos ADD COLUMN IF NOT EXISTS residuos_kg NUMERIC(8,2) NOT NULL DEFAULT 0
    CHECK (residuos_kg >= 0);

-- busca da Home ("Buscar projetos, materiais...")
ALTER TABLE aric.projetos ADD COLUMN IF NOT EXISTS busca TSVECTOR
    GENERATED ALWAYS AS (
        setweight(to_tsvector('portuguese', coalesce(titulo, '')), 'A') ||
        setweight(to_tsvector('portuguese', coalesce(subtitulo, '')), 'B') ||
        setweight(to_tsvector('portuguese', coalesce(descricao, '')), 'C')
    ) STORED;

CREATE INDEX IF NOT EXISTS idx_projetos_status_data ON aric.projetos (status, data_criacao DESC);
CREATE INDEX IF NOT EXISTS idx_projetos_autor       ON aric.projetos (autor_id);
CREATE INDEX IF NOT EXISTS idx_projetos_busca       ON aric.projetos USING GIN (busca);
CREATE INDEX IF NOT EXISTS idx_projetos_categorias_categoria ON aric.projetos_categorias (categoria_id);

-- galeria "1/6": ordem 1 é a capa
ALTER TABLE aric.projetos_midia ADD COLUMN IF NOT EXISTS ordem SMALLINT NOT NULL DEFAULT 1;

-- cards de "Materiais necessários" e de "Passo a passo" têm foto
ALTER TABLE aric.projetos_materiais ADD COLUMN IF NOT EXISTS imagem_url VARCHAR(255);
ALTER TABLE aric.projetos_materiais ADD COLUMN IF NOT EXISTS ordem SMALLINT NOT NULL DEFAULT 1;
ALTER TABLE aric.projetos_passos    ADD COLUMN IF NOT EXISTS imagem_url VARCHAR(255);

-- foto de quem fez o projeto (prova para validação)
ALTER TABLE aric.projetos_conclusoes ADD COLUMN IF NOT EXISTS foto_url VARCHAR(255);
CREATE INDEX IF NOT EXISTS idx_projetos_conclusoes_usuario ON aric.projetos_conclusoes (usuario_id, data_conclusao);
CREATE INDEX IF NOT EXISTS idx_projetos_conclusoes_data    ON aric.projetos_conclusoes (data_conclusao);

-- curtidas (coração da Home/Projeto/Perfil)
CREATE TABLE IF NOT EXISTS aric.projetos_curtidas (
    projeto_id INT NOT NULL REFERENCES aric.projetos(id) ON DELETE CASCADE,
    usuario_id INT NOT NULL REFERENCES aric.usuarios(id) ON DELETE CASCADE,
    criado_em  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (projeto_id, usuario_id)
);
CREATE INDEX IF NOT EXISTS idx_projetos_curtidas_usuario ON aric.projetos_curtidas (usuario_id);

-- salvos (marcador no topo da tela de Projeto)
CREATE TABLE IF NOT EXISTS aric.projetos_salvos (
    projeto_id INT NOT NULL REFERENCES aric.projetos(id) ON DELETE CASCADE,
    usuario_id INT NOT NULL REFERENCES aric.usuarios(id) ON DELETE CASCADE,
    criado_em  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (usuario_id, projeto_id)
);

-- avaliação 1 a 5 para filtrar projetos bons e ruins.
-- A FK para projetos_conclusoes garante que só avalia quem fez o projeto.
CREATE TABLE IF NOT EXISTS aric.projetos_avaliacoes (
    projeto_id    INT NOT NULL,
    usuario_id    INT NOT NULL,
    nota          SMALLINT NOT NULL CHECK (nota BETWEEN 1 AND 5),
    criado_em     TIMESTAMPTZ NOT NULL DEFAULT now(),
    atualizado_em TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (projeto_id, usuario_id),
    FOREIGN KEY (projeto_id, usuario_id)
        REFERENCES aric.projetos_conclusoes (projeto_id, usuario_id) ON DELETE CASCADE
);

-- comentários (tela de Projeto: "37 comentários")
CREATE TABLE IF NOT EXISTS aric.projetos_comentarios (
    id            SERIAL PRIMARY KEY,
    projeto_id    INT NOT NULL REFERENCES aric.projetos(id) ON DELETE CASCADE,
    usuario_id    INT NOT NULL REFERENCES aric.usuarios(id),
    conteudo      TEXT NOT NULL CHECK (length(trim(conteudo)) > 0),
    criado_em     TIMESTAMPTZ NOT NULL DEFAULT now(),
    atualizado_em TIMESTAMPTZ NOT NULL DEFAULT now(),
    deletado_em   TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_projetos_comentarios_projeto ON aric.projetos_comentarios (projeto_id, criado_em);

-- "Destaque da semana" da Home (um por semana, escolhido por moderador/admin)
CREATE TABLE IF NOT EXISTS aric.projetos_destaques (
    semana_inicio DATE PRIMARY KEY CHECK (EXTRACT(ISODOW FROM semana_inicio) = 1),
    projeto_id    INT NOT NULL REFERENCES aric.projetos(id),
    criado_em     TIMESTAMPTZ NOT NULL DEFAULT now()
);


-- =====================================================================
-- FÓRUM
-- tipo 'duvida' = "Perguntar" / filtro "Dúvidas abertas"
-- tipo 'dica'   = filtro "Dicas rápidas"
-- "Resolvido" = dúvida com resposta_aceita_id preenchida (não há coluna status)
-- =====================================================================

CREATE TABLE IF NOT EXISTS aric.forum_posts (
    id                 SERIAL PRIMARY KEY,
    autor_id           INT NOT NULL REFERENCES aric.usuarios(id),
    tipo               VARCHAR(10) NOT NULL DEFAULT 'duvida' CHECK (tipo IN ('duvida', 'dica')),
    titulo             VARCHAR(200) NOT NULL CHECK (length(trim(titulo)) > 0),
    conteudo           TEXT NOT NULL,
    imagem_url         VARCHAR(255),
    categoria_id       INT REFERENCES aric.categorias(id),
    resposta_aceita_id INT,
    criado_em          TIMESTAMPTZ NOT NULL DEFAULT now(),
    atualizado_em      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deletado_em        TIMESTAMPTZ,
    busca TSVECTOR GENERATED ALWAYS AS (
        setweight(to_tsvector('portuguese', coalesce(titulo, '')), 'A') ||
        setweight(to_tsvector('portuguese', coalesce(conteudo, '')), 'B')
    ) STORED,
    -- dica não tem "resposta aceita"
    CHECK (tipo = 'duvida' OR resposta_aceita_id IS NULL)
);
CREATE INDEX IF NOT EXISTS idx_forum_posts_data      ON aric.forum_posts (criado_em DESC) WHERE deletado_em IS NULL;
CREATE INDEX IF NOT EXISTS idx_forum_posts_autor     ON aric.forum_posts (autor_id);
CREATE INDEX IF NOT EXISTS idx_forum_posts_categoria ON aric.forum_posts (categoria_id);
CREATE INDEX IF NOT EXISTS idx_forum_posts_busca     ON aric.forum_posts USING GIN (busca);

CREATE TABLE IF NOT EXISTS aric.forum_respostas (
    id            SERIAL PRIMARY KEY,
    post_id       INT NOT NULL REFERENCES aric.forum_posts(id) ON DELETE CASCADE,
    autor_id      INT NOT NULL REFERENCES aric.usuarios(id),
    conteudo      TEXT NOT NULL CHECK (length(trim(conteudo)) > 0),
    criado_em     TIMESTAMPTZ NOT NULL DEFAULT now(),
    atualizado_em TIMESTAMPTZ NOT NULL DEFAULT now(),
    deletado_em   TIMESTAMPTZ,
    UNIQUE (id, post_id) -- alvo da FK composta abaixo
);
CREATE INDEX IF NOT EXISTS idx_forum_respostas_post  ON aric.forum_respostas (post_id, criado_em);
CREATE INDEX IF NOT EXISTS idx_forum_respostas_autor ON aric.forum_respostas (autor_id);

-- A resposta aceita precisa ser do próprio post (FK composta: (resposta, post) tem que existir)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_forum_posts_resposta_aceita') THEN
        ALTER TABLE aric.forum_posts
            ADD CONSTRAINT fk_forum_posts_resposta_aceita
            FOREIGN KEY (resposta_aceita_id, id)
            REFERENCES aric.forum_respostas (id, post_id);
        -- Respostas são apagadas com soft delete (deletado_em). Para apagar de verdade
        -- uma resposta aceita, limpe resposta_aceita_id antes.
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS aric.forum_posts_curtidas (
    post_id    INT NOT NULL REFERENCES aric.forum_posts(id) ON DELETE CASCADE,
    usuario_id INT NOT NULL REFERENCES aric.usuarios(id) ON DELETE CASCADE,
    criado_em  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (post_id, usuario_id)
);

CREATE TABLE IF NOT EXISTS aric.forum_respostas_curtidas (
    resposta_id INT NOT NULL REFERENCES aric.forum_respostas(id) ON DELETE CASCADE,
    usuario_id  INT NOT NULL REFERENCES aric.usuarios(id) ON DELETE CASCADE,
    criado_em   TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (resposta_id, usuario_id)
);

CREATE TABLE IF NOT EXISTS aric.forum_posts_salvos (
    post_id    INT NOT NULL REFERENCES aric.forum_posts(id) ON DELETE CASCADE,
    usuario_id INT NOT NULL REFERENCES aric.usuarios(id) ON DELETE CASCADE,
    criado_em  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (usuario_id, post_id)
);


-- =====================================================================
-- XP: log de tudo que gera (ou tira) XP
-- usuarios.xp_total é só um cache da soma: atualizar na MESMA transação
-- em que o evento é inserido. Rankings (semana, mês, bairro, cidade)
-- saem daqui filtrando por criado_em.
-- =====================================================================

CREATE TABLE IF NOT EXISTS aric.eventos_xp (
    id            BIGSERIAL PRIMARY KEY,
    usuario_id    INT NOT NULL REFERENCES aric.usuarios(id) ON DELETE CASCADE,
    quantidade    INT NOT NULL CHECK (quantidade <> 0),
    origem        VARCHAR(30) NOT NULL CHECK (origem IN (
                      'projeto_concluido',     -- referencia_id = projeto
                      'projeto_publicado',     -- referencia_id = projeto
                      'missao_concluida',      -- referencia_id = usuarios_missoes.id
                      'forum_resposta_aceita', -- referencia_id = forum_respostas.id
                      'ajuste'                 -- correção manual, referencia_id NULL
                  )),
    referencia_id INT,
    descricao     VARCHAR(255),
    criado_em     TIMESTAMPTZ NOT NULL DEFAULT now(),
    -- impede ganhar XP duas vezes pela mesma coisa (NULLs não conflitam, então 'ajuste' é livre)
    UNIQUE (usuario_id, origem, referencia_id)
);
CREATE INDEX IF NOT EXISTS idx_eventos_xp_data    ON aric.eventos_xp (criado_em, usuario_id);
CREATE INDEX IF NOT EXISTS idx_eventos_xp_usuario ON aric.eventos_xp (usuario_id, criado_em);

-- Ao inserir um evento: soma no xp_total, recalcula o nível e atualiza a sequência de dias.
-- O backend só precisa inserir em eventos_xp (na mesma transação da ação que gerou o XP).
-- Sequência: só eventos positivos que não sejam 'ajuste' contam como atividade do dia.
--   mesmo dia -> mantém | dia seguinte -> +1 | pulou um dia ou mais -> volta para 1
-- Ao EXIBIR a sequência, trate como 0 se ultima_atividade_em < hoje - 1 (o usuário quebrou a
-- sequência mas ainda não fez nada para o trigger zerar).
CREATE OR REPLACE FUNCTION aric.aplicar_evento_xp()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    v_hoje DATE := (NEW.criado_em AT TIME ZONE 'America/Sao_Paulo')::date;
    v_conta_atividade BOOLEAN := NEW.quantidade > 0 AND NEW.origem <> 'ajuste';
BEGIN
    UPDATE aric.usuarios u
    SET xp_total = GREATEST(u.xp_total + NEW.quantidade, 0),
        nivel_id = (SELECT n.id FROM aric.niveis n
                    WHERE n.minimo_xp <= GREATEST(u.xp_total + NEW.quantidade, 0)
                    ORDER BY n.minimo_xp DESC LIMIT 1),
        sequencia_dias = CASE
            WHEN NOT v_conta_atividade                    THEN u.sequencia_dias
            WHEN u.ultima_atividade_em >= v_hoje          THEN u.sequencia_dias
            WHEN u.ultima_atividade_em = v_hoje - 1       THEN u.sequencia_dias + 1
            ELSE 1
        END,
        ultima_atividade_em = CASE
            WHEN v_conta_atividade THEN GREATEST(u.ultima_atividade_em, v_hoje)
            ELSE u.ultima_atividade_em
        END,
        atualizado_em = now()
    WHERE u.id = NEW.usuario_id;
    RETURN NEW;
END $$;

DROP TRIGGER IF EXISTS trg_eventos_xp_aplicar ON aric.eventos_xp;
CREATE TRIGGER trg_eventos_xp_aplicar
    AFTER INSERT ON aric.eventos_xp
    FOR EACH ROW EXECUTE FUNCTION aric.aplicar_evento_xp();

-- O log é só de inserção: correção se faz com um evento 'ajuste' (positivo ou negativo).
-- UPDATE é proibido. DELETE é permitido (ex.: apagar usuário em cascata) e desconta do cache.
CREATE OR REPLACE FUNCTION aric.proteger_evento_xp()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF TG_OP = 'UPDATE' THEN
        RAISE EXCEPTION 'eventos_xp não pode ser alterado; insira um evento de origem ''ajuste''';
    END IF;

    UPDATE aric.usuarios u
    SET xp_total = GREATEST(u.xp_total - OLD.quantidade, 0),
        nivel_id = (SELECT n.id FROM aric.niveis n
                    WHERE n.minimo_xp <= GREATEST(u.xp_total - OLD.quantidade, 0)
                    ORDER BY n.minimo_xp DESC LIMIT 1)
    WHERE u.id = OLD.usuario_id;
    RETURN OLD;
END $$;

DROP TRIGGER IF EXISTS trg_eventos_xp_proteger ON aric.eventos_xp;
CREATE TRIGGER trg_eventos_xp_proteger
    AFTER UPDATE OR DELETE ON aric.eventos_xp
    FOR EACH ROW EXECUTE FUNCTION aric.proteger_evento_xp();


-- =====================================================================
-- LIGAS (ranking semanal)
-- Toda segunda o job fecha a semana anterior: soma o XP da semana por
-- usuário dentro da liga, grava o resultado em ligas_semanas e move
-- usuarios.liga_id (sobe os qtd_sobe primeiros, desce os qtd_desce últimos).
-- =====================================================================

CREATE TABLE IF NOT EXISTS aric.ligas (
    id        SERIAL PRIMARY KEY,
    ordem     INT NOT NULL UNIQUE,        -- 1 = mais baixa
    nome      VARCHAR(100) NOT NULL UNIQUE,
    icone     VARCHAR(50),
    qtd_sobe  INT NOT NULL DEFAULT 0 CHECK (qtd_sobe >= 0),
    qtd_desce INT NOT NULL DEFAULT 0 CHECK (qtd_desce >= 0)
);

INSERT INTO aric.ligas (ordem, nome, qtd_sobe, qtd_desce) VALUES
    (1, 'Liga Semente',  10, 0),   -- liga mais baixa: ninguém desce
    (2, 'Liga Broto',    10, 5),
    (3, 'Liga Muda',      7, 5),
    (4, 'Liga Árvore',    5, 5),
    (5, 'Liga Floresta',  0, 5)    -- liga mais alta: ninguém sobe
ON CONFLICT (ordem) DO NOTHING;

-- liga atual do usuário (novo usuário entra na de ordem 1)
ALTER TABLE aric.usuarios ADD COLUMN IF NOT EXISTS liga_id INT REFERENCES aric.ligas(id);
UPDATE aric.usuarios SET liga_id = (SELECT id FROM aric.ligas ORDER BY ordem LIMIT 1) WHERE liga_id IS NULL;

-- Usuário novo cai na liga mais baixa sem o backend precisar saber disso
CREATE OR REPLACE FUNCTION aric.definir_liga_inicial()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF NEW.liga_id IS NULL THEN
        SELECT id INTO NEW.liga_id FROM aric.ligas ORDER BY ordem LIMIT 1;
    END IF;
    RETURN NEW;
END $$;

DROP TRIGGER IF EXISTS trg_usuarios_liga_inicial ON aric.usuarios;
CREATE TRIGGER trg_usuarios_liga_inicial
    BEFORE INSERT ON aric.usuarios
    FOR EACH ROW EXECUTE FUNCTION aric.definir_liga_inicial();
CREATE INDEX IF NOT EXISTS idx_usuarios_liga ON aric.usuarios (liga_id);

-- histórico fechado de cada semana (alimenta "Tendência" e o histórico do perfil)
CREATE TABLE IF NOT EXISTS aric.ligas_semanas (
    semana_inicio   DATE NOT NULL CHECK (EXTRACT(ISODOW FROM semana_inicio) = 1),
    usuario_id      INT  NOT NULL REFERENCES aric.usuarios(id) ON DELETE CASCADE,
    liga_id         INT  NOT NULL REFERENCES aric.ligas(id),
    xp_semana       INT  NOT NULL DEFAULT 0,
    posicao         INT  NOT NULL CHECK (posicao > 0),
    resultado       VARCHAR(10) NOT NULL CHECK (resultado IN ('subiu', 'desceu', 'manteve')),
    liga_seguinte_id INT NOT NULL REFERENCES aric.ligas(id),
    PRIMARY KEY (semana_inicio, usuario_id)
);
CREATE INDEX IF NOT EXISTS idx_ligas_semanas_liga    ON aric.ligas_semanas (semana_inicio, liga_id, posicao);
CREATE INDEX IF NOT EXISTS idx_ligas_semanas_usuario ON aric.ligas_semanas (usuario_id, semana_inicio DESC);


-- =====================================================================
-- MISSÕES (semanais)
-- Cada missão ativa vale para todas as semanas; o progresso é por
-- usuário + semana. "acao" diz o que faz o progresso andar:
--   concluir_projeto / publicar_projeto / responder_forum / perguntar_forum
--     -> o backend incrementa sozinho (filtrando por categoria_id, se houver)
--   registro_manual
--     -> o usuário envia foto/quantidade em missoes_evidencias e um
--        moderador aprova (ex.: "entregue 10 garrafas no ponto de coleta")
-- =====================================================================

CREATE TABLE IF NOT EXISTS aric.missoes (
    id           SERIAL PRIMARY KEY,
    titulo       VARCHAR(150) NOT NULL,
    descricao    TEXT NOT NULL,
    imagem_url   VARCHAR(255),
    categoria_id INT REFERENCES aric.categorias(id),
    acao         VARCHAR(30) NOT NULL CHECK (acao IN (
                     'concluir_projeto', 'publicar_projeto',
                     'responder_forum', 'perguntar_forum', 'registro_manual')),
    meta         INT NOT NULL CHECK (meta > 0),          -- "10" garrafas, "1" atividade, "3" etapas
    unidade      VARCHAR(30) NOT NULL,                   -- "garrafas", "atividade", "etapas"
    xp           INT NOT NULL CHECK (xp > 0),
    ativa        BOOLEAN NOT NULL DEFAULT TRUE,
    criado_em    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS aric.usuarios_missoes (
    id            SERIAL PRIMARY KEY,
    usuario_id    INT  NOT NULL REFERENCES aric.usuarios(id) ON DELETE CASCADE,
    missao_id     INT  NOT NULL REFERENCES aric.missoes(id),
    semana_inicio DATE NOT NULL CHECK (EXTRACT(ISODOW FROM semana_inicio) = 1),
    progresso     INT  NOT NULL DEFAULT 0 CHECK (progresso >= 0),
    concluida_em  TIMESTAMPTZ,
    atualizado_em TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (usuario_id, missao_id, semana_inicio)
);
CREATE INDEX IF NOT EXISTS idx_usuarios_missoes_semana ON aric.usuarios_missoes (usuario_id, semana_inicio);

CREATE TABLE IF NOT EXISTS aric.missoes_evidencias (
    id                SERIAL PRIMARY KEY,
    usuario_missao_id INT NOT NULL REFERENCES aric.usuarios_missoes(id) ON DELETE CASCADE,
    foto_url          VARCHAR(255) NOT NULL,
    quantidade        INT NOT NULL DEFAULT 1 CHECK (quantidade > 0),
    status            VARCHAR(10) NOT NULL DEFAULT 'pendente'
                          CHECK (status IN ('pendente', 'aprovada', 'rejeitada')),
    avaliado_por      INT REFERENCES aric.usuarios(id),
    avaliado_em       TIMESTAMPTZ,
    criado_em         TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_missoes_evidencias_pendentes ON aric.missoes_evidencias (criado_em) WHERE status = 'pendente';

-- Missões de exemplo (as três da tela)
INSERT INTO aric.missoes (titulo, descricao, categoria_id, acao, meta, unidade, xp)
SELECT v.titulo, v.descricao, c.id, v.acao, v.meta, v.unidade, v.xp
FROM (VALUES
    ('Separe e entregue garrafas PET',
     'Junte pelo menos 10 garrafas PET e entregue no ponto de coleta da sua comunidade.',
     'Plástico', 'registro_manual', 10, 'garrafas', 150),
    ('Crie algo novo com papelão',
     'Transforme caixas de papelão em um objeto útil ou decorativo e compartilhe uma foto.',
     'Papelão', 'concluir_projeto', 1, 'atividade', 80),
    ('Comece a compostar',
     'Monte sua composteira caseira e registre o processo com uma foto.',
     'Horta', 'registro_manual', 3, 'etapas', 60)
) AS v(titulo, descricao, categoria, acao, meta, unidade, xp)
JOIN aric.categorias c ON c.nome = v.categoria
WHERE NOT EXISTS (SELECT 1 FROM aric.missoes m WHERE m.titulo = v.titulo);


-- =====================================================================
-- FUNÇÕES DE RANKING
-- =====================================================================

-- Ranking da aba Comunidade: a liga do usuário na semana atual.
-- Uso:
--   SELECT * FROM aric.ranking_liga((SELECT liga_id FROM aric.usuarios WHERE id = $1));
--   "Sua posição": a mesma chamada com WHERE usuario_id = $1
-- Colunas:
--   posicao    mesma ordem usada no fechamento (XP da semana; empate: cadastro mais antigo)
--   tendencia  quantas posições subiu (+) ou caiu (-) desde o fim de ontem
--   zona       'sobe' / 'desce' / 'mantem' se a semana fechasse agora (mesmas regras do fechamento)
CREATE OR REPLACE FUNCTION aric.ranking_liga(p_liga_id INT)
RETURNS TABLE (posicao BIGINT, usuario_id INT, nome VARCHAR, foto_url VARCHAR,
               bairro VARCHAR, xp_semana BIGINT, tendencia BIGINT, zona VARCHAR)
LANGUAGE sql STABLE
AS $$
    WITH limites AS (
        SELECT aric.inicio_semana(now())::timestamp AT TIME ZONE 'America/Sao_Paulo' AS inicio,
               (now() AT TIME ZONE 'America/Sao_Paulo')::date::timestamp
                   AT TIME ZONE 'America/Sao_Paulo' AS inicio_hoje
    ),
    xp AS (
        SELECT e.usuario_id,
               SUM(e.quantidade) AS xp_agora,
               SUM(e.quantidade) FILTER (WHERE e.criado_em < l.inicio_hoje) AS xp_ontem
        FROM aric.eventos_xp e, limites l
        WHERE e.criado_em >= l.inicio
        GROUP BY e.usuario_id
    ),
    membros AS (
        SELECT u.id, u.nome, u.foto_url, b.nome AS bairro, u.criado_em,
               COALESCE(x.xp_agora, 0) AS xp_agora,
               COALESCE(x.xp_ontem, 0) AS xp_ontem
        FROM aric.usuarios u
        JOIN aric.bairros b ON b.id = u.bairro_id
        LEFT JOIN xp x ON x.usuario_id = u.id
        WHERE u.liga_id = p_liga_id AND u.deletado_em IS NULL
    ),
    ranqueado AS (
        SELECT m.*,
               ROW_NUMBER() OVER (ORDER BY m.xp_agora DESC, m.criado_em, m.id) AS pos_agora,
               ROW_NUMBER() OVER (ORDER BY m.xp_ontem DESC, m.criado_em, m.id) AS pos_ontem,
               COUNT(*) OVER () AS total
        FROM membros m
    ),
    liga AS (
        SELECT l.qtd_sobe, l.qtd_desce,
               EXISTS (SELECT 1 FROM aric.ligas a WHERE a.ordem > l.ordem) AS tem_acima,
               EXISTS (SELECT 1 FROM aric.ligas a WHERE a.ordem < l.ordem) AS tem_abaixo
        FROM aric.ligas l WHERE l.id = p_liga_id
    )
    SELECT r.pos_agora, r.id, r.nome, r.foto_url, r.bairro, r.xp_agora,
           r.pos_ontem - r.pos_agora,
           (CASE
               WHEN lg.tem_acima AND r.xp_agora > 0 AND r.pos_agora <= lg.qtd_sobe THEN 'sobe'
               WHEN lg.tem_abaixo AND r.pos_agora > lg.qtd_sobe
                    AND r.pos_agora > r.total - lg.qtd_desce THEN 'desce'
               ELSE 'mantem'
           END)::VARCHAR
    FROM ranqueado r, liga lg
    ORDER BY r.pos_agora
$$;

-- A primeira versão deste script tinha ranking por bairro/cidade/mês; o ranking ficou só nas ligas.
DROP FUNCTION IF EXISTS aric.ranking_xp(TIMESTAMPTZ, TIMESTAMPTZ, INT, INT);

-- Fecha a semana das ligas. Chamar toda segunda logo após 00:00 (horário de SP)
-- passando a segunda da semana que TERMINOU:
--   SELECT aric.fechar_semana_ligas(aric.inicio_semana(now()) - 7);
-- Regras:
--   * posição na liga por XP da semana (empate: quem se cadastrou antes fica na frente)
--   * sobe: os qtd_sobe primeiros, desde que tenham feito XP na semana
--   * desce: os qtd_desce últimos, exceto quem está na zona de subida (liga pequena)
--   * liga mais alta não sobe, liga mais baixa não desce
-- Grava ligas_semanas e atualiza usuarios.liga_id. Retorna quantos usuários processou.
-- Recusa fechar a mesma semana duas vezes ou uma semana que ainda não acabou.
CREATE OR REPLACE FUNCTION aric.fechar_semana_ligas(p_semana DATE)
RETURNS INT
LANGUAGE plpgsql
AS $$
DECLARE
    v_inicio TIMESTAMPTZ := p_semana::timestamp AT TIME ZONE 'America/Sao_Paulo';
    v_fim    TIMESTAMPTZ := (p_semana + 7)::timestamp AT TIME ZONE 'America/Sao_Paulo';
    v_total  INT;
BEGIN
    IF EXTRACT(ISODOW FROM p_semana) <> 1 THEN
        RAISE EXCEPTION 'p_semana (%) precisa ser uma segunda-feira', p_semana;
    END IF;
    IF v_fim > now() THEN
        RAISE EXCEPTION 'a semana % ainda não terminou', p_semana;
    END IF;
    IF EXISTS (SELECT 1 FROM aric.ligas_semanas WHERE semana_inicio = p_semana) THEN
        RAISE EXCEPTION 'a semana % já foi fechada', p_semana;
    END IF;

    -- evita dois jobs fechando a mesma semana ao mesmo tempo
    PERFORM pg_advisory_xact_lock(hashtext('aric.fechar_semana_ligas'));

    INSERT INTO aric.ligas_semanas
        (semana_inicio, usuario_id, liga_id, xp_semana, posicao, resultado, liga_seguinte_id)
    WITH xp AS (
        SELECT e.usuario_id, SUM(e.quantidade) AS xp
        FROM aric.eventos_xp e
        WHERE e.criado_em >= v_inicio AND e.criado_em < v_fim
        GROUP BY e.usuario_id
    ),
    ligas_ord AS (
        SELECT l.id, l.qtd_sobe, l.qtd_desce,
               LEAD(l.id) OVER (ORDER BY l.ordem) AS acima,
               LAG(l.id)  OVER (ORDER BY l.ordem) AS abaixo
        FROM aric.ligas l
    ),
    ranqueado AS (
        SELECT u.id AS usuario_id, u.liga_id, COALESCE(x.xp, 0) AS xp_semana,
               lo.qtd_sobe, lo.qtd_desce, lo.acima, lo.abaixo,
               ROW_NUMBER() OVER (PARTITION BY u.liga_id
                                  ORDER BY COALESCE(x.xp, 0) DESC, u.criado_em, u.id) AS posicao,
               COUNT(*) OVER (PARTITION BY u.liga_id) AS total
        FROM aric.usuarios u
        JOIN ligas_ord lo ON lo.id = u.liga_id
        LEFT JOIN xp x ON x.usuario_id = u.id
        WHERE u.deletado_em IS NULL
          AND u.criado_em < v_fim
    ),
    resultado AS (
        SELECT r.*,
               CASE
                   WHEN r.acima IS NOT NULL AND r.xp_semana > 0 AND r.posicao <= r.qtd_sobe
                       THEN 'subiu'
                   WHEN r.abaixo IS NOT NULL AND r.posicao > r.qtd_sobe
                        AND r.posicao > r.total - r.qtd_desce
                       THEN 'desceu'
                   ELSE 'manteve'
               END AS resultado
        FROM ranqueado r
    )
    SELECT p_semana, usuario_id, liga_id, xp_semana, posicao, resultado,
           CASE resultado WHEN 'subiu' THEN acima WHEN 'desceu' THEN abaixo ELSE liga_id END
    FROM resultado;

    GET DIAGNOSTICS v_total = ROW_COUNT;

    UPDATE aric.usuarios u
    SET liga_id = ls.liga_seguinte_id
    FROM aric.ligas_semanas ls
    WHERE ls.semana_inicio = p_semana
      AND ls.usuario_id = u.id
      AND u.liga_id IS DISTINCT FROM ls.liga_seguinte_id;

    RETURN v_total;
END $$;

COMMIT;
