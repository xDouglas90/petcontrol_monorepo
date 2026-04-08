FROM golang:1.26.1-alpine3.23 AS builder

RUN apk add --no-cache git build-base

ARG SQLC_VERSION=v1.30.0
ARG MIGRATE_VERSION=v4.19.0
ARG GOPROXY_URL=https://proxy.golang.org,direct

ENV GOPROXY=${GOPROXY_URL}
ENV PATH=$PATH:/go/bin

WORKDIR /src
COPY . .

WORKDIR /src/apps/api

RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@${SQLC_VERSION}
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@${MIGRATE_VERSION}

RUN go mod download
RUN sqlc generate
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/petcontrol-api ./cmd/server

FROM alpine:3.23

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /out/petcontrol-api /app/petcontrol-api
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate
COPY --from=builder /src/infra/migrations /app/infra/migrations
COPY apps/api/entrypoint.sh /app/entrypoint.sh

RUN chmod +x /app/entrypoint.sh

EXPOSE 8080

ENTRYPOINT ["/app/entrypoint.sh"]