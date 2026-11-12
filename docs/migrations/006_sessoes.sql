-- =====================================================================
-- 006 - Invalidação de sessões
-- Pré-requisito: 003. Idempotente.
--
-- Tokens emitidos antes de sessoes_validas_desde são recusados.
-- Trocar a senha atualiza esta coluna, derrubando os outros aparelhos.
-- Conta excluída (deletado_em) também derruba todos os tokens.
-- =====================================================================

BEGIN;

ALTER TABLE aric.usuarios ADD COLUMN IF NOT EXISTS sessoes_validas_desde TIMESTAMPTZ;

COMMIT;
