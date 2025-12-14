package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// VM representa uma máquina virtual no arquivo de configuração
type VM struct {
	Name        string    `yaml:"name"`
	Distro      string    `yaml:"distro"`
	Memory      int       `yaml:"memory"`
	VCPUs       int       `yaml:"vcpus"`
	DiskSize    int       `yaml:"disk_size"`
	Username    string    `yaml:"username"`
	Group       []string  `yaml:"group"`
	SSHKeyFile  string    `yaml:"ssh_key_file"`
	Networks    []Network `yaml:"networks"`
}

// Network representa a configuração de rede de uma VM
type Network struct {
	Bridge      string   `yaml:"bridge"`
	IPv4        string   `yaml:"ipv4"`
	Gateway     string   `yaml:"gateway"`
	Nameservers []string `yaml:"nameservers"`
}

// Config representa o arquivo de configuração completo
type Config struct {
	VMs []VM `yaml:",inline"`
}

// KVMCompose é a estrutura principal do aplicativo
type KVMCompose struct {
	composeFile string
	config      Config
}

// NewKVMCompose cria uma nova instância do KVMCompose
func NewKVMCompose(composeFile string) *KVMCompose {
	return &KVMCompose{
		composeFile: composeFile,
	}
}

// loadConfig carrega o arquivo YAML de configuração
func (kvm *KVMCompose) loadConfig() error {
	data, err := ioutil.ReadFile(kvm.composeFile)
	if err != nil {
		return fmt.Errorf("erro ao ler arquivo %s: %v", kvm.composeFile, err)
	}

	err = yaml.Unmarshal(data, &kvm.config.VMs)
	if err != nil {
		return fmt.Errorf("erro ao fazer parse do YAML: %v", err)
	}

	return nil
}

// getVMByName busca uma VM pelo nome
func (kvm *KVMCompose) getVMByName(name string) (*VM, error) {
	for _, vm := range kvm.config.VMs {
		if vm.Name == name {
			return &vm, nil
		}
	}
	return nil, fmt.Errorf("VM '%s' não encontrada", name)
}

// execCommand executa um comando do sistema
func execCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// execCommandOutput executa um comando e retorna a saída
func execCommandOutput(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.Output()
	return strings.TrimSpace(string(output)), err
}

// vmExists verifica se uma VM existe no libvirt
func vmExists(name string) bool {
	_, err := execCommandOutput("virsh", "dominfo", name)
	return err == nil
}

// getVMState retorna o estado atual de uma VM
func getVMState(name string) (string, error) {
	if !vmExists(name) {
		return "not created", nil
	}
	return execCommandOutput("virsh", "domstate", name)
}

// expandPath expande o ~ no caminho para o home directory
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

// readSSHKey lê o conteúdo da chave SSH
func readSSHKey(keyFile string) (string, error) {
	expandedPath := expandPath(keyFile)
	content, err := ioutil.ReadFile(expandedPath)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(content)), nil
}

// createCloudInitFiles cria os arquivos cloud-init para uma VM
func (kvm *KVMCompose) createCloudInitFiles(vm *VM) error {
	// Aplicar valores padrão
	if vm.Memory == 0 {
		vm.Memory = 4096
	}
	if vm.VCPUs == 0 {
		vm.VCPUs = 4
	}
	if vm.DiskSize == 0 {
		vm.DiskSize = 20
	}
	if vm.Username == "" {
		vm.Username = "debian"
	}

	// Ler chave SSH
	sshKey := ""
	if vm.SSHKeyFile != "" {
		key, err := readSSHKey(vm.SSHKeyFile)
		if err != nil {
			color.Yellow("⚠️  Aviso: Não foi possível ler a chave SSH: %v", err)
		} else {
			sshKey = key
		}
	}

	// Criar user-data
	userData := fmt.Sprintf(`#cloud-config
users:
  - name: %s
    ssh_authorized_keys:
      - %s
    sudo: ['ALL=(ALL) NOPASSWD:ALL']
    shell: /bin/bash
    lock_passwd: false
`, vm.Username, sshKey)

	err := ioutil.WriteFile(vm.Name+"-user-data.yaml", []byte(userData), 0644)
	if err != nil {
		return err
	}

	// Criar network-config
	network := vm.Networks[0] // Assumindo apenas uma rede por VM
	if network.Bridge == "" {
		network.Bridge = "br0"
	}
	if network.Gateway == "" {
		network.Gateway = "192.168.1.1"
	}
	if len(network.Nameservers) == 0 {
		network.Nameservers = []string{"1.1.1.1", "8.8.8.8"}
	}

	nsYAML := ""
	for _, ns := range network.Nameservers {
		nsYAML += fmt.Sprintf("        - %s\n", ns)
	}

	networkConfig := fmt.Sprintf(`version: 2
ethernets:
  enp1s0:
    dhcp4: false
    addresses: 
      - %s/24
    gateway4: %s
    nameservers:
      addresses:
%s`, network.IPv4, network.Gateway, nsYAML)

	err = ioutil.WriteFile(vm.Name+"-network-config.yaml", []byte(networkConfig), 0644)
	if err != nil {
		return err
	}

	// Criar meta-data
	metaData := fmt.Sprintf(`instance-id: %s
local-hostname: %s
`, vm.Name, vm.Name)

	err = ioutil.WriteFile(vm.Name+"-meta-data.yaml", []byte(metaData), 0644)
	return err
}

