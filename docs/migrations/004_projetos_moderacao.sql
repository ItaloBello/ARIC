-- =====================================================================
-- 004 - Projetos: revisão antes de publicar e ajustes para o backend
-- Pré-requisito: 003. Idempotente.
--
-- Fluxo de status:
--   rascunho    -> o autor ainda está montando (só ele vê)
--   em_revisao  -> enviado; aguarda moderador (só autor e moderadores veem)
--   publicado   -> aprovado; aparece para todos
--   oculto      -> rejeitado ou tirado do ar por moderador (motivo em motivo_moderacao)
--   removido    -> apagado pelo autor ou moderador (soft delete)
-- Editar um projeto publicado faz ele voltar para em_revisao.
-- =====================================================================

BEGIN;

-- 'em_revisao' passa a existir e vira o padrão
ALTER TABLE aric.projetos DROP CONSTRAINT IF EXISTS projetos_status_check;
ALTER TABLE aric.projetos ADD CONSTRAINT projetos_status_check
    CHECK (status IN ('rascunho', 'em_revisao', 'publicado', 'oculto', 'removido'));
ALTER TABLE aric.projetos ALTER COLUMN status SET DEFAULT 'em_revisao';

-- quem moderou, quando e por quê
ALTER TABLE aric.projetos ADD COLUMN IF NOT EXISTS motivo_moderacao VARCHAR(500);
ALTER TABLE aric.projetos ADD COLUMN IF NOT EXISTS moderado_por INT REFERENCES aric.usuarios(id);
ALTER TABLE aric.projetos ADD COLUMN IF NOT EXISTS moderado_em TIMESTAMPTZ;
-- primeira aprovação (o XP de publicação é dado uma vez só)
ALTER TABLE aric.projetos ADD COLUMN IF NOT EXISTS publicado_em TIMESTAMPTZ;

-- fila de moderação (mais antigos primeiro)
CREATE INDEX IF NOT EXISTS idx_projetos_em_revisao ON aric.projetos (data_atualizacao) WHERE status = 'em_revisao';

-- Datas com fuso. Os valores antigos foram gravados com CURRENT_TIMESTAMP no fuso da sessão,
-- então a conversão sem USING (que também usa o fuso da sessão) preserva o instante certo.
DO $$
BEGIN
    IF (SELECT data_type FROM information_schema.columns
        WHERE table_schema = 'aric' AND table_name = 'projetos' AND column_name = 'data_criacao')
       = 'timestamp without time zone' THEN
        ALTER TABLE aric.projetos ALTER COLUMN data_criacao TYPE TIMESTAMPTZ;
        ALTER TABLE aric.projetos ALTER COLUMN data_atualizacao TYPE TIMESTAMPTZ;
        ALTER TABLE aric.projetos_conclusoes ALTER COLUMN data_conclusao TYPE TIMESTAMPTZ;
    END IF;
END $$;

UPDATE aric.projetos SET data_criacao = now() WHERE data_criacao IS NULL;
UPDATE aric.projetos SET data_atualizacao = data_criacao WHERE data_atualizacao IS NULL;
ALTER TABLE aric.projetos ALTER COLUMN data_criacao SET DEFAULT now();
ALTER TABLE aric.projetos ALTER COLUMN data_criacao SET NOT NULL;
ALTER TABLE aric.projetos ALTER COLUMN data_atualizacao SET DEFAULT now();
ALTER TABLE aric.projetos ALTER COLUMN data_atualizacao SET NOT NULL;

UPDATE aric.projetos_conclusoes SET data_conclusao = now() WHERE data_conclusao IS NULL;
ALTER TABLE aric.projetos_conclusoes ALTER COLUMN data_conclusao SET DEFAULT now();
ALTER TABLE aric.projetos_conclusoes ALTER COLUMN data_conclusao SET NOT NULL;
UPDATE aric.projetos_conclusoes SET validado = FALSE WHERE validado IS NULL;
ALTER TABLE aric.projetos_conclusoes ALTER COLUMN validado SET NOT NULL;

-- projetos que já estavam publicados antes deste script contam como aprovados
UPDATE aric.projetos SET publicado_em = data_criacao WHERE status = 'publicado' AND publicado_em IS NULL;

-- regras de valor
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'projetos_tempo_estimado_check') THEN
        ALTER TABLE aric.projetos ADD CONSTRAINT projetos_tempo_estimado_check CHECK (tempo_estimado > 0);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'projetos_xp_check') THEN
        ALTER TABLE aric.projetos ADD CONSTRAINT projetos_xp_check CHECK (xp >= 0);
    END IF;
END $$;

-- Os filhos de projeto saem junto quando um projeto é apagado de verdade
-- (o app usa soft delete, mas isso evita lixo em limpezas manuais).
DO $$
DECLARE
    t TEXT;
BEGIN
    FOREACH t IN ARRAY ARRAY['projetos_categorias', 'projetos_materiais', 'projetos_passos',
                             'projetos_midia', 'projetos_conclusoes']
    LOOP
        EXECUTE format('ALTER TABLE aric.%I DROP CONSTRAINT IF EXISTS %I', t, t || '_projeto_id_fkey');
        EXECUTE format('ALTER TABLE aric.%I ADD CONSTRAINT %I FOREIGN KEY (projeto_id)
                        REFERENCES aric.projetos(id) ON DELETE CASCADE', t, t || '_projeto_id_fkey');
    END LOOP;
END $$;

COMMIT;
