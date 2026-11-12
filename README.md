# ARIC

Backend em Go (Gin) + PostgreSQL (schema `aric`).

## Como rodar

1. Crie o banco (uma vez): `CREATE DATABASE aric;`
2. Conectado no banco `aric`, rode na ordem:
   - `docs/migrations/script.sql`
   - `docs/migrations/insert_estados.sql`
   - `docs/migrations/insert_cidades.sql`
   - `docs/migrations/insert_niveis.sql`
   - `docs/migrations/insert_dificuldades.sql`
   - `docs/migrations/insert_categorias.sql`
   - `docs/migrations/003_schema_completo.sql`
   - `docs/migrations/004_projetos_moderacao.sql`
   - `docs/migrations/005_missoes_evidencias.sql`
   - `docs/migrations/006_sessoes.sql`

   Banco criado com o script antigo? Rode `002_ajustes_usuarios.sql` e depois do 003 ao 006.
   Do 003 ao 006, todos podem ser rodados mais de uma vez sem erro.

   O modelo de dados, as regras de XP/ligas/missões e as consultas de cada tela estão em
   [`docs/modelo-dados.md`](docs/modelo-dados.md).
3. Copie `backend/cmd/api/.env.example` para `backend/cmd/api/.env` e preencha (o `JWT_SECRET` precisa ter 32+ caracteres).
4. Em `backend/`: `go mod tidy`
5. Em `backend/cmd/api/`: `go run .`

Testes: em `backend/`, `go test ./...`

## Autenticação

Todas as rotas ficam em `/api/v1`. Rotas protegidas exigem o header:

```
Authorization: Bearer <token>
```

O token é um JWT (HS256) que vale `JWT_EXPIRACAO_HORAS` horas (padrão 24).

A cada requisição, a API também confere no banco se a sessão ainda vale. Ela devolve `401` com
`"sessão encerrada, faça login novamente"` quando:
- **a conta foi excluída;**
- **a senha foi trocada depois que o token foi emitido.** Isso derruba os outros aparelhos.

Ao receber esse `401`, o app deve mandar o usuário para a tela de login.

### (POST) /auth/registro

Aceita `multipart/form-data` (para enviar foto) ou JSON (sem foto).

| Campo                 | Obrigatório | Observação |
|-----------------------|:-----------:|------------|
| nome                  | sim         | |
| email                 | sim         | gravado em minúsculas |
| senha                 | sim         | 8 a 72 caracteres, com maiúscula, minúscula, número e caractere especial |
| confirmarSenha        | sim         | |
| cep                   | sim         | 8 dígitos, com ou sem traço |
| numeroEndereco        | sim         | |
| complementoEndereco   | não         | |
| bio                   | não         | |
| foto                  | não         | só multipart; .jpg, .jpeg ou .png, até 5MB |
| bairro                | só se o CEP não tiver bairro | usado apenas para CEP único de cidade |
| logradouro            | só se o CEP não tiver rua    | usado apenas para CEP único de cidade |

Estado e cidade **não são enviados**: vêm do CEP (ViaCEP). O bairro é criado automaticamente
na primeira vez que aparece (`aric.bairros`, normalizado sem acento e em minúsculas).
O usuário começa no nível de menor `minimo_xp`.

A foto só é gravada depois de todas as validações, e é apagada se o INSERT falhar.

Resposta `201`: o usuário já sai logado.

```json
{
  "token": "eyJhbGciOi...",
  "expiraEm": "2026-11-04T22:00:00Z",
  "usuario": {
    "id": 1, "nome": "Maria Silva", "email": "maria@email.com",
    "fotoUrl": "/uploads/usuarios/123.jpg", "bio": null,
    "cep": "18035000", "bairroId": 4, "logradouro": "Rua X",
    "numeroEndereco": "100", "complementoEndereco": null,
    "xpTotal": 0, "nivelId": 1, "sequenciaDias": 0,
    "criadoEm": "2026-11-03T22:00:00Z"
  }
}
```

Erros (`{"error": "mensagem"}`):

- `400`: campo inválido, CEP inválido ou inexistente, cidade do CEP fora da base, CEP sem bairro/rua e campo não preenchido
- `409`: e-mail já cadastrado
- `502`: ViaCEP fora do ar