// cleanupCloudInitFiles remove os arquivos temporários de cloud-init
func cleanupCloudInitFiles(vmName string) {
	os.Remove(vmName + "-user-data.yaml")
	os.Remove(vmName + "-network-config.yaml")
	os.Remove(vmName + "-meta-data.yaml")
}

// downloadBaseImage baixa a imagem base do Debian se não existir
func downloadBaseImage() error {
	imageName := "debian-13-genericcloud-amd64.qcow2"
	if _, err := os.Stat(imageName); os.IsNotExist(err) {
		color.Cyan("📥 Baixando imagem base do Debian 13...")
		url := "https://cdimage.debian.org/images/cloud/trixie/latest/debian-13-genericcloud-amd64.qcow2"
		return execCommand("wget", "-O", imageName, url)
	}
	color.Green("✅ Imagem base já existe, pulando download...")
	return nil
}

// Up cria e inicia todas as VMs
func (kvm *KVMCompose) Up() error {
	err := kvm.loadConfig()
	if err != nil {
		return err
	}

	color.Cyan("=== Criando todas as VMs do compose ===")

	// Baixar imagem base
	if err := downloadBaseImage(); err != nil {
		return fmt.Errorf("erro ao baixar imagem base: %v", err)
	}

	createdCount := 0
	skippedCount := 0

	for _, vm := range kvm.config.VMs {
		color.White("--- Processando VM: %s ---", vm.Name)

		// Verificar se VM já existe
		if vmExists(vm.Name) {
			color.Yellow("⚠️  VM %s já existe, pulando...", vm.Name)
			skippedCount++
			fmt.Println()
			continue
		}

		// Mostrar configurações
		color.Blue("🛠️ Configurações:")
		fmt.Printf("  Usuário: %s\n", vm.Username)
		fmt.Printf("  IP: %s\n", vm.Networks[0].IPv4)
		fmt.Printf("  Memória: %dMB\n", vm.Memory)
		fmt.Printf("  vCPUs: %d\n", vm.VCPUs)
		fmt.Printf("  Disco: %dGB\n", vm.DiskSize)
		fmt.Printf("  Bridge: %s\n", vm.Networks[0].Bridge)

		// Copiar imagem
		err := execCommand("cp", "debian-13-genericcloud-amd64.qcow2", vm.Name+".qcow2")
		if err != nil {
			color.Red("❌ Erro ao copiar imagem para %s: %v", vm.Name, err)
			continue
		}

		// Ajustar permissões
		execCommand("sudo", "chmod", "644", vm.Name+".qcow2")

		// Criar arquivos cloud-init
		err = kvm.createCloudInitFiles(&vm)
		if err != nil {
			color.Red("❌ Erro ao criar arquivos cloud-init para %s: %v", vm.Name, err)
			continue
		}

		// Executar virt-install
		color.Cyan("🚀 Criando VM %s...", vm.Name)
		args := []string{
			"--name", vm.Name,
			"--memory", fmt.Sprintf("%d", vm.Memory),
			"--vcpus", fmt.Sprintf("%d", vm.VCPUs),
			"--os-variant", "debian13",
			"--disk", fmt.Sprintf("%s.qcow2,size=%d,format=qcow2", vm.Name, vm.DiskSize),
			"--network", fmt.Sprintf("bridge=%s,model=virtio", vm.Networks[0].Bridge),
			"--graphics", "spice,listen=0.0.0.0",
			"--noautoconsole",
			"--import",
			"--cloud-init",
			fmt.Sprintf("user-data=%s-user-data.yaml,network-config=%s-network-config.yaml,meta-data=%s-meta-data.yaml",
				vm.Name, vm.Name, vm.Name),
		}

		if err := execCommand("virt-install", args...); err != nil {
			color.Red("❌ Falha ao criar VM %s: %v", vm.Name, err)
		} else {
			color.Green("✅ VM %s criada com sucesso!", vm.Name)
			color.Cyan("   SSH: ssh %s@%s", vm.Username, vm.Networks[0].IPv4)
			createdCount++
		}

		// Limpar arquivos temporários
		cleanupCloudInitFiles(vm.Name)
		fmt.Println()
	}

	color.Cyan("=== Resumo ===")
	fmt.Printf("VMs criadas: %d\n", createdCount)
	fmt.Printf("VMs puladas (já existem): %d\n", skippedCount)
	fmt.Printf("Total de VMs no compose: %d\n", len(kvm.config.VMs))

	return nil
}

