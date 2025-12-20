
# KVM Compose

Versão 0.2.0 Codinome: "Gambiarra" - Dezembro de 2025

🖥️ **kvm-compose** é uma ferramenta moderna escrita em **Go** que simplifica o gerenciamento de máquinas virtuais KVM usando fluxos de trabalho similares ao Docker Compose.


## Funcionalidades

- Crie, inicie, pare e gerencie VMs KVM facilmente.
- Configuração declarativa para VMs.
- Fluxo de trabalho simplificado para desenvolvimento e testes.

---

## 📋 1. Pré-requisitos

- Linux com suporte ao KVM habilitado
- `qemu-kvm`, `libvirt-clients` e `virtinst` instalados
- Bridge de rede configurada (padrão: `br0`)
- `Go 1.21+` (para compilação)
- `wget` para baixar imagens base
- Par de chaves SSH configurado


# 🚀 Início Rápido

1. **Instalação**

```bash
# Linux
curl -L https://github.com/paulozagaloneves/kvm-compose/releases/download/0.1.0/kvm-compose-linux-amd64 -o kvm-compose

chmod +x kvm-compose
sudo mv ./kvm-compose /usr/local/bin/kvm-compose
```

2. **Edite a configuração:**

  - Crie ou modifique o arquivo `kvm-compose.yaml` para definir suas VMs.


## 🔧 Exemplo de Configuração

Aqui está um exemplo simples de arquivo `kvm-compose.yaml`:

```yaml
# Control plane do Kubernetes
- name: k8s-cp-01
  distro: debian-13
  memory: 4096
  vcpus: 4
  disk_size: 20
  username: debian
  ssh_key_file: ~/.ssh/id_ed25519.pub
  networks:
    - host_bridge: br0
      guest_ipv4: 192.168.1.40
      guest_gateway4: 192.168.1.1
      guest_nameservers: [1.1.1.1, 8.8.8.8]

# Nó worker do Kubernetes
- name: k8s-wrk-01
  distro: debian-13
  memory: 2048
  vcpus: 2
  disk_size: 15
  username: debian
  ssh_key_file: ~/.ssh/id_ed25519.pub
  networks:
    - host_bridge: br0
      guest_ipv4: 192.168.1.41
      guest_gateway4: 192.168.1.1
      guest_nameservers: [1.1.1.1, 8.8.8.8]
```


### Parâmetros de Configuração

- **name**: Identificador da VM (obrigatório)
- **memory**: RAM em MB (padrão: 2048)
- **vcpus**: Número de CPUs virtuais (padrão: 2)
- **disk_size**: Tamanho do disco em GB (padrão: 2)
- **username**: Usuário SSH (padrão do config.ini ou "debian")
- **ssh_key_file**: Caminho para a chave pública SSH (padrão no config.ini)
- **networks**: Configuração de rede
  - **host_bridge**: Bridge de rede do host (padrão: br0)
  - **guest_ipv4**: IP estático da VM
  - **guest_gateway4**: Gateway da rede da VM (padrão no config.ini)
  - **guest_nameservers**: Array de servidores DNS da VM (padrão no config.ini)


## ⚙️ Arquivo de Configuração (config.ini)

O kvm-compose agora suporta um arquivo de configuração opcional que define valores padrão. O arquivo é procurado em:

1. `./config.ini` (diretório atual)
2. `~/.config/kvm-compose/config.ini` (diretório config do usuário)

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
- 🔧 **Padrões**: Valores padrão configuráveis por projeto/usuário
- 🏠 **Diretórios**: Imagens organizadas em ~/.config/kvm-compose/
- ♻️ **Reutilização**: Imagens base compartilhadas entre projetos


## 🎯 Comandos Disponíveis

- 🆙 `up` - Cria e inicia todas as VMs definidas no arquivo compose
- ▶️ `start` - Inicia VMs existentes
- ⏹️ `stop` - Para VMs em execução (desligamento gracioso)
- ⬇️ `down` - Remove VMs e apaga arquivos de disco
- 📋 `status` - Mostra configuração e status das VMs com saída colorida
- 💻 `ssh` - Acede ao shell da VM definida 


## 💡 Exemplos de Uso

