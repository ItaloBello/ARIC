-- =====================================================================
-- 008 - Denúncias de conteúdo
-- Pré-requisito: 003. Idempotente.
--
-- Qualquer usuário denuncia um projeto, comentário de projeto, post ou resposta
-- do fórum. Moderadores veem a fila agrupada por conteúdo e decidem:
-- remover (procedente) ou manter (improcedente).
-- =====================================================================

BEGIN;

CREATE TABLE IF NOT EXISTS aric.denuncias (
    id             SERIAL PRIMARY KEY,
    denunciante_id INT NOT NULL REFERENCES aric.usuarios(id) ON DELETE CASCADE,
    alvo_tipo      VARCHAR(20) NOT NULL
                       CHECK (alvo_tipo IN ('projeto', 'projeto_comentario', 'forum_post', 'forum_resposta')),
    alvo_id        INT NOT NULL,
    motivo         VARCHAR(20) NOT NULL
                       CHECK (motivo IN ('spam', 'ofensivo', 'perigoso', 'falso', 'outro')),
    descricao      VARCHAR(500),
    status         VARCHAR(15) NOT NULL DEFAULT 'pendente'
                       CHECK (status IN ('pendente', 'procedente', 'improcedente')),
    resolvido_por  INT REFERENCES aric.usuarios(id),
    resolvido_em   TIMESTAMPTZ,
    criado_em      TIMESTAMPTZ NOT NULL DEFAULT now(),
    -- cada pessoa denuncia o mesmo conteúdo uma vez só
    UNIQUE (denunciante_id, alvo_tipo, alvo_id)
);

-- fila de moderação
CREATE INDEX IF NOT EXISTS idx_denuncias_pendentes ON aric.denuncias (alvo_tipo, alvo_id) WHERE status = 'pendente';
-- limite diário por denunciante
CREATE INDEX IF NOT EXISTS idx_denuncias_denunciante ON aric.denuncias (denunciante_id, criado_em DESC);

COMMIT;