// Start inicia todas as VMs
func (kvm *KVMCompose) Start() error {
	err := kvm.loadConfig()
	if err != nil {
		return err
	}

	color.Cyan("=== Iniciando todas as VMs do compose ===")

	startedCount := 0
	runningCount := 0
	missingCount := 0

	for _, vm := range kvm.config.VMs {
		color.White("--- Iniciando VM: %s ---", vm.Name)

		if !vmExists(vm.Name) {
			color.Yellow("⚠️  VM %s não existe. Use 'up' para criar.", vm.Name)
			missingCount++
		} else {
			state, _ := getVMState(vm.Name)
			if state == "running" {
				color.Green("🟢 VM %s já está em execução.", vm.Name)
				runningCount++
			} else {
				if err := execCommand("virsh", "start", vm.Name); err != nil {
					color.Red("❌ Falha ao iniciar VM %s: %v", vm.Name, err)
				} else {
					color.Green("✅ VM %s iniciada com sucesso!", vm.Name)
					startedCount++
				}
			}
		}
		fmt.Println()
	}

	color.Cyan("=== Resumo ===")
	fmt.Printf("VMs iniciadas: %d\n", startedCount)
	fmt.Printf("VMs já rodando: %d\n", runningCount)
	fmt.Printf("VMs não existem: %d\n", missingCount)
	fmt.Printf("Total de VMs no compose: %d\n", len(kvm.config.VMs))

	return nil
}

// Stop para todas as VMs
func (kvm *KVMCompose) Stop() error {
	err := kvm.loadConfig()
	if err != nil {
		return err
	}

	color.Cyan("=== Parando todas as VMs do compose ===")

	stoppedCount := 0
	alreadyStoppedCount := 0
	missingCount := 0

	for _, vm := range kvm.config.VMs {
		color.White("--- Parando VM: %s ---", vm.Name)

		if !vmExists(vm.Name) {
			color.Yellow("⚠️  VM %s não existe.", vm.Name)
			missingCount++
		} else {
			state, _ := getVMState(vm.Name)
			if state == "shut off" {
				color.Red("🔴 VM %s já está parada.", vm.Name)
				alreadyStoppedCount++
			} else {
				if err := execCommand("virsh", "shutdown", vm.Name); err != nil {
					color.Red("❌ Falha ao parar VM %s: %v", vm.Name, err)
				} else {
					color.Green("✅ VM %s parada com sucesso!", vm.Name)
					stoppedCount++
				}
			}
		}
		fmt.Println()
	}

	color.Cyan("=== Resumo ===")
	fmt.Printf("VMs paradas: %d\n", stoppedCount)
	fmt.Printf("VMs já paradas: %d\n", alreadyStoppedCount)
	fmt.Printf("VMs não existem: %d\n", missingCount)
	fmt.Printf("Total de VMs no compose: %d\n", len(kvm.config.VMs))

	return nil
}

// Down destrói todas as VMs
func (kvm *KVMCompose) Down() error {
	err := kvm.loadConfig()
	if err != nil {
		return err
	}

	color.Cyan("=== Destruindo todas as VMs do compose ===")

	destroyedCount := 0
	missingCount := 0

	for _, vm := range kvm.config.VMs {
		color.White("--- Destruindo VM: %s ---", vm.Name)

		if !vmExists(vm.Name) {
			color.Yellow("⚠️  VM %s não existe.", vm.Name)
			missingCount++
		} else {
			// Parar VM se estiver rodando
			state, _ := getVMState(vm.Name)
			if state == "running" {
				color.Cyan("Parando VM %s...", vm.Name)
				execCommand("virsh", "destroy", vm.Name)
			}

			// Remover VM
			if err := execCommand("virsh", "undefine", vm.Name); err != nil {
				color.Red("❌ Falha ao remover VM %s do libvirt: %v", vm.Name, err)
			} else {
				color.Green("✅ VM %s removida do libvirt", vm.Name)
				destroyedCount++
			}
		}

		// Remover arquivo de disco
		diskFile := vm.Name + ".qcow2"
		if _, err := os.Stat(diskFile); err == nil {
			os.Remove(diskFile)
			color.Blue("💾 Arquivo de disco %s removido", diskFile)
		}
		fmt.Println()
	}

	color.Cyan("=== Resumo ===")
	fmt.Printf("VMs destruídas: %d\n", destroyedCount)
	fmt.Printf("VMs não existiam: %d\n", missingCount)
	fmt.Printf("Total de VMs no compose: %d\n", len(kvm.config.VMs))

	return nil
}

