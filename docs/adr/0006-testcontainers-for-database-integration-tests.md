# ADR 0006: Testcontainers para testes de integração de banco

## Status

Aceito

## Contexto

O backend usa PostgreSQL real, migrations SQL e SQLC. Testes apenas com mocks não cobrem regressões de schema, constraints, enums, queries geradas e comportamento multi-tenant no banco.

## Decisão

Usar Testcontainers para subir PostgreSQL isolado nos testes de integração, aplicar migrations e executar cenários contra banco real.

## Consequências

- Testes de integração validam migrations, SQLC e constraints de banco com maior fidelidade.
- A suíte fica mais lenta do que testes unitários puros.
- O ambiente de CI e desenvolvimento precisa ter Docker disponível.
- Mocks continuam úteis para testes unitários rápidos e casos de erro específicos.
