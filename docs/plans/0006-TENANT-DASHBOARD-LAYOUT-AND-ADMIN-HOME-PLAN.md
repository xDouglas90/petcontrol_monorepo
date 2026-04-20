# Plano de Ação e Execução - Novo Layout da Aplicação e Dashboard Inicial por Tipo de Usuário

## Objetivo

Definir o escopo, a arquitetura de UX e o plano de implementação para a proxima PR do PetControl, cujo objetivo e reformular o layout principal da aplicação Web para aproximar a experiencia visual e estrutural da referencia `example-001.png`, preservando a identidade multi-tenant do produto e conectando o layout a dados reais do domínio.

Esta PR não deve ser tratada apenas como um "retoque visual". Ela introduz:

- um novo shell de navegação da area autenticada;
- novo comportamento do menu lateral com branding do tenant;
- bloco comercial de upgrade sensível ao plano atual do tenant;
- footer navegacional reorganizado;
- home inicial variando por tipo de usuário;
- primeira home especializada para o tipo `admin`;
- dashboard operacional e gerencial do tenant;
- area de chat textual entre `admin` e usuários do tipo `system`.

## Referencia Visual

A referencia visual desta fase e a imagem `example-001.png` fornecida no contexto desta PR.

Os elementos principais a serem reinterpretados no PetControl sao:

- sidebar branca/clara com branding no topo;
- cards de indicadores no topo do conteúdo;
- gráfico central de performance;
- lista operacional inferior;
- coluna lateral direita com contexto humano e chat;
- layout com atmosfera mais executiva e menos "painel técnicos".

Importante:

- a referencia deve orientar estrutura, densidade, hierarquia visual e fluxo;
- o produto não precisa copiar literalmente nomes, textos e entidades da imagem;
- os blocos devem ser preenchidos com dados reais do domínio PetControl.

## Contexto Atual

- O Web já possui area autenticada por `/:companySlug/...`.
- O layout autenticado atual já suporta sidebar, header, dashboard e módulos operacionais.
- O sistema já possui contexto multi-tenant com empresa corrente.
- O schema já contem colunas e tabelas relevantes para:
  - `companies.logo_url`;
  - `people_identifications.short_name`;
  - `people_identifications.image_url`;
  - `company_system_configs`;
  - `schedules`;
  - `schedule_status_history`.
- Os planos de tenant e sua hierarquia estão documentados em [plans.md](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/docs/conventions/plans.md:1).

## Resultado Esperado da PR

Ao final desta PR, o sistema deve apresentar:

- novo layout global autenticado alinhado a referencia visual;
- menu lateral do tenant com logo real da empresa;
- CTA de upgrade coerente com o plano atual do tenant;
- footer lateral com `Configurações` e `Sair`;
- definição explicita de que a home inicial depende do tipo de usuário;
- implementação da home para `admin`;
- dashboard `admin` com métricas, comparativos, gráfico e lista operacional;
- base estrutural para o chat textual com usuários `system`.

## Escopo Funcional Confirmado

## 1. Novo Menu Lateral

### Topo

- Exibir a logo da empresa corrente usando `companies.logo_url`.
- Caso não exista logo, exibir fallback visual coerente com a marca do produto.

### Navegação principal

Itens confirmados:

- Dashboard
- Agendamentos
- Clientes
- Pets

Observação:

- `Services` sai da navegação principal desta referencia inicial, salvo se a PR decidir mantê-lo em rota secundaria fora do primeiro recorte visual.

### Bloco de upgrade

Manter um bloco semelhante ao merchant da imagem de referencia, mas com comportamento orientado ao negocio:

- detectar o plano atual do tenant;
- sugerir o plano imediatamente superior;
- usar a hierarquia oficial:
  - `premium > essential > basic > starter > trial`

Regras iniciais sugeridas:

- `trial` sugere `starter`
- `starter` sugere `basic`
- `basic` sugere `essential`
- `essential` sugere `premium`
- `premium` não exibe upgrade, ou exibe CTA alternativo de suporte/expansão
- `internal` não entra na hierarquia comercial de tenant e deve ter tratamento especial

### Footer

Itens confirmados no rodapé do menu lateral:

- Configurações
- Sair

## 2. Home Inicial por Tipo de Usuário

A home inicial deve variar conforme o tipo de usuários autenticado.

Tipos explicitamente citados para mapeamento:

- `root`
- `internal`
- `admin`
- `system`
- `common`
- `free`

Decisão desta PR:

