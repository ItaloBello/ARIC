# ARIC v3.0 — Aplicativo de Reciclagem Inteligente da Comunidade

Aplicativo móvel em **React Native + Expo + TypeScript**, integrado ao backend oficial **Go + Gin + PostgreSQL** do repositório `ItaloBello/ARIC`.

A v3.0 mantém as telas e componentes principais da versão anterior, simplifica a navegação principal para **Início, Chat e Perfil** e substitui Firebase por uma API REST com JWT.

> A aba **Chat** é um espaço público de perguntas e dicas da comunidade. O backend atual não implementa mensagens privadas em tempo real.

## 1. Configurar a API

O aplicativo não se conecta diretamente ao GitHub. O repositório contém o código-fonte do servidor; para o app funcionar online, o backend Go precisa estar executando e conectado ao PostgreSQL.

No repositório `ItaloBello/ARIC`, siga o README principal para:

1. Criar o banco PostgreSQL `aric`.
2. Executar as migrations na ordem indicada pelo backend.
3. Configurar `backend/cmd/api/.env`, incluindo `JWT_SECRET` com pelo menos 32 caracteres.
4. Executar `go mod tidy` em `backend/`.
5. Iniciar a API com `go run .` em `backend/cmd/api/`.

A API local normalmente fica disponível em `http://localhost:8080`, com rotas em `/api/v1`.

## 2. Configurar o aplicativo

Copie `.env.example` para `.env` e ajuste `EXPO_PUBLIC_API_URL`:

```env
# Android Emulator
EXPO_PUBLIC_API_URL=http://10.0.2.2:8080/api/v1

# iOS Simulator
# EXPO_PUBLIC_API_URL=http://127.0.0.1:8080/api/v1

# Celular físico: use o IP local do computador na mesma rede Wi-Fi
# EXPO_PUBLIC_API_URL=http://192.168.0.10:8080/api/v1
```

Para produção, use o endereço HTTPS público do backend, sempre terminando em `/api/v1`.

Depois execute:

```bash
npm install
npm run typecheck
npx expo start -c
```

## 3. Integrações implementadas no app

- Autenticação: `POST /auth/login`, `POST /auth/registro` e `GET /auth/me`.
- Sessão: JWT armazenado no AsyncStorage e enviado como `Authorization: Bearer <token>`.
- Cadastro: nome, e-mail, senha, CEP e número do endereço. O backend consulta o CEP para obter cidade/estado/bairro.
- Perfil: consulta de gamificação e atualização de nome/bio.
- Exclusão de conta: confirmação por senha e escolha sobre manter ou apagar conteúdo publicado.
- Projetos: listagem, detalhe, favoritos/salvos e conclusão.
- Comunidade: listagem, criação de posts, curtidas e denúncias.
- Moderação: consulta e resolução de denúncias por moderadores/admins.
- Home: consulta do impacto do bairro.
- Modo visitante/demonstração: usa dados locais quando não há sessão de API.

## 4. Diferenças conhecidas entre a v2.1 e o backend Go

O app foi adaptado ao comportamento documentado do backend, sem manter promessas de recursos que a API não oferece:

- **Fórum:** publicação imediata; não existe fila de aprovação prévia.
- **IA:** o backend atual não expõe reanálise automática de posts por IA.
- **Suspensões:** não há rotas para suspender/restaurar usuários.
- **Recuperação de senha:** ainda não há rota de recuperação por e-mail. A troca de senha autenticada existe em `PUT /auth/me/senha`, mas ainda precisa de uma tela própria no app.
- **Verificação de e-mail:** não é exigida nem fornecida pelo backend atual.
- **Missões:** o backend tem missões automáticas e missões com evidências/fotos. O antigo incremento local genérico da v2.1 não deve ser usado como progresso oficial.
- **Atualização em tempo real:** a API é REST; as telas atualizam os dados por requisições, não por listeners Firestore.
- **Upload de imagens:** o backend oferece rotas de upload, mas os formulários de criação do app ainda precisam ser conectados a esses uploads antes de liberar anexos.

## 5. Segurança

- Nunca coloque `JWT_SECRET`, credenciais do PostgreSQL ou segredos do servidor no `.env` do Expo.
- O app armazena apenas o JWT de sessão; as regras de autorização são aplicadas no backend.
- Não use `localhost` no celular físico: nesse caso, configure o IP do computador na rede local.
- Em produção, publique a API atrás de HTTPS e configure CORS/proxy conforme o README do backend.

## 6. Estrutura principal

```text
src/
  context/
    AuthContext.tsx          autenticação JWT e sessão
    AppContext.tsx           perfil, favoritos e progresso local
    CommunityContext.tsx     posts, curtidas e denúncias
  hooks/
    useProjects.ts           catálogo e detalhes de projetos
  services/
    api.ts                   cliente HTTP, JWT e URLs de mídia
    community.ts             integração do fórum e moderação
    progress.ts              perfil, XP e impacto do bairro
    projects.ts              catálogo, detalhe e projetos salvos
  screens/                   telas do aplicativo
```

## 7. Validação

Execute `npm run typecheck` antes de gerar builds. A integração de rede só pode ser validada de ponta a ponta quando a API Go e o PostgreSQL estiverem executando em um endereço acessível pelo dispositivo.
