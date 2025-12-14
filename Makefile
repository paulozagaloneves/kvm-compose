# Makefile para kvm-compose

# Variáveis
BINARY_NAME=kvm-compose
VERSION=1.0.0
BUILD_DIR=build
INSTALL_DIR=/usr/local/bin

# Comandos Go
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Flags de build
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

.PHONY: all build clean test deps install uninstall help

all: deps build

## build: Compila o binário
build: deps
	@echo "🔨 Compilando $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "✅ Build concluído: $(BUILD_DIR)/$(BINARY_NAME)"

## deps: Baixa e instala dependências
deps:
	@echo "📦 Instalando dependências..."
	$(GOMOD) tidy
	$(GOMOD) download
	@echo "✅ Dependências instaladas"

## clean: Limpa arquivos de build
clean:
	@echo "🧹 Limpando arquivos de build..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	@echo "✅ Limpeza concluída"

## test: Executa testes
test:
	@echo "🧪 Executando testes..."
	$(GOTEST) -v ./...

## install: Instala o binário no sistema
install: build
	@echo "📥 Instalando $(BINARY_NAME) em $(INSTALL_DIR)..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/
	sudo chmod +x $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "✅ $(BINARY_NAME) instalado com sucesso!"
	@echo "   Use: kvm-compose --help"

## uninstall: Remove o binário do sistema
uninstall:
	@echo "🗑️  Removendo $(BINARY_NAME)..."
	sudo rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "✅ $(BINARY_NAME) removido"

## run-up: Executa 'up' diretamente
run-up: build
	./$(BUILD_DIR)/$(BINARY_NAME) up

## run-list: Executa 'list' diretamente
run-list: build
	./$(BUILD_DIR)/$(BINARY_NAME) list

## run-down: Executa 'down' diretamente
run-down: build
	./$(BUILD_DIR)/$(BINARY_NAME) down

## help: Mostra esta ajuda
help:
	@echo "Comandos disponíveis:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

# Target padrão
.DEFAULT_GOAL := help