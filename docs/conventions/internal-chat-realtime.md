# Chat Interno em Tempo Real

Esta convenção documenta o contrato inicial do chat interno em tempo real entre usuários `admin` e `system`.

## Escopo

Este recorte cobre:

- entrega em tempo real de mensagens persistidas;
- presença dinâmica básica;
- autenticação no handshake;
- isolamento multi-tenant;
- contrato de eventos WebSocket.

Fora do escopo desta fase:

- voz;
- vídeo;
- anexos;
- indicadores ricos de digitação;
- presença `busy`.

## Biblioteca

- Backend WebSocket: `github.com/coder/websocket`

## Endpoint

Formato base:

- `GET /api/v1/chat/system/:user_id/ws`

Onde:

- `:user_id` representa o contato da conversa;
- a URL deve respeitar o contexto autenticado do tenant.

## Autenticação no Handshake

Formato escolhido para esta fase:

- header `Authorization: Bearer <token>`

Motivos:

- reaproveita o mesmo modelo já usado no REST;
- evita expor token em query string;
- simplifica middleware e observabilidade.

## Subprotocol

- `petcontrol.internal-chat.v1`

## Presença

Estados confirmados para esta PR:

- `online`
- `offline`

Regra:

- `online` quando o usuário tiver pelo menos uma conexão WebSocket ativa no tenant;
- `offline` quando não houver conexões ativas.

Decisão:

- `busy` fica explicitamente fora desta PR e deve ser tratado em etapa posterior, caso exista regra clara de negócio.

## Contrato de Eventos

Todos os eventos seguem envelope comum:

```json
{
  "type": "chat.connected",
  "company_id": "uuid",
  "counterpart_user_id": "uuid",
  "emitted_at": "2026-04-20T12:00:00Z"
}
```

### `chat.connected`

Emitido quando a conexão é aceita.

Payload adicional:

```json
{
  "connection_id": "uuid-ou-id-curto",
  "viewer_user_id": "uuid",
  "viewer_role": "admin"
}
```

### `chat.message.created`

Emitido quando uma nova mensagem persistida entra na conversa.

Payload adicional:

```json
{
  "message": {
    "id": "uuid",
    "conversation_id": "uuid",
    "company_id": "uuid",
    "sender_user_id": "uuid",
    "sender_name": "Maria",
    "sender_role": "admin",
    "sender_image_url": null,
    "body": "Texto da mensagem",
    "created_at": "2026-04-20T12:00:00Z"
  }
}
```

### `chat.presence.snapshot`

Emitido logo após conectar, com o estado inicial da conversa.

Payload adicional:

```json
{
  "presences": [
    {
      "user_id": "uuid",
      "status": "online",
      "connections": 1,
      "last_changed_at": "2026-04-20T12:00:00Z"
    }
  ]
}
```

### `chat.presence.updated`

Emitido quando a presença de um participante muda.

Payload adicional:

```json
{
  "presence": {
    "user_id": "uuid",
    "status": "offline",
    "connections": 0,
    "last_changed_at": "2026-04-20T12:05:00Z"
  }
}
```

### `chat.error`

Emitido quando a conexão detecta erro recuperável de protocolo ou autorização contextual.

Payload adicional:

```json
{
  "code": "chat_forbidden",
  "message": "conversation not allowed"
}
```

## Reconexão no Web

Estratégia escolhida:

- backoff incremental curto;
- revalidação de sessão a cada tentativa;
- ressincronização via REST ao reconectar, se necessário.

Backoff inicial sugerido:

- `1000ms`
- `2000ms`
- `5000ms`
- `10000ms`

## Fonte de Verdade

Regra principal:

- o histórico persistido via REST continua sendo a fonte canônica;
- o WebSocket serve para eventos incrementais e atualização de presença.

## Segurança

- usar `wss://` fora do ambiente local;
- limitar tamanho de payload;
- rejeitar conexão sem token válido;
- rejeitar conexão fora do par autorizado `admin <-> system`;
- isolar sempre por `company_id`.