### (POST) /auth/login

JSON: `{ "email": "...", "senha": "..." }`

`200` com o mesmo formato do registro. `401` com `"e-mail ou senha inválidos"` (a mesma mensagem
para e-mail inexistente e senha errada).

### (GET) /auth/me  🔒

Devolve o objeto `usuario` do dono do token. `401` sem token ou com token inválido/expirado.

### Editar perfil  🔒

Todas devolvem o usuário atualizado (mesmo formato do `GET /auth/me`), exceto a de senha.

| Método | Rota | Corpo |
|--------|------|-------|
| PUT | `/auth/me` | `{"nome": "...", "bio": "..."}`. Bio vazia apaga a bio (máximo de 500 caracteres) |
| PUT | `/auth/me/foto` | multipart, campo `foto` (.jpg, .jpeg ou .png, até 5MB). A foto antiga é apagada do servidor |
| DELETE | `/auth/me/foto` | Remove a foto |
| PUT | `/auth/me/endereco` | `{"cep", "numeroEndereco", "complementoEndereco", "bairro", "logradouro"}`. O bairro é recalculado pelo CEP, igual ao cadastro; `bairro` e `logradouro` só são usados se o CEP não os tiver |
| PUT | `/auth/me/senha` | `{"senhaAtual", "novaSenha", "confirmarSenha"}`. Mesmas regras de senha do cadastro. Devolve um **token novo** (mesmo formato do login): o app precisa trocar o token guardado, porque o antigo deixa de valer |
| DELETE | `/auth/me` | Exclui a conta (veja abaixo) → `204` |

- **Senha atual errada** dá `400` (e não `401`), para o app não confundir com sessão expirada.
- **Troca de senha e tokens:** todos os tokens emitidos antes da troca param de valer na hora,
  em todos os aparelhos.
- **E-mail:** não muda por aqui. Trocar e-mail com segurança pede confirmação no endereço novo.

### Excluir conta  🔒

`DELETE /auth/me` com `{"senha": "...", "apagarConteudo": false}`

A senha confirma que é o dono. O app deve perguntar o que fazer com o que a pessoa publicou:

| | `apagarConteudo: false` | `apagarConteudo: true` |
|---|---|---|
| Projetos publicados | Continuam, assinados como "Usuário removido" | Removidos |
| Posts, respostas e comentários | Continuam, assinados como "Usuário removido" | Apagados. Se uma resposta era a aceita, a dúvida volta para "em aberto" |
| Imagens desses conteúdos | Ficam | Apagadas do servidor |

**Sempre acontece, nos dois casos:**
- nome, e-mail, senha, foto, bio e endereço são apagados (a conta é anonimizada);
- o e-mail fica livre para um cadastro novo;
- a foto de perfil, as fotos de "eu fiz" e as fotos de missão são apagadas do servidor;
- rascunhos e projetos em revisão ou rejeitados são removidos;
- curtidas, salvos e avaliações da pessoa somem;
- fotos de missão aguardando análise saem da fila do moderador;
- a pessoa sai do ranking da liga;
- todos os tokens dela param de valer na hora.

**O que fica, sem identificar ninguém:** o histórico de XP e de ligas, e as conclusões de
projetos. Elas continuam contando no "pessoas já fizeram" e no impacto do bairro.

`400` com `"senha incorreta"` se a senha não bater.

### Exemplo com curl

```bash
curl -X POST localhost:8080/api/v1/auth/registro \
  -F nome="Maria Silva" -F email=maria@email.com \
  -F senha='Senha@123' -F confirmarSenha='Senha@123' \
  -F cep=18035-000 -F numeroEndereco=100 -F foto=@foto.jpg

curl -X POST localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"maria@email.com","senha":"Senha@123"}'

curl localhost:8080/api/v1/auth/me -H "Authorization: Bearer $TOKEN"
```

## Projetos

Todas as rotas exigem token. Erros vêm como `{"error": "mensagem"}`.

### Fluxo de um projeto

```
rascunho ──enviar──▶ em_revisao ──aprovar──▶ publicado
                         │                      │
                         └──rejeitar──▶ oculto ◀┘ (tirar do ar)
editar (autor) ─▶ volta para em_revisao (ou rascunho)
remover (autor ou moderador) ─▶ removido
```

