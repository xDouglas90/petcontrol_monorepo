# Plano de Ação e Execução - Expansão do Núcleo Operacional do Atendimento

## Objetivo

Definir a próxima fase de desenvolvimento do PetControl com base em:

- estado atual documentado no `README.md`;
- histórico de execução já registrado em `docs/plans/0001-INITIAL_STRUCTURE_ACTION_PLAN.md`;
- consolidação funcional registrada em `docs/plans/0002-APPLICATION_NEXT_STEPS_PLAN.md`;
- evolução de navegação e UX do tenant registrada em `docs/plans/0003-COMPANY_SLUG_ROUTING_PLAN.md`;
- código atualmente implementado no monorepo.

O foco desta fase deixa de ser "endurecer a base" e passa a ser "expandir o primeiro núcleo operacional real do produto" para além de `schedules`, aproveitando tudo o que já está sólido em autenticação, tenant, auditoria, Web, Worker, Swagger, testes e CI.

## Leitura Consolidada do Estado Atual

### O que já está bem resolvido

- Monorepo consolidado com `apps/api`, `apps/web`, `apps/worker`, `libs`, `infra` e `docs`.
- API Go com autenticação JWT, tenant por `company_id`, validação de módulo, auditoria, correlation id e Swagger.
- Fluxo real ponta a ponta para `companies/current`, plano ativo, módulos ativos, `company_users` e `schedules`.
- Worker processando evento real de negócio ligado à confirmação de `schedule`.
- Web autenticado com `company_slug` nas rotas, dashboard conectado e CRUD funcional de `schedules`.
- Testes relevantes em backend, frontend, worker, queries SQLC e CI já endurecida.

### O principal gargalo funcional observado

Apesar da base estar madura, o produto ainda opera com apenas um módulo de negócio realmente exposto de ponta a ponta: `schedules`.

Esse módulo, por sua vez, ainda depende de entidades que não foram promovidas a módulos vivos da aplicação:

- o formulário Web de `schedules` ainda pede `client_id` e `pet_id` manualmente;
- não há rotas Web reais para `clients`, `pets`, `services` ou `reports`;
- a API ainda não possui handlers/services/queries dedicados para `clients`, `pets` e `services`;
- o seed local ainda não cria massa funcional desses domínios;
- `shared-types` ainda concentra auth, company e `schedules`, sem refletir o próximo bloco real do domínio.

Em outras palavras: a plataforma já está pronta para crescer, mas o valor de produto ainda está travado pela ausência do núcleo cadastral e operacional que alimenta os agendamentos.

## Conclusão: qual deve ser a próxima fase

A próxima fase recomendada é a **expansão do núcleo operacional do atendimento**, com prioridade em:

1. `clients`
2. `pets`
3. `services`
4. enriquecimento do fluxo de `schedules` para consumir essas entidades de forma real

Essa direção é a mais coerente porque:

- está explicitamente alinhada ao `README.md`, que aponta a expansão de domínios além de `schedules` como o próximo ciclo natural;
- aproveita o módulo `CRM` já seedado no ambiente local;
- remove a maior fricção atual do Web, que ainda usa UUIDs crus no fluxo principal;
- permite evoluir `schedules` sem reinventar a base técnica já pronta;
- cria fundação concreta para relatórios, notificações mais ricas, histórico de relacionamento e futuro financeiro.

## Princípios de Execução da Nova Fase

- Manter entrega vertical por domínio: banco, SQLC, service, handler, Web, testes e documentação no mesmo ciclo.
- Reaproveitar padrões já consolidados em `schedules`, `company_users` e `companies/current`.
- Preservar isolamento por tenant em toda query de leitura e mutação.
- Não expandir para muitos domínios paralelos ao mesmo tempo; fechar bem `clients`, `pets` e `services` antes de abrir `reports` ou financeiro.
- Usar esta fase também para melhorar a experiência do módulo `schedules`, em vez de tratá-lo como domínio isolado.

## Fase 17 - Contratos de Domínio e Massa de Desenvolvimento

### 17.1 - Ações

- Expandir `libs/shared-types` para incluir DTOs, enums e tipos de:
  - `clients`;
  - `pets`;
  - `services`;
  - respostas enriquecidas de `schedules`, quando passarem a exibir nomes e relacionamentos.
- Revisar `libs/shared-constants` para incluir segmentos de rota e eventuais códigos de módulo usados pelos novos fluxos.
- Atualizar o seed local para criar uma massa mínima utilizável de:
  - ao menos 1 cliente ativo;
  - ao menos 1 pet vinculado;
  - ao menos 1 serviço ativo da empresa;
  - ao menos 1 `schedule` de exemplo usando esse ecossistema quando fizer sentido.
- Garantir que `README.md` e contratos compartilhados não descrevam entidades ainda inexistentes como se já fossem rotas ativas.

### 17.2 - Checks

- [ ] `shared-types` cobre os contratos reais de `clients`, `pets` e `services`.
- [ ] O seed local permite usar o Web sem precisar descobrir UUIDs manualmente no banco.
- [ ] Os novos tipos compartilhados não duplicam contratos já existentes no backend.

## Fase 18 - Módulo Base de `clients`

### 18.1 - Ações

- Criar queries SQLC para `clients` e `company_clients`, cobrindo:
  - listagem por tenant;
  - criação;
  - consulta por id;
  - atualização;
  - soft delete ou desativação controlada, conforme a modelagem atual.
- Implementar `ClientService` e `ClientHandler` seguindo o padrão existente da API.
- Expor endpoints protegidos em `/api/v1/clients`.
- Aplicar auditoria nas mutações relevantes.
- Adicionar cobertura unitária e de integração com foco em isolamento multi-tenant.

### 18.2 - Checks

