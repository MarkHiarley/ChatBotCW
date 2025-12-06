# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copiar arquivos de dependências
COPY go.mod go.sum ./
RUN go mod download

# Copiar código fonte
COPY . .

# Build da aplicação
RUN CGO_ENABLED=0 GOOS=linux go build -o chatbot ./cmd/api

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Instalar certificados SSL para fazer requests HTTPS
RUN apk --no-cache add ca-certificates

# Copiar binário do build stage
COPY --from=builder /app/chatbot .

# Criar diretório para cache
RUN mkdir -p /app/cache

# Expor porta
EXPOSE 8080

# Comando para executar
CMD ["./chatbot"]