- implementar primeiro a home do tipo `admin`;
- deixar registrada a estrategia para os demais tipos sem necessariamente implementa-los agora.

## 3. Dashboard Inicial do Tipo `admin`

O dashboard do administrador do tenant sera a primeira implementação concreta do novo modelo.

### 3.1 Bloco superior

Exibir:

- saudação: `Ola, {people_identifications.short_name}`
- texto de apoio explicando o que esta sendo visualizado
- data atual no formato `DD de MM, AAAA`
- icone de calendário ao lado da data

### 3.2 Cards de indicadores

Devem existir tres cards principais, com:

- icone;
- titulo;
- valor principal;
- indicador positivo ou negativo;
- comparativo com período anterior.

#### Card 1 - Agendamentos/dia

Conteúdo:

- total de agendamentos do dia atual;
- atualização refletindo novos agendamentos do tenant;
- Seta de comparação com o total do dia anterior.
  - para cima e verde quando o dia atual for maior que o anterior;
  - para baixo e vermelho quando o dia atual for menor;
  - sem seta e neutro quando igual.

#### Card 2 - Agendamentos/mês

Conteúdo:

- total de agendamentos do mês atual;
- comparação com o total do mês anterior.
  - para cima e verde quando o mês atual superar o anterior;
  - para baixo e vermelho quando estiver abaixo;
  - sem seta e neutro quando igual.

#### Card 3 - Eficiência

Conteúdo:

- porcentagem de atingimento da meta operacional minima do mês.

Calculo base:

- obter em `company_system_configs`:
  - `min_schedules_per_day`
  - `schedule_days`
- calcular meta mensal minima:
  - `min_schedules_per_day * total_de_dias_de_atendimento_no_mes`
- calcular percentual realizado:
  - `total_agendamentos_mes / meta_mensal_minima * 100`

Comparativo esperado:

- seta para cima e verde, quanto passou da meta;
- seta para baixo e vermelho, quanto falta para a meta;
- sem seta e neutro, quando atinge exatamente 100%.

Observação importante:

- o plano menciona `schedule_days`, mas a implementação deve confirmar se esse campo representa uma coleção de dias da semana, dias efetivos do mês ou outra modelagem;
- caso a modelagem real exija interpretação adicional, a regra deve ser explicitada antes da implementação final.

### 3.3 Gráfico de Performance

Bloco central semelhante a referencia, adaptado para agendamentos.

#### Estrutura

- titulo `Performance`
- select no canto direito para escolher o intervalo semanal
- eixo Y baseado na janela operacional da empresa:
  - `company_system_configs.schedule_init_time`
  - `company_system_configs.schedule_end_time`
- eixo X baseado nos dias da semana selecionada, por exemplo:
  - `01-07 MM`

#### Comparação

O gráfico deve comparar:

- período selecionado do mês atual
- mesmo recorte do mês anterior

#### Interpretação visual

Como o domínio e de agendamentos por horário, a representação pode seguir uma destas abordagens:

1. densidade/volume de agendamentos por faixa horaria ao longo dos dias
2. media de ocupação por horário
3. total de agendamentos distribuídos ao longo das horas operacionais

Recomendação desta PR:

- validar qual modelo gera melhor leitura para o usuário `admin` sem distorcer o dado;
- se necessário, iniciar com um MVP de serie temporal agregada por dia e evoluir a granularidade horaria em PR seguinte.

### 3.4 Lista de agendamentos em andamento

Abaixo do gráfico, deve existir o bloco:

- `Agendamentos em andamento`

A lista deve considerar o turno atual:

- manha: agendamentos ate `12:00`
- tarde: agendamentos apos `12:00`

Cada item deve exibir:

- nome do pet em atendimento
- status atual do atendimento
- tempo de atendimento em horas

Regra de calculo do tempo:

- para `in_progress`, `waiting`, `confirmed`:
  - `horário_atual - schedules.scheduled_at`
- para `finished`, `delivered`:
  - `schedule_status_history.changed_at - schedules.scheduled_at`

Observação:

- para os estados finalizados, deve-se usar o instante relevante da mudança para o status final considerado;
- se existir mais de um evento em `schedule_status_history`, a query precisa escolher corretamente o ultimo timestamp do status alvo.

## 4. Sessão de Chat para `admin`

Na coluna da direita deve existir uma area de chat textual inspirada na referencia visual.

### Participantes

Para o usuário logado do tipo `admin`, o chat sera com usuários do tipo `system`.

### Restrições

- somente mensagens de texto
- sem chamada Telefonica
- sem videochamada

