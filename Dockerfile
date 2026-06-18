# Etapa 1: Build
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Cacheia dependências
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Binário estático (CGO desligado) — roda em qualquer base, sem libc dinâmica
RUN CGO_ENABLED=0 GOOS=linux go build -o bucket-signer .

# Etapa 2: Execução
FROM alpine:latest

# Certificados raiz (necessários para chamadas HTTPS à AWS) + usuário não-root
RUN apk add --no-cache ca-certificates && \
    addgroup -S app && adduser -S app -G app

WORKDIR /home/app

COPY --from=builder /app/bucket-signer .

# Porta padrão do serviço (sobrescrevível via env PORT)
EXPOSE 8081

USER app

CMD ["./bucket-signer"]
