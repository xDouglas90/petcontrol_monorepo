# ADR 0008: Guarda do domínio Pets pelo módulo PET

## Status

Aceito

## Contexto

Durante a expansão do núcleo operacional do PetControl, foi implementado o CRUD para as entidades `clients`, `pets` e `services`.
Observou-se um desvio histórico entre o catálogo de módulos atual e uma guarda legada aplicada no passado:

- O catálogo atual já possui um módulo dedicado `PET`.
- Havia proteção legada das rotas de `pets` por `CRM`, o que não refletia mais a modelagem modular vigente.

O conflito central passou a ser: manter compatibilidade legada (`CRM`) ou alinhar a autorização ao módulo de domínio dedicado (`PET`)?

## Decisão

O domínio `pets` passa a ficar sob a guarda primária do módulo `PET` na infraestrutura de roteamento do backend e na conceituação arquitetural.

### Justificativas

1. **Coerência de Catálogo**: O módulo `PET` existe como módulo de domínio e deve ser a referência de autorização das rotas de pets.
2. **Separação de Preocupações (SoC)**: `CLI` cobre o domínio de clientes/tutores e `PET` cobre o domínio de pets; ambos continuam relacionais, mas com fronteiras de autorização explícitas.
3. **Redução de Legado**: A remoção da guarda via `CRM` elimina ambiguidade entre seed, planos e middlewares.

## Consequências

- As rotas de `/pets` passam a requerer `RequireModule(..., "PET")`.
- O módulo `SCH` continua dependente dos dados de cliente/pet para operação, mas sem sobrepor os limites de autorização dos módulos `CLI` e `PET`.
- O seed e os testes de integração passam a usar módulos atuais, sem dependência de código legado `CRM`.