- [ ] `GET /api/v1/clients` retorna apenas clientes do tenant autenticado.
- [ ] `POST /api/v1/clients` cria cliente e vínculo com a empresa sem depender de `company_id` no body.
- [ ] `GET`, `PUT` e `DELETE` de cliente respeitam tenant e soft delete.
- [ ] Testes cobrem criação, listagem, atualização e bloqueio de acesso cruzado entre tenants.

## Fase 19 - Módulo Base de `pets`

### 19.1 - Ações

- Criar queries SQLC para `pets` com vínculo explícito ao `owner_id` e consistência com o cliente do tenant.
- Implementar `PetService` e `PetHandler`.
- Expor endpoints protegidos em `/api/v1/pets`.
- Validar regras mínimas de negócio:
  - pet precisa pertencer a um cliente visível para o tenant;
  - pet deletado/inativo não deve aparecer como opção de agendamento;
  - dados básicos do pet devem ser retornados já prontos para o Web.
- Adicionar auditoria e testes de integração.

### 19.2 - Checks

- [ ] `GET /api/v1/pets` lista apenas pets associados a clientes do tenant.
- [ ] `POST /api/v1/pets` falha ao tentar vincular pet a cliente de outro tenant.
- [ ] `PUT` e `DELETE` de pet preservam o histórico e o isolamento multi-tenant.
- [ ] O backend consegue validar ownership de pet sem depender de suposições frágeis no frontend.

## Fase 20 - Catálogo de `services` e Enriquecimento de `schedules`

### 20.1 - Ações

- Implementar o módulo base de `services` e `company_services`, com endpoints protegidos.
- Evoluir o fluxo de `schedules` para consumir relacionamentos reais do catálogo, em vez de operar apenas com ids crus.
- Revisar se esta fase deve:
  - introduzir `schedule_services` já neste ciclo; ou
  - ao menos enriquecer as respostas de `schedules` com nomes de cliente, pet e serviço.
- Atualizar queries de listagem de `schedules` para retornarem payload mais útil ao Web.
- Ajustar publicação de eventos para o Worker quando o payload puder carregar contexto mais rico de atendimento.

### 20.2 - Checks

- [ ] Existe `GET /api/v1/services` funcional por tenant.
- [ ] A tela de `schedules` deixa de exigir digitação manual de UUID para os fluxos principais.
- [ ] O payload retornado por `schedules` passa a exibir contexto humano suficiente para uso operacional.
- [ ] Eventos do Worker continuam compatíveis e versionados após o enriquecimento do domínio.

## Fase 21 - Web Operacional para `clients`, `pets` e `services`

### 21.1 - Ações

- Criar rotas autenticadas no Web para:
  - `/:companySlug/clients`;
  - `/:companySlug/pets`;
  - `/:companySlug/services`.
- Implementar camada de queries/mutations dedicada em `apps/web/src/lib/api`.
- Criar telas com listagem, formulário, estados de loading/erro/empty e ações básicas.
- Atualizar `schedules` para usar seletores reais de cliente, pet e serviço.
- Revisar navegação da sidebar e do header para acomodar os novos módulos de forma consistente.

### 21.2 - Checks

- [ ] O Web passa a ter fluxos reais de cadastro e consulta para `clients`, `pets` e `services`.
- [ ] O módulo de `schedules` usa entidades reais do domínio em vez de ids digitados manualmente.
- [ ] Query cache e invalidação seguem o padrão já usado hoje.
- [ ] Rotas com `company_slug` continuam corretas e cobertas por testes.

## Fase 22 - Fechamento de Produto, Qualidade e Direção Seguinte

### 22.1 - Ações

- Expandir Swagger para documentar os novos endpoints reais.
- Revisar `README.md` e `docs/CONTRIBUTING.md` com os novos módulos ativos.
- Adicionar testes focados em fluxos cruzados:
  - cliente -> pet -> agendamento;
  - exclusão/desativação e impacto em `schedules`;
  - uso do módulo `CRM` e `SCH` em conjunto.
- Decidir, ao fim desta fase, qual será o próximo vertical:
  - `reports` operacionais;
  - notificações mais ricas;
  - financeiro;
  - administração avançada da empresa.

### 22.2 - Checks

- [ ] Swagger cobre os módulos reais adicionados nesta fase.
- [ ] O onboarding local continua simples após a expansão do domínio.
- [ ] Existe um fluxo funcional completo de atendimento do cadastro ao agendamento.
- [ ] O projeto encerra a fase com clareza objetiva sobre o próximo vertical do produto.

## Ordem Recomendada de Execução

1. Fase 17: contratos e seed utilizável.
2. Fase 18: `clients`.
3. Fase 19: `pets`.
4. Fase 20: `services` e enriquecimento de `schedules`.
5. Fase 21: Web operacional dos novos domínios.
6. Fase 22: documentação, testes de fluxo e definição do próximo vertical.

## Resultado Esperado

Se esta fase for executada com sucesso, o PetControl deixará de ser uma plataforma com base técnica sólida e um único módulo funcional isolado, passando a operar com um núcleo real de atendimento:

- cadastro de clientes;
- cadastro de pets;
- catálogo de serviços;
- agendamento usando entidades reais e legíveis;
- terreno pronto para relatórios, notificações mais úteis e evolução comercial do produto.

## Resumo Executivo da Recomendação

O próximo passo não deve ser abrir um módulo totalmente novo e distante do fluxo atual. O maior retorno agora está em transformar `schedules` em um módulo realmente utilizável no contexto do negócio.

Por isso, a melhor `NOVA FASE` é: **fechar o núcleo operacional `clients` + `pets` + `services`, e usar essa expansão para amadurecer `schedules` de ponta a ponta**.
