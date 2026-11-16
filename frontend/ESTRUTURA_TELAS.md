# ARIC — catálogo do frontend

Este pacote contém somente o **frontend móvel** do ARIC, desenvolvido em React Native + Expo + TypeScript. O backend Go e o banco PostgreSQL não estão incluídos.

## Telas em arquivos individuais

Os componentes de tela estão em `src/screens/`:

| Arquivo | Responsabilidade | Situação na navegação v3.0 |
|---|---|---|
| `HomeScreen.tsx` | Início, pesquisa, categorias e projetos em destaque | Aba principal |
| `ForumScreen.tsx` | Chat comunitário público / fórum | Aba principal |
| `ProfileScreen.tsx` | Perfil, dados e atalhos da conta | Aba principal |
| `LoginScreen.tsx` | Entrada na conta | Fluxo de autenticação |
| `RegisterScreen.tsx` | Cadastro | Fluxo de autenticação |
| `ForgotPasswordScreen.tsx` | Interface de recuperação de acesso | Fluxo de autenticação; confirmar suporte no backend |
| `ProjectsScreen.tsx` | Catálogo e filtros de projetos | Acessada a partir de Início |
| `ProjectDetailScreen.tsx` | Detalhes de um projeto | Acessada a partir do catálogo |
| `AskQuestionScreen.tsx` | Criar publicação/pergunta no fórum | Acessada a partir do Chat |
| `MySubmissionsScreen.tsx` | Publicações do usuário | Acessada a partir do Perfil |
| `EditProfileScreen.tsx` | Edição do perfil | Acessada a partir do Perfil |
| `AdminModerationScreen.tsx` | Moderação de denúncias | Acesso de administração/moderação |
| `PrivacyScreen.tsx` | Privacidade e exclusão de conta | Fluxo de conta |
| `CommunityGuidelinesScreen.tsx` | Regras da comunidade | Fluxo de apoio ao Chat |
| `MissionsScreen.tsx` | Tela de missões da versão anterior | Arquivo mantido; não conectado à navegação principal v3.0 |
| `RankingScreen.tsx` | Tela de ranking da versão anterior | Arquivo mantido; não conectado à navegação principal v3.0 |

A navegação principal está limitada às três abas **Início, Chat e Perfil**. Os arquivos de Missões e Ranking foram mantidos no pacote como código legado/auxiliar, mas não são exibidos pelo `App.tsx` atual.

## Outras pastas importantes

- `src/components/`: componentes reutilizáveis de interface (botões, campos, cartões, cabeçalho, navegação etc.).
- `src/context/`: estados compartilhados de autenticação, aplicativo e comunidade.
- `src/services/`: comunicação HTTP com a API, projetos, comunidade, progresso e moderação.
- `src/hooks/`: hooks reutilizáveis.
- `src/navigation.ts`: tipos das rotas e das três abas.
- `src/theme.ts`: cores e tema visual.
- `assets/`: imagens e ícones usados pelo app.

## Observações importantes

1. Este ZIP é o código do frontend; não contém o servidor Go nem o banco PostgreSQL.
2. O aplicativo só realiza integração real quando `EXPO_PUBLIC_API_URL` aponta para uma API Go acessível e compatível com os endpoints usados em `src/services/`.
3. A API deve estar em execução e conectada ao PostgreSQL. O código no GitHub, isoladamente, não inicia um servidor.
4. O Chat desta versão é um fórum público de publicações e respostas, não um chat privado em tempo real.
5. O fluxo de recuperação de senha precisa ser conferido com o backend antes de ser considerado funcional; a documentação de integração registra essa limitação.
6. A integração de ponta a ponta não foi testada nesta entrega porque não há uma instância do backend em execução disponível nesta sessão.
