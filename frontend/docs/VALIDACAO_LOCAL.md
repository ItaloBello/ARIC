# Validação técnica desta entrega

Foram executadas nesta entrega:

- validação sintática de todos os arquivos `.ts` e `.tsx`;
- verificação dos caminhos de imports locais;
- verificação de todos os assets referenciados por `require()`;
- verificação de tokens de cores utilizados;
- verificação das rotas internas;
- validação JSON de `package.json`, `app.json` e `eas.json`;
- inspeção visual do conjunto de imagens locais para garantir relação com o conteúdo exibido.

O ambiente que gerou este ZIP não conseguiu concluir a instalação dos pacotes npm pela rede, portanto a validação semântica final precisa ser executada no PC de desenvolvimento após `npm install`:

```powershell
npm.cmd install
npx.cmd expo install --fix
npm.cmd run typecheck
npx.cmd expo-doctor
npx.cmd expo start -c
```

Não gere o build de produção enquanto `typecheck` ou `expo-doctor` indicarem problema crítico.
