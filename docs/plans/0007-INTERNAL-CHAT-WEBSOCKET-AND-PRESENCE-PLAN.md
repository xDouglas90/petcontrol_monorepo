# Plano de Ação e Execução - WebSocket e Presença Dinâmica no Chat Interno

## Objetivo

Definir o escopo técnico, arquitetural e operacional para evoluir o chat interno entre `admin` e `system`, saindo do modelo atual de persistência via HTTP polling para um fluxo em tempo real com WebSocket e presença dinâmica no PetControl.

Esta PR deve introduzir:

- canal WebSocket autenticado para o chat interno;
- entrega em tempo real de novas mensagens;
- presença dinâmica mínima entre `admin` e `system`;
- gerenciamento seguro de conexões por tenant;
- base observável e escalável para futuras evoluções de chat.

## Contexto Atual

- O chat interno já possui persistência básica de mensagens entre `admin` e `system`.
- O dashboard `admin` já renderiza histórico persistido e permite envio de texto.
- Ainda não existe:
  - atualização em tempo real;
  - presença online/offline/ocupado dinâmica;
  - infraestrutura de sockets no backend;
  - broadcast entre múltiplas conexões do mesmo tenant.

## Decisão de Biblioteca

Biblioteca escolhida:

- `github.com/coder/websocket`

Justificativas:

- suporte nativo a `context.Context` para cancelamento e timeouts;
- escritas concorrentes seguras;
- manutenção ativa.

Decisão explícita:

- não usar `gorilla/websocket`, por estar arquivada e sem manutenção ativa.

## Resultado Esperado da PR

Ao final desta PR, o sistema deve apresentar:

- endpoint WebSocket autenticado para chat interno;
- recebimento em tempo real de mensagens novas;
- broadcast para múltiplas sessões do mesmo usuário ou conversa;
- status de presença básica:
  - `online`
  - `offline`
  - `busy` ou equivalente, se houver ação ativa de chat;
- atualização visual do dashboard sem refresh manual;
- tratamento seguro de reconexão e cleanup de clientes.

## Escopo Funcional Confirmado

## 1. Canal WebSocket

- Criar endpoint dedicado para chat interno.
- Validar autenticação antes de aceitar upgrade da conexão.
- Restringir acesso ao recorte `admin <-> system` dentro do mesmo tenant.
- Garantir isolamento multi-tenant.

## 2. Gerenciamento de Conexões

- Adotar uma goroutine por conexão.
- Separar leitura, escrita e lifecycle management.
- Manter registro central de clientes conectados.
- Utilizar canal de broadcast para distribuição de eventos.
- Encerrar conexões com `context.Context` em casos de:
  - cancelamento;
  - timeout;
  - erro de leitura/escrita;
  - logout ou token inválido.

## 3. Presença Dinâmica

Presença mínima nesta fase:

- `online`: conexão ativa;
- `offline`: sem conexão ativa;
- `busy`: opcional nesta PR, desde que tenha semântica clara.

Regras sugeridas:

- presença associada ao par `tenant + user`;
- múltiplas abas ou dispositivos mantêm `online` enquanto houver ao menos uma conexão viva;
- desconexão deve disparar atualização de presença;
- UI do chat deve refletir a presença recebida pelo socket.

## 4. Mensagens em Tempo Real

- Novas mensagens persistidas devem disparar evento WebSocket.
- O cliente emissor deve receber confirmação consistente.
- O cliente destinatário deve receber atualização imediata.
- O histórico persistido continua sendo a fonte de verdade.

## 5. Segurança

- usar `wss://` em ambientes reais;
- autenticar por token antes de aceitar a conexão;
- validar tenant, role e participante autorizado;
- limitar tamanho de mensagem;
- limitar frequência de eventos para mitigar flood;
- fechar conexão em caso de payload inválido ou comportamento abusivo.

## 6. Escalabilidade

- evitar operações pesadas dentro da goroutine de socket;
- manter persistência e consultas fora do loop crítico quando possível;
- preparar arquitetura para funcionar atrás de Nginx ou HAProxy;
- medir:
  - conexões ativas;
  - latência de entrega;
  - erros de conexão;
  - reconexões;
  - falhas de broadcast.

