# Plano de Ação e Execução - Permissões por Módulo em `/settings`

## Objetivo

Aprimorar a seção `Permissões` da tela `/settings` para que ela deixe de exibir apenas o subconjunto atual:

- `company_settings:edit`
- `plan_settings:edit`
- `payment_settings:edit`
- `notification_settings:edit`
- `integration_settings:edit`
- `security_settings:edit`

e passe a apresentar as permissões organizadas por módulos, seguindo a estrutura definida em [permissions.md](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/docs/conventions/permissions.md:1).

Além disso, as permissões visíveis para o `admin` devem respeitar os módulos disponíveis no plano atual do tenant, conforme o recorte de módulos descrito em [modules.md](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/docs/conventions/modules.md:1), com atenção especial ao campo `Pacote Mínimo`.

## Contexto Atual

- A tela `/settings` já possui uma seção `Permissões`.
- Hoje essa seção está limitada às permissões de configurações da empresa.
- O `admin` consegue gerir permissões de usuários do tenant, mas somente nesse subconjunto reduzido.
- O catálogo completo de permissões já está documentado por módulo em `permissions.md`.
- O catálogo de módulos e o pacote mínimo de cada um já está documentado em `modules.md`.

## Problema

O modelo atual de permissões em `/settings` não representa a real matriz de autorização do sistema, porque:

- não exibe permissões por módulo;
- não acompanha a estrutura oficial de `permissions.md`;
- não diferencia claramente o que está disponível conforme o plano do tenant;
- dificulta expansão futura da gestão de acesso;
- cria uma experiência parcial e inconsistente para o `admin`.

## Resultado Esperado

Ao final desta evolução, a seção `Permissões` deve:

- listar permissões agrupadas por módulo;
- usar os nomes e agrupamentos oficiais de `permissions.md`;
- exibir ao `admin` apenas os módulos liberados pelo plano atual do tenant;
- continuar permitindo edição por usuário vinculado ao tenant;
- preservar permissões básicas já seedadas e já persistidas em `user_permissions`;
- deixar módulos indisponíveis fora da UI de edição do tenant.

## Regras Funcionais Confirmadas

### 1. Escopo de Visualização

- A seção `Permissões` continua visível apenas para usuários do tipo `admin`.
- Usuários `system`, mesmo com acesso à tela `/settings`, não devem visualizar o bloco de gestão de permissões.

### 2. Organização por Módulo

- A UI deve agrupar permissões por módulo.
- Cada grupo deve refletir a convenção de `permissions.md`.
- O agrupamento deve usar ao menos:
  - nome do módulo;
  - descrição do módulo;
  - lista de permissões pertencentes àquele módulo.

### 3. Filtro por Plano do Tenant

- O `admin` só pode ver e editar permissões de módulos disponíveis no plano atual da empresa.
- A definição de disponibilidade deve seguir `modules.md`, usando `Pacote Mínimo` como base.
- Nesta fase, considerar o plano atual do tenant como a fonte de verdade para decidir quais módulos entram na seção.

### 4. Estado Base das Permissões

- Cada usuário deve continuar carregando suas permissões atuais a partir de `user_permissions`.
- Quando não houver customização explícita, a UI deve refletir o estado básico herdado do seed/papel padrão daquele usuário.
- A ausência de permissão visível na tela não deve significar revogação automática.

### 5. Persistência

- A edição continua usando os endpoints já criados para permissões por usuário do tenant.
- Será necessário expandir o payload e a listagem para cobrir o catálogo completo de módulos liberados.
- A API deve continuar impedindo edição fora do tenant autenticado.

## Recorte Inicial de Módulos por Plano

Com base no `Pacote Mínimo` atual descrito em `modules.md`, o plano deve contemplar ao menos este comportamento:

- `starter`
  - `Configs`
  - `Users`
  - `Schedules`
  - `Services`
  - `Clients`
- `basic`
  - tudo do `starter`
  - `Services Plans`
  - `Pets`
  - `Dashboard`
  - `Reports`
- `essential`
  - tudo do `basic`
  - `Products`
  - `Professionals Schedules`
  - `Delivery Pets`
  - `Inventory`
- `premium`
  - tudo do `essential`
  - `Custom Reports`
  - `Pet Day Care`
  - `Pet Hotel`
  - `Chat`
  - `Notifications`
  - `Financial`
  - `Suppliers`
  - `External User Access`

Observação:

- módulos marcados como `internal` em `modules.md` não devem aparecer nesta gestão do tenant em `/settings`, salvo decisão explícita posterior.

## Arquitetura Recomendada

## Backend

- Criar uma camada de mapeamento entre:
  - plano atual da empresa;
  - módulos disponíveis para aquele plano;
  - permissões pertencentes a cada módulo.
