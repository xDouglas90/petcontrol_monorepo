# Plano de Ação e Execução - Rotas Web com `company_slug`

## Objetivo

Definir a evolução da navegação autenticada do PetControl para usar URLs com contexto explicito de tenant no frontend Web.

O objetivo principal deste plano e sair do modelo atual de rotas planas:

- `/login`
- `/dashboard`
- `/schedules`

e adotar um modelo em que a area autenticada carregue o `company_slug` antes de todos os recursos principais, por exemplo:

- `/company-x/dashboard`
- `/company-x/schedules`

Essa mudança deve melhorar clareza de contexto, deep link, previsibilidade de navegação e preparação para cenários multi-tenant mais explícitos, sem transferir para a URL a responsabilidade de autorização, que continua pertencendo ao backend via JWT e `company_id`.

## Contexto Atual Observado

- O backend ja autentica o usuário e emite JWT com contexto de tenant via `company_id`.
- O endpoint de empresa corrente ja existe e retorna dados da empresa autenticada, incluindo `slug`.
- O frontend Web ja possui fluxo de login real, dashboard conectada e modulo `schedules`.
- As rotas autenticadas atuais do Web ainda são planas e não carregam contexto explicito da empresa na URL.
- O frontend hoje usa a sessão persistida para auth e Query para dados de servidor, o que facilita derivar o `company_slug` da empresa corrente sem duplicar estado desnecessariamente.

## Princípios de Execução

- O `company_slug` na URL representa contexto de navegação, não autorização.
- O backend continua confiando exclusivamente no JWT e no `company_id` do token para isolar tenant.
- O frontend deve corrigir URLs inconsistentes, redirecionando para o slug real da empresa autenticada.
- O login deve continuar em rota sem slug, para evitar acoplamento artificial antes da sessão existir.
- A implementação deve ser incremental, com compatibilidade e redirecionamentos onde fizer sentido.
- Cada fase deve terminar com checks objetivos de comportamento, navegação ou testes.

## Estado Alvo

Ao final deste plano, o comportamento desejado e:

- usuário acessa `/login`;
- autentica com sucesso;
- o Web resolve a empresa corrente autenticada;
- o browser navega para `/:companySlug/dashboard`;
- todos os links internos da area autenticada preservam o prefixo `/:companySlug`;
- se o usuário acessar um slug incorreto manualmente, o frontend corrige para o slug verdadeiro da sessão;
- nenhum acesso do backend depende do slug enviado na URL do frontend.

## Riscos e Cuidados

- Usar o slug como fonte de verdade de permissão seria um erro de segurança.
- O slug pode mudar no futuro; se isso for permitido, links antigos exigirão estrategia de redirecionamento ou tolerância controlada.
- Rotas e query keys precisam evitar acoplamento desnecessária a estado duplicado.
- Testes de navegação devem cobrir redirect, mismatch de slug e carregamento inicial com sessão persistida.
- Se o produto vier a suportar múltiplas empresas por usuário na mesma sessão, a estrategia atual precisara ser expandida com seletor explicito de tenant.

## Fase 0 - Alinhamento de Contrato e Direção

### 0.1 - Ações

- Confirmar que o `slug` da empresa corrente ja esta disponível no contrato consumido pelo Web.
- Revisar se `shared-types` e `shared-constants` precisam refletir essa nova estrategia de roteamento.
- Definir a convenção oficial das rotas autenticadas:
  - `/:companySlug/dashboard`;
  - `/:companySlug/schedules`.
- Definir que `/login` permanece sem slug.
- Registrar explicitamente que o slug e contexto de UX e URL, não mecanismo de autorização.

### 0.2 - Checks

- [x] O endpoint de empresa corrente retorna `slug` de forma estável.
- [x] A documentação interna deixa claro que autorização continua baseada em JWT e `company_id`.
- [x] Existe uma convenção única de rota para os módulos autenticados.

Observação: a Fase 0 foi concluída com a validação explícita do `slug` no contrato da empresa corrente, formalização da convenção futura de rotas com `company_slug` em `shared-constants` e documentação separando claramente contexto de navegação de autorização de tenant.

## Fase 1 - Base de Roteamento com Prefixo de Tenant

### 1.1 - Ações

- Refatorar o router do Web para agrupar a area autenticada sob `/:companySlug`.
- Atualizar `APP_ROUTES` e segmentos compartilhados para refletir o novo formato.
- Adaptar `AppLayout` e links internos para navegar sempre com o slug atual.
- Garantir que a home redirecione:
  - para `/login` quando não houver sessão;
  - para `/:companySlug/dashboard` quando houver sessão valida e empresa resolvida.

### 1.2 - Checks

- [x] `dashboard` passa a responder em `/:companySlug/dashboard`.
- [x] `schedules` passa a responder em `/:companySlug/schedules`.
- [x] Links internos da sidebar preservam o prefixo `/:companySlug`.
- [x] Home redirect respeita o slug da empresa autenticada.

