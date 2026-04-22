# Plano de Refactor - Eliminar Catálogo Hardcoded de Permissões em Tenant Settings

## Objetivo

Remover a dependência do catálogo hardcoded em [tenant_settings_permissions_catalog.go](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/apps/api/internal/service/tenant_settings_permissions_catalog.go:1) e migrar a geração das permissões de `/settings` para uma fonte orientada ao banco, usando as tabelas já existentes:

- `modules`
- `permissions`
- `module_permissions`
- `plans`
- `plan_modules`
- `plan_permissions`
- `companies`

O objetivo final é fazer com que a seção `Permissões` da tela `/settings` seja derivada de dados persistidos e não de listas manuais em código.

## Problema Atual

Hoje o backend de `/settings` usa um catálogo hardcoded para:

- definir quais módulos são exibíveis;
- definir o pacote mínimo de cada módulo;
- definir o conjunto de permissões gerenciáveis;
- excluir módulos `internal` do tenant;
- montar os grupos retornados ao Web.

Isso resolve o problema funcional de curto prazo, mas cria alguns riscos:

- duplicação de verdade entre código, `permissions.md` e `modules.md`;
- alto custo de manutenção ao adicionar módulos/permissões;
- risco de divergência entre banco e runtime;
- necessidade de editar código para refletir mudanças que já deveriam viver em dados.

## Estado Atual do Projeto

O projeto já possui base de schema para essa refatoração:

- `modules` possui `code`, `name`, `description`, `min_package` e `is_active`;
- `module_permissions` liga permissões a módulos;
- `plan_modules` permite saber os módulos ligados a cada plano;
- `companies.active_package` já indica o pacote atual do tenant;
- já existe query/serviço para listar módulos ativos por empresa;
- já existe query para listar permissões por módulo.

Por outro lado, ainda há lacunas prováveis:

- o seed de módulos parece não estar completo para todo o catálogo de `modules.md`;
- o vínculo `module_permissions` pode não estar completo para todas as permissões de `permissions.md`;
- a regra de exclusão de módulos `internal` ainda não está modelada como política explícita no banco;
- a API de permissões agrupadas ainda depende do catálogo em código para o recorte funcional.

Status atual da auditoria:

- o schema já suporta uma abordagem DB-driven com `modules`, `module_permissions`, `plans`, `plan_modules`, `companies.active_package` e `company_modules`;
- a disponibilidade concreta por tenant hoje já pode ser derivada de `company_modules`, e não apenas de `active_package`;
- o seed atual de `modules` ainda é insuficiente para o catálogo completo de [modules.md](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/docs/conventions/modules.md:1):
  - hoje só entram `SCH`, `CRM` e `FIN`;
- o seed atual de `plan_modules` e `company_modules` para o tenant de desenvolvimento também é parcial:
  - hoje só entram `SCH` e `CRM`;
- o seed atual de `permissions` é amplo, mas ainda não há hidratação correspondente de `module_permissions` em [seed.sh](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/infra/scripts/seed.sh:1);
- por isso, o banco já tem estrutura suficiente, mas ainda não tem dados seedados suficientes para substituir com segurança o catálogo hardcoded.

## Resultado Esperado

Ao final deste refactor:

- o backend não dependerá mais de listas hardcoded para montar `permission_groups`;
- o agrupamento por módulo virá de `modules` + `module_permissions`;
- o filtro de módulos visíveis virá do plano/pacote do tenant;
- o backend seguirá excluindo módulos não exibíveis ao tenant, mas com regra baseada em dados e política centralizada;
- adicionar um novo módulo/permissão exigirá ajuste de dados/seed, e não alteração manual no catálogo de Go.

## Decisão Arquitetural Recomendada

### Fonte de Verdade

Usar o banco como fonte de verdade para:

- catálogo de módulos;
- pacote mínimo por módulo;
- nome e descrição do módulo;
- relacionamento entre módulo e permissão.

### Política de Elegibilidade

Separar duas camadas:

1. camada de dados
   - o banco informa quais módulos existem e quais permissões pertencem a cada módulo;
2. camada de política
   - o backend decide o que é exibível para o tenant, por exemplo:
     - módulos `internal` não aparecem;
     - módulos sem permissões gerenciáveis não aparecem;
     - módulos fora do pacote/plano do tenant não aparecem.

Essa política pode continuar em Go, mas sem listar módulo por módulo manualmente.

## Estratégia Recomendada de Refactor

### Fase 0 - Auditoria dos Dados

Objetivo:

- validar se o banco já consegue sustentar o contrato atual sem o catálogo hardcoded.

Status atual:

