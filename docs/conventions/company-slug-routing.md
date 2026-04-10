# Convencao de Rotas com `company_slug`

## Contexto

Esta documentação registra a convenção planejada para a area autenticada do Web, conforme o plano [0003-COMPANY_SLUG_ROUTING_PLAN.md](../plans/0003-COMPANY_SLUG_ROUTING_PLAN.md).

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

Após a Fase 5 do plano `0003`, a área autenticada do Web passou a ser roteada sob `/:companySlug`, com links internos preservando esse prefixo de tenant, normalização canônica em lowercase e confirmação visual do tenant atual no header.

## Canonicalização da URL

- O slug usado para navegação deve ser tratado como canônico em lowercase.
- URLs acessadas com diferença apenas de caixa devem ser redirecionadas para a forma canônica.
- A montagem de rotas autenticadas deve passar sempre por helpers compartilhados para evitar variações manuais entre apps.

## Confirmação visual de tenant

- O header da área autenticada deve exibir o nome da empresa resolvida e o `company_slug` atual.
- Essa confirmação é de UX e orientação operacional.
- Ela não substitui a validação de tenant pelo backend.

## Direção futura para mudança de slug

- Se a empresa puder alterar o slug no futuro, a fonte de verdade continua sendo a empresa corrente resolvida pela sessão.
- O frontend deve corrigir a URL ativa quando detectar divergência entre slug da rota e slug real devolvido pelo backend.
- Links compartilhados antigos podem exigir estratégia complementar de compatibilidade no backend, mas isso é uma evolução posterior e não muda a regra de autorização baseada em JWT e `company_id`.

## Relação com outros tipos de documento

- `docs/plans/`: descrevem execução por fases, checks e ordem de entrega.
- `docs/adr/`: registram decisões arquiteturais mais amplas e duradouras.
- `docs/conventions/`: documentam regras operacionais e convenções compartilhadas entre apps e libs.