// List lista todas as VMs com seus status
func (kvm *KVMCompose) List() error {
	err := kvm.loadConfig()
	if err != nil {
		return err
	}

	color.Cyan("=== VMs disponíveis no %s ===", kvm.composeFile)

	fmt.Printf("%-15s\t%-8s\t%-5s\t%-5s\t%-15s\t%-15s\n", "Nome", "Memória", "vCPUs", "Disco", "IP", "Status")
	fmt.Printf("%-15s\t%-8s\t%-5s\t%-5s\t%-15s\t%-15s\n", "----", "-------", "-----", "-----", "--", "------")

	for _, vm := range kvm.config.VMs {
		// Aplicar valores padrão
		memory := vm.Memory
		if memory == 0 {
			memory = 4096
		}
		vcpus := vm.VCPUs
		if vcpus == 0 {
			vcpus = 4
		}
		diskSize := vm.DiskSize
		if diskSize == 0 {
			diskSize = 20
		}

		ip := "N/A"
		if len(vm.Networks) > 0 {
			ip = vm.Networks[0].IPv4
		}

		// Verificar status
		statusText := "⚪ not created"
		if vmExists(vm.Name) {
			state, _ := getVMState(vm.Name)
			switch state {
			case "running":
				statusText = "🟢 running"
			case "shut off":
				statusText = "🔴 stopped"
			case "paused":
				statusText = "🟡 paused"
			case "suspended":
				statusText = "🟠 suspended"
			default:
				statusText = "❓ " + state
			}
		}

		fmt.Printf("%-15s\t%dMB\t%d\t%dGB\t%-15s\t%s\n",
			vm.Name, memory, vcpus, diskSize, ip, statusText)
	}

	return nil
}

// showBanner exibe o banner colorido
func showBanner() {
	color.Cyan("============================================================")
	color.New(color.FgGreen, color.Bold).Println("🖥️  kvm-compose - Gerenciador de VMs KVM via arquivo compose")
	color.New(color.FgYellow, color.Bold).Println("📦 Versão: 1.0.0 Codename: \"Gopher Power\" - Dezembro de 2025")
	color.Cyan("============================================================")
	fmt.Println()
}

var (
	composeFile string
	rootCmd     = &cobra.Command{
		Use:   "kvm-compose",
		Short: "Gerenciador de VMs KVM via arquivo compose",
		Long:  `kvm-compose é uma ferramenta para gerenciar múltiplas VMs KVM usando um arquivo de configuração YAML estilo Docker Compose.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			showBanner()
		},
	}

	upCmd = &cobra.Command{
		Use:   "up",
		Short: "Criar e iniciar todas as VMs do compose",
		Run: func(cmd *cobra.Command, args []string) {
			kvm := NewKVMCompose(composeFile)
			if err := kvm.Up(); err != nil {
				color.Red("Erro: %v", err)
				os.Exit(1)
			}
		},
	}

	startCmd = &cobra.Command{
		Use:   "start",
		Short: "Iniciar todas as VMs do compose",
		Run: func(cmd *cobra.Command, args []string) {
			kvm := NewKVMCompose(composeFile)
			if err := kvm.Start(); err != nil {
				color.Red("Erro: %v", err)
				os.Exit(1)
			}
		},
	}

	stopCmd = &cobra.Command{
		Use:   "stop",
		Short: "Parar todas as VMs do compose",
		Run: func(cmd *cobra.Command, args []string) {
			kvm := NewKVMCompose(composeFile)
			if err := kvm.Stop(); err != nil {
				color.Red("Erro: %v", err)
				os.Exit(1)
			}
		},
	}

	downCmd = &cobra.Command{
		Use:   "down",
		Short: "Destruir todas as VMs do compose",
		Run: func(cmd *cobra.Command, args []string) {
			kvm := NewKVMCompose(composeFile)
			if err := kvm.Down(); err != nil {
				color.Red("Erro: %v", err)
				os.Exit(1)
			}
		},
	}

	listCmd = &cobra.Command{
		Use:   "list",
		Short: "Listar VMs disponíveis no compose",
		Run: func(cmd *cobra.Command, args []string) {
			kvm := NewKVMCompose(composeFile)
			if err := kvm.List(); err != nil {
				color.Red("Erro: %v", err)
				os.Exit(1)
			}
		},
	}
)

func init() {
	// Flags globais
	rootCmd.PersistentFlags().StringVarP(&composeFile, "compose", "c", "kvm-compose.yaml", "Arquivo compose")

	// Adicionar subcomandos
	rootCmd.AddCommand(upCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(downCmd)
	rootCmd.AddCommand(listCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		color.Red("Erro: %v", err)
		os.Exit(1)
	}
}