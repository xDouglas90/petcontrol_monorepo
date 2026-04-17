# Plano - Upload Direto para Google Cloud Storage com Persistência em Banco

## Objetivo

Definir a arquitetura e o plano de implementação para permitir que o frontend envie arquivos para um bucket do Google Cloud Storage e que o backend persista a referencia desse arquivo nas colunas `_url` do banco.

Stack alvo:

- API em Go
- frontend React/Vite/TanStack
- bucket no Google Cloud Storage
- SDK Go `cloud.google.com/go/storage`

---

## Veredito da ideia

A ideia esta boa e o fluxo base esta correto:

1. o frontend pede ao backend uma URL temporária para upload
2. o backend gera a URL assinada para o bucket
3. o frontend envia o arquivo direto para o GCS
4. o backend recebe a referencia final e persiste no banco

Para este projeto, eu recomendaria um pequeno ajuste no desenho:

- nao salvar a **signed URL** no banco
- salvar apenas uma referencia **estável**, como:
  - URL canônica publica/CDN

Resumo:

- **signed URL** serve para upload temporário
- **valor salvo no banco** deve ser permanente e previsível

---

## Fluxo recomendado

### Fluxo ideal

1. frontend chama `POST /api/v1/uploads/intents`
2. backend valida tipo de arquivo, tamanho, destino e permissão
3. backend gera:
   - `upload_url`
   - `object_key`
   - `public_url` esperado
   - headers/métodos necessários
4. frontend faz upload direto para o GCS usando a `upload_url`
5. frontend chama o endpoint normal de create/update do recurso e envia:
   - `public_url`, conforme a estratégia escolhida
6. backend normaliza e salva na coluna `_url`

### Fluxo ainda melhor para robustez

1. frontend chama `POST /api/v1/uploads/intents`
2. frontend envia arquivo ao GCS
3. frontend chama `POST /api/v1/uploads/complete`
4. backend verifica se o objeto existe e se bate com o esperado
5. backend devolve a referencia canônica
6. frontend envia essa referencia no create/update do domínio

Esse fluxo com `complete` e melhor porque:

- reduz chance de salvar URL de upload que falhou
- permite validar MIME, tamanho e ownership
- facilita auditoria e evolução futura

---

## Recomendação de armazenamento no banco

### Nao recomendado

- salvar a signed URL de upload
- salvar signed URL de download

Motivo:

- expiram
- quebram histórico
- nao representam o recurso de forma estável

### Recomendado para MVP

Salvar na coluna `_url` uma URL canônica permanente, por exemplo:

- `https://storage.googleapis.com/<bucket>/<object_key>`
- ou URL de CDN/public media domain

Isso funciona bem se os arquivos forem públicos.

### Recomendado para longo prazo

Salvar internamente:

- `object_key`
- `bucket`
- `content_type`
- `size`
- `checksum`

e expor URL de leitura quando necessário.

Observação importante:

Como o schema atual usa colunas `_url`, o caminho mais pragmático para agora é:

- manter essas colunas recebendo uma URL canônica estável
- planejar evolução futura para `media_assets` ou `object_key` se surgirem arquivos privados

---

## Escopo confirmado no schema atual

Colunas `_url` do tipo `VARCHAR` encontradas no schema:

- `plans.image_url`
- `companies.logo_url`
- `people_identifications.image_url` (Avatar)
- `pets.image_url`
- `services.image_url`
- `sub_services.image_url`
- `products.image_url`
- `company_business_costs.invoice_url`
- `service_plans.image_url`
- `schedule_checkins.check_in_photo_url`
- `schedule_checkins.check_out_photo_url`
- `notifications.image_url`

Caso relacionado, mas fora do recorte inicial:

Recomendação:

- fase 1 cobre apenas colunas `_url` do tipo `VARCHAR`

---

## Decisão de arquitetura

### Modelo recomendado

Usar um **serviço genérico de upload** na API e manter a persistência final nos endpoints de domínio existentes.

Isso significa:

- upload é tratado em endpoints dedicados
- create/update de entidades do domínio continuam sendo responsáveis por salvar o `_url`

Vantagens:

- evita duplicar logica de upload em cada handler
- desacopla upload de formulários e recursos de negócio
- facilita rollout gradual por módulo

### Contrato de alto nível

#### 1. Criar upload intent

`POST /api/v1/uploads/intents`

Request sugerido:

```json
{
  "resource": "people_identifications",
  "field": "image_url",
  "file_name": "avatar.png",
  "content_type": "image/png",
  "size_bytes": 245123
}
```

Response sugerido:

