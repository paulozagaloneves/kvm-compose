#!/bin/bash


# Verificação de dependências
echo "🔎 Verificando dependências (virt-install, virsh, wget)..."
REQUIRED_CMDS=(virt-install virsh wget)
for cmd in "${REQUIRED_CMDS[@]}"; do
  if ! command -v "$cmd" >/dev/null 2>&1; then
    echo "❌ Erro: o comando '$cmd' não está instalado. Por favor, instale-o antes de continuar."
    exit 1
  fi
done
echo "✅ Todas as dependências estão instaladas."


set -e

echo "Downloading kvm-compose binary..."
curl -L "$BIN_URL" -o /tmp/kvm-compose
echo "kvm-compose installed to $BIN_DEST"

# 1. Download binary
echo "⬇️  Baixando binário do kvm-compose..."
BIN_URL="https://github.com/paulozagaloneves/kvm-compose/releases/download/0.2.0/kvm-compose-linux-amd64"
BIN_DEST="/usr/local/bin/kvm-compose"
curl -L "$BIN_URL" -o /tmp/kvm-compose
echo "🚚 Movendo binário para /usr/local/bin/kvm-compose..."
sudo mv /tmp/kvm-compose "$BIN_DEST"
echo "🔒 Atribuindo permissões de execução..."
sudo chmod +x "$BIN_DEST"
echo "✅ kvm-compose instalado em $BIN_DEST"

echo "Creating configuration directories..."

# 2. Criar diretórios de configuração
echo "📁 Criando diretórios de configuração..."
CONFIG_DIR="$HOME/.config/kvm-compose"
mkdir -p "$CONFIG_DIR/images/vm"
mkdir -p "$CONFIG_DIR/images/upstream"
mkdir -p "$CONFIG_DIR/templates"
echo "✅ Diretórios criados."

echo "Downloading config.ini.example..."
curl -L "$CONFIG_URL" -o "$CONFIG_DIR/config.ini.example"

# 3. Baixar config.ini.example
echo "⬇️  Baixando config.ini.example..."
CONFIG_URL="https://raw.githubusercontent.com/paulozagaloneves/kvm-compose/main/config.ini.example"
curl -L "$CONFIG_URL" -o "$CONFIG_DIR/config.ini.example"
echo "✅ config.ini.example salvo em $CONFIG_DIR/config.ini.example"

echo "Downloading template files..."

# 4. Baixar arquivos de template
echo "⬇️  Baixando arquivos de template..."
TEMPLATES=(
  "almalinux10.ini"
  "debian13.ini"
  "fedora43.ini"
  "meta-data.tmpl"
  "network-config-almalinux.tmpl"
  "network-config.tmpl"
  "ubuntu24.04.ini"
  "user-data.tmpl"
)
TEMPLATE_BASE="https://raw.githubusercontent.com/paulozagaloneves/kvm-compose/main/templates"
for tmpl in "${TEMPLATES[@]}"; do
  echo "  ⬇️  Baixando $tmpl..."
  curl -L "$TEMPLATE_BASE/$tmpl" -o "$CONFIG_DIR/templates/$tmpl"
done
echo "✅ Templates salvos em $CONFIG_DIR/templates"

echo "🎉 Instalação e configuração do kvm-compose concluídas!"
