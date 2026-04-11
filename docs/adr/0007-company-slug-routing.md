# ADR 0007: company_slug como contexto de navegação Web

## Status

Aceito

## Contexto

O Web precisa deixar o tenant atual visível na URL e preservar contexto ao navegar por páginas autenticadas. Ao mesmo tempo, autorização não pode depender de dados controlados pela URL.

## Decisão

Usar `/:companySlug` como prefixo das rotas autenticadas do Web, tratando o slug como contexto de navegação e UX. A fonte de verdade de autorização continua sendo o JWT e o `company_id` resolvido pelo backend.

## Consequências

- URLs autenticadas ficam mais legíveis e contextualizadas por empresa.
- O frontend deve corrigir mismatch entre slug da URL e slug real da empresa corrente.
- `/login` permanece sem slug, pois ainda não há sessão nem tenant resolvido.
- Mudanças futuras de slug exigem estratégia de redirecionamento ou compatibilidade para links antigos.
