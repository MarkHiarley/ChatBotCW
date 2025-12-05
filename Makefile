.PHONY: build run test clean

# Variáveis
BINARY_NAME=chatbot-cw
MAIN_PATH=./cmd/api

# Build do projeto
build:
	go build -o $(BINARY_NAME) $(MAIN_PATH)

# Executar o projeto
run:
	go run $(MAIN_PATH)/main.go

# Executar testes
test:
	go test -v ./...

# Limpar arquivos gerados
clean:
	go clean
	rm -f $(BINARY_NAME)

# Instalar dependências
deps:
	go mod tidy
	go mod download

# Executar com Docker
docker-build:
	docker-compose build

docker-run:
	docker-compose up

docker-down:
	docker-compose down

# Exemplo de teste da API
test-api:
	@echo "Testando health check..."
	curl -s http://localhost:8080/health
	@echo "\n\nTestando chat..."
	curl -X POST http://localhost:8080/chat \
		-H "Content-Type: application/json" \
		-d '{"question": "O que é a Cloudwalk?"}'