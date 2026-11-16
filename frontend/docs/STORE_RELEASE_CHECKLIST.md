# Checklist de publicação — ARIC v3.0

## Backend
- [ ] API Go publicada em HTTPS.
- [ ] PostgreSQL provisionado, migrations aplicadas e backup testado.
- [ ] `JWT_SECRET` forte configurado apenas no servidor.
- [ ] CORS, proxies confiáveis, rate limits e `/uploads` revisados.
- [ ] Contas de moderador/admin configuradas com acesso mínimo.
- [ ] Fila de denúncias possui responsável e SLA operacional.

## Aplicativo
- [ ] `EXPO_PUBLIC_API_URL` aponta para a API de produção.
- [ ] Login, cadastro, sessão, logout e 401 testados em Android/iOS.
- [ ] Cadastro testado com CEP sem bairro/logradouro e campos de fallback.
- [ ] Projetos, favoritos, conclusão, fórum, curtidas e denúncias testados.
- [ ] Exclusão de conta testada nos dois modos.
- [ ] Recuperação de senha não é anunciada como funcional enquanto não houver rota.
- [ ] Upload de imagens não é anunciado como funcional antes da integração.

## Privacidade e lojas
- [ ] Política de privacidade pública, revisada e coerente com o backend real.
- [ ] URL pública de exclusão de conta configurada.
- [ ] Canal de suporte configurado.
- [ ] Região da hospedagem, operadores, retenção e backups documentados.
- [ ] Classificação etária e declarações de dados preenchidas com precisão.
