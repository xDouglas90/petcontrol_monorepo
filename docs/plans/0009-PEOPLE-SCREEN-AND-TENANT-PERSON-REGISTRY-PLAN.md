# Plano de Ação e Execução - Tela de Pessoas do Tenant

## Objetivo

Definir o escopo técnico, funcional e arquitetural da próxima PR do PetControl para implementar a tela `Pessoas` em `/:companySlug/people`, permitindo listar, visualizar, criar e editar pessoas vinculadas ao tenant com comportamento orientado por `person_kind` e por `user_role_type`.

Esta PR deve introduzir:

- rota autenticada real para `/:companySlug/people`;
- item `Pessoas` no menu lateral do shell autenticado;
- listagem completa de pessoas vinculadas ao tenant corrente;
- filtros de apresentação por `person_kind`;
- painel lateral direito contextual para:
  - visualização e edição de uma pessoa selecionada;
  - formulário de criação de nova pessoa;
- fluxo de criação com seleção inicial do tipo de pessoa;
- formulários específicos por `person_kind`;
- criação opcional de usuário de sistema para pessoas elegíveis;
- aplicação das regras de acesso e escopo por `user_role_type`;
- remoção do chat da lateral direita apenas na tela `/people`.

## Contexto Atual

- O Web autenticado já possui shell com sidebar fixa e chat lateral direito para `admin` em [apps/web/src/routes/(app)/_layout.tsx](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/apps/web/src/routes/(app)/_layout.tsx:1).
- O router Web ainda não possui rota `/people`; hoje existem `dashboard`, `schedules`, `clients`, `pets`, `services` e `settings` em [apps/web/src/router.tsx](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/apps/web/src/router.tsx:1).
- O schema já possui a base de dados necessária para modelar pessoas do tenant em [schema.sql](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/schema.sql:585).
- O backend já possui queries SQLC geradas para entidades relevantes como `people`, `company_people`, `people_contacts`, `people_identifications`, `people_addresses`, `people_finances` e `user_profiles`.
- O sistema já diferencia `user_role_type` técnico do `person_kind` de negócio.
- O shell atual reserva o aside direito para o chat em telas autenticadas; esta PR exigirá exceção explícita para `/people`.

## Referências Obrigatórias

- Schema principal: [schema.sql](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/schema.sql:1)
- Layout autenticado: [apps/web/src/routes/(app)/_layout.tsx](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/apps/web/src/routes/(app)/_layout.tsx:1)
- Router Web: [apps/web/src/router.tsx](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/apps/web/src/router.tsx:1)
- Convenção de permissões: [docs/conventions/permissions.md](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/docs/conventions/permissions.md:1)
- Convenção de módulos: [docs/conventions/modules.md](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/docs/conventions/modules.md:1)
- Plano anterior de layout/dashboard: [docs/plans/0006-TENANT-DASHBOARD-LAYOUT-AND-ADMIN-HOME-PLAN.md](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/docs/plans/0006-TENANT-DASHBOARD-LAYOUT-AND-ADMIN-HOME-PLAN.md:1)
- Plano anterior de settings/permissões: [docs/plans/0008-TENANT-SETTINGS-AND-PERMISSIONS-PLAN.md](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/docs/plans/0008-TENANT-SETTINGS-AND-PERMISSIONS-PLAN.md:1)

## Resultado Esperado da PR

Ao final desta PR, o sistema deve apresentar:

- uma tela `Pessoas` navegável a partir do menu lateral;
- listagem de todas as pessoas vinculadas ao tenant, com filtro por tipo;
- abertura de um painel lateral direito para exibir ou editar a pessoa selecionada;
- abertura do mesmo painel para criação de nova pessoa;
- formulários compatíveis com os tipos:
  - `client`
  - `employee`
  - `outsourced_employee`
  - `supplier`
  - `guardian`
  - `responsible`
- criação de vínculo correto nas tabelas relacionais de acordo com o tipo criado;
- criação automática de usuário e credenciais quando aplicável;
- envio do fluxo de credenciais por email quando houver usuário de sistema;
- ausência do chat lateral nesta rota;
- restrição de visibilidade e mutação conforme o papel técnico do usuário autenticado.

## Escopo Funcional Confirmado

## 1. Rota, Navegação e Shell

Esta PR deve criar a rota `/:companySlug/people` dentro do grupo autenticado.

Também deve:

- adicionar o item `Pessoas` no menu lateral;
- manter a linguagem visual do shell atual;
- retirar o `AdminSupportChatAside` apenas nesta tela;
- reservar a lateral direita de `/people` exclusivamente para o painel contextual do módulo;
- manter o layout responsivo, com comportamento coerente em desktop e mobile.

Decisão recomendada para o shell:

- o layout autenticado deve passar a reconhecer a rota atual;
- quando a rota ativa for `/people`, o aside direito de chat não deve ser renderizado;
- o conteúdo principal da tela `/people` deve assumir a composição `lista + painel direito`.

