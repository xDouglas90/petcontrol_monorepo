# MailHog e Variáveis de Ambiente para Testes de Email

## MailHog

Para desenvolvimento e testes locais de envio de email, utilize o MailHog:

- **Suba o serviço MailHog:**

  ```sh
  docker compose -f infra/docker/docker-compose.yml up mailhog
  ```

- **Interface web:** [http://localhost:8025](http://localhost:8025)
- **SMTP:**
  - Host: `localhost`
  - Porta: `1025`

## Configuração do Backend

No ambiente de desenvolvimento, configure as seguintes variáveis no `.env`:

```text
SMTP_HOST=localhost
SMTP_PORT=1025
SMTP_USERNAME=
SMTP_PASSWORD=
SMTP_FROM_EMAIL=no-reply@petcontrol.local
SMTP_FROM_NAME=PetControl
```

## Configuração do Frontend

A UI `/people` pode ser acessada em:

- `http://localhost:5173/people` (ajuste conforme porta do Vite)

## Observações

- O worker deve apontar para o SMTP do MailHog apenas em desenvolvimento/teste.
- Não use credenciais reais em `.env` local.
- O envio de email pode ser validado visualmente na interface do MailHog.
