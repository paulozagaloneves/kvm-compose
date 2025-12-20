package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"gopkg.in/ini.v1"
	"gopkg.in/yaml.v3"
)

// KVMCompose é a estrutura principal do aplicativo
type KVMCompose struct {
	composeFile string
	config      Config
	appConfig   *AppConfig
}

// NewKVMCompose cria uma nova instância do KVMCompose
func NewKVMCompose(composeFile string) *KVMCompose {
	appConfig := loadAppConfig()
	return &KVMCompose{
		composeFile: composeFile,
		appConfig:   appConfig,
	}
}

// AppConfig representa as configurações globais da aplicação
type AppConfig struct {
	Main    MainConfig    `ini:"main"`
	Network NetworkConfig `ini:"network"`
	Images  ImagesConfig  `ini:"images"`
}

// MainConfig representa configurações principais
type MainConfig struct {
	Username   string `ini:"username"`
	SSHKeyFile string `ini:"ssh_key_file"`
}

// NetworkConfig representa configurações de rede
type NetworkConfig struct {
	Gateway     string `ini:"gateway"`
	Nameservers string `ini:"nameservers"`
}

// ImagesConfig representa configurações de imagens
type ImagesConfig struct {
	PathUpstreamImages string `ini:"path_upstream_images"`
	PathVMImages       string `ini:"path_vm_images"`
}

// VM representa uma máquina virtual no arquivo de configuração
type VM struct {
	Name       string    `yaml:"name"`
	Distro     string    `yaml:"distro"`
	Memory     int       `yaml:"memory"`
	VCPUs      int       `yaml:"vcpus"`
	DiskSize   int       `yaml:"disk_size"`
	Username   string    `yaml:"username"`
	Group      []string  `yaml:"group"`
	SSHKeyFile string    `yaml:"ssh_key_file"`
	Networks   []Network `yaml:"networks"`
}

// Network representa a configuração de rede de uma VM
type Network struct {
	HostBridge       string   `yaml:"host_bridge"`
	GuestIPv4        string   `yaml:"guest_ipv4"`
	GuestGateway4    string   `yaml:"guest_gateway4"`
	GuestNameservers []string `yaml:"guest_nameservers"`
}

// Config representa o arquivo de configuração completo
type Config struct {
	VMs []VM `yaml:",inline"`
}

// loadAppConfig carrega o arquivo de configuração INI
func loadAppConfig() *AppConfig {
	// Configurações padrão
	config := &AppConfig{
		Main: MainConfig{
			Username:   "admin",
			SSHKeyFile: "~/.ssh/id_rsa.pub",
		},
		Network: NetworkConfig{
			Gateway:     "192.168.1.1",
			Nameservers: "1.1.1.1,8.8.8.8",
		},
		Images: ImagesConfig{
			PathUpstreamImages: "~/.config/kvm-compose/images/upstream",
			PathVMImages:       "~/.config/kvm-compose/images/vm",
		},
	}

	// Caminhos para procurar o arquivo config.ini
	configPaths := []string{
		"./config.ini", // Diretório corrente
		expandPath("~/.config/kvm-compose/config.ini"), // Diretório de configuração do usuário
	}

	// Tentar carregar configuração de cada caminho
	for _, configPath := range configPaths {
		if _, err := os.Stat(configPath); err == nil {
			cfg, err := ini.Load(configPath)
			if err != nil {
				color.Yellow("⚠️  Erro ao carregar %s: %v", configPath, err)
				continue
			}

			// Fazer unmarshal das configurações
			if err := cfg.MapTo(config); err != nil {
				color.Yellow("⚠️  Erro ao fazer parse do %s: %v", configPath, err)
				continue
			}

			color.Green("✅ Configuração carregada de: %s", configPath)
			break
		}
	}

	return config
}

// getDefaultValues retorna valores padrão baseados na configuração
func (kvm *KVMCompose) getDefaultValues() (string, string, string, []string) {
	username := kvm.appConfig.Main.Username
	sshKeyFile := kvm.appConfig.Main.SSHKeyFile
	gateway := kvm.appConfig.Network.Gateway

	// Converter nameservers de string para slice
	// Remover colchetes se presentes e fazer split
	nameserversStr := strings.TrimSpace(kvm.appConfig.Network.Nameservers)
	nameserversStr = strings.Trim(nameserversStr, "[]")
	nameservers := strings.Split(nameserversStr, ",")
	for i, ns := range nameservers {
		nameservers[i] = strings.TrimSpace(ns)
	}

	return username, sshKeyFile, gateway, nameservers
}

// loadConfig carrega o arquivo YAML de configuração
func (kvm *KVMCompose) loadConfig() error {
	data, err := os.ReadFile(kvm.composeFile)
	if err != nil {
		return fmt.Errorf("erro ao ler arquivo %s: %v", kvm.composeFile, err)
	}

	err = yaml.Unmarshal(data, &kvm.config.VMs)
	if err != nil {
		return fmt.Errorf("erro ao fazer parse do YAML: %v", err)
	}

	return nil
}