### Cabeçalho do chat

Exibir para o usuário logado:

- imagem do usuário em `people_identifications.image_url`
- estado visual:
  - ocupado
  - ativo
  - offline
- nome curto:
  - `people_identifications.short_name`

Nao exibir `@username` neste momento, pois o domínio atual não tem esse identificador confirmado.

### Seletor de usuário

Abaixo, deve existir um `select` com os usuários do sistema vinculados ao tenant.

Quando um usuário for selecionado:

- o chat textual correspondente deve ser aberto.

### Dependencies de produto

Esta parte depende de validação forte de domínio, porque o sistema atual pode ainda não possuir:

- modelagem completa de conversa;
- tabela de mensagens;
- presença em tempo real;
- status online/offline/ocupado;
- canal entre `admin` e `system`.

Portanto, a PR pode precisar dividir o bloco em:

- UI completa com dados mockados/placeholder controlado
- contrato e backend em etapa posterior

ou

- MVP funcional com persistência simples de mensagens, se o domínio puder ser introduzido agora sem estourar o escopo.

## Princípios de Execução

- Preservar o contexto multi-tenant em todos os blocos.
- Tratar o novo layout como shell reutilizável, não como pagina isolada.
- Priorizar legibilidade de negocio sobre fidelidade visual cega.
- Manter o dashboard `admin` baseado em dados reais, não em números estáticos.
- Evitar acoplamento entre componentes visuais e queries SQL/REST brutas.
- Entregar de forma incremental, com fases pequenas e checks verificáveis.

## Dependencies de Dados e Contratos

Antes da implementação total, esta PR precisa validar ou explicitar:

- como obter o plano corrente do tenant de forma canônica;
- como mapear o proximo plano sugerido;
- como obter `people_identifications.short_name` do usuário logado no Web;
- como obter `companies.logo_url` com fallback seguro;
- como calcular a meta mensal a partir de `company_system_configs.schedule_days`;
- como identificar status atual de cada agendamento;
- como derivar comparativos diários e mensais com performance aceitável;
- se já existe backend suficiente para o chat ou se sera necessário novo vertical de domínio.

## Fase 0 - Descoberta Técnica e Fechamento de Contrato

### 0.1 Ações

- Mapear no backend e nos contratos compartilhados:
  - plano atual da empresa;
  - logo da empresa;
  - `short_name` do usuário autenticado;
  - dados necessários para os KPIs;
  - dados necessários para o gráfico;
  - dados necessários para a lista de agendamentos em andamento.
- Confirmar quais tipos de usuário existem de fato no contrato atual.
- Verificar se o tipo `admin` já chega de forma confiável ao frontend.
- Auditar o estado atual do suporte a chat e presença.

### 0.2 Checks

- [x] Existe fonte canônica para o plano atual do tenant.
- [x] O frontend consegue resolver `companies.logo_url` e `people_identifications.short_name`.
- [x] As formulas de KPI estão fechadas com base na modelagem real.
- [x] O escopo do chat esta classificado como `UI + contrato futuro`.

### 0.3 Conclusões da descoberta

#### Fonte canônica de plano atual e branding do tenant

- A fonte canônica atual do tenant no Web e `GET /api/v1/companies/current`.
- Esse endpoint já retorna `active_package` e `logo_url` a partir de `companies`.
- O backend resolve isso via `CompanyService.GetCurrentCompany` + query `GetCompanyByID`.
- Durante a Fase 0, os contratos compartilhados foram alinhados para refletir corretamente:
  - `logo_url` em `CompanyDTO`
  - `module_package` com suporte a `trial`

Conclusão:

- o shell novo pode usar `companies/current` como fonte inicial de plano e logo;
- o CTA de upgrade pode ser derivado inteiramente do `active_package` atual;
- tenants `premium` e `internal` exigem tratamento de UX especifico.

#### Tipos de usuário confirmados no domínio atual

No schema, os papeis sistêmicos reais sao:

- `root`
- `internal`
- `admin`
- `system`
- `common`
- `free`

No contexto de vinculo empresa/usuário (`company_users.kind`), os tipos reais sao:

- `owner`
- `employee`
- `client`
- `supplier`
- `outsourced_employee`

Durante a Fase 0, os contratos compartilhados do frontend foram corrigidos para refletir esses enums reais.

Conclusão:

- a home por tipo de usuário deve se basear em `users.role`;
- o `kind` atual continua util para contexto de negocio, mas nao substitui a decisão por papel sistêmico.

#### Estado do `admin` no frontend

