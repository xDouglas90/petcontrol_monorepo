# Plano de Ação e Execução - Tela de Pets do Tenant

## Objetivo

Definir o escopo técnico, funcional e arquitetural da próxima PR do PetControl para evoluir a tela `Pets` em `/:companySlug/pets`, aproximando sua experiência da tela `Pessoas`, porém com foco em operação animal, filtros ricos de listagem e painel lateral direito contextual para visualização, criação e edição.

Esta PR deve introduzir:

- evolução da tela `/:companySlug/pets` para o padrão `lista + aside direito`;
- filtros de listagem por:
  - `pet_size`
  - `pet_temperament`
  - `pet_kind`
  - `pets.race`
  - `pets.is_active`
- item de listagem com imagem grande e composição visual orientada ao pet;
- painel lateral direito para:
  - visualização do pet selecionado;
  - exibição de tutor e guardião(ões);
  - formulário de criação;
  - formulário de edição;
- botão adicional de `+` no menu lateral para abrir diretamente o fluxo de criação de pet;
- formulário no aside direito com seleção inicial do tutor e suporte a imagem com preview;
- fallback automático de imagem usando `GCS_PUBLIC_URL/assets/images/{kind}-default-image.png` quando o pet não tiver foto própria.

## Contexto Atual

- A rota `/:companySlug/pets` já existe no Web em [apps/web/src/routes/(app)/pets/index.tsx](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/apps/web/src/routes/(app)/pets/index.tsx:1).
- A tela atual de pets já suporta listagem simples, busca textual, criação, edição e exclusão lógica, mas ainda sem composição semelhante a `/people`.
- O router autenticado já expõe `pets` em [apps/web/src/router.tsx](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/apps/web/src/router.tsx:1).
- O shell autenticado já possui menu lateral e aside direito em [apps/web/src/routes/(app)/_layout.tsx](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/apps/web/src/routes/(app)/_layout.tsx:1).
- O backend já expõe `GET /pets`, `GET /pets/{id}`, `POST /pets`, `PUT /pets/{id}` e `DELETE /pets/{id}` em [apps/api/internal/handler/pet_handler.go](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/apps/api/internal/handler/pet_handler.go:1).
- O schema atual de `pets` já contém mais campos do que o frontend usa hoje, incluindo `race`, `color`, `sex`, `is_deceased`, `is_vaccinated`, `is_neutered`, `is_microchipped`, `microchip_number` e `microchip_expiration_date` em [schema.sql](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/schema.sql:864).
- O contrato compartilhado de pets ainda está reduzido e precisará ser expandido em [libs/shared-types/src/index.ts](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/libs/shared-types/src/index.ts:714).
- Já existem queries para `pet_guardians`, incluindo listagem tenant-scoped dos pets por guardião, em [infra/sql/queries/pet_guardians.sql](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/infra/sql/queries/pet_guardians.sql:1).

## Referências Obrigatórias

- Plano de pessoas: [0009-PEOPLE-SCREEN-AND-TENANT-PERSON-REGISTRY-PLAN.md](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/docs/plans/0009-PEOPLE-SCREEN-AND-TENANT-PERSON-REGISTRY-PLAN.md:1)
- Tela atual de pets: [apps/web/src/routes/(app)/pets/index.tsx](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/apps/web/src/routes/(app)/pets/index.tsx:1)
- Layout autenticado: [apps/web/src/routes/(app)/_layout.tsx](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/apps/web/src/routes/(app)/_layout.tsx:1)
- Router Web: [apps/web/src/router.tsx](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/apps/web/src/router.tsx:1)
- Shared types: [libs/shared-types/src/index.ts](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/libs/shared-types/src/index.ts:689)
- Handler de pets: [apps/api/internal/handler/pet_handler.go](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/apps/api/internal/handler/pet_handler.go:1)
- Queries SQLC de pets: [infra/sql/queries/pets.sql](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/infra/sql/queries/pets.sql:1)
- Queries SQLC de guardiões: [infra/sql/queries/pet_guardians.sql](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/infra/sql/queries/pet_guardians.sql:1)
- Plano de uploads GCS: [0005-GCS_DIRECT_UPLOAD_ASSET_STORAGE_PLAN.md](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/docs/plans/0005-GCS_DIRECT_UPLOAD_ASSET_STORAGE_PLAN.md:1)

## Resultado Esperado da PR

Ao final desta PR, o sistema deve apresentar:

