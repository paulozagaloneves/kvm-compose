#!/bin/bash

# Ativa modo debug se DEBUG=1
if [[ "$DEBUG" == "1" ]]; then
  echo "🔍 Modo debug ativado: comandos serão exibidos."
  set -x
fi

set -e

# Verificação de dependências
echo "🔎 Verificando dependências (virt-install, virsh, wget, curl)..."
REQUIRED_CMDS=(virt-install virsh wget curl)
for cmd in "${REQUIRED_CMDS[@]}"; do
  if ! command -v "$cmd" >/dev/null 2>&1; then
    echo "❌ Erro: o comando '$cmd' não está instalado. Por favor, instale-o antes de continuar."
    exit 1
  fi
done
echo "✅ Todas as dependências estão instaladas."




# 1. Download binary
echo "⬇️  Baixando binário do kvm-compose..."
BIN_URL="https://github.com/paulozagaloneves/kvm-compose/releases/download/0.2.0/kvm-compose-linux-amd64"
BIN_DEST="/usr/local/bin/kvm-compose"
  curl -sS -L "$BIN_URL" -o /tmp/kvm-compose
echo "🚚 Movendo binário para /usr/local/bin/kvm-compose..."
sudo mv /tmp/kvm-compose "$BIN_DEST"
echo "🔒 Atribuindo permissões de execução..."
sudo chmod +x "$BIN_DEST"
echo "✅ kvm-compose instalado em $BIN_DEST"


# 2. Criar diretórios de configuração
echo "📁 Criando diretórios de configuração..."
CONFIG_DIR="$HOME/.config/kvm-compose"
mkdir -p "$CONFIG_DIR/images/vm"
mkdir -p "$CONFIG_DIR/images/upstream"
mkdir -p "$CONFIG_DIR/templates"
echo "✅ Diretórios criados."

# 3. Baixar config.ini.example
echo "⬇️  Baixando config.ini.example..."
CONFIG_URL="https://raw.githubusercontent.com/paulozagaloneves/kvm-compose/main/config.ini.example"
  curl -sS -L "$CONFIG_URL" -o "$CONFIG_DIR/config.ini.example"
echo "✅ config.ini.example salvo em $CONFIG_DIR/config.ini.example"

# 4. Baixar config.ini default
echo "⬇️  Baixando config.ini..."
CONFIG_URL="https://raw.githubusercontent.com/paulozagaloneves/kvm-compose/main/config.ini"
  curl -sS -L "$CONFIG_URL" -o "$CONFIG_DIR/config.ini"
echo "✅ config.ini salvo em $CONFIG_DIR/config.ini"


# 5. Baixar arquivos de template
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
  curl -sS -L "$TEMPLATE_BASE/$tmpl" -o "$CONFIG_DIR/templates/$tmpl"
done
echo "✅ Templates salvos em $CONFIG_DIR/templates"

echo
echo "ℹ️  Você pode customizar os templates cloud-init (*.tmpl) localizados em: $CONFIG_DIR/templates"
echo "   Templates disponíveis: meta-data.tmpl, network-config-almalinux.tmpl, network-config.tmpl, user-data.tmpl"
echo
echo "💡 Para adicionar suporte a novas distribuições, basta criar um novo arquivo .ini na pasta de templates com as configurações desejadas."
echo
echo "Exemplo básico de kvm-compose.yaml:"
cat <<EOF
- name: minha-vm
  distro: debian-13
  memory: 2048
  vcpus: 2
  disk_size: 10
  username: debian
  ssh_key_file: ~/.ssh/id_ed25519.pub
  networks:
    - host_bridge: br0
      guest_ipv4: 192.168.1.50
      guest_gateway4: 192.168.1.1
      guest_nameservers: [1.1.1.1, 8.8.8.8]
EOF

echo "---"
echo "🎉 Instalação e configuração do kvm-compose concluídas!"