```json
{
  "upload_url": "https://storage.googleapis.com/...",
  "method": "PUT",
  "headers": {
    "Content-Type": "image/png"
  },
  "object_key": "people_identifications/image_url/2026/03/uuid-avatar.png",
  "public_url": "https://storage.googleapis.com/petcontrol-assets/people_identifications/image_url/2026/03/uuid-avatar.png",
  "expires_at": "2026-03-26T15:00:00Z"
}
```

#### 2. Completar upload

`POST /api/v1/uploads/complete`

Request sugerido:

```json
{
  "resource": "people_identifications",
  "field": "image_url",
  "object_key": "people_identifications/image_url/2026/03/uuid-avatar.png"
}
```

Response sugerido:

```json
{
  "object_key": "people_identifications/image_url/2026/03/uuid-avatar.png",
  "public_url": "https://storage.googleapis.com/petcontrol-assets/people_identifications/image_url/2026/03/uuid-avatar.png"
}
```

#### 3. Persistir no domínio

Exemplo:

- `PATCH /api/v1/people/identifications/:id`
- o frontend envia `image_url` com a `public_url` ou `object_key`

Ou, se preferirem um backend mais rigoroso:

- enviar `upload_object_key`
- backend traduz para `public_url` antes de salvar

Essa segunda opção é mais segura.

---

## Recomendação final de fluxo

### Melhor equilíbrio entre simplicidade e segurança

1. `POST /uploads/intents`
2. upload direto para GCS
3. `POST /uploads/complete`
4. create/update do recurso com `object_key`
5. backend traduz `object_key -> public_url`
6. backend salva a URL publica na coluna `_url`

Assim:

- o banco continua compatível com o schema atual
- o frontend não precisa montar URL manualmente
- o backend continua dono da regra de persistência

---

## Plano de implementação

## API

### 1. Infra de storage

Criar uma camada dedicada em `apps/api` para GCS.

Estrutura sugerida:

```text
apps/api/internal/storage/gcs/
  client.go
  signer.go
  naming.go
  validate.go
```

Responsabilidades:

- gerar signed URLs
- construir `object_key`
- validar `resource`, `field`, MIME e tamanho
- produzir `public_url`

### 2. Configuração

Adicionar configurações de ambiente no backend:

- `GCS_BUCKET_NAME`
- `GCP_PROJECT_ID`
- `GCS_UPLOADS_BASE_PATH`
- `GCS_SIGNED_URL_TTL_SECONDS`
- `GCS_PUBLIC_BASE_URL`
- `GOOGLE_APPLICATION_CREDENTIALS` ou equivalente por workload identity

Observação:

Para signed URLs em GCS, a geração precisa de capacidade de assinatura. Em produção, prefira:

- service account com permissão de assinatura
- ou mecanismo de `SignBytes` sem chave privada embutida

### 3. Endpoint de upload intent

Criar handler dedicado:

```text
apps/api/internal/handler/uploads/
  create_intent.go
  complete.go
  routes.go
  swagger_models.go
```

Usecase sugerido:

```text
apps/api/internal/usecase/uploads/
  usecase.go
```

### 4. Regras de validação

Definir whitelist por campo.

Exemplos:

- `image_url`: imagens `image/png`, `image/jpeg`, `image/webp`
- `logo_url`: imagens
- `invoice_url`: pdf ou imagens permitidas
- `check_in_photo_url`: imagens

Também definir:

- limite máximo por campo
- extensões permitidas
- prefixo de caminho por tipo

Exemplo de chave:

```text
people_identifications/image_url/{yyyy}/{mm}/{uuid}-{safe-file-name}
companies/logo_url/{yyyy}/{mm}/{uuid}-{safe-file-name}
pets/image_url/{yyyy}/{mm}/{uuid}-{safe-file-name}
```

### 5. Endpoint de complete

No `complete`, o backend deve:

- verificar se o objeto existe
- confirmar `content_type` e tamanho quando possível
- devolver `object_key` e `public_url`

### 6. Integração com endpoints de domínio

Aplicar progressivamente a persistência dos campos:

- `people_identifications.image_url`
- `companies.logo_url`
- `pets.image_url`
- `plans.image_url`
- `services.image_url`
- `products.image_url`
- `company_business_costs.invoice_url`

Recomendação:

- aceitar `upload_object_key` no payload
- o backend converte para `public_url`
- salvar a URL final na coluna existente

Isso evita confiar no cliente para montar URL.

### 7. Testes na API

Cobrir:

- geração de `object_key`
- validação de MIME e tamanho
- handler de intent
- handler de complete
- tradução `object_key -> public_url`
- create/update de domínio usando `upload_object_key`

Evitar testes dependentes de GCS real no início. Usar:

- interface de signer/storage
- mocks/fakes

---

## Web

### 1. Feature de upload

