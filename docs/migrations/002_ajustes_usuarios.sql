-- Rode este arquivo SOMENTE se o banco já foi criado com a versão antiga do script.sql.
-- Em banco novo, o script.sql já contém tudo isto.

-- Colunas que o backend usa e não existiam
ALTER TABLE aric.usuarios ADD COLUMN IF NOT EXISTS criado_em TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE aric.usuarios ADD COLUMN IF NOT EXISTS atualizado_em TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE aric.usuarios ADD COLUMN IF NOT EXISTS deletado_em TIMESTAMPTZ;

-- Evita NULL em contadores (o Go lê esses campos como int)
UPDATE aric.usuarios SET xp_total = 0 WHERE xp_total IS NULL;
UPDATE aric.usuarios SET sequencia_dias = 0 WHERE sequencia_dias IS NULL;
ALTER TABLE aric.usuarios ALTER COLUMN xp_total SET NOT NULL;
ALTER TABLE aric.usuarios ALTER COLUMN sequencia_dias SET NOT NULL;

-- E-mail sempre em minúsculas
UPDATE aric.usuarios SET email = lower(trim(email));

-- Número do nível único
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'niveis_numero_key') THEN
        ALTER TABLE aric.niveis ADD CONSTRAINT niveis_numero_key UNIQUE (numero);
    END IF;
END $$;
