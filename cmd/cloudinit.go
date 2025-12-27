package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"text/template"

	"github.com/fatih/color"
)

// createCloudInitFiles cria os arquivos cloud-init para uma VM
func (kvm *KVMCompose) createCloudInitFiles(vm *VM) error {
	// Obter valores padrão da configuração
	defaultUsername, defaultSSHKeyFile, defaultGateway, defaultNameservers := kvm.getDefaultValues()

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
		vm.Username = defaultUsername
	}

	// Determinar qual arquivo de chave SSH usar
	sshKeyFile := vm.SSHKeyFile
	if sshKeyFile == "" {
		sshKeyFile = defaultSSHKeyFile
	}

	// Ler chave SSH
	sshKey := ""
	if sshKeyFile != "" {
		key, err := readSSHKey(sshKeyFile)
		if err != nil {
			color.Yellow("⚠️  Aviso: Não foi possível ler a chave SSH %s: %v", sshKeyFile, err)
		} else {
			sshKey = key
		}
	}

	// Estrutura de dados para os templates
	type userDataVars struct {
		Username     string
		SSHPublicKey string
	}
	type networkConfigVars struct {
		GuestIPv4        string
		GuestGateway4    string
		GuestNameservers []string
	}
	type metaDataVars struct {
		InstanceID string
		Hostname   string
	}

	// Função auxiliar para procurar template
	findTemplate := func(filename string) (string, error) {
		// 1. Procurar na pasta local templates/
		local := filepath.Join("templates", filename)
		if _, err := os.Stat(local); err == nil {
			return local, nil
		}
		// 2. Procurar em ~/.config/kvm-compose/templates/
		usr, _ := user.Current()
		if usr != nil {
			home := filepath.Join(usr.HomeDir, ".config", "kvm-compose", "templates", filename)
			if _, err := os.Stat(home); err == nil {
				return home, nil
			}
		}
		return "", fmt.Errorf("template %s não encontrado", filename)
	}

	// 1. user-data
	userDataFile, err := findTemplate("user-data.tmpl")
	var userDataContent string
	if err == nil {
		tmpl, err := template.ParseFiles(userDataFile)
		if err != nil {
			return err
		}
		var buf bytes.Buffer
		err = tmpl.Execute(&buf, userDataVars{Username: vm.Username, SSHPublicKey: sshKey})
		if err != nil {
			return err
		}
		userDataContent = buf.String()
	} else {
		// Fallback para template embutido
		userDataContent = fmt.Sprintf(`#cloud-config\nusers:\n  - name: %s\n    ssh_authorized_keys:\n      - %s\n    sudo: ['ALL=(ALL) NOPASSWD:ALL']\n    shell: /bin/bash\n    lock_passwd: false\n`, vm.Username, sshKey)
	}
	err = os.WriteFile(vm.Name+"-user-data.yaml", []byte(userDataContent), 0644)
	if err != nil {
		return err
	}

	// Criar network-config
	network := vm.Networks[0] // Assumindo apenas uma rede por VM
	if network.HostBridge == "" {
		network.HostBridge = "br0"
	}
	if network.GuestGateway4 == "" {
		network.GuestGateway4 = defaultGateway
	}
	if len(network.GuestNameservers) == 0 {
		network.GuestNameservers = defaultNameservers
	}

	// 2. network-config
	networkConfigFile, err := findTemplate("network-config.tmpl")
	var networkConfigContent string
	if err == nil {

		tmpl, err := template.ParseFiles(networkConfigFile)
		if err != nil {
			return err
		}
		var buf bytes.Buffer
		err = tmpl.Execute(&buf, networkConfigVars{
			GuestIPv4:        network.GuestIPv4,
			GuestGateway4:    network.GuestGateway4,
			GuestNameservers: network.GuestNameservers,
		})
		if err != nil {
			return err
		}
		networkConfigContent = buf.String()
	} else {
		nsYAML := ""
		for _, ns := range network.GuestNameservers {
			nsYAML += fmt.Sprintf("        - %s\n", ns)
		}
		networkConfigContent = fmt.Sprintf(`version: 2\nethernets:\n  enp1s0:\n    dhcp4: false\n    addresses: \n      - %s/24\n    gateway4: %s\n    nameservers:\n      addresses:\n%s`, network.GuestIPv4, network.GuestGateway4, nsYAML)
	}
	err = os.WriteFile(vm.Name+"-network-config.yaml", []byte(networkConfigContent), 0644)
	if err != nil {
		return err
	}

	// 3. meta-data
	metaDataFile, err := findTemplate("meta-data.tmpl")
	var metaDataContent string
	if err == nil {
		tmpl, err := template.ParseFiles(metaDataFile)
		if err != nil {
			return err
		}
		var buf bytes.Buffer
		err = tmpl.Execute(&buf, metaDataVars{InstanceID: vm.Name, Hostname: vm.Name})
		if err != nil {
			return err
		}
		metaDataContent = buf.String()
	} else {
		metaDataContent = fmt.Sprintf(`instance-id: %s\nlocal-hostname: %s\n`, vm.Name, vm.Name)
	}
	return os.WriteFile(vm.Name+"-meta-data.yaml", []byte(metaDataContent), 0644)
}

// cleanupCloudInitFiles remove os arquivos temporários de cloud-init
func cleanupCloudInitFiles(vmName string) {
	os.Remove(vmName + "-user-data.yaml")
	os.Remove(vmName + "-network-config.yaml")
	os.Remove(vmName + "-meta-data.yaml")
}