- auditoria inicial concluída no schema, queries e seed;
- concluído que a remoção do catálogo ainda não é segura nesta etapa, porque faltam dados persistidos para sustentar todo o agrupamento por módulo.

### 0.1 Ações

- auditar os registros de `modules` contra [modules.md](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/docs/conventions/modules.md:1);
- auditar os registros de `permissions` contra [permissions.md](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/docs/conventions/permissions.md:1);
- auditar `module_permissions` para garantir que cada permissão exibível do tenant esteja ligada ao módulo certo;
- auditar `plan_modules` e/ou `companies.active_package` para definir a regra final de disponibilidade;
- identificar módulos que devem ser sempre excluídos do tenant:
  - `internal`
  - operação central
  - módulos `root-only`, se existirem nesse recorte.

### 0.2 Checks

- [ ] Todo módulo exibível do tenant existe na tabela `modules`.
- [ ] Toda permissão exibível do tenant possui vínculo em `module_permissions`.
- [x] Está decidido se a disponibilidade virá de `active_package`, `plan_modules` ou composição dos dois.

Conclusão da Fase 0:

- a estratégia recomendada segue sendo a composição de:
  - `company_modules` como habilitação efetiva do tenant;
  - `active_package` como teto/política complementar;
  - `plan_modules` e `plan_permissions` como base de catálogo contratual;
- antes da Fase 2, é necessário concluir a Fase 1 e preencher corretamente `modules` e `module_permissions`.

## Fase 1 - Completar Seeds e Relacionamentos

Objetivo:

- garantir que o banco tenha dados suficientes para substituir o catálogo.

Status atual:

- o seed de `modules` foi ampliado para persistir o catálogo necessário ao `/settings`, alinhado a [modules.md](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/docs/conventions/modules.md:1);
- o seed de `module_permissions` passou a hidratar os vínculos dos módulos gerenciáveis do tenant;
- o tenant seedado `petcontrol-dev` e o plano `Starter Monthly` agora recebem também os módulos `CFG`, `UCR`, `CLI` e `SVC`, além dos códigos legados ainda usados pelo app atual;
- foi mantido temporariamente o módulo legado `CRM` no seed para não quebrar o recorte operacional já em uso por rotas e middlewares existentes;
- o teste de integração do seed agora valida módulos, vínculos em `module_permissions` e módulos ativos do tenant starter.

### 1.1 Ações

- completar o seed de `modules` para refletir `modules.md`;
- completar o seed de `module_permissions` para refletir `permissions.md`;
- revisar seeds de `plan_modules` e `plan_permissions`;
- garantir que módulos por pacote estejam coerentes com `Pacote Mínimo`;
- adicionar testes de seed/integridade para evitar regressões.

### 1.2 Checks

- [x] O seed persiste todos os módulos necessários para `/settings`.
- [x] O seed persiste todos os vínculos `module_permissions` necessários.
- [x] Há teste garantindo integridade mínima do catálogo persistido.

## Fase 2 - Consultas DB-Driven para Tenant Settings

Objetivo:

- criar queries específicas para montar a visão que a tela precisa.

Status atual:

- foram adicionadas queries específicas para listar módulos elegíveis e permissões agrupáveis do tenant com base em `company_modules`, `modules` e `module_permissions`;
- `ListTenantSettingsModulesByCompanyID` já exclui módulos `internal` e ignora módulos sem permissões vinculadas;
- `ListTenantSettingsPermissionsByCompanyID` já devolve o conjunto agrupável com metadados do módulo em cada linha;
- `CompanyUserPermissionService` passou a consumir essas queries para:
  - montar `permission_groups`;
  - validar o conjunto de permissões editáveis no update;
- o catálogo hardcoded continua existindo como fallback temporário no projeto, mas a trilha principal da tela `/settings` já usa dados do banco nesta etapa.

### 2.1 Ações

- criar query para listar módulos exibíveis do tenant com base na empresa/plano;
- criar query para listar permissões agrupáveis por módulo para o tenant;
- preferir uma query já filtrada por empresa/pacote, para reduzir lógica espalhada no service;
- definir retorno contendo:
  - módulo;
  - descrição;
  - pacote mínimo;
  - permissões vinculadas.

### 2.2 Checks

- [x] Existe query que lista módulos elegíveis para `/settings`.
- [x] Existe query que lista permissões desses módulos sem depender de catálogo fixo em Go.

## Fase 3 - Refactor do Service

Objetivo:

- trocar o uso de `tenant_settings_permissions_catalog.go` por dados do banco.

Status atual:

- `CompanyUserPermissionService` já deixou de montar `permission_groups` a partir do catálogo hardcoded;
- o service agora busca:
  - módulos elegíveis via `ListTenantSettingsModulesByCompanyID`;
  - permissões agrupáveis via `ListTenantSettingsPermissionsByCompanyID`;