Observação:

- se a arquitetura passar a exigir múltiplas instâncias da API, pode ser necessário adicionar barramento externo futuro para presença e fan-out entre nós.

## 7. Boas Práticas de Código

- separar claramente:
  - upgrade/autenticação;
  - leitura;
  - escrita;
  - registro de clientes;
  - broadcast;
  - presença;
- tratar erros e cleanup de forma robusta;
- não deixar goroutines órfãs;
- cobrir cenários de carga e reconexão.

## Arquitetura Recomendada

## Backend

- `internal/realtime` ou módulo equivalente para:
  - hub de conexões;
  - cliente conectado;
  - eventos de presença;
  - broadcast de mensagens;
- handler HTTP responsável apenas por:
  - autenticar;
  - validar contexto;
  - aceitar o socket;
  - delegar ao hub.

## Frontend

- camada de client WebSocket separada do `rest-client`;
- integração com o dashboard para:
  - subscrever eventos da conversa selecionada;
  - atualizar mensagens em memória;
  - refletir presença do contato;
  - reconectar automaticamente com backoff simples.

## Fase 0 - Descoberta e Contrato

Status atual:

- Contrato inicial de eventos WebSocket documentado em [internal-chat-realtime.md](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/docs/conventions/internal-chat-realtime.md:1).
- Autenticação de handshake fechada por `Authorization: Bearer <token>`.
- Subprotocol definido como `petcontrol.internal-chat.v1`.
- Presença desta PR fechada em `online/offline`.
- Estado `busy` explicitamente adiado para etapa futura.
- Estratégia de reconexão do Web definida com backoff incremental curto.

### 0.1 Ações

- Definir contrato de eventos WebSocket.
- Confirmar formato de autenticação no handshake.
- Definir se `busy` entra nesta PR ou fica como etapa posterior.
- Mapear estratégia de reconexão no Web.

### 0.2 Checks

- [x] O contrato de eventos está documentado.
- [x] O recorte de presença desta PR está fechado.

## Fase 1 - Infraestrutura de WebSocket na API

Status atual:

- `coder/websocket` adicionado na API e pronto para uso no endpoint do chat interno.
- Novo hub em `internal/realtime` registrando conexões por `company_id + user_id`.
- Endpoint `GET /api/v1/chat/system/:user_id/ws` implementado com autenticação, validação do par `admin <-> system` e subprotocol `petcontrol.internal-chat.v1`.
- Lifecycle inicial fechado com `context.Context`, `CloseRead`, limite de payload e cleanup no disconnect.
- Evento inicial `chat.connected` já emitido após handshake bem-sucedido.
- Cobertura criada para:
  - conexão válida;
  - rejeição sem token;
  - rejeição de par proibido;
  - cleanup do hub ao desconectar.

### 1.1 Ações

- Introduzir `coder/websocket`.
- Criar endpoint WebSocket autenticado.
- Criar hub de conexões por tenant e usuário.
- Implementar lifecycle com `context.Context`.

### 1.2 Checks

- [x] A conexão autentica e conecta corretamente.
- [x] Conexões inválidas são rejeitadas.
- [x] Cleanup ocorre ao desconectar.

## Fase 2 - Broadcast de Mensagens

Status atual:

- O `POST /api/v1/chat/system/:user_id/messages` agora dispara fan-out WebSocket após persistir a mensagem.
- O hub passou a manter as conexões vivas da conversa e filtrar o broadcast pelo par exato `admin <-> system`.
- O evento `chat.message.created` já chega em tempo real para sessões conectadas da conversa.
- A emissão continua dependente da persistência bem-sucedida no banco, preservando o REST como fonte canônica.
- Cobertura criada para:
  - broadcast do hub para a conversa correta;
  - envio REST seguido de recebimento do evento no socket.

### 2.1 Ações

- Emitir evento ao persistir nova mensagem.
- Entregar evento ao emissor e ao destinatário conectado.
- Garantir consistência entre persistência e broadcast.

