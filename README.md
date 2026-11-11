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

   Banco criado com o script antigo? Rode `002_ajustes_usuarios.sql`, depois o 003 e o 004.
   O 003 e o 004 podem ser rodados mais de uma vez sem erro.

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

## Usando o login em novas rotas

```go
auth := middleware.Auth(tokens)
v1.POST("/projetos", auth, ct.Criar)

// dentro do handler:
usuarioID, ok := middleware.UsuarioID(c)
```