- Ajustar o endpoint de listagem de permissões por usuário para retornar:
  - módulos;
  - permissões agrupadas;
  - estado atual do usuário naquele módulo.
- Validar no backend que o `admin` não consegue atribuir permissões de módulos não incluídos no plano do tenant.

## Frontend

- Refatorar a seção `Permissões` de `/settings` para usar grupos por módulo.
- Exibir cada módulo em bloco próprio dentro da seção:
  - título;
  - descrição;
  - lista de toggles/checks das permissões.
- Mostrar apenas os módulos habilitados para o plano atual da empresa.
- Preservar carregamento, submit e feedback de erro por usuário selecionado.

## Fase 0 - Descoberta e Contrato

### 0.1 Ações

- Confirmar a fonte de verdade do plano atual do tenant no backend.
- Confirmar a estratégia de mapeamento entre plano e módulos disponíveis.
- Fechar quais módulos `internal` e `root-only` devem ficar fora da UI do tenant.
- Definir o novo formato de resposta do endpoint de permissões agrupadas.

### 0.2 Checks

- [ ] O contrato de módulos exibíveis por plano está fechado.
- [ ] O contrato da API agrupada por módulo está documentado.

## Fase 1 - Catálogo Agrupado por Módulo

### 1.1 Ações

- Criar o mapeamento de permissões para módulos com base em `permissions.md`.
- Criar o mapeamento de módulos disponíveis por plano com base em `modules.md`.
- Garantir exclusão explícita de módulos `internal` da UI do tenant.

### 1.2 Checks

- [ ] Toda permissão exibível em `/settings` pertence a um módulo bem definido.
- [ ] Todo módulo exibível respeita o pacote mínimo do plano atual.

## Fase 2 - API de Permissões por Módulo

### 2.1 Ações

- Expandir a listagem de permissões por usuário para retorno agrupado por módulo.
- Ajustar update para aceitar o conjunto ampliado de permissões visíveis.
- Rejeitar, no backend, permissões fora do plano da empresa.

### 2.2 Checks

- [ ] A API lista permissões agrupadas por módulo.
- [ ] A API rejeita atribuição de permissão fora do plano do tenant.

## Fase 3 - Refactor da Seção `Permissões` no Web

### 3.1 Ações

- Substituir a lista atual de permissões de configurações por uma UI modular.
- Exibir módulos liberados para o plano atual da empresa.
- Permitir alternância de permissões por usuário dentro de cada módulo.
- Manter a seção visível apenas para `admin`.

### 3.2 Checks

- [ ] A UI mostra permissões agrupadas por módulo.
- [ ] O `admin` só vê módulos disponíveis no plano do tenant.
- [ ] O `system` não vê a seção `Permissões`.

## Fase 4 - Robustez e Testes

### 4.1 Ações

- Cobrir backend para:
  - agrupamento por módulo;
  - filtro por plano;
  - rejeição de permissões fora do plano;
  - isolamento por tenant.
- Cobrir frontend para:
  - renderização modular;
  - filtragem por plano;
  - submit por usuário;
  - estados de loading, erro e readonly onde aplicável.

### 4.2 Checks

- [ ] Há testes cobrindo módulos por plano.
- [ ] Há testes cobrindo edição de permissões agrupadas.

## Riscos e Cuidados

- exibir permissões que o plano do tenant não deveria liberar;
- permitir persistência de permissões fora do escopo do plano;
- confundir permissões padrão de `role` com permissões customizadas do usuário;
- acoplar a UI diretamente a nomes de módulo sem uma camada de contrato estável;
- deixar módulos `internal` visíveis por engano no tenant.

## Decisões Recomendadas

- manter o backend como autoridade final sobre quais módulos e permissões são atribuíveis;
- usar `permissions.md` como fonte de agrupamento funcional;
- usar `modules.md` como fonte de disponibilidade por plano;
- esconder do tenant tudo que for `internal` ou exclusivo de operação central;
- tratar esta evolução como complemento da seção `Permissões`, sem misturar com as seções de `Configurações da empresa` e `Configurações de negócios`.

## Ordem Recomendada de Execução

1. Fase 0: descoberta e contrato.
2. Fase 1: catálogo agrupado por módulo.
3. Fase 2: API de permissões por módulo.
4. Fase 3: refactor da seção `Permissões` no Web.
5. Fase 4: robustez e testes.

## Resultado Final Esperado

Se este plano for executado com sucesso, a seção `Permissões` de `/settings` deixará de refletir apenas configurações da empresa e passará a funcionar como uma gestão real de acesso por módulo, coerente com o catálogo oficial de permissões e com os módulos efetivamente contratados no plano do tenant.