```bash
# Usando o binário instalado
kvm-compose up
kvm-compose list  
kvm-compose stop
kvm-compose down
kvm-compose ssh <vmname>

# Usando arquivo compose customizado
kvm-compose up --compose meu-lab.yaml

# Usando targets do Make para desenvolvimento
make run-up      # Compila e executa 'up'
make run-list    # Compila e executa 'list'  
make run-down    # Compila e executa 'down'

# Build e desenvolvimento
make build       # Compila o binário
make clean       # Limpa arquivos de build
make install     # Instala no sistema
make uninstall   # Remove do sistema
```


## 🎨 Melhorias Visuais

A versão em Go inclui saída colorida e emojis para melhor experiência:

- 🟢 VMs executando
- 🔴 VMs paradas
- 🟡 VMs pausadas
- ⚪ VMs não criadas
- ✅ Operações bem-sucedidas
- ❌ Erros e falhas
- ⚠️ Avisos importantes



### 🐧 1.1 Instalar KVM no Ubuntu/Debian

```bash
# Instale o KVM e dependências
sudo apt update
sudo apt install -y qemu-kvm libvirt-daemon libvirt-clients bridge-utils virt-manager virtinst cloud-image-utils wget
```

Para uso sem privilégios de root, adicione seu usuário aos grupos libvirt e kvm:

```bash
sudo usermod -aG libvirt,kvm $USER
```

Outros tutoriais:
- https://cloudspinx.com/install-kvm-on-debian-with-virt-manager-and-cockpit/
- https://sysguides.com/install-kvm-on-linux
- https://phoenixnap.com/kb/ubuntu-install-kvm

## 1.2 **Resolução Local de Nomes das VMs**

Para resolver os nomes das VMs localmente:

1. Instale o pacote `libnss-libvirt`:

  ```bash
  sudo apt install libnss-libvirt
  ```

2. Edite o arquivo `/etc/nsswitch.conf`, adicionando `libvirt libvirt_guest` na linha de `hosts`:

  ```
  hosts: files libvirt libvirt_guest dns
  ```

Agora você pode acessar as VMs via SSH usando o nome da máquina.

---



### 🔧 2. Configurar bridge de rede no Debian

1. **Instale os utilitários necessários**

```bash
sudo apt update
sudo apt install bridge-utils
```

2. **Identifique sua interface de rede física**

```bash
ip -f inet a s
```

3. **Configure a bridge**

Crie um arquivo de configuração para a bridge no diretório /etc/network/interfaces.d/. Por exemplo, crie um arquivo chamado br0:

```bash
sudo nano /etc/network/interfaces.d/br0
```

Adicione a configuração abaixo, substituindo **eth0** pelo nome real da sua interface e ajustando os parâmetros de IP conforme necessário:

* **Para IP estático:**


```
## Arquivo de configuração IP estático para br0 ##
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

* **Para IP dinâmico (DHCP):**

```
## Arquivo de configuração DHCP para br0 ##
auto br0
iface br0 inet dhcp
bridge_ports eth0
```

4. **Garanta que a interface física não está configurada**

  Verifique se a interface física (ex: eth0) não está configurada no arquivo principal /etc/network/interfaces. Ela deve ser gerenciada apenas pela bridge.

5. **Reinicie o serviço de rede**

```bash
sudo systemctl restart networking
```

6. **Verifique a bridge**

  Confirme que a bridge foi criada com sucesso usando o comando brctl ou bridge:

```bash
brctl show
# ou
bridge link
```

Outros tutoriais:
- https://www.cyberciti.biz/faq/how-to-configuring-bridging-in-debian-linux/


### 🛡️ 3. Criar chave SSH

```bash
# Gerar uma nova chave SSH ed25519 (recomendado)
ssh-keygen -t ed25519 -C "seu-email@exemplo.com"

# Por padrão, a chave será salva em ~/.ssh/id_ed25519
# Pressione Enter para aceitar o local padrão e defina uma senha se desejar
```


## 🏗️ Desenvolvimento

Para contribuir ou modificar o código:

1. **Clone e faça o build:**

```bash
git clone https://github.com/seuusuario/kvm-compose.git
cd kvm-compose
    
# Instale as dependências
make deps

# Desenvolvimento
make build       # Build local
make test        # Executa testes
make clean       # Limpa artefatos de build

# Teste local sem instalar
./build/kvm-compose --help
```


## Licença

Licença Pública Geral GNU Versão 3

---

© 2025 Paulo Neves. Todos os direitos reservados.