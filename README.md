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
   - Banco criado com o script antigo? Rode também `docs/migrations/002_ajustes_usuarios.sql`.
3. Copie `backend/cmd/api/.env.example` para `backend/cmd/api/.env` e preencha (o `JWT_SECRET` precisa ter 32+ caracteres).
4. Em `backend/`: `go mod tidy`
5. Em `backend/cmd/api/`: `go run .`

Testes: em `backend/`, `go test ./...`

## Autenticação

Todas as rotas ficam em `/api/v1`. Rotas protegidas exigem o header:

``` CONF
Authorization: Bearer <token>
```

O token é um JWT (HS256) que vale `JWT_EXPIRACAO_HORAS` horas (padrão 24).

### (POST) /auth/registro

Aceita `multipart/form-data` (para enviar foto) ou JSON (sem foto).

| Campo               |         Obrigatório          | Observação                                                               |
| ------------------- | :--------------------------: | ------------------------------------------------------------------------ |
| nome                |             sim              |                                                                          |
| email               |             sim              | gravado em minúsculas                                                    |
| senha               |             sim              | 8 a 72 caracteres, com maiúscula, minúscula, número e caractere especial |
| confirmarSenha      |             sim              |                                                                          |
| cep                 |             sim              | 8 dígitos, com ou sem traço                                              |
| numeroEndereco      |             sim              |                                                                          |
| complementoEndereco |             não              |                                                                          |
| bio                 |             não              |                                                                          |
| foto                |             não              | só multipart; .jpg, .jpeg ou .png, até 5MB                               |
| bairro              | só se o CEP não tiver bairro | usado apenas para CEP único de cidade                                    |
| logradouro          |  só se o CEP não tiver rua   | usado apenas para CEP único de cidade                                    |

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
    "id": 1,
    "nome": "Maria Silva",
    "email": "maria@email.com",
    "fotoUrl": "/uploads/usuarios/123.jpg",
    "bio": null,
    "cep": "18035000",
    "bairroId": 4,
    "logradouro": "Rua X",
    "numeroEndereco": "100",
    "complementoEndereco": null,
    "xpTotal": 0,
    "nivelId": 1,
    "sequenciaDias": 0,
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

### (GET) /auth/me 🔒

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

## Usando o login em novas rotas

```go
auth := middleware.Auth(tokens)
v1.POST("/projetos", auth, ct.Criar)

// dentro do handler:
usuarioID, ok := middleware.UsuarioID(c)
```
