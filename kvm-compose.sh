#!/bin/bash

# Parâmetros padrão
ACTION=""
SERVER_NAME=""
COMPOSE_FILE="kvm-compose.yaml"

# Função para ler configurações do YAML
read_vm_config() {
    local vm_name="$1"
    
    if [[ ! -f "${COMPOSE_FILE}" ]]; then
        echo "Erro: Arquivo ${COMPOSE_FILE} não encontrado."
        exit 1
    fi
    
    # Verificar se a VM existe no arquivo
    if ! grep -q "name: ${vm_name}" "${COMPOSE_FILE}"; then
        echo "Erro: VM '${vm_name}' não encontrada no arquivo ${COMPOSE_FILE}."
        echo "VMs disponíveis:"
        get_all_vms
        exit 1
    fi
    
    # Usar awk para extrair configurações de forma eficiente
    local vm_config=$(awk -v vm="$vm_name" '
    BEGIN { in_vm=0; in_networks=0 }
    /^- name: / {
        if ($3 == vm) {
            in_vm=1
            in_networks=0
            next
        } else if (in_vm) {
            exit
        }
    }
    in_vm && /^- name: / && $3 != vm { exit }
    in_vm && /^  memory: / { print "MEMORY=" $2 }
    in_vm && /^  vcpus: / { print "VCPUS=" $2 }
    in_vm && /^  disk_size: / { print "DISK_SIZE=" $2 }
    in_vm && /^  username: / { print "USERNAME=" $2 }
    in_vm && /^  ssh_key_file: / { print "SSH_KEY_FILE=" $2 }
    in_vm && /^  networks:/ { in_networks=1; next }
    in_vm && in_networks && /^    - bridge: / { print "BRIDGE=" $3 }
    in_vm && in_networks && /^      ipv4: / { print "IP_ADDRESS=" $2 }
    in_vm && in_networks && /^      gateway: / { print "GATEWAY=" $2 }
    in_vm && in_networks && /^      nameservers: / {
        gsub(/\[|\]|,/, "", $0)
        sub(/^.*nameservers: /, "", $0)
        gsub(/ /, ",", $0)
        print "NAMESERVERS=" $0
    }
    ' "${COMPOSE_FILE}")
    
    # Avaliar as variáveis extraídas
    eval "$vm_config"
    
    # Expandir ~ no caminho da chave SSH
    SSH_KEY_FILE="${SSH_KEY_FILE/#\~/$HOME}"
    
    # Ler chave SSH se especificada
    if [[ -n "$SSH_KEY_FILE" && -f "$SSH_KEY_FILE" ]]; then
        SSH_KEY_CONTENT=$(cat "$SSH_KEY_FILE")
    else
        # Chave padrão se não especificada
        echo "Aviso: Chave SSH não especificada ou arquivo não encontrado. Usando chave padrão."
    fi
    
    # Valores padrão se não especificados
    MEMORY=${MEMORY:-4096}
    VCPUS=${VCPUS:-4}
    DISK_SIZE=${DISK_SIZE:-20}
    BRIDGE=${BRIDGE:-br0}
    GATEWAY=${GATEWAY:-192.168.1.1}
    NAMESERVERS=${NAMESERVERS:-"1.1.1.1,8.8.8.8"}
    USERNAME=${USERNAME:-"debian"}
}

# Função para exibir ajuda
show_help() {
    echo "Uso: $0 <ação> [opções]"
    echo ""
    echo "Ações:"
    echo "  up      Criar e iniciar todas as VMs do compose"
    echo "  start   Iniciar todas as VMs do compose"
    echo "  stop    Parar todas as VMs do compose"
    echo "  down    Destruir todas as VMs do compose"
    echo "  list    Listar VMs disponíveis no compose"
    echo ""
    echo "Opções:"
    echo "  --compose <arquivo>     Arquivo compose (padrão: lab-compose.yaml)"
    echo "  --help                  Exibir esta ajuda"
    echo ""
    echo "Nota: O script opera em todas as VMs definidas no arquivo lab-compose.yaml."
    echo ""
    echo "Exemplos:"
    echo "  $0 up      # Cria todas as VMs"
    echo "  $0 start   # Inicia todas as VMs"
    echo "  $0 stop    # Para todas as VMs" 
    echo "  $0 down    # Destrói todas as VMs"
    echo "  $0 list    # Lista todas as VMs"
}

# Função para obter todas as VMs do compose
get_all_vms() {
    if [[ ! -f "${COMPOSE_FILE}" ]]; then
        echo "Erro: Arquivo ${COMPOSE_FILE} não encontrado."
        exit 1
    fi
    
    grep "^- name:" "${COMPOSE_FILE}" | sed 's/- name: //'
}

# Funções para cada ação
vm_up() {
    echo "=== Criando todas as VMs do compose ==="
    
    # Baixar a imagem cloud do Debian 13 (se não existir)
    if [ ! -f "debian-13-genericcloud-amd64.qcow2" ]; then
        echo "Baixando imagem base do Debian 13..."
        wget -O debian-13-genericcloud-amd64.qcow2 https://cdimage.debian.org/images/cloud/trixie/latest/debian-13-genericcloud-amd64.qcow2
    else
        echo "Imagem base já existe, pulando download..."
    fi
    echo ""
    
    # Processar cada VM do compose
    local vms=($(get_all_vms))
    local created_count=0
    local skipped_count=0
    
    for vm_name in "${vms[@]}"; do
        echo "--- Processando VM: $vm_name ---"
        
        # Verificar se VM já existe
        if virsh dominfo "$vm_name" >/dev/null 2>&1; then
            echo "⚠️  VM $vm_name já existe, pulando..."
            ((skipped_count++))
            echo ""
            continue
        fi
        
        # Ler configurações da VM
        read_vm_config "$vm_name"
        
        echo "🛠️ Configurações:"
        echo "  Usuário: ${USERNAME}"
        echo "  IP: ${IP_ADDRESS}"
        echo "  Memória: ${MEMORY}MB"
        echo "  vCPUs: ${VCPUS}"
        echo "  Disco: ${DISK_SIZE}GB"
        echo "  Bridge: ${BRIDGE}"
        echo "  Nameservers: ${NAMESERVERS}"
        
        # Copiar a imagem para criar o disco da VM
        cp debian-13-genericcloud-amd64.qcow2 ${vm_name}.qcow2
        
        # Ajustar permissões do arquivo de disco
        sudo chmod 644 ${vm_name}.qcow2

        
        # Criar arquivos de cloud-init temporários
        cat > ${vm_name}-user-data.yaml <<EOF
#cloud-config
users:
  - name: ${USERNAME}
    ssh_authorized_keys:
      - ${SSH_KEY_CONTENT}
    sudo: ['ALL=(ALL) NOPASSWD:ALL']
    shell: /bin/bash
    lock_passwd: false
EOF
    
#      cat ${vm_name}-user-data.yaml

        # Converter nameservers para formato YAML
        local ns_yaml=""
        IFS=',' read -ra NS_ARRAY <<< "${NAMESERVERS}"
        for ns in "${NS_ARRAY[@]}"; do
            ns=$(echo "$ns" | xargs)  # trim whitespace
            ns_yaml+="        - ${ns}\n"
        done

        cat > ${vm_name}-network-config.yaml <<EOF
version: 2
ethernets:
  enp1s0:
    dhcp4: false
    addresses: 
      - ${IP_ADDRESS}/24
    gateway4: ${GATEWAY}
    nameservers:
      addresses:
$(echo -e "${ns_yaml}")
EOF

#cat  ${vm_name}-network-config.yaml

        cat > ${vm_name}-meta-data.yaml <<EOF
instance-id: ${vm_name}
local-hostname: ${vm_name}
EOF

#cat  ${vm_name}-meta-data.yaml
        # Executar virt-install
        echo "Criando VM ${vm_name}..."
        if virt-install \
            --name ${vm_name} \
            --memory ${MEMORY} \
            --vcpus ${VCPUS} \
            --os-variant debian13 \
            --disk ${vm_name}.qcow2,size=${DISK_SIZE},format=qcow2 \
            --network bridge=${BRIDGE},model=virtio \
            --graphics spice,listen=0.0.0.0 \
            --noautoconsole \
            --import \
            --cloud-init user-data=${vm_name}-user-data.yaml,network-config=${vm_name}-network-config.yaml,meta-data=${vm_name}-meta-data.yaml; then
            
            echo "✅ VM ${vm_name} criada com sucesso!"
            echo "   SSH: ssh ${USERNAME}@${IP_ADDRESS}"
            ((created_count++))
        else
            echo "❌ Falha ao criar VM ${vm_name}"
        fi
        
        # Limpar arquivos temporários de cloud-init
        rm -f ${vm_name}-user-data.yaml ${vm_name}-network-config.yaml ${vm_name}-meta-data.yaml
        echo ""
    done
    
    echo "=== Resumo ==="
    echo "VMs criadas: $created_count"
    echo "VMs puladas (já existem): $skipped_count"
    echo "Total de VMs no compose: ${#vms[@]}"
}

vm_start() {
    echo "=== Iniciando todas as VMs do compose ==="
    
    local vms=($(get_all_vms))
    local started_count=0
    local running_count=0
    local missing_count=0
    
    for vm_name in "${vms[@]}"; do
        echo "--- Iniciando VM: $vm_name ---"
        
        if ! virsh dominfo "$vm_name" >/dev/null 2>&1; then
            echo "⚠️  VM $vm_name não existe. Use 'up' para criar."
            ((missing_count++))
        elif virsh domstate "$vm_name" | grep -q "running"; then
            echo "🟢 VM $vm_name já está em execução."
            ((running_count++))
        else
            if virsh start "$vm_name"; then
                echo "✅ VM $vm_name iniciada com sucesso!"
                ((started_count++))
            else
                echo "❌ Falha ao iniciar VM $vm_name"
            fi
        fi
        echo ""
    done
    
    echo "=== Resumo ==="
    echo "VMs iniciadas: $started_count"
    echo "VMs já rodando: $running_count"
    echo "VMs não existem: $missing_count"
    echo "Total de VMs no compose: ${#vms[@]}"
}

vm_stop() {
    echo "=== Parando todas as VMs do compose ==="
    
    local vms=($(get_all_vms))
    local stopped_count=0
    local already_stopped_count=0
    local missing_count=0
    
    for vm_name in "${vms[@]}"; do
        echo "--- Parando VM: $vm_name ---"
        
        if ! virsh dominfo "$vm_name" >/dev/null 2>&1; then
            echo "⚠️  VM $vm_name não existe."
            ((missing_count++))
        elif virsh domstate "$vm_name" | grep -q "shut off"; then
            echo "🔴 VM $vm_name já está parada."
            ((already_stopped_count++))
        else
            if virsh shutdown "$vm_name"; then
                echo "✅ VM $vm_name parada com sucesso!"
                ((stopped_count++))
            else
                echo "❌ Falha ao parar VM $vm_name"
            fi
        fi
        echo ""
    done
    
    echo "=== Resumo ==="
    echo "VMs paradas: $stopped_count"
    echo "VMs já paradas: $already_stopped_count"
    echo "VMs não existem: $missing_count"
    echo "Total de VMs no compose: ${#vms[@]}"
}

vm_down() {
    echo "=== Destruindo todas as VMs do compose ==="
    
    local vms=($(get_all_vms))
    local destroyed_count=0
    local missing_count=0
    
    for vm_name in "${vms[@]}"; do
        echo "--- Destruindo VM: $vm_name ---"
        
        if ! virsh dominfo "$vm_name" >/dev/null 2>&1; then
            echo "⚠️  VM $vm_name não existe."
            ((missing_count++))
        else
            # Parar VM se estiver rodando
            if virsh domstate "$vm_name" | grep -q "running"; then
                echo "Parando VM $vm_name..."
                virsh destroy "$vm_name"
            fi
            
            # Remover VM
            if virsh undefine "$vm_name"; then
                echo "✅ VM $vm_name removida do libvirt"
                ((destroyed_count++))
            else
                echo "❌ Falha ao remover VM $vm_name do libvirt"
            fi
        fi
        
        # Remover arquivo de disco se existir
        if [ -f "${vm_name}.qcow2" ]; then
            rm "${vm_name}.qcow2"
            echo "💾 Arquivo de disco ${vm_name}.qcow2 removido"
        fi
        echo ""
    done
    
    echo "=== Resumo ==="
    echo "VMs destruídas: $destroyed_count"
    echo "VMs não existiam: $missing_count"
    echo "Total de VMs no compose: ${#vms[@]}"
}

vm_list() {
    echo "=== VMs disponíveis no ${COMPOSE_FILE} ==="
    
    if [[ ! -f "${COMPOSE_FILE}" ]]; then
        echo "Erro: Arquivo ${COMPOSE_FILE} não encontrado."
        exit 1
    fi
    
    echo -e "Nome\t\tMemória\tvCPUs\tDisco\tIP\t\tStatus"
    echo -e "----\t\t-------\t-----\t-----\t--\t\t------"
    
    # Obter lista de VMs e suas configurações
    local vms=($(get_all_vms))
    
    for vm_name in "${vms[@]}"; do
        # Ler configurações da VM
        read_vm_config "$vm_name"
        
        # Verificar status da VM
        local status="not created"
        if virsh dominfo "$vm_name" >/dev/null 2>&1; then
            status=$(virsh domstate "$vm_name" 2>/dev/null)
            case "$status" in
                "running")
                    status="🟢 running"
                    ;;
                "shut off")
                    status="🔴 stopped"
                    ;;
                "paused")
                    status="🟡 paused"
                    ;;
                "suspended")
                    status="🟠 suspended"
                    ;;
                *)
                    status="❓ $status"
                    ;;
            esac
        else
            status="⚪ not created"
        fi
        
        # Imprimir linha da VM com status
        printf "%-15s\t%sMB\t%s\t%sGB\t%-15s\t%s\n" \
            "$vm_name" \
            "${MEMORY:-4096}" \
            "${VCPUS:-4}" \
            "${DISK_SIZE:-20}" \
            "${IP_ADDRESS:-N/A}" \
            "$status"
    done
}