- Só projetos **publicados** aparecem nas listas e aceitam curtida, conclusão, avaliação e comentário.
- Rascunho, em revisão e oculto: só o autor e moderadores veem. Para os outros, a rota responde 404.
- O XP do projeto vem da dificuldade (Iniciante 80, Intermediário 150, Avançado 250).
- Na **primeira** aprovação, o autor ganha 100 XP e avança missões de "publicar projeto".
  Reaprovar depois de editar não dá XP de novo.
- Quem conclui ganha o XP do projeto e avança missões de "concluir projeto" das categorias dele.
  O autor não pode concluir o próprio projeto.

### Criar um projeto (passo a passo no app)

1. Para cada foto (galeria, materiais, passos): `POST /projetos/imagens` em multipart, campo `imagem`
   (.jpg, .jpeg, .png ou .webp, até 5MB). A resposta é `201 {"url": "/uploads/projetos/123.jpg"}`.
2. `POST /projetos` com as URLs no JSON:

```json
{
  "titulo": "Horta Vertical com Garrafas PET",
  "subtitulo": "Dê uma nova vida às garrafas",
  "descricao": "Aprenda a transformar garrafas PET em uma horta vertical...",
  "dificuldadeId": 1,
  "tempoEstimado": 25,
  "residuosKg": 1.8,
  "categorias": [1, 13],
  "materiais": [
    { "nome": "Garrafas PET", "quantidade": "3 (2 litros)", "imagemUrl": "/uploads/projetos/1.jpg" },
    { "nome": "Tesoura", "quantidade": "1", "imagemUrl": "" }
  ],
  "passos": [
    { "titulo": "Corte as garrafas", "descricao": "...", "imagemUrl": "/uploads/projetos/2.jpg" }
  ],
  "imagens": ["/uploads/projetos/capa.jpg", "/uploads/projetos/3.jpg"],
  "rascunho": false
}
```

- **A primeira imagem da lista é a capa.**
- **Rascunho:** com `"rascunho": true`, o projeto pode ficar sem materiais, passos, imagens e
  categorias.
- **Envio para revisão:** com `"rascunho": false`, o projeto precisa ter pelo menos um item de cada.
- **Limites:** até 5 categorias, 30 materiais, 30 passos e 10 imagens.
- **Resposta:** `201` com o detalhe do projeto.

### Rotas

| Método | Rota | O que faz |
|--------|------|-----------|
| GET | `/categorias` | Categorias (com `tipo`, `cor`, `icone`) |
| GET | `/dificuldades` | Dificuldades e o XP de cada uma |
| GET | `/projetos` | Lista de publicados. Query: `busca`, `categoria`, `dificuldade`, `autor`, `escopo` (`bairro` ou `cidade`, relativo a quem pede), `ordem` (`recentes`, `populares`, `relevancia`), `pagina`, `limite` (máx. 50) |
| POST | `/projetos/imagens` | Envia uma imagem (multipart, campo `imagem`) |
| POST | `/projetos` | Cria (JSON acima) |
| GET | `/projetos/meus` | Meus projetos em qualquer status (com `motivoModeracao` se rejeitado) |
| GET | `/projetos/salvos` | Projetos que salvei |
| GET | `/projetos/destaque` | Destaque da semana (`204` se não houver) |
| GET | `/projetos/:id` | Detalhe: card + descrição, imagens, materiais, passos, `minhaNota`, nº de comentários, `podeEditar` |
| PUT | `/projetos/:id` | Edita (só o autor; mesmo JSON de criar). Publicado volta para revisão |
| DELETE | `/projetos/:id` | Remove (autor ou moderador) → `204` |
| POST | `/projetos/:id/conclusao` | "Eu fiz!" Corpo opcional `{"fotoUrl": "/uploads/projetos/..."}`. Resposta: `xpGanho`, `missoesConcluidas`, `xpTotal`, `nivelId`, `nivelTitulo` |
| PUT / DELETE | `/projetos/:id/curtida` | Curte / descurte → `{"curtido", "curtidas"}` |
| PUT / DELETE | `/projetos/:id/salvo` | Salva / tira dos salvos |
| PUT | `/projetos/:id/avaliacao` | `{"nota": 1..5}`. Só quem concluiu (`403` caso contrário) |
| GET | `/projetos/:id/comentarios` | Comentários (mais antigos primeiro), com `podeApagar` |
| POST | `/projetos/:id/comentarios` | `{"conteudo": "..."}` (até 2000 caracteres) |
| DELETE | `/projetos/:id/comentarios/:comentarioId` | Apaga (autor do comentário ou moderador) |

