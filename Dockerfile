# Etapa 1: Build
FROM golang:1.23-alpine AS builder

# Define diretório de trabalho
WORKDIR /app

# Copia os arquivos Go
COPY go.mod go.sum ./
RUN go mod download

COPY . .


# Compila a aplicação para produção
RUN go build -o bucket-signer ./main.go

# Etapa 2: Execução
FROM alpine:latest

WORKDIR /root/

# Copia binário da etapa de build
COPY --from=builder /app/bucket-signer .

CMD ["./bucket-signer"] buc