# Convenções do Projeto

Esta pasta concentra documentos de convenção que nao sao:

- plano de execução por fases;
- ADR formal;
- guia de onboarding.

O objetivo aqui e registrar regras compartilhadas entre apps e libs que ajudam a manter navegação, contratos de uso e organização coerentes ao longo do tempo.

## Convenções disponíveis

- [company-slug-routing.md](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/docs/conventions/company-slug-routing.md)
  Convenção de rotas autenticadas com `company_slug`, separando claramente contexto de URL/UX de autorização baseada em JWT e `company_id`.

## Relação com outras pastas de `docs/`

- `docs/plans/`: execução incremental, fases, checks e ordem de entrega.
- `docs/adr/`: decisões arquiteturais mais amplas e permanentes.
- `docs/conventions/`: regras compartilhadas de operação e organização entre apps e libs.
