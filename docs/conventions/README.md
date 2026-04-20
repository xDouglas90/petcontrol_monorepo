# Convenções do Projeto

Esta pasta concentra documentos de convenção que nao sao:

- plano de execução por fases;
- ADR formal;
- guia de onboarding.

O objetivo aqui e registrar regras compartilhadas entre apps e libs que ajudam a manter navegação, contratos de uso e organização coerentes ao longo do tempo.

## Convenções disponíveis

- [company-slug-routing.md](./company-slug-routing.md)
  Convenção de rotas autenticadas com `company_slug`, separando claramente contexto de URL/UX de autorização baseada em JWT e `company_id`.
- [internal-chat-realtime.md](./internal-chat-realtime.md)
  Contrato inicial de WebSocket, presença dinâmica e eventos do chat interno entre `admin` e `system`.
- [modules.md](./modules.md)
  Códigos, descrições e pacotes mínimos de cada módulo do sistema.
- [permissions.md](./permissions.md)
  Lista mestre de permissões e atribuição padrão por papel.
- [plans.md](./plans.md)
  Definição dos planos de assinatura (Starter, Basic, Essential, Premium).
- [users-types.md](./users-types.md)
  Diferenciação entre papéis sistêmicos (Role) e tipos de vínculo de negócio (Kind).

## Relação com outras pastas de `docs/`

- `docs/plans/`: execução incremental, fases, checks e ordem de entrega.
- `docs/adr/`: decisões arquiteturais mais amplas e permanentes.
- `docs/conventions/`: regras compartilhadas de operação e organização entre apps e libs.