- O login atual entrega ao Web:
  - `access_token`
  - `user_id`
  - `company_id`
  - `role`
  - `kind`
- O `AuthStore` persiste esse contexto de sessão com confiabilidade.

Conclusão:

- o tipo `admin` já chega de forma confiável ao frontend;
- a seleção da home inicial por role pode ser iniciada sem mudança no fluxo de login.

#### Estado atual de `short_name` do usuário autenticado

- O schema possui `people_identifications.short_name`.
- Existem queries e tabelas que manipulam `people_identifications`.
- Durante a Fase 0 foi introduzido o endpoint autenticado `GET /api/v1/users/me`, que expõe:
  - `user_id`
  - `company_id`
  - `person_id`
  - `role`
  - `kind`
  - `full_name`
  - `short_name`
  - `image_url`
- O Web passou a ter contrato, client e query para resolver esse perfil autenticado de forma canônica.

Conclusão:

- `companies.logo_url` já esta resolvível no Web;
- `people_identifications.short_name` agora também esta resolvível para o usuário autenticado;
- a Fase 1 pode consumir esse dado sem acoplar o layout ao payload de login.

#### Estado atual de `company_system_configs`

- O schema contem:
  - `schedule_init_time`
  - `schedule_end_time`
  - `min_schedules_per_day`
  - `schedule_days`
- Ja existe query SQLC `GetCompanySystemConfig`.
- Ainda nao existe exposição clara desse contrato para o Web autenticado atual.

Conclusão:

- a modelagem minima para os KPIs existe;
- a PR precisara expor `company_system_configs` ao frontend, seja por endpoint dedicado, seja por um endpoint agregador do dashboard `admin`.

#### Formulas de KPI fechadas com a modelagem atual

Decisão técnica da Fase 0:

- `Agendamentos/dia`:
  - contar `schedules` do tenant no dia atual
  - comparar com a contagem do dia anterior
- `Agendamentos/mes`:
  - contar `schedules` do tenant no mes atual
  - comparar com o mes anterior
- `Eficiência`:
  - interpretar `schedule_days` como dias da semana permitidos de atendimento
  - calcular quantos dias do mes atual pertencem a esse conjunto
  - multiplicar pelo `min_schedules_per_day`
  - usar isso como meta mensal minima
  - percentual = `agendamentos_do_mes / meta_mensal_minima * 100`

Conclusão:

- a- a formula de eficiência esta fechada no nível de domínio;
- o que falta nao e definição matemática, e sim exposição de dados e agregações para consumo do Web.

#### Status atual dos agendamentos e lista operacional

- O status atual de um `schedule` já pode ser derivado pelo ultimo evento de `schedule_status_history`.
- As queries de `schedules` já usam essa estrategia para preencher `current_status`.
- Ja existe `GetLatestScheduleStatus` e histórico completo por schedule.

Conclusão:

- a lista de "Agendamentos em andamento" e viável com o domínio atual;
- a duração por status pode ser calculada sem alterar o schema;
- sera necessário apenas definir query/DTO especifico para o dashboard.

#### Gráfico de performance

- O domínio já oferece janela operacional da empresa via `company_system_configs`.
- Ainda nao existe query pronta agregando ocupação ou distribuição por semana comparando mes atual vs mes anterior.

Conclusão:

- a PR deve criar agregação dedicada para o dashboard `admin`;
- a recomendação técnica e concentrar esse calculo no backend, nao no frontend, para evitar múltiplas consultas e regras de calendário espalhadas.

#### Chat e presença

- Nao foram encontrados no schema atual:
  - tabelas de conversas;
  - mensagens de chat;
  - presença online/offline/ocupado;
  - canal persistido entre `admin` e `system`.
- O único campo relacionado e `company_system_configs.whatsapp_conversation`, que nao resolve chat interno entre usuários do produto.

Conclusão:

- o chat desta PR deve ser tratado como `UI + contrato futuro`;
- nao ha base suficiente hoje para prometer chat funcional completo sem abrir um novo vertical de domínio;
- se o bloco visual for implementado nesta PR, ele deve deixar isso explicito no escopo.

## Fase 1 - Novo Shell de Layout Autenticado

Status atual:

- Em andamento, com shell base já implementado no Web.
- O layout autenticado agora usa:
  - branding do tenant com `companies.logo_url` e fallback por iniciais;
  - bloco de identificação do usuário com `users/me`;
  - navegação principal enxuta com `Dashboard`, `Agendamentos`, `Clientes` e `Pets`;
  - CTA de upgrade orientado por `active_package`;
  - footer lateral com `Configurações` e `Sair`;
  - suporte a desktop recolhido/expandido e drawer mobile.
