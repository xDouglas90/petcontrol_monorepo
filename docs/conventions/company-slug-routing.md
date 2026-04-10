# Convencao de Rotas com `company_slug`

## Contexto

Esta documentação registra a convenção planejada para a area autenticada do Web, conforme o plano [0003-COMPANY_SLUG_ROUTING_PLAN.md](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/docs/plans/0003-COMPANY_SLUG_ROUTING_PLAN.md).

Ela nao substitui o plano de execução. O objetivo aqui é documentar a regra funcional e arquitetural que deve permanecer valida mesmo depois da implementação.

## Regra principal

O `company_slug` na URL e contexto de navegação e UX.

Ele **nao** e mecanismo de autorização.

O backend continua confiando apenas em:

- JWT valido;
- `company_id` presente no token;
- middlewares de tenant e permissão.

## Convenção planejada para a area autenticada

- `/:companySlug/dashboard`
- `/:companySlug/schedules`

## Rotas que permanecem sem slug

- `/login`

## Estado atual

Durante a Fase 0 do plano `0003`, a convenção foi formalizada em `shared-constants`, mas as rotas vivas da aplicação ainda podem permanecer planas ate o inicio da migração do router na Fase 1.

## Relação com outros tipos de documento

- `docs/plans/`: descrevem execução por fases, checks e ordem de entrega.
- `docs/adr/`: registram decisões arquiteturais mais amplas e duradouras.
- `docs/conventions/`: documentam regras operacionais e convenções compartilhadas entre apps e libs.