# Parse dos argumentos da linha de comando
if [[ $# -eq 0 ]]; then
    show_help
    exit 1
fi

# Primeiro argumento deve ser a ação
ACTION="$1"
shift

# Parse das opções restantes
while [[ $# -gt 0 ]]; do
    case $1 in
        --compose)
            COMPOSE_FILE="$2"
            shift 2
            ;;
        --help)
            show_help
            exit 0
            ;;
        *)
            echo "Opção desconhecida: $1"
            show_help
            exit 1
            ;;
    esac
done

# Validar ação
case "${ACTION}" in
    up|start|stop|down|list)
        ;;
    *)
        echo "Erro: Ação '${ACTION}' não reconhecida."
        show_help
        exit 1
        ;;
esac

# Validar se arquivo compose existe
if [[ ! -f "${COMPOSE_FILE}" ]]; then
    echo "Erro: Arquivo ${COMPOSE_FILE} não encontrado."
    exit 1
fi

echo -e "\033[36m============================================================\033[0m"
echo -e "\033[1;32m🖥️  kvm-compose.sh - Gerenciador de VMs KVM via arquivo compose\033[0m"
echo -e "\033[1;33m📦 Versão: 0.1.0 Codename: \"Gambiarra\" - Dezembro de 2025\033[0m"
echo -e "\033[36m============================================================\033[0m"
echo ""
echo "Configuração:"
echo "  Ação: ${ACTION}"
echo "  Compose: ${COMPOSE_FILE}"
echo ""

# Executar ação correspondente
case "${ACTION}" in
    up)
        vm_up
        ;;
    start)
        vm_start
        ;;
    stop)
        vm_stop
        ;;
    down)
        vm_down
        ;;
    list)
        vm_list
        ;;
esac