- Também foi criada a rota placeholder `/:companySlug/settings` para ancorar a nova ação de configurações no shell.

### 1.1 Ações

- Refatorar o layout autenticado atual para refletir a nova estrutura visual conforme [example-001](../../example-001.png).
- Reorganizar a sidebar:
  - branding do tenant no topo;
  - navegação principal;
  - bloco de upgrade;
  - footer com configurações e logout.
- Preservar responsividade desktop, tablet e mobile.
- Garantir que estados expandido/recolhido continuem consistentes.

### 1.2 Checks

- [x] A area autenticada usa o novo shell [visual](../../example-001.png).
- [x] A logo do tenant e exibida com fallback coerente.
- [x] O CTA de upgrade respeita a hierarquia de planos.
- [x] O footer lateral contem `Configurações` e `Sair`.
- [x] Existe rota inicial de `Configurações` para sustentar a navegação do shell.

## Fase 2 - Home Inicial por Tipo de Usuário

Status atual:

- Em andamento.
- O dashboard principal já assume o recorte de `admin` como primeira home rica.
- Perfis diferentes de `admin` agora recebem um placeholder explicito no `DashboardPage`, evitando tratar todas as roles como experiencia final idêntica.

### 2.1 Ações

- Definir estrategia de roteamento/renderização da home inicial por tipo de usuário.
- Implementar o branch inicial para `admin`.
- Registrar comportamento esperado para:
  - `root`
  - `internal`
  - `system`
  - `common`
  - `free`

### 2.2 Checks

- [x] O Web não trata mais a home autenticada como única para todos.
- [x] O tipo `admin` cai na home/dash correta.
- [ ] Os demais tipos possuem direção registrada, mesmo que ainda não implementados.

## Fase 3 - Header e KPIs do Dashboard `admin`

Status atual:

- Primeiro recorte implementado.
- O dashboard `admin` agora consome:
  - `companies/current`
  - `users/me`
  - `company-system-configs/current`
  - `schedules`
- O topo já exibe saudação com `short_name`, texto contextual e data.
- Os tres cards principais já estão calculados com dados reais carregados no frontend.

### 3.1 Ações

- Implementar o topo do dashboard com saudação, texto contextual e data formatada.
- Implementar os tres cards:
  - Agendamentos/dia
  - Agendamentos/mês
  - Eficiência
- Criar queries e agregações necessárias no backend/frontend.

### 3.2 Checks

- [x] O topo exibe `short_name` real do usuário.
- [x] A data atual e exibida no formato definido.
- [x] Os tres cards refletem dados reais do tenant.
- [x] Os indicadores positivo/negativo respeitam o comparativo definido.

## Fase 4 - Gráfico de Performance

Status atual:

- Recorte visual ampliado e mais aderente ao plano.
- O seletor semanal do mês corrente continua funcional no dashboard.
- O gráfico agora usa a janela operacional do tenant (`schedule_init_time` e `schedule_end_time`) para orientar o eixo Y.
- A leitura atual representa a ocupação média por horário dentro da semana selecionada, comparando mês atual e mês anterior.

### 4.1 Ações

- Implementar bloco [visual](../../example-001.png) do gráfico.
- Construir o seletor semanal.
- Trazer dados do mês atual e do mês anterior.
- Ajustar eixo temporal conforme janela operacional do tenant.

### 4.2 Checks

- [x] O gráfico e alimentado por dados reais.
- [x] O seletor semanal funciona.
- [x] O comparativo com o mês anterior esta correto.
- [x] A leitura do gráfico permanece clara em desktop e tablet.

## Fase 5 - Lista de Agendamentos em Andamento

Status atual:

- Recorte expandido implementado.
- A lista já respeita o turno atual e agora inclui agendamentos ativos e concluídos do turno.
- O calculo de duração usa `schedule_status_history` para `finished` e `delivered`, preservando a regra `changed_at - scheduled_at`.
- A distinção visual por status agora esta refletida na lista operacional.

### 5.1 Ações

- Implementar lista operacional por turno.
- Resolver status atual dos agendamentos.
- Calcular duração conforme regra por status.
- Exibir pet, status e duração.

### 5.2 Checks

- [x] O turno atual e respeitado.
- [x] O tempo em atendimento e calculado corretamente.
- [x] A lista distingue estados em andamento e finalizados.

