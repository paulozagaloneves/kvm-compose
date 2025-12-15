# KVM Compose

Versão 0.1.0 Codename: "Gambiarra" - Dezembro de 2025

🖥️ **kvm-compose** é uma ferramenta moderna escrita em **Go** que simplifica o gerenciamento de máquinas virtuais KVM usando workflows similares ao Docker Compose.

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

### 🐧 Install KVM on Ubuntu/Debian

```bash
# Instalar KVM e dependências
sudo apt update
sudo apt install -y qemu-kvm libvirt-daemon libvirt-clients bridge-utils virt-manager virtinst wget
```

Other Tutorial sites:
- https://cloudspinx.com/install-kvm-on-debian-with-virt-manager-and-cockpit/
- https://sysguides.com/install-kvm-on-linux
- https://phoenixnap.com/kb/ubuntu-install-kvm

### 🔧 Configure network bridge on Debian

1. **Install the required utilities**

```bash
sudo apt update
sudo apt install bridge-utils
```

2. **Identify your physical network interface**

```bash
ip -f inet a s
```

3. **Configure the bridge**

Create a configuration file for the bridge in the /etc/network/interfaces.d/ directory. For example, create a file named br0

```bash
sudo nano /etc/network/interfaces.d/br0
```

Add the following configuration, replacing **eth0** with your actual interface name and adjusting the IP settings as needed:

* **For a static IP:**

  
```
## static ip config file for br0 ##
auto br0
iface br0 inet static
address 192.168.1.100
netmask 255.255.255.0
gateway 192.168.1.1
dns-nameservers 8.8.8.8 8.8.4.4
bridge_ports eth0
bridge_stp off
bridge_fd 0
```

* **For a DHCP IP**

```
## DHCP ip config file for br0 ##
auto br0
iface br0 inet dhcp
bridge_ports eth0
```

4. **Ensure the physical interface is not configured**

    Verify that the physical interface (e.g., eth0) is not configured in the main /etc/network/interfaces file. It should be managed solely by the bridge.
5. **Restart the networking service**

```bash
sudo systemctl restart networking
```

6. **Verify the bridge**

   Confirm the bridge was created successfully using the brctl or bridge command:

```bash
brctl show
# or
bridge link
```

Other tutorial sites:
- https://www.cyberciti.biz/faq/how-to-configuring-bridging-in-debian-linux/

### Create ssh key

```bash
# Gerar uma nova chave SSH ed25519 (recomendado)
ssh-keygen -t ed25519 -C "seu-email@exemplo.com"

# Por padrão, a chave será salva em ~/.ssh/id_ed25519
# Pressione Enter para aceitar o local padrão e defina uma senha se desejar
```

## 🚀 Quick Start

1. **Install**

```bash
# Linux
curl -L https://github.com/paulozagaloneves/kvm-compose/releases/download/0.1.0/kvm-compose-linux-amd64 -o kvm-compose

chmod +x kvm-compose
sudo mv ./kvm-compose /usr/local/bin/kvm-compose
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
- **memory**: RAM in MB (default: 2048)
- **vcpus**: Number of virtual CPUs (default: 2)
- **disk_size**: Disk size in GB (default: 2)
- **username**: SSH user (default from config.ini or "debian")
- **ssh_key_file**: Path to SSH public key (default from config.ini)
- **networks**: Network configuration
  - **bridge**: Network bridge (default: br0)
  - **ipv4**: Static IP address
  - **gateway**: Network gateway (default from config.ini)
  - **nameservers**: DNS servers array (default from config.ini)

## ⚙️ Configuration File (config.ini)

O kvm-compose agora suporta um arquivo de configuração opcional que define valores padrão. O arquivo é procurado em:

1. `./config.ini` (diretório atual)
2. `~/.config/kvm-compose/config.ini` (diretório do usuário)

### Exemplo de config.ini:

```ini
[main]
username = debian
ssh_key_file = ~/.ssh/id_ed25519.pub

[network]
gateway = 192.168.1.1
nameservers = 1.1.1.1, 8.8.8.8

[images]
path_upstream_images = ~/.config/kvm-compose/images/upstream
path_vm_images = ~/.config/kvm-compose/images/vm
```

### Benefícios da Configuração:

- 📂 **Organização**: Imagens separadas por tipo (base vs VMs)
- 🔧 **Defaults**: Valores padrão configuráveis por projeto/usuário
- 🏠 **Diretórios**: Imagens organizadas em ~/.config/kvm-compose/
- ♻️ **Reutilização**: Imagens base compartilhadas entre projetos

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

1. **Clone and build:**

```bash
git clone https://github.com/yourusername/kvm-compose.git
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

© 2025 Paulo Neves. Todos os direitos reservados.