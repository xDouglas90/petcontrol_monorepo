# ADR 0003: Worker como processo separado

## Status

Aceito

## Contexto

Processamento assíncrono (notificações, tarefas futuras e agendamentos) nao deve bloquear o ciclo de request/response da API.

## Decisão

Executar Worker como aplicação separada em apps/worker, integrada a Redis/Asynq.

## Consequências

- Escalabilidade independente entre API e Worker.
- Falhas do Worker nao indisponibilizam a API.
- Observabilidade e operação por fila ficam mais claras para SRE e suporte.