**Moderação** (só usuários com `papel` moderador ou admin; os outros recebem `403`):

| Método | Rota | O que faz |
|--------|------|-----------|
| GET | `/moderacao/projetos` | Fila de revisão, mais antigos primeiro |
| POST | `/moderacao/projetos/:id/aprovar` | Publica |
| POST | `/moderacao/projetos/:id/rejeitar` | `{"motivo": "..."}`. Funciona em revisão ou já publicado (tira do ar). O autor vê o motivo |
| PUT | `/moderacao/destaque` | `{"projetoId": 1}`. Define o destaque da semana atual |

Um moderador não aprova nem rejeita o próprio projeto.

Para tornar alguém moderador: `UPDATE aric.usuarios SET papel = 'moderador' WHERE email = '...';`
O `GET /auth/me` devolve o `papel`, para o app saber se mostra a área de moderação.

### Card de projeto (listas)

```json
{
  "id": 1, "titulo": "...", "subtitulo": "...", "status": "publicado",
  "xp": 80, "tempoEstimado": 25, "residuosKg": 1.8,
  "dificuldade": { "id": 1, "descricao": "Iniciante" },
  "capaUrl": "/uploads/projetos/capa.jpg",
  "autor": { "id": 3, "nome": "Juliana M.", "fotoUrl": null, "bairro": "Jardim das Flores" },
  "categorias": [{ "id": 1, "nome": "Plástico", "cor": "#F2B705" }],
  "curtidas": 24, "conclusoes": 1200, "notaMedia": 4.6,
  "curtido": false, "salvo": true, "concluido": false,
  "criadoEm": "...", "publicadoEm": "..."
}
```

`autor` é `null` em projetos oficiais do ARIC. Listas vêm como `{"itens": [...], "total", "pagina", "limite"}`.

## Fórum

Todas as rotas exigem token e ficam em `/api/v1/forum`.

### Regras

- **Publicação imediata:** o fórum não passa por revisão. Moderadores podem apagar qualquer post ou
  resposta.
- **Tipos de post:** `duvida` ("Perguntar") ou `dica`. O tipo não muda depois de criado.
- **Resolvido:** é uma dúvida com resposta aceita.
  - Só o autor da dúvida aceita uma resposta, e não pode aceitar a própria.
  - Ele pode trocar a resposta aceita ou desmarcar.
- **XP:** quem teve a resposta aceita ganha **50 XP**, uma vez por resposta. Trocar ou desmarcar
  depois não tira o XP.
- **Missões:**
  - criar uma dúvida avança as missões de "perguntar";
  - responder o post de **outra pessoa** avança as de "responder" (responder o próprio post não
    conta).
- **Resposta aceita apagada:** a dúvida volta para "em aberto".
- **Imagens:** cada post pode ter uma. Envie antes em `POST /forum/imagens` (multipart, campo
  `imagem`) e use a URL devolvida.

### Rotas