- uma tela `Pets` com listagem visualmente mais rica, tenant-scoped e filtrável;
- um aside direito próprio do módulo, em vez do padrão atual de formulário acoplado apenas à coluna da página;
- visualização consolidada do pet selecionado, do tutor e do guardião(ou guardiões);
- criação e edição completas com todos os campos relevantes da tabela `pets`;
- escolha obrigatória do tutor antes do preenchimento completo do formulário;
- seleção opcional de guardião;
- preview da imagem do pet e fallback consistente de imagem default por `kind`;
- atalho `+` na navegação lateral para abrir o formulário de novo pet.

## Escopo Funcional Confirmado

## 1. Rota, Navegação e Shell

Esta PR não cria uma rota nova; ela evolui a rota autenticada já existente `/:companySlug/pets`.

Também deve:

- manter o item `Pets` no menu lateral;
- adicionar um botão visual de `+` associado ao item de navegação de pets;
- fazer esse botão abrir diretamente o modo de criação na própria tela `/pets`;
- adotar para `/pets` a mesma linguagem estrutural de `/people`, com lista principal e aside direito contextual;
- garantir boa responsividade em desktop e mobile.

Decisão recomendada para navegação:

- clicar em `Pets` abre a listagem no estado padrão;
- clicar no `+` abre `/pets` no estado de criação;
- o estado de criação pode ser controlado por search param, para permitir navegação direta e testes mais estáveis.

## 2. Fonte de Verdade dos Pets do Tenant

A listagem principal deve continuar tenant-scoped.

Base relacional principal:

- `pets`;
- `clients` como tutor do pet via `pets.owner_id`;
- `company_clients` para garantir escopo por tenant;
- `people_identifications` para o nome do tutor;
- `pet_guardians` para o guardião associado;
- `company_people` e `people_identifications` para resolver os dados do guardião.

A listagem deve continuar partindo de pets cujo tutor pertença ao tenant corrente.

## 3. Campos Esperados na Listagem

Cada item da lista deve ser composto por:

- foto do pet em `image_url`, ocupando toda a altura útil do item e alinhada à esquerda;
- ao lado da foto, no topo:
  - nome do pet;
  - raça entre parênteses, em cinza claro e fonte menor;
- no topo à direita:
  - status `is_active`;
  - ao lado, quando aplicável, indicador de `is_deceased`;
- abaixo do nome, exibir os demais campos relevantes de `pets`, exceto:
  - `microchip_number`
  - `microchip_expiration_date`
  - `notes`
  - `created_at`
  - `updated_at`
  - `deleted_at`

Para a primeira versão da UI, isso implica exibir no item:

- tutor;
- `kind`;
- `size`;
- `temperament`;
- `color`;
- `sex`;
- `birth_date`;
- `is_vaccinated`;
- `is_neutered`;
- `is_microchipped`.

## 4. Filtros da Listagem

Devem existir filtros explícitos, combináveis, para:

- `pet_size`
- `pet_temperament`
- `pet_kind`
- `pets.race`
- `pets.is_active`

Comportamento esperado:

- os filtros devem refletir no estado da URL ou em estado persistente de navegação;
- a busca textual por nome do pet e tutor pode continuar existindo como complemento;
- `race` deve usar a coluna real de `pets.race`, com opções derivadas da base tenant-scoped;
- `is_active` deve ser exposto como filtro binário claro, por exemplo:
  - `todos`
  - `ativos`
  - `inativos`

Decisão recomendada para backend:

- evoluir `GET /pets` para aceitar filtros estruturados;
- evitar que o frontend faça filtro client-side sobre paginação já recortada.

## 5. Aside Direito do Módulo

O aside direito deve se tornar a área de contexto de `/pets`.

Estados do aside:

- vazio;
- detalhe de pet selecionado;
- criação;
- edição.

Conteúdo do detalhe:

- informações gerais do pet;
- dados do tutor;
- dados do guardião ou da lista de guardiões;
- preview de imagem;
- ações de editar e desativar quando permitido.

Conteúdo do formulário:

- criação de novo pet no próprio aside;
- edição de pet existente no mesmo container;
- possibilidade de trocar tutor;
- possibilidade de definir ou remover guardião.

## 6. Tutor e Guardiões

Na criação, o primeiro campo deve ser um `select` de clientes do tenant.

Depois da seleção do tutor, o restante do formulário do pet fica disponível.

Além disso, deve existir o check opcional:

- `Inserir guardião`

Quando ativado, a UI deve exibir um `select` com pessoas do tenant.

Importante:

- o requisito funcional do usuário usa o termo `guardians`, no plural;
- porém o schema atual de `pet_guardians` possui `pet_id` como chave primária, o que suporta apenas um guardião por pet.

Decisão necessária nesta PR:

1. opção recomendada:
   evoluir a modelagem para permitir múltiplos guardiões por pet, removendo a restrição de PK em `pet_id` e adotando chave composta ou id próprio;
2. opção mínima:
   manter apenas um guardião por pet nesta PR, mas tratar o bloco de UI e documentação como `guardião principal`.

Recomendação do plano:

- se o objetivo do produto já é plural de fato, ajustar o schema agora é melhor do que congelar uma limitação artificial na UX.

## 7. Formulário de Pets

O formulário deverá suportar todos os campos relevantes de `pets`, com ênfase visual nos obrigatórios.

Campos obrigatórios esperados:

- tutor (`owner_id`)
- nome (`name`)
- raça (`race`)
- cor (`color`)
- sexo (`sex`)
- porte (`size`)
- tipo (`kind`)
- temperamento (`temperament`)

Campos opcionais esperados:

- `image_url` ou `upload_object_key`
- `birth_date`
- `is_active`
- `is_deceased`
- `is_vaccinated`
- `is_neutered`
- `is_microchipped`
- `microchip_number`
- `microchip_expiration_date`
- `notes`

Regras de UX recomendadas:

- `microchip_number` e `microchip_expiration_date` só devem ganhar destaque quando `is_microchipped = true`;
- `notes` permanece no detalhe e no formulário, mas não aparece na listagem;
- `is_deceased = true` deve conviver com `is_active`, mas o comportamento precisa ser explicitado no backend e na UI.

## 8. Upload e Fallback de Imagem

Na foto do pet, o formulário deve oferecer:

- campo de inserção de imagem;
- preview antes do salvamento;
- reaproveitamento do fluxo de upload direto já usado em pets e pessoas.

Se nenhuma imagem for adicionada, o sistema deve resolver a imagem default em:

`{GCS_PUBLIC_URL}/assets/images/{kind}-default-image.png`

Exemplos:

- `{GCS_PUBLIC_URL}/assets/images/dog-default-image.png`
- `{GCS_PUBLIC_URL}/assets/images/cat-default-image.png`

Decisão recomendada:

- o fallback pode ser resolvido no frontend para exibição imediata;
- e também pode ser normalizado no backend no retorno, se a equipe quiser evitar duplicação de regra entre telas.

## 9. Evoluções Necessárias de Contrato

O contrato compartilhado atual de pets é insuficiente para esta tela.

Precisará ser expandido para incluir, no mínimo:

- `race`
- `color`
- `sex`
- `is_deceased`
- `is_vaccinated`
- `is_neutered`
- `is_microchipped`
- `microchip_number`
- `microchip_expiration_date`
- `created_at`
- `updated_at`
- `deleted_at`
- dados enriquecidos de guardião(ões)

Também será necessário evoluir:

- `CreatePetInput`
- `UpdatePetInput`
- response DTO de detalhe do pet
- response DTO da listagem de pets

## 10. Evoluções Necessárias de Backend

O backend deverá ser adaptado para suportar:

- filtros estruturados no list endpoint;
- leitura enriquecida do detalhe do pet com tutor e guardião(ões);
- criação e edição com todos os campos relevantes de `pets`;
- persistência de guardião(ões);
- fallback de imagem quando não houver `image_url`;
- regras coerentes para `is_active` e `is_deceased`.

Endpoints esperados ao final:

- `GET /pets`
- `GET /pets/:petId`
- `POST /pets`
- `PUT /pets/:petId`
- `DELETE /pets/:petId`

Evoluções recomendadas:

- `GET /pets` deve aceitar filtros por enum e por `race`;
- `GET /pets/:petId` deve retornar payload mais rico que o atual;
- `POST` e `PUT` devem aceitar todos os campos relevantes do domínio.

## 11. Evoluções Necessárias de Web

O frontend deverá:

- migrar a página atual para uma experiência semelhante à de `/people`;
- introduzir barra/área de filtros persistentes;
- exibir lista com foto ampla e conteúdo hierarquizado;
- suportar seleção de item e aside de detalhe;
- suportar abertura direta do aside em modo criação;
- ampliar o formulário para todos os campos necessários;
- integrar seleção de guardião;
- aplicar fallback de imagem por `kind`;
- adicionar o botão `+` no menu lateral.

## 12. Testes Esperados

Backend:

- filtros do list endpoint por `size`, `temperament`, `kind`, `race` e `is_active`;
- detalhe tenant-scoped com tutor e guardião(ões);
- criação com todos os campos obrigatórios e opcionais;
- atualização com alteração de tutor e guardião;
- fallback de imagem quando não houver imagem explícita;
- bloqueio cross-tenant para tutor e guardião;
- cobertura de schema novo caso `pet_guardians` vire relação múltipla.

Frontend:

- rota `/pets` com composição `lista + aside`;
- abertura do formulário pelo botão `+` do menu;
- filtros afetando query e renderização;
- item da lista exibindo foto, raça, status e atributos esperados;
- detalhe do pet com tutor e guardião(ões);
- formulário com obrigatórios destacados;
- preview da imagem;
- fallback visual por `kind` sem imagem;
- criação e edição bem-sucedidas;
- erro de criação/edição refletido no aside.

## 13. Riscos e Decisões em Aberto

- O schema atual de `pet_guardians` conflita com a noção de múltiplos guardiões.
- A API e os shared types ainda não expõem vários campos que a UI nova precisa.
- A tela atual de `/pets` já existe e precisará ser reestruturada sem regredir os fluxos operacionais existentes.
- O fallback de imagem baseado em `GCS_PUBLIC_URL` deve usar uma convenção estável de assets previamente publicada no bucket.

## 14. Fora de Escopo

- gestão de histórico clínico do pet;
- anexos além da foto principal;
- timeline de alterações do pet;
- múltiplas fotos por pet;
- regras de vacinação com calendário detalhado;
- módulo veterinário.

## 15. Checklist de Implementação

- [x] Definir se `pet_guardians` suportará um ou múltiplos guardiões nesta PR.
- [x] Atualizar schema/migration de `pet_guardians` caso a decisão seja suporte múltiplo.
- [x] Expandir queries SQLC de `pets` para filtros estruturados.
- [x] Expandir queries SQLC de detalhe do pet com tutor e guardião(ões).
- [x] Expandir `shared-types` para refletir todos os campos relevantes de `pets`.
- [x] Atualizar Swagger/docs de `pets`.
- [x] Evoluir `PetService` para criação e edição completas.
- [x] Evoluir `PetHandler` para aceitar e validar os novos campos.
- [x] Implementar resolução de fallback de imagem por `kind`.
- [x] Garantir validação tenant-scoped de tutor.
- [x] Garantir validação tenant-scoped de guardião(ões).
- [x] Evoluir o list endpoint para aceitar filtros por `size`, `temperament`, `kind`, `race` e `is_active`.
- [x] Reestruturar a tela `/:companySlug/pets` para padrão `lista + aside`.
- [x] Adicionar filtros visuais persistentes na tela.
- [x] Atualizar a composição visual de cada item da lista.
- [x] Implementar detalhe do pet no aside.
- [x] Exibir dados do tutor no aside.
- [x] Exibir dados do guardião ou lista de guardiões no aside.
- [x] Migrar criação e edição para o aside direito.
- [x] Adicionar seleção inicial obrigatória do tutor.
- [x] Adicionar toggle/check de inserção de guardião.
- [x] Adicionar select tenant-scoped de pessoas para guardião.
- [x] Adicionar preview de imagem no formulário.
- [x] Integrar upload de imagem com o fluxo de GCS já existente.
- [x] Aplicar fallback visual usando `GCS_PUBLIC_URL/assets/images/{kind}-default-image.png`.
- [x] Adicionar botão `+` no menu lateral para abrir o fluxo de criação de pet.
- [x] Adicionar testes de backend para filtros, detalhe e mutações.
- [x] Adicionar testes de frontend para rota, filtros, aside e formulário.

## 16. Sequência Recomendada

1. Fechar a decisão de modelagem de `guardians`.
2. Expandir schema, queries SQLC, contracts e Swagger.
3. Enriquecer backend de `pets` com filtros e detalhe.
4. Reestruturar a página `/pets` no Web.
5. Integrar imagem default e seleção de guardião.
6. Cobrir o fluxo com testes backend e frontend.

## Conclusão

Esta PR transforma `Pets` de uma tela operacional simples em um módulo contextual completo, alinhado ao padrão recente de `People`, mas respeitando as necessidades específicas do domínio animal: filtros por enum, identidade visual centrada na foto, detalhe lateral rico, tutor explícito, guardião(ões) e upload de imagem com fallback por espécie/tipo. O ponto mais importante a fechar antes da implementação é a modelagem de `pet_guardians`, porque ela impacta diretamente contrato, banco, backend e UX.
