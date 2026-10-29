CREATE DATABASE IF NOT EXISTS aric;

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
    cidade_id INT NOT NULL REFERENCES aric.cidades(id)
);

ALTER TABLE aric.bairros ADD CONSTRAINT unique_bairro_cidade UNIQUE (nome_normalizado, cidade_id);

CREATE TABLE IF NOT EXISTS aric.niveis (
    id SERIAL PRIMARY KEY,
    numero INT NOT NULL,
    titulo VARCHAR(255) NOT NULL,
    minimo_xp INT NOT NULL
);

CREATE TABLE IF NOT EXISTS aric.usuarios (
    id SERIAL PRIMARY KEY,
    nome VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    senha_hash VARCHAR(255) NOT NULL,
    foto_url VARCHAR(255),
    bio TEXT,
    cep VARCHAR(20) NOT NULL,
    bairro_id INT REFERENCES aric.bairros (id) NOT NULL,
    logradouro VARCHAR(255) NOT NULL,
    numero_endereco VARCHAR(20) NOT NULL,
    complemento_endereco VARCHAR(255),
    xp_total INT DEFAULT 0,
    nivel_id INT REFERENCES aric.niveis(id),
    sequencia_dias INT DEFAULT 0
);



CREATE TABLE IF NOT EXISTS aric.projetos (
    id SERIAL PRIMARY KEY,
    titulo VARCHAR(255) NOT NULL,
    subtitulo VARCHAR(255) NOT NULL,
    descricao TEXT NOT NULL,
    dificuldade INTEGER NOT NULL REFERENCES aric.dificuldades_projeto(numero),
    xp INTEGER NOT NULL,
    tempo_estimado INTEGER NOT NULL, -- tempo estimado em minutos
    data_criacao TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    data_atualizacao TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
