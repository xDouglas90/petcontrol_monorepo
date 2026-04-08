# ADR 0001: API em Go com Gin

## Status

Aceito

## Contexto

A API precisa de alta previsibilidade operacional, baixo overhead e facilidade de deploy em containers.

## Decisão

Adotar Go como linguagem principal do backend e Gin como framework HTTP.

## Consequências

- Boa performance com baixo consumo de memoria.
- Curva de aprendizado baixa para handlers e middlewares.
- Ecossistema maduro para observabilidade, testes e integração com Postgres.
