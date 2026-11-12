-- =====================================================================
-- 005 - Evidências de missão: motivo da rejeição
-- Pré-requisito: 003. Idempotente.
-- =====================================================================

BEGIN;

-- o usuário vê por que a foto não foi aceita
ALTER TABLE aric.missoes_evidencias ADD COLUMN IF NOT EXISTS motivo_rejeicao VARCHAR(300);

-- "minhas evidências" (mais recentes primeiro)
CREATE INDEX IF NOT EXISTS idx_missoes_evidencias_usuario_missao
    ON aric.missoes_evidencias (usuario_missao_id, criado_em DESC);

COMMIT;
