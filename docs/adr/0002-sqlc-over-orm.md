# ADR 0002: SQLC em vez de ORM

## Status

Aceito

## Contexto

O domínio requer controle explicito de queries, tuning de indices e previsibilidade de comportamento SQL em ambiente multi-tenant.

## Decisão

Usar SQLC como camada de acesso a dados, com queries SQL versionadas em infra/sql/queries e schema vindo de migrations.

## Consequências

- Queries explicitas e revisáveis no PR.
- Tipagem forte no Go gerada automaticamente.
- Menor risco de SQL inesperado em produção comparado a ORMs genéricos.
