# Makefile para kvm-compose

# Variáveis
BINARY_NAME=kvm-compose
VERSION=1.0.0
BUILD_DIR=dist
CDR=.
INSTALL_DIR=/usr/local/bin

# Flags de build
LDFLAGS := -X github.com/paulozagaloneves/kvm-compose/internal/common.Version=$(VERSION)
LDFLAGS += -X github.com/paulozagaloneves/kvm-compose/internal/common.BuildDate=$(DATE)
LDFLAGS += -X github.com/paulozagaloneves/kvm-compose/internal/common.BuildUser=$(BUILT_BY)
LDFLAGS += -X github.com/paulozagaloneves/kvm-compose/internal/common.CommitID=$(COMMITID)
LDFLAGS += -X github.com/paulozagaloneves/kvm-compose/internal/common.GoVersion=$(GO_VERSION)
LDFLAGS += -X github.com/paulozagaloneves/kvm-compose/internal/common.GoOS=$(GO_OS)
LDFLAGS += -X github.com/paulozagaloneves/kvm-compose/internal/common.GoArch=$(GO_ARCH)

# PLATAFORMAS SUPORTADAS
PLATAFORMAS = \
	linux/amd64 \
	linux/arm64 \

# Comandos Go
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod


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

# Cria os binários para múltiplas plataformas (release)
release: deps
	@echo "Construindo para múltiplas plataformas..."
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATAFORMAS); do \
		OS=$$(echo $$platform | cut -d/ -f1); \
		ARCH=$$(echo $$platform | cut -d/ -f2); \
		OUT=$(BUILD_DIR)/$(BINARY_NAME)-$$OS-$$ARCH; \
		echo "-> $$OS/$$ARCH (Saída: $$OUT)"; \
		GOOS=$$OS GOARCH=$$ARCH go build -ldflags "$(LDFLAGS)" -v -o $$OUT $(CDR)/main.go || exit 1; \
	done	

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