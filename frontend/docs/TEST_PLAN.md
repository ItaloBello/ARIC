# Plano de testes — ARIC v3.0

Os testes de integração exigem API Go e PostgreSQL ativos, com migrations executadas e endereço acessível pelo dispositivo.

## 1. Configuração e sessão

- [ ] API indisponível: app informa que não consegue alcançar o servidor.
- [ ] URL ausente: app informa que `EXPO_PUBLIC_API_URL` não foi configurada.
- [ ] Cadastro válido com CEP, número e senha que atende à política do backend.
- [ ] Cadastro com e-mail duplicado mostra mensagem adequada.
- [ ] Cadastro com senha fraca é bloqueado.
- [ ] Login válido armazena JWT e abre o app.
- [ ] Reiniciar app restaura sessão usando `GET /auth/me`.
- [ ] Resposta 401 limpa JWT e devolve o usuário ao login.
- [ ] Logout remove JWT local.

## 2. Perfil e conta

- [ ] Carregar perfil e gamificação de `/gamificacao/perfil`.
- [ ] Atualizar nome e bio.
- [ ] Excluir conta com senha errada é recusado.
- [ ] Excluir conta com `apagarConteudo=false` anonimiza autoria.
- [ ] Excluir conta com `apagarConteudo=true` remove conteúdo conforme contrato do backend.

## 3. Projetos

- [ ] Listar projetos publicados.
- [ ] Abrir detalhe de projeto e renderizar capa, materiais e passos.
- [ ] Salvar e remover projeto salvo.
- [ ] Concluir projeto uma única vez e conferir XP no perfil.
- [ ] Testar resposta 404 para projeto indisponível.

## 4. Fórum e denúncias

- [ ] Listar posts e filtrar no cliente.
- [ ] Criar dúvida válida e confirmar publicação imediata.
- [ ] Curtir e descurtir post.
- [ ] Denunciar post com motivo válido.
- [ ] Denunciar novamente o mesmo post e conferir conflito 409.
- [ ] Moderador consulta fila agrupada de denúncias.
- [ ] Moderador decide remover e confirma remoção.
- [ ] Moderador decide manter e confirma que conteúdo permanece.
- [ ] Usuário comum recebe 403 ao acessar rotas de moderação.

## 5. Limites conhecidos

- Recuperação de senha por e-mail ainda não possui rota no backend.
- Verificação de e-mail não faz parte do fluxo atual.
- Moderação prévia por IA e suspensão de usuários não estão disponíveis.
- Upload de imagens ainda precisa ser conectado nos formulários do app.