## 2. Fonte de Verdade de Pessoas do Tenant

A tela deve considerar como universo exibível a lista de pessoas vinculadas ao tenant corrente.

Base relacional principal:

- `company_people` como vínculo geral empresa ↔ pessoa;
- `people` como entidade base;
- `people_identifications` para dados civis e nome;
- `people_contacts` para email e telefones;
- `addresses` e `people_addresses` para endereços;
- `finances` e `people_finances` para dados financeiros de funcionários;
- `clients` e `company_clients` para pessoas do tipo cliente;
- `company_employees`, `employments` e `employee_benefits` para funcionários e terceirizados.

Regra de listagem:

- a lista principal de `/people` deve ser tenant-scoped;
- a listagem deve partir de `company_people`;
- tipos que possuam tabelas satélites específicas continuam aparecendo como uma única pessoa na UI;
- a UI não deve depender de consultar cada subtabela para compor a listagem básica.

Campos mínimos recomendados na listagem:

- nome completo;
- nome curto;
- `person_kind`;
- email principal;
- telefone/celular principal;
- indicador de `has_system_user`;
- status ativo;
- data de criação ou vínculo.

## 3. Filtros da Listagem

Deve existir filtro explícito por `person_kind`, incluindo:

- `all`
- `client`
- `employee`
- `outsourced_employee`
- `supplier`
- `guardian`
- `responsible`

Comportamento esperado:

- o filtro atua apenas na apresentação;
- a fonte continua sendo todas as pessoas do tenant;
- o estado do filtro deve refletir na URL ou no estado da tela de forma persistente o suficiente para navegação interna;
- usuários `system` só devem visualizar resultados dos tipos permitidos para eles.

## 4. Painel Direito do Módulo

O lado direito da tela `/people` não deve conter chat.

Estados do painel:

- vazio, quando nenhuma pessoa estiver selecionada e nenhum fluxo de criação estiver aberto;
- detalhes de pessoa selecionada;
- formulário de criação;
- formulário de edição.

Comportamento esperado:

- ao selecionar uma pessoa da listagem, carregar todos os dados relevantes dessa pessoa no painel;
- ao clicar em `Inserir pessoa`, abrir o fluxo de criação no painel;
- ao salvar com sucesso, refletir a alteração na listagem e permanecer no contexto da pessoa criada ou editada;
- o painel deve suportar readonly quando o usuário puder ver, mas não puder editar determinado tipo.

## 5. Fluxo de Criação

O fluxo de criação deve iniciar pela escolha do `person_kind`.

Depois da seleção, o sistema deve renderizar um formulário específico conforme o tipo.

Blocos comuns a praticamente todos os tipos:

- dados base em `people`;
- identificação pessoal em `people_identifications`;
- contato principal em `people_contacts`;
- endereço principal em `addresses` + tabela de vínculo pertinente;
- dados financeiros apenas para tipos de funcionário na primeira versão.

Blocos específicos por tipo:

- `client`
  - criar `people`;
  - criar `company_people`;
  - criar `clients`;
  - criar `company_clients`;
  - opcionalmente criar usuário de sistema quando permitido pela regra de papel técnico.
- `employee`
  - criar `people`;
  - criar `company_people`;
  - criar `company_employees`;
  - criar `employments`;
  - criar `employee_benefits` quando informado;
  - opcionalmente criar usuário de sistema.
- `outsourced_employee`
  - mesmo fluxo estrutural de `employee`;
  - o formulário deve deixar claro que se trata de terceirizado.
- `supplier`
  - criar `people`;
  - criar `company_people`;
  - persistir contatos, endereços e finanças conforme preenchimento.
- `guardian`
  - criar `people`;
  - criar `company_people`;
  - persistir contatos e endereço.
- `responsible`
  - criar `people`;
  - criar `company_people`;
  - persistir contatos e endereço.

Observação importante:

- como o schema não possui tabela dedicada para `supplier`, `guardian` e `responsible`, o vínculo de tenant destas categorias deve viver em `company_people`, complementado pelas tabelas comuns;
- para esta PR, `guardian` e `responsible` permanecem tipos suportados no domínio, mas fora do escopo de criação e edição por usuários `system`.

## 6. Regras de `has_system_user`

Se o tipo criado for `employee` ou `outsourced_employee`, a UI deve perguntar explicitamente se a pessoa terá usuário de sistema por meio de `has_system_user`.

Se o tipo criado for `client`, esta opção só pode existir para usuários autenticados do tipo `system`, conforme regra informada.

Regras funcionais:

- se `has_system_user = false`, o fluxo termina apenas com a criação da pessoa e seus vínculos;
- se `has_system_user = true`, o backend deve concluir também o provisionamento completo do usuário;
- a UI não deve gerar senha nem tentar montar permissões localmente;
- a UI apenas envia a intenção e os dados necessários.

## 7. Provisionamento Automático de Usuário

Quando o cadastro exigir usuário de sistema, o backend deve:

- inserir um novo registro em `users`;
- usar `people_contacts.email` como email do novo usuário;
- gerar senha automática no backend;
- persistir credencial em `user_auth`;
- vincular o novo usuário à pessoa em `user_profiles`;
- encaminhar email com:
  - link do sistema;
  - email de acesso;
  - senha temporária ou instrução equivalente definida pelo backend;
- obrigar o fluxo padrão de primeiro acesso, se o backend já suportar isso.

Regra de `role` do usuário criado:

- para pessoas `employee` e `outsourced_employee`, o papel esperado é `system`;
- para pessoa `client` criada por usuário autenticado `system`, o papel esperado é `common`.

Ponto de atenção:

- o plano informado pelo usuário exige que o novo usuário nasça com `user_permissions` baseadas nas permissões do tenant do cadastrante e nos módulos padrões do plano ativo do tenant;
- isso deve ser resolvido inteiramente no backend, reutilizando o catálogo de permissões e as regras já introduzidas nos planos `0008`, `0008_2` e `0008_3`.

## 8. Regras de Acesso por `user_role_type`

## `admin`

Usuários `admin` devem poder:

- visualizar a tela `/people`;
- listar todos os tipos de pessoa do tenant;
- criar qualquer `person_kind` suportado;
- editar qualquer `person_kind` suportado;
- criar usuário para `employee` e `outsourced_employee`.

## `system`

Usuários `system` devem poder:

- visualizar a tela `/people`;
- visualizar, criar e editar apenas pessoas dos tipos:
  - `client`
  - `supplier`
- não criar nem editar:
  - `employee`
  - `outsourced_employee`
  - `guardian`
  - `responsible`
- só poder criar usuário de sistema para pessoas do tipo `client`;
- quando criarem usuário para `client`, o novo `users.role` deve ser `common`.

Regra de UX recomendada:

- tipos não permitidos para `system` não devem aparecer como opção de criação;
- registros fora do escopo de `system` não devem aparecer como editáveis;
- para esta PR, a visualização editável de `system` deve se limitar a `client` e `supplier`.

Recomendação para esta PR:

- manter coerência forte entre backend e frontend;
- o backend deve ser a camada final de autorização por tipo;
- o frontend deve esconder ações inválidas para reduzir erro de uso.

## 9. Dados Exibidos no Painel de Pessoa

Ao selecionar uma pessoa, o painel direito deve consolidar, no mínimo:

- dados base da pessoa;
- identificação;
- contatos;
- endereço principal;
- dados financeiros principais, quando existirem para tipos de funcionário;
- vínculos específicos do tipo;
- indicador se possui usuário do sistema;
- dados resumidos do usuário vinculado, se existir.

Campos específicos adicionais por tipo:

- `client`
  - `client_since`;
  - `notes`.
- `employee` e `outsourced_employee`
  - `employment.role`;
  - `admission_date`;
  - `resignation_date`;
  - `salary`;
  - benefícios vigentes.

## 10. Estratégia de API Recomendada

Para manter a PR coesa, recomenda-se expor um contrato orientado ao módulo, ao invés de vários submits independentes no frontend.

Sugestão de endpoints:

- `GET /people`
  - lista pessoas do tenant com filtros;
- `GET /people/:personId`
  - retorna visão expandida da pessoa;
- `POST /people`
  - cria pessoa e todos os vínculos necessários conforme o payload e o `person_kind`;
- `PATCH /people/:personId`
  - atualiza pessoa e seus agregados principais.

Requisitos desses endpoints:

- sempre tenant-scoped;
- sempre validar `person_kind`;
- sempre validar autorização por `user_role_type`;
- executar criação e edição em transação;
- garantir idempotência razoável nas subtabelas one-to-one;
- impedir que o frontend grave combinações inválidas entre tipo e vínculos.

## 11. Estratégia de Persistência por Tipo

O backend deve tratar `people` como agregado de escrita.

Ordem recomendada na criação:

1. inserir `people`;
2. inserir `people_identifications`;
3. inserir `people_contacts`;
4. inserir endereço e vínculos de endereço;
5. inserir finanças e vínculos financeiros, quando a pessoa criada for do tipo `employee` ou `outsourced_employee`;
6. inserir `company_people`;
7. inserir vínculos específicos do tipo;
8. se aplicável, provisionar `users`, `user_auth`, `user_profiles` e permissões;
9. disparar email após confirmação transacional do cadastro.

## 12. Permissões e Catálogo de Módulos

Este módulo ainda não aparece formalizado em [docs/conventions/modules.md](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/docs/conventions/modules.md:1) como um módulo próprio de `People`.

Decisão recomendada:

- esta PR deve explicitar se `/people` será:
  - um módulo próprio de `People`; ou
  - uma composição de permissões de `Clients`, `Suppliers` e `Users`.

Como o comportamento pedido agrega múltiplos `person_kind` em uma tela única, a recomendação mais consistente é:

- criar um módulo próprio de `People`;
- definir permissões mínimas como `people:view`, `people:create` e `people:update`;
- manter validações complementares por `person_kind` para o papel `system`.

Se o time optar por não introduzir novo módulo nesta PR, isso deve ser assumido explicitamente como dívida técnica, porque a autorização da tela ficará espalhada entre regras de papel e regras por tipo.

## 13. Impacto no Layout e Navegação

Mudanças mínimas no Web:

- adicionar a rota `people` ao router;
- criar a página `apps/web/src/routes/(app)/people/index.tsx`;
- incluir `Pessoas` na navegação lateral;
- ajustar `AppLayout` para não renderizar chat na rota `/people`;
- assegurar que o espaço do aside direito continue previsível mesmo sem chat;
- preservar o chat no `dashboard`, conforme solicitado.

## 14. Testes Recomendados

## 14.1 Backend

- criação de `client` sem usuário;
- criação de `client` com usuário por usuário `system`, gerando `role = common`;
- criação de `employee` com e sem usuário;
- criação de `outsourced_employee` com e sem usuário;
- bloqueio de `system` tentando criar `employee` ou `outsourced_employee`;
- bloqueio de `system` tentando criar usuário para tipos diferentes de `client`;
- listagem tenant-scoped sem vazamento entre empresas;
- atualização de pessoa respeitando subtabelas e autorização;
- criação transacional com rollback se falhar a etapa de vínculo específico;
- criação de `user_profiles` e permissões padrão do plano.

## 14.2 Frontend

- menu mostra `Pessoas`;
- rota `/people` renderiza sem o chat lateral;
- lista aplica filtro por `person_kind`;
- clique em uma pessoa abre o painel direito;
- `Inserir pessoa` abre seletor de tipo;
- `system` enxerga apenas tipos permitidos no formulário de criação;
- campo `has_system_user` aparece apenas quando permitido;
- sucesso de criação atualiza lista e painel;
- estados de loading, erro e vazio ficam estáveis.

## 15. Riscos e Pontos de Atenção

- o schema usa múltiplas subtabelas e parte delas é opcional; sem contrato agregado, o frontend tende a ficar excessivamente acoplado;
- `people_contacts` não impõe unicidade global de email, então o backend precisará validar conflito antes de criar `users`;
- a criação de usuário exige integração confiável com envio de email e fluxo de senha temporária;
- é preciso evitar divergência entre `company_people` e tabelas satélites específicas como `company_clients` e `company_employees`;
- a edição de tipos distintos em uma única tela pode gerar formulários muito grandes se não houver segmentação visual;
- a modelagem atual ainda contém `company_people_addresses`, mas para esta PR a decisão é usar `people_addresses` como fonte de verdade do endereço principal; isso torna `company_people_addresses` redundante para o fluxo definido e deve levá-la a sair do escopo desta implementação;
- para `employee` e `outsourced_employee`, `employee_documents` passa a fazer parte do escopo desta PR.

## 16. Fora de Escopo Recomendado para Esta PR

- anexos e upload de documentos trabalhistas;
- histórico completo de múltiplos contatos, múltiplos endereços e múltiplas contas bancárias por pessoa;
- gestão avançada de benefícios com linha do tempo;
- redefinição manual de senha pela tela de pessoas;
- convites com aceitação por token, caso o backend ainda não tenha este fluxo;
- refatoração ampla do catálogo de permissões além do necessário para liberar `/people`.

## 17. MailHog para Testes Locais de Email

Como esta PR inclui criação opcional de usuários com envio de credenciais por email, o ambiente local deve passar a suportar inspeção real de mensagens disparadas pelo backend.

Esta PR deve incluir a configuração de `MailHog` para desenvolvimento local e testes automatizados locais do fluxo de recebimento de email, além da inspeção manual pela interface web.

Objetivos do MailHog nesta entrega:

- permitir validar localmente o disparo de email após criação de usuário;
- inspecionar assunto, destinatário e corpo da mensagem;
- validar presença do link do sistema e credenciais temporárias;
- suportar testes automatizados locais do fluxo de envio;
- evitar dependência de provedores reais de SMTP durante desenvolvimento.

Escopo recomendado:

- subir `MailHog` via `docker compose` ou mecanismo local equivalente já adotado pelo projeto;
- parametrizar o backend para usar SMTP local em ambiente de desenvolvimento;
- documentar host, portas e URL da interface web do MailHog;
- garantir que o fluxo de criação de usuário da tela `/people` seja testável apontando para MailHog;
- manter a configuração separada por ambiente, sem impactar staging ou produção.

Configuração mínima esperada:

- serviço `mailhog` exposto localmente;
- porta SMTP padrão local apontada para o backend;
- interface web acessível para leitura das mensagens capturadas;
- variáveis de ambiente explícitas para host, porta, usuário, senha e remetente, mesmo que localmente algumas delas sejam vazias.

Recomendação prática:

- usar MailHog para `development` e `test local`;
- centralizar a configuração SMTP em um único ponto do backend;
- não acoplar a implementação da feature `/people` a um provedor externo real.

## 18. Mini Plano de Remoção de `company_people_addresses`

Como a decisão funcional desta PR passa a usar `people_addresses` como fonte de verdade do endereço da pessoa, a tabela `company_people_addresses` tende a se tornar redundante.

Hoje o impacto conhecido da tabela é:

- definição no schema em [schema.sql](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/schema.sql:796);
- definição equivalente na migration inicial em [000001_init_schema.up.sql](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/infra/migrations/000001_init_schema.up.sql:796);
- remoção correspondente na down migration em [000001_init_schema.down.sql](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/infra/migrations/000001_init_schema.down.sql:81);
- arquivo de queries SQLC dedicado em [infra/sql/queries/company_people_addresses.sql](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/infra/sql/queries/company_people_addresses.sql:1);
- artefatos gerados em:
  - [apps/api/internal/db/sqlc/company_people_addresses.sql.go](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/apps/api/internal/db/sqlc/company_people_addresses.sql.go:1)
  - [apps/api/internal/db/sqlc/models.go](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/apps/api/internal/db/sqlc/models.go:1164)
  - [apps/api/internal/db/sqlc/querier.go](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/apps/api/internal/db/sqlc/querier.go:1);
- diagrama ER em [docs/diagrams/er-diagram.mmd](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/docs/diagrams/er-diagram.mmd:1).

### Objetivo da remoção

- eliminar duplicidade entre vínculo `company_people` + `people_addresses`;
- simplificar o agregado de pessoa no backend;
- reduzir superfície de queries SQLC não utilizadas;
- evitar divergência futura entre endereço da pessoa e endereço da pessoa no contexto da empresa.

### Passos Técnicos Recomendados

1. Confirmar ausência de uso funcional fora do módulo de People.
   - validar se nenhum handler, service ou fluxo legado depende de `GetCompanyPeopleAddresses`, `ListCompanyPeopleAddresses` ou correlatas;
   - validar se não há testes de integração usando essa tabela como pré-condição.

2. Remover a modelagem do schema base.
   - remover `CREATE TABLE company_people_addresses` de [schema.sql](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/schema.sql:796);
   - remover o bloco equivalente de [000001_init_schema.up.sql](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/infra/migrations/000001_init_schema.up.sql:796);
   - remover o `DROP TABLE IF EXISTS company_people_addresses;` de [000001_init_schema.down.sql](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/infra/migrations/000001_init_schema.down.sql:81).

3. Remover a fonte de queries do SQLC.
   - excluir [infra/sql/queries/company_people_addresses.sql](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/infra/sql/queries/company_people_addresses.sql:1);
   - garantir que qualquer leitura necessária passe a usar `people_addresses` e `addresses`.

4. Regenerar o SQLC.
   - executar a geração do SQLC após remover schema e query source;
   - verificar a remoção automática dos artefatos gerados relacionados.

5. Validar o impacto esperado nos artefatos gerados.
   - o arquivo gerado `company_people_addresses.sql.go` deve desaparecer;
   - o tipo `CompanyPeopleAddress` deve sair de `models.go`;
   - os métodos abaixo devem sair de `querier.go`:
     - `DeleteCompanyPeopleAddresses`
     - `GetCompanyPeopleAddresses`
     - `InsertCompanyPeopleAddresses`
     - `ListCompanyPeopleAddresses`
     - `ListCompanyPersonAddresses`
     - `UpdateCompanyPeopleAddresses`

6. Corrigir referências quebradas após regeneração.
   - ajustar imports e usos do `Querier` onde esses métodos apareçam;
   - substituir por queries baseadas em `people_addresses`, se alguma leitura ainda for necessária;
   - remover testes ou fixtures que dependam da tabela excluída.

7. Atualizar documentação estrutural.
   - remover `COMPANY_PEOPLE_ADDRESSES` do diagrama ER;
   - ajustar qualquer plano ou convenção que ainda cite essa tabela como parte do fluxo principal de pessoas.

### Impacto Esperado nas Queries SQLC

As seguintes queries deixam de existir:

- `InsertCompanyPeopleAddresses`
- `GetCompanyPeopleAddresses`
- `UpdateCompanyPeopleAddresses`
- `DeleteCompanyPeopleAddresses`
- `ListCompanyPeopleAddresses`
- `ListCompanyPersonAddresses`

Impactos colaterais esperados:

- remoção do arquivo gerado `apps/api/internal/db/sqlc/company_people_addresses.sql.go`;
- remoção do tipo `CompanyPeopleAddress` em `models.go`;
- mudança da interface `Querier`, o que pode quebrar compilações em serviços, handlers, mocks e testes que ainda referenciem esses métodos;
- eventual necessidade de regenerar mocks ou stubs baseados na interface do SQLC.

### Estratégia de Segurança