### 2.2 Checks

- [x] Novas mensagens aparecem em tempo real sem refresh.
- [x] O histórico persistido continua íntegro.

## Fase 3 - Presença Dinâmica

Status atual:

- O hub agora calcula presença `online/offline` por `company_id + user_id`, com contagem de conexões e `last_changed_at`.
- O socket passou a emitir `chat.presence.snapshot` logo após `chat.connected`.
- Mudanças reais de status são propagadas por `chat.presence.updated` quando o participante entra ou sai da conversa.
- Múltiplas conexões do mesmo usuário mantêm o estado `online` enquanto houver ao menos uma sessão ativa.
- Cobertura criada para:
  - snapshot inicial da conversa;
  - atualização `online` quando o outro participante conecta;
  - atualização `offline` quando ele desconecta.

### 3.1 Ações

- Calcular presença com base em conexões ativas.
- Emitir eventos de presença ao conectar e desconectar.
- Atualizar o dashboard para refletir presença dinâmica.

### 3.2 Checks

- [x] `online` e `offline` mudam em tempo real.
- [x] Múltiplas conexões do mesmo usuário não quebram a presença.

## Fase 4 - Robustez, Observabilidade e Carga

Status atual:

- O hub agora expõe snapshot interno de métricas com:
  - conexões ativas;
  - participantes online;
  - conexões abertas/fechadas;
  - eventos e entregas de broadcast;
  - falhas de broadcast;
  - payloads inválidos;
  - falhas de ping;
  - erros de socket.
- O socket passou a usar heartbeat com `Ping` periódico do servidor.
- Payload inesperado enviado pelo cliente agora gera `chat.error` e fechamento com `StatusPolicyViolation`.
- O módulo `internal/realtime` ganhou teste concorrente de registro/broadcast/unregister e benchmark de fan-out.
- Cobertura criada para:
  - contadores do hub;
  - rejeição de payload inválido no socket.
  - cenário concorrente com múltiplas conexões;
  - benchmark simples de broadcast.

### 4.1 Ações

- Adicionar métricas básicas.
- Cobrir timeouts, payload inválido e reconexão.
- Executar testes de carga com volume elevado de conexões.

### 4.2 Checks

- [x] Existem métricas mínimas para conexões e erros.
- [ ] O sistema suporta carga compatível com o ambiente esperado.

Observação:

- A PR já possui benchmark e teste concorrente como smoke/load check inicial.
- Ainda falta, se desejado, ampliar isso para soak test com volume significativamente maior e alvo explícito de capacidade.

## Testes Esperados

- testes unitários do hub e do registro de clientes;
- testes de handler para autenticação e rejeição de conexão inválida;
- testes de integração para broadcast entre `admin` e `system`;
- testes do Web para atualização em tempo real e presença;
- testes de carga simulando milhares de conexões.

## Riscos e Cuidados

- vazamento de goroutines por cleanup incorreto;
- race conditions em presença e broadcast;
- fan-out ineficiente em cenários com múltiplas instâncias;
- reconexão excessiva no frontend;
- uso incorreto de conexões sem TLS em produção.

## Decisões Recomendadas para Manter a PR Saudável

- implementar primeiro presença `online/offline`;
- deixar `busy` como opcional, apenas se surgir regra clara;
- manter histórico persistido via REST como fonte canônica;
- usar WebSocket para eventos e atualização incremental;
- evitar introduzir voz, vídeo ou anexos nesta fase.

## Ordem Recomendada de Execução

1. Fase 0: descoberta e contrato.
2. Fase 1: infraestrutura WebSocket na API.
3. Fase 2: broadcast de mensagens.
4. Fase 3: presença dinâmica.
5. Fase 4: robustez, observabilidade e carga.

## Resultado Esperado

Se este plano for executado com sucesso, o chat interno do PetControl deixará de ser apenas persistido e passará a oferecer experiência em tempo real, com presença dinâmica, arquitetura mais moderna e base sólida para futuras evoluções de colaboração interna.