- a validação de `permission_codes` no update também passou a usar o catálogo efetivo do tenant vindo do banco;
- o contrato público de `permission_groups` foi preservado para o frontend;
- o arquivo hardcoded ainda permanece no repositório apenas como implementação temporária residual, deixando sua remoção física para a Fase 4.

### 3.1 Ações

- refatorar `CompanyUserPermissionService` para:
  - buscar módulos elegíveis no banco;
  - buscar permissões por módulo no banco;
  - montar `permission_groups` dinamicamente;
- manter a mesma resposta pública, para não quebrar o Web;
- mover a política de exclusão de módulos não-tenant para uma função enxuta, sem catálogo manual de permissões.

### 3.2 Checks

- [x] O service não depende mais de lista manual de permissões por módulo.
- [x] O contrato de `permission_groups` permanece estável para o frontend.

## Fase 4 - Remoção do Catálogo Hardcoded

Objetivo:

- apagar a implementação temporária com segurança.

Status atual:

- `tenant_settings_permissions_catalog.go` foi removido do código da API;
- a suíte de testes legada específica desse catálogo também foi removida;
- a cobertura equivalente foi preservada por:
  - testes de seed/integridade;
  - testes de service DB-driven;
  - testes de handler do endpoint agrupado;
- a remoção foi validada sem referências residuais ao catálogo antigo no `apps/api`.

### 4.1 Ações

- remover `tenant_settings_permissions_catalog.go`;
- remover testes ligados apenas ao catálogo hardcoded;
- substituir esses testes por:
  - testes de queries;
  - testes de seed;
  - testes de service baseado em banco.

### 4.2 Checks

- [x] O arquivo hardcoded foi removido.
- [x] A cobertura equivalente foi preservada com testes baseados em dados.

## Fase 5 - Robustez e Observabilidade

Objetivo:

- garantir que mudanças de catálogo em dados não quebrem a UI silenciosamente.

Status atual:

- as queries DB-driven passaram a filtrar também por `companies.active_package`, além de excluir módulos `internal`;
- foi adicionada cobertura de integração para garantir que:
  - módulo sem permissões não aparece;
  - módulo `internal` não aparece;
  - tenant `starter` não recebe módulo `premium` mesmo se houver vínculo ativo em `company_modules`;
  - permissão sem vínculo em `module_permissions` não entra no payload;
- o smoke test do endpoint agrupado foi reforçado para validar campos contratuais do grupo:
  - `module_code`
  - `module_name`
  - `module_description`
  - `min_package`

### 5.1 Ações

- adicionar testes que verifiquem:
  - módulo sem permissões não aparece;
  - módulo `internal` não aparece;
  - permissão sem vínculo não entra no payload;
  - pacote/plano inferior não recebe módulo premium;
- adicionar smoke test de contrato para `GET /company-users/:user_id/permissions`.

### 5.2 Checks

- [x] Há testes cobrindo exclusão de módulos indevidos.
- [x] Há testes cobrindo filtro por pacote/plano.
- [x] Há teste de contrato do endpoint agrupado.

## Pergunta Arquitetural a Fechar Antes do Refactor

O filtro de disponibilidade do tenant deve vir prioritariamente de:

1. `companies.active_package`
2. `plan_modules` do plano atual contratado
3. combinação dos dois, onde:
   - `active_package` define o teto do tenant;
   - `plan_modules` define habilitações concretas

Recomendação:

- usar a opção `3`, porque é a mais robusta para evolução futura.
- enquanto isso não estiver maduro, usar `active_package` como fallback aceitável.

## Riscos e Cuidados

- seeds incompletos podem fazer módulos “sumirem” da UI;
- vínculos errados em `module_permissions` podem agrupar permissões no módulo errado;
- depender só de `active_package` pode mascarar diferenças reais entre planos;
- depender só de `plan_modules` sem higiene de seed pode deixar tenants sem módulos esperados;
- remover o catálogo antes da base de dados estar íntegra pode quebrar `/settings`.

## Ordem Recomendada de Execução

1. Fase 0: auditoria dos dados.
2. Fase 1: completar seeds e relacionamentos.
3. Fase 2: consultas DB-driven.
4. Fase 3: refactor do service.
5. Fase 4: remoção do catálogo hardcoded.
6. Fase 5: robustez e observabilidade.

## Resultado Final Esperado

Se este refactor for executado com sucesso, a seção `Permissões` do tenant deixará de depender de um catálogo manual em Go e passará a ser montada a partir do próprio modelo de dados do sistema, reduzindo duplicação, melhorando manutenção e aproximando a UI da fonte real de verdade do domínio.