- fazer a remoção em uma PR pequena e isolada, ou no começo da PR de `/people`, antes de espalhar novos usos;
- compilar a API imediatamente após regenerar o SQLC para localizar dependências ocultas;
- rodar testes de banco e integração logo após a remoção;
- se aparecer dependência real de contexto por empresa, reavaliar a exclusão antes de introduzir uma migração destrutiva.

### Critério de Conclusão da Remoção

- `company_people_addresses` deixa de existir no schema e nas migrations base;
- não existe mais query source para essa tabela em `infra/sql/queries`;
- o SQLC é regenerado sem `CompanyPeopleAddress` e sem métodos associados;
- a API compila sem referências órfãs;
- o diagrama ER deixa de exibir essa relação.

## Critérios de Aceite

- [x] Existe rota autenticada `/:companySlug/people`.
- [x] O menu lateral exibe `Pessoas`.
- [x] A tela lista pessoas vinculadas ao tenant.
- [x] A listagem possui filtro por `person_kind`.
- [x] O chat lateral não aparece em `/people`.
- [x] O painel direito mostra os dados da pessoa selecionada.
- [x] O painel direito também suporta o fluxo de criação.
- [x] O formulário muda conforme o `person_kind`.
- [x] `admin` pode criar e editar todos os tipos suportados.
- [x] `system` só pode visualizar, criar e editar `client` e `supplier`.
- [x] `system` só pode criar usuário para `client`.
- [x] `employee` e `outsourced_employee` suportam pergunta explícita de `has_system_user`.
- [x] Quando houver criação de usuário, o backend cria `users`, `user_auth` e `user_profiles`.
- [x] O email do novo usuário vem de `people_contacts.email`.
- [x] As permissões iniciais do novo usuário são derivadas do tenant/plano do cadastrante.
- [x] O ambiente local possui MailHog configurado para receber os emails disparados pelo backend.
- [x] O fluxo de criação de usuário pode ser validado localmente pela interface do MailHog.
- [x] A implementação é coberta por testes mínimos de backend e frontend.

## Sequência Recomendada de Execução

1. Definir o contrato de autorização da tela e, se aprovado, catalogar o módulo/permissões de `People`.
2. Criar os endpoints agregados de listagem, detalhe, criação e edição.
3. Implementar a transação de criação por `person_kind`, incluindo provisionamento opcional de usuário.
4. Adicionar a rota `/people`, o item de navegação e a exceção do chat no shell.
5. Implementar a tela com lista, filtros e painel lateral direito.
6. Implementar os formulários específicos por tipo.
7. Conectar mutations e estados de sucesso/erro.
8. Cobrir o fluxo com testes automatizados.

## Checklist Técnico

## Backend

- [x] Usar `People` como módulo próprio com permissões dedicadas:
  - `people:view`
  - `people:create`
  - `people:update`
- [x] Refletir o módulo `People` no catálogo de módulos, permissões e planos.
- [x] Garantir que a autorização final do recurso também valide restrições por `person_kind`, não apenas por permissão genérica.
- [x] Definir contrato DTO de listagem de pessoas tenant-scoped com filtros por `person_kind`.
- [x] Definir contrato DTO de detalhe de pessoa com agregação de:
  - `people`
  - `people_identifications`
  - `people_contacts`
  - endereços
  - finanças
  - vínculos específicos do tipo
  - resumo de usuário vinculado
- [x] Definir contrato DTO de criação de pessoa com seleção explícita de `person_kind`.
- [x] Definir contrato DTO de edição de pessoa com payload coerente por tipo.
- [x] Criar ou adaptar endpoint `GET /people` tenant-scoped.
- [x] Criar ou adaptar endpoint `GET /people/:personId` tenant-scoped.
- [x] Criar endpoint `POST /people` como escrita agregada transacional.
- [x] Criar endpoint `PATCH /people/:personId` como atualização agregada transacional.
- [x] Garantir que o list endpoint parta de `company_people`, evitando vazamento cross-tenant.
- [x] Implementar filtro por `person_kind` na listagem.
- [x] Implementar ordenação estável da listagem, preferencialmente por nome.
- [x] Validar que `system` só possa criar/editar:
  - `client`
  - `supplier`
- [x] Validar que `system` não possa criar/editar:
  - `employee`
  - `outsourced_employee`
  - `guardian`
  - `responsible`
- [x] Validar que `admin` possa criar/editar todos os tipos suportados.
- [x] Implementar fluxo de criação base em `people`.
- [x] Implementar criação de `people_identifications`.
- [x] Implementar criação e seleção de contato principal em `people_contacts`.
- [x] Implementar criação de endereço principal em `addresses`.
- [x] Implementar vínculo de endereço principal em `people_addresses` como fonte de verdade do módulo.
- [ ] Remover `company_people_addresses` do escopo desta PR e registrar sua redundância técnica na modelagem atual.
- [ ] Avaliar remoção efetiva de `company_people_addresses` do schema e das queries geradas, desde que não exista uso externo bloqueante.
- [x] Implementar criação de dados financeiros em `finances` apenas para:
  - `employee`
  - `outsourced_employee`