| Método | Rota | O que faz |
|--------|------|-----------|
| GET | `/forum/posts` | Lista. Query: `busca`, `categoria`, `tipo` (`duvida` ou `dica`), `situacao` (`abertas` ou `resolvidas`), `autor`, `ordem` (`recentes`, `votadas` ou `relevancia`), `pagina`, `limite` |
| POST | `/forum/posts` | Cria: `{"tipo", "titulo", "conteudo", "imagemUrl", "categoriaId"}`. A resposta traz `missoesConcluidas` |
| POST | `/forum/imagens` | Envia uma imagem |
| GET | `/forum/posts/salvos` | Posts que salvei |
| GET | `/forum/posts/:id` | Detalhe com o texto completo e `podeEditar`, `podeApagar` e `podeAceitar` |
| PUT | `/forum/posts/:id` | Edita título, texto, imagem e categoria (só o autor) |
| DELETE | `/forum/posts/:id` | Apaga (autor ou moderador) |
| PUT / DELETE | `/forum/posts/:id/curtida` | Curte / descurte |
| PUT / DELETE | `/forum/posts/:id/salvo` | Salva / tira dos salvos |
| PUT | `/forum/posts/:id/resposta-aceita` | `{"respostaId": 3}`: marca como resolvida |
| DELETE | `/forum/posts/:id/resposta-aceita` | Volta para "em aberto" |
| GET | `/forum/posts/:id/respostas` | Respostas: a aceita primeiro, depois as mais curtidas e depois as mais antigas |
| POST | `/forum/posts/:id/respostas` | `{"conteudo": "..."}`. A resposta traz `missoesConcluidas` |
| PUT | `/forum/respostas/:id` | Edita (só o autor) |
| DELETE | `/forum/respostas/:id` | Apaga (autor ou moderador) |
| PUT / DELETE | `/forum/respostas/:id/curtida` | Curte / descurte |

**Abas da tela:**

| Aba | Chamada |
|-----|---------|
| Mais votadas | `?ordem=votadas` |
| Recentes | `?ordem=recentes` |
| Dúvidas abertas | `?situacao=abertas` |
| Dicas rápidas | `?tipo=dica` |

Textos: título de 5 a 200 caracteres; post e resposta com até 5000.

### Card de post

```json
{
  "id": 1, "tipo": "duvida", "titulo": "O que fazer com caixas de ovos?",
  "resumo": "primeiros 200 caracteres...", "imagemUrl": null,
  "resolvido": true, "respostaAceitaId": 7,
  "autor": { "id": 3, "nome": "Ana", "fotoUrl": null, "bairro": "Centro" },
  "categoria": { "id": 2, "nome": "Papelão", "cor": "#B07A4A" },
  "curtidas": 32, "respostas": 12, "curtido": false, "salvo": false,
  "criadoEm": "...", "atualizadoEm": "..."
}
```

## Gamificação

Todas as rotas exigem token.

### Rotas

| Método | Rota | O que faz |
|--------|------|-----------|
| GET | `/gamificacao/perfil` | Meu perfil: nível, barra até o próximo, XP total, sequência, liga com posição e impacto |
| GET | `/usuarios/:id/perfil` | Perfil público de outra pessoa (sem e-mail e endereço) |
| GET | `/gamificacao/niveis` | Os 10 níveis com o XP mínimo de cada |
| GET | `/gamificacao/xp` | Meu histórico de XP (paginado) |
| GET | `/gamificacao/impacto-bairro` | Card "Impacto do bairro" da Home: kg reaproveitados pelos moradores do meu bairro hoje, na semana e no total |
| GET | `/gamificacao/ligas` | As ligas, da mais baixa para a mais alta |
| GET | `/gamificacao/ranking` | Tela Comunidade: liga, top 50, minha posição, fim da semana e resultado da semana passada |
| GET | `/gamificacao/missoes` | Missões da semana com progresso |
| POST | `/gamificacao/missoes/imagens` | Envia a foto de uma missão (multipart, campo `imagem`) |
| POST | `/gamificacao/missoes/:id/evidencias` | `{"fotoUrl", "quantidade"}`: manda a foto para análise |
| GET | `/gamificacao/evidencias` | Minhas fotos enviadas, com status e motivo da rejeição |

**Moderação** (só moderador ou admin):

| Método | Rota | O que faz |
|--------|------|-----------|
| GET | `/moderacao/evidencias` | Fotos aguardando análise, mais antigas primeiro |
| POST | `/moderacao/evidencias/:id/aprovar` | Soma no progresso. Se bater a meta, conclui a missão e dá o XP |
| POST | `/moderacao/evidencias/:id/rejeitar` | `{"motivo": "..."}`. O usuário vê o motivo |

### Perfil

