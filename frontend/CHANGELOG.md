# Changelog

## 3.0.0 — Migração para API Go + PostgreSQL

- Substituído Firebase Authentication/Firestore/Cloud Functions por API REST com JWT.
- Implementado armazenamento seguro da sessão JWT no AsyncStorage e validação inicial via `/auth/me`.
- Cadastro alinhado aos campos obrigatórios do backend: nome, e-mail, senha, CEP e número do endereço.
- Atualizadas as integrações de perfil, catálogo de projetos, projetos salvos/conclusão, fórum, curtidas, denúncias, moderação e impacto do bairro.
- Simplificada a navegação principal para Início, Chat e Perfil.
- Alinhado o fórum ao backend: publicação imediata, sem fila de aprovação prévia.
- Removidas as dependências e configurações ativas do Firebase.
- Documentados recursos ainda não disponíveis no backend atual.

## 2.1.2

- Versão anterior baseada em Firebase, com revisão prévia de publicações e gamificação local/sincronizada.
