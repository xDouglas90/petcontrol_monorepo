# ADR 0004: Multi-tenancy por company_id

## Status

Aceito

## Contexto

A plataforma atende múltiplas empresas e precisa garantir isolamento lógico entre tenants sem complexidade de banco por cliente nesta fase.

## Decisão

Usar modelo de multi-tenancy por company_id em entidades de domínio, com enforcement em middleware e queries tenant-aware.

## Consequências

- Isolamento consistente com menor custo operacional inicial.
- Necessidade de disciplina em filtros e validações de acesso.
- Facilita evolução futura para estratégias híbridas se necessário.
