FROM golang:1.24.4 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

ARG BUILD_VERSION="N/A"
ARG BUILD_DATE="N/A"
ARG BUILD_COMMIT="N/A"

COPY . .

RUN CGO_ENABLED=0 go build \
    -ldflags "\
      -X 'github.com/gdyunin/aegis-vault-keeper/internal/server/buildinfo.Version=${BUILD_VERSION}' \
      -X 'github.com/gdyunin/aegis-vault-keeper/internal/server/buildinfo.Date=${BUILD_DATE}' \
      -X 'github.com/gdyunin/aegis-vault-keeper/internal/server/buildinfo.Commit=${BUILD_COMMIT}'" \
    -o aegis_vault_keeper ./cmd/server

FROM alpine:latest
RUN for i in 1 2 3; do \
        apk update && break || sleep 2; \
    done && \
    apk --no-cache add ca-certificates
WORKDIR /app

RUN adduser -D -u 1001 appuser && \
    mkdir -p /app/filestorage && \
    mkdir -p /app/certs && \
    mkdir -p /app/config && \
    chown -R 1001:1001 /app/filestorage && \
    chown -R 1001:1001 /app/certs && \
    chown -R 1001:1001 /app/config

COPY --from=builder /app/aegis_vault_keeper .
COPY certs/ /app/certs/
COPY config/ /app/config/
RUN chown 1001:1001 /app/aegis_vault_keeper && \
    chown -R 1001:1001 /app/certs && \
    chown -R 1001:1001 /app/config

USER 1001
CMD ["/app/aegis_vault_keeper"]