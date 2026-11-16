# Endurecimento de produção — ARIC v3.0

## API e infraestrutura

- Publicar a API Go apenas por HTTPS.
- Manter `JWT_SECRET` no ambiente do servidor, com pelo menos 32 caracteres aleatórios; nunca no app Expo.
- Restringir acesso ao PostgreSQL e manter backups testados.
- Configurar `TRUSTED_PROXIES` apenas com proxies confiáveis; não confiar indiscriminadamente em `X-Forwarded-For`.
- Configurar CORS para origens necessárias, limites de requisição e monitoramento de erros.
- Proteger `/uploads` contra execução de arquivos e validar extensão, MIME type e tamanho.

## Autenticação e autorização

- Todas as rotas privadas devem validar JWT e sessão no banco.
- Manter rate limits de login/cadastro e monitorar respostas 429.
- Conceder papel `moderador`/`admin` apenas por procedimento controlado no banco.
- Nunca confiar em XP, IDs de usuário ou permissões enviados pelo cliente.

## Comunidade

- O backend atual publica posts imediatamente; denúncias são pós-publicação.
- Definir um responsável e prazo operacional para revisar denúncias.
- Não comunicar que existe moderação por IA, aprovação prévia ou suspensão de usuários enquanto essas funções não estiverem implementadas no backend.
- Implementar upload de imagem no app antes de exibir anexos como recurso disponível.

## Dados e privacidade

- Documentar hospedagem, operadores, região, backups e retenção.
- Testar exclusão de conta nos modos `apagarConteudo=false` e `true`.
- Revisar os dados preservados de XP/ligas/conclusões para garantir anonimização adequada.
- Não publicar builds com política de privacidade genérica ou URLs vazias.