## Fase 6 - Chat do `admin` com Usuários `system`

Status atual:

- Conversa persistida implementada entre `admin` e `system` no contexto do tenant.
- O seletor visual e o cabeçalho do contato seguem ocupando a coluna direita.
- O dashboard consome usuários reais do tenant via `company-users`, filtrando contatos com `role = system`.
- O seed local passou a incluir um usuário `system` vinculado ao tenant e histórico inicial de mensagens para alimentar a experiência.
- O recorte funcional desta PR ficou definido como:
  - mensagens de texto persistidas;
  - sem voz/video;
  - sem presença em tempo real.
- Ainda falta, para uma etapa futura, presença online/offline dinâmica, indicadores de ocupado e sincronização em tempo real.

### 6.1 Ações

- Implementar a UI da coluna lateral direita alinhada a referencia.
- Adicionar cabeçalho do usuário selecionado com avatar, nome curto e status.
- Adicionar seletor de usuários `system`.
- Definir e implementar o recorte técnico:
  - MVP visual;
  - MVP funcional com mensagens persistidas;
  - ou fatiamento para PR posterior.

### 6.2 Checks

- [x] Existe seletor de usuários do chat.
- [x] O contato selecionado exibe avatar e nome curto.
- [x] O escopo funcional do chat ficou explicitamente fechado.
- [x] Nao ha promessa de voz/video fora do recorte.
- [x] O histórico de mensagens entre `admin` e `system` fica persistido.
- [x] O dashboard consegue enviar novas mensagens de texto para o contato selecionado.

## Fase 7 - Testes, Observabilidade e Documentação

Status atual:

- O shell autenticado e o dashboard `admin` já possuem cobertura relevante de layout, navegação e integração.
- Os cenários de fallback agora incluem:
  - tenant sem logo;
  - usuário sem imagem;
  - tenant em `premium`;
  - tenant com configuração operacional mínima e sem dados suficientes para KPI completo.
- A documentação operacional já descreve as credenciais seedadas e o uso da home `admin` no ambiente local.

### 7.1 Ações

- Atualizar testes de layout, home e navegação.
- Adicionar testes para cards e comparativos.
- Cobrir cenários de fallback:
  - tenant sem logo;
  - usuário sem imagem;
  - tenant em `premium`;
  - tenant sem dados suficientes para KPI completo.
- Atualizar README ou docs operacionais se a home autenticada mudar substancialmente.

### 7.2 Checks

- [x] O novo layout possui cobertura de testes relevante.
- [x] O dashboard `admin` não quebra sem dados completos.
- [x] A documentação explica o novo comportamento da home.

## Riscos e Cuidados

- O chat pode extrapolar o escopo se exigir domínio novo completo.
- O calculo de eficiência depende de interpretação correta de `schedule_days`.
- O gráfico pode ficar visualmente bonito mas semanticamente fraco se a agregação não for bem definida.
- O plano corrente do tenant precisa ter fonte única e confiável.
- A home por tipo de usuário pode exigir reestruturação do roteamento autenticado.
- A referencia [visual](../../example-001.png) usa um design muito limpo; a adaptação deve evitar "slop" e preservar o contexto real do produto.

## Decisões Recomendadas para Manter a PR Saudável

- Implementar nesta PR apenas a home `admin`.
- Tratar o chat como escopo condicional:
  - funcional, se o domínio já existir ou puder ser introduzido sem risco alto;
  - visual/estrutural, se o domínio ainda não estiver pronto.
- Fechar primeiro contrato e metrics antes de estilizar o gráfico.
- Reaproveitar o shell novo para futuras homes de `root`, `internal`, `system`, `common` e `free`.

## Ordem Recomendada de Execução

1. Fase 0: descoberta técnica e fechamento de contrato.
2. Fase 1: novo shell de layout autenticado.
3. Fase 2: home inicial por tipo de usuário.
4. Fase 3: header e KPIs do dashboard `admin`.
5. Fase 4: gráfico de performance.
6. Fase 5: lista de agendamentos em andamento.
7. Fase 6: chat do `admin` com usuários `system`.
8. Fase 7: testes, observabilidade e documentação.

## Resultado Esperado

Se este plano for executado com sucesso, o PetControl dará um salto de percepção de produto:

- a area autenticada deixara de parecer um painel técnico genérico;
- o tenant passara a ter branding explicito no shell;
- a home do `admin` passara a refletir operação real, contexto diário e performance;
- o produto ganhara uma base visual e estrutural mais forte para evoluir experiencias especificas por tipo de usuário.
