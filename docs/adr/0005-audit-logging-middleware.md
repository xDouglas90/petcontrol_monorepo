# ADR 0005: Auditoria centralizada via middleware

## Status

Aceito

## Contexto

Mutações de domínio precisam registrar `audit_logs` com contexto de request, `company_id`, usuário autenticado, IP e user agent. Duplicar a persistência de auditoria em cada service aumentaria acoplamento e risco de inconsistência entre handlers.

## Decisão

Usar `middleware.AddAuditEntry(...)` nos handlers para acumular entradas de auditoria e persistir essas entradas ao final da request pelo middleware `Audit`.

## Consequências

- Services permanecem focados em regra de negócio e transações de domínio.
- Handlers declaram explicitamente o que mudou, incluindo `old_data` e `new_data`.
- O middleware padroniza enriquecimento e persistência com contexto HTTP.
- Fluxos sem request HTTP, como jobs do Worker, precisam de estratégia própria se passarem a exigir auditoria equivalente.
