-- O banco é criado uma vez, fora deste script (Postgres não tem CREATE DATABASE IF NOT EXISTS):
--   CREATE DATABASE aric;
-- Depois, conectado no banco aric:

CREATE SCHEMA IF NOT EXISTS aric;

--Tabelas básicas para armazenar informações de usuários, níveis e localização geográfica (estado, cidade, bairro).
CREATE TABLE IF NOT EXISTS aric.estados (
    id SERIAL PRIMARY KEY,
    nome VARCHAR(255) NOT NULL,
    sigla VARCHAR(3) NOT NULL
);

CREATE TABLE IF NOT EXISTS aric.cidades (
    id SERIAL PRIMARY KEY,
    nome VARCHAR(255) NOT NULL,
    estado_id INT NOT NULL REFERENCES aric.estados(id)
);

CREATE TABLE IF NOT EXISTS aric.bairros (
    id SERIAL PRIMARY KEY,
    nome VARCHAR(255) NOT NULL,
    nome_normalizado VARCHAR(255) NOT NULL,
    cidade_id INT NOT NULL REFERENCES aric.cidades(id),
    CONSTRAINT unique_bairro_cidade UNIQUE (nome_normalizado, cidade_id)
);

CREATE TABLE IF NOT EXISTS aric.niveis (
    id SERIAL PRIMARY KEY,
    numero INT NOT NULL UNIQUE,
    titulo VARCHAR(255) NOT NULL,
    minimo_xp INT NOT NULL
);

CREATE TABLE IF NOT EXISTS aric.usuarios (
    id SERIAL PRIMARY KEY,
    nome VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE, -- gravado sempre em minúsculas
    senha_hash VARCHAR(255) NOT NULL,
    foto_url VARCHAR(255),
    bio TEXT,
    cep VARCHAR(8) NOT NULL, -- só dígitos
    bairro_id INT NOT NULL REFERENCES aric.bairros(id),
    logradouro VARCHAR(255) NOT NULL,
    numero_endereco VARCHAR(20) NOT NULL,
    complemento_endereco VARCHAR(255),
    xp_total INT NOT NULL DEFAULT 0,
    nivel_id INT REFERENCES aric.niveis(id),
    sequencia_dias INT NOT NULL DEFAULT 0,
    criado_em TIMESTAMPTZ NOT NULL DEFAULT now(),
    atualizado_em TIMESTAMPTZ NOT NULL DEFAULT now(),
    deletado_em TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS aric.dificuldades_projetos (
    id INT PRIMARY KEY,
    descricao VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS aric.projetos (
    id SERIAL PRIMARY KEY,
    titulo VARCHAR(255) NOT NULL,
    subtitulo VARCHAR(255) NOT NULL,
    descricao TEXT NOT NULL,
    dificuldade INTEGER NOT NULL REFERENCES aric.dificuldades_projetos(id),
    xp INTEGER NOT NULL,
    tempo_estimado INTEGER NOT NULL, -- tempo estimado em minutos
    data_criacao TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    data_atualizacao TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS aric.categorias (
    id SERIAL PRIMARY KEY,
    nome VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS aric.projetos_categorias (
    projeto_id INT NOT NULL REFERENCES aric.projetos(id),
    categoria_id INT NOT NULL REFERENCES aric.categorias(id),
    PRIMARY KEY (projeto_id, categoria_id)
);

CREATE TABLE IF NOT EXISTS aric.projetos_materiais (
    projeto_id INT NOT NULL REFERENCES aric.projetos(id),
    material VARCHAR(255) NOT NULL,
    quantidade VARCHAR(255) NOT NULL,
    PRIMARY KEY (projeto_id, material)
);

CREATE TABLE IF NOT EXISTS aric.projetos_passos (
    projeto_id INT NOT NULL REFERENCES aric.projetos(id),
    passo_numero INT NOT NULL,
    descricao TEXT NOT NULL,
    titulo VARCHAR(255) NOT NULL,
    PRIMARY KEY (projeto_id, passo_numero)
);

CREATE TABLE IF NOT EXISTS aric.projetos_midia (
    projeto_id INT NOT NULL REFERENCES aric.projetos(id),
    midia_url VARCHAR(255) NOT NULL,
    tipo_midia VARCHAR(50) NOT NULL, -- 'imagem' ou 'video'
    PRIMARY KEY (projeto_id, midia_url)
);

CREATE TABLE IF NOT EXISTS aric.projetos_conclusoes (
    projeto_id INT NOT NULL REFERENCES aric.projetos(id),
    usuario_id INT NOT NULL REFERENCES aric.usuarios(id),
    data_conclusao TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    validado BOOLEAN DEFAULT FALSE,
    PRIMARY KEY (projeto_id, usuario_id)
);

