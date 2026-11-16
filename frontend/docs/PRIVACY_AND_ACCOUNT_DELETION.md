# Privacidade e exclusão de conta — ARIC v3.0

O app usa a API Go + Gin e o banco PostgreSQL do projeto `ItaloBello/ARIC`. Antes de publicar, a equipe precisa identificar o responsável pelo servidor, sua região de hospedagem, os operadores que terão acesso administrativo e os prazos de retenção.

## Dados tratados pelo backend

- Nome, e-mail, senha armazenada como hash, foto de perfil e bio.
- CEP, bairro, cidade/UF, logradouro, número e complemento do endereço.
- Projetos, conclusões, curtidas, favoritos/salvos, avaliações e comentários.
- Posts, respostas, curtidas e denúncias do fórum.
- Histórico de XP, nível, sequência de dias, ligas e missões mantidas no backend.
- Registros técnicos de tentativas de acesso e sessões para segurança.

## Comportamento do fórum

O backend atual publica posts e respostas imediatamente. Não existe revisão prévia por IA. Usuários podem denunciar conteúdo; moderadores/admins analisam as denúncias e decidem remover ou manter o conteúdo.

## Exclusão de conta

O app chama `DELETE /api/v1/auth/me` com a senha e o campo `apagarConteudo`.

- `apagarConteudo: false`: o conteúdo publicado permanece, mas é atribuído a “Usuário removido”.
- `apagarConteudo: true`: projetos, posts, respostas, comentários e imagens publicados pela conta são apagados conforme as regras do backend.

Nos dois casos, dados pessoais são anonimizados/removidos, sessões são invalidadas e o e-mail fica livre para novo cadastro. O histórico agregado de XP/ligas e conclusões pode permanecer sem identificar a pessoa, conforme a regra documentada no backend.

## Antes da publicação

- Preencher os dados do controlador e do canal de privacidade/suporte.
- Confirmar região e operador da hospedagem da API e PostgreSQL.
- Confirmar os logs, backups e prazos de retenção.
- Revisar a política com responsável jurídico/privacidade e adequar à LGPD.
- Publicar URLs públicas e funcionais para política e exclusão de conta.