- [x] Implementar vínculo financeiro em `people_finances` apenas para tipos de funcionário na primeira versão.
- [x] Implementar criação obrigatória de `company_people` para todo tipo suportado.
- [x] Implementar criação de `clients` para `person_kind = client`.
- [x] Implementar criação de `company_clients` após criação do cliente.
- [x] Implementar criação de `company_employees` para:
  - `employee`
  - `outsourced_employee`
- [x] Implementar criação de `employments` para tipos de funcionário.
- [x] Implementar criação opcional de `employee_benefits` quando informado.
- [x] Implementar `employee_documents` quando a pessoa criada for do tipo:
  - `employee`
  - `outsourced_employee`
- [x] Validar `has_system_user` apenas para combinações permitidas.
- [x] Permitir `has_system_user` para `employee` e `outsourced_employee` quando o usuário autenticado puder criar esses tipos.
- [x] Permitir `has_system_user` para `client` apenas quando o usuário autenticado for `system`, conforme regra informada.
- [x] Impedir `has_system_user` para `supplier`, `guardian` e `responsible` nesta PR.
- [x] Antes de criar `users`, validar unicidade/consistência de `people_contacts.email`.
- [x] Criar `users` com `role = system` para:
  - `employee`
  - `outsourced_employee`
- [x] Criar `users` com `role = common` para `client` quando esse fluxo estiver habilitado.
- [x] Criar `user_auth` com senha gerada automaticamente no backend.
- [x] Marcar comportamento de primeiro acesso, como `must_change_password`, se já fizer parte do fluxo existente.
- [x] Criar vínculo em `user_profiles`.
- [x] Aplicar bootstrap de `user_permissions` usando o plano ativo do tenant e as permissões padrão esperadas para os módulos liberados.
- [x] Garantir que o usuário recém-criado não receba permissões fora do plano ativo do tenant.
- [x] Disparar envio de email com link do sistema e credenciais de forma assíncrona após sucesso do cadastro.
- [x] Adicionar suporte de configuração SMTP por ambiente para permitir uso de MailHog localmente.
- [x] Garantir que o serviço de envio de email aceite host/porta/remetente configuráveis por env.
- [x] Garantir fallback seguro entre ambiente local e ambientes reais.
- [x] Documentar as envs necessárias para integração com MailHog.
- [x] Garantir rollback transacional se qualquer etapa crítica falhar antes da conclusão.
- [ ] Tratar falhas de email como pós-processamento assíncrono confiável, com mecanismo de retry ou compensação.
- [x] Garantir que a leitura de detalhe retorne dados suficientes para preencher painel de visualização e edição.
- [x] Garantir que a edição atualize apenas blocos permitidos para o tipo da pessoa.
- [x] Garantir que a edição não permita migrar a pessoa para um `person_kind` incompatível sem regra explícita.
- [x] Adicionar testes de serviço/handler para listagem tenant-scoped.
- [x] Adicionar testes de serviço/handler para filtros por tipo.
- [ ] Adicionar testes para criação de `client` sem usuário.
- [ ] Adicionar testes para criação de `client` com usuário `common` por usuário autenticado `system`.
- [ ] Adicionar testes para criação de `employee` com usuário `system`.
- [ ] Adicionar testes para criação de `outsourced_employee` com usuário `system`.
- [ ] Adicionar testes para bloqueio de `system` tentando criar `employee`.
- [ ] Adicionar testes para bloqueio de `system` tentando criar `outsourced_employee`.
- [ ] Adicionar testes para bloqueio de `system` tentando criar usuário para tipo não permitido.
- [ ] Adicionar testes para criação de vínculos:
  - `company_people`
  - `company_clients`
  - `company_employees`
  - `user_profiles`
- [ ] Adicionar cobertura de teste local automatizado usando MailHog para criação de usuário e recebimento do email.
- [x] Manter também um fluxo simples de validação manual pela UI do MailHog para desenvolvimento.
- [ ] Adicionar testes de rollback transacional.
- [x] Atualizar documentação da API ou swagger do domínio, se fizer parte do padrão atual.

- [x] Adicionar ou atualizar infraestrutura local para subir `MailHog`.
- [x] Se o projeto usar `docker compose`, incluir o serviço `mailhog` na composição local.
- [x] Expor no ambiente local a URL da UI do MailHog para conferência das mensagens.

## Web

- [x] Adicionar a rota `/:companySlug/people` no router autenticado.
- [x] Criar a página [apps/web/src/routes/(app)/people/index.tsx](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/apps/web/src/routes/(app)/people/index.tsx:1).
- [x] Adicionar `Pessoas` ao menu lateral do `AppLayout`.
- [x] Ajustar o `AppLayout` para não renderizar o chat lateral em `/people`.
- [x] Preservar o chat apenas no `dashboard`, conforme diretriz do produto.
- [x] Garantir que a área direita da tela `/people` fique reservada ao painel do módulo.
- [x] Definir query key e hooks de React Query para:
  - listagem de pessoas
  - detalhe da pessoa
  - criação
  - edição