```json
{
  "id": 1, "nome": "Ana", "fotoUrl": null, "bio": null,
  "bairro": "Centro", "cidade": "Sorocaba", "uf": "SP", "membroDesde": "...",
  "xpTotal": 850,
  "nivel": { "numero": 4, "titulo": "Coletor Consciente", "minimoXp": 600 },
  "proximoNivel": { "numero": 5, "titulo": "Reciclador", "minimoXp": 1000 },
  "xpParaProximo": 150, "progressoNivel": 62,
  "sequenciaDias": 18,
  "liga": { "id": 2, "nome": "Liga Broto", "ordem": 2, "icone": null, "posicao": 14, "xpSemana": 184 },
  "impacto": { "residuosKg": 12.4, "projetosConcluidos": 9, "projetosPublicados": 2, "vidasInspiradas": 12 }
}
```

- **Topo das telas:** use `xpTotal` (são os EcoPontos).
- **`sequenciaDias`:** já vem zerada se o usuário pulou um dia.
- **`proximoNivel`:** é `null` no nível máximo.
- **`vidasInspiradas`:** quantas pessoas diferentes fizeram projetos publicados por este usuário.
- **Perfil da tela:** troque "#14 no Bairro" por `liga.posicao` + `liga.nome`.

### Impacto do bairro (Home)

```json
{
  "bairroId": 1, "bairro": "Centro",
  "kgHoje": 184.2, "projetosHoje": 37, "pessoasHoje": 21,
  "kgSemana": 512.0, "kgTotal": 3120.5
}
```

- **O que soma:** o `residuosKg` de cada projeto concluído por moradores do bairro.
- **"Hoje" e "semana":** seguem o horário de São Paulo.
- **Bairro sem atividade:** tudo vem 0.
- **Na tela:** "Hoje evitamos 184 kg" usa `kgHoje`.

### Ranking (tela Comunidade)

- **Liga:** é a do usuário, com o XP desde segunda.
- **Linhas:**
  - `zona`: `sobe`, `desce` ou `mantem`, se a semana fechasse agora. Serve para desenhar as linhas
    de corte.
  - `tendencia`: posições ganhas (+) ou perdidas (-) desde ontem.
- **Liga inteira:**
  - `primeira: true`: ninguém desce.
  - `ultima: true`: ninguém sobe.
  - `fimSemana`: o contador "termina em 2d 4h".
- **`minhaPosicao`:** vem sempre, mesmo fora do top 50.
- **`ultimaSemana`:** traz o resultado da semana passada (`subiu`, `desceu` ou `manteve`, com a liga
  anterior e a nova), para mostrar "Você subiu para a Liga Broto!". É `null` antes da primeira
  semana fechada.

### Missões

- **Missões automáticas** (`enviaFoto: false`): avançam sozinhas quando o usuário conclui um projeto,
  tem um projeto aprovado, pergunta no fórum ou responde no fórum. Isso acontece nas rotas desses
  módulos, que devolvem `missoesConcluidas` quando alguma termina.
- **Missões com foto** (`enviaFoto: true`), por exemplo "entregue 10 garrafas":
  1. O app envia a foto em `/gamificacao/missoes/imagens`.
  2. O app chama `/gamificacao/missoes/:id/evidencias` com a quantidade.
  3. Um moderador aprova; a quantidade soma no progresso e, ao bater a meta, a missão conclui e dá
     o XP.
- **Limites das fotos:**
  - até 3 fotos em análise por missão;
  - quantidade de 1 até a meta;
  - não aceita foto de missão já concluída na semana.
- **Aprovação depois da virada:** uma foto aprovada na semana seguinte conta para a semana em que
  foi enviada.
- **Um moderador não analisa a própria foto.**

### Fechamento semanal das ligas

A API fecha a semana sozinha: ao subir e depois a cada hora, ela verifica se a semana anterior já
foi fechada e, se não foi, chama `aric.fechar_semana_ligas`. Se o servidor estiver fora do ar na
virada de segunda, o fechamento acontece assim que ele voltar.

Para usar o cron do servidor em vez disso, defina `LIGAS_JOB=off` e agende:

```
5 0 * * 1  psql -d aric -c "SELECT aric.fechar_semana_ligas(aric.inicio_semana(now()) - 7);"
```

Esse horário vale se o servidor estiver no fuso de São Paulo.

## Usando o login em novas rotas

```go
auth := middleware.Auth(tokens)
v1.POST("/projetos", auth, ct.Criar)

// dentro do handler:
usuarioID, ok := middleware.UsuarioID(c)
```