Observação: a Fase 1 foi concluída com a migração do router da área autenticada para `/:companySlug`, atualização das rotas compartilhadas para o novo formato, adaptação do `AppLayout` para preservar o slug nos links internos e redirecionamento da home para o dashboard da empresa corrente autenticada.

## Fase 2 - Pos-Login e Hidratação de Sessão

### 2.1 - Ações

- Ajustar o fluxo de login para redirecionar para `/:companySlug/dashboard` apos autenticação bem-sucedida.
- Garantir que o slug seja obtido de forma confiável:
  - pela query de empresa corrente; ou
  - por payload ja disponível no frontend, se isso existir sem duplicação indevida.
- Definir comportamento de carregamento enquanto sessão e empresa corrente ainda estão sendo resolvidas.
- Evitar salvar no Zustand dados de servidor que podem continuar em Query, exceto se houver justificativa clara de sessão.

### 2.2 - Checks

- [x] Login com sucesso leva para `/:companySlug/dashboard`.
- [x] Reload da pagina preserva navegação correta quando a sessão esta persistida.
- [x] O frontend não depende de mock manual do slug para navegar apos login.
- [x] Não ha espelhamento desnecessário do objeto `company` fora do Query.

Observação: a Fase 2 foi concluída com o ajuste do fluxo de rotas pós-auth e a garantia de que o slug é derivado corretamente da query de empresa corrente sem duplicação de estado.

## Fase 3 - Guardas, Correção de URL e Robustez

### 3.1 - Ações

- Implementar validação no layout autenticado para comparar:
  - slug da rota;
  - slug real da empresa corrente.
- Se o slug não corresponder, redirecionar automaticamente para o slug correto.
- Definir comportamento para casos de erro:
  - sessão invalida;
  - empresa corrente indisponível;
  - slug ausente ou malformado.
- Garantir que logout remova a sessão e redirecione para `/login`.

### 3.2 - Checks

- [x] Acesso com slug incorreto redireciona para o slug verdadeiro da empresa.
- [x] Rota autenticada sem sessão continua redirecionando para `/login`.
- [x] Logout remove o contexto autenticado e sai da area com slug.
- [x] Erros de carregamento da empresa não causam loop infinito de navegação.

Observação: a Fase 3 foi concluída com guardas robustas no `AppLayout`, incluindo correção automática de slug de forma case-insensitive, redirecionamento para `/login` sem sessão, limpeza da sessão em contexto inválido (`401`) e cobertura via Vitest para mismatch de slug, logout e saída segura da área autenticada.

## Fase 4 - Cobertura de Testes e Consistencia

### 4.1 - Ações

- Atualizar testes do router, login, layout e paginas autenticadas para o novo formato com slug.
- Adicionar testes de navegação para:
  - redirect da home;
  - redirect pos-login;
  - correção de slug inconsistente;
  - links internos preservando o slug.
- Revisar documentação de onboarding e README onde houver exemplos de URLs antigas.
- Revisar mocks, fixtures e utilitários de teste que assumem rotas planas.

### 4.2 - Checks

- [x] Testes do Web cobrem redirect pos-login e rotas com slug.
- [x] Testes cobrem mismatch entre slug da URL e slug real da empresa.
- [x] README e docs não continuam exibindo apenas rotas antigas planas.
- [x] Navegação com slug não quebra dashboard nem `schedules`.

Observação: a Fase 4 foi concluída com consolidação da suíte do Web para o fluxo com `companySlug`, incluindo integração do router para redirect da home, acesso a `/:companySlug/schedules`, preservação do slug nos links internos e revisão documental das rotas autenticadas no README e nas convenções.

## Fase 5 - Endurecimento de UX e Evolução Futura

### 5.1 - Ações

- Avaliar exibição do `company_slug` e/ou nome da empresa no header como confirmação visual de contexto.
- Revisar se a URL deve ser tratada como canonicamente minuscula e normalizada.
- Definir estrategia futura caso a empresa possa alterar o slug:
  - redirect pelo backend;
  - redirect pelo frontend;
  - invalidação e resolução por empresa corrente.
- Registrar a decisão arquitetural para reaproveitamento futuro em mobile, links compartilhados e white-label.

### 5.2 - Checks

- [ ] A URL final da area autenticada e canônica e previsível.
- [ ] Existe direção documentada para futuras mudanças de slug.
- [ ] A experiencia de navegação deixa explicito o tenant atual sem depender apenas do estado interno da aplicação.

## Ordem Recomendada de Execução

1. Fase 0: alinhamento de contrato e direção.
2. Fase 1: base de roteamento com prefixo de tenant.
3. Fase 2: pos-login e hidratação de sessão.
4. Fase 3: guardas, correção de URL e robustez.
5. Fase 4: cobertura de testes e consistencia.
6. Fase 5: endurecimento de UX e evolução futura.

## Resultado Esperado

Se este plano for executado com sucesso, o frontend do PetControl passara a expressar o tenant também na URL, com rotas mais semânticas e compartilháveis, sem degradar isolamento multi-tenant, sem duplicar responsabilidade do backend e com uma base melhor para escalar a navegação do produto.