- [x] Criar camada de client HTTP para os endpoints de `people`.
- [x] Definir tipos TS do payload de listagem, detalhe, criação e edição.
- [x] Implementar estado base da página com:
  - lista à esquerda
  - painel à direita
  - estado vazio quando nada estiver selecionado
- [x] Implementar filtro por `person_kind`.
- [x] Implementar persistência do filtro em search params ou estado equivalente.
- [x] Implementar loading state da listagem.
- [x] Implementar empty state da listagem.
- [x] Implementar error state da listagem.
- [x] Exibir na lista pelo menos:
  - nome
  - tipo
  - email principal
  - indicador de usuário do sistema
  - status
- [x] Implementar ação `Inserir pessoa`.
- [x] Ao iniciar criação, abrir painel direito no modo create.
- [x] Implementar seletor inicial de `person_kind`.
- [x] Para usuário `admin`, exibir todos os tipos permitidos no seletor.
- [x] Para usuário `system`, exibir apenas:
  - `client`
  - `supplier`
- [x] Implementar formulário base compartilhado para:
  - identificação
  - contato principal
  - endereço principal
  - finanças, quando aplicável
- [x] Implementar variação de formulário para `client`.
- [x] Implementar variação de formulário para `employee`.
- [x] Implementar variação de formulário para `outsourced_employee`.
- [x] Implementar variação de formulário para `supplier`.
- [x] Implementar variação de formulário para `guardian`.
- [x] Implementar variação de formulário para `responsible`.
- [x] Exibir o campo `has_system_user` apenas quando o tipo e o papel autenticado permitirem.
- [x] Se `person_kind = employee`, exibir pergunta sobre criação de usuário.
- [x] Se `person_kind = outsourced_employee`, exibir pergunta sobre criação de usuário.
- [x] Se `person_kind = client` e o usuário autenticado for `system`, exibir pergunta sobre criação de usuário.
- [x] Não exibir pergunta de usuário do sistema para:
  - `supplier`
  - `guardian`
  - `responsible`
- [x] Implementar submit de criação com payload agregado.
- [x] Após criação com sucesso, invalidar cache da listagem.
- [x] Após criação com sucesso, abrir no painel o detalhe da pessoa criada.
- [x] Implementar carregamento do detalhe ao selecionar pessoa na lista.
- [x] Renderizar modo readonly quando o usuário puder ver, mas não puder editar determinado registro.
- [x] Implementar modo de edição no painel direito.
- [x] Implementar submit de edição com atualização otimista ou invalidação simples, conforme padrão do projeto.
- [x] Exibir feedback de sucesso de criação e edição.
- [x] Exibir feedback de erro com mensagens utilizáveis.
- [x] Garantir responsividade da composição `lista + painel`.
- [x] Garantir usabilidade mobile sem depender do chat.
- [x] Adicionar testes da rota `/people`.
- [x] Adicionar testes do menu contendo `Pessoas`.
- [x] Adicionar testes garantindo ausência do chat em `/people`.
- [x] Adicionar testes do filtro por `person_kind`.
- [x] Adicionar testes do fluxo de seleção de pessoa e abertura do painel.
- [ ] Adicionar testes do fluxo de criação para `admin`.
- [x] Adicionar testes do fluxo de criação restrito para `system`.
- [x] Adicionar testes condicionais do campo `has_system_user`.
- [ ] Adicionar testes de sucesso e erro das mutations.

## Decisões Fechadas para Implementação

- [x] `People` será módulo próprio com permissões dedicadas, já refletido nas convenções.
- [x] Usuários `system` poderão visualizar, criar e editar apenas `client` e `supplier`.
- [x] O endereço principal será persistido em `people_addresses`.
- [x] `company_people_addresses` é redundante para o fluxo desta PR e sai do escopo da implementação; sua remoção do schema deve ser avaliada na execução, desde que não haja dependências externas relevantes.
- [x] Dados financeiros entram apenas para `employee` e `outsourced_employee` na primeira versão.
- [x] O envio de email será assíncrono.
- [x] `employee_documents` entra no escopo quando a pessoa criada for funcionário.
- [x] MailHog será usado tanto para validação manual quanto para testes automatizados locais.

## Conclusão

Esta PR estabelece a base do cadastro unificado de pessoas do tenant e conecta, em uma única experiência, clientes, funcionários, terceirizados, fornecedores, guardiões e responsáveis. Com as decisões agora fechadas, o plano passa a assumir `People` como módulo próprio, restringe o escopo de `system` a `client` e `supplier`, usa `people_addresses` como fonte de verdade para endereço, inclui `employee_documents` para funcionários e adota envio assíncrono de email com suporte local via MailHog para validação manual e automatizada.
