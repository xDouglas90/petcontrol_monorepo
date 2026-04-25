# Plano de Ação e Execução - Tela de Serviços do Tenant

## Objetivo

Definir o escopo técnico, funcional e arquitetural da próxima PR do PetControl para evoluir a tela `Serviços` em `/:companySlug/services`, aproximando sua experiência das telas `Pessoas` e `Pets`, com painel lateral contextual, criação/edição completa e modelagem relacional expandida.

Esta PR deve introduzir:

- evolução da tela `/:companySlug/services` para o padrão `lista + aside direito`;
- visualização detalhada de serviço com sub-serviços e tempos médios;
- criação e edição no aside direito, com formulário estruturado;
- persistência transacional populando as tabelas:
  - `service_types`
  - `services`
  - `sub_services`
  - `services_average_times`
  - `company_services`
- regras de autorização para que apenas `admin` tenha todas as mutações do módulo;
- ajuste de permissões padrão para usuários `system`:
  - manter somente `services:view` e `users:view` como permissões padrão derivadas de `default_roles`.

## Contexto Atual

- A rota `/:companySlug/services` já existe no Web em [apps/web/src/routes/(app)/services/index.tsx](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/apps/web/src/routes/(app)/services/index.tsx:1), porém ainda em layout antigo e sem aside contextual.
- O backend já expõe CRUD básico de serviços em [apps/api/internal/handler/service_handler.go](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/apps/api/internal/handler/service_handler.go:1), com criação hoje focada em `service_types`, `services` e `company_services`.
- O schema já contempla as entidades relacionais completas em [schema.sql](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/schema.sql:927), incluindo:
  - `sub_services`;
  - `services_average_times`.
- As permissões atuais ainda permitem mutações de `services` para `system` no seed e na convenção.

## Referências Obrigatórias

- Plano de pessoas: [0009-PEOPLE-SCREEN-AND-TENANT-PERSON-REGISTRY-PLAN.md](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/docs/plans/0009-PEOPLE-SCREEN-AND-TENANT-PERSON-REGISTRY-PLAN.md:1)
- Plano de pets: [0010-PETS-SCREEN-AND-TENANT-PET-REGISTRY-PLAN.md](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/docs/plans/0010-PETS-SCREEN-AND-TENANT-PET-REGISTRY-PLAN.md:1)
- Tela atual de serviços: [apps/web/src/routes/(app)/services/index.tsx](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/apps/web/src/routes/(app)/services/index.tsx:1)
- Handler de serviços: [apps/api/internal/handler/service_handler.go](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/apps/api/internal/handler/service_handler.go:1)
- Service de domínio: [apps/api/internal/service/service_service.go](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/apps/api/internal/service/service_service.go:1)
- Queries SQLC de serviços: [infra/sql/queries/services.sql](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/infra/sql/queries/services.sql:1)
- Queries SQLC de sub-serviços: [infra/sql/queries/sub_services.sql](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/infra/sql/queries/sub_services.sql:1)
- Convenção de permissões: [docs/conventions/permissions.md](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/docs/conventions/permissions.md:1)
- Seed de permissões e módulos: [infra/scripts/seed.sh](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/infra/scripts/seed.sh:1)

## Resultado Esperado da PR

Ao final desta PR, o sistema deve apresentar:

- uma tela `Serviços` no padrão visual e estrutural de `lista + aside`, alinhada com `/people` e `/pets`;
- listagem tenant-scoped com filtros operacionais;
- detalhe de serviço exibindo:
  - dados principais;
  - sub-serviços;
  - tempos médios por perfil de pet;
- criação transacional de serviço populando todas as tabelas-alvo do domínio;
- edição coerente dos dados principais, sub-serviços e tempos médios;
- política de acesso em que somente `admin` consegue criar/editar/remover/desativar serviços;
- usuários `system` com permissões padrão restritas a:
  - `users:view`
  - `services:view`.

## Status da Implementação

Iniciada em 24/04/2026.

