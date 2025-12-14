# KVM Compose

🖥️ **kvm-compose** é uma ferramenta moderna escrita em **Go** que simplifica o gerenciamento de máquinas virtuais KVM usando workflows similares ao Docker Compose.

## ✨ Principais Melhorias da Versão Go

- 🚀 **Performance superior** - Execução muito mais rápida que scripts Bash
- 🛡️ **Maior robustez** - Tratamento de erros mais elegante e confiável  
- 🎨 **Interface colorida** - Saída visual aprimorada com cores e emojis
- 📦 **Binário único** - Fácil instalação e distribuição
- 🔧 **Parsing YAML nativo** - Processamento mais eficiente de configurações
- ⚡ **Concurrent operations** - Operações paralelas quando possível

## Features

- Easily create, start, stop, and manage KVM VMs.
- Declarative configuration for VMs.
- Streamlined workflow for development and testing.

## 📋 Prerequisites

- Linux with KVM support enabled
- `qemu-kvm`, `libvirt-clients`, and `virtinst` installed
- Network bridge configured (default: `br0`)
- `Go 1.21+` (para compilação)
- `wget` for downloading base images
- SSH key pair configured

### 🐧 Installation on Ubuntu/Debian

```bash
# Instalar KVM e dependências
sudo apt update
sudo apt install qemu-kvm libvirt-clients virtinst bridge-utils wget

# Instalar Go (se não tiver)
sudo apt install golang-go

# Ou baixar a versão mais recente:
# https://golang.org/dl/
```

## 🚀 Quick Start

1. **Clone and build:**
    ```bash
    git clone https://github.com/yourusername/kvm-compose.git
    cd kvm-compose
    
    # Método 1: Build e instalação automática
    make install
    
    # Método 2: Build manual
    make build
    ./build/kvm-compose --help
    ```

2. **Edit the configuration:**
    - Create or modify the `kvm-compose.yaml` file to define your VMs.

## Configuration Example

Here's a simple example of a `kvm-compose.yaml` file:

```yaml
# Kubernetes control plane
- name: k8s-cp-01
  distro: debian-13
  memory: 4096
  vcpus: 4
  disk_size: 20
  username: debian
  ssh_key_file: ~/.ssh/id_ed25519.pub
  networks:
    - bridge: br0
      ipv4: 192.168.1.40
      gateway: 192.168.1.1
      nameservers: [1.1.1.1, 8.8.8.8]

# Kubernetes worker node
- name: k8s-wrk-01
  distro: debian-13
  memory: 2048
  vcpus: 2
  disk_size: 15
  username: debian
  ssh_key_file: ~/.ssh/id_ed25519.pub
  networks:
    - bridge: br0
      ipv4: 192.168.1.41
      gateway: 192.168.1.1
      nameservers: [1.1.1.1, 8.8.8.8]
```

### Configuration Parameters

- **name**: VM identifier (required)
- **memory**: RAM in MB (default: 4096)
- **vcpus**: Number of virtual CPUs (default: 4)
- **disk_size**: Disk size in GB (default: 20)
- **username**: SSH user (default: debian)
- **ssh_key_file**: Path to SSH public key
- **networks**: Network configuration
  - **bridge**: Network bridge (default: br0)
  - **ipv4**: Static IP address
  - **gateway**: Network gateway
  - **nameservers**: DNS servers array

## 🎯 Available Commands

- 🆙 `up` - Create and start all VMs defined in the compose file
- ▶️ `start` - Start existing VMs  
- ⏹️ `stop` - Stop running VMs (graceful shutdown)
- ⬇️ `down` - Destroy VMs and remove disk files
- 📋 `list` - Show VMs configuration and status with colorized output

## 💡 Usage Examples

```bash
# Usando o binário instalado
kvm-compose up
kvm-compose list  
kvm-compose stop
kvm-compose down

# Usando arquivo compose customizado
kvm-compose up --compose my-lab.yaml

# Usando make targets para desenvolvimento
make run-up      # Compila e executa 'up'
make run-list    # Compila e executa 'list'  
make run-down    # Compila e executa 'down'

# Build e desenvolvimento
make build       # Compila o binário
make clean       # Limpa arquivos de build
make install     # Instala no sistema
make uninstall   # Remove do sistema
```

## 🎨 Visual Improvements

A versão Go inclui saída colorizada e emojis para melhor experiência:

- 🟢 VMs executando
- 🔴 VMs paradas  
- 🟡 VMs pausadas
- ⚪ VMs não criadas
- ✅ Operações bem-sucedidas
- ❌ Erros e falhas
- ⚠️ Avisos importantes

## 🏗️ Development

Para contribuir ou modificar o código:

```bash
# Clone e configure
git clone <repo-url>
cd kvm-compose

# Instale dependências
make deps

# Desenvolvimento
make build       # Build local
make test        # Execute testes
make clean       # Limpe build artifacts

# Teste local sem instalar
./build/kvm-compose --help
```

## License

GNU GENERAL PUBLIC LICENSE Version 3

---

**Note:** Replace placeholders and customize instructions as needed for your script.