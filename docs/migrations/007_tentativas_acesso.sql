-- =====================================================================
-- 007 - Limite de tentativas de login e cadastro
-- Pré-requisito: 003. Idempotente.
--
-- Cada tentativa de login, de confirmação de senha (trocar senha, excluir conta)
-- e cada cadastro fica registrada aqui. O backend conta as falhas recentes
-- para bloquear força bruta. Registros com mais de 30 dias são apagados pela API.
-- =====================================================================

BEGIN;

CREATE TABLE IF NOT EXISTS aric.tentativas_acesso (
    id        BIGSERIAL PRIMARY KEY,
    tipo      VARCHAR(20) NOT NULL CHECK (tipo IN ('login', 'senha', 'cadastro')),
    email     VARCHAR(255),   -- sempre em minúsculas
    ip        INET,
    sucesso   BOOLEAN NOT NULL,
    criado_em TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_tentativas_email ON aric.tentativas_acesso (email, criado_em DESC);
CREATE INDEX IF NOT EXISTS idx_tentativas_ip    ON aric.tentativas_acesso (ip, criado_em DESC);
CREATE INDEX IF NOT EXISTS idx_tentativas_data  ON aric.tentativas_acesso (criado_em);

COMMIT;
