# Entrega — frontend ARIC v3.0

Esta pasta contém o projeto de interface do aplicativo ARIC.

## Tecnologias
- React Native
- Expo
- TypeScript
- AsyncStorage para armazenar o token de sessão
- API REST com autenticação JWT (backend Go separado)

## Pastas
- `src/screens/`: um arquivo `.tsx` para cada tela
- `src/components/`: componentes visuais reutilizáveis
- `src/services/`: integração com a API
- `src/context/`: estados globais
- `src/hooks/`: hooks
- `assets/`: imagens e identidade visual

Consulte `ESTRUTURA_TELAS.md` para a lista das telas e a situação de cada uma na navegação.

## Executar
1. Instale Node.js LTS.
2. Instale as dependências: `npm install`.
3. Copie `.env.example` para `.env`.
4. Configure `EXPO_PUBLIC_API_URL` com o endereço da API Go, terminando em `/api/v1`.
5. Execute `npm run typecheck` e depois `npx expo start -c`.

Para celular físico, use o IP local do computador na mesma rede Wi-Fi; `localhost` aponta para o próprio celular e não para o computador.

**Atenção:** o backend Go e o PostgreSQL não estão dentro desta pasta. A integração com o servidor só pode ser testada quando a API estiver no ar. O Chat representa um fórum público, não mensagens privadas em tempo real.