- Implementado: criação/edição transacional agregada em `POST/PUT /services`, incluindo `service_types`, `services`, `sub_services`, `services_average_times` e `company_services`.
- Implementado: `GET /services/:id` com retorno agregado e listagem com contadores de sub-serviços e tempos médios.
- Implementado: proteção da API pelo módulo `SVC` e permissões `services:view/create/update/delete`.
- Implementado: seed e convenção de permissões com `system` restrito a `users:view` e `services:view`.
- Implementado: contrato compartilhado e tela `/services` em layout `lista + aside`, com formulário agregado para múltiplos sub-serviços e múltiplos tempos médios.
- Implementado: teste de rollback da criação agregada quando a inserção de tempo médio falha.
- Implementado: teste de autorização bloqueando criação por usuário `system` com permissão apenas de visualização.
- Implementado: filtros estruturados na UI por tipo, status e faixa de preço.
- Implementado: detalhe agregado readonly selecionável no aside.
- Pendente: validação visual por permissão no frontend e persistência dos filtros estruturados na URL/API.

## Escopo Funcional Confirmado

## 1. Rota, Navegação e Shell

Esta PR não cria nova rota; evolui `/:companySlug/services`.

Também deve:

- manter `Serviços` no menu lateral;
- adicionar botão de `+` no menu lateral para abrir criação direta de serviço (mesmo padrão adotado em `/pets`);
- adotar composição de tela com:
  - lista principal à esquerda;
  - aside contextual à direita;
- manter responsividade desktop/mobile compatível com o padrão de `/people` e `/pets`.

## 2. Fonte de Verdade do Domínio

A criação de serviço deve ser transacional e preencher:

1. `service_types`:
   - resolver por nome existente ou criar novo tipo;
2. `services`:
   - registro principal do catálogo;
3. `sub_services`:
   - variações vinculadas ao serviço principal;
4. `services_average_times`:
   - tempos médios por `pet_size`, `pet_kind`, `pet_temperament`;
5. `company_services`:
   - vínculo do serviço com o tenant.

Decisão recomendada para simplificar consistência:

- exigir ao menos 1 `sub_service` por serviço na primeira versão;
- exigir ao menos 1 linha de `services_average_times` por `sub_service` ativo.

## 3. Campos e Estrutura da Listagem

Cada item da lista deve incluir, no mínimo:

- `title`;
- `type_name`;
- `description` resumida;
- `price` e `discount_rate`;
- `is_active`;
- quantidade de sub-serviços ativos;
- indicador resumido de tempos médios cadastrados.

Filtros recomendados:

- tipo de serviço (`service_types.name`);
- status (`is_active`);
- faixa de preço;
- busca textual (`title` e `description`).

## 4. Aside Direito do Módulo

Estados esperados:

- vazio;
- detalhe de serviço selecionado;
- criação;
- edição.

Conteúdo do detalhe:

- bloco de dados principais;
- bloco de sub-serviços;
- bloco de tempos médios por perfil de pet;
- ações de editar e desativar (quando permitido).

Conteúdo de criação/edição:

- dados principais do serviço;
- editor de sub-serviços (lista dinâmica);
- editor de tempos médios por sub-serviço (lista dinâmica).

## 5. Contratos Backend e Frontend

O contrato de `services` deve evoluir para suportar payload agregado:

- `service` (principal);
- `sub_services[]`;
- `average_times[]` ou `average_times_by_sub_service[]`.

Endpoints recomendados:

- `GET /services` com filtros estruturados;
- `GET /services/:id` retornando agregado (principal + sub-serviços + tempos);
- `POST /services` com criação transacional completa;
- `PUT /services/:id` com atualização agregada;
- `DELETE /services/:id` ou desativação lógica tenant-scoped.

## 6. Regras de Permissões e Acesso

### 6.1 Serviços

- `admin`:
  - `services:create`
  - `services:view`
  - `services:update`
  - `services:delete`
  - `services:deactivate`
  - `services:reactivate`
- `system`:
  - somente `services:view`.

### 6.2 Ajuste adicional solicitado para permissões padrão de `system`

Ao final da PR, o bootstrap de permissões por `default_roles` deve garantir que usuários `system` recebam como padrão apenas:

- `users:view`
- `services:view`

Implicações:

- remover `system` de `default_roles` em todas as demais permissões;
- adicionar `system` em `users:view` (se ainda não estiver);
- atualizar seed e convenções de permissões;
- garantir que o comportamento de bootstrap/reseed reflita a nova regra.