Criar uma feature reutilizável no frontend.

Estrutura sugerida:

```text
apps/web/src/features/uploads/
  api.ts
  hooks.ts
  types.ts
  components/
    file-upload-input.tsx
    image-upload-field.tsx
    file-upload-progress.tsx
```

### 2. Fluxo no frontend

Para um campo como `avatar_url`:

1. usuário seleciona arquivo
2. frontend chama `POST /uploads/intents`
3. frontend executa upload direto no GCS
4. frontend chama `POST /uploads/complete`
5. frontend recebe `object_key` e `public_url`
6. formulário do recurso envia `upload_object_key` ou `public_url` para o endpoint de domínio

### 3. UX minima

Implementar:

- validação de tamanho e tipo antes do upload
- loading state
- progresso visual
- erro amigável
- preview para imagens quando fizer sentido

### 4. Integração por módulo

Aplicar primeiro em:

1. `people_identifications.image_url`
2. `companies.logo_url`
3. `pets.image_url`

Depois expandir para:

- `products.image_url`
- `services.image_url`
- `plans.image_url`
- `company_business_costs.invoice_url`

### 5. Web e estado

Usar:

- TanStack Query para mutations de intent e complete
- estado local de progresso no componente ou hook

Não usar:

- Zustand para guardar arquivo enviado
- cache global para blob/file

### 6. Testes no web

Cobrir:

- validação de arquivo
- hook de intent
- hook de complete
- componente de upload com sucesso/erro

---

## Libs

## `libs/shared-types`

Adicionar contratos compartilhados da feature:

Estrutura sugerida:

```text
libs/shared-types/src/uploads/
  intent.ts
  complete.ts
  resource-field.ts
```

Tipos sugeridos:

- `UploadResource`
- `UploadField`
- `CreateUploadIntentRequest`
- `CreateUploadIntentResponse`
- `CompleteUploadRequest`
- `CompleteUploadResponse`

Também pode conter:

- unions dos campos suportados
- enums de categoria de arquivo

## `libs/shared-utils`

Adicionar apenas helpers puros e pequenos, se realmente forem compartilhados.

Possíveis candidatos:

- `getFileExtension(fileName)`
- `isImageMimeType(contentType)`
- `formatBytes(size)`

Não colocar:

- fetch/upload logic
- hooks
- integração com GCS

---

## Sequência recomendada de execução

1. definir estrategia de persistência:
   - `public_url` publica
   - ou `object_key` interno com tradução no backend
2. implementar infra de GCS na API
3. criar endpoint `uploads/intents`
4. criar endpoint `uploads/complete`
5. criar tipos compartilhados em `libs/shared-types`
6. criar feature `uploads` no frontend
7. integrar primeiro `people_profiles.avatar_url`
8. expandir para os demais campos `_url`
9. tratar `photos_urls` em fase separada

---

## Fase recomendada de rollout

### Fase 1 (Imagens de Perfil/Identidade)

- `people_identifications.image_url`
- `companies.logo_url`
- `pets.image_url`

### Fase 2 (Catálogo e Operacional)

- `products.image_url`
- `services.image_url`
- `sub_services.image_url`
- `plans.image_url`
- `service_plans.image_url`
- `company_business_costs.invoice_url`

### Fase 3 (Check-ins e Notificações)

- `schedule_checkins.check_in_photo_url`
- `schedule_checkins.check_out_photo_url`
- `notifications.image_url`

---

## Riscos e decisões importantes

### 1. Bucket publico vs privado

Se os arquivos forem públicos:

- URL canônica no banco funciona bem

Se os arquivos forem privados:

- nao use apenas colunas `_url`
- considere evoluir para armazenar `object_key` e gerar URL de leitura sob demanda

### 2. Validação por campo

Nem todo `_url` deve aceitar qualquer arquivo.

### 3. Rollout gradual

Não tente adaptar todos os handlers e telas ao mesmo tempo.

---

## Definição de pronto

Considerar essa feature pronta quando:

- API gerar signed URL de upload com segurança
- frontend conseguir subir arquivo direto ao GCS
- backend conseguir confirmar upload e produzir referência canônica
- endpoints de domínio aceitarem `upload_object_key` ou referência equivalente
- colunas `_url` passarem a ser preenchidas sem upload binário via backend
- primeiro fluxo real estiver funcionando ponta a ponta

---

## Recomendação final

Sim, esse fluxo é o caminho certo para o projeto.

Eu só faria estes ajustes como padrão oficial:

- signed URL apenas para upload
- persistir no banco uma referência estável
- preferir `object_key` no trafego interno e `public_url` na persistência atual
- rollout gradual por recurso, começando por `people_identifications.image_url`
