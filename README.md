# ARIC

## Cadastro de usuarios

### Endpoint

**(POST) /registro**

### Entradas

- nome\*
- email\*
- senha\*
- confirmarSenha\*
- foto
- bio
- cep
- bairro
- cidade
- estado
- logradouro
- numeroEndereco

**as informações com "\*" são obrigatórias**

### Funcionamento

Ao inserir as entradas, o sistema verifica se a senha esta correta e faz o hash, verifica o cep através da API viacep, faz o upload da foto e armazena o link dela  
