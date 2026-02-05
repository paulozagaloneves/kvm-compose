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
echo "⬇️  Detectando arquitetura do processador..."
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)
    ARCH_DL="amd64" ;;
  aarch64|arm64)
    ARCH_DL="arm64" ;;
  *)
    echo "❌ Arquitetura $ARCH não suportada." ; exit 1 ;;
esac

echo "⬇️  Baixando binário do kvm-compose para $ARCH_DL..."
if [[ -z "$VERSION" || "$VERSION" == "latest" ]]; then
  echo "🔎 Descobrindo a última versão do kvm-compose no GitHub..."
  VERSION=$(curl -s https://api.github.com/repos/paulozagaloneves/kvm-compose/releases/latest | grep 'tag_name' | cut -d '"' -f4)
  if [[ -z "$VERSION" ]]; then
    echo "❌ Não foi possível obter a última versão."
    exit 1
  fi
  echo "ℹ️  Última versão encontrada: $VERSION"
fi
BIN_URL="https://github.com/paulozagaloneves/kvm-compose/releases/download/${VERSION}/kvm-compose_${VERSION}_linux_${ARCH_DL}.tar.gz"
BIN_DEST="/usr/local/bin/kvm-compose"
curl -sS -L "$BIN_URL" -o /tmp/kvm-compose.tar.gz
echo "🚚 Extraindo binário para /usr/local/bin/kvm-compose..."
sudo tar -xzf /tmp/kvm-compose.tar.gz -C /usr/local/bin
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

# 4. Baixar config.ini default se não existir
if [ ! -f "$CONFIG_DIR/config.ini" ]; then
    echo "⬇️  Baixando config.ini..."
    CONFIG_URL="https://raw.githubusercontent.com/paulozagaloneves/kvm-compose/main/config.ini"
    curl -sS -L "$CONFIG_URL" -o "$CONFIG_DIR/config.ini"
    echo "✅ config.ini salvo em $CONFIG_DIR/config.ini"
else
    echo "ℹ️  config.ini já existe em $CONFIG_DIR/config.ini - pulando download"
fi


# 5. Baixar arquivos de template (apenas os que não existirem)
echo "🔄 Verificando arquivos de template..."
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

downloaded=0
already_exist=0

set +e
for tmpl in "${TEMPLATES[@]}"; do
  if [ ! -f "$CONFIG_DIR/templates/$tmpl" ]; then
    echo "  ⬇️  Baixando $tmpl..."
    if curl -sS -L "$TEMPLATE_BASE/$tmpl" -o "$CONFIG_DIR/templates/$tmpl" || true; then
      ((downloaded++))
    else
      echo "  ❌ Falha ao baixar $tmpl"
    fi
  else
    ((already_exist++))
  fi
done

if [ $downloaded -gt 0 ]; then
  echo "✅ $downloaded novos templates baixados para $CONFIG_DIR/templates"
fi
if [ $already_exist -gt 0 ]; then
  echo "ℹ️  $already_exist templates já existentes foram mantidos"
fi
set -e

echo
echo "ℹ️  Você pode customizar os templates cloud-init (*.tmpl) localizados em: $CONFIG_DIR/templates"
echo "   Templates disponíveis: meta-data.tmpl, network-config-almalinux.tmpl, network-config.tmpl, user-data.tmpl"
echo
echo "💡 Para adicionar suporte a novas distribuições, basta criar um novo arquivo .ini na pasta de templates com as configurações desejadas."
echo
echo "Exemplo básico de kvm-compose.yaml:"
cat <<EOF
- name: minha-vm
  distro: debian13
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