## 7. Igualdade de Layout com `/people` e `/pets`

Para maximizar consistência visual e de UX:

- manter mesma arquitetura de composição de tela;
- usar padrões de interação equivalentes:
  - seleção de item na lista abre detalhe no aside;
  - criação e edição dentro do aside;
  - filtros persistentes por URL/search params;
- manter o padrão de botões/ações e estados (`loading`, `empty`, `error`, `success`) já aplicado nos módulos anteriores.

## 8. Testes Esperados

Backend:

- criação transacional populando todas as tabelas alvo;
- rollback quando etapa intermediária falhar;
- listagem tenant-scoped com filtros;
- detalhe agregado tenant-scoped;
- validação de autorização:
  - `admin` com mutações permitidas;
  - `system` bloqueado para mutações;
- testes de permissão padrão para `system` (`users:view` + `services:view` apenas).

Frontend:

- render da rota `/services` no padrão `lista + aside`;
- abertura de criação pelo botão `+` da navegação;
- listagem com filtros persistentes;
- fluxo de detalhe, criação e edição agregada;
- validação de campos mínimos de `sub_services` e tempos médios;
- comportamento readonly para usuário sem permissão de mutação.

## 9. Riscos e Pontos de Atenção

- modelagem de payload agregado pode aumentar complexidade de validação no handler;
- risco de inconsistência entre `sub_services` e `services_average_times` sem validação transacional clara;
- mudança de `default_roles` para `system` pode impactar fluxos existentes que assumiam permissões mais amplas.

## 10. Fora de Escopo

- gestão de planos de serviços (`service_plans`) e bônus nesta PR;
- geração de relatórios avançados derivados de serviços;
- anexos multimídia além de imagem principal;
- automação inteligente de tempo médio com machine learning.

## 11. Checklist de Implementação

- [x] Definir DTO agregado de criação/edição de serviços.
- [x] Evoluir queries SQLC para leitura agregada de serviço + sub-serviços + tempos.
- [x] Implementar criação transacional preenchendo:
  - [x] `service_types`
  - [x] `services`
  - [x] `sub_services`
  - [x] `services_average_times`
  - [x] `company_services`
- [x] Implementar atualização agregada transacional.
- [x] Garantir rollback completo em caso de falha parcial.
- [x] Evoluir `ServiceHandler` para novo contrato.
- [ ] Atualizar Swagger/docs do módulo de serviços.
- [x] Reestruturar `/services` para layout `lista + aside`.
- [x] Implementar filtros estruturados na UI.
- [ ] Persistir filtros estruturados na URL/API.
- [x] Implementar detalhe agregado readonly no aside.
- [x] Implementar criação/edição agregada no aside.
- [x] Adicionar botão `+` no menu lateral para criação direta.
- [x] Ajustar autorização de mutações para `admin` apenas.
- [x] Garantir `system` com `services:view` apenas no módulo.
- [x] Atualizar `default_roles` para que `system` receba somente:
  - [x] `users:view`
  - [x] `services:view`
- [x] Atualizar `docs/conventions/permissions.md`.
- [x] Atualizar `infra/scripts/seed.sh`.
- [x] Adicionar/ajustar testes backend.
- [x] Adicionar/ajustar testes frontend.

## 12. Sequência Recomendada

1. Fechar contrato de permissões (`admin` mutação total, `system` apenas visualização).
2. Definir e implementar payload agregado no backend.
3. Fechar criação/edição transacional com todas as tabelas do domínio.
4. Reestruturar UI `/services` para padrão de `/people` e `/pets`.
5. Integrar filtros, detalhe e formulários dinâmicos.
6. Cobrir com testes automatizados backend/frontend.
7. Atualizar convenções e seed para novo baseline de permissões de `system`.

## Conclusão

Este plano 0011 consolida o módulo `Serviços` como catálogo operacional completo do tenant, alinhando modelagem relacional, UX contextual e política de acesso. A principal decisão estrutural é tratar a criação/edição como fluxo agregado transacional e, no mesmo ciclo, reduzir o baseline de permissões padrão de `system` para somente `users:view` e `services:view`, garantindo coerência entre segurança, produto e manutenção futura.
